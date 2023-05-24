// Slurm SRE Main ReST API Client.
// Construct with NewMain("http://localhost:5000")
package apiclient

import (
	"context"
	"fmt"
	"net/url"
)

type Main struct {
	url string
	Sender
}

func NewMain(apiurl string) (*Main, error) {
	u, err := url.Parse(apiurl)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	return &Main{url: u.String(), Sender: NewSender()}, nil
}

func (c *Main) Clone() *Main {
	return &Main{url: c.url, Sender: NewSender()}
}

func (c *Main) GetCities(ctx context.Context) ([]City, error) {
	var result struct {
		Data []City `json:"data"`
	}

	url := fmt.Sprintf("%s/cities", c.url)
	if err := c.send(ctx, "GET", url, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch list of cities: %v", err)
	}

	return result.Data, nil
}

func (c *Main) GetMovies(ctx context.Context, city string) ([]Movie, error) {
	var result struct {
		Data []Movie `json:"data"`
	}

	url := fmt.Sprintf("%s/cities/%s/movies", c.url, url.QueryEscape(city))
	if err := c.send(ctx, "GET", url, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch list of movies: %v", err)
	}

	return result.Data, nil
}

func (c *Main) GetSeances(ctx context.Context, city string, movie int) ([]Seance, error) {
	var result struct {
		Data []Seance `json:"data"`
	}

	url := fmt.Sprintf("%s/cities/%s/movies/%d/seances", c.url, url.QueryEscape(city), movie)
	if err := c.send(ctx, "GET", url, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch list of seances: %v", err)
	}

	return result.Data, nil
}

func (c *Main) Book(ctx context.Context, city string, movie, seance int, seats []int, email string, isServiceMesh bool) error {
	request := struct {
		Email string `json:"email"`
		Seats []int  `json:"seatsIds"`
		Card  string `json:"card,omitempty"`
	}{
		Email: email,
		Seats: seats,
	}

	result := struct {
		Data []int `json:"data"`
	}{}

	if isServiceMesh {
		request.Card = "anydata"
	}

	url := fmt.Sprintf("%s/cities/%s/movies/%d/seances/%d/bookings", c.url, url.QueryEscape(city), movie, seance)
	if err := c.send(ctx, "POST", url, &request, &result); err != nil {
		return fmt.Errorf("failed to make a booking: %w", err)
	}

	if len(request.Seats) != len(result.Data) {
		return fmt.Errorf("weird! amound of requested seats (%d) doesn't matched booked seats (%d)", len(request.Seats), len(result.Data))
	}

	return nil
}
