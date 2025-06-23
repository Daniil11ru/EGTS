package pg

import (
	"database/sql"
	"fmt"

	connector "github.com/daniil11ru/egts/cli/receiver/connector"
	"github.com/daniil11ru/egts/cli/receiver/source/primary"
	"github.com/lib/pq"
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

func (p *PrimarySource) GetAllVehicles() ([]primary.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}

	const q = "SELECT id, imei, oid, license_plate_number, provider_id, moderation_status FROM vehicle"
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []primary.Vehicle
	for rows.Next() {
		var v primary.Vehicle
		if err := rows.Scan(&v.ID, &v.IMEI, &v.OID, &v.LicensePlateNumber, &v.ProviderID, &v.ModerationStatus); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (p *PrimarySource) GetAllProviders() ([]primary.Provider, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}

	const q = `
		SELECT pr.id,
		       pr.name,
		       COALESCE(array_remove(array_agg(pi.ip), NULL), '{}') AS ips
		FROM provider pr
		LEFT JOIN provider_to_ip pi ON pi.provider_id = pr.id
		GROUP BY pr.id, pr.name
	`
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []primary.Provider
	for rows.Next() {
		var pr primary.Provider
		var ips pq.StringArray
		if err := rows.Scan(&pr.ID, &pr.Name, &ips); err != nil {
			return nil, err
		}
		pr.IP = []string(ips)
		providers = append(providers, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return providers, nil
}

func (p *PrimarySource) GetVehiclesByProviderIP(ip string) ([]primary.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}

	const q = `
		SELECT v.id,
		       v.imei,
		       v.oid,
		       v.license_plate_number,
		       v.provider_id,
			   v.moderation_status
		FROM vehicle v
		JOIN provider_to_ip pi ON pi.provider_id = v.provider_id
		WHERE pi.ip = $1
	`
	rows, err := db.Query(q, ip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []primary.Vehicle
	for rows.Next() {
		var v primary.Vehicle
		if err := rows.Scan(&v.ID, &v.IMEI, &v.OID, &v.LicensePlateNumber, &v.ProviderID, &v.ModerationStatus); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (p *PrimarySource) GetVehicleByID(id int32) (primary.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return primary.Vehicle{}, err
	}

	const q = `
		SELECT id, imei, oid, license_plate_number, provider_id, moderation_status
		FROM vehicle
		WHERE id = $1
	`
	row := db.QueryRow(q, id)

	var v primary.Vehicle
	if err := row.Scan(
		&v.ID,
		&v.IMEI,
		&v.OID,
		&v.LicensePlateNumber,
		&v.ProviderID,
		&v.ModerationStatus,
	); err != nil {
		if err == sql.ErrNoRows {
			return primary.Vehicle{}, fmt.Errorf("транспорт с ID %d не найден", id)
		}
		return primary.Vehicle{}, err
	}

	return v, nil
}

func (p *PrimarySource) GetVehicleByOID(oid int32) (primary.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return primary.Vehicle{}, err
	}

	const q = `
        SELECT id, imei, oid, license_plate_number, provider_id, moderation_status
        FROM vehicle
        WHERE oid = $1
    `
	rows, err := db.Query(q, oid)
	if err != nil {
		return primary.Vehicle{}, err
	}
	defer rows.Close()

	var (
		v     primary.Vehicle
		count int
	)
	for rows.Next() {
		var tmp primary.Vehicle
		if err := rows.Scan(&tmp.ID, &tmp.IMEI, &tmp.OID, &tmp.LicensePlateNumber, &tmp.ProviderID, &tmp.ModerationStatus); err != nil {
			return primary.Vehicle{}, err
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
		return primary.Vehicle{}, err
	}

	switch count {
	case 0:
		return primary.Vehicle{}, fmt.Errorf("транспорт с OID %d не найден", oid)
	case 1:
		return v, nil
	default:
		return primary.Vehicle{}, fmt.Errorf("найдено множество транспортных единиц с OID %d", oid)
	}
}

func (p *PrimarySource) GetVehicleByOIDAndProviderID(oid int32, providerID int32) (primary.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return primary.Vehicle{}, err
	}

	const q = `
		SELECT id, imei, oid, license_plate_number, provider_id, moderation_status
		FROM vehicle
		WHERE oid = $1 AND provider_id = $2
	`
	rows, err := db.Query(q, oid, providerID)
	if err != nil {
		return primary.Vehicle{}, err
	}
	defer rows.Close()

	var (
		v     primary.Vehicle
		count int
	)
	for rows.Next() {
		var tmp primary.Vehicle
		if err := rows.Scan(&tmp.ID, &tmp.IMEI, &tmp.OID, &tmp.LicensePlateNumber, &tmp.ProviderID, &tmp.ModerationStatus); err != nil {
			return primary.Vehicle{}, err
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
		return primary.Vehicle{}, err
	}

	switch count {
	case 0:
		return primary.Vehicle{}, fmt.Errorf("транспорт с OID %d и провайдером %d не найден", oid, providerID)
	case 1:
		return v, nil
	default:
		return primary.Vehicle{}, fmt.Errorf("найдено несколько транспортных единиц с OID %d и провайдером %d", oid, providerID)
	}
}

func (p *PrimarySource) AddVehicle(v primary.Vehicle) (int32, error) {
	db, err := p.db()
	if err != nil {
		return 0, err
	}

	const q = `
        INSERT INTO vehicle (imei, oid, license_plate_number, provider_id, moderation_status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	var id int32
	if err := db.QueryRow(q, v.IMEI, v.OID, v.LicensePlateNumber, v.ProviderID, v.ModerationStatus).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (p *PrimarySource) UpdateVehicleOID(id int32, oid int32) error {
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

func (p *PrimarySource) GetProviderByIP(ip string) (primary.Provider, error) {
	db, err := p.db()
	if err != nil {
		return primary.Provider{}, err
	}

	const q = `
		SELECT pr.id,
		       pr.name,
		       COALESCE(array_remove(array_agg(pi2.ip), NULL), '{}') AS ips
		FROM provider pr
		JOIN provider_to_ip pi1 ON pi1.provider_id = pr.id
		LEFT JOIN provider_to_ip pi2 ON pi2.provider_id = pr.id
		WHERE pi1.ip = $1
		GROUP BY pr.id, pr.name
	`

	row := db.QueryRow(q, ip)
	var pr primary.Provider
	var ips pq.StringArray
	if err := row.Scan(&pr.ID, &pr.Name, &ips); err != nil {
		if err == sql.ErrNoRows {
			return primary.Provider{}, fmt.Errorf("провайдер с IP %s не найден", ip)
		}
		return primary.Provider{}, err
	}
	pr.IP = []string(ips)
	return pr, nil
}

func (p *PrimarySource) GetAllIPs() ([]string, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}
	const q = `
		SELECT DISTINCT ip
		FROM provider_to_ip
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

func (p *PrimarySource) AddVehicleMovement(message interface{ ToBytes() ([]byte, error) }, vehicleID int) (int32, error) {
	if message == nil {
		return 0, fmt.Errorf("некорректная ссылка на пакет")
	}

	packet, err := message.ToBytes()
	if err != nil {
		return 0, fmt.Errorf("ошибка сериализации пакета: %v", err)
	}

	const insertQuery = `
        INSERT INTO vehicle_movement (data, vehicle_id)
        VALUES ($1, $2)
        RETURNING id
    `
	var id int32
	if err := p.connector.GetConnection().QueryRow(insertQuery, packet, vehicleID).Scan(&id); err != nil {
		return 0, fmt.Errorf("не удалось вставить запись: %v", err)
	}

	return id, nil
}
