package requestAccountInquiries

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankAccount"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/jmoiron/sqlx/types"
)

type RequestAccountInquiries struct {
	UUID                   string             `json:"uuid" db:"uuid" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	MerchantID             string             `json:"merchantId" db:"merchant_id" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	AccountInquiryId       sql.NullString     `json:"accountInquiryId" db:"account_inquiry_id" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	BeneficiaryBankCode    string             `json:"beneficiaryBankCode" db:"beneficiary_bank_code" example:"008"`
	BeneficiaryBankName    sql.NullString     `json:"beneficiaryBankName" db:"beneficiary_bank_name" example:"Bank 008"`
	BeneficiaryAccountNo   sql.NullString     `json:"beneficiaryAccountNo" db:"beneficiary_account_no" example:"8000800808"`
	BeneficiaryAccountName sql.NullString     `json:"beneficiaryAccountName" db:"beneficiary_account_name" example:"Yories Yolanda"`
	Status                 sql.NullString     `json:"status" db:"status" example:"VALID"`
	Metadata               types.NullJSONText `json:"-" db:"metadata" example:"{}"`
	CreatedAt              time.Time          `json:"createdAt" db:"created_at" example:"2021-01-01T00:00:00Z"`

	MetadataObj Metadata `json:"metadata" db:"-"`
}

type RequestAccountInquiryWithMaster struct {
	RequestAccountInquiries

	MasterBeneficiaryAccountName string `json:"-" db:"master_beneficiary_account_name"`
}

type RequestAccountInquiriesHttpRequest struct {
	RequestInquiryID   string             `json:"-" validate:"-"`
	MerchantID         string             `json:"-" validate:"required,uuid"`
	ChannelCode        string             `json:"channelCode" validate:"required"`
	ChannelInformation ChannelInformation `json:"channelInformation" validate:"required"`
	AdditionalInfo     map[string]any     `json:"additionalInfo"`

	ParentMerchantID string
}

type ChannelInformation struct {
	AccountName   string `json:"accountName" validate:"required"`
	AccountNumber string `json:"accountNumber" validate:"required,numeric" example:"8000800808"`
}

type RequestAccountInquiriesHttpResponse struct {
	UUID           string                              `json:"uuid"`
	MerchantID     string                              `json:"merchantId"`
	InquiryResult  InquiryResult                       `json:"inquiryResult"`
	AdditionalInfo *AccountInquiryResultAdditionalInfo `json:"additionalInfo,omitempty"`
}

type InquiryResult struct {
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type AccountInquiryResultAdditionalInfo struct {
	IsVirtualAccount bool `json:"isVirtualAccount"`
}

func BuildAccountInquiryAdditionalInfo(
	merchantID string,
	metadata *Metadata,
	latestSnapCoreResponse *snapCoreModel.InquiryAccountResponseData,
) *AccountInquiryResultAdditionalInfo {
	if !constant.IsAccountInquiryVirtualAccountFlagDisplayedForMerchant(merchantID) {
		return nil
	}

	return &AccountInquiryResultAdditionalInfo{
		IsVirtualAccount: ResolveIsVirtualAccount(metadata, latestSnapCoreResponse),
	}
}

func ResolveIsVirtualAccount(metadata *Metadata, latestSnapCoreResponse *snapCoreModel.InquiryAccountResponseData) bool {
	if metadata != nil && metadata.SnapCoreResponse != nil {
		return metadata.SnapCoreResponse.IsVirtualAccount
	}

	if latestSnapCoreResponse != nil {
		return latestSnapCoreResponse.IsVirtualAccount
	}

	return false
}

type Metadata struct {
	DetailStatus     string                                    `json:"detailStatus,omitempty"`
	SnapCoreResponse *snapCoreModel.InquiryAccountResponseData `json:"snapCoreResponse,omitempty"`
	OnBehalf         *merchant.OnBehalfObject                  `json:"onBehalf,omitempty"`
	FeeOnBehalf      *feeModel.TrxFeeOnBehalfMetadata          `json:"feeOnBehalf,omitempty"`
	FeeDetail        *feeModel.FeeMetadataObject               `json:"feeDetail"`
}

func NewDetailStatusRequestInquiry(status string, accountName, bankRecord, merchantId, processorResponseCode string) (string, string) {
	switch status {
	case constant.RequestAccountInquiryStatusInvalid:
		return status, util.MapAccountInquirySnapResponseToDetailStatus(processorResponseCode)
	case constant.RequestAccountInquiryStatusPending:
		return status, constant.ReqInquiryDetailStatusPending
	case constant.RequestAccountInquiryStatusWarning:
		return util.SimilarityCheck(accountName, bankRecord, status, merchantId)
	}

	return status, ""
}

func (inq *RequestAccountInquiryWithMaster) SetMetadataNullJSONText() {
	metadataB, _ := json.Marshal(inq.MetadataObj)
	inq.Metadata = types.NullJSONText{
		JSONText: metadataB,
		Valid:    true,
	}
}
