package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserRepository_ListUsers(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     int
		expected  []*userModel.User
		wantErr   bool
	}{
		{
			name: "SUCCESS: List Users",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int"),
				).Return(nil).Run(func(args mock.Arguments) {
					usersPtr := args.Get(1).(*[]*userModel.User)
					*usersPtr = []*userModel.User{user}
				})
			},
			input:    10,
			expected: []*userModel.User{user},
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("int"),
					mock.AnythingOfType("int"),
				).Return(errors.New("database error"))
			},
			input:    10,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "users")
			users, err := repo.ListUsers(ctx, tc.input, 0)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, users)
			}
		})
	}
}

func TestUserRepository_ListUsersByMerchantID(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *userModel.ListUsersByMerchantIDRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: List Users",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					usersPtr := args.Get(1).(*[]*userModel.User)
					*usersPtr = []*userModel.User{user}
				})

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter:  &userModel.ListUsersByMerchantIDRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS: List Users With Sort",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					usersPtr := args.Get(1).(*[]*userModel.User)
					*usersPtr = []*userModel.User{user}
				})

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter: &userModel.ListUsersByMerchantIDRequest{
				SortBy:    constant.UserSortColName,
				SortOrder: constant.SortOrderAsc,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get user list without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(errors.New("no rows data"))
			},
			filter:  &userModel.ListUsersByMerchantIDRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get list user with filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("no rows data"))
			},
			filter: &userModel.ListUsersByMerchantIDRequest{
				MerchantID:     uuid.NewString(),
				StartCreatedAt: &util.TimeNow,
				EndCreatedAt:   &util.TimeNow,
				Name:           "tes",
				RoleID:         uuid.NewString(),
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter:  &userModel.ListUsersByMerchantIDRequest{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "users")
			_, err := repo.ListUsersByMerchantID(ctx, tc.filter, 0, 20)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestUserRepository_FindUserByID(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		expected  *userModel.User
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find User by ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					userPtr := args.Get(1).(*userModel.User)
					*userPtr = *user
				})
			},
			input:    user.UUID,
			expected: user,
			wantErr:  false,
		},
		{
			name: "ERROR: User Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			input:    user.UUID,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			input:    user.UUID,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "user")
			userRes, err := repo.FindUserByID(ctx, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, userRes)
			}
		})
	}
}

func TestUserRepository_FindUserByEmail(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		expected  *userModel.User
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find User by Email",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					userPtr := args.Get(1).(*userModel.User)
					*userPtr = *user
				})
			},
			input:    user.UUID,
			expected: user,
			wantErr:  false,
		},
		{
			name: "ERROR: User Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			input:    user.UUID,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*user.User"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			input:    user.UUID,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "user")
			userRes, err := repo.FindUserByEmail(ctx, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, userRes)
			}
		})
	}
}

func TestFindUserTOTPDataByID(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *userModel.UserTOTPData
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(sql.ErrNoRows)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil)
			},
			wantError:  nil,
			wantResult: &userModel.UserTOTPData{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.FindUserTOTPDataByID(t.Context(), "")
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
