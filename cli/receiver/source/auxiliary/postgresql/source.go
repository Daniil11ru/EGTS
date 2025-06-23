package postgresql

import (
	"database/sql"
	"fmt"

	connector "github.com/daniil11ru/egts/cli/receiver/connector"
	aux "github.com/daniil11ru/egts/cli/receiver/source/auxiliary"
	"github.com/lib/pq"
)

type PostgresAuxSource struct {
	conn connector.Connector
}

func (p *PostgresAuxSource) Initialize(c connector.Connector) {
	p.conn = c
}

func (p *PostgresAuxSource) db() (*sql.DB, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("не удалось инициализировать подключение к базе данных")
	}
	db := p.conn.GetConnection()
	if db == nil {
		return nil, fmt.Errorf("нет активного подключения к базе данных")
	}
	return db, nil
}

func (p *PostgresAuxSource) GetAllVehicles() ([]aux.Vehicle, error) {
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

	var vehicles []aux.Vehicle
	for rows.Next() {
		var v aux.Vehicle
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

func (p *PostgresAuxSource) GetAllProviders() ([]aux.Provider, error) {
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

	var providers []aux.Provider
	for rows.Next() {
		var pr aux.Provider
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

func (p *PostgresAuxSource) GetVehiclesByProviderIP(ip string) ([]aux.Vehicle, error) {
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

	var vehicles []aux.Vehicle
	for rows.Next() {
		var v aux.Vehicle
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

func (p *PostgresAuxSource) GetVehicleByID(id int32) (aux.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return aux.Vehicle{}, err
	}

	const q = `
		SELECT id, imei, oid, license_plate_number, provider_id, moderation_status
		FROM vehicle
		WHERE id = $1
	`
	row := db.QueryRow(q, id)

	var v aux.Vehicle
	if err := row.Scan(
		&v.ID,
		&v.IMEI,
		&v.OID,
		&v.LicensePlateNumber,
		&v.ProviderID,
		&v.ModerationStatus,
	); err != nil {
		if err == sql.ErrNoRows {
			return aux.Vehicle{}, fmt.Errorf("транспорт с ID %d не найден", id)
		}
		return aux.Vehicle{}, err
	}

	return v, nil
}

func (p *PostgresAuxSource) GetVehicleByOID(oid int32) (aux.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return aux.Vehicle{}, err
	}

	const q = `
        SELECT id, imei, oid, license_plate_number, provider_id, moderation_status
        FROM vehicle
        WHERE oid = $1
    `
	rows, err := db.Query(q, oid)
	if err != nil {
		return aux.Vehicle{}, err
	}
	defer rows.Close()

	var (
		v     aux.Vehicle
		count int
	)
	for rows.Next() {
		var tmp aux.Vehicle
		if err := rows.Scan(&tmp.ID, &tmp.IMEI, &tmp.OID, &tmp.LicensePlateNumber, &tmp.ProviderID, &tmp.ModerationStatus); err != nil {
			return aux.Vehicle{}, err
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
		return aux.Vehicle{}, err
	}

	switch count {
	case 0:
		return aux.Vehicle{}, fmt.Errorf("транспорт с OID %d не найден", oid)
	case 1:
		return v, nil
	default:
		return aux.Vehicle{}, fmt.Errorf("найдено множество транспортных единиц с OID %d", oid)
	}
}

func (p *PostgresAuxSource) GetVehicleByOIDAndProviderID(oid int32, providerID int32) (aux.Vehicle, error) {
	db, err := p.db()
	if err != nil {
		return aux.Vehicle{}, err
	}

	const q = `
		SELECT id, imei, oid, license_plate_number, provider_id, moderation_status
		FROM vehicle
		WHERE oid = $1 AND provider_id = $2
	`
	rows, err := db.Query(q, oid, providerID)
	if err != nil {
		return aux.Vehicle{}, err
	}
	defer rows.Close()

	var (
		v     aux.Vehicle
		count int
	)
	for rows.Next() {
		var tmp aux.Vehicle
		if err := rows.Scan(&tmp.ID, &tmp.IMEI, &tmp.OID, &tmp.LicensePlateNumber, &tmp.ProviderID, &tmp.ModerationStatus); err != nil {
			return aux.Vehicle{}, err
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
		return aux.Vehicle{}, err
	}

	switch count {
	case 0:
		return aux.Vehicle{}, fmt.Errorf("транспорт с OID %d и провайдером %d не найден", oid, providerID)
	case 1:
		return v, nil
	default:
		return aux.Vehicle{}, fmt.Errorf("найдено несколько транспортных единиц с OID %d и провайдером %d", oid, providerID)
	}
}

func (p *PostgresAuxSource) AddVehicle(v aux.Vehicle) (int32, error) {
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

func (p *PostgresAuxSource) UpdateVehicleOID(id int32, oid int32) error {
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

func (p *PostgresAuxSource) GetProviderByIP(ip string) (aux.Provider, error) {
	db, err := p.db()
	if err != nil {
		return aux.Provider{}, err
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
	var pr aux.Provider
	var ips pq.StringArray
	if err := row.Scan(&pr.ID, &pr.Name, &ips); err != nil {
		if err == sql.ErrNoRows {
			return aux.Provider{}, fmt.Errorf("провайдер с IP %s не найден", ip)
		}
		return aux.Provider{}, err
	}
	pr.IP = []string(ips)
	return pr, nil
}

func (p *PostgresAuxSource) GetAllIPs() ([]string, error) {
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
