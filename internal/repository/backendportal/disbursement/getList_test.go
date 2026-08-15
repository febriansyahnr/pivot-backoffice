package disbursementRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *disbursementModel.GetDisbursementFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter:  &disbursementModel.GetDisbursementFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("no rows data"))
			},
			filter:  &disbursementModel.GetDisbursementFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List with filter bulk id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("no rows data"))
			},
			filter: &disbursementModel.GetDisbursementFilterRequest{
				MerchantID:     uuid.NewString(),
				BulkID:         "uuid",
				Status:         constant.DisbursementStatusWaiting,
				StartCreatedAt: &util.TimeNow,
				EndCreatedAt:   &util.TimeNow,
				Type:           constant.DisbursementTypeBulk,
				Sort:           "ASC",
				SortBy:         "updatedAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List with filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("no rows data"))
			},
			filter: &disbursementModel.GetDisbursementFilterRequest{
				MerchantID:        uuid.NewString(),
				Status:            constant.DisbursementStatusWaiting,
				StartCreatedAt:    &util.TimeNow,
				EndCreatedAt:      &util.TimeNow,
				Type:              constant.DisbursementTypeSingle,
				TransactionStatus: constant.StatusSuccess,
				Keyword:           "test",
				ReasonType:        constant.XbDisbursementReasonTypePending,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List with filter (status pending)",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					constant.StringMockType(),
				).Return(errors.New("no rows data"))
			},
			filter: &disbursementModel.GetDisbursementFilterRequest{
				MerchantID:     uuid.NewString(),
				Status:         constant.DisbursementStatusPending,
				StartCreatedAt: &util.TimeNow,
				EndCreatedAt:   &util.TimeNow,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with CANCELLED transaction status filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("no rows data"))
			},
			filter: &disbursementModel.GetDisbursementFilterRequest{
				MerchantID:        uuid.NewString(),
				StartCreatedAt:    &util.TimeNow,
				EndCreatedAt:      &util.TimeNow,
				TransactionStatus: constant.DisbursementReasonTypeCancelled,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with UUID filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &disbursementModel.GetDisbursementFilterRequest{
				MerchantID: uuid.NewString(),
				UUID:       uuid.NewString(),
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with createdAt sortBy",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &disbursementModel.GetDisbursementFilterRequest{
				MerchantID: uuid.NewString(),
				SortBy:     "createdAt",
				Sort:       "DESC",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with IsXbPayout true",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			filter: &disbursementModel.GetDisbursementFilterRequest{
				MerchantID: uuid.NewString(),
				IsXbPayout: true,
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*disbursementModel.DisbursementWithTransaction"),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("some-error"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

			},
			filter:  &disbursementModel.GetDisbursementFilterRequest{},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			_, err := repo.GetList(ctx, tc.filter, 0, 20)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestBuildRespData(t *testing.T) {
	// Create test data
	testData := []*disbursementModel.DisbursementWithTransaction{
		{
			Disbursement: disbursementModel.Disbursement{
				Status: constant.DisbursementStatusApproved,
			},
			TransactionReasonType:        stringPtr(constant.ReasonTypeBeneficiaryAccountReason),
			TransactionReasonDescription: stringPtr("Invalid account"),
		},
		{
			Disbursement: disbursementModel.Disbursement{
				Status: constant.DisbursementStatusRejected,
			},
			TransactionReasonType:        stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
			TransactionReasonDescription: stringPtr("Insufficient balance"),
		},
	}

	// Call the function
	result := BuildRespData(testData)

	// Assert the results
	assert.Equal(t, len(testData), len(result), "Result length should match input length")

	// Check first item
	assert.Equal(t, constant.DisbursementStatusApproved, result[0].Status)

	// Check second item
	assert.Equal(t, constant.DisbursementStatusRejected, result[1].Status)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestGetListWithOverFetchPagination(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		appConfig *config.AppConfig
		filter    *disbursementModel.GetDisbursementFilterRequest
		page      int64
		perPage   int64
		wantErr   bool
	}{
		{
			name: "SUCCESS: Page 1 - Small dataset",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[1].(*[]*disbursementModel.DisbursementWithTransaction)
						for i := 0; i < 5; i++ {
							*data = append(*data, &disbursementModel.DisbursementWithTransaction{})
						}
					}).Return(nil)
			},
			appConfig: &config.AppConfig{
				UseOverFetchPagination: true,
				InitialPageWindow:      3,
			},
			filter:  &disbursementModel.GetDisbursementFilterRequest{MerchantID: "test-merchant"},
			page:    1,
			perPage: 10,
			wantErr: false,
		},
		{
			name: "SUCCESS: Page 1 - Large dataset",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						data := args[1].(*[]*disbursementModel.DisbursementWithTransaction)
						for i := 0; i < 31; i++ {
							*data = append(*data, &disbursementModel.DisbursementWithTransaction{})
						}
					}).Return(nil)
			},
			appConfig: &config.AppConfig{
				UseOverFetchPagination: true,
				InitialPageWindow:      3,
			},
			filter:  &disbursementModel.GetDisbursementFilterRequest{MerchantID: "test-merchant"},
			page:    1,
			perPage: 10,
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(
				mockMysql,
				mockLogger,
				WithAppConfig(tc.appConfig),
			)

			ctx := context.Background()
			result, err := repo.GetList(ctx, tc.filter, tc.page, tc.perPage)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestGetCardFundedPayoutTransactionList(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	merchantID := "49b8cff4-48dd-4cf6-a57f-767305fb7c3a"
	startDate := time.Now().UTC().Add(-24 * time.Hour)
	endDate := time.Now().UTC()

	tests := []struct {
		name       string
		request    cardFundedPayoutModel.GetPayoutTransactionListRequest
		setupMock  func()
		wantError  error
		wantResult []cardFundedPayoutModel.GetPayoutTransactionListResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			request: cardFundedPayoutModel.GetPayoutTransactionListRequest{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
			},
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, startDate, endDate, merchantID, startDate, endDate).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:Data not found", // NOSONAR
			request: cardFundedPayoutModel.GetPayoutTransactionListRequest{
				MerchantID: merchantID,
				StartDate:  startDate,
				EndDate:    endDate,
				TrxStatus:  constant.StatusSuccess,
			},
			setupMock: func() {
				db.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, startDate, endDate, merchantID, startDate, endDate, constant.StatusSuccess).Once().Return(sql.ErrNoRows)
			},
			wantResult: []cardFundedPayoutModel.GetPayoutTransactionListResponse{},
		},
		{
			name: "SUCCESS:Data found", // NOSONAR
			request: cardFundedPayoutModel.GetPayoutTransactionListRequest{
				MerchantID:    merchantID,
				StartDate:     startDate,
				EndDate:       endDate,
				TrxStatus:     constant.StatusPending,
				TrxReasonType: constant.ReasonTypeWaitingManualAction,
			},
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, startDate, endDate, merchantID, startDate, endDate, constant.StatusPending, constant.ReasonTypeWaitingManualAction,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*[]cardFundedPayoutModel.GetPayoutTransactionListResponse) = []cardFundedPayoutModel.GetPayoutTransactionListResponse{{ID: "e8c19b62-6525-40c8-9623-0529453b505e"}}
				}).Return(nil)
			},
			wantResult: []cardFundedPayoutModel.GetPayoutTransactionListResponse{{ID: "e8c19b62-6525-40c8-9623-0529453b505e"}},
		},
		{
			name: "SUCCESS:Data found with other reason", // NOSONAR
			request: cardFundedPayoutModel.GetPayoutTransactionListRequest{
				MerchantID:    merchantID,
				StartDate:     startDate,
				EndDate:       endDate,
				TrxStatus:     constant.StatusPending,
				TrxReasonType: constant.ReasonTypeOtherReason,
			},
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, startDate, endDate, merchantID, startDate, endDate, constant.StatusPending, constant.ReasonTypeOtherReason,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*[]cardFundedPayoutModel.GetPayoutTransactionListResponse) = []cardFundedPayoutModel.GetPayoutTransactionListResponse{{ID: "e8c19b62-6525-40c8-9623-0529453b505e"}}
				}).Return(nil)
			},
			wantResult: []cardFundedPayoutModel.GetPayoutTransactionListResponse{{ID: "e8c19b62-6525-40c8-9623-0529453b505e"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := repo.GetCardFundedPayoutTransactionList(t.Context(), tt.request)
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
