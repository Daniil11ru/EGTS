package response

import "github.com/daniil11ru/egts/cli/receiver/api/model"

type GetVehicles struct {
	Vehicles []model.Vehicle
}
