package ledger_model

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"
)

type CreateNewLedgerEntryRequest struct {
	ReferenceID             string    `json:"referenceId" validate:"required"`
	MerchantReferenceID     *string   `json:"merchantReferenceId"`
	Usecase                 string    `json:"usecase" validate:"required"`
	TransactionType         string    `json:"transactionType" validate:"required"`
	Channel                 string    `json:"channel"`
	Remarks                 string    `json:"remarks"`
	TransactionTimestamp    time.Time `json:"transactionTimestamp" validate:"required"`
	Amount                  float64   `json:"amount" validate:"required"`
	Currency                string    `json:"currency" validate:"required"`
	TransferType            string    `json:"transferType" validate:"required"` // PAY_IN, PAY_OUT, P2P
	RecipientID             uuid.UUID
	RecipientAccountID      uuid.UUID `json:"recipientAccountId" validate:"uuid"`
	SenderID                uuid.UUID
	SenderAccountID         uuid.UUID `json:"senderAccountId" validate:"uuid"`
	ParentID                uuid.UUID
	ParentAccountID         uuid.UUID    `json:"parentAccountId" validate:"uuid"`
	MoneyFlowType           string       `json:"moneyFlowType"`
	Fee                     FeeRequest   `json:"fee"`
	SenderAdditionalInfo    interface{}  `json:"senderAdditionalInfo"`
	RecipientAdditionalInfo interface{}  `json:"recipientAdditionalInfo"`
	ChargeConfig            ChargeConfig `json:"chargeConfig"`
	P2PConfig               P2PConfig    `json:"p2pConfig"`
	RefundConfig            RefundConfig `json:"refundConfig"`
	ProcessorReference      string       `json:"processorReference"`
	ProcessorReferenceID    string       `json:"processorReferenceId"`
	ProcessorTransactionID  string       `json:"processorTransactionId"`
	SettlementStatus        string       `json:"settlementStatus"`
	SettlementAt            time.Time    `json:"settlementAt"`
}

type FeeRequest struct {
	Amount             float64 `json:"amount"`
	TransactionType    string  `json:"transactionType"`
	Channel            string  `json:"channel"`
	RecipientID        uuid.UUID
	RecipientAccountID uuid.UUID   `json:"recipientAccountId"`
	AdditionalInfo     interface{} `json:"additionalInfo"`
}

type ChargeConfig struct {
	BypassBalanceCheck bool `json:"bypassBalanceCheck"`
	IsDirectlyDeducted bool `json:"isDirectlyDeducted"`
}

type P2PConfig struct {
	BypassBalanceCheck bool `json:"bypassBalanceCheck"`
}

type RefundConfig struct {
	RefundToSenderFirst bool `json:"refundToSenderFirst"`
}

type UpdateLedgerEntryRequest struct {
	ReferenceID            uuid.UUID                     `json:"referenceId"`
	Usecase                string                        `json:"usecase" validate:"required"`
	Status                 string                        `json:"status" validate:"required"`
	ReasonDescription      string                        `json:"reasonDescription"`
	ReasonType             string                        `json:"reasonType"`
	AdditionalInfo         interface{}                   `json:"additionalInfo"`
	ProcessorReference     string                        `json:"processorReference"`
	ProcessorReferenceID   string                        `json:"processorReferenceId"`
	ProcessorTransactionID string                        `json:"processorTransactionId"`
	SettlementStatus       string                        `json:"settlementStatus"`
	SettlementAt           time.Time                     `json:"settlementAt"`
	SettlementModel        string                        `json:"settlementModel"`
	Conditional            *UpdateLedgerEntryConditional `json:"conditional"`
}

type BulkUpdateLedgerEntryRequest struct {
	ReferenceID uuid.UUID
	Requests    []*UpdateLedgerEntryRequest `json:"requests"`
}

type UpdateLedgerEntryConditional struct {
	CurrentStatus string `json:"currentStatus"`
	Type          string `json:"type"`
}

func (r *CreateNewLedgerEntryRequest) Validate() error {
	if err := validateTransferType(r.TransferType); err != nil {
		return err
	}

	if r.Fee.Amount < 0 {
		return constant.ErrNegativeFee
	}

	return validateUseCase(r.Usecase)
}

func validateTransferType(transferType string) error {
	upp := strings.ToUpper(transferType)
	switch upp {
	case constant.TransferTypeP2P, constant.TransferTypePayIn, constant.TransferTypePayOut, constant.TransferTypeCharge, constant.TransferTypeCancel, constant.TransferTypeRefund:
		return nil
	default:
		return constant.ErrInvalidTransferType
	}
}

func validateUseCase(useCase string) error {
	return orchestrator_model.ValidateUseCase(useCase)
}

func (r *UpdateLedgerEntryRequest) Validate() error {
	if r.ReferenceID == uuid.Nil {
		return constant.ErrInvalidReferenceID
	}
	return validateUseCase(r.Usecase)
}
