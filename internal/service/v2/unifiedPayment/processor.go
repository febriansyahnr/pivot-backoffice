package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"

	"github.com/shopspring/decimal"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	card "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	snapCoreBaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/ewallet"
	snapCoreQrModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/topUpSimulation"
	snapCoreVaModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/types"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) ProcessorInitialization(ctx context.Context, request *unifiedPaymentModel.BaseProcessorRequest) (resp *unifiedPaymentModel.ChargePaymentMethodDetails, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/ProcessorInitialization")
	defer segment.End()

	ctx = context.WithValue(ctx, c.CtxClientReqKey, &outbound.Client{
		OriginId:    request.PaymentID,
		ReferenceId: request.MerchantID,
		From:        serviceName,
	})

	switch request.PaymentMethod.Type {
	case c.UnifiedPaymentMethodVA:

		vaRequest := &unifiedPaymentModel.InitProcessorVARequest{
			BaseProcessorRequest: request,
			ExpiryAt:             request.ExpiryAt,
		}
		if request.PaymentMethodOptions.VirtualAccount != nil {
			vaRequest.VANumber = request.PaymentMethodOptions.VirtualAccount.VirtualAccountNumber
			vaRequest.VAAccountName = request.PaymentMethodOptions.VirtualAccount.VirtualAccountName
			vaRequest.Acquirer = request.PaymentMethodOptions.VirtualAccount.Channel
			vaRequest.VirtualAccountTrxType = request.PaymentMethodOptions.VirtualAccount.VirtualAccountTrxType
			vaRequest.PaymentMethodOptions.VirtualAccount.BillDetails = request.PaymentMethodOptions.VirtualAccount.BillDetails

			if request.PaymentMethodOptions.VirtualAccount.ExpiryAt != nil {
				vaRequest.ExpiryAt = *request.PaymentMethodOptions.VirtualAccount.ExpiryAt
			}
		}

		resp, err = s.initVirtualAccount(ctx, vaRequest)
		return resp, s.overridePartnerErrorToRateLimitIfNeeded(ctx, err)

	case c.UnifiedPaymentMethodQris:
		qrRequest := &unifiedPaymentModel.InitProcessorQRISRequest{
			BaseProcessorRequest: request,
			ExpiryAt:             request.ExpiryAt,
		}
		if request.PaymentMethodOptions.QR != nil && request.PaymentMethodOptions.QR.ExpiryAt != nil {
			request.ExpiryAt = *request.PaymentMethodOptions.QR.ExpiryAt
		}

		resp, err = s.initQRIS(ctx, qrRequest)
		return resp, s.overridePartnerErrorToRateLimitIfNeeded(ctx, err)

	case c.UnifiedPaymentMethodCard:
		if request.ShouldAuthenticateEncryptedCard() {
			return s.initEncryptedCardAuthentication(ctx, request)
		}
		// Let it blank for Card, will fill the data once payment notification received
		return &unifiedPaymentModel.ChargePaymentMethodDetails{
			ProcessorReference:            c.CreditCardCoreProcessor,
			ProcessorExpiredAt:            request.ExpiryAt,
			ProcessorTransactionTimestamp: time.Now().UTC(),
		}, nil

	case c.UnifiedPaymentMethodEWallet:
		resp, err = s.initEwalletPaymentLink(ctx, request)
		return resp, s.overridePartnerErrorToRateLimitIfNeeded(ctx, err)
	}

	return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, c.ErrPaymentMethodNotFound)
}

