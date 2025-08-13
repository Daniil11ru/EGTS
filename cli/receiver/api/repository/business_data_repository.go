package repository

import (
	"github.com/daniil11ru/egts/cli/receiver/api/dto/request"
	"github.com/daniil11ru/egts/cli/receiver/api/model"
	"github.com/daniil11ru/egts/cli/receiver/api/source"
)

type BusinessData interface {
	GetVehicles(request request.GetVehicles) ([]model.Vehicle, error)
	GetLocations(request request.GetLocations) ([]model.Location, error)
	UpdateVehicle(request request.UpdateVehicle) error
}

type BusinessDataSimple struct {
	PostgreSource *source.Postgre
}

func NewBusinessDataSimple(postgreSource *source.Postgre) *BusinessDataSimple {
	return &BusinessDataSimple{PostgreSource: postgreSource}
}

func (r *BusinessDataSimple) GetVehicles(request request.GetVehicles) ([]model.Vehicle, error) {
	return r.PostgreSource.GetVehicles(request)
}

func (r *BusinessDataSimple) GetLocations(request request.GetLocations) ([]model.Location, error) {
	return r.PostgreSource.GetLocations(request)
}

func (r *BusinessDataSimple) UpdateVehicle(request request.UpdateVehicle) error {
	return r.PostgreSource.UpdateVehicle(request)
}
