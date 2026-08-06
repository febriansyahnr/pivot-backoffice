package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	messagingQueueModel "github.com/paper-indonesia/pivot-backoffice/internal/model/messagingQueue"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	qrSnapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	vaSnapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/shopspring/decimal"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) PostCreateLedger(ctx context.Context, payment *paymentModel.Payment, request *paymentModel.PostCreateLedgerRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/PostCreateLedger")
	defer segment.End()

	// Find merchant data
	derivedMerchantID := payment.MerchantID
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, derivedMerchantID)
	if err != nil {
		return pkgErr.New(response.HttpErrDatabase, constant.ErrFindMerchant)
	} else if merchant == nil {
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	if merchant.ParentID.Valid {
		parentMerchant, err := s.merchantRepo.FindMerchantByID(ctx, merchant.ParentID.String)
		if err != nil {
			return pkgErr.New(response.HttpErrDatabase, constant.ErrFindParentMerchant)
		}
		if merchant.KYCStatus.String == constant.KYCStatusNotRequired {
			derivedMerchantID = parentMerchant.UUID
		}
	}

	paymentMethod, err := s.paymentMethodSvc.FindPaymentMethodByIdAndMerchant(ctx, payment.PaymentMethodID, derivedMerchantID)
	if err != nil {
		s.logger.Error(ctx, "error when FindPaymentMethodByIdAndMerchant", logger.Error(err))
		return err
	} else if paymentMethod == nil {
		s.logger.Error(ctx, "payment method not found", logger.Error(constant.ErrPaymentMethodNotFound))
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentMethodNotFound)
	}

	needSettlementProcess := true
	transactionTimestamp := payment.UpdatedAt
	if payment.TrxDatetime != nil {
		transactionTimestamp = *payment.TrxDatetime
	}

	paymentMethodType := payment.PaymentMethod.Type
	switch payment.Type {
	case constant.TypeVirtualTerminal:
		paymentMethodType = constant.TypeVirtualTerminal

	case constant.TypeCardFundedPayout:
		paymentMethodType = constant.TypeCardFundedPayout
	}
	settlementConfig := s.defineDefaultSettlementConfig(paymentMethodType, payment.PaymentMethod.Acquirer)

	trxMetadata := orchestratorModel.MetadataPayment[any]{
		ProcessorTransactionId: payment.ProcessorTransactionID,
		MethodDetail:           s.BindingPaymentMethodDetail(request.Channel, payment),
		ChargeStatus:           request.ChargeStatus,
	}

	if payment.ExpiredAt != nil {
		trxMetadata.ExpiredAt = *payment.ExpiredAt
	}
	if payment.ProcessorReferenceNumber != nil {
		trxMetadata.ReconReferenceNo = *payment.ProcessorReferenceNumber
	}
	if payment.Metadata != nil {
		if feeDetail, err := util.ConvertToStruct[feeModel.FeeMetadataObject]((*payment.Metadata)["feeDetail"]); err == nil {
			trxMetadata.FeeDetail = &feeDetail
		}
		if feeOnBehalf, err := util.ConvertToStruct[feeModel.TrxFeeOnBehalfMetadata]((*payment.Metadata)["feeOnBehalf"]); err == nil {
			trxMetadata.FeeOnBehalf = &feeOnBehalf
		}
	}

	chargeID := uuid.New()
	if request.ChargeID != "" {
		chargeID, _ = uuid.Parse(request.ChargeID)
	}

	merchantUUID, errParse := uuid.Parse(payment.MerchantID)
	if errParse != nil {
		return errParse
	}
	transactionAmount, _ := strconv.ParseFloat(request.Amount.Value, 64)
	transactionRequest := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                   chargeID,
		ReferenceID:            payment.UUID,
		Type:                   orchestratorModel.TypePayment,
		MerchantID:             merchantUUID,
		Currency:               request.Amount.Currency,
		Credit:                 transactionAmount,
		Debit:                  0.00,
		Channel:                request.Channel,
		Status:                 request.Status,
		SettlementStatus:       util.ValueToPtr(constant.StatusPending),
		Remarks:                "",
		TransactionTimestamp:   transactionTimestamp,
		Usecase:                constant.TypePayment,
		Processor:              payment.Processor,
		ProcessorID:            payment.ProcessorID,
		ProcessorTransactionID: payment.ProcessorTransactionID,
		SettlementModel:        util.ValueToPtr(constant.PaymentMethodChannelTypeAggregator),
	}

	// Set metadata for settlement with success status
	if request.Status == constant.StatusSuccess {

		transactionRequest.SettlementAt = &time.Time{}
		trxMetadata.SettlementDetail = &orchestratorModel.MetadataPaymentSettlementDetail{
			Type:   settlementConfig.Type,
			CutOff: settlementConfig.CutOff,
		}

		derivedMerchantId := payment.MerchantID
		if derivedID, ok := ctx.Value(constant.CtxDerivedMerchantID).(string); ok && derivedID != "" {
			derivedMerchantId = derivedID
		}
		getSettlementReq := merchantModel.GetSettlementConfigRequest{
			MerchantId: derivedMerchantId,
			Reference:  constant.TypePayment,
			Method:     &paymentMethodType,
			Channel:    util.ValueToPtr(strings.ToUpper(payment.PaymentMethod.Acquirer)),
		}
		if payment.CardFundedPayout != nil {
			getSettlementReq.Reference = constant.ReferencePaymentFundedPayout
			getSettlementReq.Method = util.ValueToPtr(paymentConstant.PAYMENT_METHOD_CREDIT_CARD)
			getSettlementReq.SettlementMethod = payment.CardFundedPayout.SettlementMethod
		}

		if payment.IsAutoSplitSubPayments() {
			transactionRequest.Reference = constant.ReferenceSubPayment
		}

		merchantSettlementConfig, err := s.getSettlementConfigWithParentFallback(
			ctx, payment, getSettlementReq,
		)
		if err != nil {
			s.logger.Error(ctx, "create new ledger - failed to get merchant settlement config", logger.Error(err))
			return pkgErr.New(response.HttpErrDatabase, err)

		} else if merchantSettlementConfig != nil {

			settlementConfig = merchantSettlementConfig

			trxMetadata.SettlementDetail = &orchestratorModel.MetadataPaymentSettlementDetail{
				Type:     settlementConfig.Type,
				CutOff:   settlementConfig.CutOff,
				IsOnHold: settlementConfig.IsOnHold,
			}
		}

		// Notes: need to attach `isOnHold` settlement config for multiple type payment because incoming notifications are attached to same payment ids.
		// Unlike other single payment notification which attached to a payment record. So after settlement hold, we need to store new incoming payment settlement config
		if payment.Type == constant.UnifiedPaymentTypeMultiple {
			settlementHold, err := s.settlementHoldSvc.GetSettlementHoldByPaymentID(ctx, payment.UUID)
			if err != nil {
				s.logger.Error(ctx, "error retrieve payment settlement hold, continue process", logger.Error(err))
			}
			if settlementHold != nil && settlementHold.Status == constant.SettlementHoldActionHold {
				trxMetadata.SettlementDetail.IsOnHold = true
			}
		}

		if trxMetadata.SettlementDetail.Type == constant.SettlementTypeInstant &&
			paymentMethod.ChannelType != constant.PaymentMethodChannelTypeDirect {
			transactionRequest.SettlementStatus = util.ValueToPtr(constant.StatusSuccess)
			transactionRequest.SettlementAt = util.ValueToPtr(time.Now().UTC())
		}

		// Note: When a transaction is the first authorization using the ONE_DOLLAR method or subsequent recurring payment with zero amount,
		//       the transaction is not subject to any fees and is settled immediately.
		if payment.IsFeeExempt() {
			trxMetadata.SettlementDetail = nil
			transactionRequest.SettlementAt = util.ValueToPtr(time.Now().UTC())
			transactionRequest.SettlementStatus = util.ValueToPtr(constant.StatusSuccess)
		}
		transactionRequest.AdditionalInfo.Valid = true
		transactionRequest.AdditionalInfo.JSONText, _ = json.Marshal(trxMetadata)
	}

	// Set settlement process and config to null for FACILITATOR model
	if paymentMethod.ChannelType == constant.PaymentMethodChannelTypeDirect {
		transactionRequest.SettlementModel = &paymentMethod.ChannelType
		needSettlementProcess = false
		transactionRequest.SettlementStatus = nil
		transactionRequest.SettlementAt = nil
	}

	if errOrch := s.orchestratorSvc.PostAccountTransaction(ctx, transactionRequest); errOrch != nil {
		s.logger.Error(ctx, "error when create account transaction", logger.Error(errOrch))

		return errOrch
	}

	// Record payment status history based on transaction status
	if request.Status == constant.StatusSuccess {
		s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorSystem, constant.PaymentStatusHistorySuccess)
	} else if request.Status == constant.StatusFailed {
		s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorSystem, constant.PaymentStatusHistoryFailed)
		return nil
	}

	if payment.IsFeeExempt() || payment.IsAutoSplitPaymentAuth() {
		return nil
	}

	feeTransactionID := uuid.New()
	if errFee := s.PostCreateFeeTransaction(ctx, payment, &paymentModel.PostCreateFeeTransactionRequest{
		SettlementTransactionMetadata: &settlementModel.AccountTransactionMetadataObject{
			SettlementDetail: *settlementConfig,
			ReconReferenceNo: trxMetadata.ReconReferenceNo,
		},
		FeeTransactionID:    feeTransactionID,
		LinkedTransactionID: transactionRequest.UUID,
		Status:              request.Status,
		Channel:             request.Channel,
		Currency:            request.Amount.Currency,
		TransactionAmount:   transactionAmount,
		SettlementStatus:    transactionRequest.SettlementStatus,
		SettlementAt:        transactionRequest.SettlementAt,
		SettlementModel:     transactionRequest.SettlementModel,
	}); errFee != nil {

		return errFee
	}

	newCtx := ctx
	if needSettlementProcess && trxMetadata.SettlementDetail.Type != constant.SettlementTypeInstant {
		newCtx = context.WithValue(ctx, constant.CtxSetPendingTransaction, true)
		newCtx = context.WithValue(newCtx, constant.CtxSetPendingSettlementTransaction, true)
	}

	if err = s.ProcessSplitRoute(newCtx, payment.UUID); err != nil {
		return err
	}

	if newCtx != ctx {
		s.publishPendingSettlementMessage(newCtx, transactionRequest.UUID.String(), feeTransactionID.String(), payment.MerchantID, settlementConfig)
	}

	return nil
}

