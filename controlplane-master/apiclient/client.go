package apiclient

import (
	"context"
	"time"
)

type Client struct {
	mapi   *Main
	papi   *Provider
	direct bool
}

func NewClient(mapi *Main, papi *Provider) *Client {
	return &Client{mapi: mapi, papi: papi}
}

func (c *Client) GetMovies(ctx context.Context, city string) ([]Movie, error) {
	if c.direct {
		return c.papi.GetMovies(ctx)
	} else {
		return c.mapi.GetMovies(ctx, city)
	}
}

func (c *Client) GetSeances(ctx context.Context, city string, movie int) ([]Seance, error) {
	if c.direct {
		return c.papi.GetSeances(ctx, movie)
	} else {
		return c.mapi.GetSeances(ctx, city, movie)
	}
}

func (c *Client) CreateMovie(ctx context.Context, title, description, imageUrl string) (int, error) {
	return c.papi.CreateMovie(ctx, title, description, imageUrl)
}

func (c *Client) CreateSeance(ctx context.Context, movie int, price int, time time.Time) (int, []Seat, error) {
	return c.papi.CreateSeance(ctx, movie, price, time)
}

func (c *Client) Book(ctx context.Context, city string, movie, seance int, seats []int, email string, isServiceMesh bool) error {
	if c.direct {
		return c.papi.Book(ctx, seance, seats, email)
	} else {
		return c.mapi.Book(ctx, city, movie, seance, seats, email, isServiceMesh)
	}
}

func (c *Client) SetMode(direct bool) {
	c.direct = direct
}

func (c *Client) SetRate(rate uint) {
	if c.mapi != nil {
		c.mapi.Pacer = NewPacer(rate)
	}

	if c.papi != nil {
		c.papi.Pacer = NewPacer(rate)
	}
}
