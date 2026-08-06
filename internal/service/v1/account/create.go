package accountService

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *account) CreateMerchantAccount(ctx context.Context, referenceId, userType string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/CreateAccount")
	defer segment.End()

	referenceUUID, _ := uuid.Parse(referenceId)
	request := &account_model.NewAccountRequest{
		ReferenceID: referenceUUID,
		UserType:    userType,
		Usecase:     constant.TypeDisbursement,
		Currency:    "IDR",
	}
	newAccount, err := account_model.NewAccount(request)
	if err != nil {
		s.logger.Error(ctx, "failed create account", logger.Error(err), logger.Any("request", request))
		return pkgErrors.New(response.HttpErrRequest, err)
	}
	err = s.accountRepo.Create(ctx, newAccount)
	if err != nil {
		s.logger.Error(ctx, "error when create account", logger.Error(err), logger.Any("request", newAccount))
		return pkgErrors.New(response.HttpErrInternal, constant.ErrFailedCreateMerchantAccount)
	}

	return nil
}

func (s *account) CreateAccount(ctx context.Context, request *account_model.NewAccountRequest) (*account_model.AccountResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/CreateAccount")
	defer segment.End()

	newAccount, err := account_model.NewAccount(request)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}
	err = s.accountRepo.Create(ctx, newAccount)
	if err != nil {
		s.logger.Error(ctx, "error when create account", logger.Error(err), logger.Any("request", newAccount))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrFailedCreateAccount)
	}

	return newAccount.ToResponse(), nil
}
