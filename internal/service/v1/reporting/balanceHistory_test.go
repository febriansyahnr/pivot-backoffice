package reportingService_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	cdcModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cdc"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	reportingModel "github.com/paper-indonesia/pivot-backoffice/internal/model/reporting"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/reporting"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpsertBalanceHistory(t *testing.T) {
	log := logMock.NewILogger(t)
	repo := repoMocks.NewIReportingRepository(t)
	accountRepo := repoMocks.NewIAccountRepository(t)

	service := New(log, repo, accountRepo)

	var (
		now           = time.Now()
		deletedAt     = now.Add(-1 * time.Hour)
		transactionID = "2ac93f16-93d8-4c2c-a0f2-27c48887617b"
		merchantID    = "12f513ca-d538-412a-92a2-6a02344d9b6c"
		accountID     = "ba0c388f-eeea-4322-9ac8-5c5afcb20e42"
	)

	successPayload := &cdcModel.AccountTransaction{
		UUID:       transactionID,
		MerchantID: merchantID,
		AccountID:  accountID,
		Credit:     decimal.NewFromInt(10000),
		Status:     constant.StatusSuccess,
		Type:       constant.TypePayment,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	pendingPayload := &cdcModel.AccountTransaction{
		UUID:       transactionID,
		MerchantID: merchantID,
		AccountID:  accountID,
		Credit:     decimal.NewFromInt(10000),
		Status:     constant.StatusPending,
		Type:       constant.TypePayment,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	deletedPayload := &cdcModel.AccountTransaction{
		UUID:       transactionID,
		MerchantID: merchantID,
		AccountID:  accountID,
		Credit:     decimal.NewFromInt(10000),
		Status:     constant.StatusSuccess,
		Type:       constant.TypePayment,
		DeletedAt:  &deletedAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	pendingSettlementPayload := &cdcModel.AccountTransaction{
		UUID:             transactionID,
		MerchantID:       merchantID,
		AccountID:        accountID,
		Credit:           decimal.NewFromInt(10000),
		Status:           constant.StatusSuccess,
		Type:             constant.TypePayment,
		SettlementStatus: util.ValueToPtr(constant.StatusPending),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	successSettlementPayload := &cdcModel.AccountTransaction{
		UUID:             transactionID,
		MerchantID:       merchantID,
		AccountID:        accountID,
		Credit:           decimal.NewFromInt(10000),
		Status:           constant.StatusSuccess,
		Type:             constant.TypePayment,
		SettlementStatus: util.ValueToPtr(constant.StatusSuccess),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	tests := []struct {
		name      string
		request   reportingModel.UpsertBalanceHistoryRequest
		setupMock func()
		wantError error
	}{
		{
			name: "SUCCESS:Event excluded - non-success status on create",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpCreate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       transactionID,
						MerchantID: merchantID,
						Credit:     decimal.NewFromInt(10000),
						Debit:      decimal.Zero,
						Status:     constant.StatusPending,
						Type:       constant.TypePayment,
						CreatedAt:  now,
						UpdatedAt:  now,
					},
				},
			},
			setupMock: func() {
				log.On("Info", mock.Anything, "Event not processed: data does not meet criteria", mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Event excluded - zero debit and credit",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:   cdcModel.OpCreate,
					TsMs: now.UnixMilli(),
					After: &cdcModel.AccountTransaction{
						UUID:       transactionID,
						MerchantID: merchantID,
						Status:     constant.StatusSuccess,
						Type:       constant.TypePayment,
						CreatedAt:  now, UpdatedAt: now,
					},
				},
			},
			setupMock: func() {
				log.On("Info", mock.Anything, "Event not processed: data does not meet criteria", mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Get account details",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpDelete,
					TsMs:   now.UnixMilli(),
					Before: successPayload,
				},
			},
			setupMock: func() {
				accountRepo.On("GetByUUID", mock.Anything, util.ParseUUID(accountID)).Once().Return(nil, assert.AnError)
				log.On("Error", mock.Anything, "Failed to get account details", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:Account not found",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:    cdcModel.OpCreate,
					TsMs:  now.UnixMilli(),
					After: successPayload,
				},
			},
			setupMock: func() {
				accountRepo.On("GetByUUID", mock.Anything, util.ParseUUID(accountID)).Once().Return(nil, nil)
				log.On("Info", mock.Anything, fmt.Sprintf("Account ID %s not found. Balance history report generation skipped for this transaction", accountID)).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Customer wallet transactions",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpDelete,
					TsMs:   now.UnixMilli(),
					Before: successPayload,
				},
			},
			setupMock: func() {
				accountRepo.On("GetByUUID", mock.Anything, util.ParseUUID(accountID)).Once().Return(&accountModel.Account{UserType: constant.UserTypeCustomer}, nil)
				log.On("Info", mock.Anything, "Event skipped because it is a customer wallet transaction").Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Delete operation - hard delete",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpDelete,
					TsMs:   now.UnixMilli(),
					Before: successPayload,
				},
			},
			setupMock: func() {
				accountRepo.On("GetByUUID", mock.Anything, util.ParseUUID(accountID)).Return(&accountModel.Account{UserType: constant.UserTypeMerchant}, nil)
				repo.On("HardDeleteBalanceHistory", mock.Anything, transactionID).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Delete operation - hard delete",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpUpdate,
					TsMs:   now.UnixMilli(),
					Before: successPayload,
					After:  pendingPayload,
				},
			},
			setupMock: func() {
				repo.On("HardDeleteBalanceHistory", mock.Anything, transactionID).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "ERROR:Delete operation - hard delete fails",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpDelete,
					TsMs:   now.UnixMilli(),
					Before: successPayload,
				},
			},
			setupMock: func() {
				repo.On("HardDeleteBalanceHistory", mock.Anything, transactionID).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:Update operation - soft delete (deleted_at is not nil)",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpUpdate,
					TsMs:   now.UnixMilli(),
					Before: successPayload,
					After:  deletedPayload,
				},
			},
			setupMock: func() {
				repo.On("SoftDeleteBalanceHistory", mock.Anything, transactionID, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "ERROR:Update operation - soft delete fails",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpUpdate,
					TsMs:   now.UnixMilli(),
					Before: successPayload,
					After:  deletedPayload,
				},
			},
			setupMock: func() {
				repo.On("SoftDeleteBalanceHistory", mock.Anything, transactionID, mock.Anything).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:Update operation - settlement status change from pending to success",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpUpdate,
					TsMs:   now.UnixMilli(),
					Before: pendingSettlementPayload,
					After:  successSettlementPayload,
				},
			},
			setupMock: func() {
				repo.On("UpdateSettlementBalanceHistory", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
		{
			name: "ERROR:Update operation - update settlement balance history fails",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpUpdate,
					TsMs:   now.UnixMilli(),
					Before: pendingSettlementPayload,
					After:  successSettlementPayload,
				},
			},
			setupMock: func() {
				repo.On("UpdateSettlementBalanceHistory", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed when update settlement balance history data", logger.Error(assert.AnError)).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Prepare advanced balance history data fails",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:     cdcModel.OpUpdate,
					TsMs:   now.UnixMilli(),
					Before: pendingSettlementPayload,
					After:  successSettlementPayload,
				},
			},
			setupMock: func() {
				repo.On("UpdateSettlementBalanceHistory", mock.Anything, mock.Anything).Once().Return(constant.ErrNoRowsAffected)
				repo.On("PrepareAdvancedBalanceHistoryData", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				log.On("Error", mock.Anything, "Failed when prepare advanced balance history data", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Upsert balance history fails",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:    cdcModel.OpCreate,
					TsMs:  now.UnixMilli(),
					After: successPayload,
				},
			},
			setupMock: func() {
				repo.On("PrepareAdvancedBalanceHistoryData", mock.Anything, mock.Anything).Once().Return(nil)
				repo.On("UpsertBalanceHistory", mock.Anything, mock.Anything).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:Create operation - upsert balance history",
			request: reportingModel.UpsertBalanceHistoryRequest{
				Event: &cdcModel.Event[cdcModel.AccountTransaction]{
					Op:    cdcModel.OpCreate,
					TsMs:  now.UnixMilli(),
					After: successPayload,
				},
			},
			setupMock: func() {
				repo.On("PrepareAdvancedBalanceHistoryData", mock.Anything, mock.Anything).Once().Return(nil)
				repo.On("UpsertBalanceHistory", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := service.UpsertBalanceHistory(t.Context(), &test.request)

			require.Equal(t, test.wantError, err)

			log.AssertExpectations(t)
			repo.AssertExpectations(t)
			accountRepo.AssertExpectations(t)
		})
	}
}

func TestListBalanceHistory(t *testing.T) {
	logger := logMock.NewILogger(t)
	reportingRepo := repoMocks.NewIReportingRepository(t)

	filter := &orchestratorModel.TransactionHistoryFilterRequest{}
	page, perPage := int64(1), int64(10)

	service := New(logger, reportingRepo, nil)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *commonModel.PaginationResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				logger.On("Error", mock.Anything, "Failed when list balance history via data reporting", mock.Anything).Once().Return()
				reportingRepo.On("ListBalanceHistory", mock.Anything, filter, page, perPage).Once().Return(nil, assert.AnError)
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				reportingRepo.On("ListBalanceHistory", mock.Anything, filter, page, perPage).Once().Return(&commonModel.PaginationResponse{}, nil)
			},
			wantResult: &commonModel.PaginationResponse{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.ListBalanceHistory(t.Context(), filter, page, perPage)

			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
