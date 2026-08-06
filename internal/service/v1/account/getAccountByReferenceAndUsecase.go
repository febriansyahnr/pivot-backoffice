package accountService

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *account) GetAccountByReferenceIDAndUsecase(ctx context.Context, referenceID uuid.UUID, usecase string, userType string) (*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/GetAccountByReferenceIDAndUsecase")
	defer segment.End()

	// Get Account By UUID
	accountName := account_model.GetAccountNameByUsecase(usecase)
	account, err := s.accountRepo.GetByReferenceIDAndUsecase(ctx, referenceID, accountName, userType)
	if err != nil {
		s.logger.Error(ctx, "error when get account by reference id and use case", logger.Error(err))
		return nil, err
	} else if account == nil {
		if !slices.Contains([]string{constant.TypePayment, constant.TypeDisbursement}, usecase) {
			return nil, nil
		}

		// Create new account if usecase allowed.
		account, err = account_model.NewAccount(&account_model.NewAccountRequest{
			ReferenceID: referenceID,
			Usecase:     accountName,
			Currency:    constant.CurrencyIDR,
			UserType:    userType,
		})
		if err != nil {
			return nil, pkgErrors.New(httpResponse.HttpErrRequest, err)
		}

		if err = s.accountRepo.Create(ctx, account); err != nil {
			return nil, err
		}
	}

	return account, nil
}
