package settlementService

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	orchestraModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *settlementService) ProcessSettlement(ctx context.Context, request *settlementModel.ProcessSettlementRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/settlement/ProcessSettlement")
	defer segment.End()

	s.logger.Info(ctx, "Start ProcessSettlement", logger.Any("request", request))
	defer func() {
		s.logger.Info(ctx, "Finish ProcessSettlement")
	}()

	if request.Type == constant.SettlementFeeOnly {
		return s.ProcessSettlementTransactionFee(ctx, request.MerchantID, request.TransactionID)
	}

	transaction, err := s.accountTransactionRepo.FindByID(ctx, request.TransactionID)
	if err != nil {
		return err

	} else if transaction == nil {
		s.logger.Warn(ctx, "ProcessSettlement - Find transaction by id is not found", logger.Any("request", request))
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)

	} else if transaction.MerchantID.String() != request.MerchantID {
		return pkgErr.New(response.HttpErrRequest, constant.ErrMerchantIsNotMatch)
	}

	if transaction.SettlementStatus.String != constant.SettlementStatusPending {
		s.logger.Warn(ctx, "Transaction settlement is not pending", logger.Any("request", request), logger.String("settlementStatus", transaction.SettlementStatus.String))
		return nil
	}
	additionalInfo, ok := transaction.AdditionalInfoObj.(orchestraModel.FeeTransactionMetadataObject)
	if !ok {
		s.logger.Warn(ctx, "Unable to parse transaction additional info to destined type.", logger.Any("transaction", transaction))
	}
	if !request.ByPassSettlementHold && (additionalInfo.AccountTransactionMetadataObject != nil && additionalInfo.AccountTransactionMetadataObject.SettlementDetail.IsOnHold) {
		s.logger.Info(ctx, "Current transaction is on hold, unable to settle transaction yet.", logger.Any("settlementDetail", additionalInfo.AccountTransactionMetadataObject.SettlementDetail))
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentChargeIsOnHold)
	}

	if err := s.processMainSettlementAndFee(ctx, request, transaction); err != nil {
		return err
	}

	if transaction.Type == constant.TypePayment {
		if err := s.processSplitRoute(ctx, transaction.ReferenceID); err != nil {
			s.logger.Error(ctx, "Failed to process split route, but main settlement completed",
				logger.Error(err),
				logger.String("referenceId", transaction.ReferenceID))
		}

		if transaction.Reference == constant.ReferencePaymentFundedPayout {
			err := s.cardFundedPayoutSvc.ProcessFinishCardFundedPayoutSettlement(ctx, &cardFundedPayoutModel.ProcessFinishCardFundedPayoutSettlementRequest{
				MerchantID:  transaction.MerchantID.String(),
				ReferenceID: transaction.ReferenceID,
			})
			if err != nil {
				s.logger.Warn(ctx, "error process finished card funded payout settlement", logger.Error(err))
			}
		}

	}

	return nil
}

// processMainSettlementAndFee handles main settlement and fee processing in one transaction
func (s *settlementService) processMainSettlementAndFee(ctx context.Context, request *settlementModel.ProcessSettlementRequest, transaction *orchestraModel.AccountTransactionWithUseCase) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/settlement/processMainSettlementAndFee")
	defer segment.End()

	ctxTrx, errCtx := s.accountTransactionRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.accountTransactionRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErr.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	trxMetadata := orchestraModel.MetadataPayment[any]{}

	err := s.accountTransactionRepo.UpdatePaymentTransactionStatusAndMetadataByID(
		ctxTrx, orchestraModel.UpdatePaymentTransactionRequest{
			LedgerId:         transaction.UUID.String(),
			SettlementStatus: util.ValueToPtr(constant.StatusSuccess),
			SettlementAt:     util.ValueToPtr(time.Now().UTC()),
		}, trxMetadata,
	)
	if err != nil {
		if errors.Is(err, constant.ErrDataNotFound) {
			return pkgErr.New(response.HttpErrUnprocessableContent, err)
		}
		return pkgErr.New(response.HttpErrDatabase, err)
	}

	if err := s.ProcessSettlementTransactionFee(ctxTrx, request.MerchantID, request.FeeTransactionID); err != nil {
		return err
	}

	if errCommit := s.accountTransactionRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return pkgErr.New(response.HttpErrDatabase, errCommit)
	}
	isCompleted = true

	return nil
}

// processSplitRoute handles split route processing in a separate transaction
func (s *settlementService) processSplitRoute(ctx context.Context, referenceID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/settlement/processSplitRoute")
	defer segment.End()

	ctxTrx, errCtx := s.accountTransactionRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.accountTransactionRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction for split route", logger.Error(pkgErr.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	// Process split route
	if errPrc := s.paymentSvc.ProcessSplitRoute(ctxTrx, referenceID); errPrc != nil {
		return errPrc
	}

	if errCommit := s.accountTransactionRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return pkgErr.New(response.HttpErrDatabase, errCommit)
	}
	isCompleted = true

	return nil
}

func (s *settlementService) ProcessSettlementTransactionFee(ctx context.Context, merchantId, transactionFeeId string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/settlement/ProcessSettlementTransactionFee")
	defer segment.End()

	merchant, err := s.merchantSvc.FindMerchantByID(ctx, merchantId)
	if err != nil {
		return err
	} else if merchant == nil {
		s.logger.Warn(ctx, "Merchant not found", logger.String("merchantId", merchantId))
		return pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	} else if merchant.ParentID.Valid {
		merchantId = merchant.ParentID.String
	}

	transactionFee, err := s.accountTransactionRepo.FindByID(ctx, transactionFeeId)
	if err != nil {
		s.logger.Error(ctx, "Failed while find transaction fee by id", logger.Error(err))
		return err

	} else if transactionFee == nil || transactionFee.Status != constant.StatusSuccess {
		return nil

	} else if transactionFee.MerchantID.String() != merchantId {
		return pkgErr.New(response.HttpErrRequest, constant.ErrMerchantIsNotMatch)
	}

	feeMetadataObj := orchestraModel.FeeTransactionMetadataObject{}
	if transactionFee.AdditionalInfo.Valid {
		_ = json.Unmarshal(transactionFee.AdditionalInfo.JSONText, &feeMetadataObj)
	}

	if feeMetadataObj.DeductionType == constant.MerchantFeeDeductionTypeManual {
		return nil
	}

	// update settlement_status & settlement_at
	if errUpdate := s.accountTransactionRepo.UpdateSettlementStatusAndSettlementAtByID(ctx, transactionFee.UUID.String(), constant.StatusSuccess, time.Now().UTC()); errUpdate != nil {
		s.logger.Error(ctx, "Failed while update settlement status and settlement at by id (fee)", logger.Error(errUpdate))
		return pkgErr.New(response.HttpErrDatabase, errUpdate)
	}

	// Update status fee on direct deduction
	if feeMetadataObj.DeductionType == "" || feeMetadataObj.DeductionType == constant.MerchantFeeDeductionTypeDirect {
		if err := s.accountTransactionRepo.UpdateStatusAccountTransaction(ctx, transactionFee.UUID.String(), constant.StatusSuccess, nil, nil); err != nil {
			s.logger.Error(ctx, "Failed while update status account transaction (fee)", logger.Error(err))
			return pkgErr.New(response.HttpErrDatabase, err)
		}
	}
	return nil
}
