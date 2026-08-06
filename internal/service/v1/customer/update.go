package customerService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
)

func (c *CustomerService) UpdateCustomer(ctx context.Context, request customerModel.UpdateCustomerRequest) (*customerModel.GeneralCustomerResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/UpdateCustomer")
	defer segment.End()

	customer, err := c.customerRepo.GetCustomerById(ctx, request.UUID, request.MerchantID)
	if err != nil {
		c.logger.Error(ctx, "Error get customer by by id", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}
	if customer == nil {
		return nil, errors.New(response.HttpErrNotFound, constant.ErrCustomerNotFound)
	}

	if request.PhoneNumber != nil {
		customerWithPhoneNumber, err := c.customerRepo.GetCustomerByPhoneNumber(ctx, *request.PhoneNumber, request.MerchantID)
		if err != nil {
			c.logger.Error(ctx, "Error get customer by phone number", logger.Error(err))
			return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
		}

		if customerWithPhoneNumber != nil && customerWithPhoneNumber.UUID != request.UUID {
			return nil, errors.New(response.HttpErrRequest, constant.ErrPhoneNumberAlreadyExists)
		}
	}

	customer.Update(&request)
	err = c.customerRepo.Update(ctx, customer)
	if err != nil {
		c.logger.Error(ctx, "Error when update customer data", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseUpdateCustomer)
	}

	return customer.ToGeneralCustomerResponse(), nil
}
