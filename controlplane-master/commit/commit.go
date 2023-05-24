package commit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/go-redis/redis/v7"
	"github.com/golang/glog"
	"github.com/xanzy/go-gitlab"
	"gitlab.slurm.io/sre_main/controlplane/gitlabapi"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

type Commit struct {
	glab  *gitlabapi.Client
	redis *redis.Client

	Name    string
	Message string
	Branch  string

	Diffs   []string
	Actions []*gitlab.CommitAction
}

func (c *Commit) GetType() string    { return "Commit" }
func (c *Commit) GetName() string    { return c.Name }
func (c *Commit) GetMessage() string { return c.Message }

func (c *Commit) Do(team teams.Team) error {
	if ci, err := c.load(team.Name); err != nil {
		return err
	} else if ci != nil {
		return errors.New("already commited")
	}

	project, err := c.glab.GetProject(team)
	if err != nil {
		return err
	}

	glog.Infof("[%s] preparing commit actions", team.Name)
	actions, err := c.GetCommitActions(project.ID)
	if err != nil {
		return fmt.Errorf("failed to get list of GitLab actions: %v", err)
	} else if len(actions) == 0 {
		return fmt.Errorf("nothing to commit")
	}

	glog.Infof("[%s] commiting to %s", team.Name, c.Branch)
	ci, err := c.glab.Commit(project.ID, c.Branch, c.Message, actions)
	if err != nil {
		return fmt.Errorf("failed to commit: %v", err)
	}

	if err := c.store(team.Name, ci); err != nil {
		return err
	}

	return nil
}

func (c *Commit) GetStatus(team teams.Team) (string, string, error) {
	ci, err := c.load(team.Name)
	if err != nil {
		return "", "", err
	} else if ci == nil {
		return "not commited", "", nil
	}

	commit, err := c.glab.GetCommit(ci.ProjectID, ci.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to query GitLab: %v", err)
	}

	status := "commited"
	if commit.Status != nil {
		status = status + " + " + string(*commit.Status)
	}

	if team.ProviderBackendWebURL == "" {
		return status, "", nil
	}

	return status, fmt.Sprintf("%s/commit/%s", team.ProviderBackendWebURL, ci.ID), nil
}

func (c *Commit) GetCommitActions(project int) ([]*gitlab.CommitAction, error) {
	var result []*gitlab.CommitAction

	for _, diff := range c.Diffs {
		files, _, err := gitdiff.Parse(bytes.NewBufferString(diff))
		if err != nil {
			return nil, fmt.Errorf("failed to parse diff: %v", err)
		} else if len(files) == 0 {
			continue
		}

		for _, file := range files {
			if file.IsNew || file.IsDelete || file.IsCopy || file.IsRename {
				return nil, fmt.Errorf("advance file ops via diff are not supported")
			}

			filename := file.NewName
			glog.Infof("requesting content of %q from GitLab", filename)
			body, err := c.glab.GetFileContent(project, filename, c.Branch)
			if err != nil {
				return nil, fmt.Errorf("failed to get content for file %q: %v", filename, err)
			}

			writer := bytes.NewBuffer(nil)
			appl := gitdiff.NewApplier(bytes.NewReader(body))
			if err := appl.ApplyFile(writer, file); err != nil {
				return nil, fmt.Errorf("failed to apply diff for file %q: %v", filename, err)
			}

			appl.Flush(writer)
			result = append(result, &gitlab.CommitAction{
				Action:   gitlab.FileUpdate,
				FilePath: filename,
				Content:  writer.String(),
			})

			glog.Infof("diff successfully applied to file %s", filename)
		}
	}

	if len(c.Actions) > 0 {
		result = append(result, c.Actions...)
	}

	return result, nil
}

func (c *Commit) GetFiles() string {
	var result []string
	for _, action := range c.Actions {
		result = append(result, fmt.Sprintf("%s %s", action.Action, action.FilePath))
	}

	for _, diff := range c.Diffs {
		files, _, _ := gitdiff.Parse(bytes.NewBufferString(diff))
		for _, file := range files {
			result = append(result, fmt.Sprintf("diff %s", file.NewName))
		}
	}

	return strings.Join(result, "; ")
}

func (c *Commit) store(team string, ci *gitlab.Commit) error {
	body, _ := json.Marshal(ci)
	key := fmt.Sprintf("commit-%s", c.Name)
	if _, err := c.redis.HSet(key, team, string(body)).Result(); err != nil {
		return fmt.Errorf("failed to store commit: %v", err)
	}

	return nil
}

func (c *Commit) load(team string) (*gitlab.Commit, error) {
	key := fmt.Sprintf("commit-%s", c.Name)
	body, err := c.redis.HGet(key, team).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed load commit: %v", err)
	}

	if body == "" {
		return nil, nil
	}

	var ci gitlab.Commit
	if err := json.Unmarshal([]byte(body), &ci); err != nil {
		return nil, fmt.Errorf("failed to unmarshal commit: %v", err)
	}

	return &ci, nil
}
