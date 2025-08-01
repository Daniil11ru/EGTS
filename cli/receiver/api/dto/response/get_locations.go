package response

import "github.com/daniil11ru/egts/cli/receiver/api/model"

type VehicleTrack struct {
	VehicleId int32            `json:"vehicle_id"`
	Locations []model.Location `json:"locations"`
}

type GetLocations struct {
	VehicleTracks []VehicleTrack `json:"tracks"`
}