func (s *PaymentService) UpdatePendingLedger(ctx context.Context, payment *paymentModel.Payment, request orchestratorModel.UpdatePaymentTransactionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/UpdatePendingLedger")
	defer segment.End()

	var needSettlementProcess = true

	request.TransactionTimestamp = request.UpdatedAt

	if request.TrxDatetime != nil {
		request.TransactionTimestamp = *request.TrxDatetime
	}

	request.SettlementStatus = util.ValueToPtr(constant.StatusPending)
	metadata := orchestratorModel.MetadataPayment[any]{
		ProcessorTransactionId: request.ProcessorTransactionId,
		MethodDetail:           request.MethodDetail,
		ReconReferenceNo:       payment.ReconReferenceNo,
	}
	if payment.Metadata != nil {
		feeDetail, err := util.ConvertToStruct[feeModel.FeeMetadataObject]((*payment.Metadata)["feeDetail"])
		if err == nil {
			metadata.FeeDetail = &feeDetail
		}

		feeOnBehalf, err := util.ConvertToStruct[feeModel.TrxFeeOnBehalfMetadata]((*payment.Metadata)["feeOnBehalf"])
		if err == nil {
			metadata.FeeOnBehalf = &feeOnBehalf
		}
	}

	if payment.AutoSplitPayment != nil && payment.AutoSplitPayment.Summary != nil && metadata.FeeDetail != nil {
		metadata.SubPaymentSummary = &orchestratorModel.MetadataSubPaymentSummary{
			TotalCreditAmount: payment.AutoSplitPayment.Summary.TotalSuccessfulChargeAmount.ToDecimal(),
			TotalFeeAmount:    decimal.NewFromFloat(metadata.FeeDetail.FinalAmount).Round(0),
		}
	}

	paymentMethodType := payment.PaymentMethod.Type
	switch payment.Type {
	case constant.TypeVirtualTerminal:
		paymentMethodType = constant.TypeVirtualTerminal

	case constant.TypeCardFundedPayout:
		paymentMethodType = constant.TypeCardFundedPayout
	}
	settlementConfig := s.defineDefaultSettlementConfig(paymentMethodType, payment.PaymentMethod.Acquirer)

	if request.Status == constant.StatusSuccess {
		metadata.ChargeStatus = constant.ChargeStatusSuccess
		request.SettlementAt = &time.Time{}
		metadata.SettlementDetail = &orchestratorModel.MetadataPaymentSettlementDetail{
			Type:   settlementConfig.Type,
			CutOff: settlementConfig.CutOff,
		}

		derivedMerchantId := payment.MerchantID
		if derivedID, ok := ctx.Value(constant.CtxDerivedMerchantID).(string); ok && derivedID != "" {
			derivedMerchantId = derivedID
		}
		getSettlementReq := merchantModel.GetSettlementConfigRequest{
			MerchantId: derivedMerchantId,
			Reference:  constant.TypePayment,
			Method:     &paymentMethodType,
			Channel:    util.ValueToPtr(strings.ToUpper(payment.PaymentMethod.Acquirer)),
		}
		if payment.CardFundedPayout != nil {
			getSettlementReq.Reference = constant.ReferencePaymentFundedPayout
			getSettlementReq.Method = util.ValueToPtr(paymentConstant.PAYMENT_METHOD_CREDIT_CARD)
			getSettlementReq.SettlementMethod = payment.CardFundedPayout.SettlementMethod
		}

		merchantSettlementConfig, err := s.getSettlementConfigWithParentFallback(
			ctx, payment, getSettlementReq,
		)
		if err != nil {
			s.logger.Error(ctx, "Update pending ledger - failed to get merchant settlement config", logger.Error(err))
			return pkgErr.New(response.HttpErrDatabase, err)

		} else if merchantSettlementConfig != nil {

			settlementConfig = merchantSettlementConfig

			metadata.SettlementDetail = &orchestratorModel.MetadataPaymentSettlementDetail{
				Type:   settlementConfig.Type,
				CutOff: settlementConfig.CutOff,
			}
		}

		// 3. Set settlement status
		if metadata.SettlementDetail.Type == constant.SettlementTypeInstant {
			request.SettlementStatus = util.ValueToPtr(constant.StatusSuccess)
			request.SettlementAt = util.ValueToPtr(time.Now().UTC())
		}

		// Note: When a transaction is the first authorization using the ONE_DOLLAR method or subsequent recurring payment with zero amount,
		//       the transaction is not subject to any fees and is settled immediately.
		if payment.IsFeeExempt() {
			metadata.SettlementDetail = nil
			request.SettlementAt = util.ValueToPtr(time.Now().UTC())
			request.SettlementStatus = util.ValueToPtr(constant.StatusSuccess)
		}

	} else if request.Status == constant.StatusFailed {
		metadata.ChargeStatus = constant.ChargeStatusFailed

	} else if payment.Status == constant.UnifiedPaymentSessionStatusProcessing {
		metadata.ChargeStatus = constant.ChargeStatusProcessing
	}

	if request.ChargeStatus != "" {
		metadata.ChargeStatus = request.ChargeStatus
	}

	if request.FailureCode != "" {
		metadata.FailureCode = request.FailureCode
	}

	// Set settlement process and config to null for FACILITATOR model
	if request.SettlementModel != nil && constant.IsDirectPSP(*request.SettlementModel) {
		settlementConfig = nil
		metadata.SettlementDetail = nil
		request.SettlementStatus = nil
		request.SettlementAt = nil

		needSettlementProcess = false
	}

	err := s.accountTransactionRepo.UpdatePaymentTransactionStatusAndMetadataByID(ctx, request, metadata)
	if err != nil {
		if errors.Is(err, constant.ErrDataNotFound) {
			return pkgErr.New(response.HttpErrUnprocessableContent, err)
		}
		s.logger.Error(ctx, "Update pending ledger - failed update payment transaction status and metadata", logger.Error(err))
		return pkgErr.New(response.HttpErrDatabase, err)
	}

	// Record payment status history based on transaction status update
	if request.Status == constant.StatusSuccess {
		s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorSystem, constant.PaymentStatusHistorySuccess)
	} else if request.Status == constant.StatusFailed {
		s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorSystem, constant.PaymentStatusHistoryFailed)
	} else if payment.Status == constant.UnifiedPaymentSessionStatusProcessing {
		s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorSystem, constant.PaymentStatusHistoryProcessing)
	}

	if request.Status == constant.StatusFailed || request.Status == constant.StatusPending || payment.IsFeeExempt() || payment.IsAutoSplitPaymentAuth() {
		return nil
	}

	// Find fee
	feeTrx, err := s.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.TypeFee)
	if err != nil {
		s.logger.Error(ctx, "Update pending ledger - failed to get fee transaction", logger.Error(err))
		return pkgErr.New(response.HttpErrDatabase, err)
	}

	var feeTransactionID uuid.UUID
	if feeTrx == nil {
		feeTransactionID = uuid.New()
		trxAmount, _ := decimal.NewFromString(request.Amount.Value)

		settlementTransactionMetadata := &settlementModel.AccountTransactionMetadataObject{
			ReconReferenceNo: metadata.ReconReferenceNo,
		}
		if settlementConfig != nil {
			settlementTransactionMetadata.SettlementDetail = *settlementConfig
		}

		err = s.PostCreateFeeTransaction(ctx, payment, &paymentModel.PostCreateFeeTransactionRequest{
			SettlementTransactionMetadata: settlementTransactionMetadata,
			FeeTransactionID:              feeTransactionID,
			LinkedTransactionID:           util.ParseUUID(request.LedgerId),
			Status:                        request.Status,
			Channel:                       request.Channel,
			Currency:                      request.Amount.Currency,
			TransactionAmount:             trxAmount.InexactFloat64(),
			SettlementStatus:              request.SettlementStatus,
			SettlementAt:                  request.SettlementAt,
			SettlementModel:               request.SettlementModel,
		})
		if err != nil {
			return pkgErr.New(response.HttpErrInternal, err)
		}
	} else {
		feeTransactionID = feeTrx.UUID
	}

	newCtx := ctx
	if needSettlementProcess && util.ValueOfPtr(metadata.SettlementDetail).Type != constant.SettlementTypeInstant {
		newCtx = context.WithValue(ctx, constant.CtxSetPendingTransaction, true)
		newCtx = context.WithValue(newCtx, constant.CtxSetPendingSettlementTransaction, true)
	}

	if err = s.ProcessSplitRoute(newCtx, payment.UUID); err != nil {
		return err
	}

	if newCtx != ctx {
		s.publishPendingSettlementMessage(newCtx, request.LedgerId, feeTransactionID.String(), payment.MerchantID, settlementConfig)
	}

	return nil
}

