package ledger_model

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
)

func CreatePayOutTransactions(ctx context.Context, request *CreateNewLedgerEntryRequest) ([]*orchestrator_model.AccountTransaction, error) {
	trxList := []*orchestrator_model.AccountTransaction{}

	initiatorDebitTrx, err := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
		ReferenceID:            request.ReferenceID,
		Reference:              request.Usecase,
		MerchantID:             request.SenderID,
		AccountID:              request.SenderAccountID,
		MerchantReferenceID:    request.MerchantReferenceID,
		Currency:               request.Currency,
		Debit:                  request.Amount,
		Credit:                 0,
		TransactionTimestamp:   request.TransactionTimestamp,
		Channel:                request.Channel,
		Type:                   request.TransactionType,
		Status:                 constant.StatusPending,
		Remarks:                request.Remarks,
		AdditionalInfo:         request.SenderAdditionalInfo,
		ProcessorReference:     request.ProcessorReference,
		ProcessorReferenceID:   request.ProcessorReferenceID,
		ProcessorTransactionID: request.ProcessorTransactionID,
	})
	if err != nil {
		return nil, err
	}
	trxList = append(trxList, initiatorDebitTrx)

	if request.Fee.Amount > 0 {
		senderFeeDebitTrx, _ := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
			ReferenceID:          request.ReferenceID,
			Reference:            request.Usecase,
			MerchantID:           request.SenderID,
			AccountID:            request.SenderAccountID,
			MerchantReferenceID:  request.MerchantReferenceID,
			Currency:             request.Currency,
			Debit:                request.Fee.Amount,
			Credit:               0,
			TransactionTimestamp: request.TransactionTimestamp,
			Channel:              request.Fee.Channel,
			Type:                 constant.TypeFee,
			Status:               constant.StatusPending,
			Remarks:              request.Remarks,
			AdditionalInfo:       request.Fee.AdditionalInfo,
		})
		trxList = append(trxList, senderFeeDebitTrx)

		if request.Fee.RecipientAccountID != uuid.Nil {
			recipientFeeCreditTrx, _ := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
				ReferenceID:          request.ReferenceID,
				Reference:            request.Usecase,
				MerchantID:           request.Fee.RecipientID,
				AccountID:            request.Fee.RecipientAccountID,
				MerchantReferenceID:  request.MerchantReferenceID,
				Currency:             request.Currency,
				Debit:                0,
				Credit:               request.Fee.Amount,
				TransactionTimestamp: request.TransactionTimestamp,
				Channel:              request.Fee.Channel,
				Type:                 constant.TypeFee,
				Status:               constant.StatusPending,
				Remarks:              request.Remarks,
				AdditionalInfo:       request.Fee.AdditionalInfo,
			})
			trxList = append(trxList, recipientFeeCreditTrx)
		}
	}

	if request.MoneyFlowType == constant.MoneyFlowDirect {
		return trxList, nil
	}
	if request.MoneyFlowType == constant.MoneyFlowIndirect && request.ParentAccountID != request.SenderAccountID {

		glCreditTrx, _ := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
			ReferenceID:          request.ReferenceID,
			Reference:            request.Usecase,
			MerchantID:           request.ParentID,
			AccountID:            request.ParentAccountID,
			MerchantReferenceID:  request.MerchantReferenceID,
			Currency:             request.Currency,
			Credit:               request.Amount,
			Debit:                0,
			TransactionTimestamp: request.TransactionTimestamp,
			Channel:              request.Channel,
			Type:                 request.TransactionType,
			Status:               constant.StatusPending,
			Remarks:              request.Remarks,
			AdditionalInfo:       request.SenderAdditionalInfo,
		})

		glDebitTrx, _ := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
			ReferenceID:          request.ReferenceID,
			Reference:            request.Usecase,
			MerchantID:           request.ParentID,
			AccountID:            request.ParentAccountID,
			MerchantReferenceID:  request.MerchantReferenceID,
			Currency:             request.Currency,
			Debit:                request.Amount,
			Credit:               0,
			TransactionTimestamp: request.TransactionTimestamp,
			Channel:              request.Channel,
			Type:                 request.TransactionType,
			Status:               constant.StatusPending,
			Remarks:              request.Remarks,
			AdditionalInfo:       request.SenderAdditionalInfo,
		})
		trxList = append(trxList, glCreditTrx, glDebitTrx)
	}
	return trxList, nil
}

func (request *CreateNewLedgerEntryRequest) ValidatePayOutRequest() error {
	if request.SenderAccountID == uuid.Nil || request.SenderID == uuid.Nil {
		return constant.ErrMissingSenderAccountID
	}

	if request.MoneyFlowType == constant.MoneyFlowIndirect &&
		request.SenderAccountID != request.ParentAccountID &&
		(request.ParentAccountID == uuid.Nil || request.ParentID == uuid.Nil) {
		return constant.ErrMissingParentAccountID
	}
	return nil
}
