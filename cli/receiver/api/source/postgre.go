package source

import (
	"fmt"

	"github.com/daniil11ru/egts/cli/receiver/api/dto/db/in/filter"
	"github.com/daniil11ru/egts/cli/receiver/api/dto/db/in/update"
	output "github.com/daniil11ru/egts/cli/receiver/api/dto/db/out"
	"gorm.io/gorm"
)

type Postgre struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Postgre {
	return &Postgre{db: db}
}

func (s *Postgre) GetVehicles(filter filter.Vehicles) ([]output.Vehicle, error) {
	var vehicles []output.Vehicle

	q := s.db.Table("vehicle").Select("id, imei, oid, name, provider_id, moderation_status")

	if filter.ProviderID != nil {
		q = q.Where("provider_id = ?", *filter.ProviderID)
	}

	if filter.ModerationStatus != nil {
		q = q.Where("moderation_status = ?", *filter.ModerationStatus)
	}

	if filter.IMEI != nil {
		q = q.Where("imei = ?", *filter.IMEI)
	}

	if err := q.Scan(&vehicles).Error; err != nil {
		return nil, err
	}

	return vehicles, nil
}

func (s *Postgre) GetVehicle(vehicleId int32) (output.Vehicle, error) {
	var vehicle output.Vehicle

	q := s.db.Table("vehicle").Select("id, imei, oid, name, provider_id, moderation_status").Where("id = ?", vehicleId)

	if err := q.Scan(&vehicle).Error; err != nil {
		return output.Vehicle{}, err
	}

	return vehicle, nil
}

func (s *Postgre) GetLocations(filter filter.Locations) ([]output.Location, error) {
	var locations []output.Location

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

	if filter.VehicleID != nil {
		sub = sub.Where("vehicle_id = ?", *filter.VehicleID)
	}
	if filter.SentBefore != nil {
		sub = sub.Where("sent_at < ?", *filter.SentBefore)
	}
	if filter.SentAfter != nil {
		sub = sub.Where("sent_at > ?", *filter.SentAfter)
	}
	if filter.ReceivedBefore != nil {
		sub = sub.Where("received_at < ?", *filter.ReceivedBefore)
	}
	if filter.ReceivedAfter != nil {
		sub = sub.Where("received_at > ?", *filter.ReceivedAfter)
	}

	q := s.db.Table("(?) AS ranked", sub).
		Where("rn <= ?", filter.LocationsLimit).
		Select("vehicle_id, latitude, longitude, altitude, direction, speed, satellite_count, sent_at, received_at").
		Order("vehicle_id, sent_at DESC")

	if err := q.Scan(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

func (s *Postgre) GetApiKeys() ([]output.ApiKey, error) {
	var apiKeys []output.ApiKey

	if err := s.db.Table("api_key").Select("id, name, hash").Scan(&apiKeys).Error; err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (s *Postgre) UpdateVehicleByImei(imei string, update update.VehicleByImei) error {
	updates := map[string]interface{}{}
	if update.Name != nil {
		updates["name"] = *update.Name
	}
	if update.ModerationStatus != nil {
		updates["moderation_status"] = *update.ModerationStatus
	}
	if len(updates) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var ids []int32
		if err := tx.Table("vehicle").
			Select("id").
			Where("imei = ?", imei).
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

func (s *Postgre) UpdateVehicleById(vehicleId int32, update update.VehicleById) error {
	updates := map[string]interface{}{}
	if update.Name != nil {
		updates["name"] = *update.Name
	}
	if update.ModerationStatus != nil {
		updates["moderation_status"] = *update.ModerationStatus
	}
	if update.IMEI != nil {
		updates["imei"] = *update.IMEI
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.Table("vehicle").Where("id = ?", vehicleId).Updates(updates).Error
}
