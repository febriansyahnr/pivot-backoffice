package fraudruleservice

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	fraudrulesmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/fraudRules"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestList(t *testing.T) {
	// Create sample time for testing
	sampleTime := time.Now().UTC()

	// Create sample fraud rules that will be returned by the mock repository
	sampleFraudRules := []*fraudrulesmodel.FraudRules{
		{
			UUID:          "test-uuid-1",
			RuleName:      "Test Rule 1",
			Condition:     "amount > 1000",
			Priority:      1,
			Weight:        decimal.NewFromInt(50),
			IsActive:      true,
			Provider:      sql.NullString{String: "provider-1", Valid: true},
			ReferenceType: "payment",
			CreatedAt:     sampleTime,
			UpdatedAt:     sampleTime,
			DeletedAt:     sql.NullTime{Valid: false},
		},
		{
			UUID:          "test-uuid-2",
			RuleName:      "Test Rule 2",
			Condition:     "amount < 500",
			Priority:      2,
			Weight:        decimal.NewFromInt(30),
			IsActive:      false,
			Provider:      sql.NullString{String: "provider-2", Valid: true},
			ReferenceType: "transfer",
			CreatedAt:     sampleTime,
			UpdatedAt:     sampleTime,
			DeletedAt:     sql.NullTime{Valid: false},
		},
	}

	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IFraudRulesRepository)
		query   *fraudrulesmodel.FraudRulesQuery
		want    *commonModel.PaginationResponse
		wantErr error
	}{
		{
			name: "SUCCESS: List with paging",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				query := &fraudrulesmodel.FraudRulesQuery{
					Page:     1,
					PageSize: 10,
				}
				repo.On("List", mock.Anything, query).Return(sampleFraudRules, 2, nil)
			},
			query: &fraudrulesmodel.FraudRulesQuery{
				Page:     1,
				PageSize: 10,
			},
			want: &commonModel.PaginationResponse{
				Data: sampleFraudRules,
				Meta: commonModel.Meta{
					Page:       1,
					PerPage:    10,
					TotalItems: 2,
					TotalPages: 1,
				},
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: Empty list",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				query := &fraudrulesmodel.FraudRulesQuery{
					Page:          1,
					PageSize:      10,
					ReferenceType: "nonexistent",
				}
				repo.On("List", mock.Anything, query).Return([]*fraudrulesmodel.FraudRules{}, 0, nil)
			},
			query: &fraudrulesmodel.FraudRulesQuery{
				Page:          1,
				PageSize:      10,
				ReferenceType: "nonexistent",
			},
			want: &commonModel.PaginationResponse{
				Data: []*fraudrulesmodel.FraudRules{},
				Meta: commonModel.Meta{
					Page:       1,
					PerPage:    10,
					TotalItems: 0,
					TotalPages: 0,
				},
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: Filter by rule name",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				query := &fraudrulesmodel.FraudRulesQuery{
					Page:     1,
					PageSize: 10,
					RuleName: "Test Rule 1",
				}
				repo.On("List", mock.Anything, query).Return([]*fraudrulesmodel.FraudRules{sampleFraudRules[0]}, 1, nil)
			},
			query: &fraudrulesmodel.FraudRulesQuery{
				Page:     1,
				PageSize: 10,
				RuleName: "Test Rule 1",
			},
			want: &commonModel.PaginationResponse{
				Data: []*fraudrulesmodel.FraudRules{sampleFraudRules[0]},
				Meta: commonModel.Meta{
					Page:       1,
					PerPage:    10,
					TotalItems: 1,
					TotalPages: 1,
				},
			},
			wantErr: nil,
		},
		{
			name: "ERROR: Repository failure",
			setup: func(repo *mocksRepo.IFraudRulesRepository) {
				query := &fraudrulesmodel.FraudRulesQuery{
					Page:     1,
					PageSize: 10,
				}
				repo.On("List", mock.Anything, query).Return(nil, 0,
					errors.New(response.HttpErrInternal, constant.ErrGetFraudRuleDetail))
			},
			query: &fraudrulesmodel.FraudRulesQuery{
				Page:     1,
				PageSize: 10,
			},
			want:    nil,
			wantErr: errors.New(response.HttpErrInternal, constant.ErrGetIPWhitelistConfigurationList),
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
			got, err := svc.List(context.Background(), tc.query)

			// Verify results
			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.want.Meta.Page, got.Meta.Page)
				assert.Equal(t, tc.want.Meta.PerPage, got.Meta.PerPage)
				assert.Equal(t, tc.want.Meta.TotalItems, got.Meta.TotalItems)
				assert.Equal(t, tc.want.Meta.TotalPages, got.Meta.TotalPages)

				// Compare data - no need for type conversion since Data is already an interface{}
				gotRules := got.Data.([]*fraudrulesmodel.FraudRules)
				wantRules := tc.want.Data.([]*fraudrulesmodel.FraudRules)
				assert.Equal(t, len(wantRules), len(gotRules))

				// For non-empty results, verify the first item's content
				if len(wantRules) > 0 {
					assert.Equal(t, wantRules[0].UUID, gotRules[0].UUID)
					assert.Equal(t, wantRules[0].RuleName, gotRules[0].RuleName)
					assert.Equal(t, wantRules[0].Weight.String(), gotRules[0].Weight.String())
				}
			}

			// Verify expectations were met
			repo.AssertExpectations(t)
		})
	}
}
