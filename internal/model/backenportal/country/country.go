package countryModel

import (
	"database/sql"
	"time"
)

type Country struct {
	Code      string       `db:"code"`
	Name      string       `db:"name"`
	NameID    string       `db:"name_id"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

type SearchFilterRequest struct {
	Name   string
	NameID string
	Size   int64
}

type CountryResponse struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	NameID    string    `json:"nameId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (c *Country) ToResponse() *CountryResponse {
	return &CountryResponse{
		Code:      c.Code,
		Name:      c.Name,
		NameID:    c.NameID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
