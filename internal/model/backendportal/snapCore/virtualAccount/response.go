package snapCoreModel

import (
	"time"

	"github.com/jmoiron/sqlx/types"

	snapVa "github.com/paper-indonesia/pdk/go/snap/structs/va"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	common "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

type CreateVirtualAccountResponse struct {
	Data    CreateVirtualAccountResponseData `json:"data"`
	Code    string                           `json:"code"`
	Message string                           `json:"message"`
	Error   *common.ErrorDetail              `json:"error,omitempty"`
}

type CreateVirtualAccountResponseData struct {
	ID               string                 `json:"uuid"`
	VirtualAccountNo string                 `json:"number"`
	AccountName      string                 `json:"accountName"`
	Acquirer         string                 `json:"acquirer"`
	ExpiredAt        time.Time              `json:"expiredAt"`
	Status           string                 `json:"status,omitempty"`
	TotalAmount      Amount                 `json:"totalAmount"`
	CreatedAt        time.Time              `json:"createdAt"`
	MinAmount        Amount                 `json:"minAmount"`
	MaxAmount        Amount                 `json:"maxAmount"`
	IsClosedAmount   bool                   `json:"isClosedAmount"`
	IsSingleUse      bool                   `json:"isSingleUse"`
	BillDetails      *[]snapVa.BillDetail   `json:"billDetails,omitempty"`
	AdditionalInfo   map[string]interface{} `json:"-"` // for unit testing only
}

type UpdateVirtualAccountResponse struct {
	Data    UpdateVirtualAccountResponseData `json:"data"`
	Code    string                           `json:"code"`
	Message string                           `json:"message"`
	Error   interface{}                      `json:"error,omitempty"`
}

type UpdateVirtualAccountResponseData struct {
	UUID         string       `json:"uuid"`
	Acquirer     string       `json:"acquirer"`
	Number       string       `json:"number"`
	ExpiredAt    *time.Time   `json:"expiredAt,omitempty"`
	TotalAmount  Amount       `json:"totalAmount"`
	AccountName  string       `json:"accountName"`
	AccountEmail string       `json:"accountEmail"`
	AccountPhone string       `json:"accountPhone"`
	BillDetails  []BillDetail `json:"billDetails"`
	CustomerNo   string       `json:"customerNo"`
	CreatedAt    time.Time    `json:"createdAt"`
	MaxAmount    *Amount      `json:"maxAmount,omitempty"`
	MinAmount    *Amount      `json:"minAmount,omitempty"`
}

type DeleteVirtualAccountResponse struct {
	*common.StandardResponse
	Data *DeleteVirtualAccountResponseData `json:"data"`
}

type DeleteVirtualAccountResponseData struct {
	UUID        string `json:"uuid"`
	Acquirer    string `json:"acquirer"`
	Number      string `json:"number"`
	AccountName string `json:"accountName"`
}

type VirtualAccountConfigResponse struct {
	*common.StandardResponse
	Data *VirtualAccountConfigResponseData `json:"data"`
}

type GetVirtualAccountConfigResponse struct {
	*common.StandardResponse
	Data []*VirtualAccountConfigResponseData `json:"data"`
}

type VirtualAccountConfigResponseData struct {
	UUID              string                       `json:"uuid"`
	MerchantID        string                       `json:"merchant_id"`
	MID               string                       `json:"mid"`
	BinPrefix         string                       `json:"bin_prefix"`
	BinMin            int                          `json:"bin_min"`
	BinMax            int                          `json:"bin_max"`
	Type              string                       `json:"type"`
	IntegrationType   string                       `json:"integration_type"`
	IntegrationMethod string                       `json:"integration_method"`
	ClientID          string                       `json:"client_id"`
	Credential        map[string]interface{}       `json:"credential"`
	Acquirer          string                       `json:"acquirer"`
	Status            string                       `json:"status"`
	Metadata          types.NullJSONText           `json:"metadata"`
	MetadataObj       VirtualAccountConfigMetadata `json:"-"`
}

type VirtualAccountConfigMetadata struct {
	MerchantPrefix struct {
		StartRange string `json:"start_range"`
		EndRange   string `json:"end_range"`
	} `json:"merchant_prefix"`
}

type UpdateVirtualAccountConfigPrefixResponse struct {
	*common.StandardResponse
}

type BlockVirtualAccountResponse struct {
	*common.StandardResponse
	Data []*BlockVirtualAccountResponseData `json:"data"`
}

type BlockVirtualAccountResponseData struct {
	UUID        string `json:"uuid"`
	Acquirer    string `json:"acquirer"`
	Number      string `json:"number"`
	AccountName string `json:"accountName"`
}

type UnblockVirtualAccountResponse struct {
	*common.StandardResponse
	Data []*UnblockVirtualAccountResponseData `json:"data"`
}

type UnblockVirtualAccountResponseData struct {
	Acquirer    string `json:"acquirer"`
	Number      string `json:"number"`
	AccountName string `json:"accountName"`
	HasError    bool   `json:"hasError"`
}

type InquiryStatusVAResponse struct {
	Data    InquiryStatusVAResponseData `json:"data"`
	Code    string                      `json:"code"`
	Message string                      `json:"message"`
	Error   *common.ErrorDetail         `json:"error,omitempty"`
}

type InquiryStatusVAResponseData struct {
	ResponseCode       string                  `json:"responseCode"`
	ResponseMessage    string                  `json:"responseMessage"`
	VirtualAccountData *InquiryStatusVAData    `json:"virtualAccountData,omitempty"`
	AdditionalInfo     *map[string]interface{} `json:"additionalInfo,omitempty"`
}

type InquiryStatusVAData struct {
	PartnerServiceId   string              `json:"partnerServiceId,omitempty"`
	CustomerNo         string              `json:"customerNo,omitempty"`
	VirtualAccountNo   string              `json:"virtualAccountNo,omitempty"`
	VirtualAccountName string              `json:"virtualAccountName,omitempty"`
	TrxDateTime        string              `json:"trxDateTime,omitempty"`
	ReferenceNo        string              `json:"referenceNo,omitempty"`
	PaidAmount         *commonModel.Amount `json:"paidAmount,omitempty"`
	TotalAmount        *commonModel.Amount `json:"totalAmount,omitempty"`
	InquiryRequestId   string              `json:"inquiryRequestId,omitempty"`
	PaymentRequestId   string              `json:"paymentRequestId,omitempty"`
	TransactionDate    string              `json:"transactionDate,omitempty"`
	PaidBills          string              `json:"paidBills,omitempty"`
	FlagAdvise         string              `json:"flagAdvise,omitempty"`
	PaymentFlagStatus  string              `json:"paymentFlagStatus,omitempty"`
}

func (r *InquiryStatusVAResponse) IsPaid() bool {
	if r.Data.ResponseCode == "" {
		return false
	}

	return util.IsPatternMatch(constant.SnapCoreResponseCodeSuccessPattern, r.Data.ResponseCode)
}

func (r *InquiryStatusVAResponse) IsNotFound() bool {
	if r.Data.ResponseCode == "" {
		return false
	}

	return util.IsPatternMatch(constant.SnapCoreResponseCodeVANotFoundPattern, r.Data.ResponseCode)
}

func (r *InquiryStatusVAResponse) IsConflict() bool {
	if r.Data.ResponseCode == "" {
		return false
	}

	return util.IsPatternMatch(constant.SnapCoreResponseCodeVAConflictPattern, r.Data.ResponseCode)
}
