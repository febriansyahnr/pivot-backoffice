package withdrawalService

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// RetryTransaction retries a failed/pending withdrawal bank transfer.
// It validates the withdrawal, checks retry eligibility with snap-core,
// then re-triggers the bank transfer using the existing account transaction external ID.
func (s *withdrawalService) RetryTransaction(ctx context.Context, request *withdrawal.RetryTransactionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/RetryTransaction")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	// Get the withdrawal record
	wd, err := s.repo.FindById(ctx, request.WithdrawalID, request.MerchantID)
	if err != nil {
		return pkgErrs.New(response.HttpErrDatabase, err)
	}
	if wd == nil {
		return pkgErrs.New(response.HttpErrNotFound, constant.ErrDataNotFound)
	}

	// Validate this is a bank transfer withdrawal (not balance transfer)
	if wd.Metadata.WithdrawType == constant.WithdrawalDestBalanceTransfer {
		return pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("retry is not supported for balance transfer withdrawals"))
	}

	// Get the account transaction for this withdrawal (contains the external ID for snap-core)
	accountTransaction, err := s.accountTrxRepo.FindByReference(ctx, request.WithdrawalID, constant.TypeWithdrawal)
	if err != nil {
		return pkgErrs.New(response.HttpErrDatabase, err)
	}
	if accountTransaction == nil {
		return pkgErrs.New(response.HttpErrNotFound, errors.New("account transaction not found for this withdrawal"))
	}

	// Validate transaction status is not already SUCCESS
	if accountTransaction.Status == constant.StatusSuccess {
		return pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrTransactionAlreadyInFinalStatus)
	}

	externalID := accountTransaction.UUID.String()

	s.logger.Info(ctx, "Retry withdrawal transaction",
		logger.String("withdrawalID", request.WithdrawalID),
		logger.String("externalID", externalID),
		logger.Bool("forceFailed", request.ForceFailed),
		logger.Bool("forceRetry", request.ForceRetry),
		logger.Bool("bypassProcessorCheck", request.BypassProcessorCheck),
	)

	// Store forceFailed and fromRetry flags in context for downstream passthrough
	ctx = context.WithValue(ctx, constant.CtxForceFailed, request.ForceFailed)
	ctx = context.WithValue(ctx, constant.CtxFromRetry, true)

	// Check snap-core retry eligibility (unless bypassed)
	if !request.BypassProcessorCheck {
		checkRetryResult, errCheck := s.snapCoreRepo.CheckAllowedToRetry(ctx, snapCoreModel.CheckAllowedToRetryRequest{
			ExternalID: externalID,
			MerchantId: request.MerchantID,
			Force:      request.ForceRetry,
		})
		if errCheck != nil {
			return errCheck
		}
		if !checkRetryResult.Allowed {
			return pkgErrs.New(response.HttpErrRequest, errors.New(checkRetryResult.Reason))
		}
	}

	// Re-trigger bank transfer via snap-core
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceId,
		From:        "Withdrawal-Retry",
		OriginId:    request.WithdrawalID,
		ReferenceId: request.MerchantID,
	})

	snapCoreResp, err := s.snapCoreRepo.BankTransfer(ctx, &snapCoreModel.BankTransferRequest{
		BTBeneficiaryRequest: snapCoreModel.BTBeneficiaryRequest{
			BeneficiaryBankCode:    wd.BeneficiaryBankCode,
			BeneficiaryAccountNo:   wd.BeneficiaryAccountNo,
			BeneficiaryAccountName: wd.BeneficiaryAccountName,
		},
		Currency:             constant.CurrencyIDR,
		Amount:               commonModel.Amount{Currency: constant.CurrencyIDR, Value: fmt.Sprintf("%.0f", wd.Amount)},
		Remark:               wd.Id[24:],
		PurposeOfTransaction: snapCoreModel.DefaultPurchaseOfTransaction,
		TransactionDate:      wd.CreatedAt,
	}, &snapCoreModel.BankTransferHeaderRequest{
		ExternalId: externalID,
		MerchantId: request.MerchantID,
	})

	// Update withdrawal metadata with new snap-core response
	if snapCoreResp != nil && snapCoreResp.UUID != "" {
		wd.Metadata.BankTransfer = &withdrawal.BankTransfer{
			UUID:               snapCoreResp.UUID,
			ExternalId:         snapCoreResp.ExternalID,
			BankReferenceNo:    snapCoreResp.BankReferenceNo,
			PartnerReferenceNo: snapCoreResp.PartnerReferenceNo,
		}
		if errUpdate := s.repo.UpdateMetadataById(ctx, request.WithdrawalID, &wd.Metadata); errUpdate != nil {
			s.logger.Error(ctx, "Update withdrawal metadata (retry bank transfer)", logger.Error(errUpdate))
		}

		if e := s.orchestratorSvc.UpdateProcessorAndReconReferenceByID(ctx, externalID, constant.SnapCoreProcessor, snapCoreResp.UUID, snapCoreResp.GetReconReferenceNo()); e != nil {
			s.logger.Error(ctx, "Update account transactions additional info (retry)", logger.Error(e))
		}
	}

	// Update account transaction status based on snap-core response
	if err != nil {
		accTrxStatus, reasonType, reasonDesc := constant.StatusFailed, constant.ReasonTypeOtherReason, ""
		if snapCoreResp != nil {
			accTrxStatus, reasonType, reasonDesc = snapCoreResp.MappingAccountTransactionErrStatus()
		}

		if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, externalID, accTrxStatus, &reasonType, &reasonDesc); errUpdate != nil {
			s.logger.Error(ctx, "Update status account transaction (retry failed)", logger.Error(errUpdate))
		}

		s.logger.Info(ctx, "Withdrawal retry bank transfer failed",
			logger.String("withdrawalID", request.WithdrawalID),
			logger.String("status", accTrxStatus),
			logger.String("reasonType", reasonType),
		)
		return err

	} else if snapCoreResp != nil && snapCoreResp.Status == constant.SnapCoreBankTransferStatusSuccess {
		if errUpdate := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, externalID, constant.StatusSuccess, nil, nil); errUpdate != nil {
			s.logger.Error(ctx, "Update status account transaction (retry success)", logger.Error(errUpdate))
		}
	}

	s.logger.Info(ctx, "Withdrawal retry transaction completed",
		logger.String("withdrawalID", request.WithdrawalID),
		logger.String("externalID", externalID),
	)

	// Send callback for Open API-sourced withdrawals with final status
	if wd.Metadata.Source == constant.SourceOpenApi && snapCoreResp != nil {
		status := constant.StatusPending
		if snapCoreResp.Status == constant.SnapCoreBankTransferStatusSuccess {
			status = constant.StatusSuccess
		}
		if status != constant.StatusPending {
			callbackRequest := withdrawal.WithdrawalStatusCallbackRequest{
				ID:         wd.Id,
				MerchantId: wd.MerchantId,
				Withdrawal: withdrawal.OpenAPIWithdrawalDetailResponse{
					ReferenceID:  wd.ReferenceId,
					WithdrawType: wd.Metadata.WithdrawType,
					IsFullAmount: wd.Metadata.IsFullAmount,
					Amount: &commonModel.Amount{
						Currency: constant.CurrencyIDR,
						Value:    fmt.Sprintf("%.0f", wd.Amount),
					},
					Description: wd.Description,
				},
				Status:    status,
				CreatedAt: wd.CreatedAt.Format(time.RFC3339),
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			}
			if errCallback := s.SendWithdrawalStatusCallback(ctx, callbackRequest); errCallback != nil {
				s.logger.Error(ctx, "Failed to send withdrawal final status (retry)", logger.Error(errCallback), logger.Any("callbackRequest", callbackRequest))
			}
		}
	}

	return nil
}
