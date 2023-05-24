package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.slurm.io/sre_main/controlplane/apiclient"
	"gitlab.slurm.io/sre_main/controlplane/authproblems"
	"gitlab.slurm.io/sre_main/controlplane/commit"
	"gitlab.slurm.io/sre_main/controlplane/dbreplication"
	"gitlab.slurm.io/sre_main/controlplane/gitlabapi"
	"gitlab.slurm.io/sre_main/controlplane/loadgenerator"
	redisM "gitlab.slurm.io/sre_main/controlplane/redis"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

func main() {
	addr := flag.String("addr", ":4000", "")
	flag.Parse()

	redis := MustConnectToRedis()
	k8s := MustConnectToK8s()
	glab := gitlabapi.MustConnectToGitLab()
	mapi := apiclient.MustConnectToMainAPI()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result":"ok"}`))
	})

	teams.Register(redis, k8s)
	dbreplication.Register(redis, k8s)
	loadgenerator.Register(mapi)
	redisM.Register(redis)
	authproblems.Register(redis)
	commit.Register(glab, redis)
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		glog.Infof("starting HTTP server on %q", *addr)
		if err := http.ListenAndServe(*addr, nil); err != nil {
			glog.Fatalf("failed to server HTTP traffic: %v", err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	glog.Info("Bye!")
}
