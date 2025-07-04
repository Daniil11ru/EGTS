package domain

import (
	"fmt"
	"math/bits"
	"time"

	repository "github.com/daniil11ru/egts/cli/receiver/repository/primary"
	"github.com/daniil11ru/egts/cli/receiver/repository/primary/types"
	util "github.com/daniil11ru/egts/cli/receiver/repository/util"
	"github.com/sirupsen/logrus"
)

type SavePacket struct {
	PrimaryRepository repository.PrimaryRepository

	AddVehicleMovementMonthStart int
	AddVehicleMovementMonthEnd   int

	vehicleIDToLastPosition map[int32]types.Position
}

func (domain *SavePacket) Initialize() error {
	domain.vehicleIDToLastPosition = make(map[int32]types.Position)

	vehicles, getAllVehiclesErr := domain.PrimaryRepository.GetAllVehicles()
	if getAllVehiclesErr != nil {
		return fmt.Errorf("не удалось инициализировать кэш: %w", getAllVehiclesErr)
	}

	for i := 0; i < len(vehicles); i++ {
		lastPosition, getLastPositionErr := domain.PrimaryRepository.GetLastVehiclePosition(vehicles[i].ID)
		if getLastPositionErr == nil {
			domain.vehicleIDToLastPosition[vehicles[i].ID] = lastPosition
		}
	}

	return nil
}

func isPrefixBytes(a, b uint64, n int) bool {
	width := (bits.Len64(b) + 7) / 8
	if n >= width {
		return a == b
	}
	shift := uint((width - n) * 8)
	mask := uint64(1)<<(uint(n)*8) - 1
	return a == (b>>shift)&mask
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

func (domain *SavePacket) filterVehiclesByOID(OID uint32, vehicles []types.Vehicle) ([]types.Vehicle, error) {
	var result []types.Vehicle

	for _, v := range vehicles {
		IMEI := v.IMEI

		if isPartOf(uint64(OID), uint64(IMEI)) {
			result = append(result, v)
		}
	}

	return result, nil
}

func (s *SavePacket) findVehicles(OID uint32, providerIP string) ([]types.Vehicle, error) {
	providerID, getProviderIDError := s.PrimaryRepository.GetProviderIDByIP(providerIP)
	if getProviderIDError != nil {
		return []types.Vehicle{}, getProviderIDError
	}

	vehicles, getVehiclesByOIDAndProviderIDError := s.PrimaryRepository.GetVehiclesByOIDAndProviderID(OID, providerID)
	if getVehiclesByOIDAndProviderIDError == nil {
		return vehicles, nil
	}

	vehicles, getVehiclesByProviderIPError := s.PrimaryRepository.GetVehiclesByProviderIP(providerIP)
	if getVehiclesByProviderIPError != nil {
		return []types.Vehicle{}, getVehiclesByProviderIPError
	}
	vehicles, filterVehiclesByOIDError := s.filterVehiclesByOID(OID, vehicles)
	if filterVehiclesByOIDError != nil {
		return []types.Vehicle{}, filterVehiclesByOIDError
	}

	return vehicles, nil
}

func (s *SavePacket) resolveModerationStatus(id int32) (types.ModerationStatus, error) {
	moderationStatus, err := s.PrimaryRepository.GetVehicleModerationStatus(id)
	return moderationStatus, err
}

func (s *SavePacket) Run(data *util.NavigationRecord, providerIP string) error {
	if data.SatelliteCount == 0 || data.Latitude == 0 || data.Longitude == 0 || data.OID == 0 {
		return fmt.Errorf("широта, долгота, OID и количество спутников не должны быть пустыми или иметь нулевое значение")
	}

	providerID, getProviderIDError := s.PrimaryRepository.GetProviderIDByIP(providerIP)
	if getProviderIDError != nil {
		return fmt.Errorf("не удалось определить провайдера по IP %s: %w", providerIP, getProviderIDError)
	}

	oid := data.OID

	month := int(time.Now().UTC().Month())
	if month < s.AddVehicleMovementMonthStart || month > s.AddVehicleMovementMonthEnd {
		logrus.Debug("Запись телематических данных в текущий месяц запрещена")
		return nil
	}

	var vehicleID int32
	vehicles, err := s.findVehicles(oid, providerIP)
	if err != nil {
		return fmt.Errorf("не удалось найти транспорт по OID %d: %w", oid, err)
	} else if len(vehicles) == 0 {
		var addIndefiniteVehicleErr error
		vehicleID, addIndefiniteVehicleErr = s.PrimaryRepository.AddIndefiniteVehicle(oid, providerID)
		if addIndefiniteVehicleErr != nil {
			return fmt.Errorf("не удалось добавить новый транспорт: %w", addIndefiniteVehicleErr)
		}
		logrus.Warnf("Не удалось найти транспорт по OID %d, был добавлен новый транспорт с ID %d", oid, vehicleID)
	} else if len(vehicles) > 1 {
		return fmt.Errorf("не удалось однозначно определить транспорт по OID %d", oid)
	} else if len(vehicles) == 1 {
		vehicleID = vehicles[0].ID

		// FIXME: нужно обновлять только тогда, когда OID действительно отсутствует
		s.PrimaryRepository.UpdateVehicleOID(vehicleID, oid)
	}

	moderationStatus, err := s.resolveModerationStatus(vehicleID)
	if err != nil {
		return fmt.Errorf("не удалось определить статус модерации транспорта с ID %d: %w", vehicleID, err)
	}
	if moderationStatus == types.ModerationStatusRejected {
		logrus.Debugf("Запись телематических данных для транспорта с ID %d запрещена", vehicleID)
		return nil
	}

	currentPosition := types.Position{Latitude: data.Latitude, Longitude: data.Longitude, Altitude: data.Altitude}
	lastPosition, OK := s.vehicleIDToLastPosition[vehicleID]
	if OK {
		accuracy_meters := 10.0

		equals := false
		if currentPosition.Altitude == 0 || lastPosition.Altitude == 0 {
			equals = lastPosition.EqualsHorizontallyTo(&currentPosition, accuracy_meters)
		} else {
			equals = lastPosition.EqualsTo(&currentPosition, accuracy_meters)
		}

		if equals {
			logrus.Debugf("Новое местоположение транспорта с ID %d не отличается от предыдущего", vehicleID)
			return nil
		}
	}
	s.vehicleIDToLastPosition[vehicleID] = currentPosition

	if _, err := s.PrimaryRepository.AddVehicleMovement(data, int(vehicleID)); err != nil {
		return fmt.Errorf("не удалось сохранить телематические данные для транспорта с ID %d: %w", vehicleID, err)
	}

	return nil
}
