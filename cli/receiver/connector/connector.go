package connector

import (
	"database/sql"
)

type Connector interface {
	GetConnection() *sql.DB
	Connect(map[string]string) error
	Close() error
}
