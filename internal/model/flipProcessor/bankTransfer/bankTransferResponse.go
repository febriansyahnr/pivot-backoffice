package flipProcessorModel

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	"github.com/paper-indonesia/pdk/go/snap"
)

var loc, _ = time.LoadLocation(constant.TimeLoc)

type BankTransferResponse struct {
	ID             int            `json:"id,omitempty"`
	UserID         int            `json:"user_id,omitempty"`
	Amount         int            `json:"amount,omitempty"`
	Status         string         `json:"status,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	Timestamp      string         `json:"timestamp,omitempty"`
	BankCode       string         `json:"bank_code,omitempty"`
	AccountNo      string         `json:"account_number,omitempty"`
	Remark         string         `json:"remark,omitempty"`
	Receipt        string         `json:"receipt,omitempty"`
	SenderBank     string         `json:"sender_bank,omitempty"`
	CreatedFrom    string         `json:"created_from,omitempty"`
	Fee            int            `json:"fee,omitempty"`
	Sender         map[string]any `json:"sender,omitempty"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}

func (b *BankTransferResponse) ToBankTransferResponse() *routingProcessorModel.BankTransferResponseData {
	var metadata map[string]any
	responseB, _ := json.Marshal(b)
	json.Unmarshal(responseB, &metadata)

	switch b.Status {
	case constant.FlipBankTransferStatusDone:
		b.Status = constant.SnapCoreBankTransferStatusSuccess
	case constant.FlipBankTransferStatusCancelled:
		b.Status = constant.SnapCoreBankTransferStatusFailed
	case constant.FlipBankTransferStatusPending:
		b.Status = constant.SnapCoreBankTransferStatusPending
	}

	code := snap.SNAP_SUCCESS
	switch b.Reason {
	case constant.FlipReasonInactiveAccount,
		constant.FlipReasonNotRegisteredAccount:
		code = snap.SNAP_INACTIVE_ACCOUNT
	case constant.FlipReasonBeneficiaryAccountNotVerified,
		constant.FlipReasonInvalidAccount:
		code = snap.SNAP_INVALID_ACCOUNT
	case constant.FlipReasonInsufficientBalance:
		code = snap.SNAP_INSUFFICIENT_FUND
	case constant.FlipReasonExceedAmountLimit,
		constant.FlipReasonInvalidAmount:
		code = snap.SNAP_INVALID_AMOUNT
	case constant.FlipReasonDormantAccount:
		code = snap.SNAP_DORMANT_ACCOUNT
	}

	responseCode, responseMessage := snap.GenerateResponseCode(code, snap.SNAP_SERVICE_INTERBANK_TRANSFER)

	timestamp, err := time.ParseInLocation(constant.DatetimeFormat, b.Timestamp, loc)
	if err != nil {
		timestamp = time.Now().UTC()
	}

	return &routingProcessorModel.BankTransferResponseData{
		ResponseCode:    responseCode,
		ResponseMessage: responseMessage,
		UUID:            strconv.Itoa(b.ID),
		BankReferenceNo: b.IdempotencyKey,
		Amount: commonModel.Amount{
			Value:    strconv.Itoa(b.Amount),
			Currency: "IDR",
		},
		BeneficiaryAccountNo: b.AccountNo,
		BeneficiaryBankCode:  b.BankCode,
		Status:               b.Status,
		TransferType:         b.CreatedFrom,
		ExternalID:           b.IdempotencyKey,
		Reason:               b.Reason,
		Metadata:             metadata,
		ProcessorReference:   constant.FlipPGProcessor,
		TransactionDate:      timestamp.UTC(),
	}
}
