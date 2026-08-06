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

func TestInstallmentPlanRepository_Update(t *testing.T) {
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
				SettlementType: "SETTLEMENT",
				PaymentMethod:  "CARD",
				Title:          "Updated Plan",
				Description:    "Updated Description",
				Tenor:          24,
				Status:         constant.InstallmentPlanStatusActive,
				Metadata:       metadata,
				CreatedAt:      now.Add(-time.Hour),
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
				SettlementType: "SETTLEMENT",
				PaymentMethod:  "CARD",
				Title:          "Updated Plan",
				Description:    "Updated Description",
				Tenor:          24,
				Status:         constant.InstallmentPlanStatusActive,
				Metadata:       metadata,
				CreatedAt:      now.Add(-time.Hour),
				UpdatedAt:      now,
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				dbErr := errors.New("database update failed")
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
				SettlementType: "DIRECT",
				PaymentMethod:  "CARD",
				Title:          "No UUID Plan",
				Description:    "Plan without UUID",
				Tenor:          6,
				Status:         constant.InstallmentPlanStatusActive,
				Metadata:       metadata,
				CreatedAt:      now.Add(-time.Hour),
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
			name: "minimal fields",
			plan: &installmentPlanModel.InstallmentPlan{
				UUID:           "minimal-uuid",
				MerchantID:     "minimal-merchant",
				Acquirer:       "minimal-acquirer",
				SettlementType: "DIRECT",
				PaymentMethod:  "CARD",
				Title:          "Min",
				Description:    "Min",
				Tenor:          1,
				Status:         constant.InstallmentPlanStatusActive,
				Metadata:       types.JSONText(`{}`),
				CreatedAt:      now.Add(-time.Hour),
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

			err := repo.Update(context.Background(), tt.plan)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
