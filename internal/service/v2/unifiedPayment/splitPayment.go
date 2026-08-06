package unifiedPaymentService

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/panjf2000/ants/v2"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *UnifiedPaymentService) InitiateSplitPayment(ctx context.Context, request *paymentModel.ProcessSplitPaymentRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/InitiateSplitPayment")
	defer span.End()

	parentPayment, err := s.paymentRepo.GetPaymentById(ctx, request.ParentPaymentID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get parent payment for split", logger.Error(err), logger.String("parentPaymentId", request.ParentPaymentID))
		return err
	}
	if parentPayment == nil {
		s.logger.Error(ctx, "Parent payment not found for split", logger.String("parentPaymentId", request.ParentPaymentID))
		return nil
	}

	subPayments, err := s.paymentRepo.GetAutoSplitSubPayments(ctx, &paymentModel.GetSubPaymentsRequest{
		ReferenceID: request.ParentPaymentID,
		MerchantID:  parentPayment.MerchantID,
	})
	if err != nil {
		s.logger.Error(ctx, "failed to get sub payments", logger.Error(err), logger.String("parentID", request.ParentPaymentID))
		return err
	}

	if len(subPayments) > 0 {
		s.logger.Error(ctx, "sub payments already proceed, duplicate processing avoided")
		return constant.ErrDuplicateData
	}

	unifiedPaymentMetadata := parentPayment.ToUnifiedPaymentMetadata()
	if unifiedPaymentMetadata == nil || unifiedPaymentMetadata.AutoSplitPayment == nil {
		s.logger.Info(ctx, "No split payment metadata found on parent", logger.String("parentPaymentId", request.ParentPaymentID))
		return nil
	}

	firstPayment := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{}
	autoSplitMetadata := unifiedPaymentMetadata.AutoSplitPayment
	processorLimit := decimal.NewFromFloat(autoSplitMetadata.ProcessorLimit)
	remaining := parentPayment.TotalAmount

	totalSubPayment := remaining.Div(processorLimit).Ceil().IntPart()
	s.logger.Info(ctx, "Processing split card payment",
		logger.Int("totalSubPayment", int(totalSubPayment)),
		logger.Float64("totalAmount", parentPayment.TotalAmount.InexactFloat64()),
		logger.Float64("processorLimit", processorLimit.InexactFloat64()),
		logger.String("parentPaymentId", request.ParentPaymentID),
	)

	expiryAfterMinutes := 30 // default expiry
	sequence := 1

	parentCardDetail := &unifiedPaymentModel.CardPaymentMethodDetail{}
	if unifiedPaymentMetadata.PaymentMethod.CardPaymentMethodDetail != nil {
		parentCardDetail = unifiedPaymentMetadata.PaymentMethod.CardPaymentMethodDetail
	}

	parentCardDetail.Token = request.FingerprintID

	createdBy := constant.SourceSystem
	paymentCreationCompleted := false

	defer func() {
		// skip hard delete when auth error
		if paymentCreationCompleted {
			return
		}

		err = s.paymentRepo.HardDeleteAutoSplitSubPayments(ctx, parentPayment.MerchantID, parentPayment.UUID)
		if err != nil {
			s.logger.Error(ctx, "Failed to hard delete remaining sub payments",
				logger.Error(err),
				logger.String("parentPaymentId", parentPayment.UUID),
			)
		}
	}()

	for !remaining.IsZero() {
		paymentAmount := processorLimit

		if remaining.LessThan(processorLimit) {
			paymentAmount = remaining
		}

		fee := unifiedPaymentMetadata.FeeDetail
		fee.Notes = fmt.Sprintf("Inherited from parent ID: %s", request.ParentPaymentID)
		fee.FinalAmount = decimal.NewFromFloat((fee.Percentage / 100) * paymentAmount.InexactFloat64()).Round(0).InexactFloat64()
		fee.TrxAmount = paymentAmount.InexactFloat64()

		paymentRequest := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
			PaymentID:         util.GenerateUUID().String(),
			ClientReferenceID: parentPayment.UUID,
			PaymentMethod: &unifiedPaymentModel.PaymentMethod{
				Type:                    constant.UnifiedPaymentMethodCard,
				CardPaymentMethodDetail: parentCardDetail,
			},
			PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
				Card: &unifiedPaymentModel.PaymentMethodOptionCard{
					ProcessingConfig: &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{},
				},
			},
			AutoConfirm: true,
			Mode:        constant.UnifiedPaymentModeAPI,
			ExpiryAt:    time.Now().Add(time.Duration(expiryAfterMinutes) * time.Minute).UTC(),
			Amount: unifiedPaymentModel.Amount{
				Currency: parentPayment.Currency,
				Value:    paymentAmount.InexactFloat64(),
			},
			MerchantID:  parentPayment.MerchantID,
			PaymentType: constant.UnifiedPaymentTypeSubPayment,
			CreatedBy:   createdBy,
			CreatedFrom: constant.SourceSystem,
			AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
				OrderReferenceID: request.ParentPaymentID,
			},
			FeeDetail: unifiedPaymentMetadata.FeeDetail,
		}

		if sequence == 1 {
			paymentRequest.PaymentMethodOptions.Card.ThreeDsMethod = constant.CardThreeDsMethodNever
			paymentRequest.PaymentMethodOptions.Card.ProcessingConfig.BankMerchantId = autoSplitMetadata.CITMerchantID
			paymentRequest.AutoSplitPayment.TransactionType = constant.AutoSplitPaymentTypeFirstPayment
			paymentRequest.AutoSplitPayment.Sequence = sequence
			firstPayment = paymentRequest
		} else {
			paymentRequest.PaymentMethodOptions.Card.ThreeDsMethod = constant.CardThreeDsMethodNever
			paymentRequest.PaymentMethodOptions.Card.ProcessingConfig.BankMerchantId = autoSplitMetadata.MITMerchantID
			paymentRequest.AutoSplitPayment.TransactionType = constant.AutoSplitPaymentTypeSubsequentPayment
			paymentRequest.AutoSplitPayment.Sequence = sequence
			paymentRequest.AutoSplitPayment.FirstPaymentID = firstPayment.PaymentID
		}

		_, err := s.CreateSession(ctx, paymentRequest)
		if err != nil {
			s.logger.Error(ctx, "Failed to create split child payment",
				logger.Error(err),
				logger.Int("sequence", sequence),
				logger.Float64("amount", paymentAmount.InexactFloat64()),
			)
			return err
		}

		s.logger.Info(ctx, "Created split child payment",
			logger.String("paymentId", paymentRequest.PaymentID),
			logger.Int("sequence", sequence),
			logger.Float64("amount", paymentAmount.InexactFloat64()),
		)

		remaining = remaining.Sub(paymentAmount)
		sequence++
	}

	paymentCreationCompleted = true

	reqCtx, payload, err := s.PrepareCardAuthentication(ctx, &unifiedPaymentModel.CardAuthenticationRequest{
		MerchantID:          firstPayment.MerchantID,
		PaymentID:           firstPayment.PaymentID,
		ClientTransactionID: firstPayment.ClientReferenceID,
		Amount:              firstPayment.Amount.Value,
		Currency:            firstPayment.Amount.Currency,
		ThreeDSCallbackID:   request.ThreeDSCallbackID,
		Card: &unifiedPaymentModel.CardAuthenticationRequestCard{
			Fingerprint: request.FingerprintID,
		},
		AutoSplitPayment: &unifiedPaymentModel.CardAuthAutoSplitPayment{
			TransactionType:  firstPayment.AutoSplitPayment.TransactionType,
			FirstPaymentID:   firstPayment.AutoSplitPayment.FirstPaymentID,
			Sequence:         firstPayment.AutoSplitPayment.Sequence,
			OrderReferenceID: firstPayment.AutoSplitPayment.OrderReferenceID,
		},
	})
	if err != nil {
		s.logger.Error(ctx, "failed to prepare card authentication", logger.Error(err), logger.String("paymentID", firstPayment.PaymentID))
		return err
	}

	_, err = s.creditcardSvc.Authentication(reqCtx, payload)
	if err != nil {
		s.logger.Error(ctx, "failed to authenticate card", logger.Error(err), logger.String("paymentID", firstPayment.PaymentID))
		errAbort := s.internalUnifiedPaymentSvc.AbortSplitPaymentOnCITFailure(ctx, request)
		if errAbort != nil {
			s.logger.Error(ctx, "failed to abort split payment", logger.Error(errAbort))
		}
		return err
	}

	return nil
}

