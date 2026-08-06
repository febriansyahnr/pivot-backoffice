package recurringContractService_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
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

func TestCancel(t *testing.T) {
	log := logMock.NewILogger(t)
	repo := repoMocks.NewIRecurringContractRepository(t)

	service := New(log, repo, nil)

	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	recurringID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
	updatedBy := "3bc93f16-93d8-4c2c-a0f2-27c48887617c"

	request := model.CancelRecurringContractRequest{
		MerchantID:  merchantID,
		RecurringID: recurringID,
		UpdatedBy:   updatedBy,
	}

	tests := []struct {
		name      string
		request   model.CancelRecurringContractRequest
		setupMock func()
		wantError error
	}{
		{
			name:    "ERROR:Get detail by id returns database error",
			request: request,
			setupMock: func() {
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to retrieve recurring payment contract details", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "ERROR:Recurring contract not found",
			request: request,
			setupMock: func() {
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(nil, nil)
			},
			wantError: constant.NewErrResourceNotFound("recurring payment contract", recurringID),
		},
		{
			name:    "SUCCESS:Contract already inactive",
			request: request,
			setupMock: func() {
				recurringContract := &model.RecurringContractDetail{
					UUID:   recurringID,
					Status: constant.RecurringContractStatusInactive,
				}
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(recurringContract, nil)
			},
			wantError: nil,
		},
		{
			name:    "ERROR:Update recurring contract status fails",
			request: request,
			setupMock: func() {
				recurringContract := &model.RecurringContractDetail{
					UUID:   recurringID,
					Status: constant.RecurringContractStatusActive,
				}
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Return(recurringContract, nil)
				repo.On(
					"UpdateRecurringContractStatus", mock.Anything, recurringID, constant.RecurringContractStatusInactive, updatedBy,
				).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed to update the recurring payment contract status", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name:    "SUCCESS:Active contract cancelled",
			request: request,
			setupMock: func() {
				repo.On(
					"UpdateRecurringContractStatus", mock.Anything, recurringID, constant.RecurringContractStatusInactive, updatedBy,
				).Once().Return(nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			require.Equal(t, test.wantError, service.Cancel(t.Context(), test.request))

			log.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}
