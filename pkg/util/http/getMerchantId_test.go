package httputil

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBindMerchantID(t *testing.T) {

	t.Run("MerchantID Exists", func(t *testing.T) {
		param := MockParam{
			MerchantID: uuid.Max.String(),
		}

		r, _ := http.NewRequest("GET", "http://localhost:8080", nil)
		r.Header.Set("X-Merchant-Id", "12345")

		BindMerchantID(r, &param.MerchantID)
		assert.NotEqual(t, param.MerchantID, uuid.Max.String())
		assert.Equal(t, param.MerchantID, "12345")
	})

	t.Run("MerchantID Not Exists", func(t *testing.T) {
		param := MockParam{
			MerchantID: uuid.Max.String(),
		}

		r, _ := http.NewRequest("GET", "http://localhost:8080", nil)

		BindSubmerchantID(r, &param.MerchantID)
		assert.Equal(t, param.MerchantID, uuid.Max.String())
		assert.NotEqual(t, param.MerchantID, "12345")

	})
}
