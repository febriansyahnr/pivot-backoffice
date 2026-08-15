package merchant

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx/types"
)

type UploadDocumentResp struct {
	Id string `json:"id"`
}

type UpsertBoardOfDirectorResp struct {
	Id string `json:"id"`
}

type ListBoardOfDirectorResp struct {
	Directors    []BoardOfDirectorResp `json:"directorList"`
	Commissioner []BoardOfDirectorResp `json:"commissionerList"`
}

type BoardOfDirectorResp struct {
	Id             string         `json:"id" db:"id"`
	Name           string         `json:"name" db:"name"`
	Position       string         `json:"position" db:"position"`
	PositionLong   string         `json:"positionLong" db:"position_long"`
	IdentityNumber string         `json:"identityNumber" db:"identity_number"`
	File           types.JSONText `json:"-" db:"identity_file"`
	IdentityFile   string         `json:"identityFile" db:"-"`
	Shares         float64        `json:"shares,omitempty" db:"shares"`
	CreatedBy      string         `json:"createdBy" db:"created_by"`
	CreatedAt      time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time      `json:"updatedAt" db:"updated_at"`
	// Internal Data
	ObjFile DocLocation `json:"-" db:"-"`
}

type GenSignatureResp struct {
	Signature string `json:"signature"`
}

type TransactionConfigResp struct {
	MerchantId         string             `json:"merchantId"`
	MerchantName       string             `json:"merchantName,omitempty"`
	MerchantType       string             `json:"merchantType,omitempty"`
	TransactionConfigs TransactionConfigs `json:"transactionConfigs"`
}

type RawTransactionConfig struct {
	MerchantName      string             `json:"-" db:"merchant_name"`
	MerchantType      string             `json:"-" db:"merchant_type"`
	Disbursement      types.NullJSONText `json:"-" db:"disbursement"`
	Withdrawal        types.NullJSONText `json:"-" db:"withdrawal"`
	DailyDisbursement types.NullJSONText `json:"-" db:"daily_disbursement"`
}

type DisbursementMerchantConfig struct {
	DailyLimitMerchantId   string `db:"daily_limit_merchant_id"`
	DailyLimitMerchantType string `db:"daily_limit_merchant_type"`
}

type SubMerchantResponse struct {
	UUID              string `json:"uuid"`
	Name              string `json:"name"`
	ShortName         string `json:"shortName"`
	Description       string `json:"description"`
	Website           string `json:"website"`
	Address           string `json:"address"`
	PostCode          string `json:"postCode"`
	Logo              string `json:"logo"`
	MerchantEmail     string `json:"merchantEmail"`
	MerchantPhone     string `json:"merchantPhone"`
	PICEmail          string `json:"picEmail"`
	PICPhone          string `json:"picPhone"`
	PICName           string `json:"picName"`
	PICJobTitle       string `json:"picJobTitle"`
	BusinessType      string `json:"businessType"`
	BusinessStructure string `json:"businessStructure"`
	BusinessCountry   string `json:"businessCountry"`
	ParentIndustry    string `json:"parentIndustry"`
	ChildIndustry     string `json:"childIndustry"`
	MCC               string `json:"mcc"`
	CountryOfEntity   string `json:"countryOfEntity"`
	DigitalStatus     string `json:"digitalStatus"`
	ParentID          string `json:"parentId"`

	AutoWithdrawal      *string                      `json:"autoWithdrawal,omitempty"`
	BankAccount         *MerchantBankAccountResponse `json:"bankAccount,omitempty"`
	SubAccountStatus    string                       `json:"subAccountStatus,omitempty"`
	SubAccountType      string                       `json:"subAccountType,omitempty"`
	SubAccountKycStatus string                       `json:"subAccountKycStatus,omitempty"`
}

func (m *Merchant) ToSubMerchantResponse() *SubMerchantResponse {
	return &SubMerchantResponse{
		UUID:              m.UUID,
		Name:              m.Name,
		ShortName:         m.ShortName,
		Description:       m.Description,
		Website:           m.Website,
		Logo:              m.Logo,
		MerchantEmail:     m.MerchantEmail,
		MerchantPhone:     m.MerchantPhone,
		BusinessCountry:   m.BusinessCountry.String,
		BusinessType:      m.BusinessType.String,
		BusinessStructure: m.BusinessStructure.String,
		ParentIndustry:    m.ParentIndustry.String,
		ChildIndustry:     m.ChildIndustry.String,
		MCC:               m.MCC.String,
		CountryOfEntity:   m.CountryOfEntity.String,
		DigitalStatus:     m.DigitalStatus.String,
		PICName:           m.PICName.String,
		PICEmail:          m.PICEmail,
		PICPhone:          m.PICPhone,
		PICJobTitle:       m.PICJobTitle.String,
		Address:           m.Address,
		PostCode:          m.PostCode,
		ParentID:          m.ParentID.String,
		BankAccount:       m.BankAccount,
		SubAccountStatus:  m.Status,
	}
}

