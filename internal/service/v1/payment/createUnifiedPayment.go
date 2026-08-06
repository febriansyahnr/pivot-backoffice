package paymentService

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) CreateUnifiedPayment(ctx context.Context, request *paymentModel.CreateUnifiedPaymentRequest) (*paymentModel.CreateUnifiedPaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/CreateUnifiedPayment")
	defer segment.End()

	changePaymentMethod, _ := ctx.Value(constant.CtxChangePaymentMethod).(bool)

	// Generate payment ID
	if request.PaymentID == "" {
		request.PaymentID = uuid.NewString()
	}

	if err := s.validateSplitRouteConfigurations(ctx, request); err != nil {
		return nil, err
	}

	// Validate clientReferenceID first
	if paymentByRef, err := s.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, request.MerchantID, request.ClientReferenceID); err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)

	} else if paymentByRef != nil && !changePaymentMethod {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrClientReferenceIDAlreadyExist)
	}

	paymentToken, errToken := s.GeneratePaymentToken(ctx, request.PaymentID, request.ExpiryAt)
	if errToken != nil {
		return nil, pkgErr.New(response.HttpErrInternal, errToken)
	}

	isCreatedPayment := false
	defer func() {
		if !isCreatedPayment {
			_ = s.redis.Del(ctx, fmt.Sprintf(constant.PaymentTokenCacheKey, util.HashString(paymentToken)))
		}
	}()
	request.PaymentURL = fmt.Sprintf(s.config.PaymentUIConfig.PaymentLinkURL, paymentToken)

	unifiedPaymentResp := &paymentModel.CreateUnifiedPaymentResponse{
		ID:                request.PaymentID,
		ClientReferenceID: request.ClientReferenceID,
		PaymentMethod:     request.PaymentMethod,
		Amount:            request.Amount,
		ExpiryAt:          request.ExpiryAt,
		PaymentLink:       request.PaymentURL,
	}

	// Record payment method requirement
	s.RecordPaymentStatusHistory(ctx, request.PaymentID, constant.StatusHistoryActorSystem, constant.PaymentStatusHistoryRequirePaymentMethod)

	switch request.PaymentMethod {
	case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		_, err := s.requestPaymentVirtualAccount(ctx, request)
		if err != nil {
			return nil, err
		}

	case paymentConstant.PAYMENT_METHOD_QRIS:
		_, err := s.requestPaymentQris(ctx, request)
		if err != nil {
			return nil, err
		}

	case paymentConstant.PAYMENT_METHOD_CREDIT_CARD:
		_, err := s.requestPaymentCreditCard(ctx, request)
		if err != nil {
			return nil, err
		}

		// Use CC core webview instead of payment UI
		//unifiedPaymentResp.PaymentLink = paymentCC.PaymentURL
	default:
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, fmt.Errorf("invalid payment method: %s", request.PaymentMethod))
	}

	isCreatedPayment = true
	s.publishExpiryMessage(ctx, request)

	return unifiedPaymentResp, nil
}

func (s *PaymentService) requestPaymentVirtualAccount(ctx context.Context, request *paymentModel.CreateUnifiedPaymentRequest) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/requestPaymentVirtualAccount")
	defer segment.End()

	createPaymentRequest := paymentModel.PaymentRequest{
		UUID:           request.PaymentID,
		ReferenceID:    request.ClientReferenceID,
		PaymentMethod:  request.PaymentMethod,
		TotalAmount:    request.Amount,
		Customer:       request.Customer,
		PaymentItems:   request.PaymentItems,
		IsSnap:         false,
		InitiateStatus: paymentConstant.UnifiedPaymentStatusWaitingForPayment,
		VirtualAccount: &paymentModel.PaymentMetadataVirtualAccount{
			Issuer:                request.PaymentMethodOptions.VirtualAccount.Issuer,
			VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
			ExpiredDate:           &request.ExpiryAt,
		},
		ClientRedirectUrl:          request.RedirectUrl,
		PaymentUrl:                 request.PaymentURL,
		IsUnifiedPayment:           true,
		SplitRoutingConfigurations: request.SplitRoutingConfigurations,
		CreatedBy:                  request.CreatedBy,
	}

	return s.internal.CreatePayment(ctx, request.MerchantID, createPaymentRequest)
}

func (s *PaymentService) requestPaymentQris(ctx context.Context, request *paymentModel.CreateUnifiedPaymentRequest) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/requestPaymentQris")
	defer segment.End()

	expiryInSecond := int(request.ExpiryAt.Sub(time.Now().UTC()).Seconds())
	if expiryInSecond > constant.QrisDynamicValidityPeriodMax {
		expiryInSecond = constant.QrisDynamicValidityPeriodMax
	}

	createPaymentRequest := paymentModel.PaymentRequest{
		UUID:           request.PaymentID,
		ReferenceID:    request.ClientReferenceID,
		PaymentMethod:  request.PaymentMethod,
		TotalAmount:    request.Amount,
		Customer:       request.Customer,
		PaymentItems:   request.PaymentItems,
		IsSnap:         false,
		InitiateStatus: paymentConstant.UnifiedPaymentStatusWaitingForPayment,
		Qris: &paymentModel.PaymentMetadataQris{
			QrType:         constant.QrTypeDynamic,
			QrMethodType:   constant.QrMethodTypeMPM,
			Amount:         &request.Amount,
			ValidityPeriod: expiryInSecond,
		},
		ClientRedirectUrl:          request.RedirectUrl,
		PaymentUrl:                 request.PaymentURL,
		IsUnifiedPayment:           true,
		SplitRoutingConfigurations: request.SplitRoutingConfigurations,
		CreatedBy:                  request.CreatedBy,
	}

	return s.internal.CreatePayment(ctx, request.MerchantID, createPaymentRequest)
}

