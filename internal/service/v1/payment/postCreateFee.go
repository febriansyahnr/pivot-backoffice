package paymentService

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) PostCreateFeeTransaction(ctx context.Context, payment *paymentModel.Payment, request *paymentModel.PostCreateFeeTransactionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/PostCreateFeeTransaction")
	defer segment.End()

	if payment.Metadata == nil {
		return nil
	}

	var feeMetadataObject orchestratorModel.FeeTransactionMetadataObject
	paymentMetadata := *payment.Metadata

	merchantId := payment.MerchantID
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		merchantId = parentMerchantId

		transferId := ""
		trxFeeOnBehalf := &feeModel.TrxFeeOnBehalfMetadata{}

		raw, _ := json.Marshal(paymentMetadata["feeOnBehalf"])
		_ = json.Unmarshal(raw, trxFeeOnBehalf)

		if trxFeeOnBehalf.FinalAmount > 0 {
			subMerchantUUID, _ := uuid.Parse(payment.MerchantID)
			parentMerchantUUID, _ := uuid.Parse(parentMerchantId)

			transferRequest := &transfer.TransferRequest{
				SourceMerchantID: subMerchantUUID,             // Sub-Merchant
				RecipientID:      parentMerchantUUID.String(), // Main-Merchant
				ReferenceID:      payment.UUID,
				TransferType:     constant.MoneyFlowDirect,
				Amount:           trxFeeOnBehalf.FinalAmount,
				Remarks:          fmt.Sprintf("Payment Fee Transfer - Ref: %v", payment.ReferenceID),
				ParentMerchantID: parentMerchantUUID, // Main-Merchant
				Usecase:          constant.TypePayment,
			}
			transferResult, err := s.transferSvc.Transfer(ctx, transferRequest)
			if err != nil {
				return err
			}
			transferId = transferResult.UUID.String()
		}
		if _, ok := paymentMetadata["feeDetail"].(map[string]interface{}); ok {
			paymentMetadata["feeDetail"].(map[string]interface{})["notes"] = "ON-BEHALF"
			paymentMetadata["feeDetail"].(map[string]interface{})["transferId"] = transferId
		}
	}

	metadataByte, _ := json.Marshal(paymentMetadata["feeDetail"])
	json.Unmarshal(metadataByte, &feeMetadataObject)

	feeMetadataObject.LinkedTransactionId = request.LinkedTransactionID.String()
	if request.SettlementTransactionMetadata != nil {
		feeMetadataObject.AccountTransactionMetadataObject = request.SettlementTransactionMetadata
		metadataByte, _ = json.Marshal(feeMetadataObject)
	}
	var feeAmount float64
	if payment.CardFundedPayout != nil {
		feeAmount = payment.Fee.InexactFloat64()
	} else {
		feeAmount, _ = s.feeSvc.CalculateFee(ctx, &feeModel.GetFeeRequest{
			MerchantID:      merchantId,
			Reference:       feeMetadataObject.Type,
			PaymentMethod:   feeMetadataObject.Method,
			ReferenceAmount: request.TransactionAmount,
		}, &feeMetadataObject.FeeMetadataObject)
	}
	transactionTimestamp := payment.UpdatedAt
	if payment.TrxDatetime != nil {
		transactionTimestamp = *payment.TrxDatetime
	}

	merchantUUID, _ := uuid.Parse(merchantId)
	feeRequest := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 request.FeeTransactionID,
		ReferenceID:          payment.UUID,
		Type:                 orchestratorModel.TypeFee,
		MerchantID:           merchantUUID,
		Currency:             request.Currency,
		Credit:               0.00,
		Debit:                feeAmount,
		Channel:              request.Channel,
		Status:               request.Status,
		SettlementStatus:     request.SettlementStatus,
		SettlementAt:         request.SettlementAt,
		Remarks:              "",
		TransactionTimestamp: transactionTimestamp,
		Usecase:              constant.TypePayment,
		AdditionalInfo: types.NullJSONText{
			Valid:    true,
			JSONText: metadataByte,
		},
		SettlementModel: request.SettlementModel,
	}
	switch payment.Type {
	case constant.PaymentTypeVirtualTerminal:
		feeRequest.Usecase = constant.TypeVirtualTerminal
	case constant.PaymentTypeCardFundedPayout:
		feeRequest.Usecase = constant.TypePaymentFundedPayout
	case constant.UnifiedPaymentTypeSubPayment:
		feeRequest.Reference = constant.ReferenceSubPayment
	}
	if request.Status == constant.StatusSuccess && feeMetadataObject.DeductionType == constant.MerchantFeeDeductionTypeAutomated {
		feeRequest.Status = constant.StatusPending
	}
	if request.Status == constant.StatusSuccess && feeMetadataObject.DeductionType == constant.MerchantFeeDeductionTypeManual {
		feeRequest.SettlementStatus, feeRequest.SettlementAt = util.ValueToPtr(constant.StatusPending), nil
	}

	if errOrch := s.orchestratorSvc.PostAccountTransaction(ctx, feeRequest); errOrch != nil {
		s.logger.Error(ctx, "error when create account transaction for payment fee", logger.Error(errOrch))

		return errOrch
	}

	return nil
}
