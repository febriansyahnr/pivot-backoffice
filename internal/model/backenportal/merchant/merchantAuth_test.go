package merchant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMerchantAuth(t *testing.T) {

	req := &NewMerchantAuthRequest{
		MerchantID:          "merchant-id",
		Secret:              "vault:v2:lCOQrY4A5WYCqlSJRzFi3S50nF7QlrWdRMxGZlXaa49qG0Zftt3LrPNr/CNpSz3deUE01mZZv2RGTCaqV8GG+4UM1Q0=",
		SecretVersion:       2,
		SnapPrivateKey:      "snap-private-key",
		SnapPrivateKeyValid: true,
	}
	merchantAuth := NewMerchantAuth(req)
	assert.NotEmpty(t, merchantAuth.CreatedAt)
	assert.NotEmpty(t, merchantAuth.UpdatedAt)
	assert.Empty(t, merchantAuth.DeletedAt)
	assert.NotEmpty(t, merchantAuth.MerchantID)
	assert.NotEmpty(t, merchantAuth.UUID)
	assert.NotEmpty(t, merchantAuth.Secret)
	assert.NotEmpty(t, merchantAuth.SnapPrivateKey)
}