func (s *UnifiedPaymentService) ContinueSplitPaymentExecution(ctx context.Context, request *paymentModel.ProcessSplitPaymentRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/ContinueSplitPaymentExecution")
	defer span.End()

	subPayments, err := s.paymentRepo.GetAutoSplitSubPayments(ctx, &paymentModel.GetSubPaymentsRequest{
		ReferenceID: request.ParentPaymentID,
		MerchantID:  request.MerchantID,
	})
	if err != nil {
		s.logger.Error(ctx, "failed to get sub payments", logger.Error(err), logger.String("parentID", request.ParentPaymentID))
		return err
	}

	if len(subPayments) == 0 {
		s.logger.Error(ctx, "sub payments did not exist", logger.String("parentPaymentID", request.ParentPaymentID))
		return constant.ErrDataNotFound
	}

	// handle complete process due to only one payment
	if len(subPayments) == 1 {
		s.logger.Info(ctx, "no sub payments with MIT type, skip exec remaining split payment")
		return nil
	}

	wg := new(sync.WaitGroup)
	pool, err := ants.NewPoolWithFuncGeneric(workerCount, func(payment *paymentModel.Payment) {
		defer wg.Done()

		if payment.Status != constant.UnifiedPaymentSessionStatusRequireAction {
			return
		}

		metadata := payment.ToUnifiedPaymentMetadata()

		if metadata == nil {
			return
		}

		cardFingerprint := ""
		if metadata.PaymentMethod != nil && metadata.PaymentMethod.CardPaymentMethodDetail != nil {
			cardFingerprint = metadata.PaymentMethod.CardPaymentMethodDetail.Token
		}

		reqCtx, payload, err := s.PrepareCardAuthentication(ctx, &unifiedPaymentModel.CardAuthenticationRequest{
			MerchantID:          payment.MerchantID,
			PaymentID:           payment.UUID,
			ClientTransactionID: *payment.ReferenceID,
			Amount:              payment.Amount.InexactFloat64(),
			Currency:            payment.Currency,
			Card: &unifiedPaymentModel.CardAuthenticationRequestCard{
				Fingerprint: cardFingerprint,
			},
			AutoSplitPayment: &unifiedPaymentModel.CardAuthAutoSplitPayment{
				TransactionType:  metadata.AutoSplitPayment.TransactionType,
				FirstPaymentID:   metadata.AutoSplitPayment.FirstPaymentID,
				Sequence:         metadata.AutoSplitPayment.Sequence,
				OrderReferenceID: metadata.AutoSplitPayment.OrderReferenceID,
			},
		})
		if err != nil {
			s.logger.Error(ctx, "failed to prepare card authentication", logger.Error(err), logger.String("paymentID", payment.UUID))
			return
		}

		_, err = s.creditcardSvc.Authentication(reqCtx, payload)
		if err != nil {
			s.logger.Error(ctx, "failed to authenticate card", logger.Error(err), logger.String("paymentID", payment.UUID))
			return
		}
		s.logger.Info(ctx, "sub payment executed", logger.String("paymentID", payment.UUID))
	})
	if err != nil {
		s.logger.Error(ctx, "Failed when preparing worker pool for pending subsequent payments", logger.Error(err))
		return err
	}
	defer pool.Release()

	for _, payment := range subPayments {
		wg.Add(1)
		pool.Invoke(payment)
	}

	wg.Wait()

	return nil
}

