package dbreplication

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/go-redis/redis/v7"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"

	"gitlab.slurm.io/sre_main/controlplane/sitemap"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

const (
	tmpl = `
		<html>
		<head>
			<title>DB replication</title>
			<meta http-equiv="refresh" content="30"/>
			<style>
				span.field {
					margin-right: 10px;
					display: block;
				}

				input.small {
					width: 50px;
				}
			</style>
		</head>
		<body>
		<div>
			{{ sitemap }}
			<h1>MySQL replication problems</h1>

			<form method="POST">
				<span class="field">
					Set replication delay for a slave
					<input name="default-slave-submit" type="submit" value="update" />
				</span>
				<span class="field">
					Delay, sec:
					<input type="number" name="default-slave-delay" value="0" class="small" />
				</span>
				<span class="field">
					Slave:
					<input type="radio" name="default-slave-number" value="0" checked /> 0
					<input type="radio" name="default-slave-number" value="1"/> 1
					<input type="radio" name="default-slave-number" value="2"/> 2
				</span>
			</form>

			{{ range . }}
			<hr/>
			<h3>{{ .Name }}</h3>
			<table>
			    {{ $cityName := .Name }}
				{{ range .Slaves }}  
				<form method="POST" action="?city={{ $cityName }}">
					<td>
						<span class="field">
							<strong>{{ .Name }}</strong>
							<input name="{{ .Name }}-submit" type="submit" value="update"/>
						</span>
						<span class="field">
							State:
							<strong>{{ prettyStatus .ReplicatingStatus }}</strong> (desired:
								<input type="radio" name="{{ .Name }}-replicating" value="enable" checked /> enable
								<input type="radio" name="{{ .Name }}-replicating" value="disable"/> disable
							)
						</span>
						<span class="field">
							Delay, sec:
							<strong>{{ prettySeconds .Delay }}</strong>
							(desired: <input name="{{ .Name }}-delay" type="number" value="0" class="small"/>)
						</span>
						<span class="field">
							Behind master, sec:
							<strong>{{ prettySeconds .BehindMaster }}</strong>
						</span>
					</td>
				</form>
				{{ end }}
			</table>
			{{ end }}
		</div>
		</body>
		</html>
	`

	enable  = "enable"
	disable = "disable"
)

type handler struct {
	redis *redis.Client
	k8s   *kubernetes.Clientset
}

type city struct {
	Name   string
	Slaves []SlaveStatus
}

func Register(redis *redis.Client, k8s *kubernetes.Clientset) {
	sitemap.Handle("/db-replication", &handler{
		redis: redis,
		k8s:   k8s,
	})

	for _, team := range teams.GetAll() {
		go syncSlaveStatus(redis, k8s, team)
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		query := r.URL.Query()
		if r.PostForm.Get("default-slave-submit") == "update" {
			desiredDelay, _ := strconv.Atoi(r.PostForm.Get("default-slave-delay"))
			desiredSlaveNumber, _ := strconv.Atoi(r.PostForm.Get("default-slave-number"))

			var wg sync.WaitGroup
			for _, team := range teams.GetAll() {
				wg.Add(1)
				go func(team teams.Team) {
					defer wg.Done()
					if slaves, _ := ReadSlaveStatuses(h.redis, team.Namespace); len(slaves) > desiredSlaveNumber {
						sort.Sort(SlaveStatusSorter(slaves))
						slave := slaves[desiredSlaveNumber]

						err := setSlaveReplicationState(
							h.redis, h.k8s,
							team.Namespace, slave.Name,
							true, desiredDelay,
						)

						if err != nil {
							glog.Errorf("failed to update desired state for %q: %v", team.Name, err)
						}
					}
				}(team)
			}

			wg.Wait()
		} else if team, ok := teams.Get(query.Get("city")); ok {
			for key := range r.PostForm {
				if !strings.HasSuffix(key, "-submit") {
					continue
				}

				slaveName := strings.TrimSuffix(key, "-submit")
				desiredState := r.PostForm.Get(slaveName + "-replicating")
				desiredDelay, _ := strconv.Atoi(r.PostForm.Get(slaveName + "-delay"))

				err := setSlaveReplicationState(
					h.redis, h.k8s,
					team.Namespace, slaveName,
					desiredState == enable, desiredDelay,
				)

				if err != nil {
					glog.Errorf("failed to update desired state for %q: %v", team.Name, err)
				}
			}
		}

		fallthrough

	case http.MethodGet:
		var cities []city
		for _, team := range teams.GetAll() {
			if slaves, err := ReadSlaveStatuses(h.redis, team.Namespace); err != nil {
				glog.Errorf("failed to get slave statuses: %v", err)
			} else {
				sort.Sort(SlaveStatusSorter(slaves))
				cities = append(cities, city{
					Name:   team.Name,
					Slaves: slaves,
				})
			}
		}

		funcmap := map[string]interface{}{
			"prettyStatus":  prettyStatus,
			"prettySeconds": prettySeconds,
		}

		if t, err := sitemap.Template("default").Funcs(funcmap).Parse(tmpl); err != nil {
			glog.Errorf("failed to parse template: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "text/html")
			if err := t.Execute(w, cities); err != nil {
				glog.Errorf("failed to execute template: %v", err)
			}
		}
	default:
		http.Error(w, "unknown method", http.StatusMethodNotAllowed)
	}
}

func prettySeconds(v int) template.HTML {
	color := "brown"
	if v == 0 {
		color = "green"
	} else if v <= 30 {
		color = "yellow"
	} else if v <= 60 {
		color = "orange"
	} else if v <= 100 {
		color = "red"
	}

	return template.HTML(fmt.Sprintf(`<span style="background-color: %s;">%d</span>`, color, v))
}

func prettyStatus(status string) template.HTML {
	if status == "true" {
		return template.HTML(`<span style="background-color: green">` + status + `</span>`)
	}

	return template.HTML(status)
}
