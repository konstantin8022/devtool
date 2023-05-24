package teams

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/golang/glog"
	"gitlab.slurm.io/sre_main/controlplane/sitemap"
	"k8s.io/client-go/kubernetes"
)

var (
	teams []Team
	mutex sync.Mutex
)

type Team struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	ProviderURL string `json:"providerURL"`
	ApiGateway  string `json:"apiGateway"`

	ProjectURL            string `json:"gitlabURL"`
	ProviderBackendURL    string `json:"providerBackendURL"`
	ProviderBackendWebURL string `json:"providerBackendWebURL"`
}

type handler struct{}

func Register(rd *redis.Client, k8s *kubernetes.Clientset) {
	glog.Infof("initial sync of list of teams")

	var err error
	if teams, err = list(rd, k8s); err != nil {
		glog.Errorf("failed to initially get list of teams: %v", err)
	} else if err := writeToRedis(rd, teams); err != nil {
		glog.Errorf("failed to write list of teams to Redis: %v", err)
	}

	go func() {
		for {
			time.Sleep(time.Minute)

			glog.Info("periodical scan of teams")
			if list, err := list(rd, k8s); err != nil {
				glog.Errorf("failed to get list of teams: %v", err)
			} else if err := writeToRedis(rd, list); err != nil {
				glog.Errorf("failed to write list of teams to Redis: %v", err)
			} else {
				mutex.Lock()
				teams = list
				mutex.Unlock()
			}
		}
	}()

	h := &handler{}
	sitemap.Handle("/teams", h)
	http.Handle("/list-of-teams", h)
}

func (_ *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	list := GetAll()
	body, _ := json.Marshal(&list)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func GetAll() []Team {
	mutex.Lock()
	defer mutex.Unlock()
	return teams
}

func Get(name string) (Team, bool) {
	for _, team := range GetAll() {
		if name == team.Name {
			return team, true
		}
	}

	return Team{}, false
}

func Exists(name string) bool {
	for _, team := range GetAll() {
		if name == team.Name {
			return true
		}
	}

	return false
}
