package httputil

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

type MockParam struct {
	MerchantID string
}

func TestBindSubmerchantID(t *testing.T) {

	t.Run("SubmerchantID Exists", func(t *testing.T) {
		param := MockParam{
			MerchantID: uuid.Max.String(),
		}

		r, _ := http.NewRequest("GET", "http://localhost:8080", nil)
		r.Header.Set("X-SubMerchant-Id", "12345")

		BindSubmerchantID(r, &param.MerchantID)
		assert.NotEqual(t, param.MerchantID, uuid.Max.String())
		assert.Equal(t, param.MerchantID, "12345")
	})

	t.Run("SubmerchantID Not Exists", func(t *testing.T) {
		param := MockParam{
			MerchantID: uuid.Max.String(),
		}

		r, _ := http.NewRequest("GET", "http://localhost:8080", nil)

		BindSubmerchantID(r, &param.MerchantID)
		assert.Equal(t, param.MerchantID, uuid.Max.String())
		assert.NotEqual(t, param.MerchantID, "12345")

	})

}

func TestBindLoggedInUserType(t *testing.T) {

	t.Run("SubmerchantID Exists", func(t *testing.T) {
		userType := ""
		r, _ := http.NewRequest("GET", "http://localhost:8080", nil)
		r.Header.Set("X-SubMerchant-Id", "12345")

		BindLoggedInUserType(r, &userType)
		assert.Equal(t, userType, constant.UserTypeSubMerchant)
	})

	t.Run("SubmerchantID Not Exists", func(t *testing.T) {
		userType := ""
		r, _ := http.NewRequest("GET", "http://localhost:8080", nil)

		BindLoggedInUserType(r, &userType)
		assert.Equal(t, userType, constant.UserTypeMerchant)

	})

}
