package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) ConfirmSession(ctx context.Context, request *unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/ConfirmSession")
	defer segment.End()

	var (
		merchantID               = request.MerchantID
		merchantExternalID       string
		derivedMID               string
		derivedMerchantID        string
		derivedMerchantShortname string

		chargeDetails *unifiedPaymentModel.ChargeResponse
	)

	// Set default payment method channel type
	request.PaymentMethodChannelType = constant.PaymentMethodChannelTypeAggregator

	// Get payment session by id
	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentSessionID)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	allowedStatuses := []string{
		constant.UnifiedPaymentSessionStatusRequireConfirmation,
		constant.UnifiedPaymentSessionStatusRequirePaymentMethod,
		constant.UnifiedStaticPaymentStatusActive,
	}

	paymentMetadata := payment.ToUnifiedPaymentMetadata()
	if paymentMetadata != nil {
		if util.ValueOfPtr(paymentMetadata.RetryableConfirmation) {
			allowedStatuses = append(allowedStatuses, constant.UnifiedPaymentSessionStatusRequireAction)
		}
		request.AutoSplitPayment = paymentMetadata.AutoSplitPayment
	}

	if payment.MerchantID != request.MerchantID {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantIsNotMatch)
	} else if !slices.Contains(allowedStatuses, payment.Status) || util.ValueOfPtr(payment.ProcessorReferenceNumber) != "" {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrNotAllowedToConfirmPaymentSession)
	} else if payment.PaymentMethodID == "" && request.PaymentMethod == nil {
		return nil, pkgErr.New(response.HttpErrRequest, constant.ErrConfirmShouldChoosePaymentMethod)
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
	derivedMerchantID = merchant.UUID

	if merchant.ParentID.Valid {
		parentMerchant, err := s.merchantRepo.FindMerchantByID(ctx, merchant.ParentID.String)
		if err != nil {
			return nil, pkgErr.New(response.HttpErrDatabase, constant.ErrFindParentMerchant)
		}

		if merchant.KYCStatus.String != constant.KYCStatusApproved {
			derivedMID = parentMerchant.MID.String
			derivedMerchantShortname = parentMerchant.ShortName
			derivedMerchantID = parentMerchant.UUID
			merchantExternalID = parentMerchant.ExternalId
		}
	}

	// Convert map to JSON
	unifiedPaymentResp := payment.ToUnifiedPaymentResponse()

	// Passing payment data to request
	request.MerchantExternalID = merchantExternalID
	request.DerivedMID = derivedMID
	request.DerivedMerchantID = derivedMerchantID
	request.DerivedMerchantShortname = derivedMerchantShortname
	request.Amount = unifiedPaymentResp.Amount
	request.StatementDescriptor = unifiedPaymentResp.StatementDescriptor
	request.Mode = unifiedPaymentResp.Mode
	if payment.ReferenceID != nil {
		request.ClientReferenceID = *payment.ReferenceID
	}
	if payment.ExpiredAt != nil {
		request.ExpiryAt = *payment.ExpiredAt
	}

	if request.PaymentMethod == nil {
		request.PaymentMethod = unifiedPaymentResp.PaymentMethod
	}

	basePaymentMethodOptions := unifiedPaymentResp.PaymentMethodOptions
	if request.PaymentMethodOptions != nil {
		// Merge base payment method options with request payment method options
		request.PaymentMethodOptions = request.PaymentMethodOptions.Merge(&basePaymentMethodOptions)
	} else {
		// If request doesn't have PaymentMethodOptions, use base options
		request.PaymentMethodOptions = &basePaymentMethodOptions
	}

	if request.PaymentType == "" {
		request.PaymentType = unifiedPaymentResp.PaymentType
	}

	acquirer := ""
	switch request.PaymentMethod.Type {
	case constant.UnifiedPaymentMethodVA:
		acquirer = request.PaymentMethodOptions.VirtualAccount.Channel

	case constant.UnifiedPaymentMethodEWallet:
		acquirer = strings.ToLower(request.PaymentMethodOptions.Ewallet.Channel)
	}

	splitRoute := false
	if payment.Metadata != nil && (*payment.Metadata)["splitRoutingConfigurations"] != nil {
		routes, _ := (*payment.Metadata)["splitRoutingConfigurations"].([]any)

		splitRoute = len(routes) > 0
	}
	paymentMethod, err := s.validatePaymentActivation(ctx, derivedMerchantID, merchantExternalID, request.PaymentMethod.Type, acquirer, splitRoute)
	if err != nil {
		return nil, err
	}
	payment.PaymentMethodID = paymentMethod.UUID
	request.PaymentMethodChannelType = paymentMethod.ChannelType
	cardSupportedUseCases := []*paymentMethodModel.CardSupportedUseCase{}
	if paymentMethod.MerchantConfigObj != nil {
		paymentPartnerConfig := paymentMethod.MerchantConfigObj.PartnerConfig
		request.PaymentPartnerConfig = paymentPartnerConfig

		if paymentPartnerConfig != nil && paymentPartnerConfig.Card != nil {
			for _, item := range paymentPartnerConfig.Card.Items {
				cardSupportedUseCases = append(cardSupportedUseCases, item.SupportedUseCase)
			}
		}
	}

	if request.PaymentMethod.Type == constant.UnifiedPaymentMethodCard && request.PaymentMethodOptions.Card != nil {
		s.RecordChargeStatusHistory(ctx, request.PaymentSessionID, constant.StatusHistoryActorUser, constant.ChargeStatusHistoryWaitingForAuthentication)
		if err = s.validateCard(ctx, &unifiedPaymentModel.ValidateCardRequest{
			IsConfirmStep:            true,
			Mode:                     unifiedPaymentResp.Mode,
			CardPaymentMethod:        request.PaymentMethod.CardPaymentMethodDetail,
			CardPaymentMethodOptions: request.PaymentMethodOptions.Card,
			CardSupportedUseCases:    cardSupportedUseCases,
			IsRecurringPayment:       request.RecurringID != "",
			IsVirtualTerminal:        request.VirtualTerminal != nil,
			IsAutoSplitPayment:       request.AutoSplitPayment != nil,
			HasCardOnFile:            request.HasCardOnFile(),
		}, paymentMethod); err != nil {
			return nil, err
		}
	}

	// Override Metadata
	(*payment.Metadata)["paymentMethod"] = request.PaymentMethod
	(*payment.Metadata)["paymentMethodOptions"] = request.PaymentMethodOptions

	// Handle Redirect URL for Card and EWallet
	if slices.Contains([]string{constant.UnifiedPaymentMethodCard, constant.UnifiedPaymentMethodEWallet}, request.PaymentMethod.Type) {
		token := ""
		if parsedURL, errParsed := url.Parse(payment.PaymentURL); errParsed == nil {
			token = parsedURL.Query().Get("token")
		}

		(*payment.Metadata)["redirectUrl"] = &unifiedPaymentModel.RedirectPaymentUIUrl{
			SuccessUrl: fmt.Sprintf(s.config.PaymentUIConfig.PaymentSuccessURL, token),
			FailedUrl:  fmt.Sprintf(s.config.PaymentUIConfig.PaymentFailedURL, token),
		}

		if request.Mode == constant.UnifiedPaymentModeAPI {
			(*payment.Metadata)["redirectUrl"] = &unifiedPaymentModel.RedirectPaymentUIUrl{
				SuccessUrl: unifiedPaymentResp.RedirectUrl.SuccessReturnUrl,
				FailedUrl:  unifiedPaymentResp.RedirectUrl.FailureReturnUrl,
			}
		}
		if request.AutoSplitPayment != nil {
			(*payment.Metadata)["redirectUrl"].(*unifiedPaymentModel.RedirectPaymentUIUrl).ProcessingUrl = fmt.Sprintf(s.config.PaymentUIConfig.PaymentProcessingURL, token)
		}
		request.RedirectUrl = util.ValueOfPtr((*payment.Metadata)["redirectUrl"].(*unifiedPaymentModel.RedirectPaymentUIUrl))
	}

	if request.IsStaticPayment() && payment.PaymentMethodID != "" && request.PaymentMethod != nil {
		countStaticPayment, errCount := s.paymentRepo.CountActiveStaticPayment(ctx, request.MerchantID, payment.PaymentMethodID)
		if errCount != nil {
			return nil, pkgErr.New(response.HttpErrDatabase, errCount)
		}

		partnerCfg := util.ValueOfPtr(request.PaymentPartnerConfig)
		maxQRLimit := partnerCfg.GetMaxBNCQRStaticLimit()
		if maxQRLimit == 0 {
			maxQRLimit = s.config.UnifiedPaymentConfig.QrConfig.MaxActiveStaticQRPerMerchant
		}

		if request.PaymentMethod.Type == constant.UnifiedPaymentMethodQris &&
			countStaticPayment >= maxQRLimit {

			// Update status to INACTIVE
			payment.UpdatedAt = time.Now().UTC()
			payment.Status = constant.UnifiedStaticPaymentStatusInactive
			if err = s.paymentRepo.UpdatePaymentData(ctx, payment.ToDTO()); err != nil {
				return nil, pkgErr.New(response.HttpErrDatabase, err)
			}

			return nil, pkgErr.New(response.HttpErrRequest, constant.ErrPaymentStaticQrReachMaxActivePayment)
		}
	}

	var unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	metadataB, _ := json.Marshal(payment.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)
	request.UnifiedPaymentMetadata = &unifiedPaymentMetadata

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

	if chargeDetails, err = s.processConfirm(ctx, request, payment.ToDTO()); err != nil {
		return nil, err
	}

	// Commit trx
	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, errCommit)
	}
	isCompleted = true

	// Find payment
	payment, err = s.paymentRepo.GetPaymentById(ctx, request.PaymentSessionID)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	resp := payment.ToUnifiedPaymentResponse()

	if chargeDetails != nil {
		resp.ChargeDetails = append(resp.ChargeDetails, chargeDetails)
	}

	if unifiedPaymentMetadata.MethodDetail != nil {
		if unifiedPaymentMetadata.MethodDetail.Qr != nil {
			resp.PaymentMethod.QrPaymentMethodDetail = unifiedPaymentMetadata.MethodDetail.Qr
		}

		if unifiedPaymentMetadata.MethodDetail.VirtualAccount != nil {
			resp.PaymentMethod.VAPaymentMethodDetail = unifiedPaymentMetadata.MethodDetail.VirtualAccount
		}
	}

	resp.SetPaymentURLForAPIMode()
	resp.SetPaymentSimulationForStaging(s.config)

	for _, chargeResp := range resp.ChargeDetails {
		chargeResp.RemoveUnusedResponse()
	}

	return resp, nil
}

