package cimbProcessorModel_test

import (
	"testing"

	cimbProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cimbProcessor"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"

	"github.com/stretchr/testify/assert"
)

func TestToProcessorVccTransactionInquiryResponse(t *testing.T) {
	requestPostingDate := "2026-04-30" // NOSONAR

	input := &cimbProcessorModel.InquiryTransactionCorporateCreditCardResponse{
		ResponseCode:       "2001700",    // NOSONAR
		ResponseMessage:    "Successful", // NOSONAR
		PartnerReferenceNo: "REF123",     // NOSONAR
		AdditionalInfo: cimbProcessorModel.InquiryTransactionCorporateCreditCardAdditionalInfo{
			PagingFlag:   "Y",       // NOSONAR
			PagingKey:    "10",      // NOSONAR
			ReferenceNo:  "REF-001", // NOSONAR
			RecordType:   "ALL",     // NOSONAR
			BillingCycle: "04",      // NOSONAR
			PostingDate:  20260430,  // NOSONAR
			TransactionData: []cimbProcessorModel.InquiryTransactionCorporateCreditCardTrxData{
				{
					CardHolderName:   "John Doe",                                                           // NOSONAR
					BankCardNo:       "1234567890123456",                                                   // NOSONAR
					TransactionDate:  "2026-04-28",                                                         // NOSONAR
					SettlementDate:   "2026-04-29",                                                         // NOSONAR
					AuthorizationNo:  "AUTH001",                                                            // NOSONAR
					SourceAmount:     cimbProcessorModel.CurrencyInfo{Value: "100000.00", Currency: "IDR"}, // NOSONAR
					BillingAmount:    cimbProcessorModel.CurrencyInfo{Value: "100000.00", Currency: "IDR"}, // NOSONAR
					MerchantName:     "Merchant A",                                                         // NOSONAR
					MerchantCountry:  "ID",                                                                 // NOSONAR
					MerchantCategory: "5411",                                                               // NOSONAR
					ArnNo:            "ARN001",                                                             // NOSONAR
				},
			},
		},
	}

	expected := &vccSettlement.ProcessorVccTransactionInquiryResponse{
		ReferenceNo:  "REF-001", // NOSONAR
		RecordType:   "ALL",     // NOSONAR
		BillingCycle: "04",      // NOSONAR
		PostingDate:  requestPostingDate,
		HasNextPage:  true, // NOSONAR
		Count:        10,   // NOSONAR
		TransactionData: []vccSettlement.VccTransactionInquiryTrxData{
			{
				CardHolderName:   "John Doe",         // NOSONAR
				BankCardNo:       "1234567890123456", // NOSONAR
				TransactionDate:  "2026-04-28",       // NOSONAR
				SettlementDate:   "2026-04-29",       // NOSONAR
				AuthorizationNo:  "AUTH001",
				SourceAmount:     commonModel.Amount{Value: "100000.00", Currency: "IDR"}, // NOSONAR
				BillingAmount:    commonModel.Amount{Value: "100000.00", Currency: "IDR"}, // NOSONAR
				MerchantName:     "Merchant A",                                            // NOSONAR
				MerchantCountry:  "ID",                                                    // NOSONAR
				MerchantCategory: "5411",                                                  // NOSONAR
				ArnNo:            "ARN001",                                                // NOSONAR
			},
		},
	}

	result := input.ToProcessorVccTransactionInquiryResponse(requestPostingDate)

	assert.Equal(t, expected, result)
}
