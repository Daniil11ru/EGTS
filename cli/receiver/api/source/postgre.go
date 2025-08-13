package source

import (
	"fmt"

	"github.com/daniil11ru/egts/cli/receiver/api/dto/request"
	"github.com/daniil11ru/egts/cli/receiver/api/model"
	"gorm.io/gorm"
)

type Postgre struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Postgre {
	return &Postgre{db: db}
}

func (s *Postgre) GetVehicles(request request.GetVehicles) ([]model.Vehicle, error) {
	var vehicles []model.Vehicle

	q := s.db.Table("vehicle").Select("id, imei, oid, name, provider_id, moderation_status")

	if request.ProviderID != nil {
		q = q.Where("provider_id = ?", *request.ProviderID)
	}

	if request.ModerationStatus != nil {
		q = q.Where("moderation_status = ?", *request.ModerationStatus)
	}

	if request.IMEI != nil {
		q = q.Where("imei = ?", *request.IMEI)
	}

	if err := q.Scan(&vehicles).Error; err != nil {
		return nil, err
	}

	return vehicles, nil
}

func (s *Postgre) GetVehicle(vehicleId int32) (model.Vehicle, error) {
	var vehicle model.Vehicle

	q := s.db.Table("vehicle").Select("id, imei, oid, name, provider_id, moderation_status").Where("id = ?", vehicleId)

	if err := q.Scan(&vehicle).Error; err != nil {
		return model.Vehicle{}, err
	}

	return vehicle, nil
}

func (s *Postgre) GetLocations(request request.GetLocations) ([]model.Location, error) {
	var locations []model.Location

	sub := s.db.Table("vehicle_movement").Select(`
		vehicle_id,
		latitude,
		longitude,
		altitude,
		direction,
		speed,
		satellite_count,
		sent_at,
		received_at,
		ROW_NUMBER() OVER (PARTITION BY vehicle_id ORDER BY sent_at DESC) AS rn`)

	if request.VehicleID != nil {
		sub = sub.Where("vehicle_id = ?", *request.VehicleID)
	}
	if request.SentBefore != nil {
		sub = sub.Where("sent_at < ?", *request.SentBefore)
	}
	if request.SentAfter != nil {
		sub = sub.Where("sent_at > ?", *request.SentAfter)
	}
	if request.ReceivedBefore != nil {
		sub = sub.Where("received_at < ?", *request.ReceivedBefore)
	}
	if request.ReceivedAfter != nil {
		sub = sub.Where("received_at > ?", *request.ReceivedAfter)
	}

	q := s.db.Table("(?) AS ranked", sub).
		Where("rn <= ?", request.LocationsLimit).
		Select("vehicle_id, latitude, longitude, altitude, direction, speed, satellite_count, sent_at, received_at").
		Order("vehicle_id, sent_at DESC")

	if err := q.Scan(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

func (s *Postgre) GetApiKeys() ([]model.ApiKey, error) {
	var apiKeys []model.ApiKey

	if err := s.db.Table("api_key").Select("id, name, hash").Scan(&apiKeys).Error; err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (s *Postgre) UpdateVehicleByImei(request request.UpdateVehicle) error {
	updates := map[string]interface{}{}
	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.ModerationStatus != nil {
		updates["moderation_status"] = *request.ModerationStatus
	}
	if len(updates) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var ids []int32
		if err := tx.Table("vehicle").
			Select("id").
			Where("imei = ?", request.IMEI).
			Limit(2).
			Find(&ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return gorm.ErrRecordNotFound
		}
		if len(ids) > 1 {
			return fmt.Errorf("по заданному IMEI найдено более одной транспортной единицы")
		}
		res := tx.Table("vehicle").Where("id = ?", ids[0]).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("обновление не удалось")
		}
		return nil
	})
}

func (s *Postgre) UpdateVehicleById(vehicleId int32, request request.UpdateVehicle) error {
	updates := map[string]interface{}{}
	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.ModerationStatus != nil {
		updates["moderation_status"] = *request.ModerationStatus
	}
	if request.IMEI != nil {
		updates["imei"] = *request.IMEI
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.Table("vehicle").Where("id = ?", vehicleId).Updates(updates).Error
}
