package paymentRepository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFilterPaymentHistory(t *testing.T) {
	validUUID := uuid.New().String()

	// Define test cases using the Test Data Table (TDT) approach
	tests := []struct {
		name         string
		opt          paymentModel.FilterPaymentHistoryOption
		mockSetup    func(*testing.T, *mysqlMocks.IMySqlExt)
		wantErr      string
		wantData     []paymentModel.PaymentHistoryItem
		useAppConfig bool
	}{
		{
			name: "SUCCESS: Filter by all field",
			opt: paymentModel.FilterPaymentHistoryOption{
				MerchantID:       uuid.NewString(),
				ReferenceID:      "test",
				Status:           "OKE",
				PaymentMethod:    "VA",
				StartDate:        util.TimeNow,
				EndDate:          util.TimeNow,
				PaymentStartDate: util.TimeNow,
				PaymentEndDate:   util.TimeNow,
				Sort:             "ASC",
				SortBy:           "createdAt",
				Page:             1,
				PerPage:          10,
			},
			mockSetup: func(t *testing.T, db *mysqlMocks.IMySqlExt) {
				// Mock GetContext for count query
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 1 // Setting totalRecord to 1
				}).Once()

				// Mock SelectContext for data query
				db.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]paymentModel.PaymentHistoryItem"),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).
					Run(func(args mock.Arguments) {
						*args.Get(1).(*[]paymentModel.PaymentHistoryItem) = []paymentModel.PaymentHistoryItem{
							{
								UUID:               validUUID,
								MerchantID:         "mock-merchant-id",
								ReferenceID:        "mock-reference-id",
								Method:             "CARD",
								MethodType:         "",
								Channel:            "VISA",
								ProcessorRefNumber: util.ValueToPtr("********1234"),
								Amount:             "10000.0",
								AmountCurrency:     "IDR",
								AmountPaid:         util.ValueToPtr("10000.0"),
								AmountPaidCurrency: util.ValueToPtr("IDR"),
								Status:             "SUCCESS",
								CreatedAt:          util.TimeNow,
								UpdatedAt:          util.ValueToPtr(util.TimeNow),
								HasSplitRouting:    false,
								HasInvestigation:   false,
							},
						}
					}).
					Once()
			},
			wantData: []paymentModel.PaymentHistoryItem{
				{
					UUID:               validUUID,
					MerchantID:         "mock-merchant-id",
					ReferenceID:        "mock-reference-id",
					Method:             "CARD",
					MethodType:         "",
					Channel:            "VISA",
					ProcessorRefNumber: util.ValueToPtr("********1234"),
					Amount:             "10000.0",
					AmountCurrency:     "IDR",
					AmountPaid:         util.ValueToPtr("10000.0"),
					AmountPaidCurrency: util.ValueToPtr("IDR"),
					Status:             "SUCCESS",
					CreatedAt:          util.TimeNow,
					UpdatedAt:          util.ValueToPtr(util.TimeNow),
					HasSplitRouting:    false,
					HasInvestigation:   false,
				},
			},
		},
		{
			name: "SUCCESS: Filter with default pagination",
			opt: paymentModel.FilterPaymentHistoryOption{
				MerchantID: uuid.NewString(),
				SortBy:     "amountPaid",
				Sort:       "ASC",
				Page:       -1,
				PerPage:    -10,
			},
			mockSetup: func(t *testing.T, db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 0
				}).Once()

				db.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]paymentModel.PaymentHistoryItem"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows).Once()
			},
		},
		{
			name: "ERROR: Database error",
			opt: paymentModel.FilterPaymentHistoryOption{
				MerchantID: uuid.NewString(),
				Page:       1,
				PerPage:    10,
			},
			mockSetup: func(t *testing.T, db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

				db.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]paymentModel.PaymentHistoryItem"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "SUCCESS: Sort by paymentDate",
			opt: paymentModel.FilterPaymentHistoryOption{
				MerchantID: uuid.NewString(),
				SortBy:     "paymentDate",
				Sort:       "DESC",
				Page:       1,
				PerPage:    10,
			},
			mockSetup: func(t *testing.T, db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 0
				}).Once()

				db.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]paymentModel.PaymentHistoryItem"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Once()
			},
		},
		{
			name: "SUCCESS: Status with PAID (adds SUCCESS status)",
			opt: paymentModel.FilterPaymentHistoryOption{
				MerchantID: uuid.NewString(),
				Status:     "PAID",
				Page:       1,
				PerPage:    10,
			},
			mockSetup: func(t *testing.T, db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 0
				}).Once()

				db.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]paymentModel.PaymentHistoryItem"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Once()
			},
		},
		{
			name: "SUCCESS: Status with PROCESSING (adds PENDING status)",
			opt: paymentModel.FilterPaymentHistoryOption{
				MerchantID: uuid.NewString(),
				Status:     "PROCESSING",
				Page:       1,
				PerPage:    10,
			},
			mockSetup: func(t *testing.T, db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 0
				}).Once()

				db.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]paymentModel.PaymentHistoryItem"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Once()
			},
		},
		{
			name: "SUCCESS: SettlementModel AGGREGATOR",
			opt: paymentModel.FilterPaymentHistoryOption{
				MerchantID:      uuid.NewString(),
				SettlementModel: constant.PaymentMethodChannelTypeAggregator,
				Page:            1,
				PerPage:         10,
			},
			mockSetup: func(t *testing.T, db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 0
				}).Once()

				db.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]paymentModel.PaymentHistoryItem"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Once()
			},
		},
		{
			name: "SUCCESS: SettlementModel NON-AGGREGATOR",
			opt: paymentModel.FilterPaymentHistoryOption{
				MerchantID:      uuid.NewString(),
				SettlementModel: "MERCHANT_SETTLEMENT",
				Page:            1,
				PerPage:         10,
			},
			mockSetup: func(t *testing.T, db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int64"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 0
				}).Once()

				db.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType("*[]paymentModel.PaymentHistoryItem"),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Once()
			},
		},
		{
			name: "SUCCESS: With appConfig InitialPageWindow",
			opt: paymentModel.FilterPaymentHistoryOption{
				MerchantID: uuid.NewString(),
				Page:       1,
				PerPage:    10,
			},
			useAppConfig: true,
			mockSetup: func(t *testing.T, db *mysqlMocks.IMySqlExt) {
				db.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					if ptr, ok := args.Get(1).(*int64); ok {
						*ptr = 1
					}
				}).Maybe()

				db.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Maybe()
			},
		},
	}

	// Iterate over the test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh db mock for each test case
			db := mysqlMocks.NewIMySqlExt(t)
			tc.mockSetup(t, db)

			// Create repo with or without appConfig
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			var testRepo repository.IPaymentRepository
			if tc.useAppConfig {
				testRepo = New(db, mockLogger, WithAppConfig(&config.AppConfig{
					UseOverFetchPagination: true,
					InitialPageWindow:      20,
				}))
			} else {
				testRepo = New(db, mockLogger)
			}

			result, err := testRepo.FilterPaymentHistory(context.Background(), tc.opt)
			if tc.wantErr == "" {
				require.NoError(t, err)
				if tc.wantData != nil {
					assert.Equal(t, tc.wantData, result.Data)
				}
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestGetPaymentIDsExpiringToday(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	repo := New(db, mockLogger)
	result := []*paymentModel.ExpiringPayment{
		{
			UUID:       "uuid-1",
			MerchantID: "merchant-1",
			ExpiredAt:  util.TimeNow,
		},
		{
			UUID:       "uuid-2",
			MerchantID: "merchant-2",
			ExpiredAt:  util.TimeNow,
		},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult []*paymentModel.ExpiringPayment
	}{
		{
			name: "when failed to get payment ids expiring today, then should return error",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "when the are no payment ids expiring today, then should not return error",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Once().Return(sql.ErrNoRows)

			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), constant.TimeMockType(), constant.TimeMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]*paymentModel.ExpiringPayment)) = result
				}).Return(nil)
			},
			wantResult: []*paymentModel.ExpiringPayment{
				{
					UUID:       "uuid-1",
					MerchantID: "merchant-1",
					ExpiredAt:  util.TimeNow,
				},
				{
					UUID:       "uuid-2",
					MerchantID: "merchant-2",
					ExpiredAt:  util.TimeNow,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetExpiringPayments(context.Background(), time.Now(), time.Now())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
