package merchantRcn

import (
	"reflect"
	"testing"

	cimbProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cimbProcessor"
)

func TestBuildMerchantRcnResponse(t *testing.T) {
	mockResponse := &cimbProcessorModel.InquiryCorporateCreditCardResponse{
		PartnerReferenceNo: "PRN123456",
		AdditionalInfo: cimbProcessorModel.AdditionalInfoResponse{
			PagingFlag:  "Y",
			PagingKey:   1,
			ReferenceNo: "REF123456",
			Data: cimbProcessorModel.Data{
				CompanyAccountInformation: cimbProcessorModel.CompanyAccountInfo{
					AccountName: "John Doe",
					GlobalLimit: cimbProcessorModel.CurrencyInfo{
						Value:    "1000000",
						Currency: "IDR",
					},
					CurrentBalance: cimbProcessorModel.CurrencyInfo{
						Value:    "500000",
						Currency: "IDR",
					},
				},
				AccountInformation: cimbProcessorModel.AccountInfo{
					RetailCurrBalance: cimbProcessorModel.CurrencyInfo{
						Value:    "200000",
						Currency: "IDR",
					},
					CreditLimit: cimbProcessorModel.CurrencyInfo{
						Value:    "1000000",
						Currency: "IDR",
					},
				},
				CardInformation: []cimbProcessorModel.CardInfo{
					{
						BankCardNo:    "1234567890123456",
						AnnualFeeDate: "2025-01-01",
						FullName:      "John Doe",
					},
				},
			},
		},
	}

	expectedResponse := MerchantRcnResponse{
		PartnerReferenceNo: "PRN123456",
		AdditionalInfo: AdditionalInfo{
			PagingFlag:  "Y",
			PagingKey:   1,
			ReferenceNo: "REF123456",
			CompanyAccountInformation: CompanyAccountInfo{
				AccountName: "J**n D*e",
				GlobalLimit: AmountInfo{
					Value:    "1000000",
					Currency: "IDR",
				},
				CurrentBalance: AmountInfo{
					Value:    "500000",
					Currency: "IDR",
				},
			},
			AccountInformation: AccountInfo{
				RetailCurrBalance: AmountInfo{
					Value:    "200000",
					Currency: "IDR",
				},
				CreditLimit: AmountInfo{
					Value:    "1000000",
					Currency: "IDR",
				},
			},
			CardInformation: []CardInfo{
				{
					BankCardNo:    "123456******3456",
					AnnualFeeDate: "2025-01-01",
					FullName:      "J**n D*e",
				},
			},
		},
	}

	t.Run("Valid Response", func(t *testing.T) {
		got := BuildMerchantRcnResponse(mockResponse)
		if !reflect.DeepEqual(got, expectedResponse) {
			t.Errorf("BuildMerchantRcnResponse() = %v, want %v", got, expectedResponse)
		}
	})
}
