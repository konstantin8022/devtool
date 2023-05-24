package commit

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v7"
	"github.com/golang/glog"
	"github.com/xanzy/go-gitlab"
	"gitlab.slurm.io/sre_main/controlplane/gitlabapi"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

type MergeRequest struct {
	glab  *gitlabapi.Client
	redis *redis.Client

	Name   string
	Title  string
	Branch string
	Ref    string

	Commit Commit
}

func (m *MergeRequest) GetType() string    { return "Merge Request" }
func (m *MergeRequest) GetName() string    { return m.Name }
func (m *MergeRequest) GetMessage() string { return m.Title }

func (m *MergeRequest) Do(team teams.Team) error {
	if mr, err := m.load(team.Name); err != nil {
		return err
	} else if mr != nil {
		return errors.New("already created")
	}

	project, err := m.glab.GetProject(team)
	if err != nil {
		return err
	}

	glog.Infof("[%s] creating branch %s from %s", team.Name, m.Branch, m.Ref)
	if _, err := m.glab.CreateBranch(project.ID, m.Branch, m.Ref); err != nil {
		return fmt.Errorf("failed to create branch: %v", err)
	}

	m.Commit.glab = m.glab
	m.Commit.redis = m.redis

	if m.Commit.Name == "" {
		m.Commit.Name = m.Name
	}

	if m.Commit.Message == "" {
		m.Commit.Message = m.Title
	}

	if m.Commit.Branch == "" {
		m.Commit.Branch = m.Branch
	}

	if err := m.Commit.Do(team); err != nil {
		return fmt.Errorf("failed to commit to branch: %v", err)
	}

	glog.Infof("[%s] creating merge request for %s", team.Name, m.Branch)
	mr, err := m.glab.CreateMergeRequest(project.ID, m.Branch, m.Ref, m.Title)
	if err != nil {
		return fmt.Errorf("failed to create merge request: %v", err)
	}

	if err := m.store(team.Name, mr); err != nil {
		return err
	}

	return nil
}

func (m *MergeRequest) GetStatus(team teams.Team) (string, string, error) {
	mr, err := m.load(team.Name)
	if err != nil {
		return "", "", err
	} else if mr == nil {
		return "not created", "", nil
	}

	mmr, err := m.glab.GetMergeRequest(mr.ProjectID, mr.IID)
	if err != nil {
		return "", "", fmt.Errorf("failed to query Gitlab: %v", err)
	}

	status := mmr.MergeStatus
	if mmr.Pipeline != nil {
		status = status + " + " + mmr.Pipeline.Status
	}

	return status, mmr.WebURL, nil
}

func (m *MergeRequest) GetFiles() string {
	return m.Commit.GetFiles()
}

func (m *MergeRequest) store(team string, mr *gitlab.MergeRequest) error {
	body, _ := json.Marshal(mr)
	key := fmt.Sprintf("merge-request-%s", m.Name)
	if _, err := m.redis.HSet(key, team, string(body)).Result(); err != nil {
		return fmt.Errorf("failed to store merge request: %v", err)
	}

	return nil
}

func (m *MergeRequest) load(team string) (*gitlab.MergeRequest, error) {
	key := fmt.Sprintf("merge-request-%s", m.Name)
	body, err := m.redis.HGet(key, team).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed load merge request: %v", err)
	}

	if body == "" {
		return nil, nil
	}

	var mr gitlab.MergeRequest
	body = strings.ReplaceAll(body, `"labels":"",`, "") // workaround stupid bug
	if err := json.Unmarshal([]byte(body), &mr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal merge reuqest: %v", err)
	}

	return &mr, nil
}
