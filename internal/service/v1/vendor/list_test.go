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

func TestList(t *testing.T) {
	existingVendors := []*vendorModel.Vendor{
		{
			UUID:                "test-uuid-1",
			MerchantID:          "merchant-123",
			Name:                "Test Vendor 1",
			BeneficialOwner:     "John Doe",
			BusinessCategory:    "E-Commerce",
			AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
			BankName:            "Bank ABC",
			BankCode:            "ABC",
			AccountNumber:       "1234567890",
			AccountName:         "Test Account 1",
			Status:              "ACTIVE",
		},
		{
			UUID:                "test-uuid-2",
			MerchantID:          "merchant-123",
			Name:                "Test Vendor 2",
			BeneficialOwner:     "Jane Doe",
			BusinessCategory:    "Retail",
			AvgMonthlyTpvAmount: decimal.NewFromInt(2000000),
			BankName:            "Bank XYZ",
			BankCode:            "XYZ",
			AccountNumber:       "0987654321",
			AccountName:         "Test Account 2",
			Status:              "ACTIVE",
		},
	}

	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IVendorRepository)
		query   *vendorModel.VendorQuery
		wantErr bool
	}{
		{
			name: "SUCCESS: List all vendors",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("List", mock.Anything, mock.Anything).Return(existingVendors, 2, nil)
			},
			query: &vendorModel.VendorQuery{
				Page:     1,
				PageSize: 10,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: List vendors with filter",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("List", mock.Anything, mock.Anything).Return([]*vendorModel.Vendor{existingVendors[0]}, 1, nil)
			},
			query: &vendorModel.VendorQuery{
				Name:     "Vendor 1",
				Page:     1,
				PageSize: 10,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: List empty vendors",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("List", mock.Anything, mock.Anything).Return([]*vendorModel.Vendor{}, 0, nil)
			},
			query: &vendorModel.VendorQuery{
				Page:     1,
				PageSize: 10,
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("List", mock.Anything, mock.Anything).Return(nil, 0, errors.New("database error"))
			},
			query: &vendorModel.VendorQuery{
				Page:     1,
				PageSize: 10,
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

			got, err := svc.List(context.Background(), tc.query)

			if (err != nil) != tc.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				assert.NotNil(t, got)
				assert.NotNil(t, got.Data)
				assert.NotNil(t, got.Meta)
			}
		})
	}
}
