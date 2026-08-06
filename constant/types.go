package constant

import (
	"database/sql"
	"encoding/json"
	"time"
)

type NullString struct {
	sql.NullString
}

func (n *NullString) MarshalJSON() ([]byte, error) {
	var s *string
	if n.Valid {
		s = &n.String
	}
	return json.Marshal(s)
}

type NullTime struct {
	sql.NullTime
}

func (n *NullTime) MarshalJSON() ([]byte, error) {
	var t *time.Time
	if n.Valid {
		t = &n.Time
	}
	return json.Marshal(t)
}
