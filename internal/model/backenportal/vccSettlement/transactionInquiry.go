package vccSettlement

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

type VccTransactionInquiryRequest struct {
	RcnId              string `json:"rcnId" validate:"required"`
	MerchantId         string `json:"merchantId"`
	RecordType         string `json:"recordType" validate:"required,oneof=ST"`
	BillingCycle       string `json:"billingCycle" validate:"required,number,len=2"`
	PostingDate        string `json:"postingDate" validate:"required,datetime=20060102"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
}

func (r *VccTransactionInquiryRequest) Validate() error {
	billingCycle, err := strconv.ParseInt(r.BillingCycle, 10, 8)
	if err != nil {
		return fmt.Errorf("invalid billingCycle")
	}
	if billingCycle < constant.BillingCycleFirst || billingCycle > constant.BillingCycleLast {
		return fmt.Errorf("incorrect billingCycle")
	}

	_, err = time.Parse(constant.VccDateFormat, r.PostingDate)
	if err != nil {
		return fmt.Errorf("invalid postingDate")
	}
	return nil
}

type VccTransactionInquiryResponse struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
}

type ProcessorVccTransactionInquiryResponse struct {
	HasNextPage     bool
	Count           int32
	ReferenceNo     string                         `json:"referenceNo"`
	RecordType      string                         `json:"recordType"`
	BillingCycle    string                         `json:"billingCycle"`
	PostingDate     string                         `json:"postingDate"`
	TransactionData []VccTransactionInquiryTrxData `json:"transactionData"`
}

type VccTransactionInquiryTrxData struct {
	CardHolderName   string             `json:"cardHolderName"`
	BankCardNo       string             `json:"bankCardNo"`
	TransactionDate  string             `json:"transactionDate"`
	SettlementDate   string             `json:"settlementDate"`
	AuthorizationNo  string             `json:"authorizationNo"`
	SourceAmount     commonModel.Amount `json:"sourceAmount"`
	BillingAmount    commonModel.Amount `json:"billingAmount"`
	MerchantName     string             `json:"merchantName"`
	MerchantCountry  string             `json:"merchantCountry"`
	MerchantCategory string             `json:"merchantCategory"`
	ArnNo            string             `json:"arnNo"`
}

func (r *ProcessorVccTransactionInquiryResponse) ToVccSettlementModel(rcnId string) []*VccSettlement {
	if len(r.TransactionData) == 0 {
		return nil
	}
	now := time.Now()
	settlementTrxList := make([]*VccSettlement, len(r.TransactionData))
	billingCycle, _ := strconv.ParseInt(r.BillingCycle, 0, 0)
	postingDateTime, err := time.Parse(constant.VccDateFormat, r.PostingDate)
	if err != nil {
		postingDateTime = time.Time{}
	}

	for i, val := range r.TransactionData {
		id, err := uuid.NewV7()
		if err != nil {
			id = uuid.New()
		}
		sourceAmountB, _ := val.SourceAmount.ToJson()
		billingAmountB, _ := val.BillingAmount.ToJson()
		transactionDateTime, _ := time.Parse(constant.VccDateFormat, val.TransactionDate)
		settlementDateTime, _ := time.Parse(constant.VccDateFormat, val.SettlementDate)

		settlementTrxList[i] = &VccSettlement{
			UUID:                    id.String(),
			RcnId:                   rcnId,
			AcquirerReferenceNumber: val.ArnNo,
			Status:                  constant.VccSettlementBilledStatus,
			ReferenceNo:             r.ReferenceNo,
			AuthorizationNo:         val.AuthorizationNo,
			PostingDate:             postingDateTime,
			BillingCycle:            int(billingCycle),
			SourceAmount:            sourceAmountB,
			BillingAmount:           billingAmountB,
			TransactionDate:         transactionDateTime,
			SettlementDate:          settlementDateTime,
			MerchantName:            val.MerchantName,
			MerchantCategory:        val.MerchantCategory,
			MerchantCountry:         val.MerchantCountry,
			CreatedAt:               now,
			UpdatedAt:               now,
		}
	}

	return settlementTrxList
}

type VccTransactionInquiryAlert struct {
	Title       string
	Recipient   []string
	Description string
	RcnId       string
	PostingDate string
}
