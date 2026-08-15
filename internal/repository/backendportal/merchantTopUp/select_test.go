package merchantTopUp

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchantTopUp"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByReferenceNumber(t *testing.T) {

	data := &model.MerchantTopUp{
		ID:              "uuid-uuid-uuid",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		ReferenceNumber: "reference-number",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	buf := new(bytes.Buffer)
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		expected  *model.MerchantTopUp
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find merchant top up reference by reference_number",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*model.MerchantTopUp) = *data
				})
			},
			input: data.ReferenceNumber, expected: data,
		},
		{
			name: "ERROR: Merchant top up reference not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(sql.ErrNoRows)
			},
			input: data.ReferenceNumber,
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			input:   data.ReferenceNumber,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, log)

			dataRes, err := repo.GetByReferenceNumber(context.Background(), tc.input)
			if tc.wantErr {
				assert.Error(t, err)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, dataRes)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestGetByMerchantAccountNameAndPaymentMethodId(t *testing.T) {

	data := &model.MerchantTopUp{
		ID:              "uuid-uuid-uuid",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		ReferenceNumber: "reference-number",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	buf := new(bytes.Buffer)
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt)
		merchantId string
		paymentId  string
		expected   *model.MerchantTopUp
		wantErr    bool
	}{
		{
			name: "SUCCESS: Find merchant top up reference by merchant_id and payment_method_id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*model.MerchantTopUp) = *data
				})
			},
			merchantId: "merchant-id",
			paymentId:  "payment-method-id",
			expected:   data,
		},
		{
			name: "ERROR: merchant top up reference Not Found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(sql.ErrNoRows)
			},
			merchantId: "merchant-id",
			paymentId:  "payment-method-id",
		},
		{
			name: "ERROR: Database Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			merchantId: "merchant-id",
			paymentId:  "payment-method-id",
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, log)

			dataRes, err := repo.GetByMerchantAccountNameAndPaymentMethodId(context.Background(), tc.merchantId, constant.ReferenceDisbursement, tc.paymentId)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, dataRes)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestGetList(t *testing.T) {
	now := time.Now()
	startDate := now.AddDate(0, -1, 0)
	endDate := now

	topUpList := []model.TopUpTransactionResponse{
		{
			UUID:                "transaction-uuid-1",
			ReferenceID:         "reference-uuid-1",
			MerchantReferenceID: "merchant-ref-1",
			Type:                "VA_TOP_UP",
			Channel:             "Bank Mandiri",
			Date:                now,
			Amount:              10000,
			Status:              "SUCCESS",
			BalanceType:         "Payout Balance",
		},
		{
			UUID:                "transaction-uuid-2",
			ReferenceID:         "reference-uuid-2",
			MerchantReferenceID: "-",
			Type:                "MANUAL_TOP_UP",
			Channel:             "MANUAL_TRANSFER",
			Date:                now,
			Amount:              50000,
			Status:              "SUCCESS",
			BalanceType:         "Payment Balance",
		},
	}

	buf := new(bytes.Buffer)
	log := logger.NewSlogger(logger.Config{}, logger.WithSlogOutput(buf))

	// Mock app config for testing with appConfig
	mockAppConfig := &config.AppConfig{
		UseOverFetchPagination: true,
		InitialPageWindow:      3,
	}

	testCases := []struct {
		name         string
		mockSetup    func(mysqlMock *mysqlMocks.IMySqlExt)
		request      *model.TopUpTransactionListRequest
		wantErr      bool
		useAppConfig bool
	}{
		{
			name: "SUCCESS: Get top-up transaction list",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// Mock SelectContext for getting the list
				mysqlMock.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]model.TopUpTransactionResponse) = topUpList
				}).Once()

				// Mock GetContext for counting total rows
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = int64(len(topUpList))
				}).Once()
			},
			request: &model.TopUpTransactionListRequest{
				MerchantId: "merchant-id",
				StartDate:  startDate,
				EndDate:    endDate,
				Page:       1,
				PerPage:    10,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get top-up transaction list with filters",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]model.TopUpTransactionResponse) = topUpList[:1]
				}).Once()

				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = int64(1)
				}).Once()
			},
			request: &model.TopUpTransactionListRequest{
				MerchantId:    "merchant-id",
				StartDate:     startDate,
				EndDate:       endDate,
				Status:        "SUCCESS",
				TransactionID: "transaction-uuid-1",
				Page:          1,
				PerPage:       10,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Empty result (sql.ErrNoRows)",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(sql.ErrNoRows).Once()

				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = int64(0)
				}).Once()
			},
			request: &model.TopUpTransactionListRequest{
				MerchantId: "merchant-id",
				StartDate:  startDate,
				EndDate:    endDate,
				Page:       1,
				PerPage:    10,
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error on SelectContext",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(constant.ErrSomeErrorForUnitTest).Once()

				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = int64(0)
				}).Once()
			},
			request: &model.TopUpTransactionListRequest{
				MerchantId: "merchant-id",
				StartDate:  startDate,
				EndDate:    endDate,
				Page:       1,
				PerPage:    10,
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Database error on GetContext (count) is ignored",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]model.TopUpTransactionResponse) = topUpList
				}).Once()

				// Count query fails, but pagination utility ignores this error
				mysqlMock.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			request: &model.TopUpTransactionListRequest{
				MerchantId: "merchant-id",
				StartDate:  startDate,
				EndDate:    endDate,
				Page:       1,
				PerPage:    10,
			},
			wantErr: false, // Changed from true - count errors are now ignored
		},
		{
			name: "SUCCESS: Get top-up transaction list with appConfig",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// Mock SelectContext for getting the list (overfetch will fetch more than perPage)
				mysqlMock.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]model.TopUpTransactionResponse) = topUpList
				}).Once()
			},
			request: &model.TopUpTransactionListRequest{
				MerchantId: "merchant-id",
				StartDate:  startDate,
				EndDate:    endDate,
				Page:       1,
				PerPage:    10,
			},
			wantErr:      false,
			useAppConfig: true, // Test with appConfig to cover WithAppConfig code path
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			mockMysql := mysqlMocks.NewIMySqlExt(t)
			tc.mockSetup(mockMysql)

			var repo repository.IMerchantTopUpRepository
			if tc.useAppConfig {
				repo = New(mockMysql, log, WithAppConfig(mockAppConfig))
			} else {
				repo = New(mockMysql, log)
			}

			result, err := repo.GetList(context.Background(), tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Meta)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
