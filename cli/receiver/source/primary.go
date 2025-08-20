package source

import (
	"time"

	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/filter"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/insert"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/update"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/out"
)

type Primary interface {
	GetVehicles(filter filter.Vehicles) ([]out.Vehicle, error)
	GetVehicle(id int32) (out.Vehicle, error)
	UpdateVehicleByImei(imei string, update update.VehicleByImei) error
	UpdateVehicleById(id int32, update update.VehicleById) error
	AddVehicle(v insert.Vehicle) (int32, error)

	GetLocations(filter filter.Locations) ([]out.Location, error)
	GetLastVehiclePoint(id int32) (out.Point, error)
	GetTracks(after, before time.Time) ([]out.Track, error)
	AddLocation(insert insert.Location) (int32, error)
	DeleteLocation(id int32) error

	GetProviders() ([]out.Provider, error)

	GetApiKeys() ([]out.ApiKey, error)
}
