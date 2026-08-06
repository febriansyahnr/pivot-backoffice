package customerService

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *CustomerService) CreateCustomer(ctx context.Context, request customerModel.CreateCustomerRequest) (*customerModel.GeneralCustomerResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/CreateCustomer")
	defer segment.End()

	customer, err := c.customerRepo.GetCustomerByPhoneNumber(ctx, request.PhoneNumber, request.MerchantID)

	if err != nil {
		c.logger.Error(ctx, "Error get customer by phone number", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}
	if customer != nil {
		return nil, errors.New(response.HttpErrRequest, constant.ErrCustomerAlreadyExists)
	}

	newCustomer := customerModel.CreateCustomer(&request)
	err = c.customerRepo.Create(ctx, newCustomer)
	if err != nil {
		c.logger.Error(ctx, "Error when create customer", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	err = c.createCustomerWallet(ctx, newCustomer.UUID)
	if err != nil {
		c.logger.Error(ctx, "Error when create customer wallet", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, err)
	}

	return newCustomer.ToGeneralCustomerResponse(), nil
}

func (c *CustomerService) createCustomerWallet(ctx context.Context, customerId string) error {
	accountRequest := account_model.NewAccountRequest{
		ReferenceID: uuid.MustParse(customerId),
		UserType:    constant.UserTypeCustomer,
		Usecase:     constant.ReferenceWallet,
		Currency:    constant.CurrencyIDR,
	}
	_, err := c.accountService.CreateAccount(ctx, &accountRequest)
	if err != nil {
		return err
	}
	return nil
}

func (c *CustomerService) CreateUnfiedPaymentCustomer(ctx context.Context, request customerModel.CreateUnifiedPaymentCustomerRequest) (*customerModel.GeneralCustomerResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/CreateUnfiedPaymentCustomer")
	defer segment.End()

	customer, err := c.customerRepo.GetMerchantCustomerByEmail(ctx, customerModel.GetMerchantCustomerRequest{
		MerchantID: request.MerchantID,
		Email:      request.Email,
	})

	if err != nil {
		c.logger.Error(ctx, "Error get customer of merchant", logger.Error(err), logger.String("merchantId", request.MerchantID), logger.String("email", request.Email))
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}

	if customer != nil {
		if customer.Metadata != nil {
			if request.Metadata == nil {
				request.Metadata = map[string]any{}
			}

			_ = util.MergeStructToMap(&request.Metadata, &customer.Metadata)
		}

		customer.Update(&customerModel.UpdateCustomerRequest{
			Email:            &request.Email,
			PhoneCountryCode: &request.PhoneCountryCode,
			PhoneNumber:      &request.PhoneNumber,
			FirstName:        &request.FirstName,
			LastName:         &request.LastName,
			Metadata:         request.Metadata,
		})
		err := c.customerRepo.Update(ctx, customer)
		return customer.ToGeneralCustomerResponse(), err
	}

	customer = customerModel.CreateCustomer(&customerModel.CreateCustomerRequest{
		MerchantID:       request.MerchantID,
		Email:            request.Email,
		PhoneCountryCode: request.PhoneCountryCode,
		PhoneNumber:      request.PhoneNumber,
		FirstName:        request.FirstName,
		LastName:         request.LastName,
		Metadata:         request.Metadata,
	})
	err = c.customerRepo.Create(ctx, customer)
	if err != nil {
		c.logger.Error(ctx, "Error when create customer", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseCreateCustomer)
	}

	return customer.ToGeneralCustomerResponse(), nil
}
