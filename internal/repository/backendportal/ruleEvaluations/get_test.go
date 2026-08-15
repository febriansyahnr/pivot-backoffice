package ruleevaluationsrepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ruleEvaluations"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByID(t *testing.T) {
	ruleEval := &ruleevaluationsmodel.RuleEvaluations{
		UUID:        "uuid-uuid-uuid",
		ReferenceID: "uuid-uuid-uuid",
		RuleID:      "uuid-uuid-uuid",
		Result:      "0",
		Score:       decimal.NewFromInt(0),
		Reason:      "reason",
	}

	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		merchantId string
		bankCode   string
		accountNo  string
		expected   *ruleevaluationsmodel.RuleEvaluations
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					userPtr := args.Get(1).(*ruleevaluationsmodel.RuleEvaluations)
					*userPtr = *ruleEval
				})
			},
			expected: ruleEval,
			wantErr:  false,
		},
		{
			name: "ERROR: Account Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*ruleevaluationsmodel.RuleEvaluations"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockLogger, mockMysql)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)
			userRes, err := repo.GetByID(ctx, uuid.NewString())

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, userRes)
			}
		})
	}
}

func TestGetByRefID(t *testing.T) {
	ruleEvals := &[]ruleevaluationsmodel.RuleEvaluations{
		{
			UUID:        "uuid-uuid-uuid",
			ReferenceID: "uuid-uuid-uuid",
			RuleID:      "uuid-uuid-uuid",
			Result:      "0",
			Score:       decimal.NewFromInt(0),
			Reason:      "reason",
		},
	}

	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		merchantId string
		bankCode   string
		accountNo  string
		expected   *[]ruleevaluationsmodel.RuleEvaluations
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]ruleevaluationsmodel.RuleEvaluations"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					userPtr := args.Get(1).(*[]ruleevaluationsmodel.RuleEvaluations)
					*userPtr = *ruleEvals
				})
			},
			expected: ruleEvals,
			wantErr:  false,
		},
		{
			name: "ERROR: Account Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"SelectContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*[]ruleevaluationsmodel.RuleEvaluations"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"SelectContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*[]ruleevaluationsmodel.RuleEvaluations"),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(errors.New("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockLogger, mockMysql)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)
			userRes, err := repo.GetByRefID(ctx, uuid.NewString())

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, userRes)
			}
		})
	}
}
