package account

import (
	"context"
	"encoding/json"
	"errors"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBulkCreateAccount(t *testing.T) {

	testCases := []struct {
		Name      string
		MockSetup func(accSvc *serviceMocks.IAccountService)
		Input     func() []byte
		WantErr   bool
	}{
		{
			Name: "SUCCESS: Bulk Create Account",
			MockSetup: func(accSvc *serviceMocks.IAccountService) {
				accSvc.On("BulkCreateAccount", mock.Anything, mock.Anything).Return(nil)
			},
			Input: func() []byte {
				req := &account_model.BulkCreateAccountRequest{
					Currency: "IDR",
					Usecase:  constant.ReferenceWallet,
				}
				inputBytes, _ := json.Marshal(req)
				return inputBytes
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Bulk Create Account",
			MockSetup: func(accSvc *serviceMocks.IAccountService) {
				accSvc.On("BulkCreateAccount", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			Input: func() []byte {
				req := &account_model.BulkCreateAccountRequest{
					Currency: "IDR",
					Usecase:  constant.ReferenceWallet,
				}
				inputBytes, _ := json.Marshal(req)
				return inputBytes
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accSvc := serviceMocks.NewIAccountService(t)
			logger := pdkLoggerMock.NewILogger(t)
			tc.MockSetup(accSvc)
			consumer := New(logger, accSvc)
			err := consumer.BulkCreateAccount(context.Background(), tc.Input(), "")
			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
