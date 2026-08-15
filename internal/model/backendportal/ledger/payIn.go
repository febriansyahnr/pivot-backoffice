package ledger_model

import (
	"context"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
)

func CreatePayInTransactions(ctx context.Context, request *CreateNewLedgerEntryRequest) ([]*orchestrator_model.AccountTransaction, error) {
	trxList := []*orchestrator_model.AccountTransaction{}

	if request.MoneyFlowType == constant.MoneyFlowIndirect && request.ParentAccountID != request.RecipientAccountID {
		glCreditTrx, err := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
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
			Status:               constant.StatusSuccess,
			Remarks:              request.Remarks,
			AdditionalInfo:       request.RecipientAdditionalInfo,
		})
		if err != nil {
			return nil, err
		}

		glDebitTrx, _ := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
			ReferenceID:          request.ReferenceID,
			Reference:            request.Usecase,
			MerchantID:           request.ParentID,
			AccountID:            request.ParentAccountID,
			MerchantReferenceID:  request.MerchantReferenceID,
			Currency:             request.Currency,
			Credit:               0,
			Debit:                request.Amount,
			TransactionTimestamp: request.TransactionTimestamp,
			Channel:              request.Channel,
			Type:                 request.TransactionType,
			Status:               constant.StatusSuccess,
			Remarks:              request.Remarks,
			AdditionalInfo:       request.RecipientAdditionalInfo,
		})

		trxList = append(trxList, glCreditTrx, glDebitTrx)
	}

	ledgerCreditTrx, err := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
		ReferenceID:            request.ReferenceID,
		Reference:              request.Usecase,
		MerchantID:             request.RecipientID,
		AccountID:              request.RecipientAccountID,
		MerchantReferenceID:    request.MerchantReferenceID,
		Currency:               request.Currency,
		Debit:                  0,
		Credit:                 request.Amount,
		TransactionTimestamp:   request.TransactionTimestamp,
		Channel:                request.Channel,
		Type:                   request.TransactionType,
		Status:                 constant.StatusSuccess,
		Remarks:                request.Remarks,
		AdditionalInfo:         request.RecipientAdditionalInfo,
		ProcessorReference:     request.ProcessorReference,
		ProcessorReferenceID:   request.ProcessorReferenceID,
		ProcessorTransactionID: request.ProcessorTransactionID,
	})
	if err != nil {
		return nil, err
	}
	trxList = append(trxList, ledgerCreditTrx)

	if request.Fee.Amount > 0 {
		senderFeeDebitTrx, _ := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
			ReferenceID:          request.ReferenceID,
			Reference:            request.Usecase,
			MerchantID:           request.RecipientID,
			AccountID:            request.RecipientAccountID,
			MerchantReferenceID:  request.MerchantReferenceID,
			Currency:             request.Currency,
			Debit:                request.Fee.Amount,
			Credit:               0,
			TransactionTimestamp: request.TransactionTimestamp,
			Channel:              request.Fee.Channel,
			Type:                 constant.TypeFee,
			Status:               constant.StatusSuccess,
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
				Status:               constant.StatusSuccess,
				Remarks:              request.Remarks,
				AdditionalInfo:       request.Fee.AdditionalInfo,
			})
			trxList = append(trxList, recipientFeeCreditTrx)
		}
	}

	return trxList, nil
}

func (request *CreateNewLedgerEntryRequest) ValidatePayInRequest() error {
	if request.Fee.Amount > request.Amount {
		return constant.ErrPayInFeeBiggerThanAmount
	}

	if request.RecipientAccountID == uuid.Nil || request.RecipientID == uuid.Nil {
		return constant.ErrMissingRecipientAccountID
	}

	if request.MoneyFlowType == constant.MoneyFlowIndirect &&
		request.RecipientAccountID != request.ParentAccountID &&
		(request.ParentAccountID == uuid.Nil || request.ParentID == uuid.Nil) {
		return constant.ErrMissingParentAccountID
	}
	return nil
}
