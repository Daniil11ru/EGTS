package source

import (
	"fmt"
	"time"

	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/filter"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/insert"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/in/update"
	"github.com/daniil11ru/egts/cli/receiver/dto/db/out"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DefaultPrimary struct {
	db *gorm.DB
}

func NewDefaultPrimary(dsn string) (*DefaultPrimary, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}

	return &DefaultPrimary{db: db}, nil
}

func (s *DefaultPrimary) GetVehicles(filter filter.Vehicles) ([]out.Vehicle, error) {
	var vehicles []out.Vehicle

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

	if filter.OID != nil {
		q = q.Where("oid = ?", *filter.OID)
	}

	if err := q.Scan(&vehicles).Error; err != nil {
		return nil, err
	}

	return vehicles, nil
}

func (s *DefaultPrimary) AddVehicle(v insert.Vehicle) (int32, error) {
	if v.IMEI == "" || v.ProviderID <= 0 || v.ModerationStatus == "" {
		return 0, fmt.Errorf("IMEI, ID провайдера и статус модерации не могут быть пустыми")
	}

	if v.Name != nil && *v.Name == "" {
		v.Name = nil
	}

	const q = `
		INSERT INTO vehicle (imei, oid, name, provider_id, moderation_status)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id
	`
	var id int32
	if err := s.db.Exec(q, v.IMEI, v.OID, v.Name, v.ProviderID, v.ModerationStatus).Scan(&id).Error; err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DefaultPrimary) GetVehicle(id int32) (out.Vehicle, error) {
	var vehicle out.Vehicle

	q := s.db.Table("vehicle").Select("id, imei, oid, name, provider_id, moderation_status").Where("id = ?", id)

	if err := q.Scan(&vehicle).Error; err != nil {
		return out.Vehicle{}, err
	}

	return vehicle, nil
}

func (s *DefaultPrimary) GetProviders() ([]out.Provider, error) {
	var providers []out.Provider
	if err := s.db.Table("provider").Select("id, name").Scan(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

func (s *DefaultPrimary) GetLocations(filter filter.Locations) ([]out.Location, error) {
	var locations []out.Location

	sub := s.db.Table("vehicle_movement").Select(`
		id,
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
		Select("id, vehicle_id, latitude, longitude, altitude, direction, speed, satellite_count, sent_at, received_at").
		Order("vehicle_id, sent_at DESC")

	if err := q.Scan(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

func (s *DefaultPrimary) GetApiKeys() ([]out.ApiKey, error) {
	var apiKeys []out.ApiKey

	if err := s.db.Table("api_key").Select("id, name, hash").Scan(&apiKeys).Error; err != nil {
		return nil, err
	}

	return apiKeys, nil
}

func (s *DefaultPrimary) UpdateVehicleByImei(imei string, update update.VehicleByImei) error {
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

func (s *DefaultPrimary) UpdateVehicleById(id int32, update update.VehicleById) error {
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
	if update.OID != nil {
		updates["oid"] = *update.OID
	}
	if len(updates) == 0 {
		return nil
	}
	return s.db.Table("vehicle").Where("id = ?", id).Updates(updates).Error
}

func (s *DefaultPrimary) AddLocation(insert insert.Location) (int32, error) {
	const q = `
		INSERT INTO vehicle_movement (vehicle_id, latitude, longitude, altitude, direction, speed, satellite_count, sent_at, received_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`
	var id int32
	if err := s.db.Exec(q, insert.VehicleId, insert.Latitude, insert.Longitude, insert.Altitude, insert.Direction, insert.Speed, insert.SatelliteCount, insert.SentAt, insert.ReceivedAt).Scan(&id).Error; err != nil {
		return 0, err
	}

	return id, nil
}

func (s *DefaultPrimary) GetLastVehiclePoint(id int32) (out.Point, error) {
	var point out.Point

	res := s.db.Table("vehicle_movement").
		Select("id AS location_id, latitude, longitude, altitude").
		Where("vehicle_id = ? AND sent_at IS NOT NULL", id).
		Order("sent_at DESC").
		Limit(1).
		Take(&point)

	if res.Error != nil {
		return out.Point{}, res.Error
	}
	return point, nil
}

func (s *DefaultPrimary) GetTracks(after, before time.Time) ([]out.Track, error) {
	rows, err := s.db.Table("vehicle_movement").
		Select("id, vehicle_id, latitude, longitude, altitude").
		Where("received_at BETWEEN ? AND ? AND latitude IS NOT NULL AND longitude IS NOT NULL", after, before).
		Order("vehicle_id, received_at ASC").
		Rows()
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var tracks []out.Track
	var currentTrack *out.Track

	for rows.Next() {
		var (
			id        int32
			vehicleId int32
			lat       float64
			lon       float64
			alt       *int64
		)
		if err := rows.Scan(&id, &vehicleId, &lat, &lon, &alt); err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %v", err)
		}

		if currentTrack == nil || currentTrack.VehicleId != vehicleId {
			if currentTrack != nil {
				tracks = append(tracks, *currentTrack)
			}
			currentTrack = &out.Track{
				VehicleId: vehicleId,
				Points:    []out.Point{},
			}
		}

		currentTrack.Points = append(currentTrack.Points, out.Point{
			LocationId: id,
			Latitude:   lat,
			Longitude:  lon,
			Altitude:   alt,
		})
	}

	if currentTrack != nil {
		tracks = append(tracks, *currentTrack)
	}

	return tracks, nil
}

func (s *DefaultPrimary) DeleteLocation(id int32) error {
	res := s.db.Exec("DELETE FROM vehicle_movement WHERE id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("ошибка выполнения запроса удаления: %v", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("запись о перемещении транспорта с ID %d не найдена", id)
	}
	return nil
}