func (s *PaymentService) requestPaymentCreditCard(ctx context.Context, request *paymentModel.CreateUnifiedPaymentRequest) (*creditcardModel.CreateCardPaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/requestPaymentCreditCard")
	defer segment.End()

	paymentUUID, _ := uuid.Parse(request.PaymentID)
	merchantUUID, _ := uuid.Parse(request.MerchantID)
	token, err := s.GeneratePaymentToken(ctx, paymentUUID.String(), request.ExpiryAt)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrInternal, err)
	}

	createPaymentRequest := creditcardModel.CreateCardPaymentRequest{
		PaymentUUID:          paymentUUID,
		ReferenceID:          request.ClientReferenceID,
		MerchantID:           merchantUUID,
		Currency:             request.Amount.Currency,
		Amount:               request.Amount.Value.Round(2),
		BankMerchantID:       request.PaymentMethodOptions.Card.BankMerchantID,
		AuthenticationMethod: request.PaymentMethodOptions.Card.AuthenticationMethod,
		RedirectUrl: creditcardModel.CreditcardRedirectUrlRequest{
			SuccessUrl: fmt.Sprintf(s.config.PaymentUIConfig.PaymentSuccessURL, token),
			FailedUrl:  fmt.Sprintf(s.config.PaymentUIConfig.PaymentFailedURL, token),
		},
		UnifiedPaymentRedirectUrl: creditcardModel.UnifiedPaymentRedirectUrl{
			SuccessUrl: request.RedirectUrl.SuccessUrl,
			FailedUrl:  request.RedirectUrl.FailedUrl,
		},
		IsUnifiedPayment:           true,
		SplitRoutingConfigurations: request.SplitRoutingConfigurations,
		CreatedBy:                  request.CreatedBy,
	}

	return s.creditCardSvc.CreatePayment(ctx, createPaymentRequest)
}

func (s *PaymentService) GeneratePaymentToken(ctx context.Context, paymentID string, expiryAt time.Time) (string, error) {
	// Generate JWT token
	token, err := s.jwt.GeneratePaymentToken(paymentID, expiryAt)
	if err != nil {
		s.logger.Error(ctx, "error generate payment token", logger.Error(err))
		return "", err
	}

	// Store hashed 256 jwt token to redis
	if err = s.redis.Set(ctx, fmt.Sprintf(constant.PaymentTokenCacheKey, util.HashString(token)), true, expiryAt.Sub(time.Now().UTC())).Err(); err != nil {
		s.logger.Error(ctx, "error set payment token", logger.Error(err))
		return "", err
	}

	return token, nil
}

func (s *PaymentService) publishExpiryMessage(ctx context.Context, request *paymentModel.CreateUnifiedPaymentRequest) {
	now := time.Now().UTC()

	// Set lastPublishExpiryAt, equals to 01.00 JKT time
	lastPublishExpiryAt := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
	if now.Hour() >= 18 && now.Hour() < 24 {
		lastPublishExpiryAt = lastPublishExpiryAt.Add(24 * time.Hour)
	}

	if request.ExpiryAt.Before(lastPublishExpiryAt) {
		chargeStatus := constant.ChargeStatusExpired
		if err := s.rabbitMqExt.PublishWithDelay(
			ctx,
			rabbitMqExt.PaymentExpirationRoutingKey,
			&paymentModel.ExpiringPayment{
				UUID:         request.PaymentID,
				MerchantID:   request.MerchantID,
				ExpiredAt:    request.ExpiryAt,
				ChargeStatus: chargeStatus,
			},
			request.ExpiryAt.Sub(now),
		); err != nil {
			s.logger.Error(ctx, "error publish payment expiration message", logger.Error(err))
		} else {
			s.RecordPaymentStatusHistory(ctx, request.PaymentID, constant.StatusHistoryActorSystem, constant.ChargeStatusExpired)
		}
	}
}

func (s *PaymentService) validateSplitRouteConfigurations(ctx context.Context, request *paymentModel.CreateUnifiedPaymentRequest) error {
	if !(request.SplitRoutingConfigurations != nil && len(*request.SplitRoutingConfigurations) > 0) {
		return nil
	}

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if merchant == nil {
		return pkgErr.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}

	parentMerchantID := request.MerchantID
	if merchant.ParentID.Valid {
		parentMerchantID = merchant.ParentID.String
	}

	subMerchantIDs, err := s.merchantRepo.GetSubMerchantIdListByParentId(ctx, parentMerchantID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, err)
	} else if subMerchantIDs == nil {
		subMerchantIDs = []string{}
	}

	allowedRouteMerchantIDs := subMerchantIDs
	allowedRouteMerchantIDs = append(allowedRouteMerchantIDs, parentMerchantID)

	for _, val := range *request.SplitRoutingConfigurations {
		if !slices.Contains(allowedRouteMerchantIDs, val.MerchantId) || val.MerchantId == request.MerchantID {
			err = errors.New("merchant destination is not allowed")
			s.logger.Error(ctx, "error validate split routing configurations", logger.Error(err))

			return pkgErr.New(response.HttpErrUnprocessableContent, err)
		}
	}

	return nil
}
