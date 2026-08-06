package customerService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *CustomerService) GetCustomerList(ctx context.Context, merchantId, phoneNumber string, page, perPage int64) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/customer/GetCustomerList")
	defer segment.End()

	customers, meta, err := c.customerRepo.GetCustomerList(ctx, merchantId, phoneNumber, page, perPage)
	customersResp := make([]customerModel.GeneralCustomerResponse, 0)
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}
	if customers == nil {
		return nil, errors.New(response.HttpErrNotFound, constant.ErrCustomerNotFound)
	}
	for _, customer := range customers {
		customersResp = append(customersResp, *customer.ToGeneralCustomerResponse())
	}
	return &commonModel.PaginationResponse{
		Data: customersResp,
		Meta: *meta,
	}, nil
}

func (c *CustomerService) GetCustomerById(ctx context.Context, id, merchantId string) (*customerModel.GeneralCustomerResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/customer/GetCustomerById")
	defer segment.End()

	customer, err := c.customerRepo.GetCustomerById(ctx, id, merchantId)
	if err != nil {
		c.logger.Error(ctx, "Failed when retrieving customer data using ID", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}
	if customer == nil {
		return nil, errors.New(response.HttpErrNotFound, constant.ErrCustomerNotFound)
	}
	return customer.ToGeneralCustomerResponse(), nil
}

func (c *CustomerService) GetCustomerByPhoneNumber(ctx context.Context, phoneNumber, merchantId string) (*customerModel.GeneralCustomerResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/customer/GetCustomerByPhoneNumber")
	defer segment.End()

	customer, err := c.customerRepo.GetCustomerByPhoneNumber(ctx, phoneNumber, merchantId)
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}
	if customer == nil {
		return nil, errors.New(response.HttpErrNotFound, constant.ErrCustomerNotFound)
	}
	return customer.ToGeneralCustomerResponse(), nil
}

func (c *CustomerService) FindCustomerByID(ctx context.Context, id string) (*customerModel.GeneralCustomerResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/customer/FindCustomerByID")
	defer segment.End()

	customer, err := c.customerRepo.FindCustomerById(ctx, id)
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}
	if customer == nil {
		return nil, errors.New(response.HttpErrNotFound, constant.ErrCustomerNotFound)
	}
	return customer.ToGeneralCustomerResponse(), nil

}

func (c *CustomerService) GetCustomerByIDForUnifiedPayment(ctx context.Context, id, merchantId string) (*unifiedPaymentModel.CustomerInformationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/customer/GetCustomerByIDForUnifiedPayment")
	defer segment.End()

	customer, err := c.customerRepo.GetCustomerById(ctx, id, merchantId)
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}
	if customer == nil {
		return nil, errors.New(response.HttpErrNotFound, constant.ErrCustomerNotFound)
	}
	resp := customer.ToUnifiedPaymentCustomerResponse()
	for _, method := range resp.StoredPaymentMethods {
		if method.Card != nil {
			method.Card.Fingerprint = "" // Remove fingerprint response for open API
		}
	}

	return resp, nil

}

func (c *CustomerService) GetCardFundedPayoutSavedCardList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetSavedCardList) (*commonModel.PaginationResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/customer/GetCardFundedPayoutSavedCardList")
	defer span.End()

	customers, err := c.customerRepo.GetCardFundedPayoutSavedCardList(ctx, filter)
	if err != nil {
		return nil, errors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}

	return customers, nil
}

func (c *CustomerService) GetCardFundedPayoutSavedCardDetail(ctx context.Context, request cardFundedPayoutModel.GetSavedCardDetailRequest) (*cardFundedPayoutModel.GetSavedCardResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/customer/GetCardFundedPayoutSavedCardDetail")
	defer span.End()

	card, err := c.customerRepo.GetCardFundedPayoutSavedCardDetail(ctx, request)
	if err != nil {
		c.logger.Error(ctx, "Failed to retrieve card details", logger.Error(err))
		return nil, errors.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if card == nil {
		return nil, errors.New(response.HttpErrNotFound, fmt.Errorf("card details with ID %s were not found", request.CardID))
	}
	return card, nil
}