func (s *UnifiedPaymentService) initVirtualAccount(ctx context.Context, request *unifiedPaymentModel.InitProcessorVARequest) (resp *unifiedPaymentModel.ChargePaymentMethodDetails, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/unifiedPayment/initVirtualAccount")
	defer segment.End()

	accountName := request.DerivedMerchantShortName
	if request.VAAccountName != "" {
		if constant.IsDirectPSP(request.PaymentMethodChannelType) || request.IsSnap {
			accountName = request.VAAccountName
		} else {
			accountName += " - " + request.VAAccountName
		}
	}

	// Default for unified payment
	vaTrxType := snapCoreVaModel.VaTrxType(paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC)
	if request.IsStaticPayment {
		vaTrxType = snapCoreVaModel.VaTrxType(paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC)
		if request.Amount.Value > 0 {
			vaTrxType = snapCoreVaModel.VaTrxType(paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC)
		}
	}

	// Send subCompany for AGGREGATOR
	subCompany := ""
	if request.PaymentMethodChannelType == c.PaymentMethodChannelTypeAggregator {
		subCompany = request.DerivedMID
	}

	snapCoreBillDetailsRequest := parseBillDetails(request)
	snapCoreRequest := snapCoreVaModel.CreateVirtualAccountRequest{
		MID:      request.DerivedMID,
		VaNumber: request.VANumber,
		TotalAmount: snapCoreVaModel.Amount{
			Currency: request.Amount.Currency,
			Value:    decimal.NewFromFloat(request.Amount.Value).StringFixed(2), // amount to be paid
		},
		Acquirer:      strings.ToLower(request.Acquirer),
		BillDetails:   snapCoreBillDetailsRequest,
		IsCloseAmount: vaTrxType.IsCloseAmount,
		IsSingleUse:   vaTrxType.IsSingleUsed,
		ExpiredAt:     &request.ExpiryAt,
		AccountName:   accountName,
		AdditionalInfo: &map[string]interface{}{
			c.ProcessorExternalIDKey: request.ChargeID,
		},
		MerchantID: request.DerivedMerchantID,
		SubCompany: subCompany,
		CustomerNo: request.DerivedMerchantID,
	}

	if s.config.UnifiedPaymentConfig.VirtualAccountConfig != nil && s.config.UnifiedPaymentConfig.VirtualAccountConfig.MinAmount != nil {
		(*snapCoreRequest.AdditionalInfo)[c.ProcessorMinAmountKey] = snapCoreModel.Amount{
			Currency: request.Amount.Currency,
			Value:    fmt.Sprintf("%.2f", *s.config.UnifiedPaymentConfig.VirtualAccountConfig.MinAmount),
		}
	}
	if s.config.UnifiedPaymentConfig.VirtualAccountConfig != nil && s.config.UnifiedPaymentConfig.VirtualAccountConfig.MaxAmount != nil {
		(*snapCoreRequest.AdditionalInfo)[c.ProcessorMaxAmountKey] = snapCoreModel.Amount{
			Currency: request.Amount.Currency,
			Value:    fmt.Sprintf("%.2f", *s.config.UnifiedPaymentConfig.VirtualAccountConfig.MaxAmount),
		}
	}

	processorResp, err := s.snapCoreRepo.CreateVirtualAccount(ctx, snapCoreRequest)
	if err != nil {
		return nil, s.virtualAccountErrorMapping(request, err)
	}

	return &unifiedPaymentModel.ChargePaymentMethodDetails{
		ProcessorReference:            c.SnapCoreProcessor,
		ProcessorReferenceID:          processorResp.ID,
		ProcessorTransactionTimestamp: processorResp.CreatedAt,
		ProcessorReferenceNo:          processorResp.VirtualAccountNo,
		ProcessorExpiredAt:            request.ExpiryAt,
		VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
			Channel:               request.Acquirer,
			VirtualAccountNumber:  processorResp.VirtualAccountNo,
			VirtualAccountName:    processorResp.AccountName,
			VirtualAccountTrxType: request.VirtualAccountTrxType,
			ExpiryAt:              processorResp.ExpiredAt,
		},
	}, nil
}

