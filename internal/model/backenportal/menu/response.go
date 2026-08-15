package menuModel

import "time"

type MenuResponse struct {
	UUID        string           `json:"uuid" db:"uuid"`
	Slug        string           `json:"slug" db:"slug"`
	Name        string           `json:"name" db:"name"`
	Type        string           `json:"type" db:"type"`
	Icon        string           `json:"icon" db:"icon"`
	Path        string           `json:"path" db:"path"`
	Level       int              `json:"level" db:"level"`
	Order       int              `json:"order" db:"order"`
	ParentID    *string          `json:"parentId" db:"parent_id"`
	CreatedAt   time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time        `json:"updatedAt" db:"updated_at"`
	Permissions []MenuPermission `json:"permissions" db:"permissions"`
	Children    []*MenuResponse  `json:"children"`

	AllowedProducts *[]string `json:"-"`
}

type MenuPermission struct {
	Group string `json:"group"`
	Slug  string `json:"slug"`
}
