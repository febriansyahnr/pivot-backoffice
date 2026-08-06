package vendor

import (
	"context"
	"errors"
	"testing"

	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IVendorRepository)
		input   *vendorModel.CreateVendorRequest
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			input: &vendorModel.CreateVendorRequest{
				MerchantID:          "merchant-123",
				Name:                "Test Vendor",
				BeneficialOwner:     "John Doe",
				BusinessCategory:    "E-Commerce",
				AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
				BankName:            "Bank ABC",
				BankCode:            "ABC",
				AccountNumber:       "1234567890",
				AccountName:         "Test Account",
			},
			wantErr: false,
		},
		{
			name: "ERROR: Vendor name already exists",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("Error 1062 (23000): Duplicate entry 'merchant-123-Existing Vendor' for key 'vendors.vendors_merchant_name_uniq_comp_idx'"))
			},
			input: &vendorModel.CreateVendorRequest{
				MerchantID:          "merchant-123",
				Name:                "Existing Vendor",
				BeneficialOwner:     "John Doe",
				BusinessCategory:    "E-Commerce",
				AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
				BankName:            "Bank ABC",
				BankCode:            "ABC",
				AccountNumber:       "1234567890",
				AccountName:         "Test Account",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Repository Create failure",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			input: &vendorModel.CreateVendorRequest{
				MerchantID:          "merchant-123",
				Name:                "Test Vendor",
				BeneficialOwner:     "John Doe",
				BusinessCategory:    "E-Commerce",
				AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
				BankName:            "Bank ABC",
				BankCode:            "ABC",
				AccountNumber:       "1234567890",
				AccountName:         "Test Account",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocksRepo.NewIVendorRepository(t)
			logger, _ := logger.NewZapLogger(logger.Config{})

			if tc.setup != nil {
				tc.setup(repo)
			}

			svc := New(repo, logger)

			got, err := svc.Create(context.Background(), tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				assert.NotNil(t, got)
				assert.Equal(t, tc.input.MerchantID, got.MerchantID)
				assert.Equal(t, tc.input.Name, got.Name)
				assert.Equal(t, tc.input.BeneficialOwner, got.BeneficialOwner)
				assert.Equal(t, tc.input.BusinessCategory, got.BusinessCategory)
				assert.Equal(t, tc.input.AvgMonthlyTpvAmount.String(), got.AvgMonthlyTpvAmount.String())
				assert.Equal(t, tc.input.BankName, got.BankName)
				assert.Equal(t, tc.input.BankCode, got.BankCode)
				assert.Equal(t, tc.input.AccountNumber, got.AccountNumber)
				assert.Equal(t, tc.input.AccountName, got.AccountName)
				assert.Equal(t, "ACTIVE", got.Status)
			}
		})
	}
}