type MerchantStatusResponse struct {
	UUID         string `json:"-" db:"uuid"`
	Status       string `json:"status" db:"status" redis:"status"`
	KYCStatus    string `json:"kycStatus" db:"kyc_status" redis:"kycStatus"`
	RiskLevel    string `json:"riskLevel" db:"risk_level" redis:"riskLevel"`
	ReasonStatus string `json:"reasonStatus" db:"reason_status" redis:"reasonStatus"`
}

func (o MerchantStatusResponse) MarshalBinary() ([]byte, error) {
	return json.Marshal(o)
}

func (o *MerchantStatusResponse) UnmarshalBinary(buf []byte) error {
	return json.Unmarshal(buf, o)
}

type DepositSettingResponse struct {
	MerchantName   string `json:"merchantName" db:"merchant_name"`
	AutoWithdrawal string `json:"autoWithdrawal" db:"auto_withdrawal"`
}

type CreateFeeConfigOnBehalfResponse struct {
	Id string `json:"id"`
	*CreateFeeConfigOnBehalfRequest
}

type GetFeeConfigOnBehalfResponse struct {
	*GetFeeConfigOnBehalfRequest
	Configs []FeeConfigOnBehalfResponse `json:"configs"`
}

type FeeConfigOnBehalfResponse struct {
	Id            string    `json:"id" db:"id"`
	Type          string    `json:"type" db:"type"`
	SubMerchantId *string   `json:"subMerchantId" db:"sub_merchant_id"`
	AmountType    string    `json:"amountType" db:"amount_type"`
	Amount        float64   `json:"amount" db:"amount"`
	Percentage    float64   `json:"percentage" db:"percentage"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

type UpdateMerchantKYCResponse struct {
	UUID   string `json:"id"`
	Status string `json:"status"`
}

type BillingFeeResponse struct {
	MerchantId     string                                `json:"merchantId"`
	MerchantName   string                                `json:"merchantName"`
	Total          int                                   `json:"total"`
	TotalFeeAmount float64                               `json:"totalFeeAmount"`
	Details        map[string][]BillingFeeDetailResponse `json:"details"`
	SubMerchants   []SubMerchantBillingResponse          `json:"subMerchants"`
}

type SubMerchantBillingResponse struct {
	SubMerchantId   string                                `json:"subMerchantId"`
	SubMerchantName string                                `json:"subMerchantName"`
	Total           int                                   `json:"total"`
	TotalFeeAmount  float64                               `json:"totalFeeAmount"`
	Details         map[string][]BillingFeeDetailResponse `json:"details"`
}

type BillingFeeDetailResponse struct {
	Type           string  `db:"type" json:"-"`
	Method         string  `db:"method" json:"method,omitempty"`
	Channel        string  `db:"channel" json:"channel,omitempty"`
	MerchantId     string  `db:"merchant_id" json:"-"`
	Total          int     `db:"total" json:"total"`
	TrxAmount      float64 `db:"trx_amount" json:"trxAmount"`
	FeeType        string  `db:"fee_type" json:"feeType"`
	FeeAmount      float64 `db:"fee_amount" json:"feeAmount"`
	FeePercentage  float64 `db:"fee_percentage" json:"feePercentage"`
	TotalFeeAmount float64 `db:"total_fee_amount" json:"totalFeeAmount"`
}

type BlockMerchantResponse struct {
	BlockedMerchantDetails
	SubMerchants []BlockedMerchantDetails `json:"subMerchants"`
}

type BlockedMerchantDetails struct {
	MerchantId   string    `json:"merchantId"`
	MerchantName string    `json:"merchantName"`
	BlockedAt    time.Time `json:"blockedAt"`
}

type UnblockMerchantResponse struct {
	UnblockedMerchantDetails
	SubMerchants []UnblockedMerchantDetails `json:"subMerchants"`
}

type UnblockedMerchantDetails struct {
	MerchantId   string    `json:"merchantId"`
	MerchantName string    `json:"merchantName"`
	UnblockedAt  time.Time `json:"unblockedAt"`
}

type FDSConfigResponse struct {
	FDSConfig
}

type GetFDSConfigResponse struct {
	MerchantID   string             `json:"merchantId" db:"merchant_id"`
	MerchantName string             `json:"merchantName" db:"merchant_name"`
	MerchantType string             `json:"merchantType" db:"merchant_type"`
	FDSConfig    FDSConfig          `json:"fdsConfig" db:"-"`
	RawFDSConfig types.NullJSONText `json:"-" db:"fds_configs"`
}

type PaymentInvestigationConfigResponse struct {
	Enabled             bool    `json:"enabled"`
	MerchantID          string  `json:"merchantId"`
	MerchantName        string  `json:"merchantName"`
	PivotPercentageLoss float64 `json:"pivotPercentageLoss"`
	PivotMaxLoss        float64 `json:"pivotMaxLoss"`
}
