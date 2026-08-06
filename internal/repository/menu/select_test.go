package menuRepository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAll(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(db *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "Success",
			mockSetup: func(db *mysqlMocks.IMySqlExt) {
				// Mock SelectContext
				db.On("SelectContext", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.Anything).
					Return(nil).Run(func(args mock.Arguments) {
					dataPtr := args.Get(1).(*[]*menuModel.Menu)
					*dataPtr = []*menuModel.Menu{{UUID: "1", Name: "Role1", Permissions: sql.NullString{String: "group:slug,group:slug", Valid: true}}}
				})
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*menuModel.Menu"),
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
			repo := New(mockDB, mockLogger)

			// Call the method
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "menus")
			_, err := repo.GetAll(ctx, &menuModel.GetAllFilterRequest{RoleID: uuid.NewString()})

			// Assertions
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Assert DB method invocations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestRoleRepository_FindRoleBySlug(t *testing.T) {
	now := time.Now()

	existedMenu := &menuModel.Menu{
		UUID:      "uuid-uuid-uuid",
		Name:      "name",
		Slug:      "slug",
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
					mock.AnythingOfType("*menuModel.Menu"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					menuPtr := args.Get(1).(*menuModel.Menu)
					*menuPtr = *existedMenu
				})
			},
			expectedErr: "",
			wantErr:     false,
		},
		{
			name: "ERROR: Menu Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*menuModel.Menu"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*menuModel.Menu"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
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

			_, err := repo.FindBySlug(ctx, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tc.expectedErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindBySlugWithPermissions(t *testing.T) {
	now := time.Now()

	existedMenuWithPermissions := &menuModel.Menu{
		UUID:      "uuid-uuid-uuid",
		Name:      "Home",
		Slug:      "home",
		CreatedAt: now,
		UpdatedAt: now,
		Permissions: sql.NullString{
			String: "Home:home.view",
			Valid:  true,
		},
	}

	testCases := []struct {
		name        string
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		input       string
		expectedErr string
		wantErr     bool
	}{
		{
			name: "SUCCESS: Find Menu With Permissions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*menuModel.Menu"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					menuPtr := args.Get(1).(*menuModel.Menu)
					*menuPtr = *existedMenuWithPermissions
				})
			},
			input:       "home",
			expectedErr: "",
			wantErr:     false,
		},
		{
			name: "ERROR: Menu Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*menuModel.Menu"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(sql.ErrNoRows)
			},
			input:   "nonexistent",
			wantErr: false,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*menuModel.Menu"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(errors.New("database error"))
			},
			input:       "home",
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

			result, err := repo.FindBySlugWithPermissions(ctx, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tc.expectedErr))
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if result != nil {
					assert.NotEmpty(t, result.UUID)
					assert.NotEmpty(t, result.Permissions)
				}
			}
		})
	}
}

func TestGetMenuAndPermissionIDs(t *testing.T) {

	logMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mysqlMock := mysqlMocks.NewIMySqlExt(t)

	repo := New(mysqlMock, logMock)

	tests := []struct {
		name         string
		mockModifier func()
		wantErr      string
	}{
		{
			name: "ERROR:Get context/Invalid session",
			mockModifier: func() {
				mysqlMock.On(
					"GetContext", constant.ValueCtxMockType(), constant.PtrMenuAndPermissionIDsMockType(), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(errors.New("Invalid Sesssion"))
			},
			wantErr: "Invalid Sesssion",
		},
		{
			name: "ERROR:Get context/Not row result",
			mockModifier: func() {
				mysqlMock.On(
					"GetContext", constant.ValueCtxMockType(), constant.PtrMenuAndPermissionIDsMockType(), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			mockModifier: func() {
				mysqlMock.On(
					"GetContext", constant.ValueCtxMockType(), constant.PtrMenuAndPermissionIDsMockType(), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
				).Once().Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockModifier()

			if _, err := repo.GetMenuAndPermissionIDs(context.Background(), ""); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
