package feeService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
)

func (s *FeeService) GetPayoutTransactionFee(ctx context.Context, request feeModel.GetPayoutTrxFeeRequest) (feeModel.FeeResponseder, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/fee/GetPayoutTransactionFee")
	defer segment.End()

	if request.MerchantType == constant.MerchantTypeSubMerchant {
		return s.GetTransactionFeeOnBehalf(
			ctx, &feeModel.GetTrxFeeOnBehalfRequest{
				MerchantId:    request.ParentMerchantId,
				SubMerchantId: request.MerchantId,
				Reference:     constant.ReferenceDisbursement,
			},
		)
	}

	feeRequest := &feeModel.GetFeeRequest{
		MerchantID: request.MerchantId,
		Reference:  constant.ReferenceDisbursement,
		Channel:    request.ChannelCode,
	}
	if feeRequest.Channel == "" && request.BankCode != "" {
		if bank := bankTransfer.NewBankDB().FindByCode(request.BankCode); bank != nil {
			feeRequest.Channel = bank.ChannelCode
		}
	}
	_, result, err := s.GetFeeCalculationAndDetail(ctx, feeRequest)
	return result, err
}
