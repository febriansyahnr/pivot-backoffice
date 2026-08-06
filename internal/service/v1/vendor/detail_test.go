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

func TestDetail(t *testing.T) {
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
		want    *vendorModel.Vendor
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "test-uuid").Return(existingVendor, nil)
			},
			uuid:    "test-uuid",
			want:    existingVendor,
			wantErr: false,
		},
		{
			name: "ERROR: Vendor not found",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "non-existent-uuid").Return(nil, nil)
			},
			uuid:    "non-existent-uuid",
			want:    nil,
			wantErr: true,
		},
		{
			name: "ERROR: Database error",
			setup: func(repo *mocksRepo.IVendorRepository) {
				repo.On("GetByID", mock.Anything, "test-uuid").Return(nil, errors.New("database error"))
			},
			uuid:    "test-uuid",
			want:    nil,
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

			got, err := svc.Detail(context.Background(), tc.uuid)

			if (err != nil) != tc.wantErr {
				t.Errorf("Detail() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !tc.wantErr {
				assert.NotNil(t, got)
				assert.Equal(t, tc.want.UUID, got.UUID)
				assert.Equal(t, tc.want.Name, got.Name)
			}
		})
	}
}
