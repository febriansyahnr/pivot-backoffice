package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/gcp"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *UnifiedPaymentService) CreateSession(ctx context.Context, request *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) (_ *unifiedPaymentModel.UnifiedPaymentSessionResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/CreateSession")
	defer segment.End()

	var (
		merchantID         = request.MerchantID
		merchantExternalID string

		derivedMerchantID        = request.MerchantID
		derivedMID               string
		derivedMerchantShortname string

		customer *customerModel.Customer

		paymentMethodID          string
		paymentPartnerConfig     *paymentMethodModel.SetupPaymentMethodPartnerConfigRequest
		paymentMethodChannelType = constant.PaymentMethodChannelTypeAggregator
		paymentMethod            *paymentModel.PaymentMethodWithPivot

		paymentSessionStatus = constant.UnifiedPaymentSessionStatusRequirePaymentMethod
		chargeDetails        *unifiedPaymentModel.ChargeResponse
	)

	// Generate payment ID
	if request.PaymentID == "" {
		request.PaymentID = uuid.NewString()
	}

	// Find merchant data
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, merchantID)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrFindMerchant)
	} else if merchant == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	merchantExternalID = merchant.ExternalId
	derivedMID = merchant.MID.String
	derivedMerchantShortname = merchant.ShortName

	if merchant.ParentID.Valid {
		parentMerchant, err := s.merchantRepo.FindMerchantByID(ctx, merchant.ParentID.String)
		if err != nil {
			return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrFindParentMerchant)
		}
		// When sub-merchant create payment directly
		if _, ok := ctx.Value(constant.CtxParentMerchantId).(string); !ok {
			ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)
		}
		if merchant.KYCStatus.String != constant.KYCStatusApproved {
			derivedMerchantID = parentMerchant.UUID
			derivedMID = parentMerchant.MID.String
			derivedMerchantShortname = parentMerchant.ShortName
			merchantExternalID = parentMerchant.ExternalId
		}
	}

	// Prepare data for recurring payments.
	if request.RecurringID != "" {
		if err = s.internalUnifiedPaymentSvc.PrepareRecurringPaymentRequest(ctx, request); err != nil {
			return nil, err
		}
		defer func() {
			if err == nil {
				return
			}
			request.CleanupPreparedRecurringPaymentLock(ctx)
		}()
	}

	invalidSplitPaymentRequest := request.IsAutoSplitCardPayment() &&
		request.PaymentMethod.Type != constant.UnifiedPaymentMethodCard
	if invalidSplitPaymentRequest {
		return nil, pkgErr.New(response.HttpErrRequest, errors.New("split card payment is only supported for card payment method"))
	}

	// Validate payment method
	if request.PaymentMethod != nil && request.PaymentMethod.Type != "" {
		acquirer := ""

		switch request.PaymentMethod.Type {
		case constant.UnifiedPaymentMethodVA:
			acquirer = request.PaymentMethodOptions.VirtualAccount.Channel

		case constant.UnifiedPaymentMethodEWallet:
			acquirer = strings.ToLower(request.PaymentMethodOptions.Ewallet.Channel)
		}

		splitRoute := len(util.ValueOfPtr(request.SplitRoutingConfigurations)) > 0

		paymentMethod, err = s.validatePaymentActivation(ctx, derivedMerchantID, merchantExternalID, request.PaymentMethod.Type, acquirer, splitRoute)
		if err != nil {
			return nil, err
		}
		if err := s.ValidatePaymentExpiry(ctx, paymentModel.PaymentRequestExpiryValidation{
			MerchantID:            merchantID,
			Method:                request.PaymentMethod.Type,
			UnifiedPaymentRequest: request,
			PaymentMethod:         paymentMethod,
		}); err != nil {
			return nil, err
		}
		paymentMethodID = paymentMethod.UUID
		paymentMethodChannelType = paymentMethod.ChannelType
		cardSupportedUseCases := []*paymentMethodModel.CardSupportedUseCase{}
		paymentPartnerConfigs := []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{}
		if paymentMethod.MerchantConfigObj != nil {
			paymentPartnerConfig = paymentMethod.MerchantConfigObj.PartnerConfig

			if request.PaymentMethod.Type == constant.UnifiedPaymentMethodQris &&
				paymentPartnerConfig != nil &&
				paymentPartnerConfig.Qris != nil {
				paymentPartnerConfig.Qris.Acquirer = strings.ToUpper(paymentMethod.Acquirer)
			}

			if paymentPartnerConfig != nil && paymentPartnerConfig.Card != nil {
				for _, item := range paymentPartnerConfig.Card.Items {
					paymentPartnerConfigs = append(paymentPartnerConfigs, item)
					cardSupportedUseCases = append(cardSupportedUseCases, item.SupportedUseCase)
				}
			}
		}

		if request.PaymentMethod.Type == constant.UnifiedPaymentMethodCard &&
			request.PaymentMethodOptions.Card != nil {
			if err = s.validateCard(ctx, &unifiedPaymentModel.ValidateCardRequest{
				IsConfirmStep:            false,
				Mode:                     request.Mode,
				CardPaymentMethod:        request.PaymentMethod.CardPaymentMethodDetail,
				CardPaymentMethodOptions: request.PaymentMethodOptions.Card,
				CardSupportedUseCases:    cardSupportedUseCases,
				IsRecurringPayment:       request.RecurringID != "",
				IsVirtualTerminal:        request.VirtualTerminal != nil,
				IsCardFundedPayout:       request.CardFundedPayout != nil,
				IsAutoSplitPayment:       request.AutoSplitPayment != nil && request.AutoSplitPayment.TransactionType != constant.AutoSplitPaymentTypeAuthentication,
				HasCardOnFile:            request.HasCardOnFile(),
			}, paymentMethod); err != nil {
				return nil, err
			}
		}

		// At this time, recurring payments are only supported for card payment methods.
		if request.RecurringID != "" && request.PaymentMethod.Type == constant.UnifiedPaymentMethodCard {
			idx := slices.IndexFunc(paymentPartnerConfigs, func(p paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj) bool {
				if request.InitiateFirstAuthorization {
					return p.RecurringType == constant.CardTransactionTypeCIT && p.IsActive
				}
				return p.RecurringType == constant.CardTransactionTypeMIT && p.IsActive
			})
			if idx < 0 {
				return nil, pkgErr.New(response.HttpErrRequest, fmt.Errorf("%s", "Recurring payments are not allowed"))
			}
			if request.PaymentMethodOptions.Card == nil {
				request.PaymentMethodOptions.Card = &unifiedPaymentModel.PaymentMethodOptionCard{
					ProcessingConfig: &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{},
				}
			} else if request.PaymentMethodOptions.Card.ProcessingConfig == nil {
				request.PaymentMethodOptions.Card.ProcessingConfig = &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{}
			}
			request.PaymentMethodOptions.Card.ProcessingConfig.BankMerchantId = paymentPartnerConfigs[idx].AcquirerMerchantID
		}
	}

	if request.IsAutoSplitCardPayment() {
		if !paymentMethod.EnableSplitCardPayment() {
			return nil, pkgErr.New(response.HttpErrRequest, errors.New("split card payments are not allowed"))
		}
		err = request.PrepareAutoSplitCardPayment(
			paymentMethod.MerchantConfigObj.SplitCardPaymentConfig, paymentMethod.MerchantConfigObj.PartnerConfig, s.config.AutoSplitPayment.ProcessorLimitDefault,
		)
		if err != nil {
			return nil, err
		}
	}

	// set payment status on selected method
	paymentSessionStatus = constant.UnifiedPaymentSessionStatusRequireConfirmation

	if request.IsStaticPayment() {
		paymentSessionStatus = constant.UnifiedStaticPaymentStatusActive
	}

	// Validate payment session
	if err = s.validateCreatePaymentSession(ctx, request); err != nil {
		return nil, err
	}

	// Generate payment token
	paymentToken, errToken := s.generatePaymentToken(ctx, request.PaymentID, request.ExpiryAt)
	if errToken != nil {
		return nil, pkgErr.New(response.HttpErrInternal, errToken)
	}

	// handle payment customer
	if request.CustomerID != "" {
		customer, err = s.customerRepo.GetCustomerById(ctx, request.CustomerID, request.MerchantID)
		if err != nil {
			return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
		}

		// check if customer is blocked and not allowed to create payment
		if customer.IsBlocked {
			s.logger.Error(ctx, "customer is blocked", logger.Error(pkgErr.New(response.HttpErrRequest, constant.ErrCustomerIsBlocked)))
			return nil, pkgErr.New(response.HttpErrRequest, constant.ErrCustomerIsBlocked)
		}
	}

	// TODO: Implement distributed lock here.

	isCreatedPayment := false
	defer func() {
		if !isCreatedPayment {
			_ = s.redis.Del(ctx, fmt.Sprintf(constant.PaymentTokenCacheKey, util.HashString(paymentToken)))
		}
	}()
	// TODO: Shorten it later
	request.PaymentURL = fmt.Sprintf(s.config.PaymentUIConfig.PaymentLinkURL, paymentToken)
	if request.CreatedFrom == constant.SourceMerchantPortal {
		shortLink, err := s.shortLinkSvc.Create(ctx, &shortLinkModel.CreateShortLink{
			Reference:      constant.ReferencePayment,
			UniqueID:       request.PaymentID,
			DestinationURL: request.PaymentURL,
			ExpiredAt:      request.ExpiryAt,
		})
		if err != nil {
			return nil, err
		}
		request.ShortPaymentURL = shortLink.ShortLinkURL
	}

	unifiedPaymentResp := &unifiedPaymentModel.UnifiedPaymentSessionResponse{
		ClientReferenceID:          request.ClientReferenceID,
		Amount:                     request.Amount,
		AutoConfirm:                request.AutoConfirm,
		Mode:                       request.Mode,
		RedirectUrl:                request.RedirectUrl,
		PaymentType:                request.PaymentType,
		PaymentMethod:              request.PaymentMethod,
		PaymentMethodOptions:       request.PaymentMethodOptions,
		StatementDescriptor:        request.StatementDescriptor,
		SplitRoutingConfigurations: request.SplitRoutingConfigurations,
		SaveForFutureUse:           request.SaveForFutureUse,
		ShowSavedPayment:           request.ShowSavedPayment,
		ExpirationMode:             request.ExpirationMode,
		RecurringID:                request.RecurringID,
		InitiateFirstAuthorization: request.InitiateFirstAuthorization,
		Metadata:                   request.Metadata,

		ID:               request.PaymentID,
		BypassStatusPage: request.BypassStatusPage,
	}

	if !request.ExpiryAt.IsZero() {
		unifiedPaymentResp.ExpiryAt = &request.ExpiryAt
	}

	// Set encryption key for CARD API
	unifiedPaymentResp.SetEncryptionKeyForCard()

	if customer != nil {
		unifiedPaymentResp.CustomerId = &customer.UUID
		unifiedPaymentResp.CustomerInformation = customer.ToUnifiedPaymentCustomerResponse()
	}

	// Begin trx
	ctxTrx, errCtx := s.paymentRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return nil, errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.paymentRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErr.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	amount := decimal.NewFromFloat(request.Amount.Value)
	discount := decimal.NewFromFloat(0)
	fee := decimal.NewFromFloat(0)
	if request.CardFundedPayout != nil {
		fee = decimal.NewFromFloat(request.CardFundedPayout.FeeAmount)
	}
	totalAmount := amount.Add(fee).Sub(discount)

	unifiedPaymentMetadata := &unifiedPaymentModel.MetadataUnifiedPayment{
		IsUnifiedPaymentV2: true,
		IsMigratingFromV1:  request.IsMigratingFromV1,

		AutoConfirm:          request.AutoConfirm,
		Mode:                 request.Mode,
		PaymentMethod:        request.PaymentMethod,
		PaymentMethodOptions: request.PaymentMethodOptions,
		StatementDescriptor:  request.StatementDescriptor,
		ClientMetadata:       request.Metadata,
		PaymentOrder:         request.OrderInformation,
		SaveForFutureUse:     request.SaveForFutureUse,
		ShowSavedPayment:     request.ShowSavedPayment,
		ExpirationMode:       request.ExpirationMode,
		ShortPaymentURL:      request.ShortPaymentURL,
		BypassStatusPage:     request.BypassStatusPage,
		AutoSplitPayment:     request.AutoSplitPayment,
		FeeDetail:            request.FeeDetail,
	}
	if request.RecurringID != "" {
		unifiedPaymentMetadata.RecurringPayment = &unifiedPaymentModel.MetadataRecurringPayment{
			InitiateFirstAuthorization: request.InitiateFirstAuthorization,
			FirstAuthorizationMethod:   request.FirstAuthorizationMethod,
			FirstAuthorizationOrderID:  request.FirstAuthorizationOrderID,
			BillingCycle:               request.RecurringBillingCycle,
		}
	}
	if request.VirtualTerminal != nil {
		unifiedPaymentMetadata.VirtualTerminal = request.VirtualTerminal
		paymentSessionStatus = constant.UnifiedPaymentSessionStatusProcessing
	}
	if request.OneDollarAuthorization != nil {
		unifiedPaymentMetadata.OneDollarAuthorization = request.OneDollarAuthorization
	}
	if cardFunded := request.CardFundedPayout; cardFunded != nil {
		unifiedPaymentMetadata.CardFundedPayout = &unifiedPaymentModel.MetadataCardFundedPayout{
			Sequence:         cardFunded.Sequence,
			Count:            cardFunded.Count,
			FirstPaymentID:   cardFunded.FirstPaymentID,
			SettlementMethod: cardFunded.SettlementMethod,
			CardID:           cardFunded.CardID,
		}
		unifiedPaymentMetadata.FeeDetail = &cardFunded.FeeConfig
	}
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		unifiedPaymentMetadata.OnBehalf = &merchantModel.OnBehalfObject{
			ParentMerchantId: parentMerchantId,
		}
	}

	if request.Mode == constant.UnifiedPaymentModeRedirect || request.Mode == constant.UnifiedPaymentModeAPI {

		unifiedPaymentMetadata.ClientRedirectUrl = &request.RedirectUrl

		if request.PaymentMethod != nil && slices.Contains([]string{constant.UnifiedPaymentMethodCard, constant.UnifiedPaymentMethodEWallet}, request.PaymentMethod.Type) {
			unifiedPaymentMetadata.RedirectPaymentUIUrl = &unifiedPaymentModel.RedirectPaymentUIUrl{
				SuccessUrl: fmt.Sprintf(s.config.PaymentUIConfig.PaymentSuccessURL, paymentToken),
				FailedUrl:  fmt.Sprintf(s.config.PaymentUIConfig.PaymentFailedURL, paymentToken),
			}
		}
		if request.Mode == constant.UnifiedPaymentModeAPI {
			unifiedPaymentMetadata.RedirectPaymentUIUrl = &unifiedPaymentModel.RedirectPaymentUIUrl{
				SuccessUrl: request.RedirectUrl.SuccessReturnUrl,
				FailedUrl:  request.RedirectUrl.FailureReturnUrl,
			}
		}
		if request.IsAutoSplitCardPayment() {
			unifiedPaymentMetadata.RedirectPaymentUIUrl.ProcessingUrl = fmt.Sprintf(s.config.PaymentUIConfig.PaymentProcessingURL, paymentToken)
		}
	}
	if request.SplitRoutingConfigurations != nil && len(*request.SplitRoutingConfigurations) > 0 {
		unifiedPaymentMetadata.SplitRoutingConfigurations = request.SplitRoutingConfigurations
	}
	if unifiedPaymentResp.EncryptionKey != nil {

		secret := commonModel.EncryptionSecret{}

		version, err := gcp.GetGlobalSecretManagerClient().
			GetLatestSecretValueJSON(
				ctx, s.config.GCPConfig.ProjectId, s.config.GCPConfig.SecretManager.EncryptionSecretName, &secret,
			)
		if err != nil {
			s.logger.Error(ctxTrx, "error when get latest secret value", logger.Error(err))
			return nil, err
		}

		if unifiedPaymentMetadata.EncryptedEncryptionKey, err = unifiedPaymentResp.EncryptionKey.NewWithEncryptedKeyPair(version, secret.Payment.KeyEncryptionKey); err != nil {
			s.logger.Error(ctxTrx, "error when generate RSA key pair on unified payment", logger.Error(err))
			return nil, err
		}
	}
	if request.IsSnap {
		unifiedPaymentMetadata.IsSnap = &request.IsSnap
	}

	metaDataB, _ := json.Marshal(unifiedPaymentMetadata)
	metaDataString := string(metaDataB)

	// Insert into payment table
	paymentDTO := paymentModel.PaymentDTO{
		UUID:                     request.PaymentID,
		ReferenceID:              &request.ClientReferenceID,
		MerchantID:               request.MerchantID,
		PaymentMethodID:          paymentMethodID,
		ProcessorReferenceNumber: nil,
		RecurringContractID:      nil,
		Currency:                 request.Amount.Currency,
		Amount:                   amount,
		Fee:                      &fee,
		Discount:                 &discount,
		TotalAmount:              totalAmount,
		Status:                   paymentSessionStatus,
		Type:                     request.PaymentType,
		Metadata:                 &metaDataString,
		CreatedBy:                &request.CreatedBy,
		CreatedAt:                time.Now().UTC(),
		UpdatedAt:                time.Now().UTC(),
		PaymentURL:               request.PaymentURL,
		CreatedFrom:              request.CreatedFrom,
	}
	if request.RecurringID != "" {
		paymentDTO.RecurringContractID = &request.RecurringID
	}

	if request.CreatedFrom == "" {
		paymentDTO.CreatedFrom = constant.SourceOpenApi
	}

	if !request.ExpiryAt.IsZero() {
		paymentDTO.ExpiredAt = &request.ExpiryAt
	}

	if customer != nil {
		paymentDTO.CustomerID = customer.UUID
	}

	// Active static payment limit
	if request.IsStaticPayment() && paymentMethodID != "" && request.PaymentMethod != nil {
		if request.PaymentMethod.Type == constant.UnifiedPaymentMethodQris &&
			paymentMethod != nil &&
			!strings.EqualFold(paymentMethod.Acquirer, constant.BANK_ACQUIRER_BNC) {
			return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentStaticQrAcquirerNotSupported)
		}

		countStaticPayment, errCount := s.paymentRepo.CountActiveStaticPayment(ctx, request.MerchantID, paymentMethodID)
		if errCount != nil {
			return nil, pkgErr.New(response.HttpErrDatabase, errCount)
		}

		maxQRLimit := paymentPartnerConfig.GetMaxBNCQRStaticLimit()

		if maxQRLimit == 0 {
			maxQRLimit = s.config.UnifiedPaymentConfig.QrConfig.MaxActiveStaticQRPerMerchant
		}

		if request.PaymentMethod.Type == constant.UnifiedPaymentMethodQris &&
			countStaticPayment >= maxQRLimit {
			return nil, pkgErr.New(response.HttpErrRequest, constant.ErrPaymentStaticQrReachMaxActivePayment)
		}
	}

	if err = s.paymentRepo.CreatePayment(ctxTrx, &paymentDTO); err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	switch paymentSessionStatus {
	case constant.UnifiedPaymentSessionStatusRequirePaymentMethod:
		s.RecordChargeStatusHistory(ctx, request.PaymentID, constant.StatusHistoryActorUser, constant.ChargeStatusHistoryWaitingForUserAction)
		s.paymentSvc.RecordPaymentStatusHistory(ctx, request.PaymentID, constant.StatusHistoryActorUser, constant.PaymentStatusHistoryRequirePaymentMethod)
	case constant.UnifiedPaymentSessionStatusRequireConfirmation:
		s.RecordChargeStatusHistory(ctx, request.PaymentID, constant.StatusHistoryActorUser, constant.ChargeStatusHistoryWaitingForUserAction)
		s.paymentSvc.RecordPaymentStatusHistory(ctx, request.PaymentID, constant.StatusHistoryActorUser, constant.PaymentStatusHistoryRequireConfirmation)
	case constant.UnifiedStaticPaymentStatusActive:
		s.RecordChargeStatusHistory(ctx, request.PaymentID, constant.StatusHistoryActorUser, constant.ChargeStatusHistoryWaitingForUserAction)
	}

	if request.RecurringID != "" && request.RecurringStatus == constant.RecurringContractStatusCreated {
		err = s.recurringContractRepo.UpdateRecurringContractStatus(ctxTrx, request.RecurringID, constant.RecurringContractStatusPendInitialAuth, constant.UserSystemType)
		if err != nil {
			s.logger.Error(ctx, "Failed to update the recurring payment contract status", logger.Error(err))
			return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
		}
	}

	// If autoConfirm, request to processor
	if request.AutoConfirm {
		request.Amount.Value = totalAmount.Round(2).InexactFloat64()
		if chargeDetails, err = s.processConfirm(ctxTrx,
			&unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest{
				PaymentSessionID:           request.PaymentID,
				PaymentMethod:              request.PaymentMethod,
				PaymentMethodOptions:       &request.PaymentMethodOptions,
				ClientReferenceID:          request.ClientReferenceID,
				Fee:                        fee,
				Amount:                     request.Amount,
				ExpiryAt:                   request.ExpiryAt,
				StatementDescriptor:        request.StatementDescriptor,
				RecurringID:                request.RecurringID,
				InitiateFirstAuthorization: request.InitiateFirstAuthorization,
				FirstAuthorizationMethod:   request.FirstAuthorizationMethod,
				FirstAuthorizationOrderID:  request.FirstAuthorizationOrderID,
				RecurringBillingCycle:      request.RecurringBillingCycle,
				VirtualTerminal:            request.VirtualTerminal,
				CardFundedPayout:           request.CardFundedPayout,
				OneDollarAuthorization:     request.OneDollarAuthorization,
				AutoSplitPayment:           request.AutoSplitPayment,

				MerchantID:               merchantID,
				ParentMerchantID:         request.ParentMerchantID,
				MerchantExternalID:       merchantExternalID,
				DerivedMID:               derivedMID,
				DerivedMerchantID:        derivedMerchantID,
				DerivedMerchantShortname: derivedMerchantShortname,
				PaymentMethodChannelType: paymentMethodChannelType,
				Mode:                     request.Mode,
				PaymentPartnerConfig:     paymentPartnerConfig,
				RedirectUrl:              util.ValueOfPtr(unifiedPaymentMetadata.RedirectPaymentUIUrl),
				PaymentType:              request.PaymentType,
				UnifiedPaymentMetadata:   unifiedPaymentMetadata,
			},
			&paymentDTO,
		); err != nil {
			return nil, err
		}
	}

	// Commit trx
	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, errCommit)
	}
	isCompleted = true
	isCreatedPayment = true

	// Publish expiry use plugin delayed message
	s.publishExpiryMessage(ctx, request)

	// Passing response
	if unifiedPaymentResp.PaymentMethod != nil && unifiedPaymentResp.PaymentMethod.CardPaymentMethodDetail != nil {
		unifiedPaymentResp.PaymentMethod.CardPaymentMethodDetail.CVC = ""
	}
	unifiedPaymentResp.CreatedAt = paymentDTO.CreatedAt
	unifiedPaymentResp.UpdatedAt = paymentDTO.UpdatedAt
	unifiedPaymentResp.Status = paymentDTO.Status

	// dana api mode need the initiate frontend url to trigger processing event
	if request.Mode == constant.UnifiedPaymentModeRedirect ||
		(request.IsAPIMode() && request.AutoConfirm && request.GetEwalletChannel() == constant.UnifiedPaymentEWalletDanaAcquirer) {
		unifiedPaymentResp.PaymentUrl = request.PaymentURL
	}
	if chargeDetails != nil {
		unifiedPaymentResp.ChargeDetails = append(unifiedPaymentResp.ChargeDetails, chargeDetails)
	}

	if unifiedPaymentResp.IsSubsequentRecurringPayment() {
		unifiedPaymentResp.Status = constant.UnifiedPaymentSessionStatusProcessing
	}

	if unifiedPaymentMetadata.MethodDetail != nil {
		if unifiedPaymentMetadata.MethodDetail.Qr != nil {
			unifiedPaymentResp.PaymentMethod.QrPaymentMethodDetail = unifiedPaymentMetadata.MethodDetail.Qr
		}

		if unifiedPaymentMetadata.MethodDetail.VirtualAccount != nil {
			unifiedPaymentResp.PaymentMethod.VAPaymentMethodDetail = unifiedPaymentMetadata.MethodDetail.VirtualAccount
		}
	}

	unifiedPaymentResp.SetPaymentURLForAPIMode()
	unifiedPaymentResp.ShortPaymentUrl = request.ShortPaymentURL
	unifiedPaymentResp.SetPaymentSimulationForStaging(s.config)

	isFirstPaymentForCardFundedPayout := request.CardFundedPayout != nil &&
		request.CardFundedPayout.Sequence == 1 && len(unifiedPaymentResp.ChargeDetails) > 0
	if isFirstPaymentForCardFundedPayout {
		unifiedPaymentResp.PaymentUrl = util.ValueOfPtr(unifiedPaymentResp.ChargeDetails[0].Card).ACSURL
	}

	for _, chargeResp := range unifiedPaymentResp.ChargeDetails {
		chargeResp.RemoveUnusedResponse()
	}

	return unifiedPaymentResp, nil
}

