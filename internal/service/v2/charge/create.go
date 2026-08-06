package chargeMoneyFlowService

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

func (s *ChargeMoneyFlowService) CreateTransactions(ctx context.Context, request *ledger_model.CreateNewLedgerEntryRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/charge/CreateTransactions")
	defer segment.End()

	err := request.ValidateChargeRequest()
	if err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	if !request.ChargeConfig.BypassBalanceCheck {
		balance, err := s.ledgerSvc.GetLedgerBalance(ctx, request.SenderAccountID)
		if err != nil {
			return pkgErrors.New(response.HttpErrInternal, constant.ErrFailedValidateCharge)
		}
		if balance.Balance < request.Amount {
			return pkgErrors.New(response.HttpErrRequest, constant.ErrInsufficientBalance)
		}
	}

	trxList, err := ledger_model.CreateChargeTransactions(ctx, request)
	if err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	pubPendSettlementIfFeeMerchantPaymentReq := s.markForMerchantSettlement(request, trxList)

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
		}
	}

	if err = s.repo.BulkInsert(ctx, trxList); err != nil {
		return pkgErrors.New(response.HttpErrInternal, constant.ErrCreateLedgerEntry)
	}

	_ = pubPendSettlementIfFeeMerchantPaymentReq(ctx)

	return nil
}

func (s *ChargeMoneyFlowService) markForMerchantSettlement(request *ledger_model.CreateNewLedgerEntryRequest, transactions []*orchestratorModel.AccountTransaction) (process func(context.Context) error) {

	process = func(context.Context) error { return nil }

	if len(transactions) == 0 {
		return

	} else if request.TransactionType != constant.TypeFee || request.SenderAdditionalInfo == nil {
		return
	}

	additionalInfo := map[string]any{}

	switch val := request.SenderAdditionalInfo.(type) {
	default:
		raw, _ := json.Marshal(val)
		_ = json.Unmarshal(raw, &additionalInfo)

	case map[string]any:
		additionalInfo = val
	}

	if t, _ := additionalInfo["referenceType"].(string); t != constant.WalletTrxMerchantPaymentType {
		return
	}

	// For now only for fees earned from MERCHANT_PAYMENT type.
	// Config:
	// 	  Type: "T+1"
	// 	  DayType: "ANYDAY"
	//    PendingTime: "24 Hours"

	settlementRequests := make([]settlementModel.ProcessSettlementRequest, len(transactions))

	for i := range transactions {

		metadata := map[string]any{}

		_ = json.Unmarshal(transactions[i].AdditionalInfo.JSONText, &metadata)

		metadata["settlementDetail"] = orchestratorModel.MetadataPaymentSettlementDetail{Type: constant.SettlementTypeTimePlus01}

		transactions[i].Status = constant.StatusPending
		transactions[i].AdditionalInfo.Valid = true
		transactions[i].AdditionalInfo.JSONText, _ = json.Marshal(metadata)
		transactions[i].SettlementStatus = sql.NullString{String: constant.StatusPending, Valid: true}

		settlementRequests[i] = settlementModel.ProcessSettlementRequest{
			MerchantID:    transactions[i].MerchantID.String(),
			Type:          constant.SettlementFeeOnly,
			TransactionID: transactions[i].UUID.String(),
		}
	}
	return func(ctx context.Context) error {
		for _, request := range settlementRequests {
			if err := s.queues.PublishForSettlementProcess(ctx, messagingQueueModel.PublishSettlementProcessPayload{
				SettlementType: constant.SettlementTypeTimePlus01,
				Day:            1,
				Payload:        request,
				MessageTTL:     time.Hour * 24,
				ModifyMessage:  nil,
			}); err != nil {
				s.logger.Error(ctx, "Failed while publish pending settlement process (charge)", logger.Error(err))
			}
		}
		return nil
	}
}
