package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/mock"
)

func TestMerchantRepositoryGetMerchantAuthByMerchantId(t *testing.T) {
	clientID := "client-id"

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get merchant auth by ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					merchantAuth := args.Get(1).(*merchant.MerchantAuth)
					*merchantAuth = merchant.MerchantAuth{
						UUID:       clientID,
						MerchantID: "merchant-id",
					}
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Merchant auth not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(errors.New("invalid-query"))
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
			_, err := repo.GetMerchantAuthByMerchantId(context.Background(), clientID)
			if (err != nil) != tc.wantErr {
				t.Errorf("MerchantRepository.GetMerchantAuthByMerchantId() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestMerchantRepositoryGetMerchantPKCS8(t *testing.T) {
	clientID := "client-id"

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get merchant auth by ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					merchantAuth := args.Get(1).(*merchant.MerchantAuth)
					*merchantAuth = merchant.MerchantAuth{
						UUID:       clientID,
						MerchantID: "merchant-id",
					}
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Merchant auth not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(errors.New("invalid-query"))
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
			_, err := repo.GetMerchantSnapPKCS8KeyByMerchantID(context.Background(), clientID)
			if (err != nil) != tc.wantErr {
				t.Errorf("MerchantRepository.GetMerchantSnapPKCS8KeyByMerchantID() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
