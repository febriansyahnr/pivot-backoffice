package menuModel

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
)

type Menu struct {
	UUID        string         `json:"uuid" db:"uuid"`
	Slug        string         `json:"slug" db:"slug"`
	Name        string         `json:"name" db:"name"`
	Type        string         `json:"type" db:"type"`
	Icon        string         `json:"icon" db:"icon"`
	Path        string         `json:"path" db:"path"`
	Level       int            `json:"level" db:"level"`
	Order       int            `json:"order" db:"order"`
	ParentID    *string        `json:"parentId" db:"parent_id"`
	CreatedAt   time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time      `json:"updatedAt" db:"updated_at"`
	Permissions sql.NullString `json:"-" db:"permissions"`
	Children    []*Menu        `json:"children"`

	AllowedProducts types.NullJSONText `json:"-" db:"allowed_products"`
}

func (m *Menu) ToResponse() *MenuResponse {
	response := &MenuResponse{
		UUID:        m.UUID,
		Slug:        m.Slug,
		Name:        m.Name,
		Type:        m.Type,
		Icon:        m.Icon,
		Path:        m.Path,
		Level:       m.Level,
		Order:       m.Order,
		ParentID:    m.ParentID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Permissions: []MenuPermission{},
	}

	if m.Permissions.Valid {
		menuPermissions := strings.Split(m.Permissions.String, ",") // format = group:slug

		permissions := make([]MenuPermission, len(menuPermissions))
		if len(menuPermissions) > 0 {
			for i, menuPermission := range menuPermissions {
				splitMenuPermission := strings.Split(menuPermission, ":")
				permissions[i] = MenuPermission{
					Group: splitMenuPermission[0],
					Slug:  splitMenuPermission[1],
				}
			}
		}
		response.Permissions = permissions
	}

	if m.AllowedProducts.Valid {
		var allowedProducts []string
		if err := json.Unmarshal(m.AllowedProducts.JSONText, &allowedProducts); err == nil {
			response.AllowedProducts = &allowedProducts
		}
	}

	children := make([]*MenuResponse, len(m.Children))
	for i, child := range m.Children {
		children[i] = child.ToResponse()
	}
	response.Children = children

	return response
}

type MenuAndPermissionIDs struct {
	MenuID         string          `db:"menu_id"`
	MenuName       string          `db:"menu_name"`
	PermissionsStr string          `db:"permissions"`
	Permissions    []ObjPermission `db:"-"`
}

type ObjPermission struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
