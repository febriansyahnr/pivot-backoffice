package paymentRepository_test

import (
	"context"
	"database/sql"
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

func TestFilterStaticVaList(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    paymentModel.StaticVaFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get static VA list without filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static VA list with status filter",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
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
			name: "SUCCESS: Get static VA list with query filter",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				ID:         "VAname234",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static VA list with date range filter",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				StartDate:  time.Now().AddDate(0, 0, -7),
				EndDate:    time.Now(),
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
				).Return(errors.New("database error"))

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
			wantErr: true,
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
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
				).Return(errors.New("select query failed"))
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get static VA list with SortBy status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "ASC",
				SortBy:     "status",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static VA list with SortBy referenceId",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "referenceId",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static VA list with default SortBy (invalid value)",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "invalidSortBy",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static VA list with BankName filter",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				BankName:   "BCA",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static VA list with sql.ErrNoRows from SelectContext",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "createdAt",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static VA list with sql.ErrNoRows from GetContext",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType(constant.MockTypeIntReference),
					constant.StringMockType(),
					mock.Anything,
				).Return(sql.ErrNoRows)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaListResponse"),
					constant.StringMockType(),
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaFilterRequest{
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "ASC",
				SortBy:     "status",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := paymentRepository.New(mysqlMock, logger)
			result, err := repo.FilterStaticVaList(context.Background(), tc.filter)

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

func TestGetStaticVaDetail(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		request   paymentModel.StaticVaDetailRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get static VA detail",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.StaticVaDetailResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			request: paymentModel.StaticVaDetailRequest{
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
					mock.AnythingOfType("*paymentModel.StaticVaDetailResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(errors.New("record not found"))
			},
			request: paymentModel.StaticVaDetailRequest{
				PaymentID:  "invalid-payment",
				MerchantID: "merchant-123",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Database connection error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*paymentModel.StaticVaDetailResponse"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
				).Return(errors.New("database connection failed"))
			},
			request: paymentModel.StaticVaDetailRequest{
				PaymentID:  "payment-123",
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
			result, err := repo.GetStaticVaDetail(context.Background(), tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.IsType(t, &paymentModel.StaticVaDetailResponse{}, result)
			}

			mysqlMock.AssertExpectations(t)
		})
	}
}

func TestGetStaticVaTransactions(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    paymentModel.StaticVaTransactionFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get static VA transactions",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaTransactionFilterRequest{
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
			name: "SUCCESS: Get static VA transactions with status filter",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaTransactionFilterRequest{
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
			name: "SUCCESS: Get static VA transactions with ID filter",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				ID:         "transaction-456",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "paymentDate",
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
					mock.Anything,
				).Return(errors.New("database connection failed"))

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
			wantErr: true,
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
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"SelectContext",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*[]paymentModel.StaticVaTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("select query failed"))
			},
			filter: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Get static VA transactions with SortBy createdAt",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaTransactionFilterRequest{
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
			name: "SUCCESS: Get static VA transactions with SortBy amount",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
			filter: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "amount",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get static VA transactions with amount population",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					// Populate the slice with transactions that have empty amount fields
					transactions := args.Get(1).(*[]paymentModel.StaticVaTransactionItem)
					*transactions = []paymentModel.StaticVaTransactionItem{
						{
							UUID:            "transaction-1",
							ReferenceID:     "payment-123",
							AmountValue:     "",
							AmountCurrency:  "",
							Status:          "SUCCESS",
						},
						{
							UUID:            "transaction-2",
							ReferenceID:     "payment-123",
							AmountValue:     "10000",
							AmountCurrency:  "",
							Status:          "SUCCESS",
						},
						{
							UUID:            "transaction-3",
							ReferenceID:     "payment-123",
							AmountValue:     "",
							AmountCurrency:  "USD",
							Status:          "SUCCESS",
						},
					}
				}).Return(nil)
			},
			filter: paymentModel.StaticVaTransactionFilterRequest{
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
			name: "SUCCESS: Get static VA transactions with sql.ErrNoRows",
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
					mock.AnythingOfType("*[]paymentModel.StaticVaTransactionItem"),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(sql.ErrNoRows)
			},
			filter: paymentModel.StaticVaTransactionFilterRequest{
				PaymentID:  "payment-123",
				MerchantID: "merchant-123",
				Page:       1,
				PerPage:    10,
				Sort:       "DESC",
				SortBy:     "paymentDate",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mysqlMock := &mysqlMocks.IMySqlExt{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mysqlMock)

			repo := paymentRepository.New(mysqlMock, logger)
			result, err := repo.GetStaticVaTransactions(context.Background(), tc.filter)

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