func (s *PaymentService) publishPendingSettlementMessage(ctx context.Context, transactionId, feeTransactionId, merchantId string, settlementConfig *merchantModel.SettlementConfig) {
	if settlementConfig.Type == constant.SettlementTypeInstant {
		s.logger.Info(ctx, "skip publishPendingSettlementMessage because it is INSTANT type")
		return
	}

	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/publishPendingSettlementMessage")
	defer segment.End()

	processTime := time.Now().UTC()
	if ok := s.tryProcessSettlementCutOff(ctx, transactionId, feeTransactionId, merchantId, processTime, settlementConfig); ok {
		return
	}

	settlementTime, _ := settlementConfig.GetSettlementTime(processTime) // error already handled in tryProcessSettlementCutoff
	if util.IsSettlementTimeDayBased(settlementConfig.Type) {
		// Notes: reason using PublishWithDelay instead of using PublishForSettlementProcess, to avoid D+ settlement process hold by earlier msg published into RMQ (FIFO)
		err := s.rabbitMqExt.PublishWithDelay(ctx, rabbitMqExt.SettlementProcessingRoutingKey, settlementModel.ProcessSettlementRequest{
			TransactionID:    transactionId,
			FeeTransactionID: feeTransactionId,
			MerchantID:       merchantId,
			Type:             constant.SettlementTransaction,
		}, settlementTime.Sub(processTime))
		if err != nil {
			s.logger.Error(ctx, "Failed while publish day based settlement process", logger.Error(err))
		}

		/// Notes: Record settlementTime for day based type settlement for:
		/// * Settlement hold release action
		s.logger.Info(ctx, "store estimate settlement at for day based settlement")
		ids := []string{transactionId, feeTransactionId}
		updateRequest := orchestratorModel.UpdateSettlementDetailRequest{
			EstimateSettlementAt: util.ValueToPtr(settlementTime.In(time.UTC)),
		}
		if err := s.accountTransactionRepo.UpdateSettlementDetailByIDs(ctx, ids, updateRequest); err != nil {
			s.logger.Warn(ctx, "Settlement detail update failed, but the process is considered successful as the message was published", logger.Error(err))
		}
	} else {
		err := s.rabbitMqExt.PublishForSettlementProcess(ctx, messagingQueueModel.PublishSettlementProcessPayload{
			SettlementType: settlementConfig.Type,
			MessageTTL:     settlementTime.Sub(processTime),
			Payload: &settlementModel.ProcessSettlementRequest{
				TransactionID:    transactionId,
				FeeTransactionID: feeTransactionId,
				MerchantID:       merchantId,
				Type:             constant.SettlementTransaction,
			},
			Day:           settlementConfig.GetSettlementDay(),
			ModifyMessage: nil,
		})
		if err != nil {
			s.logger.Error(ctx, "Failed while publish settlement process", logger.Error(err))
		}
	}

}

