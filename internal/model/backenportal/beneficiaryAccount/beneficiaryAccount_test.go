package beneficiaryAccountModel

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
)

func TestAccount_ToBeneficiaryAccount(t *testing.T) {
	// Setup test data
	now := time.Now()
	maxAmount := decimal.NewFromFloat(100000.0)

	account := &Account{
		UUID:                   "test-uuid",
		MerchantID:             "merchant-123",
		BeneficiaryAccountNo:   "1234567890",
		BeneficiaryAccountName: "Test Account",
		BeneficiaryBankCode:    "014",
		BeneficiaryBankName:    "Test Bank",
		CreatedAt:              now,
		UpdatedAt:              now,
		MetadataObj: Metadata{
			IsXb:          false,
			IsOverbooking: true,
			MaxAmount:     maxAmount,
			OnBehalf: &merchantModel.OnBehalfObject{
				ParentMerchantId: "parent-merchant-123",
			},
		},
	}

	// Convert Account to BeneficiaryAccount
	result := account.ToBeneficiaryAccount()

	// Assertions
	assert.Equal(t, account.UUID, result.UUID)
	assert.Equal(t, account.BeneficiaryAccountNo, result.BeneficiaryAccountNo)
	assert.Equal(t, account.BeneficiaryAccountName, result.BeneficiaryAccountName)
	assert.Equal(t, account.BeneficiaryBankCode, result.BeneficiaryBankCode)
	assert.Equal(t, account.BeneficiaryBankName, result.BeneficiaryBankName)
	assert.Equal(t, account.CreatedAt, result.CreatedAt)
	assert.Equal(t, account.UpdatedAt, result.UpdatedAt)
	assert.Equal(t, account.MerchantID, result.MerchantID)
	assert.Equal(t, account.MetadataObj, result.MetadataObj)
}
