package userRole

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	userId := uuid.NewString()
	roleId := uuid.NewString()

	userRole := New(userId, roleId)
	assert.NotEqual(t, uuid.Nil, userRole.UUID)
	assert.Equal(t, userId, userRole.UserID)
	assert.Equal(t, roleId, userRole.RoleID)
}
