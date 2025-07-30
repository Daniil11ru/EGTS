package primary

import (
	"database/sql"
	"time"

	"github.com/daniil11ru/egts/cli/receiver/repository/primary/types"
	"github.com/daniil11ru/egts/cli/receiver/repository/util"
	"github.com/daniil11ru/egts/cli/receiver/source/primary"
)

type PrimaryRepository struct {
	Source primary.PrimarySource
}

func (p *PrimaryRepository) GetAllVehicles() ([]types.Vehicle, error) {
	return p.Source.GetAllVehicles()
}

func (p *PrimaryRepository) GetVehicleModerationStatus(id int32) (types.ModerationStatus, error) {
	vehicle, err := p.Source.GetVehicleByID(id)
	return vehicle.ModerationStatus, err
}

func (p *PrimaryRepository) GetVehiclesByProviderIP(ip string) ([]types.Vehicle, error) {
	return p.Source.GetVehiclesByProviderIP(ip)
}

func (p *PrimaryRepository) GetVehiclesByOIDAndProviderID(OID uint32, providerID int32) ([]types.Vehicle, error) {
	return p.Source.GetVehiclesByOIDAndProviderID(OID, providerID)
}

func (p *PrimaryRepository) AddIndefiniteVehicle(OID uint32, providerID int32) (int32, error) {
	return p.Source.AddVehicle(types.Vehicle{
		IMEI:             int64(OID),
		Name:             sql.NullString{String: "", Valid: false},
		ProviderID:       providerID,
		ModerationStatus: types.ModerationStatusPending,
	},
	)
}

func (p *PrimaryRepository) UpdateVehicleOID(id int32, OID uint32) error {
	return p.Source.UpdateVehicleOID(id, OID)
}

func (p *PrimaryRepository) GetAllProviders() ([]types.Provider, error) {
	return p.Source.GetAllProviders()
}

func (p *PrimaryRepository) GetProviderIDByIP(ip string) (int32, error) {
	provider, err := p.Source.GetProviderByIP(ip)
	return provider.ID, err
}

func (p *PrimaryRepository) AddVehicleMovement(data *util.NavigationRecord, vehicleID int) (int32, error) {
	return p.Source.AddVehicleMovement(data, vehicleID)
}

func (p *PrimaryRepository) GetLastVehiclePosition(vehicleID int32) (types.Position3D, error) {
	return p.Source.GetLastVehiclePosition(vehicleID)
}

func (p *PrimaryRepository) GetTracks2DOfAllVehicles(after, before time.Time) ([]types.Track2D, error) {
	return p.Source.GetTracks2DOfAllVehicles(after, before)
}

func (p *PrimaryRepository) DeleteVehicleMovement(vehicleMovementID int32) error {
	return p.Source.DeleteVehicleMovement(vehicleMovementID)
}
