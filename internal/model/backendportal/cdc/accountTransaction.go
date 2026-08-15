package cdcModel

import (
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"

	"github.com/shopspring/decimal"
)

// AccountTransaction represents the account_transactions table row
type AccountTransaction struct {
	UUID                string                           `json:"uuid" db:"uuid"`
	ReferenceID         string                           `json:"reference_id" db:"reference_id"`
	MerchantID          string                           `json:"merchant_id" db:"merchant_id"`
	AccountID           string                           `json:"account_id" db:"account_id"`
	MerchantReferenceID *string                          `json:"merchant_reference_id" db:"merchant_reference_id"`
	Currency            string                           `json:"currency" db:"currency"`
	Credit              decimal.Decimal                  `json:"credit" db:"credit"`
	Debit               decimal.Decimal                  `json:"debit" db:"debit"`
	Type                string                           `json:"type" db:"type"`
	Channel             string                           `json:"channel" db:"channel"`
	Status              string                           `json:"status" db:"status"`
	ReasonType          *string                          `json:"reason_type" db:"reason_type"`
	ReasonDescription   *string                          `json:"reason_description" db:"reason_description"`
	Remarks             string                           `json:"remarks" db:"remarks"`
	SettlementAt        *time.Time                       `json:"settlement_at" db:"settlement_at"`
	SettlementStatus    *string                          `json:"settlement_status" db:"settlement_status"`
	SettlementModel     *string                          `json:"settlement_model" db:"settlement_model"`
	AdditionalInfo      AccountTransactionAdditionalInfo `json:"additional_info" db:"-"`
	CreatedAt           time.Time                        `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time                        `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time                       `json:"deleted_at" db:"deleted_at"`
	Reference           *string                          `json:"reference" db:"reference"`
	RawAdditionalInfo   *string                          `json:"-" db:"additional_info"` // Only for direct data retrieval to the database
}

type aliasAccountTransaction AccountTransaction

func (a *AccountTransaction) UnmarshalJSON(data []byte) error {

	temp := struct {
		aliasAccountTransaction
		AdditionalInfo *string `json:"additional_info"`
	}{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*a = AccountTransaction(temp.aliasAccountTransaction)

	if temp.AdditionalInfo == nil {
		return nil
	}
	return json.Unmarshal([]byte(*temp.AdditionalInfo), &a.AdditionalInfo)
}

func (a *AccountTransaction) GetSettlementModel() string {
	if a.SettlementModel == nil {
		return constant.SettlementModelAggregator
	}

	switch *a.SettlementModel {
	default:
		return constant.SettlementModelAggregator

	case constant.SettlementModelFacilitator, constant.SettlementModelDirect:
		return constant.SettlementModelDirect
	}
}

type AccountTransactionAdditionalInfo struct {
	Type              string                                       `json:"type,omitempty"`
	ReferenceType     string                                       `json:"referenceType,omitempty"`
	FeeDetail         *FeeDetail                                   `json:"feeDetail,omitempty"`
	SettlementDetail  *SettlementDetail                            `json:"settlementDetail,omitempty"`
	Notes             string                                       `json:"notes,omitempty"`
	SubPaymentSummary *orchestratorModel.MetadataSubPaymentSummary `json:"subPaymentSummary,omitempty"`
}

type FeeDetail struct {
	Type          string  `json:"type,omitempty"`
	Method        string  `json:"method,omitempty"`
	Channel       string  `json:"channel,omitempty"`
	DeductionType string  `json:"deductionType,omitempty"`
	Notes         string  `json:"notes,omitempty"`
	FinalAmount   float64 `json:"finalAmount,omitempty"`
}

type SettlementDetail struct {
	Type                 string     `json:"type"`
	EstimateSettlementAt *time.Time `json:"estimateSettlementAt,omitempty"`
}
