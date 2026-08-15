package merchant

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/mock"
)

func TestCreateAuth(t *testing.T) {
	merchant := &merchantModel.MerchantAuth{
		UUID:       uuid.NewString(),
		MerchantID: uuid.NewString(),
		Secret:     "generated-secret",
		CreatedAt:  util.TimeNow,
		UpdatedAt:  util.TimeNow,
	}

	testCases := []struct {
		name      string
		merchant  *merchantModel.MerchantAuth
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name:     "SUCCESS: create merchant",
			merchant: merchant,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:     "FAIL: Failure Insert to Database",
			merchant: merchant,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(false, errors.New("insert error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No Rows Affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(false, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.CreateMerchantAuth(context.Background(), tc.merchant)

			if (err != nil) != tc.wantErr {
				t.Errorf("MerchantRepository.CreateMerchantAuth() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
