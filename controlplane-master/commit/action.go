package commit

import "gitlab.slurm.io/sre_main/controlplane/teams"

type Action interface {
	GetType() string
	GetName() string
	GetMessage() string
	GetFiles() string

	Do(team teams.Team) error
	GetStatus(team teams.Team) (string, string, error)
}