func (s *UnifiedPaymentService) EvaluateSplitPaymentOutcome(ctx context.Context, payment *paymentModel.Payment) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/EvaluateSplitPaymentOutcome")
	defer span.End()

	lockKey := fmt.Sprintf(constant.LockKeyAutoSplitPaymentCompleteKey, *payment.ReferenceID)
	mutex := s.cache.NewMutex(
		lockKey,
		redsync.WithExpiry(60*time.Second),
		redsync.WithRetryDelay(80*time.Millisecond),
		redsync.WithFailFast(true),
		redsync.WithTries(256),
	)
	if err := mutex.LockContext(ctx); err != nil {
		s.logger.Error(ctx, "failed to lock auto split mutex ", logger.Error(err), logger.String("paymentID", payment.UUID), logger.String("parentPaymentID", util.ValueOfPtr(payment.ReferenceID)))
		return err
	}
	defer func() {
		if _, unlockErr := mutex.UnlockContext(ctx); unlockErr != nil {
			s.logger.Error(ctx, "failed to release distributed lock for auto split payment", logger.Error(unlockErr))
		}
	}()

	summary, err := s.paymentRepo.GetSummaryAutoSplitPayment(ctx, &paymentModel.GetAutoSplitPaymentSummaryRequest{
		ReferenceID:     *payment.ReferenceID,
		MerchantID:      payment.MerchantID,
		MaxDateCreation: maxPaymentCreatedDays,
	})
	if err != nil {
		s.logger.Error(ctx, "failed to get auto split summary", logger.Error(err))
		return err
	}

	if summary == nil {
		return nil
	}

	// calculate payment final status
	isComplete := (summary.NumberOfFailedCharges + summary.NumberOfSuccessfulCharges) == summary.NumberOfCharges
	if summary.NumberOfCharges > 0 && isComplete {
		return s.internalUnifiedPaymentSvc.FinalizeSplitPayment(ctx, &paymentModel.ProcessSplitPaymentRequest{
			ParentPaymentID: *payment.ReferenceID,
			MerchantID:      payment.MerchantID,
			Summary:         summary,
		})
	}

	s.logger.Info(ctx, "auto split payment still in progress", logger.Any("summary", summary))
	return nil
}

