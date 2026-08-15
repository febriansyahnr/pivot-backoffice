package userRole

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	UUID      string       `json:"uuid" db:"uuid"`
	UserID    string       `json:"userId" db:"user_id"`
	RoleID    string       `json:"roleId" db:"role_id"`
	CreatedAt time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time    `json:"updatedAt" db:"updated_at"`
	DeletedAt sql.NullTime `json:"deletedAt" db:"deleted_at"`
}

func New(userId, roleId string) *UserRole {
	return &UserRole{
		UUID:      uuid.New().String(),
		UserID:    userId,
		RoleID:    roleId,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}