func (*UnifiedPaymentService) virtualAccountErrorMapping(request *unifiedPaymentModel.InitProcessorVARequest, err error) error {
	httpError, originalErr := pkgErr.ExtractError(err)
	if isPartnerDownstreamError(httpError) {
		return pkgErr.New(httpError, c.ErrPartnerInGeneral)
	}
	if httpError != httpResponse.HttpErrRequest {
		return pkgErr.New(httpResponse.HttpErrInternal, c.ErrPartnerInGeneral)
	}

	// Check if originalErr is nil to prevent panic
	if originalErr == nil {
		return pkgErr.New(httpResponse.HttpErrInternal, c.ErrPartnerInGeneral)
	}

	// This is payment type from POV product.
	paymentType := "dynamic"
	if request.VANumber != "" {
		paymentType = "static"
	}
	if strings.Contains(originalErr.Error(), "number out of range") {
		return pkgErr.New(httpResponse.HttpErrRequest, fmt.Errorf(c.ErrDetailMsgVaNumberIsOutsideValidRangeFmt, paymentType))

	} else if strings.Contains(originalErr.Error(), "va number still in use") && request.VANumber != "" {
		return pkgErr.New(httpResponse.HttpErrRequest, errors.New(c.ErrDetailMsgVaNumberStillInUse))

	} else if strings.Contains(originalErr.Error(), "va number still in use") && request.VANumber == "" {
		return pkgErr.New(httpResponse.HttpErrRequest, fmt.Errorf(c.ErrDetailMsgNoAvailableVaNumberToAssignFmt, paymentType))
	}
	return err
}

func isPartnerDownstreamError(httpError string) bool {
	switch httpError {
	case httpResponse.HttpErrRequestLimitExceeded,
		httpResponse.HttpErrTooManyRequest,
		httpResponse.HttpErrRequestTimeout,
		httpResponse.HttpErrBadGateway,
		httpResponse.HttpErrServiceUnavailable,
		httpResponse.HttpErrThirdParty:
		return true
	default:
		return false
	}
}

func (s *UnifiedPaymentService) initQRIS(ctx context.Context, request *unifiedPaymentModel.InitProcessorQRISRequest) (resp *unifiedPaymentModel.ChargePaymentMethodDetails, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/unifiedPayment/initQRIS")
	defer segment.End()

	snapCoreRequest, err := s.getQRSnapBasicPayload(ctx, request)
	if err != nil {
		return nil, err
	}

	merchantID := request.MerchantID
	derivedID, ok := ctx.Value(constant.CtxDerivedMerchantID).(string)
	if ok && derivedID != "" {
		merchantID = derivedID
	}

	// Check if NEW flow (multi-acquirer routing) is enabled for this merchant
	if constant.IsQrMultiAcquirerRoutingEnabled(merchantID) {
		// NEW flow: Send merchant UUID to SNAP Core for priority-based routing
		snapCoreRequest.MerchantID = merchantID
		snapCoreRequest.Acquirer = "" // Clear acquirer - SNAP Core will route based on priority
	}

	// Request to snap core
	processorResp, err := s.snapCoreRepo.GenerateQrMpm(ctx, snapCoreRequest)
	if err != nil {
		httpErrorStr, _ := pkgErr.ExtractError(err)
		switch {
		case httpErrorStr == httpResponse.HttpErrRequest:
			return nil, err
		case isPartnerDownstreamError(httpErrorStr):
			return nil, pkgErr.New(httpErrorStr, c.ErrPartnerInGeneral)
		}

		return nil, pkgErr.New(httpResponse.HttpErrInternal, c.ErrPartnerInGeneral)
	}

	// Use actual acquirer from SNAP Core response, fallback to request acquirer if empty
	actualAcquirer := snapCoreRequest.Acquirer
	if processorResp.Acquirer != "" {
		actualAcquirer = processorResp.Acquirer
	}

	return &unifiedPaymentModel.ChargePaymentMethodDetails{
		ProcessorReference:            c.SnapCoreProcessor,
		ProcessorReferenceID:          processorResp.UUID,
		ProcessorTransactionTimestamp: processorResp.CreatedAt,
		ProcessorReferenceNo:          processorResp.ReferenceNo,
		ProcessorExpiredAt:            request.ExpiryAt,
		Qr: &unifiedPaymentModel.ChargePaymentMethodDetailQr{
			Acquirer:                 actualAcquirer, // Use actual acquirer from SNAP Core response
			RetrievalReferenceNumber: processorResp.ReferenceNo,
			QrContent:                processorResp.QrContent,
			QrUrl:                    processorResp.QrUrl,
			QrType:                   snapCoreRequest.QrType,
			ExpiryAt:                 processorResp.ExpiredAt,
			MerchantName:             request.DerivedMerchantShortName,
			StoreID:                  processorResp.StoreID,
		},
	}, nil
}

