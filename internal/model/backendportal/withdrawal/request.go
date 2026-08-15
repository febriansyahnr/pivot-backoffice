package withdrawal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
)

type PreparationRequest struct {
	MerchantId  string `json:"-" validate:"required,uuid"`
	AccountName string `json:"-" validate:"required,oneof=PAYMENT WALLET VIRTUAL_TERMINAL"`
}

type WithdrawalRequest struct {
	AccountName            string  `json:"accountName" validate:"required,oneof=PAYMENT WALLET VIRTUAL_TERMINAL"`
	ReferenceID            string  `json:"-"` // for OPEN API request
	IsFullAmount           bool    `json:"-"` // for OPEN API request
	Type                   string  `json:"-" validate:"required,oneof=MANUAL AUTOMATED"`
	Amount                 float64 `json:"amount" validate:"required,min=1"`
	Destination            string  `json:"destination" validate:"required,oneof=BANK_TRANSFER BALANCE_TRANSFER"`
	BeneficiaryBankCode    string  `json:"beneficiaryBankCode" validate:"required_if=Destination BANK_TRANSFER,omitempty,numeric"`
	BeneficiaryAccountNo   string  `json:"beneficiaryAccountNo" validate:"required_if=Destination BANK_TRANSFER,omitempty,numeric"`
	DestinationAccountName string  `json:"destinationAccountName" validate:"required_if=Destination BALANCE_TRANSFER,omitempty,oneof=DISBURSEMENT"`
	UserId                 string  `json:"-" validate:"required,uuid"`
	MerchantId             string  `json:"-" validate:"required,uuid"`
	Reason                 string  `json:"-" validate:"-"`
	Description            string  `json:"description"`

	Source string `json:"-"` // Used Internally
}

type OpenAPIWithdrawalRequest struct {
	ReferenceId      string             `json:"referenceId" validate:"required"`
	WithdrawType     string             `json:"withdrawType" validate:"required,oneof=BANK_TRANSFER BALANCE_TRANSFER"`
	BalanceType      string             `json:"balanceType" validate:"required_if=WithdrawType BALANCE_TRANSFER,omitempty,oneof=PAYOUT_BALANCE"`
	Amount           commonModel.Amount `json:"amount" validate:"required_if=IsFullAmount false"`
	IsFullAmount     bool               `json:"isFullAmount"`
	Description      string             `json:"description"`
	MerchantId       string             `json:"-" validate:"required,uuid"`
	ParentMerchantId string             `json:"-" validate:"required,uuid"`
}

func (r *OpenAPIWithdrawalRequest) ToWithdrawalRequest() *WithdrawalRequest {
	amount, _ := strconv.ParseFloat(r.Amount.Value, 64)

	request := &WithdrawalRequest{
		AccountName:            constant.AccountNamePayment,
		ReferenceID:            r.ReferenceId,
		Destination:            r.WithdrawType,
		BeneficiaryBankCode:    "",
		BeneficiaryAccountNo:   "",
		DestinationAccountName: "",
		Amount:                 amount,
		IsFullAmount:           r.IsFullAmount,
		Description:            r.Description,
		MerchantId:             r.MerchantId,
		Type:                   constant.WithdrawalManual,
		Source:                 constant.SourceOpenApi,
	}

	if r.WithdrawType == constant.WithdrawalDestBalanceTransfer {
		request.DestinationAccountName = constant.AccountNameDisbursement
	}

	return request
}

type WithdrawalHistoryRequest struct {
	*WithdrawalListRequest
	Page    int `json:"page,omitempty" validate:"min=0"`
	PerPage int `json:"perPage,omitempty" validate:"min=0"`
}

type WithdrawalListRequest struct {
	AccountName  string    `json:"accountName" validate:"required,oneof=PAYMENT DISBURSEMENT"`
	StrStartDate string    `json:"startDate" validate:"required,datetime=2006-01-02"`
	StrEndDate   string    `json:"endDate" validate:"required,datetime=2006-01-02"`
	Status       string    `json:"status,omitempty" validate:"omitempty,oneof=PENDING SUCCESS FAILED"`
	Id           string    `json:"id,omitempty" validate:"omitempty,uuid"`
	Sort         string    `json:"sort,omitempty" validate:"omitempty,oneof=date -date"` // Sortable: -date (DESC) date (ASC)
	MerchantId   string    `json:"-" validate:"required,uuid"`
	StartDate    time.Time `json:"-" validate:"-"`
	EndDate      time.Time `json:"-" validate:"-"`
}

type WithdrawalDetailRequest struct {
	Id          string `json:"id"`
	AccountName string `json:"accountName"`
	MerchantId  string `json:"merchantId"`
}

type InquiryTransactionRequest struct {
	Id         string `json:"-" validate:"required,uuid"`
	MerchantId string `json:"merchantId" validate:"required,uuid"`
}

type RetryTransactionRequest struct {
	WithdrawalID         string `json:"-"`
	MerchantID           string `json:"merchantId" validate:"required,uuid"`
	ForceFailed          bool   `json:"forceFailed"`
	ForceRetry           bool   `json:"forceRetry"`
	BypassProcessorCheck bool   `json:"bypassProcessorCheck"`
}

type WithdrawalInsightRequest struct {
	MerchantID string
	Status     string
}

type WithdrawalInsightQuery struct {
	Status      string  `db:"status"`
	Total       int     `db:"total"`
	TotalAmount float64 `db:"total_amount"`
	Currency    string  `db:"currency"`
}

type WithdrawalTransferBalanceRequest struct {
	UserID     string `json:"-" validate:"required,uuid"`
	MerchantID string `json:"-" validate:"required,uuid"`

	SourceAccountName      string  `json:"source" validate:"required,oneof=PAYMENT"`
	DestinationAccountName string  `json:"destination" validate:"required,oneof=DISBURSEMENT"`
	Amount                 float64 `json:"amount" validate:"required,min=1"`
}

func (w *WithdrawalListRequest) HashFilterKey(endDate time.Time) string {
	sha := sha256.New()
	_, _ = fmt.Fprintf(sha,
		"%s:%s:%v:%v:%s:%s:%s", w.MerchantId, w.AccountName, w.StartDate, endDate, w.Status, w.Id, w.Sort)

	hash := sha.Sum(nil)
	return hex.EncodeToString(hash)
}

type FailedWithdrawalAlertRequest struct {
	AlertTitle                 string
	WithdrawalID               string
	MerchantID                 string
	BalanceName                string
	BeneficiaryAccountName     string
	BeneficiaryAccountNo       string
	BeneficiaryAccountBankName string
	WithdrawType               string
	Status                     string
	Amount                     float64
	Reason                     string
}

type WithdrawalChangeStatusRequest struct {
	WithdrawalID      string  `json:"-"`
	MerchantID        string  `json:"merchantId" validate:"required,uuid"`
	ReasonType        *string `json:"reasonType" validate:"required_if=Status FAILED,omitempty,oneof=OTHER INSUFFICIENT_ESCROW_FUND INVALID_ACCOUNT BLOCKED_BY_HARSYA REVERSAL BLOCKED_BY_BANK"`
	ReasonDescription *string `json:"reasonDescription,omitempty"`
	Status            string  `json:"status" validate:"required,oneof=FAILED SUCCESS"`
}
