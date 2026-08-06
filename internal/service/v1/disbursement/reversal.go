package disbursementService

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
)

// Note: Disbursement transaction reversal can only be performed for transactions with a SUCCESS status.
func (s *DisbursementService) Reversal(ctx context.Context, request *disbursementModel.ReversalTransactionReq) (result *disbursementModel.ReversalTransactionResp, err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/Reversal")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	disbursement, err := s.disbursementRepo.FindForReversalDisbursementById(ctx, request.MerchantId, request.DisbursementId)
	if err != nil {
		s.logger.Error(ctx, "Getting disbursement data", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))

	} else if disbursement == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("disbursement data not found"))

	} else if disbursement.Status != constant.DisbursementStatusApproved {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("disbursement status must be approved"))

	} else if disbursement.ReasonType.String == constant.ReasonTypeReversal {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("disbursement has been REVERSAL"))

	} else if disbursement.Transaction.Status != constant.StatusSuccess {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("transaction must be have SUCCESS status"))
	}

	isCompleted := false
	merchant, err := s.merchantSvc.FindMerchantByID(ctx, disbursement.MerchantId)
	if err != nil {
		s.logger.Error(ctx, "Getting merchant data", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	if merchant == nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("disbursement merchant data not found"))
	}
	var parentMerchantId string
	if merchant.ParentID.String != "" {
		parentMerchantId = merchant.ParentID.String
	}

	ctxTrx, err := s.disbursementRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "Start session transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}
	defer func() {
		if isCompleted {
			return
		}
		if e := s.disbursementRepo.RollbackTransaction(ctxTrx); e != nil {
			s.logger.Error(ctx, "Rollback active transaction", logger.Error(e))
		}
	}()

	reversalReasonType := constant.ReasonTypeReversal
	merchantId, _ := uuid.Parse(disbursement.MerchantId)

	rawReversalMetadata, _ := json.Marshal(&disbursementModel.ReversalMetadataObject{
		TransactionId: disbursement.Transaction.Id, FeeId: disbursement.Fee.Id,
	})

	accountTransaction := &orchestratorModel.CreateAccountTransactionRequest{
		UUID:                 uuid.New(),
		ReferenceID:          disbursement.Id,
		Type:                 constant.TypeDisbursement,
		Channel:              constant.ChannelManualAction,
		ReasonType:           &reversalReasonType,
		MerchantID:           merchantId,
		Currency:             disbursement.Currency,
		Credit:               disbursement.Amount,
		Status:               constant.StatusSuccess,
		Remarks:              request.Reason,
		TransactionTimestamp: time.Now().UTC(),
		AdditionalInfo: types.NullJSONText{
			Valid: true, JSONText: rawReversalMetadata,
		},
	}

	if disbursement.IsFeeStatus(constant.StatusPending) &&
		disbursement.IsFeeDeductionType(constant.MerchantFeeDeductionTypeAutomated, constant.MerchantFeeDeductionTypeManual) {

		// Note: In this section, it needs to be clarified whether when the deduction type is MANUAL or AUTOMATED (Pending),
		//       the transaction status should remain the same or otherwise.
		// 		 A discussion with the product team is necessary to determine the desired behavior.
		if err = s.accountTransactionRepo.CancelIndirectTransactionFee(ctxTrx, disbursement.Fee.Id, time.Now().UTC()); err != nil {
			s.logger.Error(ctx, "Cancel indirect transaction fee", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
		}

	} else if disbursement.IsFeeStatus(constant.StatusSuccess) &&
		disbursement.IsFeeDeductionType(constant.MerchantFeeDeductionTypeDirect, constant.MerchantFeeDeductionTypeAutomated) {

		if parentMerchantId == "" {
			accountTransaction.Credit += disbursement.Fee.Amount
		}
	}

	if parentMerchantId != "" {
		if disbursement.Fee.Amount > 0 {
			s.logger.Info(ctx, "Reverse parent merchant disbursement fee")
			feeMetadataB, _ := json.Marshal(disbursement.Fee.Metadata)
			parentFeeAccountTrxReq := &orchestratorModel.CreateAccountTransactionRequest{
				UUID:                 uuid.New(),
				ReferenceID:          disbursement.Id,
				Type:                 constant.TypeFee,
				Channel:              constant.ChannelManualAction,
				ReasonType:           &reversalReasonType,
				MerchantID:           util.ParseUUID(parentMerchantId),
				Currency:             disbursement.Currency,
				Credit:               disbursement.Fee.Amount,
				Status:               constant.StatusSuccess,
				Remarks:              request.Reason,
				TransactionTimestamp: time.Now().UTC(),
				AdditionalInfo: types.NullJSONText{
					Valid: true, JSONText: feeMetadataB,
				},
			}
			if err = s.orchestratorSvc.PostAccountTransaction(ctxTrx, parentFeeAccountTrxReq); err != nil {
				s.logger.Error(ctx, "Error reverse parent merchant disbursement fee", logger.Error(err))
				return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
			}
		}

		s.logger.Info(ctx, "Reverse platform transfer transaction")
		_, err = s.transferSvc.ReverseTransfer(ctxTrx, &transfer.ReverseTransferRequest{
			ReferenceID:      disbursement.Id,
			MerchantID:       disbursement.MerchantId,
			ParentMerchantID: parentMerchantId,
			Usecase:          constant.UseCaseDisbursement,
			Remarks:          reversalReasonType,
		})
		if err != nil {
			s.logger.Error(ctx, "Error trigger reverse platform transfer", logger.Error(err))
			return nil, pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))
		}

		s.logger.Info(ctx, "Finish reverse platform transfer transaction")
	}

	if err = s.orchestratorSvc.PostAccountTransaction(ctxTrx, accountTransaction); err != nil {
		s.logger.Error(ctx, "Post account transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	err = s.disbursementRepo.UpdateReversalTransaction(ctxTrx, disbursement.Id, constant.ReasonTypeReversal, request.Reason, request.CreatedBy)
	if err != nil {
		s.logger.Error(ctx, "Update disbursement reason type & description", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	if err = s.disbursementRepo.CommitTransaction(ctxTrx); err != nil {
		s.logger.Error(ctx, "Commit transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	isCompleted = true

	return &disbursementModel.ReversalTransactionResp{
		Id:             accountTransaction.UUID.String(),
		DisbursementId: disbursement.Id,
		ReversalAmount: accountTransaction.Credit,
	}, nil
}
