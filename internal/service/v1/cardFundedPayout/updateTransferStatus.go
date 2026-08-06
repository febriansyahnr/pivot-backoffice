package cardFundedPayoutService

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *service) UpdateBankTransferStatus(ctx context.Context, request *routingProcessorModel.BankTransferResponseData) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/UpdateBankTransferStatus")
	defer span.End()

	// In this case, only transactions with a pending status can have their status updated.
	if request.Transaction.Status != constant.StatusPending {
		s.logger.Info(ctx, "Transaction status is final, status cannot be updated. Bank transfer status update ignored", logger.String("status", request.Transaction.Status))
		return nil
	}

	payout, err := s.disbursementRepo.GetDetailForCardFundedPayoutByID(ctx, request.Transaction.ReferenceID)
	if err != nil {
		s.logger.Error(ctx, "Failed when get detail for card-funded payout", logger.Error(err))
		return err
	}
	if payout == nil {
		return constant.ErrPayoutIsNotFound
	}

	err = s.disbursementRepo.UpdateProcessorReferenceIdAndBankReferenceNo(
		ctx, payout.UUID, request.UUID, request.BankReferenceNo,
	)
	if err != nil {
		return fmt.Errorf("failed update payout processor reference id and bank reference no: %w", err)
	}

	err = s.orchestratorSvc.UpdateProcessorAndReconReferenceByID(
		ctx, request.Transaction.UUID.String(), request.ProcessorReference, request.UUID, request.GetReconReferenceNo(),
	)
	if err != nil {
		return fmt.Errorf("failed update processor and recon reference: %w", err)
	}

	bankTransferResponse := request.ToSnapBankTransferResponseData()

	var (
		status = constant.StatusSuccess
		// Only not null when the status is not SUCCESS
		reasonType, reasonDescription *string
	)
	if bankTransferResponse.Status != constant.SnapCoreBankTransferStatusSuccess {
		// Make memory allocations because the default value of the following variables is nil.
		reasonType, reasonDescription = util.ValueToPtr(""), util.ValueToPtr("")

		// Mapping transfer status to transaction status
		status, *reasonType, *reasonDescription = bankTransferResponse.MappingAccountTransactionErrStatus()
	}
	if err := s.orchestratorSvc.UpdateStatusAccountTransaction(ctx, request.Transaction.UUID.String(), status, reasonType, reasonDescription); err != nil {
		return fmt.Errorf("failed update status account transaction: %w", err)
	}

	if status == constant.StatusPending {
		return nil
	}

	_ = s.recordStatusHistory(ctx, payout.UUID, status, constant.StatusHistoryActorSystem, "")

	s.logger.Info(
		ctx, "Update bank transfer status for card-funded payout transaction",
		logger.Any("details", map[string]string{
			"id":                     payout.UUID,
			"merchantId":             payout.MerchantID,
			"beneficiaryBankName":    util.ValueOfPtr(payout.BeneficiaryBankName),
			"beneficiaryAccountNo":   payout.BeneficiaryAccountNo,
			"beneficiaryAccountName": payout.BeneficiaryAccountName,
			"currency":               payout.Currency,
			"amount":                 fmt.Sprintf("%.2f", payout.Amount.InexactFloat64()),
		}),
		logger.String("status", status),
		logger.String("reasonType", util.ValueOfPtr(reasonType)),
		logger.String("reasonDescription", util.ValueOfPtr(reasonDescription)),
	)
	return nil
}

func (s *service) UpdatePayoutTransactionStatus(ctx context.Context, request model.PatchPayoutTransactionStatusRequest) (*model.PayoutActionResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/UpdatePayoutTransactionStatus")
	defer span.End()

	payout, err := s.disbursementRepo.GetDetailForCardFundedPayoutByID(ctx, request.PayoutID)
	if err != nil {
		s.logger.Error(ctx, "Failed to retrieve card-funded payout detail", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	if payout == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("payout with ID %s not found", request.PayoutID))
	}
	if payout.Status != constant.DisbursementStatusApproved {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("payout must be in APPROVED status; current status is %s", payout.Status))
	}

	transaction, err := s.orchestratorSvc.FindByReference(ctx, payout.UUID, constant.ReferenceDisbursement)
	if err != nil {
		s.logger.Error(ctx, "Failed to find transaction by reference and reference type", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	if transaction == nil {
		return nil, pkgErrs.New(response.HttpErrNotFound, fmt.Errorf("transaction with reference id %s and reference type %s not found", request.PayoutID, constant.ReferenceDisbursement))
	}
	if transaction.Status != constant.StatusPending {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, fmt.Errorf("transaction must be in PENDING status; current status is %s", transaction.Status))
	}

	ctxTx, err := s.disbursementRepo.BeginTransaction(ctx)
	if err != nil {
		s.logger.Error(ctx, "Failed begin transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	isComplete := false
	defer func() {
		if isComplete {
			return
		}
		if rollbackErr := s.disbursementRepo.RollbackTransaction(ctxTx); rollbackErr != nil {
			s.logger.Error(ctx, "Failed rollback transaction", logger.Error(rollbackErr))
		}
	}()

	err = s.disbursementRepo.UpdateProcessorReferenceIdAndBankReferenceNo(ctxTx, payout.UUID, "", request.BankReferenceNo)
	if err != nil {
		s.logger.Error(ctx, "Failed update payout processor reference id and bank reference no", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	err = s.orchestratorSvc.UpdateProcessorAndReconReferenceByID(ctxTx, transaction.UUID.String(), constant.ManualProcessor, "", request.ReconReferenceNo)
	if err != nil {
		s.logger.Error(ctx, "Failed update processor and recon reference", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	err = s.orchestratorSvc.UpdateStatusAccountTransaction(ctxTx, transaction.UUID.String(), request.Status, request.ReasonType, request.ReasonDescription)
	if err != nil {
		s.logger.Error(ctx, "Failed update status account transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	if err = s.recordStatusHistory(ctxTx, payout.UUID, request.Status, constant.StatusHistoryActorSystem, ""); err != nil {
		s.logger.Warn(ctx, "Failed to record status history", logger.Error(err))
	}

	if err := s.disbursementRepo.CommitTransaction(ctxTx); err != nil {
		s.logger.Error(ctx, "Failed commit transaction", logger.Error(err))
		return nil, pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	isComplete = true

	cardFundedDetail := payout.GetCardFundedPayoutDetail()
	return &model.PayoutActionResponse{
		ID:            payout.UUID,
		VendorID:      cardFundedDetail.VendorID,
		VendorName:    cardFundedDetail.VendorName,
		ReferenceID:   payout.ReferenceID,
		BankCode:      payout.BeneficiaryBankCode,
		BankName:      util.ValueOfPtr(payout.BeneficiaryBankName),
		AccountNumber: payout.BeneficiaryAccountNo,
		AccountName:   payout.BeneficiaryAccountName,
		Amount: commonModel.AmountRequest{
			Currency: payout.Currency,
			Value:    transaction.Debit,
		},
		Remarks:           util.ValueOfPtr(payout.Remark),
		SettlementMethod:  cardFundedDetail.SettlementMethod,
		CardID:            cardFundedDetail.Card.ID,
		CardName:          cardFundedDetail.Card.CardName,
		MerchantID:        payout.MerchantID,
		Status:            request.Status,
		BankReferenceNo:   request.BankReferenceNo,
		ReconReferenceNo:  request.ReconReferenceNo,
		ReasonType:        request.ReasonType,
		ReasonDescription: request.ReasonDescription,
	}, nil
}
