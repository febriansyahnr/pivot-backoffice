package unifiedPaymentService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	splitRoutingPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/splitRoutingPayment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) CreateDashboardPaymentLink(ctx context.Context, request *unifiedPaymentModel.DashboardPaymentLinkCreateRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/CreateDashboardPaymentLink")
	defer segment.End()

	requesterMerchant, err := s.merchantSvc.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return nil, err
	}
	if requesterMerchant == nil {
		return nil, errPkg.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	customer, err := s.customerSvc.CreateUnfiedPaymentCustomer(ctx, customerModel.CreateUnifiedPaymentCustomerRequest{
		MerchantID: request.MerchantID,
		Email:      request.Customer.Email,
	})
	if err != nil {
		return nil, err
	}

	unifiedPaymentRequest := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
		MerchantID:        request.MerchantID,
		ClientReferenceID: request.ClientReferenceID,
		Amount: unifiedPaymentModel.Amount{
			Value:    request.Amount.Value,
			Currency: constant.CurrencyIDR,
		},
		AutoConfirm:         false,
		ExpiryAt:            request.ExpiredAt,
		ExpirationMode:      constant.UnifiedPaymentExpirationModeLoose,
		Mode:                constant.UnifiedPaymentModeRedirect,
		PaymentType:         constant.UnifiedPaymentTypeSingle,
		StatementDescriptor: "",
		SaveForFutureUse:    util.BoolPtr(false),
		ShowSavedPayment:    util.BoolPtr(false),
		RedirectUrl:         unifiedPaymentModel.RedirectUrl{},
		CustomerInformation: &unifiedPaymentModel.CustomerInformation{
			Email: request.Customer.Email,
		},
		CustomerID:  customer.UUID,
		CreatedFrom: constant.SourceMerchantPortal,
		CreatedBy:   request.UserID,
	}
	if requesterMerchant.ParentID.Valid && requesterMerchant.ParentID.String != "" {
		unifiedPaymentRequest.ParentMerchantID = requesterMerchant.ParentID.String
		unifiedPaymentRequest.SplitRoutingConfigurations = &[]splitRoutingPaymentModel.PaymentSplitRoutingConfiguration{
			{
				MerchantId:  requesterMerchant.ParentID.String,
				FixedAmount: request.Amount.Value,
				Currency:    constant.CurrencyIDR,
				Type:        constant.SplitRoutingPaymentTypeFixed,
			},
		}
	}

	data, err := s.internalUnifiedPaymentSvc.CreateSession(ctx, unifiedPaymentRequest)
	if err != nil {
		s.logger.Error(ctx, "error when create unified payment session", logger.Error(err), logger.Any("request", request))
		return nil, err
	}

	return data, nil
}
