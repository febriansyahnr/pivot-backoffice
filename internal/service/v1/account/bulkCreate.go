package accountService

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *account) BulkCreateAccount(ctx context.Context, request *account_model.BulkCreateAccountRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/BulkCreateAccount")
	defer segment.End()

	err := s.bulkCreateMerchantAccount(ctx, request)
	if err != nil {
		return err
	}

	err = s.bulkCreateCustomerAccount(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (s *account) bulkCreateMerchantAccount(ctx context.Context, request *account_model.BulkCreateAccountRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/BulkCreateAccount")
	defer segment.End()

	for {
		merchants, err := s.accountRepo.GetMerchantsWithoutAccount(ctx, account_model.GetEntityWithoutAccountRequest{
			MerchantID: request.MerchantID,
			Usecase:    request.Usecase,
			Limit:      constant.DefaultPageSize,
		})
		if err != nil {
			return constant.ErrDatabaseGetData
		}
		if merchants == nil {
			break
		}

		accountList := []*account_model.Account{}
		for _, merchant := range merchants {
			merchantUUID, _ := uuid.Parse(merchant.UUID)
			accountReq := &account_model.NewAccountRequest{
				ReferenceID: merchantUUID,
				UserType:    constant.UserTypeSubMerchant,
				Usecase:     request.Usecase,
				Currency:    request.Currency,
			}
			if !merchant.ParentID.Valid || merchant.ParentID.String == "" {
				accountReq.UserType = constant.UserTypeMerchant
			}

			newAccount, err := account_model.NewAccount(accountReq)
			if err != nil {
				s.logger.Error(ctx, "failed create account for submerchant", logger.Error(err), logger.Any("submerchantID", merchantUUID))
				return err
			}
			accountList = append(accountList, newAccount)
		}

		err = s.accountRepo.BulkInsert(ctx, accountList)
		if err != nil {
			return constant.ErrFailedBulkCreateAccounts
		}
	}

	return nil
}

func (s *account) bulkCreateCustomerAccount(ctx context.Context, request *account_model.BulkCreateAccountRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/account/bulkCreateCustomerAccount")
	defer segment.End()

	for {
		customers, err := s.accountRepo.GetCustomersWithoutAccount(ctx, account_model.GetEntityWithoutAccountRequest{
			MerchantID: request.MerchantID,
			Usecase:    request.Usecase,
			Limit:      constant.DefaultPageSize,
		})
		if err != nil {
			return constant.ErrDatabaseGetData
		}
		if customers == nil {
			break
		}

		accountList := []*account_model.Account{}
		for _, customer := range customers {
			customerID, _ := uuid.Parse(customer.UUID)
			accountReq := &account_model.NewAccountRequest{
				ReferenceID: customerID,
				UserType:    constant.UserTypeCustomer,
				Usecase:     request.Usecase,
				Currency:    request.Currency,
			}

			newAccount, err := account_model.NewAccount(accountReq)
			if err != nil {
				s.logger.Error(ctx, "failed create accounts for customers", logger.Error(err), logger.Any("customerID", customerID))
				return err
			}
			accountList = append(accountList, newAccount)
		}

		err = s.accountRepo.BulkInsert(ctx, accountList)
		if err != nil {
			return constant.ErrFailedBulkCreateAccounts
		}
	}

	return nil
}