func (s *PaymentService) defineDefaultSettlementConfig(paymentMethodType, channel string) *merchantModel.SettlementConfig {

	var (
		ok     bool
		config config.SettlementConfigDetail
	)

	switch paymentMethodType {
	case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		config, ok = s.config.PaymentSettlementConfig.VirtualAccount[strings.ToLower(channel)]
		if !ok {
			config = s.config.PaymentSettlementConfig.VirtualAccount["other_channel"]
		}

	case paymentConstant.PAYMENT_METHOD_QRIS:
		config, ok = s.config.PaymentSettlementConfig.Qris[strings.ToLower(channel)]
		if !ok {
			config = s.config.PaymentSettlementConfig.Qris["other_channel"]
		}

	case paymentConstant.PAYMENT_METHOD_CREDIT_CARD:
		config, ok = s.config.PaymentSettlementConfig.CreditCard[strings.ToLower(channel)]
		if !ok {
			config = s.config.PaymentSettlementConfig.CreditCard["other_channel"]
		}

	case paymentConstant.PAYMENT_METHOD_VIRTUAL_TERMINAL:
		config = s.config.PaymentSettlementConfig.VirtualTerminal

	case paymentConstant.PAYMENT_METHOD_EWALLET:
		config, ok = s.config.PaymentSettlementConfig.Ewallet[strings.ToLower(channel)]
		if !ok {
			config = s.config.PaymentSettlementConfig.Ewallet["other_channel"]
		}

	case paymentConstant.PAYMENT_METHOD_CARD_FUNDED_PAYOUT:
		return &merchantModel.SettlementConfig{
			Type: s.config.PaymentSettlementConfig.CardFundedPayout.Type,
			CutOff: &merchantModel.SettlementConfigCutOff{
				Window: merchantModel.SettlementConfigCutOffWindow{
					StartTime: s.config.PaymentSettlementConfig.CardFundedPayout.CutOff.Window.StartTime,
					EndTime:   s.config.PaymentSettlementConfig.CardFundedPayout.CutOff.Window.EndTime,
				},
				Deferral: merchantModel.SettlementConfigCutOffDeferral{
					OffsetDays:    s.config.PaymentSettlementConfig.CardFundedPayout.CutOff.Deferral.OffsetDays,
					ExecutionTime: s.config.PaymentSettlementConfig.CardFundedPayout.CutOff.Deferral.ExecutionTime,
				},
			},
			SettlementTime: s.config.PaymentSettlementConfig.CardFundedPayout.SettlementTime,
		}

	default:
		return &merchantModel.SettlementConfig{
			Type: constant.SettlementTypeInstant,
		}
	}

	return &merchantModel.SettlementConfig{Type: config.Type}
}

