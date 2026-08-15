package user_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/user"

	"github.com/skip2/go-qrcode"
	"github.com/stretchr/testify/assert"
)

func TestEnrollTOTPRequest(t *testing.T) {
	tests := []struct {
		request         EnrollTOTPRequest
		wantQrCodeLevel qrcode.RecoveryLevel
	}{
		{
			request:         EnrollTOTPRequest{},
			wantQrCodeLevel: qrcode.Medium,
		},
		{
			request: EnrollTOTPRequest{
				QrCodeLevel: "Low",
			},
			wantQrCodeLevel: qrcode.Low,
		},
		{
			request: EnrollTOTPRequest{
				QrCodeLevel: "Medium",
			},
			wantQrCodeLevel: qrcode.Medium,
		},
		{
			request: EnrollTOTPRequest{
				QrCodeLevel: "High",
			},
			wantQrCodeLevel: qrcode.High,
		},
		{
			request: EnrollTOTPRequest{
				QrCodeLevel: "Highest",
			},
			wantQrCodeLevel: qrcode.Highest,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantQrCodeLevel, test.request.GetQrCodeLevel())
	}
}
