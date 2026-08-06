package orchestrator_service

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (os *OrchestratorService) CreateAccountTransaction(ctx context.Context, request *orchestrator_model.CreateAccountTransactionRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/CreateAccountTransaction")
	defer segment.End()

	if ctx, err = os.accountTransactionRepo.BeginTransaction(ctx); err != nil {
		return errors.New(responseHttp.HttpErrDatabase, err)
	}

	if err := os.PostAccountTransaction(ctx, request); err != nil {
		return errors.New(responseHttp.HttpErrDatabase, err)
	}

	if err := os.accountTransactionRepo.CommitTransaction(ctx); err != nil {
		return errors.New(responseHttp.HttpErrDatabase, err)
	}
	return nil
}

func (os *OrchestratorService) PostAccountTransaction(ctx context.Context, request *orchestrator_model.CreateAccountTransactionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/postAccountTransaction")
	defer segment.End()

	name := account_model.GetAccountNameByUsecase(request.Usecase)
	account, err := os.accountRepo.FindMerchantAccountByName(ctx, request.MerchantID, name)
	if err != nil {
		return errors.New(responseHttp.HttpErrDatabase, err)
	} else if account == nil {
		account, err = account_model.NewAccount(&account_model.NewAccountRequest{
			ReferenceID: request.MerchantID,
			UserType:    constant.UserTypeMerchant,
			Usecase:     name,
			Currency:    request.Currency,
		})
		if err != nil {
			return errors.New(responseHttp.HttpErrRequest, err)
		}

		if err = os.accountRepo.Create(ctx, account); err != nil {
			return errors.New(responseHttp.HttpErrDatabase, err)
		}
	}
	return os.accountTransactionRepo.Create(ctx, request.ToAccountTransactionDTO(account))
}
