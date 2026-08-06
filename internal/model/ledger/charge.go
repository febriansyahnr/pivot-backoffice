package ledger_model

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
)

func CreateChargeTransactions(ctx context.Context, request *CreateNewLedgerEntryRequest) ([]*orchestrator_model.AccountTransaction, error) {
	// Validate the request first
	if err := request.ValidateChargeRequest(); err != nil {
		return nil, err
	}

	// Validate fee AdditionalInfo if present
	if request.Fee.Amount > 0 {
		// Check if AdditionalInfo is a valid type (map or nil)
		switch request.Fee.AdditionalInfo.(type) {
		case nil, map[string]interface{}:
			// These types are valid
		default:
			// Any other type is invalid
			return nil, errors.New("invalid fee additional info type")
		}
	}

	trxStatus := constant.StatusSuccess
	if !request.ChargeConfig.IsDirectlyDeducted {
		trxStatus = constant.StatusPending
	}

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
		Status:                 trxStatus,
		Remarks:                request.Remarks,
		AdditionalInfo:         request.SenderAdditionalInfo,
		ProcessorReference:     request.ProcessorReference,
		ProcessorReferenceID:   request.ProcessorReferenceID,
		ProcessorTransactionID: request.ProcessorTransactionID,
		SettlementStatus:       trxStatus,
		SettlementAt:           time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}
	trxList = append(trxList, senderDebitTrx)

	if request.Fee.Amount > 0 {
		senderFeeDebitTrx, err := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
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
			Status:               trxStatus,
			Remarks:              request.Remarks,
			AdditionalInfo:       request.Fee.AdditionalInfo,
			SettlementStatus:     trxStatus,
			SettlementAt:         time.Now().UTC(),
		})
		if err != nil {
			return nil, err
		}
		trxList = append(trxList, senderFeeDebitTrx)

		if request.Fee.RecipientAccountID != uuid.Nil {
			recipientFeeCreditTrx, err := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
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
				Status:               trxStatus,
				Remarks:              request.Remarks,
				AdditionalInfo:       request.Fee.AdditionalInfo,
				SettlementStatus:     trxStatus,
				SettlementAt:         time.Now().UTC(),
			})
			if err != nil {
				return nil, err
			}
			trxList = append(trxList, recipientFeeCreditTrx)
		}
	}

	if request.RecipientAccountID != uuid.Nil {
		recipientCreditTrx, err := orchestrator_model.NewAccountTransaction(&orchestrator_model.CreateNewTransactionRequest{
			ReferenceID:          request.ReferenceID,
			Reference:            request.Usecase,
			MerchantID:           request.RecipientID,
			AccountID:            request.RecipientAccountID,
			MerchantReferenceID:  request.MerchantReferenceID,
			Currency:             request.Currency,
			Debit:                0,
			Credit:               request.Amount,
			TransactionTimestamp: request.TransactionTimestamp,
			Channel:              request.Channel,
			Type:                 request.TransactionType,
			Status:               trxStatus,
			Remarks:              request.Remarks,
			AdditionalInfo:       request.RecipientAdditionalInfo,
			SettlementStatus:     trxStatus,
			SettlementAt:         time.Now().UTC(),
		})
		if err != nil {
			return nil, err
		}
		trxList = append(trxList, recipientCreditTrx)
	}

	return trxList, nil
}

func (request *CreateNewLedgerEntryRequest) ValidateChargeRequest() error {
	if request.SenderAccountID == uuid.Nil || request.SenderID == uuid.Nil {
		return constant.ErrMissingSenderAccountID
	}

	if request.Amount <= 0 {
		return constant.ErrInvalidAmount
	}

	if request.Currency == "" {
		return errors.New("invalid currency")
	}

	if request.TransactionTimestamp.IsZero() {
		return constant.ErrInvalidTransactionTimestamp
	}

	return nil
}
