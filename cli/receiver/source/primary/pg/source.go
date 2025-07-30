package pg

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	connector "github.com/daniil11ru/egts/cli/receiver/connector"
	"github.com/daniil11ru/egts/cli/receiver/repository/primary/types"
	"github.com/daniil11ru/egts/cli/receiver/repository/util"
)

type PrimarySource struct {
	connector connector.Connector
}

func (p *PrimarySource) Initialize(c connector.Connector) {
	p.connector = c
}

func (p *PrimarySource) db() (*sql.DB, error) {
	if p.connector == nil {
		return nil, fmt.Errorf("не удалось инициализировать подключение к базе данных")
	}
	db := p.connector.GetConnection()
	if db == nil {
		return nil, fmt.Errorf("нет активного подключения к базе данных")
	}
	return db, nil
}

func (p *PrimarySource) GetAllVehicles() ([]types.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}

	const q = "SELECT id, imei, oid, name, provider_id, moderation_status FROM vehicle"
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []types.Vehicle
	for rows.Next() {
		var v types.Vehicle
		if err := rows.Scan(&v.ID, &v.IMEI, &v.OID, &v.Name, &v.ProviderID, &v.ModerationStatus); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (p *PrimarySource) GetAllProviders() ([]types.Provider, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}

	const q = `SELECT id, name, ip FROM provider`
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []types.Provider
	for rows.Next() {
		var pr types.Provider
		var ip sql.NullString

		if err := rows.Scan(&pr.ID, &pr.Name, &ip); err != nil {
			return nil, err
		}

		if ip.Valid {
			pr.IP = ip.String
		} else {
			pr.IP = ""
		}

		providers = append(providers, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return providers, nil
}

func (p *PrimarySource) GetVehiclesByProviderIP(ip string) ([]types.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}

	const q = `
        SELECT v.id,
               v.imei,
               v.name,
               v.provider_id,
               v.moderation_status
        FROM vehicle v
        JOIN provider p ON p.id = v.provider_id
        WHERE p.ip = $1
    `

	rows, err := db.Query(q, ip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []types.Vehicle
	for rows.Next() {
		var v types.Vehicle
		if err := rows.Scan(
			&v.ID,
			&v.IMEI,
			&v.Name,
			&v.ProviderID,
			&v.ModerationStatus,
		); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return vehicles, nil
}

func (p *PrimarySource) GetVehicleByID(id int32) (types.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return types.Vehicle{}, err
	}

	const q = `
		SELECT id, imei, name, provider_id, moderation_status
		FROM vehicle
		WHERE id = $1
	`
	row := db.QueryRow(q, id)

	var v types.Vehicle
	if err := row.Scan(
		&v.ID,
		&v.IMEI,
		&v.Name,
		&v.ProviderID,
		&v.ModerationStatus,
	); err != nil {
		if err == sql.ErrNoRows {
			return types.Vehicle{}, fmt.Errorf("транспорт с ID %d не найден", id)
		}
		return types.Vehicle{}, err
	}

	return v, nil
}

func (p *PrimarySource) GetVehicleByOID(oid uint32) (types.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return types.Vehicle{}, err
	}

	const q = `
        SELECT id, imei, name, provider_id, moderation_status
        FROM vehicle
        WHERE oid = $1
    `
	rows, err := db.Query(q, oid)
	if err != nil {
		return types.Vehicle{}, err
	}
	defer rows.Close()

	var (
		v     types.Vehicle
		count int
	)
	for rows.Next() {
		var tmp types.Vehicle
		if err := rows.Scan(&tmp.ID, &tmp.IMEI, &tmp.Name, &tmp.ProviderID, &tmp.ModerationStatus); err != nil {
			return types.Vehicle{}, err
		}
		if count == 0 {
			v = tmp
		}
		count++
		if count > 1 {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return types.Vehicle{}, err
	}

	switch count {
	case 0:
		return types.Vehicle{}, fmt.Errorf("транспорт с OID %d не найден", oid)
	case 1:
		return v, nil
	default:
		return types.Vehicle{}, fmt.Errorf("найдено множество транспортных единиц с OID %d", oid)
	}
}

func (p *PrimarySource) GetVehiclesByOIDAndProviderID(oid uint32, providerID int32) ([]types.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return []types.Vehicle{}, err
	}

	const q = `
		SELECT id, imei, name, provider_id, moderation_status
		FROM vehicle
		WHERE oid = $1 AND provider_id = $2
	`
	rows, err := db.Query(q, oid, providerID)
	if err != nil {
		return []types.Vehicle{}, err
	}
	defer rows.Close()

	var vehicles []types.Vehicle
	for rows.Next() {
		var v types.Vehicle
		if err := rows.Scan(&v.ID, &v.IMEI, &v.Name, &v.ProviderID, &v.ModerationStatus); err != nil {
			return []types.Vehicle{}, err
		}
		vehicles = append(vehicles, v)
	}
	if err := rows.Err(); err != nil {
		return []types.Vehicle{}, err
	}

	if len(vehicles) > 0 {
		return vehicles, nil
	} else {
		return []types.Vehicle{}, fmt.Errorf("транспорт с OID %d и ID провайдера %d не найден", oid, providerID)
	}
}

func (p *PrimarySource) AddVehicle(v types.Vehicle) (int32, error) {
	db, err := p.db()
	if err != nil {
		return 0, err
	}

	const q = `
        INSERT INTO vehicle (imei, name, provider_id, moderation_status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	var id int32
	if err := db.QueryRow(q, v.IMEI, v.Name, v.ProviderID, v.ModerationStatus).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (p *PrimarySource) UpdateVehicleOID(id int32, oid uint32) error {
	db, err := p.db()
	if err != nil {
		return err
	}

	const q = `
		UPDATE vehicle
		SET oid = $1
		WHERE id = $2
	`
	res, err := db.Exec(q, oid, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("транспорт с ID %d не найден", id)
	}

	return nil
}

func (p *PrimarySource) GetProviderByIP(ip string) (types.Provider, error) {
	db, err := p.db()
	if err != nil {
		return types.Provider{}, err
	}

	const exactQuery = `
        SELECT id, name, ip
        FROM provider
        WHERE ip = $1
    `
	row := db.QueryRow(exactQuery, ip)

	var exactProvider types.Provider
	var exactNullableIP sql.NullString

	err = row.Scan(&exactProvider.ID, &exactProvider.Name, &exactNullableIP)
	if err == nil {
		if exactNullableIP.Valid {
			exactProvider.IP = exactNullableIP.String
		}
		return exactProvider, nil
	} else if err != sql.ErrNoRows {
		return types.Provider{}, err
	}

	const maskQuery = `
        SELECT id, name, ip
        FROM provider
        WHERE ip LIKE '%*%'
    `
	rows, err := db.Query(maskQuery)
	if err != nil {
		return types.Provider{}, err
	}
	defer rows.Close()

	var matchingProviders []types.Provider
	for rows.Next() {
		var pr types.Provider
		var nullableIP sql.NullString
		if err := rows.Scan(&pr.ID, &pr.Name, &nullableIP); err != nil {
			return types.Provider{}, err
		}

		if nullableIP.Valid {
			pr.IP = nullableIP.String
		} else {
			continue
		}

		if ipMatchesMask(ip, pr.IP) {
			matchingProviders = append(matchingProviders, pr)
		}
	}

	if err := rows.Err(); err != nil {
		return types.Provider{}, err
	}

	switch len(matchingProviders) {
	case 0:
		return types.Provider{}, fmt.Errorf("провайдер с IP %s не найден", ip)
	case 1:
		return matchingProviders[0], nil
	default:
		ids := make([]int32, 0, len(matchingProviders))
		for _, p := range matchingProviders {
			ids = append(ids, p.ID)
		}
		return types.Provider{}, fmt.Errorf(
			"найдено несколько провайдеров для IP %s: %v",
			ip, ids,
		)
	}
}

func ipMatchesMask(ip, mask string) bool {
	if !strings.Contains(mask, "*") {
		return ip == mask
	}

	ipParts := strings.Split(ip, ".")
	maskParts := strings.Split(mask, ".")

	if len(ipParts) != 4 || len(maskParts) > 4 {
		return false
	}

	for i, maskPart := range maskParts {
		if i >= len(ipParts) {
			return false
		}

		if maskPart == "*" {
			continue
		}

		if maskPart != ipParts[i] {
			return false
		}
	}

	return true
}

func (p *PrimarySource) GetAllIPs() ([]string, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}

	const q = `
        SELECT DISTINCT ip
        FROM provider
        WHERE ip IS NOT NULL AND ip <> ''
    `

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ips, nil
}

func (p *PrimarySource) AddVehicleMovement(data *util.NavigationRecord, vehicleID int) (int32, error) {
	if data == nil {
		return 0, fmt.Errorf("некорректная ссылка на пакет")
	}

	var sentTime time.Time
	if data.SentTimestamp == 0 {
		sentTime = time.Time{}
	} else {
		sentTime = time.Unix(data.SentTimestamp, 0)
	}
	receivedTime := time.Unix(data.ReceivedTimestamp, 0)

	const insertQuery = `
        INSERT INTO vehicle_movement (
            vehicle_id,
            latitude,
            longitude,
            altitude,
            direction,
            speed,
            satellite_count,
            sent_at,
            received_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id
    `

	var id int32
	err := p.connector.GetConnection().QueryRow(
		insertQuery,
		vehicleID,
		data.Latitude,
		data.Longitude,
		data.Altitude,
		data.Direction,
		data.Speed,
		data.SatelliteCount,
		sentTime,
		receivedTime,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("не удалось вставить запись: %v", err)
	}

	return id, nil
}

func (p *PrimarySource) GetLastVehiclePosition(vehicleID int32) (types.Position3D, error) {
	db, err := p.db()
	if err != nil {
		return types.Position3D{}, err
	}

	const q = `
    SELECT
        latitude,
        longitude,
        altitude
    FROM vehicle_movement
    WHERE vehicle_id = $1
        AND sent_at IS NOT NULL
        AND latitude IS NOT NULL
        AND longitude IS NOT NULL
    ORDER BY sent_at DESC
    LIMIT 1
    `

	var pos types.Position3D
	var altitude sql.NullInt16

	err = db.QueryRow(q, vehicleID).Scan(&pos.Latitude, &pos.Longitude, &altitude)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.Position3D{}, fmt.Errorf("нет записей местоположения для транспорта с ID %d", vehicleID)
		}
		return types.Position3D{}, err
	}

	if altitude.Valid {
		pos.Altitude = uint32(altitude.Int16)
	}

	return pos, nil
}

func (p *PrimarySource) GetTracks2DOfAllVehicles(after, before time.Time) ([]types.Track2D, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}

	const q = `
        SELECT 
			id,
            vehicle_id,
            latitude,
            longitude
        FROM vehicle_movement
        WHERE 
            received_at BETWEEN $1 AND $2
            AND latitude IS NOT NULL
            AND longitude IS NOT NULL
        ORDER BY vehicle_id, received_at ASC
    `

	rows, err := db.Query(q, after, before)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer rows.Close()

	var tracks []types.Track2D
	var currentTrack *types.Track2D = nil

	for rows.Next() {
		var (
			id        int32
			vehicleID int32
			lat       float64
			lon       float64
		)

		if err := rows.Scan(&id, &vehicleID, &lat, &lon); err != nil {
			return nil, fmt.Errorf("ошибка чтения строки: %v", err)
		}

		if currentTrack == nil || currentTrack.VehicleID != vehicleID {
			if currentTrack != nil {
				tracks = append(tracks, *currentTrack)
			}

			currentTrack = &types.Track2D{
				VehicleID: vehicleID,
				Movements: []types.Movement2D{},
			}
		}

		currentTrack.Movements = append(currentTrack.Movements, types.Movement2D{
			ID: id,
			Position2D: types.Position2D{
				Latitude:  lat,
				Longitude: lon,
			},
		})
	}

	if currentTrack != nil {
		tracks = append(tracks, *currentTrack)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка обработки результатов: %v", err)
	}

	return tracks, nil
}

func (p *PrimarySource) DeleteVehicleMovement(vehicleMovementID int32) error {
	db, err := p.db()
	if err != nil {
		return err
	}

	const q = "DELETE FROM vehicle_movement WHERE id = $1"

	result, err := db.Exec(q, vehicleMovementID)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса удаления: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества удаленных строк: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("запись о перемещении транспорта с ID %d не найдена", vehicleMovementID)
	}

	return nil
}
