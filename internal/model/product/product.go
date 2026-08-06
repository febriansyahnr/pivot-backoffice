package product

import "time"

type Product struct {
	UUID      string    `json:"uuid" db:"uuid"`
	Name      string    `json:"name" db:"name"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type UpdateProductRequest struct {
	ID     string `json:"id" validate:"required"`
	Active bool   `json:"active" `
}
