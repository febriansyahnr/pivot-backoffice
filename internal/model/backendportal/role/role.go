package role

import (
	"database/sql"
	"strings"
	"time"
)

type Role struct {
	UUID        string         `json:"uuid" db:"uuid"`
	MerchantID  sql.NullString `json:"merchantId" db:"merchant_id"`
	Name        string         `json:"name" db:"name"`
	Slug        string         `json:"slug" db:"slug"`
	Type        string         `json:"type" db:"type"`
	CreatedAt   time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time      `json:"updatedAt" db:"updated_at"`
	DeletedAt   sql.NullTime   `json:"deletedAt" db:"deleted_at"`
	Permissions sql.NullString `db:"permissions" json:"-"`
}

type RoleRequest struct {
	Name string `json:"name" validate:"required"`
}

type RoleResponse struct {
	UUID        string    `json:"uuid"`
	MerchantID  *string   `json:"merchantId"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Type        string    `json:"type"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Permissions []string  `json:"permissions"`
}

func (r *Role) ToResponse() *RoleResponse {
	response := &RoleResponse{
		UUID:        r.UUID,
		Name:        r.Name,
		Slug:        r.Slug,
		Type:        r.Type,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		Permissions: []string{},
	}

	if r.MerchantID.Valid {
		response.MerchantID = &r.MerchantID.String
	}
	if r.Permissions.Valid {
		response.Permissions = strings.Split(r.Permissions.String, ",")
	}

	return response
}
