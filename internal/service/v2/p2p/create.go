package p2pMoneyFlowService

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	messagingQueueModel "github.com/paper-indonesia/pivot-backoffice/internal/model/messagingQueue"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *P2PMoneyFlowService) CreateTransactions(ctx context.Context, request *ledger_model.CreateNewLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/p2p/CreateTransactions")
	defer segment.End()

	err := request.ValidateP2PRequest()
	if err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	if !request.P2PConfig.BypassBalanceCheck {
		balance, err := s.ledgerSvc.GetLedgerBalance(ctx, request.SenderAccountID)
		if err != nil {
			return pkgErrors.New(response.HttpErrInternal, constant.ErrFailedValidateP2PTransfer)
		}
		if balance.Balance < (request.Amount + request.Fee.Amount) {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrInsufficientBalance)
		}
	}

	trxList, err := ledger_model.CreateP2PTransactions(ctx, request)
	if err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}
	// Sorting transaction mutations so that when a customer tops up via a merchant, the balance is increased first before the top-up fee is charged.
	if request.TransactionType == constant.WalletTrxTopUpType && len(trxList) == 4 {
		trxList[1], trxList[2], trxList[3] = trxList[3], trxList[1], trxList[2]

		updatedAt := time.Now().UTC().Add(time.Millisecond)
		trxList[2].UpdatedAt, trxList[3].UpdatedAt = updatedAt, updatedAt
	}

	pubPendSettlementIfMerchantPaymentReq := s.markForMerchantSettlement(request, trxList)

	if set, _ := ctx.Value(constant.CtxSetPendingTransaction).(bool); set {
		for i := range trxList {
			trxList[i].Status = constant.StatusPending
		}
	}
	if set, _ := ctx.Value(constant.CtxSetPendingSettlementTransaction).(bool); set {
		for i := range trxList {
			trxList[i].SettlementStatus = sql.NullString{
				Valid:  true,
				String: constant.StatusPending,
			}
			trxList[i].SettlementAt = sql.NullTime{
				Valid: false,
			}
		}
	}
	if err = s.repo.BulkInsert(ctx, trxList); err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrCreateLedgerEntry)
	}

	if err = pubPendSettlementIfMerchantPaymentReq(ctx); err != nil {
		s.logger.Error(ctx, "Failed while publish pending settlement process (p2p)", logger.Error(err))
	}
	return nil
}

func (s *P2PMoneyFlowService) markForMerchantSettlement(request *ledger_model.CreateNewLedgerEntryRequest, transactions []*orchestratorModel.AccountTransaction) (process func(context.Context) error) {

	process = func(context.Context) error { return nil }

	if len(transactions) == 0 {
		return

	} else if request.TransactionType != constant.WalletTrxMerchantPaymentType {
		return
	}

	// For now only MERCHANT_PAYMENT transactions.
	// Config:
	// 	  Type: "T+1"
	// 	  DayType: "ANYDAY"
	//    PendingTime: "24 Hours"

	settlementRequest := settlementModel.ProcessSettlementRequest{
		MerchantID: transactions[0].MerchantID.String(),
		Type:       constant.SettlementTransaction,
	}

	for i := range transactions {
		if transactions[i].AccountID != request.RecipientAccountID {
			continue
		}
		metadata := map[string]any{}

		_ = json.Unmarshal(transactions[i].AdditionalInfo.JSONText, &metadata)

		metadata["settlementDetail"] = orchestratorModel.MetadataPaymentSettlementDetail{Type: constant.SettlementTypeTimePlus01}

		transactions[i].AdditionalInfo.Valid = true
		transactions[i].AdditionalInfo.JSONText, _ = json.Marshal(metadata)
		transactions[i].SettlementStatus = sql.NullString{String: constant.StatusPending, Valid: true}

		switch transactions[i].Type {
		case constant.WalletTrxMerchantPaymentType:
			settlementRequest.TransactionID = transactions[i].UUID.String()

		case constant.TypeFee:
			transactions[i].Status = constant.StatusPending
			settlementRequest.FeeTransactionID = transactions[i].UUID.String()
		}
	}
	return func(ctx context.Context) error {
		return s.queues.PublishForSettlementProcess(ctx, messagingQueueModel.PublishSettlementProcessPayload{
			SettlementType: constant.SettlementTypeTimePlus01,
			Day:            1,
			Payload:        &settlementRequest,
			MessageTTL:     time.Hour * 24,
			ModifyMessage:  nil,
		})
	}
}
