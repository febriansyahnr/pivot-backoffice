package fraudrulesrepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindByQuery(t *testing.T) {
	testCases := []struct {
		name           string
		setupMock      func(mockDB *mysqlMocks.IMySqlExt)
		query          *fraudrulesmodel.FraudRulesQuery
		expectedResult []*fraudrulesmodel.FraudRules
		wantErr        bool
	}{
		{
			name: "success find by empty query",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// Mock the COUNT query
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 0
					}).
					Return(nil)

				// Mock the SELECT query
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*fraudrulesmodel.FraudRules) // fix here
						*resultPtr = []*fraudrulesmodel.FraudRules{}
					}).
					Return(nil)
			},
			query:          &fraudrulesmodel.FraudRulesQuery{},
			expectedResult: []*fraudrulesmodel.FraudRules{},
			wantErr:        false,
		},
		{
			name: "success find by query",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 1
					}).
					Return(nil)

				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*fraudrulesmodel.FraudRules) // fix here
						*resultPtr = []*fraudrulesmodel.FraudRules{}
					}).
					Return(nil)
			},
			query: &fraudrulesmodel.FraudRulesQuery{
				RuleName: "test",
			},
			expectedResult: []*fraudrulesmodel.FraudRules{},
			wantErr:        false,
		},
		{
			name: "failed find by query",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("error db"))
			},
			query: &fraudrulesmodel.FraudRulesQuery{
				RuleName: "test",
			},
			expectedResult: []*fraudrulesmodel.FraudRules{},
			wantErr:        true,
		},
		{
			name: "success with pagination - page 1",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// Mock the COUNT query
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 100
					}).
					Return(nil)

				// Mock the SELECT query
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*fraudrulesmodel.FraudRules)
						*resultPtr = []*fraudrulesmodel.FraudRules{}
					}).
					Return(nil)
			},
			query: &fraudrulesmodel.FraudRulesQuery{
				Page:     1,
				PageSize: 10,
			},
			expectedResult: []*fraudrulesmodel.FraudRules{},
			wantErr:        false,
		},
		{
			name: "success with pagination - page 2",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// Mock the COUNT query
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 100
					}).
					Return(nil)

				// Mock the SELECT query
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*fraudrulesmodel.FraudRules)
						*resultPtr = []*fraudrulesmodel.FraudRules{}
					}).
					Return(nil)
			},
			query: &fraudrulesmodel.FraudRulesQuery{
				Page:     2,
				PageSize: 10,
			},
			expectedResult: []*fraudrulesmodel.FraudRules{},
			wantErr:        false,
		},
		{
			name: "error on SelectContext",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// Mock the COUNT query
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 10
					}).
					Return(nil)

				// Mock the SELECT query with error
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("select error"))
			},
			query: &fraudrulesmodel.FraudRulesQuery{
				Page:     1,
				PageSize: 10,
			},
			expectedResult: []*fraudrulesmodel.FraudRules{},
			wantErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.setupMock(mockDB)

			repository := New(mockLogger, mockDB)
			ctx := context.WithValue(context.Background(), constant.CtxSQLTableNameKey, "bank")
			result, _, err := repository.List(ctx, tc.query)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tc.expectedResult), len(result))
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	ruleEval := &fraudrulesmodel.FraudRules{
		UUID: "uuid-uuid-uuid",
	}

	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		merchantId string
		bankCode   string
		accountNo  string
		expected   *fraudrulesmodel.FraudRules
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*fraudrulesmodel.FraudRules"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					userPtr := args.Get(1).(*fraudrulesmodel.FraudRules)
					*userPtr = *ruleEval
				})
			},
			expected: ruleEval,
			wantErr:  false,
		},
		{
			name: "ERROR: Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*fraudrulesmodel.FraudRules"),
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
						mock.AnythingOfType("*fraudrulesmodel.FraudRules"),
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
