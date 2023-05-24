package teams

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/golang/glog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func list(redis *redis.Client, k8s *kubernetes.Clientset) ([]Team, error) {
	nss, err := k8s.CoreV1().Namespaces().List(v1.ListOptions{
		LabelSelector: "slurm.io/is_student_namespace==true",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var teams []Team
	for _, ns := range nss.Items {
		if tname := ns.Labels["slurm.io/team_name"]; tname != "" {
			team := Team{Name: tname, Namespace: ns.Name}
			team.ProviderURL = fmt.Sprintf("http://%s.%s", tname, ns.Name)

			val, err := redis.Get("is_service_mesh").Result()
			if err != nil {
				glog.Errorf("Error getting redis value %s: %s", "is_service_mesh", err)
			}
			if val == "true" {
				team.ApiGateway = fmt.Sprintf("http://%s.%s", "api-gateway", ns.Name)
			}

			if url := ns.Annotations["slurm.io/gitlab_url"]; url != "" {
				if !strings.HasSuffix(url, "/provider_backend.git") {
					url = strings.TrimSuffix(url, "/") + "/provider_backend.git"
				}

				team.ProviderBackendWebURL = strings.TrimSuffix(url, ".git")
				team.ProviderBackendURL = url
				team.ProjectURL = url

			}

			teams = append(teams, team)
		}
	}

	sort.Sort(TeamSorter(teams))
	return teams, nil
}

func writeToRedis(rd *redis.Client, teams []Team) error {
	tvals := make(map[string]interface{})
	for _, team := range teams {
		body, _ := json.Marshal(team)
		tvals[team.Name] = string(body)
	}

	nvals := make(map[string]interface{})
	for _, team := range teams {
		body, _ := json.Marshal(team)
		nvals[team.Namespace] = string(body)
	}

	_, err := rd.TxPipelined(func(pipe redis.Pipeliner) error {
		pipe.Del("list-of-teams", "list-of-namespaces").Result()
		pipe.HMSet("list-of-teams", tvals)
		pipe.HMSet("list-of-namespaces", nvals)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to update the list of names/namespaces in Redis: %v", err)
	}

	return nil
}

type TeamSorter []Team

func (t TeamSorter) Len() int {
	return len(t)
}

func (t TeamSorter) Less(i, j int) bool {
	return t[i].Name < t[j].Name
}

func (t TeamSorter) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
