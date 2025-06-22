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

	const q = "SELECT id, imei, license_plate_number, vehicle_directory_id FROM vehicle"
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []aux.Vehicle
	for rows.Next() {
		var v aux.Vehicle
		if err := rows.Scan(&v.ID, &v.IMEI, &v.LicensePlateNumber, &v.DirectoryID); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (p *PostgresAuxSource) GetAllDirectories() ([]aux.VehicleDirectory, error) {
	db, err := p.db()
	if err != nil {
		return nil, err
	}

	const q = "SELECT id, provider_id FROM vehicle_directory"
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dirs []aux.VehicleDirectory
	for rows.Next() {
		var d aux.VehicleDirectory
		if err := rows.Scan(&d.ID, &d.ProviderID); err != nil {
			return nil, err
		}
		dirs = append(dirs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dirs, nil
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
		       v.license_plate_number,
		       v.vehicle_directory_id
		FROM vehicle v
		JOIN vehicle_directory vd ON vd.id = v.vehicle_directory_id
		JOIN provider_to_ip pi     ON pi.provider_id = vd.provider_id
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
		if err := rows.Scan(&v.ID, &v.IMEI, &v.LicensePlateNumber, &v.DirectoryID); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return vehicles, nil
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
