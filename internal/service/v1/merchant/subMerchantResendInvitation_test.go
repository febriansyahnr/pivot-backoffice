package merchant_test

import (
	"context"
	"database/sql"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	userMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubMerchantResendInvitation(t *testing.T) {

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	userSvc := userMocks.NewIUserService(t)
	repo := repoMocks.NewIMerchantRepository(t)

	service := New(repo, logger, nil, nil, nil, nil, WithUserService(userSvc))

	subMerchant := uuid.NewString()
	parentMerchant := uuid.NewString()
	request := &merchant.ResendInvitationRequest{
		Email:            "hero@email.com",
		MerchantId:       subMerchant,
		ParentMerchantId: parentMerchant,
	}
	tests := []struct {
		name      string
		request   *merchant.ResendInvitationRequest
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Find merchant by id",
			setupMock: func() {
				repo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name: "ERROR:Merchant not found",
			setupMock: func() {
				repo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrNotFound, c.ErrMerchantNotFound), // NOSONAR
		},
		{
			name: "ERROR:Merchant not allowed",
			setupMock: func() {
				repo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&merchant.Merchant{ParentID: sql.NullString{String: "ABC"}}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrForbidden, c.ErrMerchantNotAllowedPerformAction), // NOSONAR
		},
		{
			name:    "ERROR:Find user by email",
			request: request,
			setupMock: func() {
				repo.On(
					"FindMerchantByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&merchant.Merchant{
					UUID:     subMerchant,
					ParentID: sql.NullString{String: parentMerchant},
				}, nil)

				userSvc.On(
					"FindUserByEmail", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest), // NOSONAR
		},
		{
			name:    "ERROR:User with email not found",
			request: request,
			setupMock: func() {
				userSvc.On(
					"FindUserByEmail", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrNotFound, c.ErrUserNotFound), // NOSONAR
		},
		{
			name:    "ERROR:User not allowed",
			request: request,
			setupMock: func() {
				userSvc.On(
					"FindUserByEmail", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&user.User{}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrForbidden, c.ErrMerchantNotAllowedPerformAction), // NOSONAR
		},
		{
			name:    "ERROR:User already activated",
			request: request,
			setupMock: func() {
				userSvc.On(
					"FindUserByEmail", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(&user.User{MerchantId: subMerchant, Status: c.UserStatusActive}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrUserAlreadyActivated), // NOSONAR
		},
		{
			name:    "ERROR:Send generated invitation URL",
			request: request,
			setupMock: func() {
				userSvc.On(
					"FindUserByEmail", c.ValueCtxMockType(), c.StringMockType(),
				).Return(&user.User{MerchantId: subMerchant, Status: c.UserStatusInvited}, nil)

				userSvc.On(
					"SendGeneratedInvitationURL", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest, // NOSONAR
		},
		{
			name:    "SUCCESS",
			request: request,
			setupMock: func() {
				userSvc.On("SendGeneratedInvitationURL", c.ValueCtxMockType(), mock.Anything).Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			if test.request == nil {
				test.request = &merchant.ResendInvitationRequest{}
			}

			assert.Equal(t, test.wantErr, service.SubMerchantResendInvitation(context.Background(), test.request))
		})
	}
}
