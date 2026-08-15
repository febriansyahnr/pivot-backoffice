package disbursementService

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	common "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pdkLogMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExternalFDS(t *testing.T) {
	logger := pdkLogMock.NewILogger(t)
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	selfSvc := serviceMocks.NewIDisbursementInternalService(t)
	workflowFDSRepo := repositoryMocks.NewIWorkflowFDSRepository(t)
	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	accountTxRepo := repositoryMocks.NewIAccountTransactionRepository(t)

	conf := &config.Config{
		Environment: constant.EnvironmentTest,
	}

	service := &DisbursementService{
		config:                 conf,
		logger:                 logger,
		merchantRepo:           merchantRepo,
		disbursementRepo:       disbursementRepo,
		workflowFDSRepo:        workflowFDSRepo,
		accountTransactionRepo: accountTxRepo,
		self:                   selfSvc,
	}

	merchantID := uuid.NewString()
	payoutID := uuid.NewString()
	transactionID := uuid.NewString()

	validMerchant := &merchant.Merchant{
		UUID:      merchantID,
		Name:      "Test Merchant",                            // NOSONAR
		RiskLevel: sql.NullString{String: "LOW", Valid: true}, // NOSONAR
	}

	validPayout := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:                   payoutID,
			MerchantID:             merchantID,
			ReferenceID:            "REF00011110100293", // NOSONAR
			Amount:                 decimal.NewFromFloat(12_250),
			Currency:               "IDR",          // NOSONAR
			BeneficiaryBankCode:    "002",          // NOSONAR
			BeneficiaryAccountNo:   "029930495555", // NOSONAR
			BeneficiaryAccountName: "John Doe",     // NOSONAR
			CreatedAt:              time.Date(2026, 2, 3, 2, 34, 2, 0, time.UTC),
			UpdatedAt:              time.Date(2026, 2, 3, 2, 34, 2, 0, time.UTC),
			CreatedFrom:            &[]string{"OPEN_API"}[0], // NOSONAR
		},
	}

	validLedger := &orchestratorModel.TransactionAndFeeObject{
		TransactionID: transactionID,
		FeeID:         uuid.NewString(),
		MerchantID:    merchantID,
	}

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name:      "SUCCESS: External FDS ignored", // NOSONAR
			setupMock: func() { conf.Environment = constant.EnvironmentProduction },
			wantError: nil,
		},
		{
			name: "ERROR: Failed to fetch merchant data", // NOSONAR
			setupMock: func() {
				conf.Environment = constant.EnvironmentTest

				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID).Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to fetch merchant data", mock.Anything).Once().Return()
			},
			wantError: fmt.Errorf("%s:%s", "External FDS:", assert.AnError),
		},
		{
			name: "ERROR: Merchant not found", // NOSONAR
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID).Once().Return(nil, nil)
			},
			wantError: fmt.Errorf("%s:%s", "External FDS:", constant.ErrMerchantNotFound),
		},
		{
			name: "ERROR: FDS request error (HttpErrRequest)", // NOSONAR
			setupMock: func() {
				merchantRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(validMerchant, nil)

				requestErr := pkgErrs.New(response.HttpErrRequest, errors.New("bad request"))
				workflowFDSRepo.On("AssessPayoutTransaction", mock.Anything, mock.Anything).Once().Return(nil, requestErr)
				accountTxRepo.On("UpdateFDSRiskAssessmentResultByID", mock.Anything, transactionID, mock.Anything).Once().Return(nil)
				selfSvc.On("FailTransactionByFDSResult", mock.Anything, payoutID, validLedger).Once().Return(nil)
			},
			wantError: constant.ErrBlockedByFDS,
		},
		{
			name: "ERROR: FDS error but continue with ERROR result", // NOSONAR
			setupMock: func() {
				workflowFDSRepo.On(
					"AssessPayoutTransaction", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)

				accountTxRepo.On(
					"UpdateFDSRiskAssessmentResultByID", mock.Anything, transactionID, mock.Anything,
				).Once().Run(func(args mock.Arguments) {

					fdsResponse := args.Get(2).(*fdscommon.TransactionAssessmentResponse)
					require.Equal(t, constant.WorkflowFDSResultError, fdsResponse.Result)

				}).Return(nil)

				selfSvc.On(
					"FailTransactionByFDSResult", mock.Anything, payoutID, validLedger,
				).Once().Return(nil)
			},
			wantError: constant.ErrBlockedByFDS,
		},
		{
			name: "ERROR: Failed to update risk assessment", // NOSONAR
			setupMock: func() {
				fdsResponse := &fdscommon.TransactionAssessmentResponse{
					Result: constant.WorkflowFDSResultApprove,
				}
				workflowFDSRepo.On(
					"AssessPayoutTransaction", mock.Anything, mock.Anything,
				).Once().Return(fdsResponse, nil)

				accountTxRepo.On(
					"UpdateFDSRiskAssessmentResultByID", mock.Anything, transactionID, fdsResponse,
				).Once().Return(assert.AnError)
				logger.On("Error", mock.Anything, "Failed to update risk assessment", mock.Anything).Once().Return()
			},
			wantError: fmt.Errorf("%s:%s", "External FDS:", assert.AnError),
		},
		{
			name: "ERROR: FDS rejected and fail to update transaction status", // NOSONAR
			setupMock: func() {
				fdsResponse := &fdscommon.TransactionAssessmentResponse{
					Result: constant.WorkflowFDSResultReject,
				}

				workflowFDSRepo.On(
					"AssessPayoutTransaction", mock.Anything, mock.Anything,
				).Once().Return(fdsResponse, nil)
				accountTxRepo.On(
					"UpdateFDSRiskAssessmentResultByID", mock.Anything, transactionID, fdsResponse,
				).Once().Return(nil)
				selfSvc.On(
					"FailTransactionByFDSResult", mock.Anything, payoutID, validLedger,
				).Once().Return(assert.AnError)
				logger.On(
					"Error", mock.Anything, "Failed to update transaction status to blocked based on FDS risk assessment", mock.Anything,
				).Once().Return()
			},
			wantError: fmt.Errorf("%s:%s", "External FDS:", assert.AnError),
		},
		{
			name: "ERROR: FDS rejected - blocked by FDS", // NOSONAR
			setupMock: func() {

				fdsResponse := &fdscommon.TransactionAssessmentResponse{
					Result: constant.WorkflowFDSResultReject,
				}
				workflowFDSRepo.On(
					"AssessPayoutTransaction", mock.Anything, mock.Anything,
				).Once().Return(fdsResponse, nil)
				accountTxRepo.On(
					"UpdateFDSRiskAssessmentResultByID", mock.Anything, transactionID, fdsResponse,
				).Once().Return(nil)
				selfSvc.On(
					"FailTransactionByFDSResult", mock.Anything, payoutID, validLedger,
				).Return(nil)
			},
			wantError: constant.ErrBlockedByFDS,
		},
		{
			name: "SUCCESS: FDS approved", // NOSONAR
			setupMock: func() {

				fdsResponse := &fdscommon.TransactionAssessmentResponse{
					Result: constant.WorkflowFDSResultApprove, // NOSONAR
				}
				workflowFDSRepo.On(
					"AssessPayoutTransaction", mock.Anything, mock.Anything,
				).Once().Return(fdsResponse, nil)
				accountTxRepo.On(
					"UpdateFDSRiskAssessmentResultByID", mock.Anything, transactionID, fdsResponse,
				).Once().Return(nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			err := service.ExternalFDS(context.Background(), validPayout, validLedger)
			if test.wantError != nil {
				assert.EqualError(t, err, test.wantError.Error())
			} else {
				assert.NoError(t, err)
			}

			logger.AssertExpectations(t)
			selfSvc.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			accountTxRepo.AssertExpectations(t)
			workflowFDSRepo.AssertExpectations(t)
			disbursementRepo.AssertExpectations(t)
		})
	}
}

