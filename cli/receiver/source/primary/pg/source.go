package pg

import (
	"database/sql"
	"fmt"
	"strconv"

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
               v.oid,
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
			&v.OID,
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
		SELECT id, imei, oid, name, provider_id, moderation_status
		FROM vehicle
		WHERE id = $1
	`
	row := db.QueryRow(q, id)

	var v types.Vehicle
	if err := row.Scan(
		&v.ID,
		&v.IMEI,
		&v.OID,
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
        SELECT id, imei, oid, name, provider_id, moderation_status
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
		if err := rows.Scan(&tmp.ID, &tmp.IMEI, &tmp.OID, &tmp.Name, &tmp.ProviderID, &tmp.ModerationStatus); err != nil {
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

func (p *PrimarySource) GetVehicleByOIDAndProviderID(oid uint32, providerID int32) (types.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return types.Vehicle{}, err
	}

	const q = `
		SELECT id, imei, oid, name, provider_id, moderation_status
		FROM vehicle
		WHERE oid = $1 AND provider_id = $2
	`
	rows, err := db.Query(q, oid, providerID)
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
		if err := rows.Scan(&tmp.ID, &tmp.IMEI, &tmp.OID, &tmp.Name, &tmp.ProviderID, &tmp.ModerationStatus); err != nil {
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
		return types.Vehicle{}, fmt.Errorf("транспорт с OID %d и провайдером %d не найден", oid, providerID)
	case 1:
		return v, nil
	default:
		return types.Vehicle{}, fmt.Errorf("найдено несколько транспортных единиц с OID %d и провайдером %d", oid, providerID)
	}
}

func (p *PrimarySource) AddVehicle(v types.Vehicle) (int32, error) {
	db, err := p.db()
	if err != nil {
		return 0, err
	}

	const q = `
        INSERT INTO vehicle (imei, oid, name, provider_id, moderation_status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	var id int32
	if err := db.QueryRow(q, v.IMEI, v.OID, v.Name, v.ProviderID, v.ModerationStatus).Scan(&id); err != nil {
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

	const q = `
        SELECT id, name, ip
        FROM provider
        WHERE ip = $1
    `

	row := db.QueryRow(q, ip)
	var pr types.Provider
	var nullableIP sql.NullString

	if err := row.Scan(&pr.ID, &pr.Name, &nullableIP); err != nil {
		if err == sql.ErrNoRows {
			return types.Provider{}, fmt.Errorf("провайдер с IP %s не найден", ip)
		}
		return types.Provider{}, err
	}

	if nullableIP.Valid {
		pr.IP = nullableIP.String
	} else {
		pr.IP = ""
	}

	return pr, nil
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

	const insertQuery = `
        INSERT INTO vehicle_movement (
		vehicle_id,
		oid,
		latitude,
		longitude,
		altitude,
		direction,
		speed,
		navigation_system,
		satellite_count,
		sent_unix_time,
		received_unix_time
		)
        VALUES (
		$1, 
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11)
        RETURNING id
    `
	var id int32
	if err := p.connector.GetConnection().QueryRow(insertQuery, vehicleID, data.OID,
		data.Latitude, data.Longitude, data.Altitude, data.Direction,
		data.Speed, data.NavigationSystem, data.SatelliteCount,
		data.SentTimestamp, data.ReceivedTimestamp).Scan(&id); err != nil {
		return 0, fmt.Errorf("не удалось вставить запись: %v", err)
	}

	return id, nil
}

func (p *PrimarySource) GetLastVehiclePosition(vehicleID int32) (types.Position, error) {
	db, err := p.db()
	if err != nil {
		return types.Position{}, err
	}

	const q = `
	SELECT
		latitude,
		longitude,
		altitude
	FROM vehicle_movement
	WHERE vehicle_id = $1
		AND sent_unix_time IS NOT NULL
		AND latitude IS NOT NULL
		AND longitude IS NOT NULL
	ORDER BY sent_unix_time DESC
	LIMIT 1
	`
	var latStr, lonStr, altStr sql.NullString
	if err := db.QueryRow(q, vehicleID).Scan(&latStr, &lonStr, &altStr); err != nil {
		if err == sql.ErrNoRows {
			return types.Position{}, fmt.Errorf("нет записей местоположения для транспорта с ID %d", vehicleID)
		}
		return types.Position{}, err
	}

	var pos types.Position
	if latStr.Valid {
		if v, err := strconv.ParseFloat(latStr.String, 64); err == nil {
			pos.Latitude = v
		}
	}
	if lonStr.Valid {
		if v, err := strconv.ParseFloat(lonStr.String, 64); err == nil {
			pos.Longitude = v
		}
	}
	if altStr.Valid {
		if v, err := strconv.ParseUint(altStr.String, 10, 32); err == nil {
			pos.Altitude = uint32(v)
		}
	}

	return pos, nil
}
