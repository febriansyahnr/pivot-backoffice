package installmentPlan

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInstallmentPlanRepository_Create(t *testing.T) {
	now := time.Now().UTC()
	metadata := types.JSONText(`{"card":{"midId":"test-mid","allowedBins":["123456"],"interest":2.5}}`)

	tests := []struct {
		name      string
		plan      *installmentPlanModel.InstallmentPlan
		mockSetup func(*mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "success",
			plan: &installmentPlanModel.InstallmentPlan{
				UUID:           "test-uuid",
				MerchantID:     "test-merchant",
				Acquirer:       "test-acquirer",
				SettlementType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
				Title:          "Test Plan",
				Description:    "Test Description",
				Tenor:          12,
				Status:         constant.InstallmentPlanStatusActive,
				Metadata:       metadata,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			plan: &installmentPlanModel.InstallmentPlan{
				UUID:           "test-uuid",
				MerchantID:     "test-merchant",
				Acquirer:       "test-acquirer",
				SettlementType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
				Title:          "Test Plan",
				Description:    "Test Description",
				Tenor:          12,
				Status:         constant.InstallmentPlanStatusActive,
				Metadata:       metadata,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				dbErr := errors.New("database connection failed")
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, dbErr)
			},
			wantErr: true,
		},
		{
			name: "empty uuid",
			plan: &installmentPlanModel.InstallmentPlan{
				UUID:           "",
				MerchantID:     "test-merchant",
				Acquirer:       "test-acquirer",
				SettlementType: constant.PaymentMethodChannelTypeAggregator,
				PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
				Title:          "Test Plan",
				Description:    "Test Description",
				Tenor:          12,
				Status:         constant.InstallmentPlanStatusActive,
				Metadata:       metadata,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return(false, nil)
			},
			wantErr: false,
		},
		{
			name: "minimal required fields",
			plan: &installmentPlanModel.InstallmentPlan{
				UUID:           "minimal-uuid",
				MerchantID:     "minimal-merchant",
				Acquirer:       "minimal-acquirer",
				SettlementType: "DIRECT",
				PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
				Title:          "Minimal",
				Description:    "Minimal",
				Tenor:          1,
				Status:         constant.InstallmentPlanStatusActive,
				Metadata:       types.JSONText(`{}`),
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tt.mockSetup(mockDB)

			repo := &InstallmentPlanRepository{
				db:     mockDB,
				logger: mockLogger,
			}

			err := repo.Create(context.Background(), tt.plan)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