func TestFailTransactionByFDSResult(t *testing.T) {
	conf := &config.Config{
		Environment: constant.EnvironmentStaging,
	}
	logger := pdkLogMock.NewILogger(t)
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
	orchestratorSvc := serviceMocks.NewIOrchestratorService(t)
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)

	svc := &DisbursementService{
		config:              conf,
		logger:              logger,
		merchantRepo:        merchantRepo,
		disbursementRepo:    disbursementRepo,
		orchestratorSvc:     orchestratorSvc,
		statusHistoriesRepo: statusHistoriesRepo,
	}

	payoutID := uuid.NewString()
	transactionID := uuid.NewString()
	reasonType := constant.ReasonTypeBlockedByFDS
	reasonDesc := constant.ReasonDescBlockedByFDS

	validLedger := &orchestratorModel.TransactionAndFeeObject{
		TransactionID: transactionID,
	}

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR: Failed to begin transaction", // NOSONAR
			setupMock: func() {
				disbursementRepo.On("BeginTransaction", mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR: Failed to update transaction status", // NOSONAR
			setupMock: func() {

				disbursementRepo.On("BeginTransaction", mock.Anything).Return(t.Context(), nil)

				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", mock.Anything, transactionID, constant.StatusFailed, &reasonType, &reasonDesc,
				).Once().Return(assert.AnError)

				disbursementRepo.On("RollbackTransaction", mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				logger.On("Error", mock.Anything, "Failed to rollback payout transaction status update", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR: Failed to commit transaction", // NOSONAR
			setupMock: func() {
				orchestratorSvc.On(
					"UpdateStatusAccountTransaction", mock.Anything, transactionID, constant.StatusFailed, &reasonType, &reasonDesc,
				).Return(nil)
				statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil)
				disbursementRepo.On("CommitTransaction", mock.Anything).Once().Return(assert.AnError)
				disbursementRepo.On("RollbackTransaction", mock.Anything).Once().Return(nil)
			},
			wantError: assert.AnError,
		},
		{
			name:      "SUCCESS: Transaction failed by FDS", // NOSONAR
			setupMock: func() { disbursementRepo.On("CommitTransaction", mock.Anything).Once().Return(nil) },
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			err := svc.FailTransactionByFDSResult(t.Context(), payoutID, validLedger)
			assert.Equal(t, test.wantError, err)
		})
	}
}

