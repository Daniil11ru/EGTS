package repository

import (
	"strconv"
	"time"

	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/filter"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/insert"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/update"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/out"
	"github.com/daniil11ru/egts/cli/receiver/dto/other"
	"github.com/daniil11ru/egts/cli/receiver/source"
)

type Primary struct {
	Source source.Primary
}

func (p *Primary) GetAllVehicles() ([]out.Vehicle, error) {
	return p.Source.GetVehicles(filter.Vehicles{})
}

func (p *Primary) GetVehicleModerationStatus(id int32) (other.ModerationStatus, error) {
	vehicle, err := p.Source.GetVehicle(id)
	return vehicle.ModerationStatus, err
}

func (p *Primary) GetVehiclesByProviderId(providerId int32) ([]out.Vehicle, error) {
	return p.Source.GetVehicles(filter.Vehicles{ProviderId: &providerId})
}

func (p *Primary) GetVehiclesByOIDAndProviderId(oid int64, providerId int32) ([]out.Vehicle, error) {
	return p.Source.GetVehicles(filter.Vehicles{
		OID:        &oid,
		ProviderId: &providerId,
	})
}

func (p *Primary) AddIndefiniteVehicle(oid int64, providerId int32) (int32, error) {
	return p.Source.AddVehicle(insert.Vehicle{
		IMEI:             strconv.FormatInt(oid, 10),
		OID:              &oid,
		ProviderId:       providerId,
		ModerationStatus: other.ModerationStatusPending,
	})
}

func (p *Primary) UpdateVehicleOid(id int32, oid int64) error {
	return p.Source.UpdateVehicleById(id, update.VehicleById{
		OID: &oid,
	})
}

func (p *Primary) GetAllProviders() ([]out.Provider, error) {
	return p.Source.GetProviders()
}

func (p *Primary) AddLocation(data *other.PacketData, vehicleId int32) (int32, error) {
	speed := int32(data.Speed)
	altitude := int64(data.Altitude)
	oid := int64(data.OID)

	return p.Source.AddLocation(insert.Location{
		VehicleId: vehicleId,
		OID:       oid,
		Latitude:  data.Latitude,
		Longitude: data.Longitude,
		Speed:     &speed,
		Altitude:  &altitude,
	})
}

func (p *Primary) GetLastVehiclePoint(vehicleId int32) (out.Point, error) {
	return p.Source.GetLastVehiclePoint(vehicleId)
}

func (p *Primary) GetTracks(after, before time.Time) ([]out.Track, error) {
	return p.Source.GetTracks(after, before)
}

func (p *Primary) DeleteLocation(locationId int32) error {
	return p.Source.DeleteLocation(locationId)
}