func (s *UnifiedPaymentService) processConfirm(ctx context.Context, request *unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest, paymentDTO *paymentModel.PaymentDTO) (data *unifiedPaymentModel.ChargeResponse, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/processConfirm")
	defer segment.End()

	defer func() {
		metricData := monitoring.CustomMetric{
			ComponentName:        constant.ComponentNameUnifiedPayment,
			MetricName:           constant.MetricNameUnifiedPaymentConfirmSession,
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			Attributes: map[string]any{
				"merchantId":          request.ParentMerchantID,
				"onBehalfSubmerchant": request.ParentMerchantID != request.MerchantID,
				"mode":                request.Mode,
				"recurringPayment":    request.RecurringID != "",
				"virtualTerminal":     request.VirtualTerminal != nil,
				"cardFundedPayout":    request.CardFundedPayout != nil,
			},
		}
		if request.PaymentMethod != nil {
			metricData.Attributes["paymentMethod"] = request.PaymentMethod.Type
			metricData.Attributes["paymentMethodTypeDetail"] = request.PaymentMethod.GetPaymentMethodTypeDetail()
			if request.PaymentMethodOptions != nil {
				metricData.Attributes["paymentMethodOptionDetail"] = request.PaymentMethodOptions.GetPaymentMethodOptionDetail()
			}
		}
		if err != nil {
			errType, errDetail := pkgErr.ExtractError(err)
			metricData.Attributes["errorDetail"] = errDetail.Error()
			metricData.Attributes["errorType"] = errType
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			s.logger.Warn(ctx, "failed to record confirmed unified payment custom metric", logger.Error(errMetric), logger.Any("metricData", metricData))
		}

	}()

	// Confirm should have chosen payment method
	if request.PaymentMethod == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrConfirmShouldChoosePaymentMethod)
	}

	// Init charge ID
	chargeID := uuid.New()
	accountTransactions, err := s.accountTransactionRepo.FindByReference(ctx, request.PaymentSessionID, constant.TypePayment)
	if err != nil {
		s.logger.Error(ctx, "failed to find account transaction by reference", logger.Error(err), logger.String("payment_session_id", request.PaymentSessionID), logger.String("payment_method_type", request.PaymentMethod.Type))
	}
	if accountTransactions != nil {
		chargeID = accountTransactions.UUID
		request.SkipInsertAccountTransaction = true
		s.logger.Info(ctx, "skipping account transaction creation, using existing transaction", logger.String("charge_id", chargeID.String()))
	}
	chargeIDStr := chargeID.String()
	if request.IsStaticPayment() {
		chargeIDStr = ""
	}

	// Get FDS Config
	fdsConfig, err := s.merchantSvc.GetFDSConfig(ctx, request.DerivedMerchantID)
	if err != nil && !strings.Contains(err.Error(), constant.ErrMerchantNotFound.Error()) {
		return nil, err
	} else if fdsConfig != nil {
		request.BypassExternalFDS = fdsConfig.FDSConfig.BypassExternalPaymentCheck
	}

	// Create initiated charge
	if err = s.createInitiatedCharge(ctx, chargeID, request); err != nil {
		return nil, pkgErr.New(response.HttpErrInternal, constant.ErrUpdateUnifiedPaymentSessionLedger)
	}

	// Call processor request
	isSnap := false
	if request.UnifiedPaymentMetadata != nil && request.UnifiedPaymentMetadata.IsSnap != nil {
		isSnap = *request.UnifiedPaymentMetadata.IsSnap
	}

	initProcessorRequest := &unifiedPaymentModel.BaseProcessorRequest{
		PaymentMethod:            request.PaymentMethod,
		Fee:                      request.Fee,
		Amount:                   request.Amount,
		ExpiryAt:                 request.ExpiryAt,
		PaymentMethodType:        request.PaymentMethod.Type,
		PaymentMethodOptions:     request.PaymentMethodOptions,
		PaymentMethodChannelType: request.PaymentMethodChannelType,
		PaymentPartnerConfig:     request.PaymentPartnerConfig,

		PaymentID:                request.PaymentSessionID,
		ClientReferenceID:        request.ClientReferenceID,
		ChargeID:                 chargeIDStr,
		MerchantID:               request.MerchantID,
		MerchantExternalID:       request.MerchantExternalID,
		DerivedMID:               request.DerivedMID,
		DerivedMerchantID:        request.DerivedMerchantID,
		DerivedMerchantShortName: request.DerivedMerchantShortname,
		SuccessRedirectUrl:       request.RedirectUrl.SuccessUrl,
		Mode:                     request.Mode,
		PaymentURL:               paymentDTO.PaymentURL,
		IsStaticPayment:          request.IsStaticPayment(),
		IsSnap:                   isSnap,
		// Recurring Payment
		RecurringID:                request.RecurringID,
		InitiateFirstAuthorization: request.InitiateFirstAuthorization,
		FirstAuthorizationMethod:   request.FirstAuthorizationMethod,
		FirstAuthorizationOrderID:  request.FirstAuthorizationOrderID,
		RecurringBillingCycle:      request.RecurringBillingCycle,
		// Card Funded Payout
		CardFundedPayout: request.CardFundedPayout,
		// Auto Split Payment
		AutoSplitPayment: request.AutoSplitPayment,
	}
	if request.HasCardOnFile() {
		initProcessorRequest.CardOnFile = &unifiedPaymentModel.CardOnFile{
			Initiator:                    request.PaymentMethodOptions.Card.CardOnFile.Initiator,
			Type:                         request.PaymentMethodOptions.Card.CardOnFile.Type,
			PreviousNetworkTransactionID: request.PaymentMethodOptions.Card.CardOnFile.PreviousNetworkTransactionID,
		}
	}
	processor, err := s.ProcessorInitialization(ctx, initProcessorRequest)
	if err != nil {
		return nil, err
	}

	// If multi-acquirer routing is enabled and processor returned a different acquirer,
	// find the correct payment method that matches the actual acquirer used
	if constant.IsQrMultiAcquirerRoutingEnabled(request.MerchantID) {
		// For QRIS payments, extract actual acquirer from processor response
		if processor.Qr != nil && processor.Qr.Acquirer != "" &&
			request.PaymentPartnerConfig != nil && request.PaymentPartnerConfig.Qris != nil {
			// Clean acquirer name (BNC_QRIS -> bnc)
			responseAcquirer := util.CleanAcquirerName(processor.Qr.Acquirer)
			originalAcquirer := strings.ToLower(request.PaymentPartnerConfig.Qris.Acquirer)

			// Compare with original payment method acquirer
			if responseAcquirer != originalAcquirer {
				// Find payment method by: category=PAYMENT, type=QRIS, acquirer=<from response>
				matchedPaymentMethod, err := s.paymentMethodRepo.GetPaymentMethodByCategoryTypeAndAcquirer(
					ctx,
					paymentConstant.PAYMENT_METHOD_CATEGORY_PAYMENT,
					paymentConstant.PAYMENT_METHOD_QRIS,
					responseAcquirer,
				)

				if err != nil {
					s.logger.Error(ctx, "error finding payment method by acquirer from processor response",
						logger.Error(err),
						logger.String("rawAcquirer", processor.Qr.Acquirer),
						logger.String("cleanedAcquirer", responseAcquirer),
						logger.String("originalAcquirer", originalAcquirer))
				} else if matchedPaymentMethod != nil {
					s.logger.Info(ctx, "using payment method matched from processor acquirer response",
						logger.String("originalPaymentMethodID", paymentDTO.PaymentMethodID),
						logger.String("originalAcquirer", originalAcquirer),
						logger.String("matchedPaymentMethodID", matchedPaymentMethod.UUID),
						logger.String("rawAcquirer", processor.Qr.Acquirer),
						logger.String("cleanedAcquirer", responseAcquirer))

					// Update payment DTO with matched payment method ID
					paymentDTO.PaymentMethodID = matchedPaymentMethod.UUID
				}
			}
		}
	}

	statementDescriptor := processor.StatementDescriptor
	if statementDescriptor == "" {
		statementDescriptor = request.DerivedMerchantShortname
	}

	if request.IsStaticPayment() {
		request.UnifiedPaymentMetadata.MethodDetail = processor
		if metaDataB, errMarshal := json.Marshal(request.UnifiedPaymentMetadata); errMarshal != nil {
			return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrInvalidMarshalData)
		} else if metaDataB != nil {
			metadataStr := string(metaDataB)
			paymentDTO.Metadata = &metadataStr
		}
	} else {
		if request.VirtualTerminal != nil {
			s.paymentSvc.RecordPaymentStatusHistory(ctx, request.PaymentSessionID, constant.StatusHistoryActorUser, constant.PaymentStatusHistoryProcessing)
		} else {
			paymentDTO.Status = constant.UnifiedPaymentSessionStatusRequireAction
			s.paymentSvc.RecordPaymentStatusHistory(ctx, request.PaymentSessionID, constant.StatusHistoryActorSystem, constant.PaymentStatusHistoryRequireAction)
		}
		accountTrxMetadata := orchestratorModel.MetadataPayment[any]{
			ReconReferenceNo:    processor.ProcessorReferenceNo,
			ExpiredAt:           processor.ProcessorExpiredAt,
			MethodDetail:        processor,
			StatementDescriptor: statementDescriptor,
		}

		updateRequest := orchestratorModel.UpdatePaymentTransactionRequest{
			LedgerId:               chargeIDStr,
			TransactionTimestamp:   processor.ProcessorTransactionTimestamp,
			ProcessorReferenceName: processor.ProcessorReference,
			ProcessorReferenceId:   processor.ProcessorReferenceID,
			MethodDetail:           processor,
		}
		if err = s.accountTransactionRepo.UpdatePaymentTransactionStatusAndMetadataByID(
			ctx, updateRequest, accountTrxMetadata); err != nil {
			s.logger.Error(ctx, "failed to update payment transaction status and metadata", logger.Error(err))
			return nil, pkgErr.New(response.HttpErrDatabase, err)
		}
	}

	// Update payment processor reference number
	paymentDTO.UpdatedAt = time.Now().UTC()
	paymentDTO.ProcessorReferenceNumber = &processor.ProcessorReferenceNo
	if err = s.paymentRepo.UpdatePaymentData(ctx, paymentDTO); err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	resp := &unifiedPaymentModel.ChargeResponse{
		ID:                              chargeID.String(),
		PaymentSessionID:                request.PaymentSessionID,
		PaymentSessionClientReferenceID: request.ClientReferenceID,
		Amount:                          request.Amount,
		StatementDescriptor:             statementDescriptor,
		Status:                          constant.ChargeStatusWaitingForUserAction,
		CreatedAt:                       time.Now().UTC(), // Need to update with actual value
		UpdatedAt:                       time.Now().UTC(), // Need to update with actual value
		ChargePaymentMethodDetails:      processor,
	}

	// Make charge null for static payment
	if request.IsStaticPayment() {
		resp = nil
	}

	return resp, nil
}

