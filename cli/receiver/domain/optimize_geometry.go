package domain

import (
	"fmt"
	"math"
	"time"

	repository "github.com/daniil11ru/egts/cli/receiver/repository/primary"
	"github.com/daniil11ru/egts/cli/receiver/repository/primary/types"
	"github.com/daniil11ru/egts/cli/receiver/repository/util"
)

type LineMovement struct {
	First  types.Movement2D
	Second types.Movement2D
}

func (lineMovement LineMovement) DistanceToPosition(position types.Position2D) float64 {
	a, b, c := lineMovement.Coefficients()
	return math.Abs(a*position.Longitude+b*position.Latitude+c) / math.Sqrt(a*a+b*b)
}

func (lineMovement LineMovement) Coefficients() (a, b, c float64) {
	a = lineMovement.First.Latitude - lineMovement.Second.Latitude
	b = lineMovement.First.Longitude - lineMovement.Second.Longitude
	c = lineMovement.First.Longitude*lineMovement.Second.Latitude - lineMovement.First.Longitude*lineMovement.Second.Latitude

	return a, b, c
}

type OptimizeGeometry struct {
	PrimaryRepository repository.PrimaryRepository
}

func (s *OptimizeGeometry) Run() error {
	tracks, err := s.PrimaryRepository.GetTracks2DOfAllVehicles(time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		return fmt.Errorf("не удалось получить треки: %w", err)
	}

	for _, track := range tracks {
		simplifiedTrack := GetSimplifiedTrack(track.Movements, 0.0001)

		for _, movement := range track.Movements {
			exists := false

			for _, movementID := range simplifiedTrack {
				if movement.ID == movementID {
					exists = true
					break
				}
			}

			if !exists {
				s.PrimaryRepository.DeleteVehicleMovement(movement.ID)
			}
		}
	}

	return nil
}

func GetSimplifiedTrack(movements []types.Movement2D, ep float64) []int32 {
	if len(movements) <= 2 {
		result := util.Map(movements, func(item types.Movement2D) int32 { return item.ID })
		return result
	}

	line := LineMovement{First: movements[0], Second: movements[len(movements)-1]}

	idx, maxDist := seekMostDistantPoint(line, movements)
	if maxDist >= ep {
		left := GetSimplifiedTrack(movements[:idx+1], ep)
		right := GetSimplifiedTrack(movements[idx:], ep)
		return append(left[:len(left)-1], right...)
	}

	return []int32{movements[0].ID, movements[len(movements)-1].ID}
}

func seekMostDistantPoint(lineMovement LineMovement, movements []types.Movement2D) (idx int, maxDist float64) {
	for i := 0; i < len(movements); i++ {
		d := lineMovement.DistanceToPosition(types.Position2D{Latitude: movements[i].Latitude, Longitude: movements[i].Longitude})
		if d > maxDist {
			maxDist = d
			idx = i
		}
	}

	return idx, maxDist
}
