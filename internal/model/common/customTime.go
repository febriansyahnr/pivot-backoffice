package commonModel

import (
	"database/sql"
	"encoding/json"
	"time"
)

// CustomTime struct embeds sql.NullTime
type CustomNullTime struct {
	sql.NullTime
}

type SqlNulTimeInString struct {
	Valid bool   `json:"Valid"`
	Time  string `json:"Time"`
}

// MarshalJSON customizes the JSON output for CustomTime
func (ct CustomNullTime) MarshalJSON() ([]byte, error) {
	if !ct.Valid {
		return json.Marshal(SqlNulTimeInString{
			Valid: false,
			Time:  "",
		})
	}
	return json.Marshal(SqlNulTimeInString{
		Valid: true,
		Time:  ct.Time.Format(time.RFC3339),
	})
}

// Scan implements the Scanner interface for CustomTime
func (ct *CustomNullTime) Scan(value interface{}) error {
	return ct.NullTime.Scan(value)
}
