package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/paper-indonesia/pivot-backoffice/config"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	userRepoMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerPkgMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiY2U5Mzc1OGUtYjE5Yi00M2MyLTk3MmEtZjY0YmUxNmYwOGNmIiwiaWRlbnRpZmllciI6IjMxYjNkMDk5LTg1MDgtNGU5My1hODBjLTA0NzFlZjIzN2E0YyJ9.awoAo8i2uIvROrmxDqS1aa45N4OOPJU7KHuiTwFaBls"

func TestForgotPassword(pt *testing.T) {

	otpSvcMock := serviceMocks.NewIOTP(pt)
	logMock, _ := loggerPkgMock.NewZapLogger(loggerPkgMock.Config{})
	userMock := userRepoMock.NewIUserRepository(pt)

	service := New(
		&config.Config{}, nil, logMock, userMock, nil, WithOTPService(otpSvcMock),
	)

	tests := []struct {
		name      string
		mockSetup func()
		wantErr   string
		wantToken string
	}{
		{
			name: "ERROR:Find user by email",
			mockSetup: func() {
				userMock.On("FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType()).Once().Return(nil, assert.AnError)
			},
			wantErr: assert.AnError.Error(),
		},
		{
			name: "ERROR:Email not registered",
			mockSetup: func() {
				userMock.On(
					"FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: "email not registered",
		},
		{
			name: "ERROR:User has been blocked",
			mockSetup: func() {
				userMock.On(
					"FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(&userModel.User{
					Blocked: sql.NullTime{Time: time.Now(), Valid: true},
				}, nil)
			},
			wantErr: "user has been blocked",
		},
		{
			name: "ERROR:User has not completed onboarding",
			mockSetup: func() {
				userMock.On(
					"FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(&userModel.User{
					Status: constant.UserStatusInvited,
				}, nil)
			},
			wantErr: constant.ErrUserInvitedStatus.Error(),
		},
		{
			name: "ERROR:Merchant is blocked",
			mockSetup: func() {
				userMock.On(
					"FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(&userModel.User{
					MerchantStatus: sql.NullString{Valid: true, String: constant.MerchantStatusBlocked},
				}, nil)
			},
			wantErr: "merchant is blocked", // NOSONAR
		},
		{
			name: "ERROR:Merchant is inactive",
			mockSetup: func() {
				userMock.On(
					"FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(&userModel.User{
					MerchantStatus: sql.NullString{Valid: true, String: constant.MerchantStatusInactive},
				}, nil)
			},
			wantErr: "merchant is inactive", // NOSONAR
		},
		{
			name: "ERROR:Merchant is deactivated",
			mockSetup: func() {
				userMock.On(
					"FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(&userModel.User{
					MerchantStatus: sql.NullString{Valid: true, String: constant.MerchantStatusDeactivated},
				}, nil)
			},
			wantErr: "merchant is deactivated", // NOSONAR
		},
		{
			name: "ERROR:Merchant is closed",
			mockSetup: func() {
				userMock.On(
					"FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(&userModel.User{
					MerchantStatus: sql.NullString{Valid: true, String: constant.MerchantStatusClosed},
				}, nil)
			},
			wantErr: "Merchant status is closed. Reason: ", // NOSONAR
		},
		{
			name: "ERROR:Generate OTP code",
			mockSetup: func() {
				userMock.On(
					"FindUserByEmail", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&userModel.User{}, nil)

				otpSvcMock.On(
					"GenerateOTPCode", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Once().Return("", errors.New("your request has exceeded the limit"))
			},
			wantErr: "your request has exceeded the limit",
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				otpSvcMock.On(
					"GenerateOTPCode", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.AnythingOfType("constant.OTPIdentifier"),
				).Return(token, nil)
			},
			wantToken: token,
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			test.mockSetup()

			token, err := service.ForgotPassword(context.Background(), "email@example.id")
			if test.wantErr == "" {
				require.Nil(t, err)
				assert.Equal(t, test.wantToken, token)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestResetPassword(pt *testing.T) {
	userMock := userRepoMock.NewIUserRepository(pt)

	service := New(&config.Config{}, nil, nil, userMock, nil)

	tests := []struct {
		name      string
		mockSetup func(u *userRepoMock.IUserRepository)
		wantErr   string
	}{
		{
			name: "ERROR:Change password",
			mockSetup: func(u *userRepoMock.IUserRepository) {
				u.On(
					"ChangePassword", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Once().Return(false, errors.New("invalid db session"))
			},
			wantErr: "invalid db session",
		},
		{
			name: "SUCCESS",
			mockSetup: func(u *userRepoMock.IUserRepository) {
				u.On(
					"ChangePassword", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(true, nil)
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			test.mockSetup(userMock)

			err := service.ResetPassword(context.Background(), "unique-id", "text-password")
			if test.wantErr == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