func (s *UnifiedPaymentService) FinalizeSplitPayment(ctx context.Context, request *paymentModel.ProcessSplitPaymentRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/FinalizeSplitPayment")
	defer span.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.ParentPaymentID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get parent payment for split", logger.Error(err), logger.String("parentPaymentId", request.ParentPaymentID))
		return err
	}
	if payment == nil {
		s.logger.Error(ctx, "Parent payment not found for split", logger.String("parentPaymentId", request.ParentPaymentID))
		return constant.ErrDataNotFound
	}

	if payment.Status == constant.UnifiedPaymentSessionStatusPaid {
		s.logger.Warn(ctx, "auto split payment already paid, skip the process")
		return nil
	}

	autoSplitSummary, err := s.internalUnifiedPaymentSvc.GetAutoSplitPaymentDetail(ctx, &paymentModel.GetAutoSplitPaymentSummaryRequest{
		ReferenceID:              request.ParentPaymentID,
		MerchantID:               request.MerchantID,
		MaxDateCreation:          maxPaymentCreatedDays,
		ExcludeParentCalculation: true,
	})
	if err != nil {
		return err
	}

	unifiedPaymentMetadata := payment.ToUnifiedPaymentMetadata()
	totalFee := (unifiedPaymentMetadata.FeeDetail.Percentage / 100) * autoSplitSummary.TotalSuccessfulChargeAmount.ToDecimal().InexactFloat64()
	unifiedPaymentMetadata.AutoSplitPayment.Summary = autoSplitSummary
	unifiedPaymentMetadata.FeeDetail.TrxAmount = autoSplitSummary.TotalSuccessfulChargeAmount.ToDecimal().InexactFloat64()
	unifiedPaymentMetadata.FeeDetail.FinalAmount = decimal.NewFromFloat(totalFee).Round(0).InexactFloat64()
	metadataMap, _ := util.ConvertToStruct[*map[string]any](unifiedPaymentMetadata)

	paymentLedger, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.ReferencePayment)
	if err != nil {
		s.logger.Error(ctx, "failed to get ledger", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if paymentLedger == nil {
		s.logger.Error(ctx, "ledger not found", logger.String("parentPaymentId", request.ParentPaymentID))
		return constant.ErrDataNotFound
	}
	ctxTrx, err := s.paymentRepo.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	isCompleted := false
	defer func() {
		if isCompleted {
			return
		}

		err := s.paymentRepo.RollbackTransaction(ctxTrx)
		if err != nil {
			s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErrors.New(response.HttpErrDatabase, err)))
		}
	}()

	now := time.Now().UTC()
	payment.SetStatusByAutoSplitStatus(autoSplitSummary.Status)
	payment.UpdatedAt = now
	payment.Metadata = metadataMap
	payment.AutoSplitPayment = unifiedPaymentMetadata.AutoSplitPayment

	err = s.paymentRepo.UpdatePaymentData(ctxTrx, payment.ToDTO())
	if err != nil {
		s.logger.Error(ctx, "failed to update payment", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	updateRequest := orchestratorModel.UpdatePaymentTransactionRequest{
		ProcessorReferenceId:   payment.ProcessorID,
		ProcessorTransactionId: payment.ProcessorTransactionID,
		LedgerId:               paymentLedger.UUID.String(),
		UpdatedAt:              payment.UpdatedAt,
		TrxDatetime:            payment.TrxDatetime,
		Status:                 payment.GetLedgerStatus(),
		Channel:                constant.ChannelCard,
		Amount: commonModel.Amount{
			Currency: payment.Currency,
			Value:    "0",
		},
	}

	if paymentLedger.SettlementModel.Valid {
		updateRequest.SettlementModel = util.ValueToPtr(paymentLedger.SettlementModel.String)
	}
	if err := s.paymentSvc.UpdatePendingLedger(ctxTrx, payment, updateRequest); err != nil {
		return err
	}

	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return pkgErrors.New(response.HttpErrDatabase, errCommit)
	}

	isCompleted = true

	s.SendCallback(ctx, payment)

	// stomp notification
	identifier := util.ValueOfPtr(payment.ReferenceID)
	subject, message := constant.GetNotificationMessage(identifier, payment.Status)
	err = s.rabbitMqExt.PushNotification(ctx, &notification.PushNotification{
		RoutingKey: fmt.Sprintf(constant.NotificationRoutingKeyFmt, identifier),
		Payload: notification.PushNotificationPayload{
			ID:        util.GenerateUUID().String(),
			Subject:   subject,
			Type:      constant.CreateCardPaymentNotifType,
			Message:   message,
			CreatedAt: time.Now().UTC(),
			Status:    payment.Status,
		},
	})
	if err != nil {
		s.logger.Error(ctx, "failed to push notification for payment", logger.Error(err), logger.String("paymentID", request.ParentPaymentID))
	}

	return nil
}

