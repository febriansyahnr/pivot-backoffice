package fraudruleservice

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDetail(t *testing.T) {
	// Create a sample fraud rule that will be returned by the mock repository
	sampleUUID := "test-uuid-123"
	sampleTime := time.Now().UTC()
	sampleFraudRule := &fraudrulesmodel.FraudRules{
		UUID:          sampleUUID,
		RuleName:      "Test Rule",
		Condition:     "amount > 1000",
		Priority:      1,
		Weight:        decimal.NewFromInt(50),
		IsActive:      true,
		Provider:      sql.NullString{String: "provider-name", Valid: true},
		ReferenceType: "payment",
		CreatedAt:     sampleTime,
		UpdatedAt:     sampleTime,
		DeletedAt:     sql.NullTime{Valid: false},
	}

	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IFraudRulesRepository)
		uuid    string
		want    *fraudrulesmodel.FraudRules
		wantErr error
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				repo.On("GetByID", mock.Anything, sampleUUID).Return(sampleFraudRule, nil)
			},
			uuid:    sampleUUID,
			want:    sampleFraudRule,
			wantErr: nil,
		},
		{
			name: "ERROR: Fraud rule not found",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				repo.On("GetByID", mock.Anything, "non-existent-uuid").Return(nil, nil)
			},
			uuid:    "non-existent-uuid",
			want:    nil,
			wantErr: errors.New(response.HttpErrRequest, constant.ErrFraudRulesNotFound),
		},
		{
			name: "ERROR: Repository failure",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				repo.On("GetByID", mock.Anything, "error-uuid").Return(nil,
					errors.New(response.HttpErrInternal, constant.ErrGetFraudRuleDetail))
			},
			uuid:    "error-uuid",
			want:    nil,
			wantErr: errors.New(response.HttpErrInternal, constant.ErrGetFraudRuleDetail),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			repo := mocksRepo.NewIFraudRulesRepository(t)
			testLogger, _ := logger.NewZapLogger(logger.Config{})

			// Setup expectations
			if tc.setup != nil {
				tc.setup(repo)
			}

			// Create service
			svc := New(testLogger, repo)

			// Call method
			got, err := svc.Detail(context.Background(), tc.uuid)

			// Verify results
			if tc.wantErr != nil {
				assert.Error(t, err)
				// Since errors.New returns a new error instance each time,
				// we need to compare error messages instead of the error itself
				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			// Verify expectations were met
			repo.AssertExpectations(t)
		})
	}
}
