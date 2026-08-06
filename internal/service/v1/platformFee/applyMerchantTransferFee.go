package platformFeeService

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platformFee"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PlatformFeeService) ApplyMerchantTransferFee(ctx context.Context, req platformFee.PlatformFeeRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/platformFee/ApplyMerchantTransferFee")
	defer segment.End()

	return s.ApplyMerchantFee(ctx, req, constant.ReferencePlatformTransfer)
}

func (s *PlatformFeeService) ApplyMerchantTransactionFee(ctx context.Context, req platformFee.PlatformFeeRequest) error {

	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/platformFee/ApplyMerchantTransferFee")
	defer segment.End()

	return s.ApplyMerchantFee(ctx, req, constant.ReferencePlatformTransaction)
}

func (s *PlatformFeeService) ApplyMerchantFee(ctx context.Context, req platformFee.PlatformFeeRequest, referenceType string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/platformFee/ApplyMerchantTransferFee")
	defer segment.End()

	feeAmount, feeDetail, err := s.feeSvc.GetFeeCalculationAndDetail(ctx, &feeModel.GetFeeRequest{
		MerchantID:      req.MerchantID,
		Reference:       referenceType,
		ReferenceAmount: req.Amount,
	})
	if err != nil {
		s.logger.Error(ctx, "error when get platform fee detail.", logger.Error(err), logger.Any("referenceType", referenceType), logger.Any("request", req))
		return err
	}
	if feeAmount <= 0 {
		return nil
	}

	merchantUUID := uuid.MustParse(req.MerchantID)
	account, err := s.accountSvc.GetAccountByReferenceIDAndUsecase(ctx, merchantUUID, req.Usecase, constant.UserTypeMerchant)
	if err != nil {
		s.logger.Error(ctx, "error when get merchant account", logger.Error(err), logger.Any("referenceType", referenceType), logger.Any("request", req))
		return err
	}

	ledgerRequest := &ledger_model.CreateNewLedgerEntryRequest{
		ReferenceID:          req.ReferenceID,
		Usecase:              constant.ReferencePlatform,
		TransactionType:      constant.TypeFee,
		Channel:              "",
		TransactionTimestamp: time.Now().UTC(),
		Amount:               feeAmount,
		Currency:             constant.CurrencyIDR,
		TransferType:         constant.TransferTypeCharge,
		SenderAccountID:      account.UUID,
		SenderID:             account.ReferenceID,
		MoneyFlowType:        constant.MoneyFlowDirect,
		SenderAdditionalInfo: feeDetail,
		ChargeConfig: ledger_model.ChargeConfig{
			BypassBalanceCheck: true,
			IsDirectlyDeducted: false,
		},
	}
	if feeDetail.DeductionType == constant.MerchantFeeDeductionTypeDirect {
		ledgerRequest.TransferType = constant.TransferTypeCharge
		ledgerRequest.ChargeConfig.IsDirectlyDeducted = true
	}

	err = s.ledgerSvc.RecordTransaction(ctx, req.MerchantID, ledgerRequest)
	if err != nil {
		s.logger.Error(ctx, "Error when apply fee to merchant in platform operation.", logger.Error(err), logger.Any("referenceType", referenceType), logger.Any("request", req))
		return err
	}
	return nil

}
