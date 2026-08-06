package role

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const InvalidSession = "Invalid Session"

func TestRoleRepository_GetList(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(db *mysqlMocks.IMySqlExt)
		page      int64
		perPage   int64
		wantErr   bool
	}{
		{
			name: "Success",
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				// Mock SelectContext
				db.On("SelectContext", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(nil).Run(func(args mock.Arguments) {
					dataPtr := args.Get(1).(*[]*role.Role)
					*dataPtr = []*role.Role{{UUID: "1", Name: "Role1", Permissions: sql.NullString{String: "Permission1,Permission2", Valid: true}}}
				})

				// Mock GetContext
				db.On("GetContext", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(nil).Run(func(args mock.Arguments) {
					totalItemsPtr := args.Get(1).(*int64)
					*totalItemsPtr = 1
				})
			},
			page:    1,
			perPage: 10,
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*role.Role"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock DB and Logger
			mockDB := &mysqlMocks.IMySqlExt{}
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Set up mocks
			tc.mockSetup(mockDB)

			// Create RoleRepository instance
			repo := &RoleRepository{
				db:     mockDB,
				logger: mockLogger,
			}

			// Call the method
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "roles")
			resp, err := repo.GetList(ctx, &role.GetRoleFilterRequest{MerchantID: uuid.NewString()}, tc.page, tc.perPage)

			// Assertions
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				respData, ok := resp.Data.([]*role.RoleResponse)
				assert.True(t, ok, "response data type assertion failed")
				assert.NotNil(t, resp)
				assert.Equal(t, int64(1), resp.Meta.TotalItems)
				assert.Equal(t, int64(1), resp.Meta.TotalPages)
				assert.Len(t, resp.Data, 1)
				assert.Equal(t, "Role1", respData[0].Name)
				assert.ElementsMatch(t, []string{"Permission1", "Permission2"}, respData[0].Permissions)
			}

			// Assert DB method invocations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestRoleRepository_FindRoleByID(t *testing.T) {
	now := time.Now()

	existedRole := &role.Role{
		UUID:      "uuid-uuid-uuid",
		Name:      "role-name",
		Slug:      "role-slug",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		name        string
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		input       string
		expected    *role.Role
		expectedErr string
		wantErr     bool
	}{
		{
			name: "SUCCESS: Find Role By ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*role.Role"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					rolePtr := args.Get(1).(*role.Role)
					*rolePtr = *existedRole
				})
			},
			input:       existedRole.UUID,
			expected:    existedRole,
			expectedErr: "",
			wantErr:     false,
		},
		{
			name: "ERROR: Role Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*role.Role"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			input:    existedRole.UUID,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*role.Role"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			input:       existedRole.UUID,
			expected:    nil,
			expectedErr: "database error",
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := New(mysqlMock, loggerMock)

			ctx := context.Background()

			roleRes, err := repo.FindRoleByID(ctx, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, roleRes)
				require.True(t, strings.Contains(err.Error(), tc.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, roleRes)
			}
		})
	}
}

func TestRoleRepository_FindRoleBySlug(t *testing.T) {
	now := time.Now()

	existedRole := &role.Role{
		UUID:      "uuid-uuid-uuid",
		Name:      "role-name",
		Slug:      "role-slug",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		name        string
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		input       string
		expected    *role.Role
		expectedErr string
		wantErr     bool
	}{
		{
			name: "SUCCESS: Find Role By Slug",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*role.Role"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					rolePtr := args.Get(1).(*role.Role)
					*rolePtr = *existedRole
				})
			},
			input:       existedRole.Slug,
			expected:    existedRole,
			expectedErr: "",
			wantErr:     false,
		},
		{
			name: "ERROR: Role Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*role.Role"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			input:    existedRole.UUID,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*role.Role"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			input:       existedRole.UUID,
			expected:    nil,
			expectedErr: "database error",
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := New(mysqlMock, loggerMock)

			ctx := context.Background()

			roleRes, err := repo.FindRoleBySlug(ctx, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, roleRes)
				require.True(t, strings.Contains(err.Error(), tc.expectedErr))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, roleRes)
			}
		})
	}
}

func TestTotalRoleByMerchantID(t *testing.T) {
	logMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mysqlMock := mysqlMocks.NewIMySqlExt(t)

	repo := New(mysqlMock, logMock)

	tests := []struct {
		name         string
		mockModifier func()
		wantErr      string
	}{
		{
			name: "ERROR:Invalid session",
			mockModifier: func() {
				mysqlMock.On(
					"GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*uint64"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(errors.New(InvalidSession))
			},
			wantErr: InvalidSession,
		},
		{
			name: "SUCCESS",
			mockModifier: func() {
				mysqlMock.On(
					"GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*uint64"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockModifier()

			if _, err := repo.TotalRoleByMerchantID(context.Background(), "unique-id"); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestCheckAvailableRoleName(t *testing.T) {
	logMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mysqlMock := mysqlMocks.NewIMySqlExt(t)

	repo := New(mysqlMock, logMock)

	tests := []struct {
		name         string
		mockModifier func()
		wantErr      string
	}{
		{
			name: "ERROR:Invalid session",
			mockModifier: func() {
				mysqlMock.On(
					"GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*bool"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(errors.New(InvalidSession))
			},
			wantErr: InvalidSession,
		},
		{
			name: "SUCCESS",
			mockModifier: func() {
				mysqlMock.On(
					"GetContext", constant.ValueCtxMockType(), mock.AnythingOfType("*bool"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockModifier()

			if _, err := repo.CheckAvailableRoleName(context.Background(), "id", "name"); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
