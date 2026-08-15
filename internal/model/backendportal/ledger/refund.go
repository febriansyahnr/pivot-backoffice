package ledger_model

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"
)

func NewRefundTransactions(request *CreateNewLedgerEntryRequest) ([]*orchestratorModel.AccountTransaction, error) {
	var (
		trxList = []*orchestratorModel.AccountTransaction{}
	)

	if request.RefundConfig.RefundToSenderFirst {
		refundToDebitTrx, err := orchestratorModel.NewAccountTransaction(&orchestratorModel.CreateNewTransactionRequest{
			ReferenceID:          request.ReferenceID,
			MerchantID:           request.SenderID,
			AccountID:            request.SenderAccountID,
			MerchantReferenceID:  request.MerchantReferenceID,
			Currency:             constant.CurrencyIDR,
			Credit:               request.Amount,
			Debit:                0,
			Type:                 request.TransactionType,
			Channel:              request.Channel,
			Status:               constant.StatusSuccess,
			ReasonType:           "",
			ReasonDescription:    "",
			Reference:            request.Usecase,
			Remarks:              request.Remarks,
			TransactionTimestamp: request.TransactionTimestamp,
			AdditionalInfo:       request.SenderAdditionalInfo,
		})
		if err != nil {
			return trxList, err
		}
		trxList = append(trxList, refundToDebitTrx)
	}
	debitTrx, err := orchestratorModel.NewAccountTransaction(&orchestratorModel.CreateNewTransactionRequest{
		ReferenceID:          request.ReferenceID,
		MerchantID:           request.SenderID,
		AccountID:            request.SenderAccountID,
		MerchantReferenceID:  request.MerchantReferenceID,
		Currency:             constant.CurrencyIDR,
		Credit:               0,
		Debit:                request.Amount,
		Type:                 request.TransactionType,
		Channel:              request.Channel,
		Status:               constant.StatusSuccess,
		ReasonType:           "",
		ReasonDescription:    "",
		Reference:            request.Usecase,
		Remarks:              request.Remarks,
		TransactionTimestamp: request.TransactionTimestamp,
		AdditionalInfo:       request.SenderAdditionalInfo,
	})
	if err != nil {
		return trxList, err
	}
	trxList = append(trxList, debitTrx)

	creditTrx, err := orchestratorModel.NewAccountTransaction(&orchestratorModel.CreateNewTransactionRequest{
		ReferenceID:          request.ReferenceID,
		MerchantID:           request.RecipientID,
		AccountID:            request.RecipientAccountID,
		MerchantReferenceID:  request.MerchantReferenceID,
		Currency:             constant.CurrencyIDR,
		Credit:               request.Amount,
		Debit:                0,
		Type:                 request.TransactionType,
		Channel:              request.Channel,
		Status:               constant.StatusSuccess,
		ReasonType:           "",
		ReasonDescription:    "",
		Reference:            request.Usecase,
		Remarks:              request.Remarks,
		TransactionTimestamp: request.TransactionTimestamp,
		AdditionalInfo:       request.SenderAdditionalInfo,
	})
	if err != nil {
		return trxList, err
	}
	trxList = append(trxList, creditTrx)

	if request.Fee.Amount > 0 {
		feeTrx, err := orchestratorModel.NewAccountTransaction(&orchestratorModel.CreateNewTransactionRequest{
			ReferenceID:          request.ReferenceID,
			MerchantID:           request.Fee.RecipientID,
			AccountID:            request.Fee.RecipientAccountID,
			MerchantReferenceID:  request.MerchantReferenceID,
			Currency:             constant.CurrencyIDR,
			Credit:               request.Fee.Amount,
			Debit:                0,
			Type:                 request.Fee.TransactionType, // in wallet case, could be FEE_REVERSAL
			Channel:              request.Fee.Channel,
			Status:               constant.StatusSuccess,
			ReasonType:           "",
			ReasonDescription:    "",
			Reference:            request.Usecase,
			Remarks:              request.Remarks,
			TransactionTimestamp: request.TransactionTimestamp,
			AdditionalInfo:       request.Fee.AdditionalInfo,
		})
		if err != nil {
			return trxList, err
		}
		trxList = append(trxList, feeTrx)
	}

	return trxList, nil
}