func (s *UnifiedPaymentService) initEwalletPaymentLink(ctx context.Context, request *unifiedPaymentModel.BaseProcessorRequest) (resp *unifiedPaymentModel.ChargePaymentMethodDetails, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/unifiedPayment/initEwalletPaymentLink")
	defer segment.End()

	// Create shortlink for Dana API
	successRedirectUrl := request.SuccessRedirectUrl
	acquirer := strings.ToUpper(request.PaymentMethodOptions.Ewallet.Channel)

	if acquirer == constant.UnifiedPaymentEWalletDanaAcquirer && request.Mode == c.UnifiedPaymentModeAPI {
		shortLink, err := s.shortLinkSvc.Create(ctx, &shortLinkModel.CreateShortLink{
			Reference:      constant.ReferencePayment,
			UniqueID:       request.PaymentID,
			DestinationURL: request.PaymentURL,
			ExpiredAt:      request.ExpiryAt.Add(30 * time.Minute), // add extra time for processing state / user try to confirm the payment in dana
		})
		if err != nil {
			return nil, err
		}

		successRedirectUrl = shortLink.ShortLinkURL
	}

	paymentRequest := &ewallet.EwalletPaymentRequest{
		OriginalReferenceId: request.ChargeID,
		Acquirer:            acquirer,
		Amount: commonModel.Amount{
			Currency: request.Amount.Currency,
			Value:    decimal.NewFromFloat(request.Amount.Value).StringFixed(2),
		},
		UrlParams: []snapCoreBaseModel.SnapURLParam{
			{
				URL:        successRedirectUrl,
				IsDeepLink: "N",
			},
		},
		ValidUpTo: "",
	}
	if !request.ExpiryAt.IsZero() {
		paymentRequest.ValidUpTo = util.ConvertToJakarta(request.ExpiryAt).Format(constant.DatetimeWithTimezone)
	}
	if partnerConfig := request.PaymentPartnerConfig; partnerConfig != nil && partnerConfig.EWallet != nil {
		switch paymentRequest.Acquirer {
		case c.UnifiedPaymentEWalletDanaAcquirer:
			paymentRequest.SubMerchantId = partnerConfig.EWallet.SubMerchantID

		case c.UnifiedPaymentEWalletShopeePayAcquirer:
			paymentRequest.MerchantId = partnerConfig.EWallet.ExternalMerchantID
			paymentRequest.ExternalStoreId = partnerConfig.EWallet.ExternalStoreID
		}
	}

	if s.config.Environment != constant.EnvironmentProduction && s.isEWalletPaymentSimulationFlowEnabled(ctx, request.MerchantID) {
		ctx = context.WithValue(ctx, constant.CtxPaymentSimulationMode, strconv.FormatBool(true))
		token := ""
		if parsedURL, errParsed := url.Parse(request.PaymentURL); errParsed == nil {
			token = parsedURL.Query().Get("token")
		}
		ctx = context.WithValue(ctx, constant.CtxPaymentSimulationToken, token)
	}

	paymentLinkResponse, err := s.snapCoreRepo.CreateEWalletPaymentLink(ctx, paymentRequest)
	if err != nil {
		return nil, err
	}

	return &unifiedPaymentModel.ChargePaymentMethodDetails{
		Ewallet: &unifiedPaymentModel.ChargePaymentMethodDetailEwallet{
			Channel:            strings.ToUpper(request.PaymentMethodOptions.Ewallet.Channel),
			AppRedirectURL:     paymentLinkResponse.AppRedirectionURL,
			WebRedirectURL:     paymentLinkResponse.WebRedirectionURL,
			ReferenceNo:        paymentLinkResponse.ReferenceNo,
			PartnerReferenceNo: paymentLinkResponse.PartnerReferenceNo,
		},
		ProcessorReference:            c.SnapCoreProcessor,
		ProcessorReferenceID:          paymentLinkResponse.UUID,
		ProcessorTransactionID:        paymentLinkResponse.UUID,
		ProcessorTransactionTimestamp: time.Now().UTC(),
		ProcessorExpiredAt:            request.ExpiryAt,
	}, nil
}

