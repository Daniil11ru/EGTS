package auxiliary

import (
	"database/sql"

	aux "github.com/daniil11ru/egts/cli/receiver/source/auxiliary"
)

type AuxiliaryInformationRepository struct {
	Source aux.AuxiliaryInformationSource
}

func (auxInfoRepo *AuxiliaryInformationRepository) GetAllVehicles() ([]aux.Vehicle, error) {
	return auxInfoRepo.Source.GetAllVehicles()
}

func (auxInfoRepo *AuxiliaryInformationRepository) GetVehiclesByProviderIP(ip string) ([]aux.Vehicle, error) {
	return auxInfoRepo.Source.GetVehiclesByProviderIP(ip)
}

func (auxInfoRepo *AuxiliaryInformationRepository) GetVehicleByOID(OID int32) (aux.Vehicle, error) {
	return auxInfoRepo.Source.GetVehicleByOID(OID)
}

func (auxInfoRepo *AuxiliaryInformationRepository) GetVehicleByOIDAndProviderID(OID int32, providerID int32) (aux.Vehicle, error) {
	return auxInfoRepo.Source.GetVehicleByOIDAndProviderID(OID, providerID)
}

func (auxInfoRepo *AuxiliaryInformationRepository) AddIndefiniteVehicle(OID int32, providerID int32) (int32, error) {
	return auxInfoRepo.Source.AddVehicle(aux.Vehicle{
		IMEI:               int64(OID),
		OID:                sql.NullInt32{Int32: OID, Valid: true},
		LicensePlateNumber: sql.NullString{String: "", Valid: false},
		ProviderID:         providerID,
		ModerationStatus:   aux.ModerationStatusPending,
	},
	)
}

func (auxInfoRepo *AuxiliaryInformationRepository) UpdateVehicleOID(id int32, OID int32) error {
	return auxInfoRepo.Source.UpdateVehicleOID(id, OID)
}

func (auxInfoRepo *AuxiliaryInformationRepository) GetProviderIDByIP(ip string) (int32, error) {
	provider, err := auxInfoRepo.Source.GetProviderByIP(ip)
	return provider.ID, err
}
