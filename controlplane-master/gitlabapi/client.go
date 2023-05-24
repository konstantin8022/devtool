package gitlabapi

import (
	"fmt"
	"net/http"

	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

const (
	refreshInterval = 60
)

var (
	authorName  = "Василий Дубосекин"
	authorEmail = "v.dubosekin@slurm.io"
)

type Client struct {
	client *gitlab.Client
}

func NewClient(url, token string) *Client {
	hc := &http.Client{Transport: http.DefaultTransport}
	client := gitlab.NewClient(hc, token)
	client.SetBaseURL(url)

	return &Client{client: client}
}

func (c *Client) Commit(project int, branch, message string, actions []*gitlab.CommitAction) (*gitlab.Commit, error) {
	if len(actions) == 0 {
		return nil, nil
	}

	commit, _, err := c.client.Commits.CreateCommit(project, &gitlab.CreateCommitOptions{
		Branch:        &branch,
		CommitMessage: &message,
		AuthorName:    &authorName,
		AuthorEmail:   &authorEmail,
		Actions:       actions,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}

	return commit, err
}

func (c *Client) GetFileContent(project int, file, ref string) ([]byte, error) {
	body, _, err := c.client.RepositoryFiles.GetRawFile(project, file, &gitlab.GetRawFileOptions{Ref: &ref})
	return body, err
}

func (c *Client) CreateBranch(project int, branch, ref string) (*gitlab.Branch, error) {
	br, _, err := c.client.Branches.CreateBranch(project, &gitlab.CreateBranchOptions{Branch: &branch, Ref: &ref})
	return br, err
}

func (c *Client) CreateMergeRequest(project int, source, target string, title string) (*gitlab.MergeRequest, error) {
	mr, _, err := c.client.MergeRequests.CreateMergeRequest(project, &gitlab.CreateMergeRequestOptions{
		Title:        &title,
		SourceBranch: &source,
		TargetBranch: &target,
	})

	return mr, err
}

func (c *Client) GetCommit(project int, sha string) (*gitlab.Commit, error) {
	commit, _, err := c.client.Commits.GetCommit(project, sha)
	return commit, err
}

func (c *Client) GetMergeRequest(project int, mrequest int) (*gitlab.MergeRequest, error) {
	mr, _, err := c.client.MergeRequests.GetMergeRequest(project, mrequest, &gitlab.GetMergeRequestsOptions{})
	return mr, err
}

func (c *Client) GetProject(team teams.Team) (*gitlab.Project, error) {
	projects, err := c.GetProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.HTTPURLToRepo == team.ProviderBackendURL {
			return project, nil
		}
	}

	return nil, fmt.Errorf("project not found for team %q (%s)", team.Name, team.ProviderBackendURL)
}

func (c *Client) GetProjects() ([]*gitlab.Project, error) {
	projects, _, err := c.client.Projects.ListProjects(&gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	})

	return projects, err
}
