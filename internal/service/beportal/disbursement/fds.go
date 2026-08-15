package disbursementService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	common "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementService) ExternalFDS(ctx context.Context, payout *disbursementModel.DisbursementWithTransaction, ledger *orchestratorModel.TransactionAndFeeObject) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ExternalFDS")
	defer segment.End()

	if !constant.IsPayoutExternalFDSEnabled(s.config.Environment) {
		return nil
	}

	merchant, err := s.merchantRepo.FindMerchantByID(ctx, payout.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "Failed to fetch merchant data", logger.Error(err))
		return fmt.Errorf("%s:%s", "External FDS:", err)

	} else if merchant == nil {
		return fmt.Errorf("%s:%s", "External FDS:", constant.ErrMerchantNotFound)
	}

	request := s.toAssessPayoutTransactionRequest(merchant, payout)

	fdsResponse, err := s.workflowFDSRepo.AssessPayoutTransaction(ctx, request)
	if err != nil {
		// Ignore external FDS errors and mark the result as ERROR
		fdsResponse = &fdscommon.TransactionAssessmentResponse{
			Result: constant.WorkflowFDSResultError,
		}
	}
	if err := s.accountTransactionRepo.UpdateFDSRiskAssessmentResultByID(ctx, ledger.TransactionID, fdsResponse); err != nil {
		s.logger.Error(ctx, "Failed to update risk assessment", logger.Error(err))
		return fmt.Errorf("%s:%s", "External FDS:", err)
	}

	if constant.IsPayoutFDSResultAllowed(fdsResponse.Result) {
		return nil
	}

	if err := s.self.FailTransactionByFDSResult(ctx, payout.UUID, ledger); err != nil {
		s.logger.Error(ctx, "Failed to update transaction status to blocked based on FDS risk assessment", logger.Error(err))
		return fmt.Errorf("%s:%s", "External FDS:", err)
	}
	return constant.ErrBlockedByFDS
}

func (s *DisbursementService) FailTransactionByFDSResult(ctx context.Context, payoutID string, ledger *orchestratorModel.TransactionAndFeeObject) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/FailTransactionByFDSResult")
	defer segment.End()

	if ctx, err = s.disbursementRepo.BeginTransaction(ctx); err != nil {
		return err
	}
	defer func() {
		if err == nil {
			return
		}

		if errRollback := s.disbursementRepo.RollbackTransaction(ctx); errRollback != nil {
			s.logger.Error(ctx, "Failed to rollback payout transaction status update", logger.Error(errRollback))
		}
	}()

	reasonType := constant.ReasonTypeBlockedByFDS
	reasonDesc := constant.ReasonDescBlockedByFDS
	if err = s.updateTransactionStatusWithHistory(ctx, ledger, constant.StatusFailed, &reasonType, &reasonDesc, payoutID); err != nil {
		return err
	}
	return s.disbursementRepo.CommitTransaction(ctx)
}

func (DisbursementService) toAssessPayoutTransactionRequest(merchant *merchant.Merchant, payout *disbursementModel.DisbursementWithTransaction) fdscommon.AssessPayoutTransactionRequest {
	return fdscommon.AssessPayoutTransactionRequest{
		Merchant: fdscommon.Merchant{
			ID:        merchant.UUID,
			Name:      merchant.Name,
			RiskLevel: merchant.RiskLevel.String,
		},
		Transaction: fdscommon.Transaction{
			ID:                payout.UUID,
			ClientReferenceID: payout.ReferenceID,
			Amount: common.Amount2{
				Value:    payout.Amount.InexactFloat64(),
				Currency: payout.Currency,
			},
			CreatedAt:   payout.CreatedAt,
			UpdatedAt:   payout.UpdatedAt,
			CreatedFrom: util.ValueOfPtr(payout.CreatedFrom),
		},
		Destination: fdscommon.PayoutDestination{
			BankCode:      payout.BeneficiaryBankCode,
			AccountNumber: payout.BeneficiaryAccountNo,
			AccountName:   payout.BeneficiaryAccountName,
		},
		Metadata: map[string]any{},
	}
}
