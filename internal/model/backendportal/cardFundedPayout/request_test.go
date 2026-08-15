package cardFundedPayoutModel_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/vendor"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPayoutActionRequest(t *testing.T) {
	userID := "f00edbd2-c694-4edd-9320-7a3d4f8a388e"
	merchantID := "d77820fc-05b5-41d8-a61a-7ec2cf49b914"
	userName := "John Doe"

	request := PayoutActionRequest{}
	request.SetUserID(userID)
	request.SetUserName(userName)
	request.SetMerchantID(merchantID)

	assert.Equal(t, userID, request.UserID)
	assert.Equal(t, merchantID, request.MerchantID)
	assert.Equal(t, userName, request.UserName)
}

func TestCreatePayoutRequestToCreateDisbursement(t *testing.T) {
	vendorID := "019cfa95-de34-7fd4-b56d-4db4d41cb715"
	cardID := "019cf96e-a0e6-7582-ade9-9da6ddc7e76d"
	merchantID := "aec6636d-7a02-4d93-a4c5-006b9c235068"
	userID := "c4b9ad6a-3773-4d40-8e90-3eae3fb8e4b4"

	request := CreatePayoutRequest{
		VendorID:    vendorID,
		ReferenceID: "REF/AIRLINES/202603/0001", // NOSONAR
		Amount: commonModel.AmountRequest{
			Currency: constant.CurrencyIDR,
			Value:    100_000, // NOSONAR
		},
		Remarks:          "Test remarks", // NOSONAR
		SettlementMethod: constant.PaymentSettlementMethodStandard,
		CardID:           cardID,
		PayoutActionRequest: PayoutActionRequest{
			MerchantID: merchantID,
			UserID:     userID,
		},
	}

	vendorData := &vendor.Vendor{
		UUID:          vendorID,
		MerchantID:    merchantID,
		Name:          "PT Test Vendor", // NOSONAR
		BankCode:      "014",            // NOSONAR
		BankName:      "BCA",            // NOSONAR
		AccountNumber: "1234567890",     // NOSONAR
		AccountName:   "Test Account",   // NOSONAR
		Status:        constant.StatusActive,
	}

	card := &GetSavedCardResponse{
		ID:             cardID,
		CardName:       "Test Card",     // NOSONAR
		CardOrigin:     "LOCAL",         // NOSONAR
		PaymentChannel: "VISA",          // NOSONAR
		IssuingBank:    "BNI",           // NOSONAR
		Last4:          "4242",          // NOSONAR
		ExpiryMonth:    "12",            // NOSONAR
		ExpiryYear:     "2025",          // NOSONAR
		MerchantName:   "Test Merchant", // NOSONAR
	}

	fee := &feeModel.FeeMetadataObject{
		Type:          constant.ReferencePaymentFundedPayout,
		Method:        "CREDIT_CARD", // NOSONAR
		Channel:       "LOCAL_VISA",  // NOSONAR
		DeductionType: "DIRECT",      // NOSONAR
		AmountType:    "FIXED",       // NOSONAR
		Amount:        5000,          // NOSONAR
		FinalAmount:   5000,          // NOSONAR
	}

	result := request.ToCreateDisbursement(vendorData, card, fee)

	// Assert basic fields
	assert.NoError(t, uuid.Validate(result.UUID), "UUID should be generated")
	assert.Equal(t, request.ReferenceID, result.ReferenceID)
	assert.Equal(t, merchantID, result.MerchantID)
	assert.Equal(t, card.MerchantName, result.SenderName)
	assert.Equal(t, vendorData.BankCode, result.BeneficiaryBankCode)
	assert.Equal(t, vendorData.BankName, *result.BeneficiaryBankName)
	assert.Equal(t, vendorData.AccountNumber, result.BeneficiaryAccountNo)
	assert.Equal(t, vendorData.AccountName, result.BeneficiaryAccountName)
	assert.Equal(t, request.Amount.Currency, result.Currency)

	// Assert amounts
	assert.True(t, decimal.NewFromFloat(request.Amount.Value).Equal(result.Amount))
	assert.True(t, decimal.NewFromFloat(fee.FinalAmount).Equal(*result.Fee))
	assert.True(t, decimal.NewFromFloat(request.Amount.Value+fee.FinalAmount).Equal(result.TotalAmount))

	// Assert status and metadata
	assert.Equal(t, constant.DisbursementStatusWaiting, result.Status)
	assert.Equal(t, request.Remarks, *result.Remark)
	assert.True(t, result.Metadata.Valid)

	// Assert metadata object
	assert.Equal(t, vendorData.UUID, result.MetadataObj.CardFundedDetail.VendorID)
	assert.Equal(t, vendorData.Name, result.MetadataObj.CardFundedDetail.VendorName)
	assert.Equal(t, request.SettlementMethod, result.MetadataObj.CardFundedDetail.SettlementMethod)
	assert.Equal(t, card.ID, result.MetadataObj.CardFundedDetail.Card.ID)
	assert.Equal(t, card.CardName, result.MetadataObj.CardFundedDetail.Card.CardName)
	assert.Equal(t, card.PaymentChannel, result.MetadataObj.CardFundedDetail.Card.PaymentChannel)
	assert.Equal(t, card.IssuingBank, result.MetadataObj.CardFundedDetail.Card.IssuingBank)
	assert.Equal(t, card.Last4, result.MetadataObj.CardFundedDetail.Card.Last4Digits)
	assert.Equal(t, card.ExpiryMonth, result.MetadataObj.CardFundedDetail.Card.ExpiryMonth)
	assert.Equal(t, card.ExpiryYear, result.MetadataObj.CardFundedDetail.Card.ExpiryYear)

	// Assert fee detail in metadata
	assert.Equal(t, fee.Type, result.MetadataObj.FeeDetail.Type)
	assert.Equal(t, fee.FinalAmount, result.MetadataObj.FeeDetail.FinalAmount)
}
