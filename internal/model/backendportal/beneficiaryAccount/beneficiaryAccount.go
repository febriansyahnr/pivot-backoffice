package beneficiaryAccountModel

import (
	"database/sql"
	"time"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/xb"

	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

type BeneficiaryAccount struct {
	UUID                   string    `json:"uuid" db:"uuid" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	BeneficiaryAccountNo   string    `json:"beneficiaryAccountNo" db:"beneficiary_account_no" example:"8000800808"`
	BeneficiaryAccountName string    `json:"beneficiaryAccountName" db:"beneficiary_account_name" example:"Yories Yolanda"`
	BeneficiaryBankCode    string    `json:"beneficiaryBankCode" db:"beneficiary_bank_code" example:"008"`
	BeneficiaryBankName    string    `json:"beneficiaryBankName" db:"beneficiary_bank_name" example:"Bank 008"`
	CreatedAt              time.Time `json:"createdAt" db:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt              time.Time `json:"updatedAt" db:"updated_at" example:"2021-01-01T00:00:00Z"`

	MerchantID string             `json:"-" db:"merchant_id"`
	DeletedAt  sql.NullTime       `json:"-" db:"deleted_at"`
	Metadata   types.NullJSONText `json:"-" db:"metadata"`

	MetadataObj Metadata `json:"metadata" db:"-"`
}

type Account struct {
	UUID                   string                 `json:"uuid" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	MerchantID             string                 `json:"merchantId"`
	BeneficiaryAccountNo   string                 `json:"beneficiaryAccountNo" example:"8000800808"`
	BeneficiaryAccountName string                 `json:"beneficiaryAccountName" example:"Yories Yolanda"`
	BeneficiaryBankCode    string                 `json:"beneficiaryBankCode" example:"008"`
	BeneficiaryBankName    string                 `json:"beneficiaryBankName" example:"Bank 008"`
	CreatedAt              time.Time              `json:"createdAt" example:"2021-01-01T00:00:00Z"`
	UpdatedAt              time.Time              `json:"updatedAt" example:"2021-01-01T00:00:00Z"`
	MetadataObj            Metadata               `json:"metadata"`
	AdditionalInfo         *AccountAdditionalInfo `json:"additionalInfo,omitempty"`
}

type AccountAdditionalInfo struct {
	IsVirtualAccount bool `json:"isVirtualAccount"`
}

type Metadata struct {
	IsXb                       bool                                                `json:"isXb"`
	XbDetail                   *xbModel.BeneficiaryDataResponse                    `json:"xbDetail,omitempty"`
	OnBehalf                   *merchantModel.OnBehalfObject                       `json:"onBehalf,omitempty"`
	RequestInquiryStatus       string                                              `json:"requestInquiryStatus,omitempty"`
	IsVirtualAccount           bool                                                `json:"isVirtualAccount,omitempty"`
	BeneficiaryPayoutLimitRule *disbursementModel.BeneficiaryPayoutLimitRuleConfig `json:"beneficiaryPayoutLimitRule,omitempty"`
	IsOverbooking              bool                                                `json:"isOverbooking"`
	MaxAmount                  decimal.Decimal                                     `json:"maxAmount"`
	PayoutFeeAmount            float64                                             `json:"payoutFeeAmount"`
}

func (a *Account) ToBeneficiaryAccount() *BeneficiaryAccount {
	return &BeneficiaryAccount{
		UUID:                   a.UUID,
		BeneficiaryAccountNo:   a.BeneficiaryAccountNo,
		BeneficiaryAccountName: a.BeneficiaryAccountName,
		BeneficiaryBankCode:    a.BeneficiaryBankCode,
		BeneficiaryBankName:    a.BeneficiaryBankName,
		CreatedAt:              a.CreatedAt,
		UpdatedAt:              a.UpdatedAt,
		MerchantID:             a.MerchantID,
		MetadataObj:            a.MetadataObj,
	}
}
