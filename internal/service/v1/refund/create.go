package refundService

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *RefundService) Create(ctx context.Context, request *refundModel.CreateRefundRequest) (*refundModel.RefundResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refund/Create")
	defer span.End()

	// Generate refund ID for status history tracking
	var (
		refundID, _  = uuid.NewV7()
		refundAmount float64
	)

	// Record refund creation
	s.RecordRefundStatusHistory(ctx, refundID.String(), constant.StatusHistoryActorUser, constant.RefundStatusHistoryPending)

	// Validate clientReferenceId
	if isExist, err := s.refundRepo.ExistsByClientReferenceAndMerchantID(ctx, request.ClientReferenceID, request.MerchantID); err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	} else if isExist {
		s.logger.Warn(ctx, "[CreateRefund] Client reference id is exist")
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrClientReferenceIDAlreadyExist)
	}

	// Find payment by paymentID
	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentSessionID)
	if err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	} else if payment == nil {
		s.logger.Warn(ctx, "[CreateRefund] Payment not found")
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	if payment.MerchantID != request.MerchantID {
		s.logger.Warn(ctx, "[CreateRefund] MerchantID is not match with payment session id", logger.String("paymentMerchantID", payment.MerchantID), logger.String("requestMerchantID", request.MerchantID))
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrMerchantIsNotMatch)
	}

	// Find charge by paymentID and chargeID
	var paymentCharge *orchestratorModel.AccountTransactionWithUseCase
	if request.ChargeID != "" {
		paymentCharge, err = s.accountTransactionRepo.FindByID(ctx, request.ChargeID)
	} else {
		paymentCharge, err = s.accountTransactionRepo.FindByReference(ctx, request.PaymentSessionID, constant.TypePayment)
	}

	if err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	} else if paymentCharge == nil {
		s.logger.Warn(ctx, "[CreateRefund] Payment charge not found")
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentChargeNotFound)
	} else if paymentCharge.ReferenceID != request.PaymentSessionID {
		s.logger.Warn(ctx, "[CreateRefund] Payment charge is not match with payment session id")
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentChargeNotFound)
	} else if paymentCharge.Status != constant.StatusSuccess {
		s.logger.Info(ctx, "[CreateRefund] Payment charge status is not successful yet")
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentChargeNotSettled)
	}

	if !request.IsFullAmount {
		refundAmount, err = strconv.ParseFloat(request.Amount.Value, 64)
		if err != nil {
			return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, err)
		}
	}

	// Validate payment refund status by charge ID
	var (
		totalRefundedAmount float64
		refundCurrency      = paymentCharge.Currency
	)
	totalRefundedAmount, err = s.refundRepo.GetTotalRefundedAmount(ctx, payment.UUID)
	if err != nil {
		s.logger.Error(ctx, "[CreateRefund] error when getting total refunded amount", logger.Error(err))
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	if request.IsFullAmount {
		refundAmount = paymentCharge.Credit - totalRefundedAmount
		if refundAmount == 0 {
			s.logger.Info(ctx, "total refund amount is 0", logger.Float64("paymentAmount", paymentCharge.Credit), logger.Float64("totalRefundedAmount", totalRefundedAmount))
			return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrPaymentAlreadyRefunded)
		}
	}
	if totalRefundedAmount >= paymentCharge.Credit || (totalRefundedAmount+refundAmount) > paymentCharge.Credit {
		s.logger.Info(ctx, "[CreateRefund] Total refund amount exceed payment charge credit")
		return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrRefundAmountExceedPaymentCharge)
	}

	// Get the payment MDR fee (percentage of the merchant_fees) to the merchant
	var feeCharge *orchestratorModel.AccountTransactionWithUseCase
	if !constant.IsDirectPSP(paymentCharge.SettlementModel.String) {
		var feeAmount float64
		feeCharge, err = s.accountTransactionRepo.FindByReference(ctx, request.PaymentSessionID, constant.TypeFee)
		if err != nil {
			s.logger.Error(ctx, "[CreateRefund] error when find fee charge", logger.Error(err))
			return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
		}
		if feeCharge == nil {
			// notes: no error returned under the assumption that fee might be abstained in some special cases
			s.logger.Warn(ctx, "[CreateRefund] Fee charge not found, continuing")
		} else {
			feeAmount, _, _ = s.calculateMDRFee(ctx, feeCharge, refundAmount)
			s.logger.Info(ctx, "[CreateRefund] Fee amount", logger.Float64("feeAmount", feeAmount))
		}

		if paymentCharge.SettlementStatus.String == constant.SettlementStatusPending {
			s.logger.Info(ctx, "[CreateRefund] Payment charge is unsettled, skip check balance")
		} else {
			mutex := s.redis.NewMutex(
				"backend-portal:merchant-balances:"+request.MerchantID+":deduct:"+constant.TypePayment,
				redsync.WithExpiry(30*time.Second),
				redsync.WithRetryDelay(50*time.Millisecond),
				redsync.WithFailFast(true),
				redsync.WithTries(256),
			)
			if errLock := mutex.LockContext(ctx); errLock != nil {
				s.logger.Error(ctx, "[CreateRefund] Error acquire distributed lock for balance deduction", logger.Error(errLock))
				return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrReduceBalance)
			}
			s.logger.Info(ctx, "[CreateRefund] Success acquire lock for balance deduction")
			defer func() {
				ok, errUnlock := mutex.UnlockContext(ctx)
				if errUnlock != nil {
					s.logger.Warn(ctx, "[CreateRefund] Error release unlock distributed lock", logger.Error(errUnlock))
					return
				}
				if !ok {
					s.logger.Warn(ctx, "[CreateRefund] Failed to release distributed lock")
					return
				}
				s.logger.Info(ctx, "[CreateRefund] Success release lock for balance deduction")
			}()

			availableBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, request.MerchantID, constant.TypePayment)
			if err != nil {
				s.logger.Error(ctx, "[CreateRefund] Get available merchant balance", logger.Error(err))
				return nil, pkgErr.New(httpResponse.HttpErrDatabase, constant.ErrValidateBalance)
			}
			if availableBalance < refundAmount-feeAmount {
				s.logger.Warn(ctx, "[CreateRefund] Insufficient balance", logger.Float64("availableBalance", availableBalance), logger.Float64("refundAmount", refundAmount), logger.Float64("feeAmount", feeAmount))
				return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrInsufficientBalance)
			}
		}
	}

	// Validate payment method for CRM request & Open API
	// Validate whether payment method channel type is Facilitator / Direct PSP type
	var settlementModel string
	if paymentCharge.SettlementModel.Valid {
		settlementModel = paymentCharge.SettlementModel.String
	}
	if constant.IsDirectPSP(settlementModel) {
		if request.Method == constant.RefundMethodTransferOnly {
			return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrRefundIncorrectRequestMethodForFacilitator)
		}
		var channelSupportedPaymentMethod = []string{paymentConstant.PAYMENT_METHOD_CREDIT_CARD, paymentConstant.PAYMENT_METHOD_QRIS, paymentConstant.PAYMENT_METHOD_EWALLET}
		if !slices.Contains(channelSupportedPaymentMethod, payment.PaymentMethod.Type) {
			s.logger.Warn(ctx, "[CreateRefund] Not allowed to refund payment with payment method facilitator configuration", logger.String("paymentMethodType", payment.PaymentMethod.Type))
			return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrRefundNotAllowedForPaymentMethodFacilitatorConfig)
		}
		if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_QRIS && !strings.EqualFold(payment.PaymentMethod.Acquirer, paymentConstant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC) {
			s.logger.Warn(ctx, "[CreateRefund] Not allowed to refund payment with QRIS from non BNC acquirer", logger.String("paymentMethodAcquirer", payment.PaymentMethod.Acquirer))
			return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, constant.ErrRefundNotAllowedForPaymentMethodFacilitatorConfig)
		}
	}

	// Block refund for recent credit card payments within 24 hours
	if payment.PaymentMethod.Type == paymentConstant.PAYMENT_METHOD_CREDIT_CARD &&
		util.IsWithinLast24Hours(paymentCharge.TransactionTimestamp) {

		isFull := request.IsFullAmount
		sameAsPayment := refundAmount == payment.Amount.InexactFloat64()
		isPartial := !request.IsFullAmount || refundAmount < paymentCharge.Credit

		var (
			msg string
			err error
		)

		switch {
		case isFull && !sameAsPayment:
			msg = "[CreateRefund] Refund is not yet available"
			err = constant.ErrRefundIsNotYetAvailable

		case isPartial:
			msg = "[CreateRefund] Partial amount is not yet available"
			err = constant.ErrRefundPartialIsNotYetAvailable
		}

		if msg != "" {
			s.logger.Warn(ctx, msg)
			return nil, pkgErr.New(httpResponse.HttpErrUnprocessableContent, err)
		}
	}

	destinationType := constant.RefundDestinationTypeChannel
	if request.Method == constant.RefundMethodTransferOnly {
		destinationType = constant.RefundDestinationTypeAccount
	}

	// Begin trx
	ctxTrx, errCtx := s.paymentRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, errCtx)
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.paymentRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErr.New(httpResponse.HttpErrDatabase, e)))
			}
		}
	}()

	// Insert to refund table
	// refundID already declared above for status history
	metadataObj := &refundModel.MetadataObj{
		TransferDestination: request.TransferDestination,
		ClientMetadata:      request.Metadata,
	}
	metadataJson, _ := json.Marshal(metadataObj)

	// Build refund model
	refund := &refundModel.Refund{
		UUID:              refundID.String(),
		MerchantID:        request.MerchantID,
		ClientReferenceID: request.ClientReferenceID,
		PaymentID:         request.PaymentSessionID,
		PaymentChargeID:   paymentCharge.UUID.String(),
		Currency:          refundCurrency,
		Amount:            refundAmount,
		Status:            constant.RefundStatusPending,
		Reason:            request.Reason,
		Description:       request.Description,
		DestinationType:   destinationType,
		Method:            request.Method,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
		Metadata: types.NullJSONText{
			Valid:    true,
			JSONText: metadataJson,
		},
	}
	if err = s.refundRepo.Insert(ctxTrx, refund); err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}

	// Insert refund payment ledger
	refundLedgerID, _ := uuid.NewV7()
	refundLedger := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 refundLedgerID,
		ReferenceID:          refund.UUID,
		Type:                 constant.TypeRefund,
		MerchantID:           util.ParseUUID(request.MerchantID),
		Currency:             refundCurrency,
		Debit:                refundAmount,
		Status:               constant.StatusPending,
		TransactionTimestamp: refund.CreatedAt,
		Usecase:              constant.TypePayment,
		SettlementModel:      &settlementModel,
		AdditionalInfo: types.NullJSONText{
			Valid: true,
		},
	}
	refundLedger.AdditionalInfo.JSONText, _ = json.Marshal(orchestratorModel.MetadataRefund{
		PaymentSessionID: request.PaymentSessionID,
		PaymentChargeID:  paymentCharge.UUID.String(),
	})
	if err = s.orchestratorSvc.PostAccountTransaction(ctx, refundLedger); err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}

	// Insert refund fee ledger
	feeChargeID := ""
	if feeCharge != nil {
		feeChargeID = feeCharge.UUID.String()
	}
	s.calculateAndInsertMDRFee(ctx, feeCharge, refundAmount, request.PaymentSessionID, feeChargeID, refund)

	if err = s.refundRepo.CommitTransaction(ctxTrx); err != nil {
		return nil, pkgErr.New(httpResponse.HttpErrDatabase, err)
	}
	isCompleted = true

	// Record refund waiting for bank transfer
	s.RecordRefundStatusHistory(ctx, refundID.String(), constant.StatusHistoryActorSystem, constant.RefundStatusHistoryWaitingBankTransfer)

	// Publish to process refund
	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.RefundProcessRoutingKey, nil, &refundModel.RefundProcessRequest{
		RefundID:                 refund.UUID,
		PaymentMethodChannelType: settlementModel,
	})

	refundResponse := refundModel.RefundResponse{
		ID:                refundID.String(),
		MerchantID:        request.MerchantID,
		ClientReferenceID: request.ClientReferenceID,
		PaymentSessionID:  request.PaymentSessionID,
		ChargeID:          paymentCharge.UUID.String(),
		CapturedAmount: commonModel.Amount{
			Currency: paymentCharge.Currency,
			Value:    fmt.Sprintf("%.2f", paymentCharge.Credit),
		},
		IsFullAmount: request.IsFullAmount,
		Amount: commonModel.Amount{
			Currency: refund.Currency,
			Value:    fmt.Sprintf("%.2f", refundAmount),
		},
		Status:              refund.Status,
		Reason:              refund.Reason,
		Description:         refund.Description,
		Method:              refund.Method,
		DestinationType:     "", // Value is available after refund reached to final status, indicating the refund destination OR if status=WAITING_BANK_TRANSFER then value will always be ACCOUNT
		TransferDestination: request.TransferDestination,
		CreatedAt:           refund.CreatedAt,
		UpdatedAt:           refund.UpdatedAt,
		Metadata:            request.Metadata,
	}

	errCallback := s.SendCallback(ctx, refundID.String(), request.MerchantID)
	if errCallback != nil {
		s.logger.Error(ctx, "error when send callback", logger.Error(errCallback))
	}

	return &refundResponse, nil
}

