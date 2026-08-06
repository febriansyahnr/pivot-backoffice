package cimbProcessorModel

import (
	"strconv"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
)

type InquiryCorporateCreditCardResponse struct {
	ResponseCode       string                 `json:"responseCode"`
	ResponseMessage    string                 `json:"responseMessage"`
	PartnerReferenceNo string                 `json:"partnerReferenceNo"`
	AdditionalInfo     AdditionalInfoResponse `json:"additionalInfo"`
}
type AdditionalInfoResponse struct {
	PagingFlag  string `json:"pagingFlag"`
	PagingKey   int    `json:"pagingKey"`
	ReferenceNo string `json:"referenceNo"`
	Data        Data   `json:"data"`
}
type Data struct {
	CompanyAccountInformation CompanyAccountInfo `json:"companyAccountInformation"`
	AccountInformation        AccountInfo        `json:"accountInformation"`
	CardInformation           []CardInfo         `json:"cardInformation"`
}

type CompanyAccountInfo struct {
	AccountName        string       `json:"accountName"`
	GlobalLimit        CurrencyInfo `json:"globalLimit"`
	GlobalLimitCash    CurrencyInfo `json:"globalLimitCash"`
	CurrentBalance     CurrencyInfo `json:"currentBalance"`
	CurrentBalanceCash CurrencyInfo `json:"currentBalanceCash"`
	AvailableLimit     CurrencyInfo `json:"availableLimit"`
	AvailableLimitCash CurrencyInfo `json:"availableLimitCash"`
}

type AccountInfo struct {
	RetailCurrBalance  CurrencyInfo `json:"retailCurrBalance"`
	CashCurrBalance    CurrencyInfo `json:"cashCurrBalance"`
	DisputeBalance     CurrencyInfo `json:"disputeBalance"`
	PayDueDate         string       `json:"payDueDate"`
	CreditLimit        CurrencyInfo `json:"creditLimit"`
	CashLimit          CurrencyInfo `json:"cashLimit"`
	OsBalance          CurrencyInfo `json:"osBalance"`
	CashLimitBalance   CurrencyInfo `json:"cashLimitBalance"`
	CreditLimitBalance CurrencyInfo `json:"creditLimitBalance"`
}

type CardInfo struct {
	BankCardNo    string `json:"bankCardNo"`
	AnnualFeeDate string `json:"annualFeeDate"`
	FullName      string `json:"fullName"`
}

type CurrencyInfo struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type InquiryTransactionCorporateCreditCardResponse struct {
	ResponseCode       string                                              `json:"responseCode"`
	ResponseMessage    string                                              `json:"responseMessage"`
	PartnerReferenceNo string                                              `json:"partnerReferenceNo"`
	AdditionalInfo     InquiryTransactionCorporateCreditCardAdditionalInfo `json:"additionalInfo"`
}

type InquiryTransactionCorporateCreditCardAdditionalInfo struct {
	PagingFlag      string                                         `json:"pagingFlag"`
	PagingKey       string                                         `json:"pagingKey"`
	ReferenceNo     string                                         `json:"referenceNo"`
	RecordType      string                                         `json:"recordType"`
	BillingCycle    string                                         `json:"billingCycle"`
	PostingDate     int                                            `json:"postingDate"`
	TransactionData []InquiryTransactionCorporateCreditCardTrxData `json:"transactionData"`
}

type InquiryTransactionCorporateCreditCardTrxData struct {
	CardHolderName   string       `json:"cardHolderName"`
	BankCardNo       string       `json:"bankCardNo"`
	TransactionDate  string       `json:"transactionDate"`
	SettlementDate   string       `json:"settlementDate"`
	AuthorizationNo  string       `json:"authorizationNo"`
	SourceAmount     CurrencyInfo `json:"sourceAmount"`
	BillingAmount    CurrencyInfo `json:"billingAmount"`
	MerchantName     string       `json:"merchantName"`
	MerchantCountry  string       `json:"merchantCountry"`
	MerchantCategory string       `json:"merchantCategory"`
	ArnNo            string       `json:"arnNo"`
}

func (r *InquiryTransactionCorporateCreditCardResponse) ToProcessorVccTransactionInquiryResponse(requestPostingDate string) *vccSettlement.ProcessorVccTransactionInquiryResponse {
	processorTrx := make([]vccSettlement.VccTransactionInquiryTrxData, len(r.AdditionalInfo.TransactionData))
	for i, val := range r.AdditionalInfo.TransactionData {
		processorTrx[i] = vccSettlement.VccTransactionInquiryTrxData{
			CardHolderName:  val.CardHolderName,
			BankCardNo:      val.BankCardNo,
			TransactionDate: val.TransactionDate,
			SettlementDate:  val.SettlementDate,
			AuthorizationNo: val.AuthorizationNo,
			SourceAmount: commonModel.Amount{
				Value:    val.SourceAmount.Value,
				Currency: val.SourceAmount.Currency,
			},
			BillingAmount: commonModel.Amount{
				Value:    val.BillingAmount.Value,
				Currency: val.BillingAmount.Currency,
			},
			MerchantName:     val.MerchantName,
			MerchantCountry:  val.MerchantCountry,
			MerchantCategory: val.MerchantCategory,
			ArnNo:            val.ArnNo,
		}
	}

	hasNextPage := r.AdditionalInfo.PagingFlag == "Y"
	count, _ := strconv.ParseInt(r.AdditionalInfo.PagingKey, 0, 32)
	return &vccSettlement.ProcessorVccTransactionInquiryResponse{
		ReferenceNo:     r.AdditionalInfo.ReferenceNo,
		RecordType:      r.AdditionalInfo.RecordType,
		BillingCycle:    r.AdditionalInfo.BillingCycle,
		PostingDate:     requestPostingDate,
		TransactionData: processorTrx,
		HasNextPage:     hasNextPage,
		Count:           int32(count),
	}
}
