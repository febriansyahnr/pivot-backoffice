package recurringContractService_test

import (
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/recurringContract"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetRecurringByID(t *testing.T) {
	log := logMock.NewILogger(t)
	repo := repoMocks.NewIRecurringContractRepository(t)

	service := New(log, repo, nil)

	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	recurringID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
	customerID := "3bc93f16-93d8-4c2c-a0f2-27c48887617c"

	request := model.GetRecurringByIDRequest{
		MerchantID:  merchantID,
		RecurringID: recurringID,
	}

	validRecurringContract := &model.RecurringContractDetail{
		UUID:              recurringID,
		MerchantID:        merchantID,
		CustomerID:        customerID,
		ClientReferenceID: "client-ref-123",
		Plan: model.Plan{
			PlanId:   "plan-1",
			PlanName: "Premium Plan",
		},
		Trials: []model.Trial{
			{
				TrialStart: 1,
				TrialEnd:   3,
				Type:       constant.RecurringContractTrialTypePercentage,
				Percentage: 50,
			},
		},
		Billing: model.Billing{
			Interval:     1,
			IntervalUnit: "MONTH",
			Count:        12,
		},
		Currency:  "IDR",
		Amount:    100000.00,
		Status:    constant.RecurringContractStatusActive,
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name       string
		request    model.GetRecurringByIDRequest
		setupMock  func()
		wantError  error
		wantResult *model.GetRecurringByIDDashboardResponse
	}{
		{
			name:    "ERROR:Get detail by ID returns error",
			request: request,
			setupMock: func() {
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "error when get recurring contract detail by ID", mock.Anything).Once().Return()
			},
			wantError:  assert.AnError,
			wantResult: nil,
		},
		{
			name:    "ERROR:Recurring contract not found",
			request: request,
			setupMock: func() {
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(nil, nil)
				log.On("Error", mock.Anything, "recurring contract not found", mock.Anything).Once().Return()
			},
			wantError:  pkgErrs.New(response.HttpErrNotFound, errors.New("recurring contract not found")),
			wantResult: nil,
		},
		{
			name:    "SUCCESS:Get recurring by ID",
			request: request,
			setupMock: func() {
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(validRecurringContract, nil)
			},
			wantError: nil,
			wantResult: &model.GetRecurringByIDDashboardResponse{
				UUID:              recurringID,
				MerchantID:        merchantID,
				CustomerID:        customerID,
				ClientReferenceID: "client-ref-123",
				Plan: model.Plan{
					PlanId:   "plan-1",
					PlanName: "Premium Plan",
				},
				Trials: []model.Trial{
					{
						TrialStart: 1,
						TrialEnd:   3,
						Type:       constant.RecurringContractTrialTypePercentage,
						Percentage: 50,
					},
				},
				Billing: model.Billing{
					Interval:     1,
					IntervalUnit: "MONTH",
					Count:        12,
				},
				Amount: commonModel.Amount{
					Currency: "IDR",
					Value:    "100000.00",
				},
				Status:    constant.RecurringContractStatusActive,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetRecurringByID(t.Context(), test.request)

			require.Equal(t, test.wantError, err)
			if test.wantResult != nil {
				require.NotNil(t, result)
				assert.Equal(t, test.wantResult.UUID, result.UUID)
				assert.Equal(t, test.wantResult.MerchantID, result.MerchantID)
				assert.Equal(t, test.wantResult.CustomerID, result.CustomerID)
				assert.Equal(t, test.wantResult.ClientReferenceID, result.ClientReferenceID)
				assert.Equal(t, test.wantResult.Plan, result.Plan)
				assert.Equal(t, test.wantResult.Trials, result.Trials)
				assert.Equal(t, test.wantResult.Billing, result.Billing)
				assert.Equal(t, test.wantResult.Amount, result.Amount)
				assert.Equal(t, test.wantResult.Status, result.Status)
				assert.Equal(t, test.wantResult.CreatedAt, result.CreatedAt)
				assert.Equal(t, test.wantResult.UpdatedAt, result.UpdatedAt)
			} else {
				require.Nil(t, result)
			}

			log.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}
