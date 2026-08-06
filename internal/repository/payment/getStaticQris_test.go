package paymentRepository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestFilterStaticQrisList(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    paymentModel.StaticQrisFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get static QRIS list without filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS list with status filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Status:     "ACTIVE",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error on count query",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database error"))

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get static QRIS list with derived merchant ID context",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS list with status sort",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "ASC",
				SortBy:     "status",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS list with referenceId sort",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "ASC",
				SortBy:     "referenceId",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS list with ID filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				ID:         "search-term",
				Page:       1,
				PerPage:    10,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS list with date filters",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				StartDate:  time.Now().AddDate(0, 0, -7),
				EndDate:    time.Now(),
				Page:       1,
				PerPage:    10,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS list with payment method ID filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID:      "merchant-123",
				PaymentMethodID: "payment-method-123",
				Page:            1,
				PerPage:         10,
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error on select query",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(errors.New("select error"))
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get static QRIS list with unknown sort option (default case)",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "ASC",
				SortBy:     "unknownField", // This will trigger the default case
			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			ctx := context.Background()
			if tc.name == "SUCCESS: Get static QRIS list with derived merchant ID context" {
				ctx = context.WithValue(ctx, constant.CtxDerivedMerchantID, "parent-merchant-123")
			}

			repo := paymentRepository.New(mysqlMock, logger)
			result, err := repo.FilterStaticQrisList(ctx, tc.filter)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.IsType(t, &commonModel.PaginationResponse{}, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}

func TestGetStaticQrisDetail(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		request   paymentModel.StaticQrisDetailRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get static QRIS detail",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.StaticQrisDetailResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			request: paymentModel.StaticQrisDetailRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS detail with empty total amount",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.StaticQrisDetailResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					resp := args.Get(1).(*paymentModel.StaticQrisDetailResponse)
					resp.TotalAmountValue = ""
				}).Return(nil)
			},
			request: paymentModel.StaticQrisDetailRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
			},
			wantErr: false,
		},
		{
			name: "ERROR: Payment not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.StaticQrisDetailResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(errors.New("record not found"))
			},
			request: paymentModel.StaticQrisDetailRequest{
				PaymentID:  "invalid-payment",
				MerchantID: "merchant-123",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := paymentRepository.New(mysqlMock, logger)
			result, err := repo.GetStaticQrisDetail(context.Background(), tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.IsType(t, &paymentModel.StaticQrisDetailResponse{}, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}

func TestGetStaticQrisTransactions(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    paymentModel.StaticQrisTransactionFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get static QRIS transactions",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "paymentDate",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS transactions with status filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Status:     "SUCCESS",
				StartDate:  time.Now().AddDate(0, 0, -7),
				EndDate:    time.Now(),
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "paymentDate",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS transactions with createdAt sort",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					transactions := args.Get(1).(*[]paymentModel.StaticQrisTransactionItem)
					*transactions = append(*transactions, paymentModel.StaticQrisTransactionItem{
						AmountValue:    "",
						AmountCurrency: "",
					})
				}).Return(nil)
			},
			filter: paymentModel.StaticQrisTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "ASC",
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS transactions with amount sort",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "ASC",
				SortBy:     "amount",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static QRIS transactions with ID filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				ID:         "transaction-search",
				Page:       1,
				PerPage:    10,
				Sort:       "ASC",
				SortBy:     "paymentDate",
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database connection failed"))

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticQrisTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticQrisTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := paymentRepository.New(mysqlMock, logger)
			result, err := repo.GetStaticQrisTransactions(context.Background(), tc.filter)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.IsType(t, &commonModel.PaginationResponse{}, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}
