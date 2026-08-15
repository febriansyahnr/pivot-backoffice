package orchestrator_model

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/common"
)

type ExportTransactionHistoryResponse struct {
	Url string `json:"url"`
}

type TransactionHistoryOpenApiResponse struct {
	MerchantReferenceID    string              `json:"referenceId"`
	Date                   time.Time           `json:"date"`
	SettlementAt           *time.Time          `json:"settlementDate"`
	Status                 string              `json:"settlementStatus"`
	BalanceType            string              `json:"balanceType"`
	TrxChannel             string              `json:"channel"`
	TrxType                string              `json:"transactionType"`
	Amount                 commonModel.Amount2 `json:"amount"`
	Fee                    commonModel.Amount2 `json:"fee"`
	CreatedBy              string              `json:"createdBy,omitempty"`
	ID                     string              `json:"transactionId"`
	BeneficiaryAccountName string              `json:"beneficiaryAccountName,omitempty"`
	ApprovedBy             string              `json:"approvedBy,omitempty"`
	ApprovedAt             *time.Time          `json:"approvedAt,omitempty"`
}

func ToTransactionHistoryOpenApiResponse(useCase *AccountTransactionWithUseCase) TransactionHistoryOpenApiResponse {
	result := TransactionHistoryOpenApiResponse{
		ID:                  useCase.UUID.String(),
		TrxType:             TransactionTypeForUser(useCase.Type, useCase.Channel),
		TrxChannel:          FormatChannelName(useCase.Channel),
		Date:                useCase.UpdatedAt,
		Status:              useCase.Status,
		CreatedBy:           useCase.CreatedBy,
		ApprovedBy:          useCase.ApprovedBy,
		ApprovedAt:          useCase.ApprovedAt,
		BalanceType:         useCase.BalanceType,
		MerchantReferenceID: useCase.MerchantReferenceID,
		Amount: commonModel.Amount2{
			Value:    -1 * useCase.Debit,
			Currency: useCase.Currency,
		},
		Fee: commonModel.Amount2{
			Value:    useCase.Fee,
			Currency: useCase.Currency,
		},
	}
	if useCase.BeneficiaryAccountName != "" && useCase.BeneficiaryAccountName != "-" {
		result.BeneficiaryAccountName = useCase.BeneficiaryAccountName
	}
	if !useCase.SettlementAt.Time.IsZero() {
		result.SettlementAt = &useCase.SettlementAt.Time
	}
	if useCase.Credit != 0 {
		result.Amount.Value = useCase.Credit
	}
	return result
}

type TransactionHistoryResponse struct {
	ID                     string     `json:"id"`
	TrxID                  string     `json:"trxId"`
	TrxType                string     `json:"type"`
	TrxChannel             string     `json:"channel"`
	Date                   time.Time  `json:"date"`
	BeneficiaryAccountName string     `json:"beneficiaryAccountName"`
	Amount                 float64    `json:"amount"`
	Status                 string     `json:"status"`
	CreatedBy              string     `json:"createdBy"`
	ApprovedBy             string     `json:"approvedBy"`
	ApprovedAt             *time.Time `json:"approvedAt"`
	BalanceType            string     `json:"balanceType"`
	MerchantReferenceID    string     `json:"merchantReferenceId"`
	SettlementAt           *time.Time `json:"settlementAt"`
	Fee                    float64    `json:"fee"`
	SettlementStatus       string     `json:"settlementStatus"`
	SettlementModel        string     `json:"settlementModel"`
}

func ToTransactionHistoryResponse(useCase *AccountTransactionWithUseCase) TransactionHistoryResponse {
	result := TransactionHistoryResponse{
		ID:                     useCase.UUID.String(),
		TrxID:                  useCase.ReferenceID,
		TrxType:                useCase.Type,
		TrxChannel:             useCase.Channel,
		Date:                   useCase.UpdatedAt,
		BeneficiaryAccountName: useCase.BeneficiaryAccountName,
		Status:                 useCase.Status,
		CreatedBy:              useCase.CreatedBy,
		ApprovedBy:             useCase.ApprovedBy,
		ApprovedAt:             useCase.ApprovedAt,
		BalanceType:            useCase.BalanceType,
	}
	if !useCase.SettlementAt.Time.IsZero() {
		result.SettlementAt = &useCase.SettlementAt.Time
	}
	if useCase.SettlementStatus.String != "" {
		result.SettlementStatus = useCase.SettlementStatus.String
	}
	result.MerchantReferenceID = useCase.MerchantReferenceID
	result.Amount = -1 * useCase.Debit
	result.Fee = useCase.Fee
	if useCase.Credit != 0 {
		result.Amount = useCase.Credit
	}
	if useCase.SettlementModel.String != "" {
		result.SettlementModel = useCase.SettlementModel.String
	}
	return result
}

