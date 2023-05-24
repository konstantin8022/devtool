package loadgenerator

// this strategy does the following steps
// 1. create a new movie
// 2. create a random number of seances [0, maxSeances)
// 3. book all the seats in all the seance in a random chunks
// 4. there is a fixed (lifetime) time period to perform this actions,

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gitlab.slurm.io/sre_main/controlplane/apiclient"
	"gitlab.slurm.io/sre_main/controlplane/teams"
)

const (
	email           = "load.generator@slurm.io"
	maxSeances      = 10
	defaultRate     = 10
	defaultLifetime = 120 // * time.Second

	movie   = "movie"
	seance  = "seance"
	seat    = "seat"
	success = "success"
	failure = "failure"
	remain  = "remain"
)

var (
	created = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "loadgenerator",
			Subsystem: "generic_strategy",
			Name:      "items_created",
			Help:      "Number of items available",
		},
		[]string{"city", "type", "result"},
	)

	booked = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "loadgenerator",
			Subsystem: "generic_strategy",
			Name:      "items_booked",
			Help:      "Number of items that were booked",
		},
		[]string{"city", "type", "result"},
	)

	requests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "loadgenerator",
			Subsystem: "generic_strategy",
			Name:      "requests_total",
			Help:      "Number of sent request",
		},
		[]string{"city"},
	)
)

type Generic struct {
	city          string
	client        *apiclient.Client
	params        GenericParams
	isServiceMesh bool
}

type GenericParams struct {
	RPS      uint `json:"rps"`
	Lifetime uint `json:"lifetime"`
	Direct   bool `json:"direct"`
}

func NewGenericStrategy(team teams.Team, mapi *apiclient.Main) (*Generic, error) {
	papi, err := apiclient.NewProvider(team.ProviderURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider API client for team %v", team)
	}

	return &Generic{
		city:   team.Name,
		client: apiclient.NewClient(mapi.Clone(), papi),
		params: GenericParams{
			RPS:      defaultRate,
			Lifetime: defaultLifetime,
		},
		isServiceMesh: len(team.ApiGateway) > 0,
	}, nil
}

func (s *Generic) Name() string {
	return "General load for team: " + strings.Title(s.city)
}

func (s *Generic) Type() string {
	return fmt.Sprintf("%T", Generic{})
}

func (s *Generic) RunOnce(ctx context.Context) {
	lifetime := time.Duration(s.params.Lifetime) * time.Second
	ctx, cancel := context.WithTimeout(ctx, lifetime)
	defer cancel()

	id := fmt.Sprintf("%s-%s", s.city, ctx.Value("worker"))
	glog.Infof("[%s] start new round", id)
	defer glog.Infof("[%s] finished round", id)

	requestsWithLabel := requests.WithLabelValues(s.city)

	var err error
	var movieID int
	film := GetRandomFilm()
	err = apiclient.Retry3(func() error {
		requestsWithLabel.Inc()
		movieID, err = s.client.CreateMovie(
			ctx,
			film.Title,
			film.Title, // TODO
			film.IMDBLink(),
		)
		return err
	})

	if err != nil {
		glog.Errorf("[%s] failed to create a movie: %v", id, err)
		created.WithLabelValues(s.city, movie, failure).Inc()
		return
	}

	glog.Infof("[%s] created a movie %q (ID: %d)", id, film.Title, movieID)
	created.WithLabelValues(s.city, movie, success).Inc()

	// keep track of seances and how many seats we didn't book in the end
	remainAvailableSeats := 0
	createdSeats := 0

	defer func() {
		if remainAvailableSeats > 0 {
			booked.WithLabelValues(s.city, seat, remain).Add(float64(remainAvailableSeats))
			glog.Infof("[%s] finishing round with reamining available seats: %d", id, remainAvailableSeats)
		}
	}()

	for i := rand.Intn(maxSeances); i >= 0; i = i - 1 {
		if shouldStop(ctx) {
			return
		}

		price := GeneratePrice()
		time := GenerateFutureDateTime()

		var seanceID int
		var seats []apiclient.Seat
		err := apiclient.Retry3(func() error {
			requestsWithLabel.Inc()
			seanceID, seats, err = s.client.CreateSeance(ctx, movieID, price, time)
			return err
		})

		if err != nil {
			created.WithLabelValues(s.city, seance, failure).Inc()
			glog.Errorf(
				"[%s] failed to create a seance for movie %d: %v",
				id, movieID, err,
			)
		} else {
			glog.Infof("[%s] created a seance %d", id, seanceID)
			created.WithLabelValues(s.city, seance, success).Inc()

			ln := len(seats)
			created.WithLabelValues(s.city, seat, success).Add(float64(ln))
			remainAvailableSeats += ln
			createdSeats += ln
		}
	}

	for {
		if shouldStop(ctx) {
			return
		}

		requestsWithLabel.Inc()
		seances, err := s.client.GetSeances(ctx, s.city, movieID)
		if err != nil {
			glog.Errorf(
				"[%s] failed to get seances for movie %d: %v",
				id, movieID, err,
			)
			continue
		}

		max := 1 + rand.Intn(10) // how many seats to book at this iteration [1.. 10]
		seanceID, seatIDs, totalSeats := SelectRandomSeats(seances, max)
		if len(seatIDs) == 0 {
			if totalSeats == createdSeats {
				// this condition means that all seats in all seances have been booked
				break
			} else {
				// this is only possible due to replication delay, wait until it catch up
				glog.Infof("[%s] no seats to book, there should be %d seats bookable, likely due to replication delay", id, remainAvailableSeats)
				continue
			}
		}

		ln := len(seatIDs)
		requestsWithLabel.Inc()

		if err := s.client.Book(ctx, s.city, movieID, seanceID, seatIDs, email, s.isServiceMesh); err != nil {
			glog.Errorf(
				"[%s] failed to book %d seats for movie %d seance %d: %v",
				id, ln, movieID, seanceID, err,
			)

			booked.WithLabelValues(s.city, seat, failure).Add(float64(ln))
		} else {
			glog.Infof(
				"[%s] successfully booked %d seats for movie %d seance %d",
				id, ln, movieID, seanceID,
			)

			booked.WithLabelValues(s.city, seat, success).Add(float64(ln))
			remainAvailableSeats -= ln
		}
	}
}

func (s *Generic) SetParams(params string) {
	if params != "" {
		if err := json.Unmarshal([]byte(params), &s.params); err != nil {
			glog.Errorf("failed to parse params: %v", err)
			return
		}

		s.client.SetMode(s.params.Direct)
		s.client.SetRate(s.params.RPS)
	}
}

func (s *Generic) GetParams() string {
	body, _ := json.Marshal(&s.params)
	return string(body)
}
