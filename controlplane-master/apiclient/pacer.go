package apiclient

import (
	"time"
)

type Pacer struct {
	rps    uint
	ticker *time.Ticker
}

func NewPacer(rps uint) *Pacer {
	if rps == 0 {
		return nil
	}

	rate := time.Second / time.Duration(rps)
	return &Pacer{rps: rps, ticker: time.NewTicker(rate)}
}

func (p *Pacer) Step() {
	if p == nil {
		return
	}

	<-p.ticker.C
}

func (p *Pacer) RPS() uint {
	if p == nil {
		return 0
	}

	return p.rps
}
