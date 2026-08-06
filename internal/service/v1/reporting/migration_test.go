package reportingService_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	cdcModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cdc"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/reporting"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMigrateBalanceHistoryToDataReporting(t *testing.T) {
	tz, _ := time.LoadLocation(constant.TimeLoc)

	logger := loggerMock.NewILogger(t)
	repo := repoMocks.NewIReportingRepository(t)
	accountRepo := repoMocks.NewIAccountRepository(t)

	service := New(logger, repo, accountRepo)

	var (
		now           = time.Now()
		startDate     = time.Date(2026, 3, 10, 17, 0, 0, 0, time.UTC)
		endDate       = time.Date(2026, 3, 11, 16, 59, 59, 0, time.UTC)
		transactionID = "2ac93f16-93d8-4c2c-a0f2-27c48887617b" // NOSONAR
		merchantID    = "12f513ca-d538-412a-92a2-6a02344d9b6c" // NOSONAR
		accountID     = "ba0c388f-eeea-4322-9ac8-5c5afcb20e42" // NOSONAR
	)

	transaction := cdcModel.AccountTransaction{
		UUID:              transactionID,
		MerchantID:        merchantID,
		AccountID:         accountID,
		Credit:            decimal.NewFromInt(10000),
		Status:            constant.StatusSuccess,
		Type:              constant.TypePayment,
		SettlementStatus:  util.ValueToPtr(constant.StatusSuccess),
		RawAdditionalInfo: util.ValueToPtr(`{"notes": ""}`),
		CreatedAt:         now, UpdatedAt: now,
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		setupMock func()
		wantError error
	}{
		{
			name:      "ERROR:Invalid start and end date",
			startDate: time.Date(2026, 3, 11, 17, 0, 0, 0, time.UTC),
			endDate:   time.Date(2026, 3, 10, 17, 0, 0, 0, time.UTC),
			setupMock: func() { /* Empty Function */ },
			wantError: nil,
		},
		{
			name:      "ERROR:List account transactions data",
			startDate: startDate,
			endDate:   endDate,
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, fmt.Sprintf("Starting the process from %s to %s", startDate.In(tz).Format(time.DateTime), endDate.In(tz).Format(time.DateTime)),
				).Return()
				repo.On("ListAccountTransactionsForMigration", mock.Anything, startDate, endDate).Once().Return(nil, assert.AnError)

				logger.On("Error", mock.Anything, "Failed to list account transactions for migration", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name:      "SUCCESS:Transaction not found",
			startDate: startDate,
			endDate:   endDate,
			setupMock: func() {
				repo.On("ListAccountTransactionsForMigration", mock.Anything, startDate, endDate).Once().Return([]cdcModel.AccountTransaction{}, nil)

				logger.On("Info", mock.Anything, "Transaction not found, data migration process completed").Once().Return()
			},
			wantError: nil,
		},
		{
			name:      "ERROR:Context canceled",
			startDate: startDate,
			endDate:   endDate,
			setupMock: func() {

				cancel()

				repo.On("ListAccountTransactionsForMigration", mock.Anything, startDate, endDate).Return([]cdcModel.AccountTransaction{transaction}, nil)
			},
			wantError: nil,
		},
		{
			name:      "ERROR:Get account details",
			startDate: startDate,
			endDate:   endDate,
			setupMock: func() {

				ctx = t.Context()

				accountRepo.On("GetByUUID", mock.Anything, util.ParseUUID(accountID)).Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to get account details", mock.Anything).Once().Return()
				logger.On("Error", mock.Anything, "Failed to upsert balance history data", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name:      "SUCCESS",
			startDate: startDate,
			endDate:   endDate,
			setupMock: func() {
				accountRepo.On("GetByUUID", mock.Anything, util.ParseUUID(accountID)).Once().Return(&accountModel.Account{UserType: constant.UserTypeMerchant}, nil)
				repo.On("PrepareAdvancedBalanceHistoryData", mock.Anything, mock.Anything).Once().Return(nil)
				repo.On("UpsertBalanceHistory", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			assert.Equal(t, tt.wantError, service.MigrateBalanceHistoryToDataReporting(ctx, tt.startDate, tt.endDate))

			repo.AssertExpectations(t)
			logger.AssertExpectations(t)
			accountRepo.AssertExpectations(t)
		})
	}
}
