package snapCoreModel

import (
	"time"

	constant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
)

type CreateVirtualAccountRequest struct {
	CustomerNo     string                  `json:"customerNo" validate:"required"`
	MID            string                  `json:"mid"`      // 4 digit MID related to Merchant
	VaNumber       string                  `json:"vaNumber"` // not FQVA, only random VA
	AccountName    string                  `json:"accountName" validate:"required"`
	AccountEmail   string                  `json:"accountEmail"`
	AccountPhone   string                  `json:"accountPhone"`
	SubCompany     string                  `json:"subCompany"`
	TotalAmount    Amount                  `json:"totalAmount" validate:"required"`
	FeeAmount      Amount                  `json:"feeAmount"`
	ExpiredAt      *time.Time              `json:"expiredAt"`
	Acquirer       string                  `json:"acquirer" validate:"required"`
	BillDetails    []BillDetail            `json:"billDetails"`
	FreeTexts      []Description           `json:"freeTexts"`
	IsCloseAmount  bool                    `json:"isClosedAmount"`
	IsSingleUse    bool                    `json:"isSingleUse"`
	AdditionalInfo *map[string]interface{} `json:"additionalInfo"`
	Purpose        string                  `json:"purpose"`

	MerchantID string `json:"-"`
}

type UpdateVirtualAccountRequest struct {
	Number       string       `json:"vaNumber" validate:"required" example:"7663123400000012"`
	TotalAmount  *Amount      `json:"totalAmount"`
	ExpiredAt    *time.Time   `json:"expiredAt" example:"2024-03-10T17:02:03+07:00"`
	AccountEmail string       `json:"accountEmail" example:"updateva@email.com"`
	AccountPhone string       `json:"accountPhone" example:"081239123"`
	AccountName  string       `json:"accountName" example:"Tester VA"`
	BillDetails  []BillDetail `json:"billDetails"`
	CustomerNo   string       `json:"customerNo"`
	MaxAmount    *Amount      `json:"maxAmount"`
	MinAmount    *Amount      `json:"minAmount"`
}

type DeleteVirtualAccountRequest struct {
	UUID   string `json:"uuid"`
	Number string `json:"number"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type Description struct {
	English   string `json:"english"`
	Indonesia string `json:"indonesia"`
}

type BillDetail struct {
	BillerReferenceId string                  `json:"billerReferenceId"`
	BillCode          string                  `json:"billCode"`
	BillNo            string                  `json:"billNo"`
	BillName          string                  `json:"billName"`
	BillShortName     string                  `json:"billShortName"`
	BillDescription   Description             `json:"billDescription"`
	BillSubCompany    string                  `json:"billSubCompany"`
	BillAmount        Amount                  `json:"billAmount"`
	AdditionalInfo    *map[string]interface{} `json:"additionalInfo"`
}

type VirtualAccountTrxType struct {
	IsCloseAmount bool
	IsSingleUsed  bool
}

var trxType = map[string]VirtualAccountTrxType{
	constant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC:    {IsCloseAmount: false, IsSingleUsed: false},
	constant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC:  {IsCloseAmount: true, IsSingleUsed: false},
	constant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC: {IsCloseAmount: true, IsSingleUsed: true},
}

func VaTrxType(name string) VirtualAccountTrxType {
	return trxType[name]
}

func FindVaTrxTypeByCriteria(isCloseAmount, isSingleUsed bool) string {
	for key, val := range trxType {
		if val.IsCloseAmount == isCloseAmount && val.IsSingleUsed == isSingleUsed {
			return key
		}
	}

	return ""
}

type CreateVirtualAccountConfigRequest struct {
	MerchantID      string `json:"merchant_id"`
	MID             string `json:"mid"`
	BinPrefix       string `json:"bin_prefix"`
	Type            string `json:"type"`
	IntegrationType string `json:"integration_type"`
	Acquirer        string `json:"acquirer"`
}

type GetVirtualAccountConfigRequest struct {
	MerchantID      string `url:"merchant_id"`
	MID             string `url:"mid"`
	BinPrefix       string `url:"bin_prefix"`
	Type            string `url:"type"`
	IntegrationType string `url:"integration_type"`
	Acquirer        string `url:"acquirer"`
	Status          string `url:"status"`
}

type UpdateVirtualAccountConfigPrefixRequest struct {
	MerchantID      string                                    `json:"merchant_id"`
	Acquirer        string                                    `json:"acquirer"`
	IntegrationType string                                    `json:"integration_type"`
	Detail          []*UpdateVirtualAccountConfigPrefixDetail `json:"detail"`
}

type UpdateVirtualAccountConfigPrefixDetail struct {
	Type       string `json:"type"`
	StartRange string `json:"start_range"`
	EndRange   string `json:"end_range"`
}

type BlockVirtualAccountRequest struct {
	MerchantID string `json:"merchantId"`
}

type UnblockVirtualAccountRequest struct {
	MerchantID string `json:"merchantId"`
}

type InquiryStatusVARequest struct {
	VirtualAccount string `json:"virtualAccount"`
	SkipPublish    bool   `json:"skipPublish"`
	ExternalID     string `json:"externalId"`
}
