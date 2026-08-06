package customerService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *CustomerService) DeleteCustomer(ctx context.Context, id, merchantId string) (*customerModel.GeneralCustomerResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/DeleteCustomer")
	defer segment.End()

	customer, err := c.customerRepo.GetCustomerById(ctx, id, merchantId)
	if err != nil {
		c.logger.Error(ctx, "Error when get customer", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}

	if customer == nil {
		return nil, errors.New(response.HttpErrNotFound, constant.ErrCustomerNotFound)
	}

	err = c.customerRepo.Delete(ctx, id, merchantId)
	if err != nil {
		c.logger.Error(ctx, "Error when delete customer", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseDeleteCustomer)
	}

	return customer.ToGeneralCustomerResponse(), nil
}
