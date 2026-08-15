package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRepository_UpdateMerchantAuth(t *testing.T) {
	now := time.Now()

	dataUpdate := &merchantModel.MerchantAuth{
		UUID:       "uuid-uuid-uuid",
		MerchantID: "merchant-id",
		Secret:     "secret",
		MerchantPublicKey: sql.NullString{
			String: "public-key",
			Valid:  true,
		},
		SnapPrivateKey: sql.NullString{
			String: "private-key",
			Valid:  true,
		},
		CreatedAt: now,
	}

	testCases := []struct {
		desc      string
		input     *merchantModel.MerchantAuth
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			desc:  "SUCCESS: update merchant_auths",
			input: dataUpdate,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(true, nil)
			},
		},
		{
			desc:  "ERROR: update merchant_auths",
			input: dataUpdate,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*merchant.MerchantAuth"),
				).Return(false, errors.New("update error"))
			},
			wantErr: true,
		},
		{
			desc: "ERROR: No Rows Affected",
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
		t.Run(tc.desc, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.UpdateMerchantAuth(context.Background(), tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("MerchantRepository.CreateMerchantAuth() got errpr: %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
