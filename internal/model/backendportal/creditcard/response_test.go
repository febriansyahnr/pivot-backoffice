package card_test

import (
	"testing"

	"github.com/google/uuid"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	card "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcardCoreProcessor"
	"github.com/stretchr/testify/assert"
)

func TestToCreditcardCoreGetTransactionListResponse(t *testing.T) {
	mockUUID := uuid.New()

	// Test case 1: Response with all data fields populated
	mockResponseWithAllData := &creditcardCoreProcessorModel.GetTransactionDataList{
		Result: []*creditcardCoreProcessorModel.GetTransactionListResult{
			{
				TransactionDate:       "2025-02-23",
				PaymentUUID:           mockUUID,
				ClientTransactionID:   "TX12345",
				AcquirerTransactionID: []string{"ACQ123"},
				PayerTransactionID:    "PAY123",
				VoidTransactionID:     "VOID123",
				AuthorizationData: &creditcardCoreProcessorModel.AuthorizationData{
					AuthorizationResult:   "Success",
					OrderID:               "ORD123",
					TransactionStatus:     "Completed",
					AuthorizationID:       "AUTH123",
					ApprovalCode:          "AP123",
					BankMerchantID:        "BM123",
					AcquirerTransactionID: "ACQ123",
					TransactionReference:  "REF123",
					CvvResult:             "Match",
					AcquirerResponseCode:  "00",
					Stan:                  "STAN123",
					AvsResult:             "Y",
					ErrorMessage:          "",
				},
				AuthenticationData: &creditcardCoreProcessorModel.AuthenticationData{
					AuthenticationResult: "Success",
					AuthenticationID:     "AUTH123",
					PaRes:                "PARES123",
					VeRes:                "VERES123",
					XID:                  "XID123",
					CAVV:                 "CAVV123",
					EciCode:              "05",
					ThreeDsVer:           "2.0",
					ChallengeCode:        "CC123",
				},
				CardData: &creditcardCoreProcessorModel.CardData{
					First8Digit: "12345678",
					Last4Digit:  "1234",
					CardType:    "Credit",
					CardBrand:   "Visa",
					CardIssuing: "Test Bank",
					CountryCode: "ID",
					Fingerprint: "FP123",
				},
				IssuingBank:     "Test Bank",
				ChargeStatus:    "Success",
				ChargeAt:        "2025-02-23T10:00:00Z",
				VoidStatus:      "Failed",
				VoidAt:          "",
				TransactionType: []string{"SALE"},
				FDS:             "Passed",
			},
		},
		Pagination: creditcardCoreProcessorModel.PaginationResponse{
			PageNumber:  1,
			PageLimit:   10,
			TotalRecord: 1,
			TotalPage:   1,
		},
	}

	// Test case 2: Response with no data fields
	mockResponseWithNoData := &creditcardCoreProcessorModel.GetTransactionDataList{
		Result: []*creditcardCoreProcessorModel.GetTransactionListResult{
			{
				TransactionDate:       "2025-02-23",
				PaymentUUID:           mockUUID,
				ClientTransactionID:   "TX12345",
				AcquirerTransactionID: []string{"ACQ123"},
				PayerTransactionID:    "PAY123",
				VoidTransactionID:     "VOID123",
				AuthorizationData:     nil,
				AuthenticationData:    nil,
				CardData:              nil,
				IssuingBank:           "Test Bank",
				ChargeStatus:          "Success",
				ChargeAt:              "2025-02-23T10:00:00Z",
				VoidStatus:            "Failed",
				VoidAt:                "",
				TransactionType:       []string{"SALE"},
				FDS:                   "Passed",
			},
		},
		Pagination: creditcardCoreProcessorModel.PaginationResponse{
			PageNumber:  1,
			PageLimit:   10,
			TotalRecord: 1,
			TotalPage:   1,
		},
	}

	// Test with all data fields
	resultWithAllData := card.ToCreditcardCoreGetTransactionListResponse(mockResponseWithAllData)
	assert.NotNil(t, resultWithAllData)
	assert.Len(t, resultWithAllData.Data, 1)

	// Check that all data was properly converted
	transactionWithAllData := resultWithAllData.Data.([]*card.GetTransactionListResult)[0]

	// Verify authorization data
	assert.NotNil(t, transactionWithAllData.AuthorizationData)
	assert.Equal(t, "Success", transactionWithAllData.AuthorizationData.AuthorizationResult)
	assert.Equal(t, "ORD123", transactionWithAllData.AuthorizationData.OrderID)
	assert.Equal(t, "Completed", transactionWithAllData.AuthorizationData.TransactionStatus)
	assert.Equal(t, "AUTH123", transactionWithAllData.AuthorizationData.AuthorizationID)
	assert.Equal(t, "AP123", transactionWithAllData.AuthorizationData.ApprovalCode)
	assert.Equal(t, "BM123", transactionWithAllData.AuthorizationData.BankMerchantID)
	assert.Equal(t, "ACQ123", transactionWithAllData.AuthorizationData.AcquirerTransactionID)
	assert.Equal(t, "REF123", transactionWithAllData.AuthorizationData.TransactionReference)
	assert.Equal(t, "Match", transactionWithAllData.AuthorizationData.CvvResult)
	assert.Equal(t, "00", transactionWithAllData.AuthorizationData.AcquirerResponseCode)
	assert.Equal(t, "STAN123", transactionWithAllData.AuthorizationData.Stan)
	assert.Equal(t, "Y", transactionWithAllData.AuthorizationData.AvsResult)
	assert.Equal(t, "", transactionWithAllData.AuthorizationData.ErrorMessage)

	// Verify authentication data
	assert.NotNil(t, transactionWithAllData.AuthenticationData)
	assert.Equal(t, "Success", transactionWithAllData.AuthenticationData.AuthenticationResult)
	assert.Equal(t, "AUTH123", transactionWithAllData.AuthenticationData.AuthenticationID)
	assert.Equal(t, "PARES123", transactionWithAllData.AuthenticationData.PaRes)
	assert.Equal(t, "VERES123", transactionWithAllData.AuthenticationData.VeRes)
	assert.Equal(t, "XID123", transactionWithAllData.AuthenticationData.XID)
	assert.Equal(t, "CAVV123", transactionWithAllData.AuthenticationData.CAVV)
	assert.Equal(t, "05", transactionWithAllData.AuthenticationData.EciCode)
	assert.Equal(t, "2.0", transactionWithAllData.AuthenticationData.ThreeDsVer)
	assert.Equal(t, "CC123", transactionWithAllData.AuthenticationData.ChallengeCode)

	// Verify card data
	assert.NotNil(t, transactionWithAllData.CardData)
	assert.Equal(t, "12345678", transactionWithAllData.CardData.First8Digit)
	assert.Equal(t, "1234", transactionWithAllData.CardData.Last4Digit)
	assert.Equal(t, "Credit", transactionWithAllData.CardData.CardType)
	assert.Equal(t, "Visa", transactionWithAllData.CardData.CardBrand)
	assert.Equal(t, "Test Bank", transactionWithAllData.CardData.CardIssuing)
	assert.Equal(t, "ID", transactionWithAllData.CardData.CountryCode)
	assert.Equal(t, "FP123", transactionWithAllData.CardData.Fingerprint)

	// Test with no data fields
	resultWithNoData := card.ToCreditcardCoreGetTransactionListResponse(mockResponseWithNoData)
	assert.NotNil(t, resultWithNoData)
	assert.Len(t, resultWithNoData.Data, 1)

	// Check that nil fields were properly handled
	transactionWithNoData := resultWithNoData.Data.([]*card.GetTransactionListResult)[0]
	assert.Nil(t, transactionWithNoData.AuthorizationData)
	assert.Nil(t, transactionWithNoData.AuthenticationData)
	assert.Nil(t, transactionWithNoData.CardData)

	// Verify pagination data
	assert.Equal(t, commonModel.Meta{
		Page:       1,
		PerPage:    10,
		TotalItems: 1,
		TotalPages: 1,
	}, resultWithNoData.Meta)
}
