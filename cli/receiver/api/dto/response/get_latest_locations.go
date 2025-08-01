package response

import "github.com/daniil11ru/egts/cli/receiver/api/model"

type GetLatestLocations struct {
	LatestLocations []model.LatestLocation
}
