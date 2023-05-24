package dbreplication

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
)

const redisKeyPrefix = "slave-status"

type SlaveStatus struct {
	Name   string
	Status map[string]string
}

func (ss *SlaveStatus) Delay() int {
	v, _ := strconv.Atoi(ss.Status["SQL_Delay"])
	return v
}

func (ss *SlaveStatus) BehindMaster() int {
	v, _ := strconv.Atoi(ss.Status["Seconds_Behind_Master"])
	return v
}

func (ss *SlaveStatus) IsReplicating() bool {
	return strings.ToLower(ss.Status["Slave_IO_Running"]) == "yes" && strings.ToLower(ss.Status["Slave_SQL_Running"]) == "yes"
}

func (ss *SlaveStatus) ReplicatingStatus() string {
	mhost := ss.MasterHost()
	mhost = strings.TrimSuffix(mhost, "-master")
	mhost = strings.TrimSuffix(mhost, "-primary")

	if mhost != "" && !strings.Contains(ss.Name, mhost) {
		return "wrong master"
	}

	if ss.IsReplicating() {
		return "true"
	}

	return "false"
}

func (ss *SlaveStatus) LastUpdate() string {
	return ss.Status["__updated_at"]
}

func (ss *SlaveStatus) GTIDEnabled() bool {
	return ss.Status["Retrieved_Gtid_Set"] != "" || ss.Status["Executed_Gtid_Set"] != ""
}

func (ss *SlaveStatus) AutoPositionEnabled() bool {
	return ss.Status["Auto_Position"] == "1"
}

func (ss *SlaveStatus) MasterHost() string {
	return ss.Status["Master_Host"]
}

func WriteSlaveStatus(rd *redis.Client, name string, status SlaveStatus) error {
	now := time.Now().Format(time.RFC3339)
	key := redisKeyPrefix + "-" + name + "-" + status.Name
	_, err := rd.TxPipelined(func(pipe redis.Pipeliner) error {
		vals := make(map[string]interface{})
		for k, v := range status.Status {
			vals[k] = v
		}

		vals["__updated_at"] = now
		vals["__name"] = status.Name

		pipe.Del(key)
		pipe.HMSet(key, vals)
		return nil
	})

	return err
}

func WriteSlaveStatuses(rd *redis.Client, name string, statuses []SlaveStatus) error {
	now := time.Now().Format(time.RFC3339)
	prefix := redisKeyPrefix + "-" + name

	_, err := rd.TxPipelined(func(pipe redis.Pipeliner) error {
		keys, _ := pipe.Keys(prefix + "-*").Result()
		for _, key := range keys {
			pipe.Del(key)
		}

		for _, status := range statuses {
			vals := make(map[string]interface{})
			for k, v := range status.Status {
				vals[k] = v
			}

			vals["__updated_at"] = now
			vals["__name"] = status.Name

			pipe.HMSet(prefix+"-"+status.Name, vals)
		}

		return nil
	})

	return err
}

func ReadSlaveStatuses(rd *redis.Client, name string) ([]SlaveStatus, error) {
	var result []SlaveStatus
	prefix := redisKeyPrefix + "-" + name

	keys, _ := rd.Keys(prefix + "-*").Result()
	for _, key := range keys {
		status, _ := rd.HGetAll(key).Result()
		if len(status) != 0 {
			result = append(result, SlaveStatus{
				Name:   status["__name"],
				Status: status,
			})
		}
	}

	return result, nil
}

type SlaveStatusSorter []SlaveStatus

func (sss SlaveStatusSorter) Len() int {
	return len(sss)
}

func (sss SlaveStatusSorter) Less(i, j int) bool {
	return sss[i].Name < sss[j].Name
}

func (sss SlaveStatusSorter) Swap(i, j int) {
	sss[i], sss[j] = sss[j], sss[i]
}
