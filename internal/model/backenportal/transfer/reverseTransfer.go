package transfer

import (
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
)

type ReverseTransferRequest struct {
	Usecase           string
	ReferenceID       string
	MerchantID        string
	ParentMerchantID  string
	Remarks           string
	ReasonDescription string
}

func (t *Transfer) ReverseTransfer(request *ReverseTransferRequest) *Transfer {
	now := time.Now().UTC()
	return &Transfer{
		UUID:              uuid.New(),
		MerchantID:        t.RecipientID,
		RecipientID:       t.MerchantID,
		ReferenceID:       request.ReferenceID,
		TransferType:      t.TransferType,
		Currency:          t.Currency,
		Amount:            t.Amount,
		Status:            constant.TransferStatusPending,
		Remarks:           request.Remarks,
		ReasonDescription: request.ReasonDescription,
		CreatedAt:         now,
		UpdatedAt:         now,
		DeletedAt:         nil,
	}
}