// getSettlementConfigWithParentFallback looks up the settlement config for baseReq.
// When the sub-merchant has no config and the payment was created by a sub-merchant
// (onBehalf metadata present), it retries the lookup using the parent merchant id so
// that approved sub-merchants inherit their parent's settlement config, consistent
// with how fees are resolved. Returns nil only when neither the sub-merchant nor the
// parent (when applicable) has a config.
func (s *PaymentService) getSettlementConfigWithParentFallback(
	ctx context.Context,
	payment *paymentModel.Payment,
	baseReq merchantModel.GetSettlementConfigRequest,
) (*merchantModel.SettlementConfig, error) {
	cfg, err := s.merchantRepo.GetSettlementConfig(ctx, baseReq)
	if err != nil || cfg != nil {
		return cfg, err
	}

	parentID := parentMerchantIDFromOnBehalf(payment)
	if parentID == "" || parentID == baseReq.MerchantId {
		return nil, nil
	}

	s.logger.Warn(ctx, "sub-merchant has no settlement config, falling back to parent merchant", logger.String("parentMerchantID", parentID))
	baseReq.MerchantId = parentID
	parentCfg, err := s.merchantRepo.GetSettlementConfig(ctx, baseReq)
	if err != nil || parentCfg != nil {
		return parentCfg, err
	}

	s.logger.Warn(ctx, "parent merchant also has no settlement config, keeping default")
	return nil, nil
}

