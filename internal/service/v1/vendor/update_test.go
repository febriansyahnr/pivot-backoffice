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

func TestUpdate(t *testing.T) {
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

	newName := "Updated Vendor"

	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IVendorRepository)
		input   *vendorModel.UpdateVendorRequest
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "test-uuid").Return(existingVendor, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			input: &vendorModel.UpdateVendorRequest{
				UUID: "test-uuid",
				Name: &newName,
			},
			wantErr: false,
		},
		{
			name: "ERROR: Vendor not found",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "non-existent-uuid").Return(nil, nil)
			},
			input: &vendorModel.UpdateVendorRequest{
				UUID: "non-existent-uuid",
				Name: &newName,
			},
			wantErr: true,
		},
		{
			name: "ERROR: GetByID database error",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "test-uuid").Return(nil, errors.New("database error"))
			},
			input: &vendorModel.UpdateVendorRequest{
				UUID: "test-uuid",
				Name: &newName,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Update database error",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "test-uuid").Return(existingVendor, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))
			},
			input: &vendorModel.UpdateVendorRequest{
				UUID: "test-uuid",
				Name: &newName,
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

			got, err := svc.Update(context.Background(), tc.input)

			if (err != nil) != tc.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				assert.NotNil(t, got)
			}
		})
	}
}
