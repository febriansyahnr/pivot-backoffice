package vendor

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/vendor"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByID(t *testing.T) {
	vendor := &vendorModel.Vendor{
		UUID:       "uuid-uuid-uuid",
		MerchantID: "merchant-123",
		Name:       "Test Vendor",
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *vendorModel.Vendor
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*vendor.Vendor"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					vendorPtr := args.Get(1).(*vendorModel.Vendor)
					*vendorPtr = *vendor
				})
			},
			expected: vendor,
			wantErr:  false,
		},
		{
			name: "SUCCESS: Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*vendor.Vendor"),
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
						mock.AnythingOfType("*vendor.Vendor"),
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
			result, err := repo.GetByID(ctx, uuid.NewString())

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestList(t *testing.T) {
	testCases := []struct {
		name           string
		setupMock      func(mockDB *mysqlMocks.IMySqlExt)
		query          *vendorModel.VendorQuery
		expectedResult []*vendorModel.Vendor
		wantErr        bool
	}{
		{
			name: "success find by empty query",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// No args for count query (empty filter)
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 0
					}).
					Return(nil)

				// 2 args for select query (pageSize, offset)
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*vendorModel.Vendor)
						*resultPtr = []*vendorModel.Vendor{}
					}).
					Return(nil)
			},
			query:          &vendorModel.VendorQuery{Page: 1, PageSize: 10},
			expectedResult: []*vendorModel.Vendor{},
			wantErr:        false,
		},
		{
			name: "success find by query with name filter",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// 1 arg for count query (name)
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 1
					}).
					Return(nil)

				// 3 args for select query (name, pageSize, offset)
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*vendorModel.Vendor)
						*resultPtr = []*vendorModel.Vendor{}
					}).
					Return(nil)
			},
			query: &vendorModel.VendorQuery{
				Name:     "test",
				Page:     1,
				PageSize: 10,
			},
			expectedResult: []*vendorModel.Vendor{},
			wantErr:        false,
		},
		{
			name: "success find by query with status filter",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// 1 arg for count query (status)
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 1
					}).
					Return(nil)

				// 3 args for select query (status, pageSize, offset)
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*vendorModel.Vendor)
						*resultPtr = []*vendorModel.Vendor{}
					}).
					Return(nil)
			},
			query: &vendorModel.VendorQuery{
				Status:   "ACTIVE",
				Page:     1,
				PageSize: 10,
			},
			expectedResult: []*vendorModel.Vendor{},
			wantErr:        false,
		},
		{
			name: "failed find by query - count error",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// 1 arg for count query (name)
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("error db"))
			},
			query: &vendorModel.VendorQuery{
				Name:     "test",
				Page:     1,
				PageSize: 10,
			},
			expectedResult: nil,
			wantErr:        true,
		},
		{
			name: "success with pagination - page 1",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// No args for count query (empty filter)
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 100
					}).
					Return(nil)

				// 2 args for select query (pageSize, offset)
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*vendorModel.Vendor)
						*resultPtr = []*vendorModel.Vendor{}
					}).
					Return(nil)
			},
			query: &vendorModel.VendorQuery{
				Page:     1,
				PageSize: 10,
			},
			expectedResult: []*vendorModel.Vendor{},
			wantErr:        false,
		},
		{
			name: "success with pagination - page 2",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// No args for count query (empty filter)
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 100
					}).
					Return(nil)

				// 2 args for select query (pageSize, offset)
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						resultPtr := args.Get(1).(*[]*vendorModel.Vendor)
						*resultPtr = []*vendorModel.Vendor{}
					}).
					Return(nil)
			},
			query: &vendorModel.VendorQuery{
				Page:     2,
				PageSize: 10,
			},
			expectedResult: []*vendorModel.Vendor{},
			wantErr:        false,
		},
		{
			name: "error on SelectContext",
			setupMock: func(mockDB *mysqlMocks.IMySqlExt) {
				// No args for count query (empty filter)
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						totalPtr := args.Get(1).(*int)
						*totalPtr = 10
					}).
					Return(nil)

				// 2 args for select query (pageSize, offset)
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("select error"))
			},
			query: &vendorModel.VendorQuery{
				Page:     1,
				PageSize: 10,
			},
			expectedResult: nil,
			wantErr:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.setupMock(mockDB)

			repository := New(mockLogger, mockDB)
			ctx := context.WithValue(context.Background(), constant.CtxSQLTableNameKey, "vendors")
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

func TestGetByName(t *testing.T) {
	vendor := &vendorModel.Vendor{
		UUID:       "uuid-uuid-uuid",
		MerchantID: "merchant-123",
		Name:       "Test Vendor",
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *vendorModel.Vendor
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*vendor.Vendor"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					vendorPtr := args.Get(1).(*vendorModel.Vendor)
					*vendorPtr = *vendor
				})
			},
			expected: vendor,
			wantErr:  false,
		},
		{
			name: "SUCCESS: Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.
					On(
						"GetContext",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*vendor.Vendor"),
						mock.AnythingOfType("string"),
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
						mock.AnythingOfType("*vendor.Vendor"),
						mock.AnythingOfType("string"),
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
			result, err := repo.GetByName(ctx, "merchant-123", "Test Vendor")

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