func (s *UnifiedPaymentService) createInitiatedCharge(ctx context.Context, chargeID uuid.UUID, request *unifiedPaymentModel.ConfirmUnifiedPaymentSessionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/createInitiatedCharge")
	defer segment.End()

	if request.IsStaticPayment() {
		s.logger.Info(ctx, "no need to create initiated charge for static payment")
		return nil
	}

	// Insert into account_transactions as charge
	accountTrxMetadata := orchestratorModel.MetadataPayment[any]{
		ChargeStatus:      constant.ChargeStatusWaitingForUserAction,
		BypassExternalFDS: request.BypassExternalFDS,
	}

	rawAccountTrxMetadata, _ := json.Marshal(accountTrxMetadata)
	trxRequest := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 chargeID,
		ReferenceID:          request.PaymentSessionID,
		Type:                 constant.TypePayment,
		MerchantID:           util.ParseUUID(request.MerchantID),
		Currency:             request.Amount.Currency,
		Credit:               request.Amount.Value,
		Channel:              request.PaymentMethod.Type,
		Status:               constant.StatusPending,
		TransactionTimestamp: time.Now().UTC(),
		SettlementStatus:     util.ValueToPtr(constant.StatusPending),
		Usecase:              constant.TypePayment,
		AdditionalInfo: types.NullJSONText{
			Valid: true, JSONText: rawAccountTrxMetadata,
		},
		SettlementModel: util.ValueToPtr(request.PaymentMethodChannelType),
	}
	if request.VirtualTerminal != nil {
		trxRequest.Usecase = constant.TypeVirtualTerminal
	} else if request.CardFundedPayout != nil || (request.OneDollarAuthorization != nil && request.OneDollarAuthorization.UseCase == constant.UnifiedPaymentUseCaseCardFundedPayoutSavedCards) {
		trxRequest.Usecase = constant.TypePaymentFundedPayout
	}
	if request.IsAutoSplitPaymentAuth() {
		trxRequest.Credit = 0
	}

	if request.IsAutoSplitSubPayments() {
		trxRequest.Reference = constant.ReferenceSubPayment
	}

	// Handle FACILITATOR settlement model
	if constant.IsDirectPSP(request.PaymentMethodChannelType) {
		trxRequest.SettlementStatus = nil
		trxRequest.SettlementModel = util.ValueToPtr(constant.PaymentMethodChannelTypeDirect)
	}

	// Set charge amount to 0 for captureMethod = MANUAL
	if request.PaymentMethodOptions.Card != nil && strings.ToUpper(request.PaymentMethodOptions.Card.CaptureMethod) == constant.UnifiedPaymentCardCaptureMethodManual {
		trxRequest.Credit = 0
	}

	// Action For Post Transaction
	if request.SkipInsertAccountTransaction {
		err := s.orchestratorSvc.UpdateTransaction(ctx, &orchestratorModel.UpdateTransactionRequest{
			TransactionID: chargeID.String(),
			Channel:       request.PaymentMethod.Type,
		})
		if err != nil {
			s.logger.Error(ctx, "error when update existing unified payment charges", logger.Error(err), logger.String("ledgerId", chargeID.String()))
			return err
		}
		return nil
	}
	if err := s.orchestratorSvc.PostAccountTransaction(ctx, trxRequest); err != nil {
		return err
	}

	return nil
}
