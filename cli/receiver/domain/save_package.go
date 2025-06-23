package domain

import (
	"fmt"
	"math/bits"
	"time"

	repository "github.com/daniil11ru/egts/cli/receiver/repository/primary"
	util "github.com/daniil11ru/egts/cli/receiver/repository/util"
	source "github.com/daniil11ru/egts/cli/receiver/source/primary"
)

type SavePackage struct {
	PrimaryRepository repository.PrimaryRepository

	AddVehicleMovementMonthStart int
	AddVehicleMovementMonthEnd   int
}

func isPrefixBytes(a, b uint64, n int) bool {
	width := (bits.Len64(b) + 7) / 8
	shift := uint((width - n) * 8)
	return a == b>>shift
}

func isSuffixBytes(a, b uint64, n int) bool {
	width := (bits.Len64(b) + 7) / 8
	if n >= width {
		return a == b
	}
	return a == b&((1<<(uint(n)*8))-1)
}

func isPrefixDigits(a, b uint64, n int) bool {
	digits := 1
	for tmp := b; tmp >= 10; tmp /= 10 {
		digits++
	}
	div := uint64(1)
	for i := 0; i < digits-n; i++ {
		div *= 10
	}
	return a == b/div
}

func isSuffixDigits(a, b uint64, n int) bool {
	mod := uint64(1)
	for i := 0; i < n; i++ {
		mod *= 10
	}
	return a == b%mod
}

func digitCount(n uint64) int {
	if n == 0 {
		return 1
	}
	count := 0
	for n > 0 {
		count++
		n /= 10
	}
	return count
}

func byteCount(n uint64) int {
	if n == 0 {
		return 1
	}
	return (bits.Len64(n) + 7) / 8
}

func isPartOf(a, b uint64) bool {
	byteCount := byteCount(a)
	digitCount := digitCount(a)

	if isSuffixBytes(a, b, byteCount) || isPrefixBytes(a, b, byteCount) ||
		isSuffixDigits(a, b, digitCount) || isPrefixDigits(a, b, digitCount) {
		return true
	}

	return false
}

func (domain *SavePackage) getVehicleIDByOID(OID int32, vehicles []source.Vehicle) (int32, error) {
	isFound := false
	var id int32 = 0

	for _, v := range vehicles {
		IMEI := v.IMEI

		if isPartOf(uint64(OID), uint64(IMEI)) {
			if !isFound {
				id = v.ID
			} else {
				id = 0
				return id, fmt.Errorf("не удалось однозначно определить IMEI")
			}
		}
	}

	return id, fmt.Errorf("не удалось определить IMEI")
}

func (domain *SavePackage) getVehicleIDByOIDAndProviderIDFromStorage(OID int32, providerID int32) (int32, error) {
	vehicle, err := domain.PrimaryRepository.GetVehicleByOIDAndProviderID(OID, providerID)
	return vehicle.ID, err
}

func (s *SavePackage) resolveVehicleID(OID int32, providerIP string) (int32, error) {
	providerID, getProviderIDError := s.PrimaryRepository.GetProviderIDByIP(providerIP)
	if getProviderIDError != nil {
		return providerID, getProviderIDError
	}

	id, err := s.getVehicleIDByOIDAndProviderIDFromStorage(OID, providerID)
	if err == nil {
		return id, nil
	}

	vehicles, auxErr := s.PrimaryRepository.GetVehiclesByProviderIP(providerIP)
	if auxErr != nil {
		return 0, auxErr
	}

	id, err = s.getVehicleIDByOID(OID, vehicles)
	if err != nil {
		id, err = s.PrimaryRepository.AddIndefiniteVehicle(OID, providerID)
		return id, err
	}

	s.PrimaryRepository.UpdateVehicleOID(id, OID)
	return id, nil
}

func (s *SavePackage) resolveModerationStatus(id int32) (source.ModerationStatus, error) {
	moderationStatus, err := s.PrimaryRepository.GetVehicleModerationStatus(id)
	return moderationStatus, err
}

func (s *SavePackage) Run(data *util.NavRecord, providerIP string) error {
	oid := int32(data.Client)

	month := int(time.Now().UTC().Month())
	if month < s.AddVehicleMovementMonthStart || month > s.AddVehicleMovementMonthEnd {
		return nil
	}

	vehicleID, err := s.resolveVehicleID(oid, providerIP)
	if err != nil {
		return fmt.Errorf("не удалось найти транспорт по OID %d: %w", oid, err)
	}

	moderationStatus, err := s.resolveModerationStatus(vehicleID)
	if err != nil {
		return fmt.Errorf("не удалось определить статус модерации транспорта с ID %d: %w", vehicleID, err)
	}
	if moderationStatus == source.ModerationStatusRejected {
		return fmt.Errorf("запись телематических данных для транспорта с ID %d запрещена", vehicleID)
	}

	if _, err := s.PrimaryRepository.AddVehicleMovement(data, int(vehicleID)); err != nil {
		return fmt.Errorf("не удалось сохранить телематические данные для транспорта с ID %d: %w", vehicleID, err)
	}

	return nil
}
