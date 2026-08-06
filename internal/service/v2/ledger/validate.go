package ledgerService

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *LedgerService) ValidateTransaction(ctx context.Context, merchantID string, request *ledger_model.CreateNewLedgerEntryRequest) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v2/ledger/ValidateTransaction")
	defer segment.End()

	var (
		accountList = []*account_model.Account{}
	)

	if err := request.Validate(); err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	if request.RecipientAccountID != uuid.Nil {
		recipientAccount, err := s.accountRepo.GetByUUID(ctx, request.RecipientAccountID)
		if err != nil {
			return pkgErrors.New(response.HttpErrInternal, errors.New("error find recipient account"))
		}
		if recipientAccount == nil {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrRecipientAccountNotFound)
		}
		request.RecipientID = recipientAccount.ReferenceID
		if recipientAccount.UserType == constant.UserTypeCustomer {
			customer, err := s.customerSvc.FindCustomerByID(ctx, recipientAccount.ReferenceID.String())
			if err != nil {
				return err
			}
			request.RecipientID = uuid.MustParse(customer.MerchantID)
		}
		accountList = append(accountList, recipientAccount)
	}

	if request.SenderAccountID != uuid.Nil {
		senderAccount, err := s.accountRepo.GetByUUID(ctx, request.SenderAccountID)
		if err != nil {
			return pkgErrors.New(response.HttpErrInternal, errors.New("error find sender account"))
		}
		if senderAccount == nil {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrSenderAccountNotFound)
		}
		request.SenderID = senderAccount.ReferenceID
		if senderAccount.UserType == constant.UserTypeCustomer {
			customer, err := s.customerSvc.FindCustomerByID(ctx, senderAccount.ReferenceID.String())
			if err != nil {
				return err
			}
			request.SenderID = uuid.MustParse(customer.MerchantID)
		}
		accountList = append(accountList, senderAccount)
	}

	if request.ParentAccountID != uuid.Nil {
		parentAccount, err := s.accountRepo.GetByUUID(ctx, request.ParentAccountID)
		if err != nil {
			return pkgErrors.New(response.HttpErrInternal, errors.New("error find parent account"))
		}
		if parentAccount == nil {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrParentAccountNotFound)
		}

		request.ParentID = parentAccount.ReferenceID
		accountList = append(accountList, parentAccount)
	}

	if request.Fee.RecipientAccountID != uuid.Nil {
		feeRecipientAccount, err := s.accountRepo.GetByUUID(ctx, request.Fee.RecipientAccountID)
		if err != nil {
			return pkgErrors.New(response.HttpErrInternal, errors.New("error find recipient fee account"))
		}
		if feeRecipientAccount == nil {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrFeeRecipientAccountNotFound)
		}
		if feeRecipientAccount.UserType == constant.UserTypeCustomer {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrFeeRecipientIsNotMerchant)
		}
		request.Fee.RecipientID = feeRecipientAccount.ReferenceID
		accountList = append(accountList, feeRecipientAccount)
	}

	if len(accountList) == 0 {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrAccountNotSpecified)
	}

	return nil
}
