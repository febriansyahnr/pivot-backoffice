package recurringContractService_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/recurringContract"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateRecurringPayment(t *testing.T) {
	log := logMock.NewILogger(t)
	repo := repoMocks.NewIRecurringContractRepository(t)

	service := New(log, repo, nil)

	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	recurringID := "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
	transactionID := "trx-123"    // NOSONAR
	paymentTokenID := "token-456" // NOSONAR
	paymentMethodID := "pm-789"   // NOSONAR
	updatedBy := "3bc93f16-93d8-4c2c-a0f2-27c48887617c"

	request := model.UpdateRecurringPaymentRequest{
		MerchantID:      merchantID,
		RecurringID:     recurringID,
		TransactionID:   transactionID,
		PaymentTokenID:  paymentTokenID,
		PaymentMethodID: paymentMethodID,
		RecurringPayment: &unifiedPaymentModel.MetadataRecurringPayment{
			InitiateFirstAuthorization: false,
			BillingCycle: unifiedPaymentModel.RecurringBillingCycle{
				Count: 1,
			},
		},
		UpdatedBy: updatedBy,
	}

	activeRecurringContract := &model.RecurringContractDetail{
		UUID:   recurringID,
		Status: constant.RecurringContractStatusActive,
	}

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Get detail by id returns database error",
			setupMock: func() {
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to retrieve recurring payment contract details", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Recurring contract not found",
			setupMock: func() {
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(nil, nil)
			},
			wantError: constant.NewErrResourceNotFound("recurring payment contract", recurringID),
		},
		{
			name: "ERROR:Update recurring contract fails with database error",
			setupMock: func() {
				request.RecurringPayment.InitiateFirstAuthorization = true

				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(activeRecurringContract, nil)
				repo.On("UpdateRecurringContract", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed to update the recurring payment contract", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Update recurring contract returns no rows affected",
			setupMock: func() {
				request.RecurringPayment.InitiateFirstAuthorization = false

				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(activeRecurringContract, nil)
				repo.On("UpdateRecurringContract", mock.Anything, mock.Anything).Once().Return(constant.ErrNoRowsAffected)
			},
			wantError: pkgErrs.New(response.HttpErrNotFound, constant.ErrNoRowsAffected),
		},
		{
			name: "SUCCESS:Update without initiating first authorization, contract not in first auth state",
			setupMock: func() {
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(activeRecurringContract, nil)
				repo.On(
					"UpdateRecurringContract", mock.Anything, mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					data := args.Get(1).(model.UpdateRecurringContractRequest)
					require.Equal(t, data.RecurringID, recurringID)
					require.Equal(t, data.BillingCycleCount, uint16(1))
					require.Empty(t, data.TransactionID)
					require.Empty(t, data.PaymentTokenID)
					require.Empty(t, data.PaymentMethodID)
					require.Empty(t, data.UpdatedBy)
					require.Empty(t, data.Status)
					require.True(t, data.UpdatedAt.IsZero())
					require.True(t, data.ActivatedAt.IsZero())
				}).Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Update with initiating first authorization, contract not in first auth state",
			setupMock: func() {
				request.RecurringPayment.InitiateFirstAuthorization = true

				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(activeRecurringContract, nil)
				repo.On(
					"UpdateRecurringContract", mock.Anything, mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					data := args.Get(1).(model.UpdateRecurringContractRequest)
					require.Equal(t, data.RecurringID, recurringID)
					require.Equal(t, data.BillingCycleCount, uint16(1))
					require.Equal(t, data.TransactionID, transactionID)
					require.Equal(t, data.PaymentTokenID, paymentTokenID)
					require.Equal(t, data.PaymentMethodID, paymentMethodID)
					require.Equal(t, data.UpdatedBy, updatedBy)
					require.Empty(t, data.Status)
					require.False(t, data.UpdatedAt.IsZero())
					require.True(t, data.ActivatedAt.IsZero())
				}).Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Update with initiating first authorization, contract is in first auth state (should activate)",
			setupMock: func() {
				request.RecurringPayment.BillingCycle.Count = 0

				recurringContract := &model.RecurringContractDetail{
					UUID:   recurringID,
					Status: constant.RecurringContractStatusCreated,
				}
				repo.On(
					"GetDetailByID", mock.Anything, merchantID, recurringID,
				).Once().Return(recurringContract, nil)
				repo.On(
					"UpdateRecurringContract", mock.Anything, mock.Anything,
				).Once().Run(func(args mock.Arguments) {
					data := args.Get(1).(model.UpdateRecurringContractRequest)
					require.Equal(t, data.RecurringID, recurringID)
					require.Equal(t, data.BillingCycleCount, uint16(0))
					require.Equal(t, data.TransactionID, transactionID)
					require.Equal(t, data.PaymentTokenID, paymentTokenID)
					require.Equal(t, data.PaymentMethodID, paymentMethodID)
					require.Equal(t, data.UpdatedBy, updatedBy)
					require.Equal(t, data.Status, constant.RecurringContractStatusActive)
					require.False(t, data.UpdatedAt.IsZero())
					require.False(t, data.ActivatedAt.IsZero())
				}).Return(nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := service.UpdateRecurringPayment(t.Context(), request)

			require.Equal(t, test.wantError, err)

			log.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}
