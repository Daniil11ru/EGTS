package primary

import (
	"database/sql"

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

func (p *PrimaryRepository) GetVehicleByOIDAndProviderID(OID uint32, providerID int32) (types.Vehicle, error) {
	return p.Source.GetVehicleByOIDAndProviderID(OID, providerID)
}

func (p *PrimaryRepository) AddIndefiniteVehicle(OID uint32, providerID int32) (int32, error) {
	return p.Source.AddVehicle(types.Vehicle{
		IMEI:             int64(OID),
		OID:              sql.NullInt64{Int64: int64(OID), Valid: true},
		Name:             sql.NullString{String: "", Valid: false},
		ProviderID:       providerID,
		ModerationStatus: types.ModerationStatusPending,
	},
	)
}

func (p *PrimaryRepository) UpdateVehicleOID(id int32, OID uint32) error {
	return p.Source.UpdateVehicleOID(id, OID)
}

func (p *PrimaryRepository) GetProviderIDByIP(ip string) (int32, error) {
	provider, err := p.Source.GetProviderByIP(ip)
	return provider.ID, err
}

func (p *PrimaryRepository) AddVehicleMovement(data *util.NavigationRecord, vehicleID int) (int32, error) {
	return p.Source.AddVehicleMovement(data, vehicleID)
}

func (p *PrimaryRepository) GetLastVehiclePosition(vehicle_id int32) (types.Position, error) {
	return p.Source.GetLastVehiclePosition(vehicle_id)
}
