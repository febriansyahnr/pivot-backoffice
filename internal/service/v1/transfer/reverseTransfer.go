package transferService

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platformFee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *TransferService) ReverseTransfer(ctx context.Context, request *transfer.ReverseTransferRequest) (*transfer.Transfer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/transfer/ReverseTransfer")
	defer segment.End()

	transfer, err := s.repo.GetByReferenceID(ctx, request.MerchantID, request.ParentMerchantID, request.ReferenceID)
	if err != nil {
		s.logger.Error(ctx, "error when retrieve existing transfer data", logger.Error(err), logger.Any("request", request))
		return nil, err
	}
	if transfer == nil {
		s.logger.Error(ctx, "transfer not found", logger.Any("request", request))
		return nil, constant.ErrTransferNotFound
	}

	reverseTransfer := transfer.ReverseTransfer(request)
	if err = s.repo.Create(ctx, reverseTransfer); err != nil {
		s.logger.Error(ctx, "error store reverse transfer", logger.Error(err))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrCreateTransfer)
	}

	err = s.processLedgerReverseTransfer(ctx, request, reverseTransfer)
	if err != nil {
		s.logger.Error(ctx, "error process ledger reverse transfer", logger.Error(err))
		reverseTransfer.Update(constant.TransferStatusFailed, err.Error())
		updateErr := s.repo.Update(ctx, reverseTransfer)
		if updateErr != nil {
			s.logger.Error(ctx, "error update failed transfer", logger.Error(updateErr))
			return nil, errPkg.New(response.HttpErrInternal, constant.ErrUpdateTransfer)
		}
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrReverseTransfer)
	}

	reverseTransfer.Update(constant.TransferStatusSuccess, "")
	updateErr := s.repo.Update(ctx, reverseTransfer)
	if updateErr != nil {
		s.logger.Error(ctx, "error update successful transfer", logger.Error(updateErr))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrUpdateTransfer)
	}

	platformFeeReversalRequest := &platformFee.PlatformReversalFeeRequest{
		ReferenceID:         transfer.UUID.String(),
		ReversalReferenceID: reverseTransfer.UUID.String(),
		Remarks:             request.Remarks,
	}
	err = s.platformFeeSvc.ReverseMerchantFee(ctx, *platformFeeReversalRequest)
	if err != nil {
		s.logger.Error(ctx, "error when trigger reverse merchant platform fee", logger.Error(err))
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrReverseMerchantPlatformFee)
	}

	return transfer, nil
}

func (s *TransferService) processLedgerReverseTransfer(ctx context.Context, request *transfer.ReverseTransferRequest, reverseTransfer *transfer.Transfer) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/transfer/processLedgerReverseTransfer")
	defer segment.End()

	merchantIds := []uuid.UUID{
		util.ParseUUID(request.MerchantID),
		util.ParseUUID(request.ParentMerchantID),
	}
	usecase := constant.ReferencePlatform
	if request.Usecase != "" {
		usecase = request.Usecase
	}
	accountMap, err := s.accountSvc.GetMerchantAccounts(ctx, merchantIds, usecase)
	if err != nil {
		s.logger.Error(ctx, "error get merchant accounts", logger.Error(err))
		return errPkg.New(response.HttpErrInternal, constant.ErrGetAccounts)
	}

	sourceAccount := accountMap[util.ParseUUID(request.ParentMerchantID)]
	if sourceAccount == nil {
		return constant.ErrSenderAccountNotFound
	}
	recipientAccount := accountMap[util.ParseUUID(request.MerchantID)]
	if recipientAccount == nil {
		return constant.ErrRecipientAccountNotFound
	}

	ledgerRequest := &ledger_model.CreateNewLedgerEntryRequest{
		ReferenceID:          reverseTransfer.UUID.String(),
		Usecase:              constant.ReferencePlatform,
		Channel:              "",
		TransactionType:      constant.TypeTransfer,
		Remarks:              request.Remarks,
		Amount:               reverseTransfer.Amount,
		TransactionTimestamp: reverseTransfer.CreatedAt,
		Currency:             constant.CurrencyIDR,
		TransferType:         constant.TransferTypeP2P,
		MoneyFlowType:        constant.MoneyFlowDirect,
		RecipientID:          util.ParseUUID(request.MerchantID),
		RecipientAccountID:   recipientAccount.UUID,
		SenderID:             util.ParseUUID(request.ParentMerchantID),
		SenderAccountID:      sourceAccount.UUID,
		P2PConfig: ledger_model.P2PConfig{
			BypassBalanceCheck: true,
		},
	}

	err = s.ledgerSvc.RecordTransaction(ctx, request.MerchantID, ledgerRequest)
	if err != nil {
		s.logger.Error(ctx, "error when record reverse transfer transaction into ledger", logger.Error(err), logger.Any("request", ledgerRequest))
		return err
	}

	return nil
}
