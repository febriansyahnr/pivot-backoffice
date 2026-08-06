package adjustment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/adjustment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/adjustment"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func TestCreateMerchantBalanceAdjustment(t *testing.T) {
	adjustRepoMock := repoMocks.NewIAdjustmentRepository(t)
	merchantRepoMock := repoMocks.NewIMerchantRepository(t)
	orchestratorMock := serviceMocks.NewIOrchestratorService(t)
	loggerMock := loggerMocks.NewILogger(t)

	// Set up logger mock to avoid nil pointer dereference
	loggerMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()

	service := New(config.SlackConfig{}, adjustRepoMock, merchantRepoMock)
	WithOrchestratorService(service, orchestratorMock)
	WithLogger(service, loggerMock)

	ctx := context.Background()
	traceId := uuid.NewString()
	ctxWithTrace := context.WithValue(ctx, pdkConst.CtxTraceIdKey, traceId)
	ctxTx := context.WithValue(ctxWithTrace, mySqlExt.CtxSqlTx, &sqlx.Tx{})

	validMerchant := &merchant.Merchant{
		UUID: uuid.NewString(),
		Name: "Test Merchant",
	}

	validRequest := &adjustment.MerchantBalanceAdjustmentRequest{
		MerchantId:  uuid.NewString(),
		ReferenceId: "REF-123",
		BalanceType: c.AdjustmentPayoutBalanceDestination,
		Currency:    "IDR",
		Credit:      100000,
		Debit:       0,
		CreatedBy:   "test-admin",
		Remarks:     "Test adjustment",
	}

	tests := []struct {
		name         string
		request      *adjustment.MerchantBalanceAdjustmentRequest
		setupMock    func()
		wantErr      bool
		wantErrCode  string
		validateResp func(*testing.T, *adjustment.ManualAdjustmentHistory)
	}{
		{
			name:    "ERROR: Find merchant by ID - database error",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(nil, errors.New("database connection failed"))
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name:    "ERROR: Merchant not found",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(nil, nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrUnprocessableContent,
		},
		{
			name:    "ERROR: Begin transaction failed",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(nil, errors.New("begin transaction failed"))
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name: "ERROR: Invalid amount - both credit and debit are zero",
			request: &adjustment.MerchantBalanceAdjustmentRequest{
				MerchantId:  uuid.NewString(),
				ReferenceId: "REF-123",
				BalanceType: c.AdjustmentPayoutBalanceDestination,
				Currency:    "IDR",
				Credit:      0,
				Debit:       0,
				CreatedBy:   "test-admin",
				Remarks:     "Test adjustment",
			},
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(validMerchant, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrRequest,
		},
		{
			name: "ERROR: Invalid balance type",
			request: &adjustment.MerchantBalanceAdjustmentRequest{
				MerchantId:  uuid.NewString(),
				ReferenceId: "REF-123",
				BalanceType: "INVALID_BALANCE",
				Currency:    "IDR",
				Credit:      100000,
				Debit:       0,
				CreatedBy:   "test-admin",
				Remarks:     "Test adjustment",
			},
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(validMerchant, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrRequest,
		},
		{
			name:    "ERROR: Create adjustment failed",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				adjustRepoMock.On(
					"CreateAdjustment", mock.Anything, mock.AnythingOfType("*adjustment.ManualAdjustmentHistory"),
				).Once().Return(errors.New("create adjustment failed"))
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name:    "ERROR: Post account transaction failed",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				adjustRepoMock.On(
					"CreateAdjustment", mock.Anything, mock.AnythingOfType("*adjustment.ManualAdjustmentHistory"),
				).Once().Return(nil)
				orchestratorMock.On(
					"PostAccountTransaction", mock.Anything, mock.Anything,
				).Once().Return(errors.New("post account transaction failed"))
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name:    "ERROR: Commit transaction failed",
			request: validRequest,
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, validRequest.MerchantId,
				).Once().Return(validMerchant, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				adjustRepoMock.On(
					"CreateAdjustment", mock.Anything, mock.AnythingOfType("*adjustment.ManualAdjustmentHistory"),
				).Once().Return(nil)
				orchestratorMock.On(
					"PostAccountTransaction", mock.Anything, mock.Anything,
				).Once().Return(nil)
				adjustRepoMock.On(
					"CommitTransaction", mock.Anything,
				).Once().Return(errors.New("commit transaction failed"))
				adjustRepoMock.On(
					"RollbackTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr:     true,
			wantErrCode: response.HttpErrDatabase,
		},
		{
			name: "SUCCESS: Credit adjustment for payout balance",
			request: &adjustment.MerchantBalanceAdjustmentRequest{
				MerchantId:  uuid.NewString(),
				ReferenceId: "REF-CREDIT-123",
				BalanceType: c.AdjustmentPayoutBalanceDestination,
				Currency:    "IDR",
				Credit:      100000,
				Debit:       0,
				CreatedBy:   "test-admin",
				Remarks:     "Credit adjustment test",
			},
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(validMerchant, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				adjustRepoMock.On(
					"CreateAdjustment", mock.Anything, mock.AnythingOfType("*adjustment.ManualAdjustmentHistory"),
				).Once().Return(nil)
				orchestratorMock.On(
					"PostAccountTransaction", mock.Anything, mock.Anything,
				).Once().Return(nil)
				adjustRepoMock.On(
					"CommitTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *adjustment.ManualAdjustmentHistory) {
				assert.NotEmpty(t, resp.UUID)
				assert.Equal(t, float64(100000), resp.Amount)
				assert.Equal(t, c.AccountNameDisbursement, resp.Type)
				assert.Equal(t, "REF-CREDIT-123", resp.ReferenceID)
				assert.Equal(t, "Credit adjustment test", resp.Notes)
			},
		},
		{
			name: "SUCCESS: Debit adjustment for payment balance",
			request: &adjustment.MerchantBalanceAdjustmentRequest{
				MerchantId:  uuid.NewString(),
				ReferenceId: "REF-DEBIT-123",
				BalanceType: c.AdjustmentPaymentBalanceDestination,
				Currency:    "IDR",
				Credit:      0,
				Debit:       50000,
				CreatedBy:   "test-admin",
				Remarks:     "Debit adjustment test",
			},
			setupMock: func() {
				merchantRepoMock.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(validMerchant, nil)
				adjustRepoMock.On(
					"BeginTransaction", mock.Anything,
				).Once().Return(ctxTx, nil)
				adjustRepoMock.On(
					"CreateAdjustment", mock.Anything, mock.AnythingOfType("*adjustment.ManualAdjustmentHistory"),
				).Once().Return(nil)
				orchestratorMock.On(
					"PostAccountTransaction", mock.Anything, mock.Anything,
				).Once().Return(nil)
				adjustRepoMock.On(
					"CommitTransaction", mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *adjustment.ManualAdjustmentHistory) {
				assert.NotEmpty(t, resp.UUID)
				assert.Equal(t, float64(-50000), resp.Amount) // Debit amount should be negative
				assert.Equal(t, c.AccountNamePayment, resp.Type)
				assert.Equal(t, "REF-DEBIT-123", resp.ReferenceID)
				assert.Equal(t, "Debit adjustment test", resp.Notes)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := service.CreateMerchantBalanceAdjustment(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				if tt.wantErrCode != "" {
					errCode, _ := pkgErrs.ExtractError(err)
					assert.Equal(t, tt.wantErrCode, errCode)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				// Common validations
				assert.NotEmpty(t, result.UUID)
				assert.Equal(t, tt.request.MerchantId, result.MerchantID)
				assert.Equal(t, tt.request.Currency, result.Currency)
				assert.Equal(t, tt.request.CreatedBy, result.CreatedBy)
				assert.Equal(t, tt.request.ReferenceId, result.ReferenceID)
				assert.WithinDuration(t, time.Now().UTC(), result.CreatedAt, 5*time.Second)
				assert.WithinDuration(t, time.Now().UTC(), result.UpdatedAt, 5*time.Second)

				// Custom validations if provided
				if tt.validateResp != nil {
					tt.validateResp(t, result)
				}
			}

			// Verify all expectations were met
			merchantRepoMock.AssertExpectations(t)
			adjustRepoMock.AssertExpectations(t)
			orchestratorMock.AssertExpectations(t)
		})
	}
}
