package userRole

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserRoleRepository_FindUserRoleByUserID(t *testing.T) {
	now := time.Now()

	existedRole := &userRole.UserRole{
		UUID:      "uuid-uuid-uuid",
		UserID:    "user-id",
		RoleID:    "role-id",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		name        string
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		input       string
		expected    *userRole.UserRole
		expectedErr string
		wantErr     bool
	}{
		{
			name: "SUCCESS: find user_role By user_id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*userRole.UserRole"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					rolePtr := args.Get(1).(*userRole.UserRole)
					*rolePtr = *existedRole
				})
			},
			input:       existedRole.UUID,
			expected:    existedRole,
			expectedErr: "",
			wantErr:     false,
		},
		{
			name: "ERROR: user_role not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*userRole.UserRole"),
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
					mock.AnythingOfType("*userRole.UserRole"),
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

			roleRes, err := repo.FindUserRoleByUserID(ctx, tc.input)

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

func TestTotalActiveUsersByRoleID(t *testing.T) {
	mysqlMock := mysqlMocks.NewIMySqlExt(t)

	repo := New(mysqlMock, nil)

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
				).Once().Return(errors.New("invalid session"))
			},
			wantErr: "invalid session",
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

			if _, err := repo.TotalActiveUsersByRoleID(context.Background(), "role-id"); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