func TestToAssessPayoutTransactionRequest(t *testing.T) {
	merchantID := uuid.NewString()
	payoutID := uuid.NewString()

	merchant := &merchant.Merchant{
		UUID:      merchantID,
		Name:      "Test Merchant",                            // NOSONAR
		RiskLevel: sql.NullString{String: "LOW", Valid: true}, // NOSONAR
	}

	payout := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:                   payoutID,
			ReferenceID:            "REF00011110100293", // NOSONAR
			Amount:                 decimal.NewFromFloat(12_250),
			Currency:               "IDR",          // NOSONAR
			BeneficiaryBankCode:    "002",          // NOSONAR
			BeneficiaryAccountNo:   "029930495555", // NOSONAR
			BeneficiaryAccountName: "John Doe",     // NOSONAR
			CreatedAt:              time.Date(2026, 2, 3, 2, 34, 2, 0, time.UTC),
			UpdatedAt:              time.Date(2026, 2, 3, 2, 34, 2, 0, time.UTC),
			CreatedFrom:            &[]string{"OPEN_API"}[0], // NOSONAR
		},
	}

	expectedRequest := fdscommon.AssessPayoutTransactionRequest{
		Merchant: fdscommon.Merchant{
			ID:        merchantID,
			Name:      "Test Merchant", // NOSONAR
			RiskLevel: "LOW",           // NOSONAR
		},
		Transaction: fdscommon.Transaction{
			ID:                payoutID,
			ClientReferenceID: "REF00011110100293", // NOSONAR
			Amount: common.Amount2{
				Value:    12_250,
				Currency: "IDR", // NOSONAR
			},
			CreatedAt:   time.Date(2026, 2, 3, 2, 34, 2, 0, time.UTC),
			UpdatedAt:   time.Date(2026, 2, 3, 2, 34, 2, 0, time.UTC),
			CreatedFrom: "OPEN_API", // NOSONAR
		},
		Destination: fdscommon.PayoutDestination{
			BankCode:      "002",          // NOSONAR
			AccountNumber: "029930495555", // NOSONAR
			AccountName:   "John Doe",     // NOSONAR
		},
		Metadata: map[string]any{},
	}

	svc := &DisbursementService{}
	result := svc.toAssessPayoutTransactionRequest(merchant, payout)

	assert.Equal(t, expectedRequest, result)
}