func (s *UnifiedPaymentService) initEncryptedCardAuthentication(ctx context.Context, request *unifiedPaymentModel.BaseProcessorRequest) (resp *unifiedPaymentModel.ChargePaymentMethodDetails, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/unifiedPayment/initEncryptedCardAuthentication")
	defer segment.End()

	// The merchantID variable is used to set the temporary payment session in the cache and to construct requests to the processor.
	// If the request.DerivedMerchantID parameter is not empty, the merchantID value will be replaced with the value of that parameter,
	// indicating that the processor request uses the parent merchant ID identifier.
	var merchantID = request.MerchantID
	if request.DerivedMerchantID != "" {
		s.logger.Debug(ctx, "using derived merchant ID for authentication",
			logger.String("merchant_id", request.MerchantID),
			logger.String("derived_merchant_id", request.DerivedMerchantID),
		)
		merchantID = request.DerivedMerchantID
	}

	payment, err := s.paymentRepo.GetPaymentByIdAndMerchantId(ctx, request.PaymentID, request.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "payment not found", logger.String("payment_id", request.PaymentID), logger.String("merchant_id", request.MerchantID))
		return nil, pkgErr.New(httpResponse.HttpErrInternal, c.ErrDatabaseGetData)
	}

	// Store payment details in the cache for the payment session creation process, which is forwarded directly to the next stage (a new HTTP request to the processor)
	// before the commit to the session database is performed.
	cacheKey := fmt.Sprintf(c.TemporaryPaymentRecordCacheKey, merchantID, request.PaymentID)
	b, _ := json.Marshal(payment)
	_, err = s.redis.Set(ctx, cacheKey, string(b), c.TemporaryPaymentRecordTTL).Result()
	if err != nil {
		s.logger.Error(ctx, "failed to set temporary payment record cache", logger.Error(err), logger.String("cache_key", cacheKey))
		return nil, pkgErr.New(httpResponse.HttpErrInternal, c.ErrStoreInCache)
	}

	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, errMarshal := json.Marshal(payment.Metadata)
	if errMarshal != nil {
		return nil, pkgErr.New(httpResponse.HttpErrInternal, c.ErrInvalidMarshalData)
	}

	if errUnmarshal := json.Unmarshal(metadataB, &unifiedPaymentMetadata); errUnmarshal != nil {
		return nil, pkgErr.New(httpResponse.HttpErrInternal, c.ErrInvalidUnmarshalJSON)
	}

	payload := &card.EncryptedCardAuthenticationRequest{
		PaymentID:           request.PaymentID,
		MerchantID:          merchantID,
		ClientTransactionID: request.ClientReferenceID,
		CVC:                 request.PaymentMethod.CardPaymentMethodDetail.CVC,
		Fee:                 request.Fee.InexactFloat64(),
		Amount:              request.Amount.Value,
		Currency:            request.Amount.Currency,
		SavedFutureUse:      unifiedPaymentMetadata.SaveForFutureUse,
	}
	if request.CardOnFile != nil {
		payload.CardOnFile = &card.CardOnFile{
			Initiator:                    request.CardOnFile.Initiator,
			Type:                         request.CardOnFile.Type,
			PreviousNetworkTransactionID: request.CardOnFile.PreviousNetworkTransactionID,
		}
	}
	if unifiedPaymentMetadata.PaymentOrder != nil && unifiedPaymentMetadata.PaymentOrder.BillingInformation != nil {
		billingInfo := unifiedPaymentMetadata.PaymentOrder.BillingInformation
		payload.BillingInformation = &card.BillingInformation{
			GivenName:     billingInfo.GivenName,
			SureName:      billingInfo.GetSurname(),
			Email:         billingInfo.Email,
			Address1:      billingInfo.Address1,
			Address2:      billingInfo.Address2,
			City:          billingInfo.City,
			ProvinceState: billingInfo.ProvinceState,
			Country:       billingInfo.Country,
			PostalCode:    billingInfo.PostalCode,
		}

		if billingInfo.PhoneNumber != nil {
			payload.BillingInformation.PhoneNumber = &card.PhoneNumber{
				CountryCode: billingInfo.PhoneNumber.CountryCode,
				Number:      billingInfo.PhoneNumber.Number,
			}
		}
	}

	if request.PaymentMethodOptions != nil && request.PaymentMethodOptions.Card != nil {
		// Set cardPaymentMethodOptions from request
		cardPaymentMethodOptions := *request.PaymentMethodOptions.Card

		// Set ThreeDsMethod from metadata
		payload.ThreeDsMethod = cardPaymentMethodOptions.ThreeDsMethod

		// Set ThreeDsInfo if threeDsMethod = EXTERNAL
		if cardPaymentMethodOptions.ThreeDsInfo != nil && payload.ThreeDsMethod == constant.CardThreeDsMethodExternal {
			externalThreeDsInfo := cardPaymentMethodOptions.ThreeDsInfo
			payload.ExternalThreeDsInfo = &card.ExternalThreeDsInfo{
				TransactionID:        externalThreeDsInfo.TransactionID,
				ThreeDSVersion:       externalThreeDsInfo.ThreeDSVersion,
				ECI:                  externalThreeDsInfo.ECI,
				TransactionStatus:    externalThreeDsInfo.TransactionStatus,
				AuthenticationScheme: externalThreeDsInfo.AuthenticationScheme,
				ACSReference:         externalThreeDsInfo.ACSReference,
				ACSTransactionID:     externalThreeDsInfo.ACSTransactionID,
				CAVV:                 externalThreeDsInfo.CAVV,
				Time:                 externalThreeDsInfo.AuthenticationTime,
			}
		}

	}

	if request.PaymentMethod.CardPaymentMethodDetail.Token != "" {
		customer, err := s.customerRepo.GetCustomerById(ctx, payment.CustomerID, payment.MerchantID)
		if err != nil {
			s.logger.Error(ctx, "failed to get customer by id", logger.Error(err), logger.String("customer_id", payment.CustomerID))
			return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
		} else if customer == nil {
			s.logger.Info(ctx, "failed to get customer by id", logger.String("customer_id", payment.CustomerID))
			return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, c.ErrCustomerNotFound)
		}

		customerPaymentMethods, _ := util.ConvertToStruct[[]*unifiedPaymentModel.CustomerPaymentMethod](customer.Metadata["paymentMethods"])
		for _, paymentMethod := range customerPaymentMethods {
			if paymentMethod.Card != nil && paymentMethod.Token == request.PaymentMethod.CardPaymentMethodDetail.Token {
				payload.Fingerprint = paymentMethod.Card.Fingerprint
				payload.CardHolderName = strings.TrimSpace(fmt.Sprintf("%s %s", paymentMethod.Card.CardHolderFirstName, paymentMethod.Card.CardHolderLastName))
				break
			}
		}

		if payload.Fingerprint == "" {
			s.logger.Warn(ctx, "card token is not found", logger.String("customer_id", payment.CustomerID))
			return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, fmt.Errorf("invalid card token"))
		}
	} else {
		payload.EncryptedCard = request.PaymentMethod.CardPaymentMethodDetail.EncryptedCard
		payload.EncryptedEncryptionKey = util.ValueOfPtr(unifiedPaymentMetadata.EncryptedEncryptionKey).GetPrivateKey() // The private key sent will remain encrypted using AES GCM.
	}

	// Prepare recurring payment data.
	if request.RecurringID != "" {
		payload.RecurringID = request.RecurringID
		payload.InitiateFirstAuthorization = &request.InitiateFirstAuthorization
		payload.FirstAuthorizationMethod = request.FirstAuthorizationMethod
		payload.FirstAuthorizationOrderID = request.FirstAuthorizationOrderID
		payload.RecurringBillingCycle = &card.RecurringBillingCycle{
			Interval:     request.RecurringBillingCycle.Interval,
			IntervalUnit: request.RecurringBillingCycle.IntervalUnit,
			Count:        request.RecurringBillingCycle.Count,
		}
	}

	// Card-funded payout data.
	if cardFunded := request.CardFundedPayout; cardFunded != nil {
		payload.CardFundedPayout = &creditcardCoreProcessorModel.CardFundedPayout{
			Sequence:       cardFunded.Sequence,
			Count:          cardFunded.Count,
			VendorID:       cardFunded.VendorID,
			VendorName:     cardFunded.VendorName,
			FirstPaymentID: cardFunded.FirstPaymentID,
		}
	}
	response, err := s.creditcardSvc.CreateEncryptedCardAuthenticationLink(ctx, payload)
	if err != nil {
		return nil, err
	}

	if response.AuthenticationResponse.Status == c.CreditCardProcessorStatusFailed {
		s.logger.Error(ctx, "failed to create encrypted card authentication link",
			logger.String("payment_id", request.PaymentID),
			logger.String("merchant_id", request.MerchantID),
			logger.String("status", response.AuthenticationResponse.Status),
			logger.String("message", response.AuthenticationResponse.Message))

		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, fmt.Errorf("failed to create encrypted card authentication link: %s", response.AuthenticationResponse.Message))
	}

	var authenticationData *unifiedPaymentModel.ChargePaymentMethodDetailCardAuthenticationResult
	if response.AuthenticationResponse.AuthenticationData != nil {
		authenticationData = &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthenticationResult{
			ThreeDsVersion: response.AuthenticationResponse.AuthenticationData.ThreeDsVer,
			ThreeDsResult:  response.AuthenticationResponse.AuthenticationData.AuthenticationResult,
			EciCode:        response.AuthenticationResponse.AuthenticationData.EciCode,
		}
	}

	return &unifiedPaymentModel.ChargePaymentMethodDetails{
		ProcessorReference: c.CreditCardCoreProcessor,
		Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
			First6:      response.CardInfo.First6Digits,
			Last4:       response.CardInfo.Last4Digits,
			First8:      response.CardInfo.First8Digits,
			ExpMonth:    types.String(response.CardInfo.ExpiryMonth),
			ExpYear:     types.String(response.CardInfo.ExpiryYear),
			Fingerprint: response.CardInfo.Fingerprint,
			BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
				Type:        response.Bin.CardType,
				IssuingBank: response.Bin.IssuerName,
				Brand:       response.Bin.CardBrand,
				Country:     response.Bin.IssuerCountry,
			},
			ACSURL:               response.AuthenticationResponse.AuthenticationURL.URL,
			AuthenticationResult: authenticationData,
		},
		ProcessorExpiredAt:            request.ExpiryAt,
		ProcessorTransactionTimestamp: time.Now().UTC(),
		ProcessorReferenceID:          response.AuthenticationResponse.AcquirerTransactionID,
	}, nil

}