func (s *UnifiedPaymentService) GetAutoSplitPaymentDetail(ctx context.Context, request *paymentModel.GetAutoSplitPaymentSummaryRequest) (*unifiedPaymentModel.AutoSplitPaymentSummary, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/GetAutoSplitPaymentDetail")
	defer span.End()

	parentPayment, err := s.paymentRepo.GetPaymentById(ctx, request.ReferenceID)
	if err != nil {
		s.logger.Error(ctx, "failed to get parent payment", logger.Error(err))
		return nil, err
	}

	if parentPayment == nil {
		return nil, constant.ErrDataNotFound
	}

	if parentPayment.IsFinalStatus() && parentPayment.AutoSplitPayment != nil && parentPayment.AutoSplitPayment.Summary != nil {
		return parentPayment.AutoSplitPayment.Summary, nil
	}

	summary, err := s.paymentRepo.GetSummaryAutoSplitPayment(ctx, &paymentModel.GetAutoSplitPaymentSummaryRequest{
		ReferenceID:     request.ReferenceID,
		MerchantID:      request.MerchantID,
		MaxDateCreation: maxPaymentCreatedDays,
	})
	if err != nil {
		s.logger.Error(ctx, "failed to get auto split summary", logger.Error(err))
		return nil, err
	}

	if summary == nil {
		return nil, nil
	}

	autoSplitSummary := &unifiedPaymentModel.AutoSplitPaymentSummary{
		Status:                    summary.GetFinalStatus(),
		NumberOfCharges:           summary.NumberOfCharges,
		NumberOfSuccessfulCharges: summary.NumberOfSuccessfulCharges,
		NumberOfFailedCharges:     summary.NumberOfFailedCharges,
		NumberOfInProcessCharges:  summary.NumberOfInProcessCharges,
		TotalFailedChargeAmount: commonModel.Amount{
			Currency: constant.CurrencyIDR,
			Value:    decimal.NewFromFloat(summary.TotalFailedChargeAmount).StringFixed(2),
		},
		TotalSuccessfulChargeAmount: commonModel.Amount{
			Currency: constant.CurrencyIDR,
			Value:    decimal.NewFromFloat(summary.TotalSuccessfulChargeAmount).StringFixed(2),
		},
		TotalInProgressChargeAmount: commonModel.Amount{
			Currency: constant.CurrencyIDR,
			Value:    decimal.NewFromFloat(summary.TotalInProgressChargeAmount).StringFixed(2),
		},
	}

	now := time.Now().UTC()
	charges, err := s.paymentRepo.GetCharges(ctx, &unifiedPaymentModel.FilterChargeRequest{
		MerchantID:        request.MerchantID,
		ClientReferenceID: request.ReferenceID,
		StartCreatedAt:    now.AddDate(0, 0, -request.MaxDateCreation),
		PaymentTypes:      []string{constant.UnifiedPaymentTypeSubPayment},
		EndCreatedAt:      now,
	})
	if err != nil {
		s.logger.Error(ctx, "failed to get charges", logger.Error(err), logger.String("referenceID", request.ReferenceID))
		return nil, err
	}

	autoSplitSummary.ChargeDetails = charges

	return autoSplitSummary, nil
}

