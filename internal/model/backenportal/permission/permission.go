package permissionModel

import "time"

type Permission struct {
	UUID        string    `json:"uuid" db:"uuid"`
	Slug        string    `json:"slug" db:"slug"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Group       string    `json:"group" db:"group"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

func (p *Permission) UpdatePermission(req *Permission) {
	p.Slug = req.Slug
	p.Name = req.Name
	p.Description = req.Description
	p.Group = req.Group
	p.UpdatedAt = time.Now().UTC()
}
