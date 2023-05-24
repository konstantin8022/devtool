package authproblems

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
)

const (
	authProblemsRedisKeyPrefix = "authproblems"
)

type state struct {
	Name      string
	Namespace string

	SlowdownProbability int
	SlowdownMin         int
	SlowdownMax         int

	ErrorProbability int

	Active bool

	LastUpdate string

	Errors []error
}

func NewState(name string) state {
	return state{
		Name: name,

		SlowdownProbability: 0,
		SlowdownMin:         0,
		SlowdownMax:         0,

		ErrorProbability: 0,
	}
}

func (s *state) Save(rd *redis.Client) error {
	vals := map[string]interface{}{
		"updated_at":           time.Now().Format("Jan 2 15:04:05"),
		"slowdown-probability": strconv.Itoa(s.SlowdownProbability),
		"slowdown-min":         strconv.Itoa(s.SlowdownMin),
		"slowdown-max":         strconv.Itoa(s.SlowdownMax),
		"error-probability":    strconv.Itoa(s.ErrorProbability),
		"active":               fmt.Sprintf("%v", s.Active),
	}

	key := authProblemsRedisKeyPrefix + "-" + s.Name

	return rd.HMSet(key, vals).Err()
}

func (s *state) ReadState(rd *redis.Client) error {
	key := authProblemsRedisKeyPrefix + "-" + s.Name

	// redis structure: hash called "authproblems-$team", keys: "slowdown-probability", "slowdown-min", "slowdown-max", "error-probability"
	vals, err := rd.HMGet(key,
		"slowdown-probability",
		"slowdown-min",
		"slowdown-max",
		"error-probability",
		"active",
		"updated_at",
	).Result()

	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to fetch info from Redis: %v", err)
	} else if len(vals) != 6 {
		return fmt.Errorf("unexpected result length")
	}

	if vals[0] != nil {
		slowdownProbRaw := fmt.Sprintf("%v", vals[0])
		slowdownProb, err := strconv.Atoi(slowdownProbRaw)
		if err != nil {
			return fmt.Errorf("team %q could not convert slowdownProb %q to int", s.Name, slowdownProbRaw)
		}

		s.SlowdownProbability = slowdownProb
	}

	if vals[1] != nil {
		slowdownMinRaw := fmt.Sprintf("%v", vals[1])
		slowdownMin, err := strconv.Atoi(slowdownMinRaw)
		if err != nil {
			return fmt.Errorf("team %q could not convert slowdownMin %q to int", s.Name, slowdownMinRaw)
		}

		s.SlowdownMin = slowdownMin
	}

	if vals[2] != nil {
		slowdownMaxRaw := fmt.Sprintf("%v", vals[2])
		slowdownMax, err := strconv.Atoi(slowdownMaxRaw)
		if err != nil {
			return fmt.Errorf("team %q could not convert slowdownMax %q to int", s.Name, slowdownMaxRaw)
		}

		s.SlowdownMax = slowdownMax
	}

	if vals[3] != nil {
		errorProbRaw := fmt.Sprintf("%v", vals[3])
		errorProb, err := strconv.Atoi(errorProbRaw)
		if err != nil {
			return fmt.Errorf("team %q could not convert errorProb %q to int", s.Name, errorProbRaw)
		}

		s.ErrorProbability = errorProb
	}

	if vals[4] != nil {
		// need to do this because it's an interface{}
		activeRaw := fmt.Sprintf("%v", vals[4])
		if activeRaw == "true" || s.Name == "default" {
			s.Active = true
		} else {
			s.Active = false
		}
	}

	if vals[5] == nil {
		s.LastUpdate = "never"
	} else {
		// need to do this because it's an interface{} and I prefer this over casting
		s.LastUpdate = fmt.Sprintf("%v", vals[5])
	}

	return nil
}
