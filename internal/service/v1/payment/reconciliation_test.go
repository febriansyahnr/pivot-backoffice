package paymentService_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessInvestigationMonthlyReconciliation(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	ledgerSvc := serviceMocks.NewIOrchestratorService(t)

	service := New(paymentRepo, logger, nil, nil, nil, nil, nil, WithOrchestratorService(ledgerSvc))

	ctxTrx := context.WithValue(t.Context(), pdkConst.CtxSqlTx, struct{}{})

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Calculate investigation monthly reconciliation",
			setupMock: func() {
				paymentRepo.On(
					"CalculateInvestigationMonthlyReconciliation", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
				logger.On(
					"Error", mock.Anything, "Failed to calculate investigation monthly reconciliation", mock.Anything,
				).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, assert.AnError),
		},
		{
			name: "SUCCESS:No transactions occurred",
			setupMock: func() {
				paymentRepo.On(
					"CalculateInvestigationMonthlyReconciliation", mock.Anything, mock.Anything,
				).Once().Return(nil, nil)
				logger.On(
					"Info", mock.Anything, "There are no transactions resulting from the investigation that need to be reconciled",
				).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Begin transaction",
			setupMock: func() {
				paymentRepo.On(
					"CalculateInvestigationMonthlyReconciliation", mock.Anything, mock.Anything,
				).Return([]model.CalculateInvestigationMonthlyReconciliation{
					{MerchantID: "6baf8fb5-4a3a-4f72-bb48-4df2f7123152", GrossAmount: 10_000, NetAmount: 10_000, MerchantLossAmount: 10_000},
				}, nil)
				paymentRepo.On("BeginTransaction", mock.Anything).Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to process reconciliation from payment investigation", mock.Anything, mock.Anything).Return()
				logger.On("Info", mock.Anything, "Process monthly reconciliation completed", pdkLog.Any("summaries", map[string]any{"total": 1, "success": 0, "failed": 1})).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Insert investigation monthly reconciliation",
			setupMock: func() {
				paymentRepo.On("BeginTransaction", mock.Anything).Return(ctxTrx, nil)
				paymentRepo.On("InsertInvestigationMonthlyReconciliation", ctxTrx, mock.Anything).Once().Return(assert.AnError)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(constant.ErrSomeErrorForUnitTest)
				logger.On("Error", mock.Anything, "Failed to rollback transaction", pdkLog.Error(constant.ErrSomeErrorForUnitTest)).Once().Return()
				logger.On("Info", mock.Anything, "Process monthly reconciliation completed", pdkLog.Any("summaries", map[string]any{"total": 1, "success": 0, "failed": 1})).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Update payment investigation reconciliation",
			setupMock: func() {
				paymentRepo.On("InsertInvestigationMonthlyReconciliation", ctxTrx, mock.Anything).Return(nil)
				paymentRepo.On("UpdatePaymentInvestigationReconciliation", ctxTrx, mock.Anything).Once().Return(assert.AnError)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(nil)
				logger.On("Info", mock.Anything, "Process monthly reconciliation completed", pdkLog.Any("summaries", map[string]any{"total": 1, "success": 0, "failed": 1})).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Create account transaction",
			setupMock: func() {
				paymentRepo.On("UpdatePaymentInvestigationReconciliation", ctxTrx, mock.Anything).Return(nil)
				ledgerSvc.On("CreateAccountTransaction", ctxTrx, mock.Anything).Once().Return(assert.AnError)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(nil)
				logger.On("Info", mock.Anything, "Process monthly reconciliation completed", pdkLog.Any("summaries", map[string]any{"total": 1, "success": 0, "failed": 1})).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Commit transaction",
			setupMock: func() {
				ledgerSvc.On("CreateAccountTransaction", ctxTrx, mock.Anything).Return(nil)
				paymentRepo.On("CommitTransaction", ctxTrx).Once().Return(assert.AnError)
				paymentRepo.On("RollbackTransaction", ctxTrx).Once().Return(nil)
				logger.On("Info", mock.Anything, "Process monthly reconciliation completed", pdkLog.Any("summaries", map[string]any{"total": 1, "success": 0, "failed": 1})).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				ledgerSvc.On("CreateAccountTransaction", ctxTrx, mock.Anything).Return(nil)
				paymentRepo.On("CommitTransaction", ctxTrx).Once().Return(nil)
				logger.On("Info", mock.Anything, "Process monthly reconciliation completed", pdkLog.Any("summaries", map[string]any{"total": 1, "success": 1, "failed": 0})).Once().Return()
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, service.ProcessInvestigationMonthlyReconciliation(t.Context(), model.MonthlyReconciliationRequest{}))

			logger.AssertExpectations(t)
			ledgerSvc.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
		})
	}
}
