package loadgenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/golang/glog"
	"gitlab.slurm.io/sre_main/controlplane/apiclient"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

type SeanceCreator struct {
	city   string
	client *apiclient.Client
	params struct {
		RPS     uint `json:"rps"`
		Exactly bool `json:"exactly"`
	}
}

func NewSeanceCreatorStrategy(team teams.Team, mapi *apiclient.Main) (*SeanceCreator, error) {
	papi, err := apiclient.NewProvider(team.ProviderURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider API client for team %v", team)
	}

	sc := &SeanceCreator{
		city:   team.Name,
		client: apiclient.NewClient(mapi.Clone(), papi),
	}

	sc.params.RPS = defaultRate
	return sc, nil
}

func (sc *SeanceCreator) RunOnce(ctx context.Context) {
	id := fmt.Sprintf("%s-%s", sc.city, ctx.Value("worker"))
	glog.Infof("[%s] start new round", id)
	defer glog.Infof("[%s] finished round", id)

	price := GeneratePrice()
	time := GenerateFutureDateTime()

	var movieID int

	if sc.params.Exactly {
		movies, _ := sc.client.GetMovies(ctx, sc.city)
		if len(movies) > 0 {
			var upTo int

			if len(movies) > 4 {
				upTo = 4
			} else {
				upTo = len(movies)
			}

			movieID = movies[rand.Intn(upTo)].ID
		} else {
			glog.Errorf("Can't get movie list for %s", sc.city)
			return
		}
	} else {
		movieID = GenerateRandomID()
	}

	expectedErr := func(err error) bool {
		return strings.Contains(err.Error(), "Movie must exist")
	}

	cnt := 100
	if sc.params.RPS > 0 {
		cnt = int(sc.params.RPS)
	}

	for i := 0; i < cnt; i++ {
		if _, _, err := sc.client.CreateSeance(ctx, movieID, price, time); err != nil && !expectedErr(err) {
			glog.Errorf("[%s] unexpected error: %+#v", id, err)
		} else if err == nil {
			glog.Infof("[%s] accidental access, movie ID %d", id, movieID)
		}
	}
}

func (sc *SeanceCreator) Name() string {
	return "Create random seances for team: " + strings.Title(sc.city)
}

func (s *SeanceCreator) Type() string {
	return fmt.Sprintf("%T", SeanceCreator{})
}

func (sc *SeanceCreator) SetParams(params string) {
	if params != "" {
		if err := json.Unmarshal([]byte(params), &sc.params); err != nil {
			glog.Errorf("failed to parse params: %v", err)
			return
		}

		sc.client.SetRate(sc.params.RPS)
	}
}

func (sc *SeanceCreator) GetParams() string {
	body, _ := json.Marshal(&sc.params)
	return string(body)
}
