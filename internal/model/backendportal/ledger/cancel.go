package ledger_model

import (
	"database/sql"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"
)

func CancelTransactions(trxList []*orchestratorModel.AccountTransaction) []*orchestratorModel.AccountTransaction {
	for _, trx := range trxList {
		trx.Status = constant.StatusFailed
		trx.UpdatedAt = time.Now().UTC()
		trx.ReasonType = sql.NullString{String: constant.ReasonTypeCancelMerchantPayment, Valid: true}
		trx.ReasonDescription = sql.NullString{String: constant.ReasonDescCancelMerchantPayment, Valid: true}
		trx.SettlementStatus = sql.NullString{String: constant.SettlementStatusCancelled, Valid: true}
	}
	return trxList
}

func CreateNewCancelTransaction(request *CreateNewLedgerEntryRequest) *orchestratorModel.AccountTransaction {
	trx, _ := orchestratorModel.NewAccountTransaction(&orchestratorModel.CreateNewTransactionRequest{
		ReferenceID:          request.ReferenceID,
		MerchantID:           request.SenderID,
		AccountID:            request.RecipientAccountID,
		MerchantReferenceID:  request.MerchantReferenceID,
		Currency:             constant.CurrencyIDR,
		Credit:               request.Amount,
		Debit:                0,
		Type:                 constant.TypeVoid,
		Channel:              constant.ChannelMerchantPayment,
		Status:               constant.StatusSuccess,
		ReasonType:           "",
		ReasonDescription:    "",
		Reference:            request.Usecase,
		Remarks:              request.Remarks,
		TransactionTimestamp: request.TransactionTimestamp,
		AdditionalInfo:       request.RecipientAdditionalInfo,
	})

	return trx
}
