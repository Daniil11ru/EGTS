package repository

import (
	"github.com/daniil11ru/egts/cli/receiver/api/dto/db/in/filter"
	"github.com/daniil11ru/egts/cli/receiver/api/dto/db/in/update"
	output "github.com/daniil11ru/egts/cli/receiver/api/dto/db/out"
	"github.com/daniil11ru/egts/cli/receiver/api/source"
)

type BusinessData interface {
	GetVehicle(vehicleId int32) (output.Vehicle, error)
	GetVehicles(filter filter.Vehicles) ([]output.Vehicle, error)
	GetLocations(filter filter.Locations) ([]output.Location, error)
	UpdateVehicleByImei(imei string, update update.VehicleByImei) error
	UpdateVehicleById(vehicleId int32, update update.VehicleById) error
}

type BusinessDataSimple struct {
	PostgreSource *source.Postgre
}

func NewBusinessDataSimple(postgreSource *source.Postgre) *BusinessDataSimple {
	return &BusinessDataSimple{PostgreSource: postgreSource}
}

func (r *BusinessDataSimple) GetVehicles(filter filter.Vehicles) ([]output.Vehicle, error) {
	return r.PostgreSource.GetVehicles(filter)
}

func (r *BusinessDataSimple) GetVehicle(vehicleId int32) (output.Vehicle, error) {
	return r.PostgreSource.GetVehicle(vehicleId)
}

func (r *BusinessDataSimple) GetLocations(filter filter.Locations) ([]output.Location, error) {
	return r.PostgreSource.GetLocations(filter)
}

func (r *BusinessDataSimple) UpdateVehicleById(vehicleId int32, update update.VehicleById) error {
	return r.PostgreSource.UpdateVehicleById(vehicleId, update)
}

func (r *BusinessDataSimple) UpdateVehicleByImei(imei string, update update.VehicleByImei) error {
	return r.PostgreSource.UpdateVehicleByImei(imei, update)
}
