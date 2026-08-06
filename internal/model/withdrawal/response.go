package withdrawal

import (
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/bankAccount"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

	"github.com/jmoiron/sqlx/types"
)

type PreparationResponse struct {
	MerchantId       string                            `json:"merchantId"`
	AccountName      string                            `json:"accountName"`
	AvailableBalance float64                           `json:"availableBalance"`
	BankAccounts     []bankAccount.BankAccountResponse `json:"bankAccounts"`
}

type WithdrawalProcessResponse struct {
	Id          string             `json:"id"`
	Type        string             `json:"type"`
	AccountName string             `json:"accountName"`
	Status      string             `json:"status"`
	Reason      string             `json:"reason,omitempty"`
	Amount      commonModel.Amount `json:"-"`
	CreatedAt   time.Time          `json:"-"`
	UpdatedAt   time.Time          `json:"-"`
}

type WithdrawalHistoryResponse struct {
	Id                     string    `json:"id" db:"id"`
	Date                   time.Time `json:"date" db:"created_at"`
	Type                   string    `json:"type" db:"type"`
	Amount                 float64   `json:"amount" db:"amount"`
	BeneficiaryBankName    string    `json:"beneficiaryBankName" db:"beneficiary_bank_name"`
	BeneficiaryAccountName string    `json:"beneficiaryAccountName" db:"beneficiary_account_name"`
	Status                 string    `json:"status" db:"status"`
	CreatedBy              string    `json:"createdBy" db:"created_by"`
	BalanceType            string    `json:"balanceType" db:"balance_type"`
	// Support Excel Export
	UpdatedAt            time.Time `json:"-" db:"updated_at"`
	BankReference        string    `json:"-" db:"bank_reference"`
	BeneficiaryAccountNo string    `json:"-" db:"beneficiary_account_no"`
}

type WithdrawalDetailResponse struct {
	Id                     string    `json:"id" db:"id"`
	CreatedAt              time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt              time.Time `json:"updatedAt" db:"updated_at"`
	CreatedBy              string    `json:"createdBy" db:"created_by"`
	Type                   string    `json:"type" db:"type"`
	Amount                 float64   `json:"amount" db:"amount"`
	Status                 string    `json:"status" db:"status"`
	BankReferenceNo        string    `json:"bankReferenceNo" db:"bank_reference_no"`
	BeneficiaryBankName    string    `json:"beneficiaryBankName" db:"beneficiary_bank_name"`
	BeneficiaryAccountNo   string    `json:"beneficiaryAccountNo" db:"beneficiary_account_no"`
	BeneficiaryAccountName string    `json:"beneficiaryAccountName" db:"beneficiary_account_name"`
	// Used Internally
	TransactionID       string             `json:"-" db:"transaction_id"`
	BankTransferUUID    string             `json:"-" db:"bank_transfer_uuid"`
	ExternalID          string             `json:"-" db:"external_id"`
	ReferenceID         string             `json:"-" db:"reference_id"`
	MerchantID          string             `json:"-" db:"merchant_id"`
	Currency            string             `json:"-" db:"currency"`
	Description         string             `json:"-" db:"description"`
	BeneficiaryBankCode string             `json:"-" db:"beneficiary_bank_code"`
	RawMetadata         types.NullJSONText `json:"-" db:"metadata"`
	Metadata            Metadata           `json:"-" db:"-"`
}