func (s *RefundService) calculateAndInsertMDRFee(ctx context.Context, paymentFeeLedger *orchestratorModel.AccountTransactionWithUseCase, refundAmount float64, paymentSessionID, feeChargeID string, refundObj *refundModel.Refund) error {
	if paymentFeeLedger == nil || refundObj == nil {
		return nil
	}

	refundOfPaymentFee, paymentFeePercentage, _ := s.calculateMDRFee(ctx, paymentFeeLedger, refundAmount)
	if refundOfPaymentFee > 0.0 {
		// Create refund fee (return the payment fee to the client)
		refundFeeLedgerID, _ := uuid.NewV7()
		refundFeeLedger := &orchestratorModel.CreateAccountTransactionRequest{
			UUID:                 refundFeeLedgerID,
			ReferenceID:          refundObj.UUID,
			Type:                 constant.TypeFeeRefund,
			MerchantID:           paymentFeeLedger.MerchantID,
			Currency:             paymentFeeLedger.Currency,
			Credit:               decimal.NewFromFloat(refundOfPaymentFee).Round(0).InexactFloat64(),
			Status:               constant.StatusPending,
			TransactionTimestamp: refundObj.CreatedAt,
			Usecase:              constant.TypePayment,
			AdditionalInfo: types.NullJSONText{
				Valid: true,
			},
		}
		refundFeeLedger.AdditionalInfo.JSONText, _ = json.Marshal(orchestratorModel.MetadataRefundOfPaymentFee{
			PaymentSessionID:   paymentSessionID,
			PaymentChargeID:    feeChargeID,
			PaymentFeeLedgerID: paymentFeeLedger.UUID.String(),
			FeeDetail: &feeModel.FeeMetadataObject{
				Type:        constant.TypeRefund,
				AmountType:  constant.MerchantFeePercentageType,
				Percentage:  paymentFeePercentage,
				FinalAmount: refundOfPaymentFee,
			},
		})

		if err := s.orchestratorSvc.PostAccountTransaction(ctx, refundFeeLedger); err != nil {
			return pkgErr.New(httpResponse.HttpErrDatabase, constant.ErrRefundPaymentProcess)
		}
	}
	return nil
}

func (s *RefundService) calculateMDRFee(ctx context.Context, paymentFeeLedger *orchestratorModel.AccountTransactionWithUseCase, refundAmount float64) (float64, float64, error) {
	if paymentFeeLedger == nil {
		return 0, 0, nil
	}

	// Check fee object
	paymentFeeLedgerMetadata := orchestratorModel.FeeTransactionMetadataObject{}
	_ = json.Unmarshal(paymentFeeLedger.AdditionalInfo.JSONText, &paymentFeeLedgerMetadata)

	// Product team, call it MDR
	paymentFeePercentage := 0.0
	if paymentFeeLedgerMetadata.AmountType != constant.MerchantFeeAmountType {
		paymentFeePercentage = paymentFeeLedgerMetadata.Percentage
	}

	// Value of payment fee that refunded to the merchant based on MDR percentage
	refundOfPaymentFee := 0.0
	if paymentFeePercentage > 0.0 {
		refundOfPaymentFee = (paymentFeePercentage / 100) * refundAmount
	}

	return refundOfPaymentFee, paymentFeePercentage, nil
}
