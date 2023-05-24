package loadgenerator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

type MainRunners interface {
	isMainRunner() bool
	SetWorkers(count int)
	GetType() string
	GetName() string
	SetParams(string)
}

type mainGenericRunner struct {
	Workers int
	params  struct {
		RPS      uint `json:"rps"`
		Lifetime uint `json:"lifetime"`
		Direct   bool `json:"direct"`
	}
}

func (s *mainGenericRunner) GetParams() string {
	body, _ := json.Marshal(&s.params)
	return string(body)
}

func (s *mainGenericRunner) SetParams(params string) {
	if params != "" {
		if err := json.Unmarshal([]byte(params), &s.params); err != nil {
			glog.Errorf("failed to parse params: %v", err)
			return
		}
	}
}
func (s *mainGenericRunner) isMainRunner() bool {
	return true
}

func (s *mainGenericRunner) SetWorkers(count int) {
	s.Workers = count
}

func (s *mainGenericRunner) GetType() string {
	return fmt.Sprintf("%T", Generic{})
}

func (s *mainGenericRunner) GetName() string {
	return "Main general loadgenerator"
}

type mainSeancesRunner struct {
	Workers int
	params  struct {
		RPS     uint `json:"rps"`
		Exactly bool `json:"exactly"`
	}
}

func (s *mainSeancesRunner) GetParams() string {
	body, _ := json.Marshal(&s.params)
	return string(body)
}

func (s *mainSeancesRunner) SetParams(params string) {
	if params != "" {
		if err := json.Unmarshal([]byte(params), &s.params); err != nil {
			glog.Errorf("failed to parse params: %v", err)
			return
		}
	}
}

func (s *mainSeancesRunner) isMainRunner() bool {
	return true
}

func (s *mainSeancesRunner) SetWorkers(count int) {
	s.Workers = count
}

func (s *mainSeancesRunner) GetType() string {
	return fmt.Sprintf("%T", SeanceCreator{})
}

func (s *mainSeancesRunner) GetName() string {
	return "Main seances loadgenerator"
}

func runMainLoadGenerator(h *handler, r *http.Request) {
	handler := func(action string, mainRunner MainRunners) {
		workers, _ := strconv.Atoi(r.PostFormValue(mainRunner.GetName() + "-workers"))
		params := r.PostFormValue(mainRunner.GetName() + "-params")

		switch action {
		case "Go":
			for _, runner := range h.runners {
				if strings.Contains(runner.strategy.Type(), mainRunner.GetType()) {
					mainRunner.SetParams(params)
					runner.Start(workers, params)
				}
			}
		case "Stop":
			for _, runner := range h.runners {
				if strings.Contains(runner.strategy.Type(), mainRunner.GetType()) {
					runner.Stop()
				}
			}
		case "Adjust":
			for _, runner := range h.runners {
				if strings.Contains(runner.strategy.Type(), mainRunner.GetType()) {
					mainRunner.SetWorkers(workers)
					runner.workers = workers
				}
			}
		}
	}

	if mainGenericRunner := r.PostFormValue(h.mainGenericRunner.GetName()); mainGenericRunner != "" {
		handler(mainGenericRunner, h.mainGenericRunner)
	}

	if mainSeancesRunner := r.PostFormValue(h.mainSeancesRunner.GetName()); mainSeancesRunner != "" {
		handler(mainSeancesRunner, h.mainSeancesRunner)
	}
}
