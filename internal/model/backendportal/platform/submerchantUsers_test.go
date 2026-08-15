package platform

import (
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/stretchr/testify/assert"
)

func TestToSubMerchantUserResponse(t *testing.T) {
	now := time.Now()
	t.Run("SUCCESS: Should return a valid SubMerchantUserResponse", func(t *testing.T) {
		u := &user.User{
			UUID:  "12345",
			Name:  "John Doe",
			Email: "john@example.com",
			Role: sql.NullString{
				String: constant.RoleAdmin,
				Valid:  true,
			},
			Status: constant.UserStatusActive,
			LastLoginAt: commonModel.CustomNullTime{
				NullTime: sql.NullTime{Time: now, Valid: true},
			},
		}

		expectedResponse := &SubMerchantUserResponse{
			UUID:        "12345",
			Name:        "John Doe",
			Email:       "john@example.com",
			Role:        constant.RoleAdmin,
			Status:      constant.UserStatusActive,
			LastLoginAt: now,
		}

		actualResponse := ToSubMerchantUserResponse(u)

		assert.Equal(t, expectedResponse, actualResponse)
	})

	t.Run("SUCCESS: should return a valid SubMerchantUserResponse with empty fields", func(t *testing.T) {
		u := &user.User{}

		expectedResponse := &SubMerchantUserResponse{}

		actualResponse := ToSubMerchantUserResponse(u)

		assert.Equal(t, expectedResponse, actualResponse)
	})
}
