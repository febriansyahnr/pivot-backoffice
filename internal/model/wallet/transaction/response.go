package walletTransactionModel

import (
	"slices"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/jmoiron/sqlx/types"
)

type MerchantTransactionHistoryListResp struct {
	Id               string    `json:"id" db:"uuid"`
	ReferenceId      string    `json:"referenceId" db:"merchant_reference_id"`
	Type             string    `json:"type" db:"type"`
	Channel          string    `json:"channel" db:"channel"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time `json:"updatedAt" db:"updated_at"`
	Amount           float64   `json:"amount" db:"amount"`
	Status           string    `json:"status" db:"status"`
	SettlementStatus string    `json:"settlementStatus" db:"settlement_status"`
	// Data is only displayed in the report download results
	CreatedBy         string `json:"-" db:"created_by"`
	BankAccountNumber string `json:"-" db:"beneficiary_account_no"`
	BankAccountName   string `json:"-" db:"beneficiary_account_name"`
}

type MerchantTransactionDetailResp struct {
	Id                string         `json:"id" db:"uuid"`
	ReferenceId       string         `json:"referenceId" db:"merchant_reference_id"`
	Type              string         `json:"type" db:"type"`
	Channel           string         `json:"channel" db:"channel"`
	CreatedAt         string         `json:"createdAt" db:"created_at"`
	UpdatedAt         string         `json:"updatedAt" db:"updated_at"`
	Amount            float64        `json:"amount" db:"amount"`
	CreatedBy         string         `json:"createdBy" db:"created_by"`
	AdditionalInfo    any            `json:"additionalInfo" db:"-"`
	RawAdditionalInfo types.JSONText `json:"-" db:"additional_info"`
	Status            string         `json:"status" db:"status"`
	SettlementStatus  string         `json:"settlementStatus" db:"settlement_status"`
}

var transactionTypeTitles = map[string]string{
	"FEE_WALLET_TRANSACTION": "Wallet Transaction Fee",
	"FEE_MERCHANT_PAYMENT":   "Merchant Payment Fee",
	"FEE_TOP_UP":             "Top Up Fee",
	"FEE_BANK_TRANSFER":      "Bank Transfer Fee",
	"FEE_BILL":               "PPOB Fee",
	"FEE_WITHDRAWAL":         "Withdrawal Fee",
	"MERCHANT_PAYMENT":       "Merchant Payment",
	"WITHDRAWAL":             "Withdrawal",
	"MERCHANT_TOP_UP":        "Merchant Top Up",
}

func TransactionTypeToTitle(typ string) string {
	title, ok := transactionTypeTitles[typ]
	if !ok {
		return util.ToTitle(typ)
	}
	return title
}

func (t *MerchantTransactionHistoryListResp) TransactionTypeToTitle() string {
	typ := t.Type
	if typ == constant.TypeFee {
		typ += "_" + t.Channel
	}
	return TransactionTypeToTitle(typ)
}

func (t *MerchantTransactionHistoryListResp) GetReferenceID() string {
	if slices.Contains([]string{constant.TypeMerchantTopUp, constant.TypeWithdrawal}, t.Type) {
		return "-"
	}
	return t.ReferenceId
}
