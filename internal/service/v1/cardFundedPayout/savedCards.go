package cardFundedPayoutService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

func (s *service) CreateSavedCard(ctx context.Context, request *cardFundedPayoutModel.CreateSavedCardRequest) (*cardFundedPayoutModel.CreateSavedCardResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/CreateSavedCard")
	defer span.End()

	// Create generated customer
	customer, err := s.customerSvc.CreateUnfiedPaymentCustomer(ctx, customerModel.CreateUnifiedPaymentCustomerRequest{
		MerchantID: request.MerchantID,
		Email:      fmt.Sprintf("auto-generated-%s@example.com", request.ReferenceID),
	})
	if err != nil {
		return nil, err
	}

	// Define max expiry config
	expiryConfig := paymentModel.PaymentMethodExpiryConfig{
		Duration: s.config.UnifiedPaymentConfig.CardConfig.MaxExpiryDuration,
		Unit:     s.config.UnifiedPaymentConfig.CardConfig.MaxExpiryDurationUnit,
	}

	// Create Payment to use One Dollar Authorization Flow
	session, err := s.unifiedPaymentSvc.CreateSession(ctx, &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
		MerchantID:        request.MerchantID,
		ClientReferenceID: request.ReferenceID,
		Amount: unifiedPaymentModel.Amount{
			Value:    constant.UnifiedPaymentOneDollarAuthorizationAmount,
			Currency: constant.CurrencyIDR,
		},
		AutoConfirm:    true,
		ExpiryAt:       expiryConfig.ToDateTime(),
		ExpirationMode: constant.UnifiedPaymentExpirationModeLoose,
		Mode:           constant.UnifiedPaymentModeRedirect,
		PaymentType:    constant.UnifiedPaymentOneDollarAuthorization,
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{
			Type: constant.UnifiedPaymentMethodCard,
		},
		PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
			Card: &unifiedPaymentModel.PaymentMethodOptionCard{
				CaptureMethod: constant.UnifiedPaymentCardCaptureMethodManual,
				ThreeDsMethod: constant.CardThreeDsMethodChallenge,
			},
		},
		StatementDescriptor: "",
		SaveForFutureUse:    util.BoolPtr(true),
		ShowSavedPayment:    util.BoolPtr(false),
		RedirectUrl: unifiedPaymentModel.RedirectUrl{
			SuccessReturnUrl: s.config.MerchantPortalConfig.CardFundedPayoutURL,
		},
		CustomerID:  customer.UUID,
		CreatedFrom: constant.SourceMerchantPortal,
		CreatedBy:   request.CreatedBy,
		OneDollarAuthorization: &unifiedPaymentModel.OneDollarAuthorization{
			UseCase: constant.UnifiedPaymentUseCaseCardFundedPayoutSavedCards,
		},
	})
	if err != nil {
		return nil, err
	}

	return &cardFundedPayoutModel.CreateSavedCardResponse{
		ReferenceID: request.ReferenceID,
		PaymentUrl:  session.ShortPaymentUrl,
	}, nil
}

func (s *service) GetSavedCardList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetSavedCardList) (*commonModel.PaginationResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/GetSavedCardList")
	defer span.End()

	resp, err := s.customerSvc.GetCardFundedPayoutSavedCardList(ctx, filter)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
