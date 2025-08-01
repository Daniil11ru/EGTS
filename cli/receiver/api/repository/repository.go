package repository

import (
	"github.com/daniil11ru/egts/cli/receiver/api/dto/request"
	"github.com/daniil11ru/egts/cli/receiver/api/model"
)

type Repository interface {
	GetVehicles(request request.GetVehicles) ([]model.Vehicle, error)
	GetLocations(request request.GetLocations) ([]model.Location, error)
	GetLatestLocations(request request.GetLatestLocations) ([]model.LatestLocation, error)
}
