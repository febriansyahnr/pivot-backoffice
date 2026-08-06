package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/types"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *UnifiedPaymentService) InquiryCardPayment(ctx context.Context, payment *paymentModel.Payment) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/unifiedPayment/InquiryCardPayment")
	defer segment.End()

	var (
		isUnifiedPaymentV2 bool
		merchantID         = payment.MerchantID
	)

	if payment.Metadata != nil {
		if isUnifiedPaymentOk, ok := (*payment.Metadata)["isUnifiedPaymentV2"].(bool); ok {
			isUnifiedPaymentV2 = isUnifiedPaymentOk
		}
	}

	if !isUnifiedPaymentV2 {
		s.logger.Info(ctx, "ignore non unified payment cc of inquiry")
		return payment, pkgErrors.New(response.HttpErrForbidden, errors.New("invalid payment version"))
	}

	charge, err := s.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.ReferencePayment)
	if err != nil {
		s.logger.Error(ctx, "failed to get payment charge", logger.Error(err))
		return payment, nil
	}

	if charge == nil {
		s.logger.Error(ctx, "payment charge not found", logger.String("paymentID", payment.UUID))
		return payment, nil
	}

	if parentID := payment.GetOnBehalfParentID(); parentID != "" {
		merchantID = parentID
	}

	inquiry, err := s.creditcardSvc.InquiryTransaction(ctx, &creditcardModel.InquiryTransactionRequest{
		MerchantID:           merchantID,
		ClientReferenceID:    util.ValueOfPtr(payment.ReferenceID),
		ProcessorReferenceID: charge.ProcessorReferenceId,
	})

	if err != nil {
		s.logger.Error(ctx, "inquiry result return error and force payment to failed")
		payment.Status = constant.UnifiedPaymentSessionStatusCancelled
		return payment, err
	}

	if inquiry.PaymentStatus == constant.UnifiedPaymentSessionStatusProcessing {
		s.logger.Info(ctx, "card payment still processing", logger.String("paymentID", payment.UUID))
		payment.Status = constant.UnifiedPaymentSessionStatusProcessing
		return payment, nil
	}

	ctxTrx, errCtx := s.paymentRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return payment, errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.paymentRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErrors.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	if payment.PaymentMethod.Type == constant.ChannelCreditCard {
		channel := "FOREIGN_"
		merchant, err := s.merchantRepo.FindMerchantByID(ctx, payment.MerchantID)
		if err != nil {
			s.logger.Error(ctx, "Failed while find merchant by id", logger.Error(err))
			return payment, err
		}

		if merchant.BusinessCountry.String == inquiry.CardData.IssuingCountry {
			channel = "LOCAL_"
		}

		channel += strings.ToUpper(inquiry.CardData.CardBrand)
		payment.PaymentMethod.Acquirer = channel // For calculation payment fee per channel
	}

	if err := s.paymentSvc.DeterminePaymentFee(&ctxTrx, payment); err != nil {
		return payment, pkgErrors.New(response.HttpErrDatabase, err)
	}

	chargeMethodDetails := &unifiedPaymentModel.ChargePaymentMethodDetails{}
	_ = json.Unmarshal(charge.AdditionalInfo.JSONText, &struct {
		MethodDetail interface{} `json:"methodDetail"`
	}{
		MethodDetail: chargeMethodDetails,
	})

	if chargeMethodDetails.Card != nil {
		payment.ReconReferenceNo = chargeMethodDetails.Card.AuthorizationResult.AuthorizationID
		if util.ValueOfPtr(chargeMethodDetails.Card.SaveForFutureUse) {
			customerUseCase := payment.GetOneDollarAuthorizationUseCase()
			if err = s.storeFutureUseOfCustomerPaymentMethodCard(ctx, payment.CustomerID, customerUseCase, chargeMethodDetails.Card); err != nil {
				return payment, err
			}
		}
	}

	if inquiry.CardData != nil {
		first6 := inquiry.CardData.First8Digit
		if len(first6) > 6 {
			first6 = first6[:6]
		}

		chargeMethodDetails.Card.First6 = first6
		chargeMethodDetails.Card.First8 = inquiry.CardData.First8Digit
		chargeMethodDetails.Card.Last4 = inquiry.CardData.Last4Digit
		chargeMethodDetails.Card.ExpMonth = types.String(inquiry.CardData.ExpiryMonth)
		chargeMethodDetails.Card.ExpYear = types.String(inquiry.CardData.ExpiryYear)
		chargeMethodDetails.Card.Fingerprint = inquiry.CardData.Fingerprint
		chargeMethodDetails.Card.BankMerchantID = inquiry.BankMerchantID
		chargeMethodDetails.Card.SaveForFutureUse = util.ValueToPtr(inquiry.CardData.SavedFutureUse)
		chargeMethodDetails.Card.CardHolderName = inquiry.CardData.CardHolderName

		chargeMethodDetails.Card.BinInformations = unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
			Type:        inquiry.CardData.CardType,
			IssuingBank: inquiry.CardData.CardIssuing,
			Brand:       inquiry.CardData.CardBrand,
			Country:     inquiry.CardData.IssuingCountry,
		}
	}

	if inquiry.AuthenticationData != nil {
		chargeMethodDetails.Card.AuthenticationResult = &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthenticationResult{
			ThreeDsVersion: inquiry.AuthenticationData.ThreeDsVer,
			ThreeDsResult:  inquiry.AuthenticationData.AuthenticationResult,
			ThreeDsMethod:  inquiry.AuthenticationMethod,
			EciCode:        inquiry.AuthenticationData.EciCode,
		}
	}

	if inquiry.AuthorizationData != nil {
		chargeMethodDetails.Card.AuthorizationResult = &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
			AcquirerReferenceNumber:  inquiry.AuthorizationData.AcquirerTransactionID,
			RetrievalReferenceNumber: inquiry.AuthorizationData.TransactionReference,
			Stan:                     inquiry.AuthorizationData.Stan,
			AvsResult:                inquiry.AuthorizationData.AvsResult,
			CvvResult:                inquiry.AuthorizationData.CvvResult,
			AuthorizedAmount: unifiedPaymentModel.Amount{
				Currency: inquiry.Currency,
				Value:    inquiry.Amount.InexactFloat64(),
			},
			IssuerAuthorizationCode: inquiry.AuthorizationData.AcquirerResponseCode,
			AuthorizationID:         inquiry.AuthorizationData.AuthorizationID,
		}
	}
	if inquiry.ResponseCode != nil {
		chargeMethodDetails.Card.ResponseCode = &unifiedPaymentModel.ChargePaymentMethodDetailCardResponseCode{
			GatewayCode:           inquiry.ResponseCode.GatewayCode,
			GatewayRecommendation: inquiry.ResponseCode.GatewayRecommendation,
		}
	}

	switch inquiry.PaymentStatus {
	case constant.ChargeStatusSuccess:
		payment, err = s.processCardPaidInquiryPayment(ctxTrx, payment, charge, chargeMethodDetails, inquiry)
	case constant.ChargeStatusFailed:
		payment, err = s.processCardFailedInquiryPayment(ctxTrx, payment, charge, chargeMethodDetails, inquiry)
	}

	if err != nil {
		s.logger.Error(ctx, "failed to process inquiry result", logger.String("paymentID", payment.UUID))
		return nil, err
	}

	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return payment, pkgErrors.New(response.HttpErrDatabase, errCommit)
	}

	isCompleted = true
	s.SendCallback(ctx, payment)

	return payment, nil
}

