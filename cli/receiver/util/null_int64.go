package util

import (
	"database/sql"
	"fmt"
)

type NullInt64 struct {
	sql.NullInt64
}

func (ni *NullInt64) Scan(value interface{}) error {
	if value == nil {
		ni.Int64, ni.Valid = 0, false
		return nil
	}

	switch v := value.(type) {
	case int64:
		ni.Int64 = v
		ni.Valid = true
	case int32:
		ni.Int64 = int64(v)
		ni.Valid = true
	case int:
		ni.Int64 = int64(v)
		ni.Valid = true
	default:
		return fmt.Errorf("неподдерживаемый тип для OID: %T", value)
	}

	return nil
}
