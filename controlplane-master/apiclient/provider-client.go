package apiclient

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"gitlab.slurm.io/sre_main/controlplane/teams"
)

type Provider struct {
	url string
	Sender
}

func NewProvider(apiurl string) (*Provider, error) {
	u, err := url.Parse(apiurl)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	return &Provider{url: u.String(), Sender: NewSender()}, nil
}

func NewCityProvider(city string) (*Provider, error) {
	team, ok := teams.Get(city)
	if !ok {
		return nil, fmt.Errorf("city/team %q not found", city)
	}

	return NewProvider(team.ProviderURL)
}

func (c *Provider) GetMovies(ctx context.Context) ([]Movie, error) {
	var result struct {
		Data []Movie `json:"data"`
	}

	url := fmt.Sprintf("%s/movies", c.url)
	if err := c.send(ctx, "GET", url, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch list of movies: %v", err)
	}

	return result.Data, nil
}

func (c *Provider) GetSeances(ctx context.Context, movie int) ([]Seance, error) {
	var result struct {
		Data []Seance `json:"data"`
	}

	url := fmt.Sprintf("%s/movies/%d/seances", c.url, movie)
	if err := c.send(ctx, "GET", url, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch list of seances: %v", err)
	}

	return result.Data, nil
}

func (c *Provider) CreateMovie(ctx context.Context, title, description, imageUrl string) (int, error) {
	request := struct {
		Movie
	}{
		Movie: Movie{
			Name:        title,
			Description: description,
			ImageURL:    imageUrl,
		},
	}

	var response struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}

	url := fmt.Sprintf("%s/movies/", c.url)
	if err := c.send(ctx, "POST", url, &request, &response); err != nil {
		return 0, fmt.Errorf("failed to create movie: %v", err)
	}

	return response.Data.ID, nil
}

func (c *Provider) CreateSeance(ctx context.Context, movie int, price int, time time.Time) (int, []Seat, error) {
	request := struct {
		Seance
	}{
		Seance: Seance{
			Price:    price,
			DateTime: time.Format("2006-01-02T15:04:05.000Z07:00"),
		},
	}

	var response struct {
		Data struct {
			ID    int    `json:"id"`
			Seats []Seat `json:"seats"`
		} `json:"data"`
	}

	url := fmt.Sprintf("%s/movies/%d/seances", c.url, movie)
	if err := c.send(ctx, "POST", url, &request, &response); err != nil {
		return 0, nil, fmt.Errorf("failed to create seance: %v", err)
	}

	if len(response.Data.Seats) == 0 {
		return 0, nil, fmt.Errorf("created a senace with no seats!")
	}

	return response.Data.ID, response.Data.Seats, nil
}

func (c *Provider) Book(ctx context.Context, seance int, seats []int, email string) error {
	request := struct {
		SeanceID int    `json:"seance_id"`
		Seats    []int  `json:"seatsIds"`
		Email    string `json:"email"`
	}{
		SeanceID: seance,
		Seats:    seats,
		Email:    email,
	}

	result := struct {
		Data []int `json:"data"`
	}{}

	url := fmt.Sprintf("%s/bookings", c.url)
	if err := c.send(ctx, "POST", url, &request, &result); err != nil {
		return fmt.Errorf("failed to make a booking: %w", err)
	}

	if len(request.Seats) != len(result.Data) {
		return fmt.Errorf("weird! amound of requested seats (%d) doesn't matched booked seats (%d)", len(request.Seats), len(result.Data))
	}

	return nil
}
