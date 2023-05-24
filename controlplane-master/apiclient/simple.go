package apiclient

import (
	"os"

	"github.com/golang/glog"
)

func MustConnectToMainAPI() *Main {
	url := os.Getenv("MAIN_API")
	if url == "" {
		glog.Exit("no MAIN_API")
	}

	glog.Infof("connecting to main API using %q", url)
	api, err := NewMain(url)
	if err != nil {
		glog.Exitf("failed to create main API client: %v", err)
	}

	return api
}
