package recurringContractService

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) Create(ctx context.Context, request model.CreateRecurringContractRequest) (*model.CreateRecurringContractResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/recurringContract/Create")
	defer span.End()

	if request.CustomerID != nil {
		if _, err := s.customerSvc.GetCustomerById(ctx, *request.CustomerID, request.MerchantID); err != nil {
			return nil, err // Error information has been wrapped
		}

	} else {
		createCustomerPayload := customerModel.CreateUnifiedPaymentCustomerRequest{
			MerchantID: request.MerchantID,
			FirstName:  request.Customer.GivenName,
			LastName:   request.Customer.GetSurname(),
			Email:      request.Customer.Email,
		}
		if request.Customer.PhoneNumber != nil {
			createCustomerPayload.PhoneNumber = request.Customer.PhoneNumber.Number
			createCustomerPayload.PhoneCountryCode = request.Customer.PhoneNumber.CountryCode
		}
		customer, err := s.customerSvc.CreateUnfiedPaymentCustomer(ctx, createCustomerPayload)
		if err != nil {
			return nil, err // Error information has been wrapped
		}
		request.CustomerID = &customer.UUID
	}

	recurringContract := model.RecurringContract{
		UUID:              util.GenerateUUID().String(),
		MerchantID:        request.MerchantID,
		ClientReferenceID: request.ClientReferenceID,
		CustomerID:        *request.CustomerID,
		AuthMethod:        request.FirstAuthorization,
		SchedulerMode:     request.Mode,
		Plan:              request.Plan,
		Trials:            request.Trials,
		Billing: model.Billing{
			Interval:     request.BillingInterval,
			IntervalUnit: request.BillingIntervalUnit,
		},
		Currency:  request.Amount.Currency,
		Amount:    float64(request.Amount.Value),
		Status:    constant.RecurringContractStatusCreated,
		CreatedBy: request.CreatedBy,
		CreatedAt: time.Now().UTC(),
		UpdatedBy: request.CreatedBy,
		UpdatedAt: time.Now().UTC(),
	}
	recurringContract.EndDate, _ = time.Parse(time.RFC3339, request.EndDate)

	if err := s.repo.Insert(ctx, recurringContract); err != nil {
		if errors.Is(err, constant.ErrClientReferenceIDAlreadyExist) {
			return nil, pkgErrs.New(response.HttpErrDupCheck, err)
		}
		s.log.Error(ctx, "Failed when inserting recurring contract data", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	return &model.CreateRecurringContractResponse{
		RecurringID: recurringContract.UUID,
		CustomerID:  *request.CustomerID,
	}, nil
}