type TransactionAndFeeObject struct {
	TransactionID string
	FeeID         string
	// Support to update transfer status
	MerchantID    string
	TransferFeeID string
}

type TransactionHistoryDetailResp struct {
	Id                  string      `json:"id" db:"id"`
	CreatedAt           time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time   `json:"updatedAt" db:"updated_at"`
	Type                string      `json:"type" db:"type"`
	Channel             string      `json:"channel" db:"channel"`
	Remarks             string      `json:"remarks" db:"remarks"`
	Amount              float64     `json:"amount" db:"amount"`
	Status              string      `json:"status" db:"status"`
	ReferenceId         string      `json:"-" db:"reference_id"`
	FailedReason        string      `json:"failedReason" db:"reason_description"`
	LinkedTransactionId string      `json:"-" db:"linked_transaction_id"`
	Details             interface{} `json:"details" db:"-"` // Nullable

	MerchantID string `json:"-" db:"merchant_id"`
}

type TransactionDisbursementResp struct {
	SenderName     string              `json:"senderName" db:"sender_name"`
	CreatedFrom    string              `json:"createdFrom" db:"created_from"`
	CreatedBy      string              `json:"createdBy" db:"created_by"`
	ReferenceId    string              `json:"referenceId" db:"reference_id"`
	BankReference  constant.NullString `json:"bankReference" db:"bank_reference_no"`
	BankName       string              `json:"bankName" db:"beneficiary_bank_name"`
	AccountNumber  string              `json:"accountNumber" db:"beneficiary_account_no"`
	Beneficiary    string              `json:"beneficiary" db:"beneficiary_account_name"`
	Fee            float64             `json:"fee" db:"fee"`
	Total          float64             `json:"total" db:"total_amount"`
	ApprovalStatus string              `json:"approvalStatus" db:"status"`
	ApprovalType   constant.NullString `json:"approvalType" db:"reason_type"`
	ApprovalDesc   constant.NullString `json:"approvalDesc" db:"reason_description"`
	ApprovalDate   constant.NullTime   `json:"approvalDate" db:"approved_at"`
	ApprovalBy     constant.NullString `json:"approvalBy" db:"approved_by"`
	BulkId         constant.NullString `json:"bulkId" db:"bulk_id"`
	Currency       string              `json:"currency" db:"currency"`
}

type TransactionWithdrawalResp struct {
	CreatedBy     string              `json:"createdBy" db:"created_by"`
	BankReference constant.NullString `json:"bankReference" db:"bank_reference_no"`
	BankName      string              `json:"bankName" db:"beneficiary_bank_name"`
	AccountNumber string              `json:"accountNumber" db:"beneficiary_account_no"`
	Beneficiary   string              `json:"beneficiary" db:"beneficiary_account_name"`
}

type TransactionFeeResp struct {
	LinkedID string `json:"linkedId" db:"linked_id"`
}

type TransactionPaymentResp struct {
	PaymentMethodType        string `json:"paymentMethodType" db:"type"`
	PaymentMethodCategory    string `json:"paymentMethodCategory" db:"category"`
	PaymentMethodName        string `json:"paymentMethodName" db:"name"`
	PaymentMethodDescription string `json:"paymentMethodDescription" db:"description"`
}

type TransactionTransferResp struct {
	MerchantID  string  `json:"merchantId" db:"merchant_id"`
	RecipientID string  `json:"recipientId" db:"recipient_id"`
	Currency    string  `json:"currency" db:"currency"`
	Amount      float64 `json:"amount" db:"amount"`

	Type string `json:"type"`
}

type GetMerchantBalanceResponse struct {
	AvailableBalance commonModel.Amount `json:"availableBalance"`
	PendingBalance   commonModel.Amount `json:"pendingBalance"`
	TotalBalance     commonModel.Amount `json:"-"`
}
