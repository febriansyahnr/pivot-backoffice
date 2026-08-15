package permissionModel_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"

	"github.com/stretchr/testify/assert"
)

func TestPermissionUpdatePermission(t *testing.T) {
	new := Permission{
		Slug:        "Slup",
		Name:        "Name",
		Description: "Description",
		Group:       "Group",
	}

	old := Permission{}
	old.UpdatePermission(&new)

	assert.Equal(t, new.Slug, old.Slug)
	assert.Equal(t, new.Name, old.Name)
	assert.Equal(t, new.Description, old.Description)
	assert.Equal(t, new.Group, old.Group)
	assert.False(t, old.UpdatedAt.IsZero())
}
