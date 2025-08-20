package repository

import (
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/filter"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/update"
	output "github.com/daniil11ru/egts/cli/receiver/dto/db/out"
	"github.com/daniil11ru/egts/cli/receiver/source"
)

type BusinessData interface {
	GetVehicle(vehicleId int32) (output.Vehicle, error)
	GetVehicles(filter filter.Vehicles) ([]output.Vehicle, error)
	GetLocations(filter filter.Locations) ([]output.Location, error)
	UpdateVehicleByImei(imei string, update update.VehicleByImei) error
	UpdateVehicleById(vehicleId int32, update update.VehicleById) error
}

type BusinessDataDefault struct {
	PostgreSource source.Primary
}

func NewBusinessDataDefault(postgreSource source.Primary) *BusinessDataDefault {
	return &BusinessDataDefault{PostgreSource: postgreSource}
}

func (r *BusinessDataDefault) GetVehicles(filter filter.Vehicles) ([]output.Vehicle, error) {
	return r.PostgreSource.GetVehicles(filter)
}

func (r *BusinessDataDefault) GetVehicle(vehicleId int32) (output.Vehicle, error) {
	return r.PostgreSource.GetVehicle(vehicleId)
}

func (r *BusinessDataDefault) GetLocations(filter filter.Locations) ([]output.Location, error) {
	return r.PostgreSource.GetLocations(filter)
}

func (r *BusinessDataDefault) UpdateVehicleById(vehicleId int32, update update.VehicleById) error {
	return r.PostgreSource.UpdateVehicleById(vehicleId, update)
}

func (r *BusinessDataDefault) UpdateVehicleByImei(imei string, update update.VehicleByImei) error {
	return r.PostgreSource.UpdateVehicleByImei(imei, update)
}
