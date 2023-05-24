package dbreplication

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.slurm.io/sre_main/controlplane/teams"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	mysqlTimeout    = 10 * time.Second
	refreshInterval = 10
	schema          = "information_schema"
)

var (
	replicationEnabled = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dbreplication",
			Subsystem: "slave",
			Name:      "replication_enabled",
			Help:      "MySQL slave replication status",
		},
		[]string{"team_name", "namespace", "slave"},
	)

	replicationDelay = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dbreplication",
			Subsystem: "slave",
			Name:      "replication_delay_seconds",
			Help:      "MySQL slave replication delay",
		},
		[]string{"team_name", "namespace", "slave"},
	)

	replicationBehindMaster = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "dbreplication",
			Subsystem: "slave",
			Name:      "replication_behind_master_seconds",
			Help:      "MySQL slave replication behind master",
		},
		[]string{"team_name", "namespace", "slave"},
	)
)

type MySQLSlave struct {
	Name     string
	IP       string
	User     string
	Password string
}

func syncSlaveStatus(rd *redis.Client, k8s *kubernetes.Clientset, team teams.Team) {
	time.Sleep(time.Duration(rand.Intn(refreshInterval)))

	for {
		glog.Infof("periodical sync slave status for %q", team.Name)
		if err := syncSlaveStatusIteration(rd, k8s, team); err != nil {
			glog.Errorf("failed to sync slave status for %q: %v", team.Name, err)
		} else {
			glog.Infof("successfully synced slave status for %q", team.Name)
		}

		rnd := rand.Intn(refreshInterval)
		jittedDelay := (refreshInterval * time.Second) + (time.Duration(rnd) * time.Second)
		time.Sleep(jittedDelay)
	}
}

func syncSlaveStatusIteration(rd *redis.Client, k8s *kubernetes.Clientset, team teams.Team) error {
	slaves, err := discoverMySQLSlaves(k8s, team.Namespace)
	if err != nil {
		return err
	}

	var statuses []SlaveStatus
	for _, slave := range slaves {
		db, err := connectToMySQL(slave.IP, slave.User, slave.Password, schema)
		if err != nil {
			glog.Error(err)
			continue
		}

		defer db.Close()
		st, err := getSlaveStatus(db)
		if err != nil {
			glog.Error(err)
			continue
		}

		status := SlaveStatus{
			Name:   slave.Name,
			Status: st,
		}

		statuses = append(statuses, status)

		if status.GTIDEnabled() && !status.AutoPositionEnabled() {
			glog.Infof("enabling auto position for %q", slave.Name)
			if err := enableAutoPosition(db); err != nil {
				glog.Errorf("failed to enable auto position for %q: %v", slave.Name, err)
			}
		}
	}

	if err := WriteSlaveStatuses(rd, team.Namespace, statuses); err != nil {
		return err
	}

	for _, status := range statuses {
		replicationDelay.WithLabelValues(team.Name, team.Namespace, status.Name).Set(float64(status.Delay()))
		replicationBehindMaster.WithLabelValues(team.Name, team.Namespace, status.Name).Set(float64(status.BehindMaster()))

		if status.IsReplicating() {
			replicationEnabled.WithLabelValues(team.Name, team.Namespace, status.Name).Set(1)
		} else {
			replicationEnabled.WithLabelValues(team.Name, team.Namespace, status.Name).Set(0)
		}
	}

	return nil
}

func discoverMySQLSlaves(k8s *kubernetes.Clientset, namespace string) ([]MySQLSlave, error) {
	secretSelector := "app.kubernetes.io/name=mysql"
	services, err := k8s.CoreV1().Services(namespace).List(metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=mysql,app.kubernetes.io/component=secondary",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch list of services in namespace %q: %v", namespace, err)
	} else if len(services.Items) == 0 {
		secretSelector = "app=mysql"
		services, err = k8s.CoreV1().Services(namespace).List(metav1.ListOptions{
			LabelSelector: "app=mysql,component=slave",
		})

		if err != nil {
			return nil, fmt.Errorf("failed to fetch list of services in namespace %q: %v", namespace, err)
		} else if len(services.Items) == 0 {
			return nil, fmt.Errorf("failed to find service for mysql in namespace %q", namespace)
		}
	}

	service := services.Items[0]
	secrets, err := k8s.CoreV1().Secrets(namespace).List(metav1.ListOptions{
		LabelSelector: secretSelector,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to fetch list of secrets in namespace %q: %v", namespace, err)
	} else if len(secrets.Items) == 0 {
		return nil, fmt.Errorf("failed to find secret for mysql in namespace %q", namespace)
	}

	secret := secrets.Items[0]
	password := string(secret.Data["mysql-root-password"])
	username := "root"

	endpoints, err := k8s.CoreV1().Endpoints(service.Namespace).Get(service.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to find subset for service %q", service.Name)
	}

	var result []MySQLSlave
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			if address.TargetRef != nil {
				result = append(result, MySQLSlave{
					Name:     address.TargetRef.Name,
					IP:       address.IP,
					User:     username,
					Password: password,
				})
			}
		}
	}

	return result, nil
}

