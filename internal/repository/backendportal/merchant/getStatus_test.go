package merchant_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/merchant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/mock"
)

func TestFindStatusByID(t *testing.T) {
	merchantId := uuid.NewString()

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get merchant status by merchant ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrMerchantStatusResponseMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					merchantStatus := args.Get(1).(*merchant.MerchantStatusResponse)
					*merchantStatus = merchant.MerchantStatusResponse{
						UUID:         "merchant-id",
						Status:       constant.MerchantStatusActive,
						ReasonStatus: "",
					}
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Merchant status not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrMerchantStatusResponseMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrMerchantStatusResponseMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
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
			_, err := repo.FindStatusByID(context.Background(), merchantId)
			if (err != nil) != tc.wantErr {
				t.Errorf("MerchantRepository.FindStatusByID() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
