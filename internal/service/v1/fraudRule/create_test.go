package fraudruleservice

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IFraudRulesRepository)
		input   *fraudrulesmodel.CreateFraudRuleRequest
		want    *fraudrulesmodel.FraudRulesResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			input: &fraudrulesmodel.CreateFraudRuleRequest{
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      1,
				Weight:        decimal.NewFromInt(50),
				IsActive:      true,
				Provider:      sql.NullString{String: "provider-name", Valid: true},
				ReferenceType: "payment",
			},
			want: &fraudrulesmodel.FraudRulesResponse{
				UUID:          mock.Anything,
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      1,
				Weight:        decimal.NewFromInt(50),
				IsActive:      true,
				Provider:      "provider-name",
				ReferenceType: "payment",
			},
			wantErr: false,
		},
		{
			name: "ERROR: Invalid request",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("validation error"))
			},
			input: &fraudrulesmodel.CreateFraudRuleRequest{
				// Missing required fields like RuleName and Weight
				Condition:     "amount > 1000",
				Priority:      1,
				IsActive:      true,
				ReferenceType: "payment",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Repository failure",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			input: &fraudrulesmodel.CreateFraudRuleRequest{
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      1,
				Weight:        decimal.NewFromInt(50),
				IsActive:      true,
				Provider:      sql.NullString{String: "provider-name", Valid: true},
				ReferenceType: "payment",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			repo := mocksRepo.NewIFraudRulesRepository(t)
			logger, _ := logger.NewZapLogger(logger.Config{})

			// Setup expectations
			if tc.setup != nil {
				tc.setup(repo)
			}

			// Create service
			svc := New(logger, repo)

			// Call method
			got, err := svc.Create(context.Background(), tc.input)

			// Verify results
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// If we expect success, verify the returned fraud rule
			if !tc.wantErr {
				assert.NotNil(t, got)
				assert.Equal(t, tc.input.RuleName, got.RuleName)
				assert.Equal(t, tc.input.Condition, got.Condition)
				assert.Equal(t, tc.input.Priority, got.Priority)
				assert.Equal(t, tc.input.Weight.String(), got.Weight.String())
				assert.Equal(t, tc.input.IsActive, got.IsActive)
				assert.Equal(t, tc.input.Provider.String, got.Provider)
				assert.Equal(t, tc.input.ReferenceType, got.ReferenceType)
			}
		})
	}
}