func setSlaveReplicationState(
	rd *redis.Client, k8s *kubernetes.Clientset,
	namespace string, slaveName string,
	desiredState bool, desiredDelay int,
) error {
	slaves, err := discoverMySQLSlaves(k8s, namespace)
	if err != nil {
		return err
	}

	var slave MySQLSlave
	for _, sl := range slaves {
		if sl.Name == slaveName {
			slave = sl
		}
	}

	if slave.Name == "" {
		return fmt.Errorf("failed to discover slave %q in %q", slaveName, namespace)
	}

	db, err := connectToMySQL(slave.IP, slave.User, slave.Password, schema)
	if err != nil {
		return err
	}

	defer db.Close()
	status, err := getSlaveStatus(db)
	if err != nil {
		return err
	}

	slaveStatus := SlaveStatus{
		Name:   slave.Name,
		Status: status,
	}

	if slaveStatus.Delay() != desiredDelay {
		glog.Infof("setting replication delay to %d (current one %d) for %q", desiredDelay, slaveStatus.Delay(), slave.Name)
		if err := setReplicationDelay(db, desiredDelay); err != nil {
			return err
		}
	}

	if slaveStatus.GTIDEnabled() && !slaveStatus.AutoPositionEnabled() {
		glog.Infof("enabling auto position for %q", slave.Name)
		if err := enableAutoPosition(db); err != nil {
			return err
		}
	}

	if slaveStatus.IsReplicating() != desiredState {
		if desiredState {
			glog.Infof("starting replication for %q", slave.Name)
			if err := startReplication(db); err != nil {
				return err
			}
		} else {
			glog.Infof("stopping replication for %q", slave.Name)
			if err := stopReplication(db); err != nil {
				return err
			}
		}
	}

	status, err = getSlaveStatus(db)
	if err != nil {
		return err
	}

	return WriteSlaveStatus(rd, namespace, SlaveStatus{
		Name:   slave.Name,
		Status: status,
	})
}

func connectToMySQL(host, user, passwd, schema string) (*sql.DB, error) {
	connstr := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s", user, passwd, host, schema, mysqlTimeout)
	glog.Infof("connecting to MySQL using connection string %q", connstr)

	db, err := sql.Open("mysql", connstr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %q: %v", host, err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}

	return db, nil
}

func getSlaveStatus(db *sql.DB) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mysqlTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SHOW SLAVE STATUS")
	if err != nil {
		return nil, fmt.Errorf("failed to get slave state: %v", err)
	}

	defer rows.Close()
	return scanMap(rows)
}

func scanMap(rows *sql.Rows) (map[string]string, error) {
	//https://github.com/gdaws/mysql-slave-status/blob/ec74abd477f86372fe567c2126ed04600397f71f/scan_map.go#L5
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		} else {
			return nil, nil
		}
	}

	values := make([]interface{}, len(columns))
	for index := range values {
		values[index] = new(sql.NullString)
	}

	if err := rows.Scan(values...); err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for index, cname := range columns {
		if nstr := *values[index].(*sql.NullString); nstr.Valid {
			result[cname] = nstr.String
		}
	}

	return result, nil
}

func startReplication(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), mysqlTimeout)
	defer cancel()

	if _, err := db.ExecContext(ctx, "START SLAVE;"); err != nil {
		return fmt.Errorf("failed to stop slave: %v", err)
	}

	return nil
}

func stopReplication(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), mysqlTimeout)
	defer cancel()

	if _, err := db.ExecContext(ctx, "STOP SLAVE;"); err != nil {
		return fmt.Errorf("failed to stop slave: %v", err)
	}

	return nil
}

func setReplicationDelay(db *sql.DB, delay int) error {
	ctx, cancel := context.WithTimeout(context.Background(), mysqlTimeout)
	defer cancel()

	if _, err := db.ExecContext(ctx, "STOP SLAVE;"); err != nil {
		return fmt.Errorf("failed to stop slave: %v", err)
	}

	_, rerr := db.ExecContext(ctx, fmt.Sprintf("CHANGE MASTER TO MASTER_DELAY=%d;", delay))

	if _, err := db.ExecContext(ctx, "START SLAVE;"); err != nil {
		return fmt.Errorf("failed to start slave: %v", err)
	}

	if rerr != nil {
		return fmt.Errorf("failed to set replication delay: %v", rerr)
	}

	return nil
}

func enableAutoPosition(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), mysqlTimeout)
	defer cancel()

	if _, err := db.ExecContext(ctx, "STOP SLAVE;"); err != nil {
		return fmt.Errorf("failed to stop slave: %v", err)
	}

	_, rerr := db.ExecContext(ctx, "CHANGE MASTER TO MASTER_AUTO_POSITION=1")

	if _, err := db.ExecContext(ctx, "START SLAVE;"); err != nil {
		return fmt.Errorf("failed to start slave: %v", err)
	}

	if rerr != nil {
		return fmt.Errorf("failed to set auto position: %v", rerr)
	}

	return nil
}