func (s *UnifiedPaymentService) processCardPaidInquiryPayment(ctx context.Context, payment *paymentModel.Payment, charge *orchestrator_model.AccountTransactionWithUseCase, methodDetails *unifiedPaymentModel.ChargePaymentMethodDetails, inquiry *creditcardModel.PaymentNotificationDataRequest) (*paymentModel.Payment, error) {

	payment.Status = constant.UnifiedPaymentSessionStatusPaid
	payment.UpdatedAt = time.Now().UTC()
	err := s.paymentRepo.UpdatePaymentStatus(ctx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt)
	if err != nil {
		return payment, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if !inquiry.Updated.IsZero() {
		payment.TrxDatetime = &inquiry.Updated
	} else {
		inquiry.Updated = time.Now().UTC()
	}
	payment.ProcessorTransactionID = inquiry.TransactionID.String()
	payment.Processor = constant.CreditCardCoreProcessor
	payment.ProcessorID = inquiry.AcquirerTransactionID
	payment.TrxDatetime = &inquiry.Updated

	if payment.ProcessorReferenceNumber != nil {
		payment.ReconReferenceNo = *payment.ProcessorReferenceNumber
	}

	paidAmount := commonModel.Amount{
		Currency: inquiry.Currency,
		Value:    decimal.NewFromFloat(inquiry.Amount.InexactFloat64()).StringFixed(2),
	}

	updateRequest := orchestrator_model.UpdatePaymentTransactionRequest{
		ProcessorReferenceId:   inquiry.AcquirerTransactionID,
		ProcessorTransactionId: inquiry.TransactionID.String(),
		LedgerId:               charge.UUID.String(),
		UpdatedAt:              payment.UpdatedAt,
		TrxDatetime:            payment.TrxDatetime,
		Status:                 constant.StatusSuccess,
		Channel:                constant.UnifiedPaymentMethodCard,
		Amount:                 paidAmount,
		MethodDetail:           methodDetails,
	}

	if charge.SettlementModel.Valid {
		updateRequest.SettlementModel = util.ValueToPtr(charge.SettlementModel.String)
	}
	if err := s.paymentSvc.UpdatePendingLedger(ctx, payment, updateRequest); err != nil {
		return payment, err
	}

	// Send callback on paid charge
	return payment, nil
}

func (s *UnifiedPaymentService) processCardFailedInquiryPayment(ctx context.Context, payment *paymentModel.Payment, charge *orchestrator_model.AccountTransactionWithUseCase, methodDetails *unifiedPaymentModel.ChargePaymentMethodDetails, inquiry *creditcardModel.PaymentNotificationDataRequest) (*paymentModel.Payment, error) {

	payment.Status = constant.UnifiedPaymentSessionStatusCancelled
	payment.UpdatedAt = time.Now().UTC()
	err := s.paymentRepo.UpdatePaymentStatus(ctx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt)
	if err != nil {
		s.logger.Error(ctx, "Failed to update payment status", logger.Error(err))
		return payment, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if charge == nil {
		s.logger.Info(ctx, "charge data was missing", logger.String("paymentID", payment.UUID))
		return payment, pkgErrors.New(response.HttpErrNotFound, constant.ErrPaymentChargeNotFound)
	}

	paidAmount := commonModel.Amount{
		Currency: inquiry.Currency,
		Value:    decimal.NewFromFloat(inquiry.Amount.InexactFloat64()).StringFixed(2),
	}

	updateRequest := orchestrator_model.UpdatePaymentTransactionRequest{
		ProcessorReferenceId:   inquiry.AcquirerTransactionID,
		ProcessorTransactionId: inquiry.TransactionID.String(),
		LedgerId:               charge.UUID.String(),
		UpdatedAt:              payment.UpdatedAt,
		TrxDatetime:            payment.TrxDatetime,
		Status:                 constant.StatusFailed,
		Channel:                constant.UnifiedPaymentMethodCard,
		Amount:                 paidAmount,
		MethodDetail:           methodDetails,
	}

	if charge.SettlementModel.Valid {
		updateRequest.SettlementModel = util.ValueToPtr(charge.SettlementModel.String)
	}

	if err := s.paymentSvc.UpdatePendingLedger(ctx, payment, updateRequest); err != nil {
		s.logger.Error(ctx, "Failed to update pending ledger", logger.Error(err))
		return payment, err
	}

	return payment, nil
}
