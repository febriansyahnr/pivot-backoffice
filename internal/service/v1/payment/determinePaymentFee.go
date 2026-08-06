package paymentService

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) DeterminePaymentFee(ctx *context.Context, payment *paymentModel.Payment) (err error) {

	// Fee exemptions: ONE_DOLLAR authorization and zero-amount recurring transactions carry no fees.
	// Auto-split sub-payments inherit fee details from their parent payment during creation.
	if payment.IsFeeExempt() || payment.IsAutoSplitSubPayments() {
		return nil
	}

	// For card-funded payout transactions, the fee calculation is performed when the payout is initially created by the maker.
	if payment.CardFundedPayout != nil {
		(*payment.Metadata)["feeDetail"] = &payment.CardFundedPayout.FeeConfig
		return nil
	}

	paymentMethodChannelType := constant.PaymentMethodChannelTypeAggregator
	paymentLedger, _ := s.orchestratorSvc.FindByReference(*ctx, payment.UUID, constant.TypePayment)
	if paymentLedger != nil && paymentLedger.SettlementModel.Valid {
		paymentMethodChannelType = paymentLedger.SettlementModel.String
	}

	feeMerchantId := payment.MerchantID
	paymentAmount := payment.Amount.InexactFloat64()
	if paymentLedger != nil {
		paymentAmount = paymentLedger.Credit
	}

	if payment.Metadata != nil {
		if onBehalf, ok := (*payment.Metadata)["onBehalf"].(map[string]any); ok {
			parentMerchantId, _ := onBehalf["parentMerchantId"].(string)

			feeMerchantId = parentMerchantId
			*ctx = context.WithValue(*ctx, constant.CtxParentMerchantId, parentMerchantId)

			// Disable fee on behalf for payments and move using split routing (Innovation Sprint 34).
			// (*payment.Metadata)["feeOnBehalf"], err = s.feeSvc.GetTransactionFeeOnBehalf(
			// 	*ctx, &feeModel.GetTrxFeeOnBehalfRequest{
			// 		MerchantId:        parentMerchantId,
			// 		SubMerchantId:     payment.MerchantID,
			// 		Reference:         constant.ReferencePayment,
			// 		PaymentMethod:     payment.PaymentMethod.Type,
			// 		TransactionAmount: paymentAmount,
			// 	},
			// )
			// if err != nil {
			// 	s.logger.Error(*ctx, "Failed while get transaction fee on behalf", logger.Error(err))
			// 	return err
			// }
			//////////////////////////////////////////////////////////////////////////////////////////////
		}
	}

	if payment.Metadata == nil {
		payment.Metadata = &map[string]any{}
	}

	getFeeRequest := &feeModel.GetFeeRequest{
		MerchantID:      feeMerchantId,
		Reference:       constant.ReferencePayment,
		PaymentMethod:   payment.PaymentMethod.Type,
		Channel:         strings.ToUpper(payment.PaymentMethod.Acquirer),
		ReferenceAmount: paymentAmount,
		SettlementModel: paymentMethodChannelType,
	}
	if payment.Type == constant.TypeVirtualTerminal {
		getFeeRequest.PaymentMethod = constant.TypeVirtualTerminal
	}

	if payment.AutoSplitPayment != nil {
		getFeeRequest.PaymentMethod = constant.TypeSplitPayment
	}

	_, feeDetail, err := s.feeSvc.GetFeeCalculationAndDetail(*ctx, getFeeRequest)
	if err != nil {
		s.logger.Error(*ctx, "Failed while get fee calculation and detail", logger.Error(err))
		return err
	}

	// Increment LADDER tiering counter after successful fee resolution.
	s.feeSvc.IncrementLadderCounter(*ctx, feeDetail.LadderCounterKey, feeDetail.LadderCounterIncrement)

	if constant.IsDirectPSP(paymentMethodChannelType) {
		feeDetail.DeductionType = constant.MerchantFeeDeductionTypeManual
	}
	(*payment.Metadata)["feeDetail"] = feeDetail

	paymentMetadataReq := paymentModel.UpdatePaymentMetadataRequest{
		FeeDetail:   (*payment.Metadata)["feeDetail"],
		FeeOnBehalf: (*payment.Metadata)["feeOnBehalf"],
	}
	return s.paymentRepo.UpdatePaymentMetadataById(*ctx, payment.UUID, paymentMetadataReq)
}
