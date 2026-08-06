package fraudruleservice

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	customErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	now := time.Now().UTC()
	ruleUUID := "test-uuid"
	weight := decimal.NewFromInt(80)
	provider := "provider-abc"

	existingRule := &fraudrulesmodel.FraudRules{
		UUID:          ruleUUID,
		RuleName:      "Old Rule",
		Condition:     "amount < 500",
		Priority:      1,
		Weight:        decimal.NewFromInt(20),
		IsActive:      false,
		Provider:      sql.NullString{Valid: false},
		ReferenceType: "invoice",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	tests := []struct {
		name      string
		setup     func(repo *mocksRepo.IFraudRulesRepository)
		input     *fraudrulesmodel.UpdateFraudRuleRequest
		wantErr   bool
		errorType error
	}{
		{
			name: "SUCCESS: Update fraud rule",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				repo.On("GetByID", mock.Anything, ruleUUID).Return(existingRule, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			input: &fraudrulesmodel.UpdateFraudRuleRequest{
				UUID:          ruleUUID,
				RuleName:      util.ValueToPtr("Updated Rule"),
				Condition:     util.ValueToPtr("amount > 1000"),
				Priority:      util.ValueToPtr(5),
				Weight:        &weight,
				IsActive:      util.ValueToPtr(true),
				Provider:      &provider,
				ReferenceType: util.ValueToPtr("payment"),
			},
			wantErr: false,
		},
		{
			name: "ERROR: Rule not found",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				repo.On("GetByID", mock.Anything, ruleUUID).Return(nil, errors.New("not found"))
			},
			input: &fraudrulesmodel.UpdateFraudRuleRequest{
				UUID: ruleUUID,
			},
			wantErr:   true,
			errorType: customErr.New(response.HttpErrNotFound, errors.New("not found")),
		},
		{
			name: "ERROR: Repository update failure",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				repo.On("GetByID", mock.Anything, ruleUUID).Return(existingRule, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			input: &fraudrulesmodel.UpdateFraudRuleRequest{
				UUID:     ruleUUID,
				RuleName: util.ValueToPtr("Another Rule"),
			},
			wantErr:   true,
			errorType: customErr.New(response.HttpErrInternal, errors.New("update failed")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocksRepo.NewIFraudRulesRepository(t)
			log, _ := logger.NewZapLogger(logger.Config{})
			if tt.setup != nil {
				tt.setup(repo)
			}
			svc := New(log, repo)

			res, err := svc.Update(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.input.UUID, res.UUID)
				assert.Equal(t, *tt.input.RuleName, res.RuleName)
				assert.Equal(t, *tt.input.Condition, res.Condition)
				assert.Equal(t, *tt.input.Priority, res.Priority)
				assert.Equal(t, tt.input.Weight.String(), res.Weight.String())
				assert.Equal(t, *tt.input.IsActive, res.IsActive)
				assert.Equal(t, *tt.input.Provider, res.Provider)
				assert.Equal(t, *tt.input.ReferenceType, res.ReferenceType)
			}
		})
	}
}