// parentMerchantIDFromOnBehalf extracts the parent merchant id stored on the payment's
// onBehalf metadata. Returns an empty string when the metadata is absent or malformed.
func parentMerchantIDFromOnBehalf(payment *paymentModel.Payment) string {
	if payment.Metadata == nil {
		return ""
	}
	onBehalf, err := util.ConvertToStruct[merchantModel.OnBehalfObject]((*payment.Metadata)["onBehalf"])
	if err != nil {
		return ""
	}
	return onBehalf.ParentMerchantId
}

func (s *PaymentService) BindingPaymentMethodDetail(channel string, payment *paymentModel.Payment) any {
	if payment.Metadata == nil {
		return nil
	}

	switch channel {
	case constant.ChannelQris:
		if (*payment.Metadata)["snapCore"] == nil {
			return nil
		}
		snapCoreResp := qrSnapCoreModel.GenerateQrMpmResponseData{}

		raw, _ := json.Marshal((*payment.Metadata)["snapCore"])
		_ = json.Unmarshal(raw, &snapCoreResp)

		result := orchestratorModel.MetadataPaymentMethodQRIS{
			StoreID:        snapCoreResp.StoreID,
			MerchantID:     snapCoreResp.MerchantID,
			MerchantName:   snapCoreResp.MerchantName,
			QrUrl:          snapCoreResp.QrUrl,
			QrContent:      snapCoreResp.QrContent,
			AdditionalInfo: snapCoreResp.AdditionalInfo,
		}
		if payment.ReferenceID != nil {
			result.PartnerReferenceNo = *payment.ReferenceID
		}
		result.QrType, _ = (*payment.Metadata)["qrType"].(string)
		result.QrMethodType, _ = (*payment.Metadata)["qrMethodType"].(string)

		return result

	case constant.ChannelVirtualAccount:
		if (*payment.Metadata)["snapCore"] == nil {
			return nil
		}
		snapCoreResp := vaSnapCoreModel.CreateVirtualAccountResponseData{}

		raw, _ := json.Marshal((*payment.Metadata)["snapCore"])
		_ = json.Unmarshal(raw, &snapCoreResp)

		result := orchestratorModel.MetadataPaymentMethodVA{
			AccountName:    snapCoreResp.AccountName,
			Acquirer:       snapCoreResp.Acquirer,
			Status:         snapCoreResp.Status,
			CreatedAt:      snapCoreResp.CreatedAt,
			IsClosedAmount: snapCoreResp.IsClosedAmount,
			IsSingleUse:    snapCoreResp.IsSingleUse,
			AdditionalInfo: snapCoreResp.AdditionalInfo,
		}
		return result
	}

	return nil // Options when payment channel is not registered
}
