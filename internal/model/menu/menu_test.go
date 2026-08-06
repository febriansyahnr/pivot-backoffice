package menuModel

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMenuToResponse(t *testing.T) {
	parentID := uuid.NewString()
	childID := uuid.NewString()

	menu := Menu{
		UUID:      parentID,
		Slug:      "test",
		Name:      "name",
		Icon:      "icon",
		Path:      "/menu",
		Level:     1,
		Order:     1,
		ParentID:  nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Permissions: sql.NullString{
			String: "group:slug",
			Valid:  true,
		},

		Children: []*Menu{
			{
				UUID: childID,
			},
		},
	}

	response := menu.ToResponse()

	assert.Equal(t, menu.UUID, response.UUID)
	assert.Equal(t, menu.Slug, response.Slug)
	assert.Equal(t, menu.Name, response.Name)
	assert.Equal(t, menu.Icon, response.Icon)
	assert.Equal(t, menu.Path, response.Path)
	assert.Equal(t, menu.Level, response.Level)
	assert.Equal(t, menu.Order, response.Order)
	assert.Equal(t, menu.ParentID, response.ParentID)
	assert.Equal(t, menu.CreatedAt, response.CreatedAt)
	assert.Equal(t, menu.UpdatedAt, response.UpdatedAt)
}
