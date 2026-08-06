package accountService

import (
	"context"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (a *account) GetMerchantAccounts(ctx context.Context, merchantIDs []uuid.UUID, usecase string) (map[uuid.UUID]*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/GetMerchantAccounts")
	defer segment.End()

	if len(merchantIDs) == 0 {
		return nil, nil
	}

	accountName := account_model.GetAccountNameByUsecase(usecase)
	accountMap, err := a.accountRepo.GetEntityAccounts(ctx, merchantIDs, constant.UserTypeMerchant, accountName)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrInternal, constant.ErrGetAccounts)
	}

	for _, merchantID := range merchantIDs {
		if _, exists := accountMap[merchantID]; !exists {
			acc, err := account_model.NewAccount(&account_model.NewAccountRequest{
				ReferenceID: merchantID,
				UserType:    constant.UserTypeMerchant,
				Usecase:     accountName,
				Currency:    constant.CurrencyIDR,
			})
			if err != nil {
				return nil, pkgErr.New(response.HttpErrRequest, err)
			}

			if err := a.accountRepo.Create(ctx, acc); err != nil {
				return nil, pkgErr.New(response.HttpErrDatabase, err)
			}

			// Add the newly created account to the map
			accountMap[merchantID] = acc
		}
	}

	return accountMap, nil
}

func (a *account) GetWalletMerchantAccount(ctx context.Context, parentMerchantId, merchantID uuid.UUID) (*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/GetWalletMerchantAccount")
	defer segment.End()

	if merchantID != parentMerchantId {
		err := a.merchantSvc.ValidateSubMerchantParent(ctx, parentMerchantId.String(), merchantID.String())
		if err != nil {
			return nil, err
		}
	}

	account, err := a.GetAccountByReferenceIDAndUsecase(ctx, merchantID, constant.ReferenceWallet, constant.UserTypeMerchant)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrInternal, constant.ErrGetAccount)
	}
	if account == nil {
		return nil, pkgErr.New(response.HttpErrNotFound, constant.ErrAccountNotFound)
	}

	return account, nil
}