func (s *UnifiedPaymentService) getQRSnapBasicPayload(ctx context.Context, request *unifiedPaymentModel.InitProcessorQRISRequest) (snapCoreQrModel.GenerateQrMpmRequest, error) {
	var (
		snapCoreRequest = snapCoreQrModel.GenerateQrMpmRequest{
			PartnerReferenceNo: request.ClientReferenceID,
			Amount: commonModel.Amount{
				Currency: request.Amount.Currency,
				Value:    decimal.NewFromFloat(request.Amount.Value).StringFixed(2),
			},
			QrType:         c.QrTypeDynamic,
			ValidityPeriod: int(math.Round(request.ExpiryAt.Sub(time.Now().UTC()).Seconds())),
			AdditionalInfo: map[string]interface{}{
				c.ProcessorExternalIDKey: request.ChargeID,
			},
		}
	)

	if request.IsStaticPayment {
		snapCoreRequest.QrType = c.QrTypeStatic
		snapCoreRequest.ValidityPeriod = 0
	} else if snapCoreRequest.ValidityPeriod > c.QrisDynamicValidityPeriodMax {
		snapCoreRequest.ValidityPeriod = c.QrisDynamicValidityPeriodMax
		request.ExpiryAt = time.Now().UTC().Add(time.Duration(c.QrisDynamicValidityPeriodMax) * time.Second)
	}

	// when the partner config exist, then should use this instead of DB
	if util.ValueOfPtr(request.PaymentPartnerConfig).Qris != nil {
		snapCoreRequest.SubMerchantID = request.PaymentPartnerConfig.Qris.AcquirerMerchantID
		snapCoreRequest.TerminalID = request.PaymentPartnerConfig.Qris.AcquirerTerminalID
		snapCoreRequest.MerchantID = request.PaymentPartnerConfig.Qris.AcquirerMerchantID
		snapCoreRequest.Acquirer = request.PaymentPartnerConfig.Qris.Acquirer

		return snapCoreRequest, nil
	}

	// TODO: handle the qris registration to use acquirer to get valid data
	qrRegistration, err := s.qrisSvc.FindQrRegistrationByExternalID(ctx, request.MerchantExternalID)
	if err != nil {
		if err.Error() == pkgErr.New(httpResponse.HttpErrNotFound, c.ErrDataNotFound).Error() || qrRegistration == nil {
			return snapCoreRequest, pkgErr.New(httpResponse.HttpErrUnprocessableContent, c.ErrMerchantNotRegisteredQR)
		}

		s.logger.Error(ctx, "failed to get qr registration", logger.Error(err))
		return snapCoreRequest, err
	}

	snapCoreRequest.SubMerchantID = util.ValueOfPtr(qrRegistration.AcquirerMerchantId)
	snapCoreRequest.MerchantID = util.ValueOfPtr(qrRegistration.AcquirerMerchantId)
	snapCoreRequest.TerminalID = util.ValueOfPtr(qrRegistration.AcquirerTerminalId)
	snapCoreRequest.Acquirer = qrRegistration.Acquirer

	if qrRegistration.MerchantType != c.QrMerchantTypeMerchant {
		snapCoreRequest.SubMerchantID = qrRegistration.AcquirerParentMerchantId
		snapCoreRequest.StoreID = util.ValueOfPtr(qrRegistration.AcquirerMerchantId)
	}

	return snapCoreRequest, nil
}

func parseBillDetails(request *unifiedPaymentModel.InitProcessorVARequest) []snapCoreVaModel.BillDetail {
	if request == nil || request.PaymentMethodOptions == nil || request.PaymentMethodOptions.VirtualAccount == nil {
		return nil
	}

	billDetails := request.PaymentMethodOptions.VirtualAccount.BillDetails

	if billDetails == nil {
		return nil
	}

	var result []snapCoreVaModel.BillDetail
	for _, bd := range *billDetails {
		result = append(result, snapCoreVaModel.BillDetail{
			BillerReferenceId: bd.BillerReferenceId,
			BillCode:          bd.BillCode,
			BillNo:            bd.BillNo,
			BillName:          bd.BillName,
			BillShortName:     bd.BillShortName,
			BillDescription: snapCoreVaModel.Description{
				English:   bd.BillDescription.English,
				Indonesia: bd.BillDescription.Indonesia,
			},
			BillSubCompany: bd.BillSubCompany,
			BillAmount: snapCoreVaModel.Amount{
				Value:    bd.BillAmount.Value,
				Currency: bd.BillAmount.Currency,
			},
		})
	}

	return result
}