func (s *UnifiedPaymentService) generatePaymentToken(ctx context.Context, paymentID string, expiryAt time.Time) (string, error) {
	// the expiry time following the expiring backoff time
	// Calculate total backoff time from config with minute buffer
	totalBackoffMinutes := 10
	for _, backoff := range s.config.UnifiedPaymentConfig.ExpiringProcessedBackoffMinutes {
		totalBackoffMinutes += backoff
	}
	expiryAt = expiryAt.Add(time.Duration(totalBackoffMinutes) * time.Minute)
	tokenExpiry := expiryAt.Sub(time.Now().UTC())

	// Generate JWT token
	token, err := s.jwt.GeneratePaymentToken(paymentID, expiryAt)
	if err != nil {
		s.logger.Error(ctx, "error generate payment token", logger.Error(err))
		return "", err
	}

	// Store hashed 256 jwt token to redis
	if err = s.redis.Set(ctx, fmt.Sprintf(constant.PaymentTokenCacheKey, util.HashString(token)), true, tokenExpiry).Err(); err != nil {
		s.logger.Error(ctx, "error set payment token", logger.Error(err))
		return "", err
	}

	return token, nil
}

func (s *UnifiedPaymentService) publishExpiryMessage(ctx context.Context, request *unifiedPaymentModel.CreateUnifiedPaymentSessionRequest) {
	if request.ExpiryAt.IsZero() {
		return
	}

	// Set lastPublishExpiryAt, equals to 01.00 JKT time
	now := time.Now().UTC()
	lastPublishExpiryAt := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
	if now.Hour() >= 18 && now.Hour() < 24 {
		lastPublishExpiryAt = lastPublishExpiryAt.Add(24 * time.Hour)
	}

	if request.ExpiryAt.Before(lastPublishExpiryAt) {
		if err := s.rabbitMqExt.PublishWithDelay(
			ctx,
			rabbitMqExt.PaymentExpirationRoutingKey,
			&paymentModel.ExpiringPayment{
				UUID:       request.PaymentID,
				MerchantID: request.MerchantID,
				ExpiredAt:  request.ExpiryAt,
			},
			request.ExpiryAt.Sub(now),
		); err != nil {
			s.logger.Error(ctx, "error publish payment expiration message", logger.Error(err))
		}
	}
}
