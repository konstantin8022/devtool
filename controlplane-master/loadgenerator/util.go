package loadgenerator

import (
	"context"
	"math/rand"
	"time"

	"gitlab.slurm.io/sre_main/controlplane/apiclient"
)

func GetRandomFilm() Film {
	idx := rand.Intn(len(Films))
	return Films[idx]
}

func GeneratePrice() int {
	return 100 + (rand.Intn(8) * 100)
}

func GenerateFutureDateTime() time.Time {
	return time.Now().Add(24 * time.Hour * time.Duration(rand.Intn(30)))
}

func GenerateRandomID() int {
	return rand.Int()
}

// This function return upto 'max' seats in a seance. Seats and Seance selected randomly.
// This function also retrun boolean variable reflecting weather all seances contain information
// about all expected seats i.e. len(seances) * defaultAmountOfSeats
func SelectRandomSeats(seances []apiclient.Seance, max int) (int, []int, int) {
	rand.Shuffle(len(seances), func(i, j int) {
		seances[i], seances[j] = seances[j], seances[i]
	})

	totalSeats := 0
	for _, seance := range seances {
		seats := seance.Seats
		rand.Shuffle(len(seats), func(i, j int) {
			seats[i], seats[j] = seats[j], seats[i]
		})

		totalSeats += len(seats)

		var seatIDs []int
		for _, seat := range seats {
			if seat.Vacant {
				seatIDs = append(seatIDs, seat.ID)

				max--
				if max <= 0 {
					break
				}
			}
		}

		if len(seatIDs) > 0 {
			return seance.ID, seatIDs, 0
		}
	}

	return 0, nil, totalSeats
}

func shouldStop(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
