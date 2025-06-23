package primary

import (
	"database/sql"

	"github.com/daniil11ru/egts/cli/receiver/source/primary"
)

type PrimaryRepository struct {
	Source primary.AuxiliaryInformationSource
}

func (p *PrimaryRepository) GetAllVehicles() ([]primary.Vehicle, error) {
	return p.Source.GetAllVehicles()
}

func (p *PrimaryRepository) GetVehicleModerationStatus(id int32) (primary.ModerationStatus, error) {
	vehicle, err := p.Source.GetVehicleByID(id)
	return vehicle.ModerationStatus, err
}

func (p *PrimaryRepository) GetVehiclesByProviderIP(ip string) ([]primary.Vehicle, error) {
	return p.Source.GetVehiclesByProviderIP(ip)
}

func (p *PrimaryRepository) GetVehicleByOIDAndProviderID(OID int32, providerID int32) (primary.Vehicle, error) {
	return p.Source.GetVehicleByOIDAndProviderID(OID, providerID)
}

func (p *PrimaryRepository) AddIndefiniteVehicle(OID int32, providerID int32) (int32, error) {
	return p.Source.AddVehicle(primary.Vehicle{
		IMEI:               int64(OID),
		OID:                sql.NullInt32{Int32: OID, Valid: true},
		LicensePlateNumber: sql.NullString{String: "", Valid: false},
		ProviderID:         providerID,
		ModerationStatus:   primary.ModerationStatusPending,
	},
	)
}

func (p *PrimaryRepository) UpdateVehicleOID(id int32, OID int32) error {
	return p.Source.UpdateVehicleOID(id, OID)
}

func (p *PrimaryRepository) GetProviderIDByIP(ip string) (int32, error) {
	provider, err := p.Source.GetProviderByIP(ip)
	return provider.ID, err
}

func (p *PrimaryRepository) AddVehicleMovement(message interface{ ToBytes() ([]byte, error) }, vehicleID int) (int32, error) {
	return p.Source.AddVehicleMovement(message, vehicleID)
}
