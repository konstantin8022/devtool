package gitlabapi

import (
	"os"

	"github.com/golang/glog"
)

func MustConnectToGitLab() *Client {
	url := os.Getenv("GITLAB_API")
	if url == "" {
		glog.Exit("no GITLAB_API")
	}

	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		glog.Exit("no GITLAB_TOKEN")
	}

	glog.Infof("connecting to GitLab using %q", url)
	return NewClient(url, token)
}
