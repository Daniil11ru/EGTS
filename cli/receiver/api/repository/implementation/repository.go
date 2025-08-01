package implementation

import (
	"github.com/daniil11ru/egts/cli/receiver/api/dto/request"
	"github.com/daniil11ru/egts/cli/receiver/api/model"
	postgre "github.com/daniil11ru/egts/cli/receiver/api/source/postgre"
)

type Repository struct {
	PostgreSource *postgre.Source
}

func New(postgreSource *postgre.Source) *Repository {
	return &Repository{PostgreSource: postgreSource}
}

func (r *Repository) GetVehicles(request request.GetVehicles) ([]model.Vehicle, error) {
	return r.PostgreSource.GetVehicles(request)
}

func (r *Repository) GetLocations(request request.GetLocations) ([]model.Location, error) {
	return r.PostgreSource.GetLocations(request)
}

func (r *Repository) GetLatestLocations(request request.GetLatestLocations) ([]model.LatestLocation, error) {
	return r.PostgreSource.GetLatestLocations(request)
}
