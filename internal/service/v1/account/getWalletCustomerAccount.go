package accountService

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *account) GetWalletCustomerAccount(ctx context.Context, req *account_model.GetCustomerAccountRequest) (*account_model.Account, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/GetWalletCustomerAccount")
	defer segment.End()

	_, err := s.customerSvc.GetCustomerById(ctx, req.CustomerID, req.MerchantID)
	if err != nil {
		return nil, err
	}
	custId := uuid.MustParse(req.CustomerID)

	account, err := s.GetAccountByReferenceIDAndUsecase(ctx, custId, constant.TypeWallet, constant.UserTypeCustomer)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrInternalServerForUser)
	}
	if account == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrAccountNotFound)
	}

	return account, nil
}
