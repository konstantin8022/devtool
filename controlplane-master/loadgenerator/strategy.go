package loadgenerator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Implement RunOnce() to perform the sequence of actions once, e.g.
// create a seance/tickets followed by buying all the tickets.
type Strategy interface {
	Name() string
	RunOnce(ctx context.Context)
	Type() string

	SetParams(params string)
	GetParams() string
}

// StrategyRunner keeps track of a running strategy. Can start/stop strategy execution.
type StrategyRunner struct {
	strategy   Strategy
	cancelFunc func()
	workers    int
	current    map[string]func()
	sync.Mutex
}

func NewStrategyRunner(strategy Strategy) *StrategyRunner {
	return &StrategyRunner{strategy: strategy, workers: 1}
}

func (s *StrategyRunner) Start(workers int, params string) {
	if s.IsRunning() {
		return
	}

	if workers < 1 {
		workers = 1
	}

	s.Lock()
	defer s.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel
	s.workers = workers
	s.current = make(map[string]func())

	s.strategy.SetParams(params)

	adjust := func() {
		s.Lock()
		defer s.Unlock()

		if len(s.current) < s.workers {
			worker := fmt.Sprintf("%04d", rand.Intn(1000))
			ctxl, cancel := context.WithCancel(ctx)
			s.current[worker] = cancel
			go s.worker(ctxl, worker)
		} else if len(s.current) > s.workers {
			for worker, cancel := range s.current {
				delete(s.current, worker)
				cancel()
				break
			}
		}
	}

	go func() {
		for {
			adjust()

			select {
			case <-time.After(300 * time.Millisecond):
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *StrategyRunner) Adjust(workers int) {
	s.Lock()
	defer s.Unlock()
	if s.cancelFunc == nil {
		// no started or stopped
		return
	}

	if workers > 0 {
		s.workers = workers
	}
}

func (s *StrategyRunner) worker(ctx context.Context, worker string) {
	defer func() {
		s.Lock()
		delete(s.current, worker)
		s.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.strategy.RunOnce(context.WithValue(ctx, "worker", worker))
		}
	}
}

func (s *StrategyRunner) Stop() {
	s.Lock()
	defer s.Unlock()
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}
}

func (s *StrategyRunner) IsRunning() bool {
	return s.WorkersRunning() > 0
}

func (s *StrategyRunner) WorkersRunning() int {
	s.Lock()
	defer s.Unlock()
	return len(s.current)
}

func (s *StrategyRunner) DesiredWorkers() int {
	s.Lock()
	defer s.Unlock()
	return s.workers
}

func (s *StrategyRunner) Strategy() Strategy {
	return s.strategy
}
