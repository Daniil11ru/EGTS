package domain

import (
	"fmt"
	"math/bits"
	"strconv"
	"time"

	"github.com/daniil11ru/egts/cli/receiver/dto/db/out"
	util "github.com/daniil11ru/egts/cli/receiver/dto/other"
	repository "github.com/daniil11ru/egts/cli/receiver/server/repository"
	cron "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type SavePacket struct {
	PrimaryRepository repository.Primary

	AddVehicleMovementMonthStart int
	AddVehicleMovementMonthEnd   int

	vehicleIdToLastPosition map[int32]out.Point

	cronScheduler *cron.Cron
}

func (domain *SavePacket) fillVehicleIdToLastPosition() error {
	domain.vehicleIdToLastPosition = make(map[int32]out.Point)

	vehicles, getAllVehiclesErr := domain.PrimaryRepository.GetAllVehicles()
	if getAllVehiclesErr != nil {
		return fmt.Errorf("не удалось получить список транспорта: %w", getAllVehiclesErr)
	}

	for i := 0; i < len(vehicles); i++ {
		lastPosition, getLastPositionErr := domain.PrimaryRepository.GetLastVehiclePosition(vehicles[i].ID)
		if getLastPositionErr == nil {
			domain.vehicleIdToLastPosition[vehicles[i].ID] = lastPosition
		}
	}

	return nil
}

func (domain *SavePacket) Initialize() error {
	if err := domain.fillVehicleIdToLastPosition(); err != nil {
		return fmt.Errorf("не удалось инициализировать кэш транспорта: %w", err)
	}

	loc, loadLocationErr := time.LoadLocation("Europe/Moscow")
	if loadLocationErr != nil {
		return fmt.Errorf("не удалось загрузить временную зону Europe/Moscow: %w", loadLocationErr)
	}
	domain.cronScheduler = cron.New(cron.WithLocation(loc))

	_, err := domain.cronScheduler.AddFunc("0 3 * * *", func() {
		logrus.Info("Запуск запланированного обновления кэша транспорта")
		if err := domain.fillVehicleIdToLastPosition(); err != nil {
			logrus.Errorf("Ошибка обновления кэша транспорта: %v", err)
		} else {
			logrus.Info("Кэш транспорта успешно обновлен")
		}
	})

	if err != nil {
		return fmt.Errorf("ошибка при настройке cron-задачи: %w", err)
	}

	domain.cronScheduler.Start()
	logrus.Info("Запланировано ежедневное обновление кэша провайдеров в 03:00")

	return nil
}

func (domain *SavePacket) Shutdown() {
	if domain.cronScheduler != nil {
		domain.cronScheduler.Stop()
		logrus.Info("Cron-планировщик остановлен")
	}
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

func (domain *SavePacket) filterVehiclesByOID(OID int64, vehicles []out.Vehicle) ([]out.Vehicle, error) {
	var result []out.Vehicle

	for _, v := range vehicles {
		imeiStr := v.IMEI
		imei, err := strconv.ParseUint(imeiStr, 10, 64)
		if err != nil {
			logrus.Warnf("Не удалось преобразовать IMEI '%s' в число: %v", imeiStr, err)
			continue
		}
		if isPartOf(uint64(OID), imei) {
			result = append(result, v)
		}
	}

	return result, nil
}

func (s *SavePacket) findVehicles(OID int64, providerID int32) ([]out.Vehicle, error) {
	vehicles, err := s.PrimaryRepository.GetVehiclesByOIDAndProviderID(OID, providerID)
	if err == nil {
		return vehicles, nil
	}

	vehicles, err = s.PrimaryRepository.GetVehiclesByProviderID(providerID)
	if err != nil {
		return []out.Vehicle{}, err
	}
	vehicles, err = s.filterVehiclesByOID(OID, vehicles)
	if err != nil {
		return []out.Vehicle{}, err
	}

	return vehicles, nil
}

func (s *SavePacket) resolveModerationStatus(id int32) (util.ModerationStatus, error) {
	moderationStatus, err := s.PrimaryRepository.GetVehicleModerationStatus(id)
	return moderationStatus, err
}

func (s *SavePacket) Run(data *util.PacketData, providerID int32) error {
	if data.Latitude == 0 || data.Longitude == 0 || data.OID == 0 {
		logrus.Debugf("OID: %d, широта: %f, долгота: %f", data.OID, data.Latitude, data.Longitude)
		return fmt.Errorf("широта, долгота и OID не должны быть пустыми или иметь нулевое значение")
	}

	oid := data.OID

	month := int(time.Now().UTC().Month())
	if month < s.AddVehicleMovementMonthStart || month > s.AddVehicleMovementMonthEnd {
		logrus.Debug("Запись телематических данных в текущий месяц запрещена")
		return nil
	}

	var vehicleID int32
	vehicles, err := s.findVehicles(int64(oid), providerID)
	if err != nil {
		return fmt.Errorf("не удалось найти транспорт по OID %d: %w", oid, err)
	} else if len(vehicles) == 0 {
		var addIndefiniteVehicleErr error
		vehicleID, addIndefiniteVehicleErr = s.PrimaryRepository.AddIndefiniteVehicle(int64(oid), providerID)
		if addIndefiniteVehicleErr != nil {
			return fmt.Errorf("не удалось добавить новый транспорт: %w", addIndefiniteVehicleErr)
		}
		logrus.Warnf("Не удалось найти транспорт по OID %d, был добавлен новый транспорт с ID %d", oid, vehicleID)
	} else if len(vehicles) > 1 {
		return fmt.Errorf("не удалось однозначно определить транспорт по OID %d", oid)
	} else if len(vehicles) == 1 {
		vehicleID = vehicles[0].ID

		// FIXME: нужно обновлять только тогда, когда OID действительно отсутствует
		s.PrimaryRepository.UpdateVehicleOID(vehicleID, int64(oid))
	}

	moderationStatus, err := s.resolveModerationStatus(vehicleID)
	if err != nil {
		return fmt.Errorf("не удалось определить статус модерации транспорта с ID %d: %w", vehicleID, err)
	}
	if moderationStatus == util.ModerationStatusRejected {
		logrus.Debugf("Запись телематических данных для транспорта с ID %d запрещена", vehicleID)
		return nil
	}

	altitude := int64(data.Altitude)
	currentPosition := out.Point{Latitude: data.Latitude, Longitude: data.Longitude, Altitude: &altitude}
	lastPosition, OK := s.vehicleIdToLastPosition[vehicleID]
	if OK {
		accuracy_meters := 10.0

		equals := false
		if *currentPosition.Altitude == 0 || *lastPosition.Altitude == 0 {
			equals = lastPosition.EqualsHorizontallyTo(&currentPosition, accuracy_meters)
		} else {
			equals = lastPosition.EqualsTo(&currentPosition, accuracy_meters)
		}

		if equals {
			logrus.Debugf("Новое местоположение транспорта с ID %d не отличается от предыдущего", vehicleID)
			return nil
		}
	}
	s.vehicleIdToLastPosition[vehicleID] = currentPosition

	if _, err := s.PrimaryRepository.AddLocation(data, vehicleID); err != nil {
		return fmt.Errorf("не удалось сохранить телематические данные для транспорта с ID %d: %w", vehicleID, err)
	}

	return nil
}
