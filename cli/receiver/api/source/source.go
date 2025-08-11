package source

import (
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

func (s *Postgre) GetLocations(request request.GetLocations) ([]model.Location, error) {
	var locations []model.Location

	q := s.db.Table("vehicle_movement").Select("vehicle_id, latitude, longitude, sent_at, received_at")

	if request.VehicleID != nil {
		q = q.Where("vehicle_id = ?", *request.VehicleID)
	}
	if request.SentBefore != nil {
		q = q.Where("sent_at < ?", *request.SentBefore)
	}
	if request.SentAfter != nil {
		q = q.Where("sent_at > ?", *request.SentAfter)
	}
	if request.ReceivedBefore != nil {
		q = q.Where("received_at < ?", *request.ReceivedBefore)
	}
	if request.ReceivedAfter != nil {
		q = q.Where("received_at > ?", *request.ReceivedAfter)
	}

	if err := q.Scan(&locations).Error; err != nil {
		return nil, err
	}

	return locations, nil
}

func (s *Postgre) GetLatestLocations(request request.GetLatestLocations) ([]model.LatestLocation, error) {
	var latestLocations []model.LatestLocation

	q := s.db.Table("vehicle_movement").Select("vehicle_id, latitude, longitude, altitude, direction, speed, satellite_count, sent_at, received_at")

	if request.VehicleID != nil {
		q = q.Where("vehicle_id = ?", *request.VehicleID)
	}

	if err := q.Scan(&latestLocations).Error; err != nil {
		return nil, err
	}

	return latestLocations, nil
}

func (s *Postgre) GetApiKeys() ([]model.ApiKey, error) {
	var apiKeys []model.ApiKey

	if err := s.db.Table("api_key").Select("id, name, hash").Scan(&apiKeys).Error; err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (s *Postgre) UpdateVehicle(request request.UpdateVehicle) error {
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
	return s.db.Table("vehicle").Where("imei = ?", request.IMEI).Updates(updates).Error
}
