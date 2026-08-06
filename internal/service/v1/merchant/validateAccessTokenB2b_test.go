package merchant

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateAccessTokenB2b(t *testing.T) {
	merchantId := uuid.NewString()
	claims := &merchantModel.MerchantAuthTokenClaims{
		MerchantId: merchantId,
	}
	testCases := []struct {
		name    string
		request merchantModel.ValidateAccessTokenB2bRequest
		setup   func(
			jwtMock *mockJWT.IJwt,
		)
		wantErr bool
	}{
		{
			name: "SUCCESS: Validate merchant token",
			request: merchantModel.ValidateAccessTokenB2bRequest{
				MerchantId:  merchantId,
				AccessToken: "token",
			},
			setup: func(jwtMock *mockJWT.IJwt) {
				jwtMock.On(
					"VerifyMerchantToken",
					mock.Anything,
					mock.Anything,
				).Return(claims, nil)
			},
		},
		{
			name: "ERROR: Expired merchant token",
			request: merchantModel.ValidateAccessTokenB2bRequest{
				MerchantId:  merchantId,
				AccessToken: "token",
			},
			setup: func(jwtMock *mockJWT.IJwt) {
				jwtMock.On(
					"VerifyMerchantToken",
					mock.Anything,
					mock.Anything,
				).Return(nil, constant.ErrExpiredMerchantAuth)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Validate merchant token",
			request: merchantModel.ValidateAccessTokenB2bRequest{
				MerchantId:  merchantId,
				AccessToken: "token",
			},
			setup: func(jwtMock *mockJWT.IJwt) {
				jwtMock.On(
					"VerifyMerchantToken",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Incorrect merchant",
			request: merchantModel.ValidateAccessTokenB2bRequest{
				MerchantId:  uuid.NewString(),
				AccessToken: "token",
			},
			setup: func(jwtMock *mockJWT.IJwt) {
				jwtMock.On(
					"VerifyMerchantToken",
					mock.Anything,
					mock.Anything,
				).Return(claims, nil)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := logger.NewZapLogger(logger.Config{})
			jwt := mockJWT.NewIJwt(t)
			tc.setup(jwt)

			svc := New(nil, logger, nil, jwt, nil, nil)
			_, err := svc.ValidateAccessTokenB2b(context.TODO(), &tc.request)
			if tc.wantErr {
				assert.NotNil(t, err)
			}
		})
	}

}
