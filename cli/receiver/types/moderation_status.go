package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ModerationStatus string

const (
	ModerationStatusPending  ModerationStatus = "pending"
	ModerationStatusApproved ModerationStatus = "approved"
	ModerationStatusRejected ModerationStatus = "rejected"
)

var moderationSet = map[ModerationStatus]struct{}{
	ModerationStatusPending:  {},
	ModerationStatusApproved: {},
	ModerationStatusRejected: {},
}

func (ms ModerationStatus) IsValid() bool {
	_, ok := moderationSet[ms]
	return ok
}

func ParseModerationStatus(s string) (ModerationStatus, error) {
	v := ModerationStatus(s)
	if !v.IsValid() {
		return "", fmt.Errorf("недопустимый moderation_status: %q", s)
	}
	return v, nil
}

func (ms *ModerationStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	v := ModerationStatus(s)
	if !v.IsValid() {
		return fmt.Errorf("недопустимый moderation_status: %q", s)
	}
	*ms = v
	return nil
}

func (ms ModerationStatus) MarshalJSON() ([]byte, error) {
	if !ms.IsValid() {
		return nil, fmt.Errorf("недопустимый moderation_status: %q", string(ms))
	}
	return json.Marshal(string(ms))
}

func (ms *ModerationStatus) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		*ms = ModerationStatus(string(v))
	case string:
		*ms = ModerationStatus(v)
	default:
		return fmt.Errorf("невозможно извлечь ModerationStatus из %T", value)
	}
	if !ms.IsValid() {
		return fmt.Errorf("недопустимый ModerationStatus: %q", string(*ms))
	}
	return nil
}

func (ms ModerationStatus) Value() (driver.Value, error) {
	if !ms.IsValid() {
		return nil, fmt.Errorf("недопустимый ModerationStatus: %q", string(ms))
	}
	return string(ms), nil
}
