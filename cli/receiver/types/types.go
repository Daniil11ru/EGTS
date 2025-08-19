package types

import (
	"database/sql"
)

type Track2D struct {
	VehicleID int32
	Movements []Movement2D
}

type Vehicle struct {
	ID               int32
	IMEI             int64
	OID              sql.NullInt64
	Name             sql.NullString
	ProviderID       int32
	ModerationStatus ModerationStatus
}

type Provider struct {
	ID   int32
	Name string
}

type Movement2D struct {
	Position2D
	ID int32
}