func (s *UnifiedPaymentService) AbortSplitPaymentOnCITFailure(ctx context.Context, request *paymentModel.ProcessSplitPaymentRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/AbortSplitPaymentOnCITFailure")
	defer span.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.ParentPaymentID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get parent payment for split", logger.Error(err), logger.String("parentPaymentId", request.ParentPaymentID))
		return err
	}
	if payment == nil {
		s.logger.Error(ctx, "Parent payment not found for split", logger.String("parentPaymentId", request.ParentPaymentID))
		return nil
	}

	subPayments, err := s.paymentRepo.GetAutoSplitSubPayments(ctx, &paymentModel.GetSubPaymentsRequest{
		ReferenceID: request.ParentPaymentID,
		MerchantID:  request.MerchantID,
	})
	if err != nil {
		s.logger.Error(ctx, "failed to get sub payments", logger.Error(err), logger.String("parentID", request.ParentPaymentID))
		return err
	}

	totalSubPayment := len(subPayments)
	if totalSubPayment == 0 {
		s.logger.Error(ctx, "sub payments did not exist", logger.String("parentPaymentID", request.ParentPaymentID))
		return constant.ErrDataNotFound
	}

	wg := new(sync.WaitGroup)
	pool, err := ants.NewPoolWithFuncGeneric(workerCount, func(payment *paymentModel.Payment) {
		defer wg.Done()

		if payment.Status != constant.UnifiedPaymentSessionStatusRequireAction {
			return
		}

		paymentLedger, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.ReferencePayment)
		if err != nil {
			s.logger.Error(ctx, "failed to get ledger", logger.Error(err))
			return
		}

		ctxTrx, err := s.paymentRepo.BeginTransaction(ctx)
		if err != nil {
			return
		}

		isCompleted := false
		defer func() {
			if !isCompleted {
				err := s.paymentRepo.RollbackTransaction(ctxTrx)
				if err != nil {
					s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErrors.New(response.HttpErrDatabase, err)))
				}
			}
		}()

		payment.Status = constant.UnifiedPaymentSessionStatusCancelled
		payment.UpdatedAt = time.Now().UTC()

		err = s.paymentRepo.UpdatePaymentData(ctxTrx, payment.ToDTO())
		if err != nil {
			s.logger.Error(ctx, "failed to update payment", logger.Error(err))
			return
		}

		updateRequest := orchestratorModel.UpdatePaymentTransactionRequest{
			ProcessorReferenceId:   payment.ProcessorID,
			ProcessorTransactionId: payment.ProcessorTransactionID,
			LedgerId:               paymentLedger.UUID.String(),
			UpdatedAt:              payment.UpdatedAt,
			TrxDatetime:            payment.TrxDatetime,
			Status:                 payment.GetLedgerStatus(),
			Channel:                constant.ChannelCard,
			Amount: commonModel.Amount{
				Currency: payment.Currency,
				Value:    "0",
			},
		}

		if request.MethodDetail != nil {
			updateRequest.MethodDetail = request.MethodDetail
		}

		if err := s.paymentSvc.UpdatePendingLedger(ctxTrx, payment, updateRequest); err != nil {
			s.logger.Error(ctx, "failed to update ledger", logger.Error(err))
			return
		}

		if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
			s.logger.Error(ctx, "failed to commit transaction", logger.Error(err))
			return
		}

		isCompleted = true

		s.logger.Info(ctx, "sub payment updated", logger.String("paymentID", payment.UUID))
	})
	if err != nil {
		s.logger.Error(ctx, "Failed when preparing worker pool for invalidate subsequent payments", logger.Error(err))
		return err
	}
	defer pool.Release()

	request.Summary = &paymentModel.AutoSplitPaymentSummary{
		NumberOfCharges:       totalSubPayment,
		NumberOfFailedCharges: totalSubPayment,
	}

	for _, payment := range subPayments {
		wg.Add(1)
		pool.Invoke(payment)
		request.Summary.TotalFailedChargeAmount += payment.TotalAmount.InexactFloat64()
	}

	wg.Wait()

	err = s.internalUnifiedPaymentSvc.FinalizeSplitPayment(ctx, request)
	if err != nil {
		return err
	}

	return nil
}
