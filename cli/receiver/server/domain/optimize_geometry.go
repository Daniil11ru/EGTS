package domain

import (
	"fmt"
	"math"
	"time"

	"github.com/daniil11ru/egts/cli/receiver/dto/db/out"
	repository "github.com/daniil11ru/egts/cli/receiver/server/repository"
	"github.com/daniil11ru/egts/cli/receiver/util"
	"github.com/sirupsen/logrus"
)

type Line struct {
	First  out.Point
	Second out.Point
}

func (lineMovement Line) DistanceToPosition(position out.Point) float64 {
	a, b, c := lineMovement.Coefficients()
	return math.Abs(a*position.Longitude+b*position.Latitude+c) / math.Sqrt(a*a+b*b)
}

func (lineMovement Line) Coefficients() (a, b, c float64) {
	a = lineMovement.First.Latitude - lineMovement.Second.Latitude
	b = lineMovement.First.Longitude - lineMovement.Second.Longitude
	c = lineMovement.First.Longitude*lineMovement.Second.Latitude - lineMovement.First.Longitude*lineMovement.Second.Latitude

	return a, b, c
}

type OptimizeGeometry struct {
	PrimaryRepository repository.Primary
}

func (s *OptimizeGeometry) Run() error {
	logrus.Info("Запуск запланированной задачи оптимизации транспортных треков")

	tracks, err := s.PrimaryRepository.GetTracks(time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		return fmt.Errorf("не удалось получить треки: %w", err)
	}

	count := 0
	for _, track := range tracks {
		simplifiedTrack := GetSimplifiedTrack(track.Points, 0.0001)

		for _, movement := range track.Points {
			exists := false

			for _, movementID := range simplifiedTrack {
				if movement.LocationId == movementID {
					exists = true
					break
				}
			}

			if !exists {
				count++
				s.PrimaryRepository.DeleteLocation(movement.LocationId)
			}
		}
	}

	logrus.Info("Оптимизация транспортных треков завершена, удалено точек: ", count)

	return nil
}

func GetSimplifiedTrack(movements []out.Point, ep float64) []int32 {
	if len(movements) <= 2 {
		result := util.Map(movements, func(item out.Point) int32 { return item.LocationId })
		return result
	}

	line := Line{First: movements[0], Second: movements[len(movements)-1]}

	idx, maxDist := seekMostDistantPoint(line, movements)
	if maxDist >= ep {
		left := GetSimplifiedTrack(movements[:idx+1], ep)
		right := GetSimplifiedTrack(movements[idx:], ep)
		return append(left[:len(left)-1], right...)
	}

	return []int32{movements[0].LocationId, movements[len(movements)-1].LocationId}
}

func seekMostDistantPoint(lineMovement Line, movements []out.Point) (idx int, maxDist float64) {
	for i := 0; i < len(movements); i++ {
		d := lineMovement.DistanceToPosition(out.Point{Latitude: movements[i].Latitude, Longitude: movements[i].Longitude})
		if d > maxDist {
			maxDist = d
			idx = i
		}
	}

	return idx, maxDist
}
