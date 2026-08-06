package transferService

import (
	"context"
	stdErrors "errors"
	"strings"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platformFee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *TransferService) Transfer(ctx context.Context, request *transfer.TransferRequest) (*transfer.Transfer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/transfer/Transfer")
	defer segment.End()

	if recipientMerchant, err := s.merchantSvc.FindMerchantByID(ctx, request.RecipientID); err != nil {
		return nil, errPkg.New(response.HttpErrDatabase, err)

	} else if recipientMerchant == nil {
		return nil, errPkg.New(response.HttpErrNotFound, constant.ErrRecipientIdNotFound)
	}

	transferData, err := transfer.NewTransfer(request)
	if err != nil {
		// Same-merchant transfer is a business rule violation (valid request, invalid state) → 422.
		// Other NewTransfer errors (invalid amount, invalid transfer type, etc.) are format issues → 400.
		if stdErrors.Is(err, constant.ErrSameMerchant) {
			return nil, errPkg.New(response.HttpErrUnprocessableContent, err)
		}
		return nil, errPkg.New(response.HttpErrRequest, err)
	}

	merchantIds := []uuid.UUID{
		request.SourceMerchantID,
		util.ParseUUID(request.RecipientID),
		request.ParentMerchantID,
	}

	usecase := constant.ReferencePlatform
	if request.Usecase != "" {
		usecase = request.Usecase
	}
	accountMap, err := s.accountSvc.GetMerchantAccounts(ctx, merchantIds, usecase)
	if err != nil {
		return nil, errPkg.New(response.HttpErrDatabase, constant.ErrGetAccounts)
	}

	if request.Usecase != constant.TypePayment {
		existingTransfer, err := s.repo.GetByReferenceID(ctx, request.SourceMerchantID.String(), request.RecipientID, request.ReferenceID)
		if err != nil {
			return nil, errPkg.New(response.HttpErrDatabase, constant.ErrCheckReference)
		}
		if existingTransfer != nil {
			// Duplicate referenceID is a conflict on an existing resource → 409.
			return nil, errPkg.New(response.HttpErrDupCheck, constant.ErrReferenceIdExist)
		}
	}

	ledgerRequest, err := s.getLedgerRequest(ctx, transferData, accountMap, request.ParentMerchantID)
	if err != nil {
		// Sender/recipient account not found for the requested usecase → 404.
		if stdErrors.Is(err, constant.ErrSenderAccountNotFound) || stdErrors.Is(err, constant.ErrRecipientAccountNotFound) {
			return nil, errPkg.New(response.HttpErrNotFound, err)
		}
		return nil, errPkg.New(response.HttpErrRequest, err)
	}

	if err = s.repo.Create(ctx, transferData); err != nil {
		return nil, errPkg.New(response.HttpErrDatabase, constant.ErrCreateTransfer)
	}

	err = s.ledgerSvc.RecordTransaction(ctx, request.ParentMerchantID.String(), ledgerRequest)
	if err != nil {
		s.logger.Error(ctx, "error when record transaction into ledger", logger.Error(err))

		transferData.Update(constant.TransferStatusFailed, err.Error())
		updateErr := s.repo.Update(ctx, transferData)
		if updateErr != nil {
			return nil, errPkg.New(response.HttpErrDatabase, constant.ErrUpdateTransfer)
		}
		// Insufficient balance bubbles up from ledger.RecordTransaction → 403 for consistency
		// with payout/inquiry handling (see /payouts/{id}/retry and /inquiry-account).
		if stdErrors.Is(err, constant.ErrInsufficientBalance) {
			return nil, errPkg.New(response.HttpErrForbidden, err)
		}
		return nil, err
	}

	if set, _ := ctx.Value(constant.CtxSetPendingTransaction).(bool); !set {

		transferData.Update(constant.TransferStatusSuccess, "")
		if err := s.repo.Update(ctx, transferData); err != nil {
			return nil, errPkg.New(response.HttpErrDatabase, constant.ErrUpdateTransfer)
		}
	}

	applyFeeReq := platformFee.PlatformFeeRequest{
		MerchantID:  request.ParentMerchantID.String(),
		Amount:      transferData.Amount,
		ReferenceID: transferData.UUID.String(),
		Usecase:     usecase,
	}
	s.platformFeeSvc.ApplyMerchantTransferFee(ctx, applyFeeReq)
	s.platformFeeSvc.ApplyMerchantTransactionFee(ctx, applyFeeReq)

	return transferData, nil
}

func (s *TransferService) getLedgerRequest(ctx context.Context, request *transfer.Transfer, accountMap map[uuid.UUID]*account_model.Account, parentId uuid.UUID) (*ledger_model.CreateNewLedgerEntryRequest, error) {
	sourceAccount := accountMap[request.MerchantID]
	if sourceAccount == nil {
		return nil, constant.ErrSenderAccountNotFound
	}
	recipientAccount := accountMap[request.RecipientID]
	if recipientAccount == nil {
		return nil, constant.ErrRecipientAccountNotFound
	}

	ledgerRequest := &ledger_model.CreateNewLedgerEntryRequest{
		ReferenceID:          request.UUID.String(),
		Usecase:              constant.ReferencePlatform,
		Channel:              "",
		TransactionType:      constant.TypeTransfer,
		Remarks:              request.Remarks,
		Amount:               request.Amount,
		TransactionTimestamp: request.CreatedAt,
		Currency:             constant.CurrencyIDR,
		TransferType:         constant.TransferTypeP2P,
		MoneyFlowType:        strings.ToUpper(request.TransferType),
		RecipientID:          request.RecipientID,
		RecipientAccountID:   recipientAccount.UUID,
		SenderID:             request.MerchantID,
		SenderAccountID:      sourceAccount.UUID,
		ParentID:             parentId,
		ParentAccountID:      accountMap[parentId].UUID,
	}

	if set, _ := ctx.Value(constant.CtxSetBypassBalanceCheckTransaction).(bool); set {
		ledgerRequest.P2PConfig = ledger_model.P2PConfig{
			BypassBalanceCheck: set,
		}
	}

	return ledgerRequest, nil
}

func (s *TransferService) UpdateTransferStatus(ctx context.Context, merchantId, transferId, status string, description *string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/transfer/UpdateTransferStatus")
	defer segment.End()

	request := &transfer.Transfer{
		MerchantID: util.ParseUUID(merchantId),
		UUID:       util.ParseUUID(transferId),
		Status:     status,
	}
	if description != nil {
		request.ReasonDescription = *description
	}
	return s.repo.Update(ctx, request)
}
