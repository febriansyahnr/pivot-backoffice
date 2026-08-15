package transferRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/transfer"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Count total items",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(errors.New("errors"))

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Get List",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(errors.New("errors"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No rows",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)

			_, _, err := repo.GetList(ctx, &transfer.GetTransferListRequest{
				MerchantID:  uuid.New().String(),
				ReferenceID: uuid.New().String(),
			}, 1, 10)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestGenerateMerchantClauseAndParam(t *testing.T) {
	repo := transferRepository{}

	tests := []struct {
		name        string
		req         *transfer.GetTransferListRequest
		wantClauses []string
		wantParams  []interface{}
	}{
		{
			name: "ParentID and Type empty",
			req: &transfer.GetTransferListRequest{
				MerchantID: "merchant1",
			},
			wantClauses: []string{"(t.merchant_id = ? OR t.recipient_id = ?)"},
			wantParams:  []interface{}{"merchant1", "merchant1"},
		},
		{
			name: "ParentID present, Type empty",
			req: &transfer.GetTransferListRequest{
				ParentID:   "parent1",
				MerchantID: "merchant1",
			},
			wantClauses: []string{"(t.merchant_id IN (?, ?) AND t.recipient_id IN (?, ?))"},
			wantParams:  []interface{}{"parent1", "merchant1", "parent1", "merchant1"},
		},
		{
			name: "ParentID present, Type IN",
			req: &transfer.GetTransferListRequest{
				ParentID:   "parent1",
				MerchantID: "merchant1",
				Type:       constant.TransferTypeIN,
			},
			wantClauses: []string{"t.recipient_id = ? AND t.merchant_id = ?"},
			wantParams:  []interface{}{"merchant1", "parent1"},
		},
		{
			name: "ParentID present, Type OUT",
			req: &transfer.GetTransferListRequest{
				ParentID:   "parent1",
				MerchantID: "merchant1",
				Type:       constant.TransferTypeOUT,
			},
			wantClauses: []string{"t.merchant_id = ? AND t.recipient_id = ?"},
			wantParams:  []interface{}{"merchant1", "parent1"},
		},
		{
			name: "Type IN",
			req: &transfer.GetTransferListRequest{
				MerchantID: "merchant1",
				Type:       constant.TransferTypeIN,
			},
			wantClauses: []string{"t.recipient_id = ?"},
			wantParams:  []interface{}{"merchant1"},
		},
		{
			name: "Type OUT",
			req: &transfer.GetTransferListRequest{
				MerchantID: "merchant1",
				Type:       constant.TransferTypeOUT,
			},
			wantClauses: []string{"t.merchant_id = ? "},
			wantParams:  []interface{}{"merchant1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clauses, params := repo.GenerateMerchantClauseAndParam(context.Background(), test.req)
			assert.Equal(t, test.wantClauses, clauses)
			assert.Equal(t, test.wantParams, params)
		})
	}
}
