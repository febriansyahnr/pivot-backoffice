package transferRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByUUID(t *testing.T) {

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get by ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*transfer.Transfer"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Transfer Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*transfer.Transfer"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database Get By UUID Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*transfer.Transfer"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)
			_, err := repo.GetByID(ctx, uuid.NewString(), uuid.NewString())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
func TestGetTransferTransaction(t *testing.T) {
	mockMysql := mysqlMocks.NewIMySqlExt(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	repo := New(mockMysql, mockLogger)

	req := transfer.GetTransferTransactionRequest{
		MerchantID:    uuid.NewString(),
		TransactionID: uuid.NewString(),
	}

	testCases := []struct {
		name      string
		req       transfer.GetTransferTransactionRequest
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		shouldErr bool
		wantErr   error
	}{
		{
			name: "SUCCESS: Get Transfer Transaction",
			req:  req,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					req.MerchantID,
					req.MerchantID,
					req.MerchantID,
					req.ParentID,
					req.MerchantID,
					req.MerchantID,
					req.TransactionID,
				).Return(nil).Once()
			},
			shouldErr: false,
		},
		{
			name: "ERROR: Transfer Transaction Not Found",
			req:  req,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					req.MerchantID,
					req.MerchantID,
					req.MerchantID,
					req.ParentID,
					req.MerchantID,
					req.MerchantID,
					req.TransactionID,
				).Return(sql.ErrNoRows).Once()
			},
			shouldErr: false,
		},
		{
			name: "ERROR: Database Get Transfer Transaction Error",
			req:  req,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.StringMockType(),
					req.MerchantID,
					req.MerchantID,
					req.MerchantID,
					req.ParentID,
					req.MerchantID,
					req.MerchantID,
					req.TransactionID,
				).Return(errors.New("database error")).Once()
			},
			shouldErr: true,
			wantErr:   errors.New("database error"),
		},
		{
			name:      "ERROR: Invalid Request Parameters",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {},
			shouldErr: true,
			wantErr:   errors.New("invalid request parameters"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup(mockMysql)

			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)

			_, err := repo.GetTransferTransaction(ctx, tc.req)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.NoError(t, err)
			mockMysql.AssertExpectations(t)
		})
	}
}
