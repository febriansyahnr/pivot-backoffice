package merchantRcn

import (
	cimbProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cimbProcessor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

type MerchantRcnResponse struct {
	PartnerReferenceNo string         `json:"partnerReferenceNo"`
	AdditionalInfo     AdditionalInfo `json:"additionalInfo"`
}

type AdditionalInfo struct {
	PagingFlag                string             `json:"pagingFlag"`
	PagingKey                 int                `json:"pagingKey"`
	ReferenceNo               string             `json:"referenceNo"`
	CompanyAccountInformation CompanyAccountInfo `json:"companyAccountInformation"`
	AccountInformation        AccountInfo        `json:"accountInformation"`
	CardInformation           []CardInfo         `json:"cardInformation"`
}

type CompanyAccountInfo struct {
	AccountName        string     `json:"accountName"`
	GlobalLimit        AmountInfo `json:"globalLimit"`
	GlobalLimitCash    AmountInfo `json:"globalLimitCash"`
	CurrentBalance     AmountInfo `json:"currentBalance"`
	CurrentBalanceCash AmountInfo `json:"currentBalanceCash"`
	AvailableLimit     AmountInfo `json:"availableLimit"`
	AvailableLimitCash AmountInfo `json:"availableLimitCash"`
}

type AccountInfo struct {
	RetailCurrBalance  AmountInfo `json:"retailCurrBalance"`
	CashCurrBalance    AmountInfo `json:"cashCurrBalance"`
	DisputeBalance     AmountInfo `json:"disputeBalance"`
	PayDueDate         string     `json:"payDueDate"`
	CreditLimit        AmountInfo `json:"creditLimit"`
	CashLimit          AmountInfo `json:"cashLimit"`
	OsBalance          AmountInfo `json:"osBalance"`
	CashLimitBalance   AmountInfo `json:"cashLimitBalance"`
	CreditLimitBalance AmountInfo `json:"creditLimitBalance"`
}

type CardInfo struct {
	BankCardNo    string `json:"bankCardNo"`
	AnnualFeeDate string `json:"annualFeeDate"`
	FullName      string `json:"fullName"`
}

type AmountInfo struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

func BuildMerchantRcnResponse(cimbReponse *cimbProcessorModel.InquiryCorporateCreditCardResponse) MerchantRcnResponse {
	companyAccountInfo := CompanyAccountInfo{
		AccountName: util.MaskFullName(cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.AccountName),
		GlobalLimit: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.GlobalLimit.Value,
			Currency: cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.GlobalLimit.Currency,
		},
		GlobalLimitCash: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.GlobalLimitCash.Value,
			Currency: cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.GlobalLimitCash.Currency,
		},
		CurrentBalance: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.CurrentBalance.Value,
			Currency: cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.CurrentBalance.Currency,
		},
		CurrentBalanceCash: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.CurrentBalanceCash.Value,
			Currency: cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.CurrentBalanceCash.Currency,
		},
		AvailableLimit: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.AvailableLimit.Value,
			Currency: cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.AvailableLimit.Currency,
		},
		AvailableLimitCash: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.AvailableLimitCash.Value,
			Currency: cimbReponse.AdditionalInfo.Data.CompanyAccountInformation.AvailableLimitCash.Currency,
		},
	}
	accountInfo := AccountInfo{
		RetailCurrBalance: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.AccountInformation.RetailCurrBalance.Value,
			Currency: cimbReponse.AdditionalInfo.Data.AccountInformation.RetailCurrBalance.Currency,
		},
		CashCurrBalance: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.AccountInformation.CashCurrBalance.Value,
			Currency: cimbReponse.AdditionalInfo.Data.AccountInformation.CashCurrBalance.Currency,
		},
		DisputeBalance: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.AccountInformation.DisputeBalance.Value,
			Currency: cimbReponse.AdditionalInfo.Data.AccountInformation.DisputeBalance.Currency,
		},
		PayDueDate: "",
		CreditLimit: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.AccountInformation.CreditLimit.Value,
			Currency: cimbReponse.AdditionalInfo.Data.AccountInformation.CreditLimit.Currency,
		},
		CashLimit: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.AccountInformation.CashLimit.Value,
			Currency: cimbReponse.AdditionalInfo.Data.AccountInformation.CashLimit.Currency,
		},
		OsBalance: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.AccountInformation.OsBalance.Value,
			Currency: cimbReponse.AdditionalInfo.Data.AccountInformation.OsBalance.Currency,
		},
		CashLimitBalance: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.AccountInformation.CashLimitBalance.Value,
			Currency: cimbReponse.AdditionalInfo.Data.AccountInformation.CashLimitBalance.Currency,
		},
		CreditLimitBalance: AmountInfo{
			Value:    cimbReponse.AdditionalInfo.Data.AccountInformation.CreditLimitBalance.Value,
			Currency: cimbReponse.AdditionalInfo.Data.AccountInformation.CreditLimitBalance.Currency,
		},
	}

	var cardInfo []CardInfo
	for _, card := range cimbReponse.AdditionalInfo.Data.CardInformation {
		cardInfo = append(cardInfo, CardInfo{
			BankCardNo:    util.MaskCreditCardNumber(card.BankCardNo),
			AnnualFeeDate: card.AnnualFeeDate,
			FullName:      util.MaskFullName(card.FullName),
		})
	}

	return MerchantRcnResponse{
		PartnerReferenceNo: cimbReponse.PartnerReferenceNo,
		AdditionalInfo: AdditionalInfo{
			PagingFlag:                cimbReponse.AdditionalInfo.PagingFlag,
			PagingKey:                 cimbReponse.AdditionalInfo.PagingKey,
			ReferenceNo:               cimbReponse.AdditionalInfo.ReferenceNo,
			CompanyAccountInformation: companyAccountInfo,
			AccountInformation:        accountInfo,
			CardInformation:           cardInfo,
		},
	}
}
