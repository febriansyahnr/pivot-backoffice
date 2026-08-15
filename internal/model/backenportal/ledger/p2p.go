package ledger_model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
)

func CreateP2PTransactions(ctx context.Context, request *CreateNewLedgerEntryRequest) ([]*orchestrator_model.AccountTransaction, error) {

	trxList := []*orchestrator_model.AccountTransaction{}
	senderDebitTrx, err := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
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
		Status:                 constant.StatusSuccess,
		Remarks:                request.Remarks,
		AdditionalInfo:         request.SenderAdditionalInfo,
		ProcessorReference:     request.ProcessorReference,
		ProcessorReferenceID:   request.ProcessorReferenceID,
		ProcessorTransactionID: request.ProcessorTransactionID,
		SettlementStatus:       constant.StatusSuccess,
		SettlementAt:           time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}
	trxList = append(trxList, senderDebitTrx)

	if request.Fee.Amount > 0 {
		senderAccountId := request.SenderAccountID
		if request.TransactionType == constant.WalletTrxTopUpType {
			senderAccountId = request.RecipientAccountID
		}
		senderFeeDebitTrx, _ := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
			ReferenceID:          request.ReferenceID,
			Reference:            request.Usecase,
			MerchantID:           request.SenderID,
			AccountID:            senderAccountId,
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
			SettlementStatus:     constant.StatusSuccess,
			SettlementAt:         time.Now().UTC(),
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
				SettlementStatus:     constant.StatusSuccess,
				SettlementAt:         time.Now().UTC(),
			})
			trxList = append(trxList, recipientFeeCreditTrx)
		}
	}

	if request.MoneyFlowType == constant.MoneyFlowIndirect &&
		request.ParentAccountID != uuid.Nil &&
		request.ParentAccountID != request.SenderAccountID &&
		request.ParentAccountID != request.RecipientAccountID {

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
			Status:               constant.StatusSuccess,
			Remarks:              request.Remarks,
			AdditionalInfo:       request.SenderAdditionalInfo,
			SettlementStatus:     constant.StatusSuccess,
			SettlementAt:         time.Now().UTC(),
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
			Status:               constant.StatusSuccess,
			Remarks:              request.Remarks,
			AdditionalInfo:       request.SenderAdditionalInfo,
			SettlementStatus:     constant.StatusSuccess,
			SettlementAt:         time.Now().UTC(),
		})

		trxList = append(trxList, glCreditTrx, glDebitTrx)
	}

	recipientCreditTrx, err := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
		ReferenceID:          request.ReferenceID,
		Reference:            request.Usecase,
		MerchantID:           request.RecipientID,
		AccountID:            request.RecipientAccountID,
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
		SettlementStatus:     constant.StatusSuccess,
		SettlementAt:         time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}

	trxList = append(trxList, recipientCreditTrx)
	return trxList, nil
}

func (request *CreateNewLedgerEntryRequest) ValidateP2PRequest() error {
	if request.RecipientAccountID == uuid.Nil || request.RecipientID == uuid.Nil {
		return constant.ErrMissingRecipientAccountID
	}

	if request.RecipientAccountID == request.SenderAccountID {
		return constant.ErrSenderSameWithRecipient
	}

	if request.SenderAccountID == uuid.Nil || request.SenderID == uuid.Nil {
		return constant.ErrMissingSenderAccountID
	}

	if request.MoneyFlowType == constant.MoneyFlowIndirect &&
		request.SenderAccountID != request.ParentAccountID &&
		request.RecipientAccountID != request.ParentAccountID &&
		(request.ParentAccountID == uuid.Nil || request.ParentID == uuid.Nil) {
		return constant.ErrParentAccountNotFound
	}

	return nil
}
