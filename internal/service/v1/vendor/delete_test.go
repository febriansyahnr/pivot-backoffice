package vendor

import (
	"context"
	"errors"
	"testing"

	vendorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	existingVendor := &vendorModel.Vendor{
		UUID:                "test-uuid",
		MerchantID:          "merchant-123",
		Name:                "Test Vendor",
		BeneficialOwner:     "John Doe",
		BusinessCategory:    "E-Commerce",
		AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
		BankName:            "Bank ABC",
		BankCode:            "ABC",
		AccountNumber:       "1234567890",
		AccountName:         "Test Account",
		Status:              "ACTIVE",
	}

	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IVendorRepository)
		uuid    string
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "test-uuid").Return(existingVendor, nil)
				repo.On("Delete", mock.Anything, "test-uuid").Return(nil)
			},
			uuid:    "test-uuid",
			wantErr: false,
		},
		{
			name: "ERROR: Vendor not found",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "non-existent-uuid").Return(nil, nil)
			},
			uuid:    "non-existent-uuid",
			wantErr: true,
		},
		{
			name: "ERROR: GetByID database error",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "test-uuid").Return(nil, errors.New("database error"))
			},
			uuid:    "test-uuid",
			wantErr: true,
		},
		{
			name: "ERROR: Delete database error",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "test-uuid").Return(existingVendor, nil)
				repo.On("Delete", mock.Anything, "test-uuid").Return(errors.New("delete error"))
			},
			uuid:    "test-uuid",
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

			err := svc.Delete(context.Background(), tc.uuid)

			if (err != nil) != tc.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
