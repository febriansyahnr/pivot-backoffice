package role

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRole_ToResponse(t *testing.T) {
	// Create a Role instance
	role := &Role{
		UUID:       "1",
		MerchantID: sql.NullString{String: "merchant1", Valid: true},
		Name:       "Role1",
		Slug:       "role1-slug",
		Type:       "type1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Permissions: sql.NullString{
			String: "Permission1,Permission2",
			Valid:  true,
		},
	}

	// Call ToResponse method
	response := role.ToResponse()

	// Assertions
	assert.NotNil(t, response)
	assert.Equal(t, "1", response.UUID)
	assert.NotNil(t, response.MerchantID)
	assert.Equal(t, "merchant1", *response.MerchantID)
	assert.Equal(t, "Role1", response.Name)
	assert.Equal(t, "role1-slug", response.Slug)
	assert.Equal(t, "type1", response.Type)
	assert.True(t, response.CreatedAt.Before(time.Now()))
	assert.True(t, response.UpdatedAt.Before(time.Now()))
	assert.Len(t, response.Permissions, 2)
	assert.ElementsMatch(t, []string{"Permission1", "Permission2"}, response.Permissions)
}
