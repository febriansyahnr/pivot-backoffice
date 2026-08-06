package platformFeeService

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platformFee"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PlatformFeeService) ReverseMerchantFee(ctx context.Context, req platformFee.PlatformReversalFeeRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/platformFee/ReverseMerchantFee")
	defer segment.End()

	ledgerDetails, err := s.ledgerSvc.GetLedgerDetail(ctx, req.ReferenceID)
	if err != nil {
		s.logger.Error(ctx, "error when get merchant platform fee ledger details", logger.Error(err), logger.Any("request", req))
		return err
	}
	if len(ledgerDetails) == 0 {
		s.logger.Info(ctx, "no merchant platform fee ledger details found", logger.Any("request", req))
		return nil
	}

	for _, ledgerDetail := range ledgerDetails {
		if ledgerDetail.Reference == constant.ReferencePlatform && ledgerDetail.Type == constant.TypeFee {
			ledgerRequest := &ledger_model.CreateNewLedgerEntryRequest{
				ReferenceID:             req.ReversalReferenceID,
				Usecase:                 constant.ReferencePlatform,
				TransactionType:         constant.TypeFee,
				Channel:                 "",
				TransactionTimestamp:    time.Now().UTC(),
				Amount:                  ledgerDetail.Debit,
				Currency:                constant.CurrencyIDR,
				TransferType:            constant.TransferTypePayIn,
				RecipientAccountID:      ledgerDetail.AccountID,
				RecipientID:             ledgerDetail.MerchantID,
				MoneyFlowType:           constant.MoneyFlowDirect,
				RecipientAdditionalInfo: ledgerDetail.AdditionalInfo,
				Remarks:                 req.Remarks,
				ChargeConfig: ledger_model.ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			}

			err = s.ledgerSvc.RecordTransaction(ctx, req.MerchantID, ledgerRequest)
			if err != nil {
				s.logger.Error(ctx, "Error when credit fee to merchant in platform reverse operation.", logger.Error(err), logger.Any("request", ledgerRequest))
				return err
			}
		}
	}
	s.logger.Info(ctx, "finish reverse platform merchant fee")

	return nil

}
