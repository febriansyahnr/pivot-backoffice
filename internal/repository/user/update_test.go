package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Update(t *testing.T) {
	expectedUser := &userModel.User{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "pass",
		PinHash:    sql.NullString{Valid: true, String: "pin"},
		Blocked:    sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  time.Now(),
	}

	// Define the test cases
	testCases := []struct {
		name      string
		inputUser *userModel.User
		result    error
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr:   false,
			inputUser: expectedUser,
			result:    nil,
		},
		{
			name: "FAILED: Error when updating blocked status in user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, errors.New("error database"))
			},
			wantErr:   true,
			inputUser: nil,
			result:    errors.New("error database"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "users")
			err := repo.Update(ctx, user)

			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.UpdateBlocked() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserRepository_UpdateRefreshToken(t *testing.T) {
	// Define the test cases
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		userId    string
		token     string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update refresh token in user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
			wantErr: false,
			userId:  user.UUID,
			token:   "refresh_token",
		},
		{
			name: "FAILED: Error when updating refresh token in user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(false, errors.New("error database"))
			},
			wantErr: true,
			userId:  user.UUID,
			token:   "refresh_token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "users")
			err := repo.UpdateRefreshToken(ctx, tc.userId, tc.token)

			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.UpdateRefreshToken() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserRepository_ChangePassword(t *testing.T) {
	// Define the test cases
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		userId    string
		password  string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update change password in user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
			wantErr:  false,
			userId:   user.UUID,
			password: "password",
		},
		{
			name: "ERROR: Update change password in user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(false, errors.New("some error"))

			},
			wantErr:  true,
			userId:   user.UUID,
			password: "password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "users")
			_, err := repo.ChangePassword(ctx, tc.userId, tc.password)

			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.ChangePassword() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestBlockUser(pt *testing.T) {
	mockMysql := mysqlMocks.NewIMySqlExt(pt)

	repo := New(mockMysql, nil)

	tests := []struct {
		name      string
		id        string
		blocked   sql.NullTime
		mockSetup func(m *mysqlMocks.IMySqlExt)
		wantErr   string
	}{
		{
			name: "ERROR:Invalid session",
			mockSetup: func(m *mysqlMocks.IMySqlExt) {
				m.On(
					"ExecContext", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("sql.NullTime"), mock.AnythingOfType("string"),
				).Once().Return(false, errors.New("invalid session"))
			},
			wantErr: "invalid session",
		},
		{
			name: "SUCCESS",
			mockSetup: func(m *mysqlMocks.IMySqlExt) {
				m.On(
					"ExecContext", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.AnythingOfType("string"), mock.AnythingOfType("sql.NullTime"), mock.AnythingOfType("string"),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			test.mockSetup(mockMysql)

			err := repo.BlockUser(context.Background(), test.id, test.blocked)
			if test.wantErr == "" {
				assert.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestUpdateUserTOTPData(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		request   *userModel.UpdateUserTOTPDataRequest
		setupMock func()
		wantError error
	}{
		{
			name:      "ERROR:Nil request parameter",
			setupMock: func() { /* Empty */ },
			wantError: errors.New("request parameters can't be nil"),
		},
		{
			name:    "ERROR:Some error", // NOSONAR
			request: &userModel.UpdateUserTOTPDataRequest{},
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			request: &userModel.UpdateUserTOTPDataRequest{
				WrappedSecret: "secret-key", // NOSONAR
			},
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, "UPDATE users SET updated_at = :updated_at, totp_wrapped_secret = :totp_wrapped_secret WHERE uuid = :uuid;", mock.Anything).Once().Return(false, nil)
			},
			wantError: constant.ErrUserNotFound,
		},
		{
			name: "SUCCESS", // NOSONAR
			request: &userModel.UpdateUserTOTPDataRequest{
				WrappedSecret:  "secret-key", // NOSONAR
				EncryptVersion: 1,
				Status:         constant.TOTPStatusEnrolled,
			},
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, "UPDATE users SET updated_at = :updated_at, totp_wrapped_secret = :totp_wrapped_secret, totp_encrypt_version = :totp_encrypt_version, totp_status = :totp_status WHERE uuid = :uuid;", mock.Anything).Once().Return(false, nil)
			},
			wantError: constant.ErrUserNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.UpdateUserTOTPData(t.Context(), test.request))
		})
	}
}

func TestUpdateUserPreferred2FAMethod(t *testing.T) {
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, mockLogger)

	tests := []struct {
		name               string
		userId             string
		preferred2FAMethod string
		setupMock          func()
		wantError          error
	}{
		{
			name:               "SUCCESS: Update preferred 2FA method",
			userId:             "user-uuid-123",
			preferred2FAMethod: "OTP",
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).Once().Return(true, nil)
			},
			wantError: nil,
		},
		{
			name:               "ERROR: Database error",
			userId:             "user-uuid-123",
			preferred2FAMethod: "TOTP",
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).Once().Return(false, errors.New("database error"))
			},
			wantError: errors.New("database error"),
		},
		{
			name:               "ERROR: User not found",
			userId:             "non-existent-user",
			preferred2FAMethod: "OTP",
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).Once().Return(false, nil)
			},
			wantError: constant.ErrUserNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.UpdateUserPreferred2FAMethod(context.Background(), test.userId, test.preferred2FAMethod)

			if test.wantError != nil {
				require.NotNil(t, err)
				assert.Equal(t, test.wantError.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