func (r *WithdrawalDetailResponse) ToOpenAPIWithdrawalResponse() OpenAPIWithdrawalResponse {
	if r.RawMetadata.Valid {
		_ = r.RawMetadata.Unmarshal(&r.Metadata)
	}
	if r.Metadata.WithdrawType == "" {
		r.Metadata.WithdrawType = constant.WithdrawalDestBankTransfer
		if strings.TrimSpace(r.BeneficiaryAccountName) == "Payout Balance" {
			r.Metadata.WithdrawType = constant.WithdrawalDestBalanceTransfer
			r.Metadata.BalanceType = constant.WithdrawalPayoutBalanceDestination
		}
	}
	return OpenAPIWithdrawalResponse{
		ID:         r.Id,
		MerchantId: r.MerchantID,
		Withdrawal: OpenAPIWithdrawalDetailResponse{
			ReferenceID:  r.ReferenceID,
			WithdrawType: r.Metadata.WithdrawType,
			BalanceType:  r.Metadata.BalanceType,
			IsFullAmount: r.Metadata.IsFullAmount,
			Amount: &commonModel.Amount{
				Currency: r.Currency,
				Value:    fmt.Sprintf("%.0f", r.Amount),
			},
			Description: r.Description,
		},
		Status:    r.Status,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
		UpdatedAt: r.UpdatedAt.Format(time.RFC3339),
	}
}

type InquiryTransactionResponse struct {
	*WithdrawalDetailResponse
	UpdatedAt         time.Time `json:"updatedAt"`
	Status            string    `json:"status"`
	ReasonType        string    `json:"reasonType"`
	ReasonDescription string    `json:"reasonDescription"`
}

type WithdrawalInsightResponse struct {
	TodayTotalSuccess *WithdrawalInsightItem `json:"todayTotalSuccess"`
	TodayTotalPending *WithdrawalInsightItem `json:"todayTotalPending"`
	TodayTotalFailed  *WithdrawalInsightItem `json:"todayTotalFailed"`
}

type WithdrawalInsightItem struct {
	Total       int                `json:"total"`
	TotalAmount commonModel.Amount `json:"totalAmount"`
}

type WithdrawalDownloadResponse struct {
	URL string `json:"url"`
}

type WithdrawalTransferBalanceResponse struct {
	Id string `json:"id"`
}

type OpenAPIWithdrawalResponse struct {
	ID         string                          `json:"id"`
	MerchantId string                          `json:"merchantId"`
	Withdrawal OpenAPIWithdrawalDetailResponse `json:"withdrawal"`
	Status     string                          `json:"status"`
	Reason     string                          `json:"reason,omitempty"`
	CreatedAt  string                          `json:"createdAt,omitempty"`
	UpdatedAt  string                          `json:"updatedAt,omitempty"`
}

type OpenAPIWithdrawalDetailResponse struct {
	ReferenceID  string              `json:"referenceId"`
	WithdrawType string              `json:"withdrawType"`
	BalanceType  string              `json:"balanceType"`
	IsFullAmount bool                `json:"isFullAmount"`
	Amount       *commonModel.Amount `json:"amount,omitempty"`
	Description  string              `json:"description"`
}

func (w *WithdrawalProcessResponse) ToOpenAPIWithdrawalResponse(r *OpenAPIWithdrawalRequest) *OpenAPIWithdrawalResponse {
	return &OpenAPIWithdrawalResponse{
		ID:         w.Id,
		MerchantId: r.MerchantId,
		Withdrawal: OpenAPIWithdrawalDetailResponse{
			ReferenceID:  r.ReferenceId,
			WithdrawType: r.WithdrawType,
			BalanceType:  r.BalanceType,
			IsFullAmount: r.IsFullAmount,
			Description:  r.Description,
			Amount:       &w.Amount,
		},
		Status:    w.Status,
		CreatedAt: w.CreatedAt.Format(time.RFC3339),
		UpdatedAt: w.UpdatedAt.Format(time.RFC3339),
	}
}

type WithdrawalStatusCallbackRequest OpenAPIWithdrawalResponse

type OpenAPIBankAccountListResponse struct {
	BankAccounts []bankAccount.BankAccountResponse `json:"bankAccounts"`
}

type WithdrawalChangeStatusResponse struct {
	ID         string `json:"id"`
	MerchantID string `json:"merchantId"`
	Status     string `json:"status"`
}
