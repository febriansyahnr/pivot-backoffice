package refundRepository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/refund"
	pdkLogMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExistsByClientReferenceAndMerchantID(t *testing.T) {
	clientRefID := uuid.NewString()
	merchantID := uuid.NewString()

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantExist bool
		wantErr   bool
	}{
		{
			name: "SUCCESS: Refund exists",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,               // ctx
					mock.AnythingOfType("*int"), // pointer to exists
					mock.AnythingOfType("string"),
					clientRefID,
					merchantID,
				).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*int)
					*ptr = 1
				}).Return(nil)
			},
			wantExist: true,
			wantErr:   false,
		},
		{
			name: "SUCCESS: Refund does not exist",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					clientRefID,
					merchantID,
				).Return(sql.ErrNoRows)
			},
			wantExist: false,
			wantErr:   false,
		},
		{
			name: "ERROR: Database failure",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*int"),
					mock.AnythingOfType("string"),
					clientRefID,
					merchantID,
				).Return(errors.New("db error"))
			},
			wantExist: false,
			wantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), pdkConst.CtxSQLTableNameKey, "refunds")
			exists, err := repo.ExistsByClientReferenceAndMerchantID(ctx, clientRefID, merchantID)

			assert.Equal(t, tc.wantExist, exists)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)
		})
	}
}

func TestExistsByClientReferenceAndMerchantID_QueryStatusExclusion(t *testing.T) {
	clientRefID := uuid.NewString()
	merchantID := uuid.NewString()

	// Capture the SQL once; every status assertion below inspects the same query.
	var capturedQuery string
	mockMysql := mysqlMocks.NewIMySqlExt(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	mockMysql.On(
		"GetContext",
		mock.Anything,               // ctx
		mock.AnythingOfType("*int"), // pointer to exists
		mock.AnythingOfType("string"),
		clientRefID,
		merchantID,
	).Run(func(args mock.Arguments) {
		capturedQuery = args.Get(2).(string)
	}).Return(sql.ErrNoRows)

	repo := New(mockMysql, mockLogger)
	ctx := context.WithValue(context.Background(), pdkConst.CtxSQLTableNameKey, "refunds")

	exists, err := repo.ExistsByClientReferenceAndMerchantID(ctx, clientRefID, merchantID)
	assert.NoError(t, err)
	assert.False(t, exists)
	mockMysql.AssertExpectations(t)

	// A FAILED refund must NOT reserve the client_reference_id, so it can be retried
	// with the same reference. PENDING / WAITING_BANK_TRANSFER stay excluded too;
	// only a SUCCESS refund reserves the reference.
	testCases := []struct {
		name         string
		status       string
		wantExcluded bool
	}{
		{
			name:         "PENDING is excluded (in-flight, reusable)",
			status:       "'PENDING'",
			wantExcluded: true,
		},
		{
			name:         "WAITING_BANK_TRANSFER is excluded (in-flight, reusable)",
			status:       "'WAITING_BANK_TRANSFER'",
			wantExcluded: true,
		},
		{
			name:         "FAILED is excluded (retryable with same reference)",
			status:       "'FAILED'",
			wantExcluded: true,
		},
		{
			name:         "SUCCESS is not excluded (reserves the reference)",
			status:       "'SUCCESS'",
			wantExcluded: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantExcluded {
				assert.Contains(t, capturedQuery, tc.status)
			} else {
				assert.NotContains(t, capturedQuery, tc.status)
			}
		})
	}
}

func TestFindByID(t *testing.T) {
	refundID := uuid.NewString()

	mockRefund := refundModel.Refund{
		UUID:              refundID,
		MerchantID:        uuid.NewString(),
		ClientReferenceID: "client-ref-123",
		PaymentID:         uuid.NewString(),
		Currency:          constant.CurrencyIDR,
		Amount:            10000.00,
		Status:            constant.RefundStatusPending,
		Reason:            "DUPLICATE",
		Description:       "Refund for duplicate transaction",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantNil   bool
		wantErr   bool
	}{
		{
			name: "SUCCESS: Refund found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.Refund"),
					mock.AnythingOfType("string"),
					refundID,
				).Run(func(args mock.Arguments) {
					ref := args.Get(1).(*refundModel.Refund)
					*ref = mockRefund
				}).Return(nil)
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "SUCCESS: Refund not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.Refund"),
					mock.AnythingOfType("string"),
					refundID,
				).Return(sql.ErrNoRows)
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.Refund"),
					mock.AnythingOfType("string"),
					refundID,
				).Return(errors.New("database error"))
			},
			wantNil: true,
			wantErr: true,
		},
		{
			name: "SUCCESS: Refund found with valid metadata",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.Refund"),
					mock.AnythingOfType("string"),
					refundID,
				).Run(func(args mock.Arguments) {
					ref := args.Get(1).(*refundModel.Refund)
					*ref = mockRefund
					ref.Metadata = types.NullJSONText{
						JSONText: []byte(`{"key":"value","test":true}`),
						Valid:    true,
					}
				}).Return(nil)
			},
			wantNil: false,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), pdkConst.CtxSQLTableNameKey, "refunds")
			result, err := repo.FindByID(ctx, refundID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, mockRefund.UUID, result.UUID)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestGetRefundList(t *testing.T) {
	now := time.Now()
	refundResp := &refundModel.RefundResponse{
		ID:                uuid.NewString(),
		ClientReferenceID: "client-ref-001",
		PaymentSessionID:  uuid.NewString(),
		ChargeID:          uuid.NewString(),
		Amount: commonModel.Amount{
			Value:    "100000.00",
			Currency: constant.CurrencyIDR,
		},
		CapturedAmount: commonModel.Amount{
			Value:    "100000.00",
			Currency: constant.CurrencyIDR,
		},
		IsFullAmount:    true,
		Status:          constant.RefundStatusSuccess,
		Reason:          "DUPLICATE",
		Description:     "Refund issued",
		DestinationType: "bank_account",
		Method:          "manual",
		FailureCode:     "NONE",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	filter := refundModel.FilterRefundRequest{
		Page:       1,
		PerPage:    10,
		MerchantID: "merchant-id",
		Sort:       "desc",
		SortBy:     "created_at",
	}

	tests := []struct {
		name           string
		mockSetup      func(mockDB *mysqlMocks.IMySqlExt)
		expectedErr    bool
		expectedResult []*refundModel.RefundResponse
		expectedTotal  int64
	}{
		{
			name: "success - return list and count",
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On("GetContext", mock.Anything, mock.AnythingOfType("*int64"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ptr := args.Get(1).(*int64)
						*ptr = 1
					}).Return(nil)

				mockDB.On("SelectContext", mock.Anything, mock.AnythingOfType("*[]*refundModel.RefundResponse"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ptr := args.Get(1).(*[]*refundModel.RefundResponse)
						*ptr = []*refundModel.RefundResponse{refundResp}
					}).Return(nil)
			},
			expectedErr:    false,
			expectedResult: []*refundModel.RefundResponse{refundResp},
			expectedTotal:  1,
		},
		{
			name: "error - count query fails",
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On("GetContext", mock.Anything, mock.AnythingOfType("*int64"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("db error"))

				mockDB.On("SelectContext", mock.Anything, mock.AnythingOfType("*[]*refundModel.RefundResponse"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ptr := args.Get(1).(*[]*refundModel.RefundResponse)
						*ptr = []*refundModel.RefundResponse{refundResp}
					}).Return(nil)
			},
			expectedErr:    true,
			expectedResult: nil,
			expectedTotal:  0,
		},
		{
			name: "error - select query fails",
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On("GetContext", mock.Anything, mock.AnythingOfType("*int64"), mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ptr := args.Get(1).(*int64)
						*ptr = 1
					}).Return(nil)

				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("select error"))
			},
			expectedErr:    true,
			expectedResult: nil,
			expectedTotal:  0,
		},
		{
			name: "success - count query returns sql.ErrNoRows",
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On("GetContext", mock.Anything, mock.AnythingOfType("*int64"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sql.ErrNoRows)

				mockDB.On("SelectContext", mock.Anything, mock.AnythingOfType("*[]*refundModel.RefundResponse"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ptr := args.Get(1).(*[]*refundModel.RefundResponse)
						*ptr = []*refundModel.RefundResponse{}
					}).Return(nil)
			},
			expectedErr:    false,
			expectedResult: []*refundModel.RefundResponse{},
			expectedTotal:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.mockSetup(mockDB)

			repo := New(mockDB, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "refunds")
			res, err := repo.GetRefundList(ctx, filter)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.expectedTotal, res.Meta.TotalItems)
				assert.Len(t, res.Data, len(tt.expectedResult))
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestFindRefundByChargeID(t *testing.T) {
	chargeID := uuid.NewString()

	mockRefund := &refundModel.Refund{
		UUID:              uuid.NewString(),
		MerchantID:        uuid.NewString(),
		ClientReferenceID: "client-ref-123",
		PaymentID:         uuid.NewString(),
		PaymentChargeID:   chargeID,
		Currency:          "IDR",
		Amount:            10000,
		Status:            "PENDING",
		Reason:            "duplicate",
		Description:       "Duplicate transaction",
		DestinationType:   "BANK_ACCOUNT",
		Method:            "MANUAL",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Metadata:          types.NullJSONText{Valid: false},
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		want      *refundModel.Refund
		wantErr   bool
	}{
		{
			name: "SUCCESS: Refund found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.Refund"),
					mock.AnythingOfType("string"),
					chargeID,
				).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*refundModel.Refund)
					*arg = *mockRefund
				}).Return(nil)
			},
			want:    mockRefund,
			wantErr: false,
		},
		{
			name: "SUCCESS: Refund not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.Refund"),
					mock.AnythingOfType("string"),
					chargeID,
				).Return(sql.ErrNoRows)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "ERROR: DB failure",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.Refund"),
					mock.AnythingOfType("string"),
					chargeID,
				).Return(errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), pdkConst.CtxSQLTableNameKey, "refunds")
			result, err := repo.FindRefundByChargeID(ctx, chargeID)

			assert.Equal(t, tc.want, result)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)
		})
	}
}

func TestGetRefundListWithDefaultsAndFilters(t *testing.T) {
	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	testCases := []struct {
		name      string
		filter    refundModel.FilterRefundRequest
		mockSetup func(mockDB *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Default values for Page, PerPage, Sort, SortBy",
			filter: refundModel.FilterRefundRequest{
				MerchantID: "merchant-123",
				// Page and PerPage not set (will default)
				// Sort and SortBy not set (will default)
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On("GetContext", mock.Anything, mock.AnythingOfType("*int64"), mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ptr := args.Get(1).(*int64)
						*ptr = 5
					}).Return(nil)

				mockDB.On("SelectContext", mock.Anything, mock.AnythingOfType("*[]*refundModel.RefundResponse"), mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ptr := args.Get(1).(*[]*refundModel.RefundResponse)
						*ptr = []*refundModel.RefundResponse{}
					}).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: With all filters - UUID, PaymentSessionID, ClientReferenceID, Status, dates",
			filter: refundModel.FilterRefundRequest{
				MerchantID:        "merchant-123",
				UUID:              uuid.NewString(),
				PaymentSessionID:  uuid.NewString(),
				ClientReferenceID: "client-ref-001",
				Status:            constant.RefundStatusSuccess,
				StartCreatedAt:    &startDate,
				EndCreatedAt:      &endDate,
				Page:              1,
				PerPage:           10,
				Sort:              "asc",
				SortBy:            "updated_at",
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				// Count query
				mockDB.On("GetContext", mock.Anything, mock.AnythingOfType("*int64"), mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ptr := args.Get(1).(*int64)
						*ptr = 10
					}).Return(nil)

				// Select query
				mockDB.On("SelectContext", mock.Anything, mock.AnythingOfType("*[]*refundModel.RefundResponse"), mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						ptr := args.Get(1).(*[]*refundModel.RefundResponse)
						*ptr = []*refundModel.RefundResponse{}
					}).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockDB)

			repo := New(mockDB, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "refunds")
			result, err := repo.GetRefundList(ctx, tc.filter)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetRefundByID(t *testing.T) {
	refundID := uuid.NewString()
	merchantID := uuid.NewString()

	mockRefundResponse := refundModel.RefundResponse{
		ID:              refundID,
		MerchantID:      merchantID,
		Status:          constant.RefundStatusSuccess,
		DestinationType: "CHANNEL",
		PaymentChannel:  "BNC",
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantNil   bool
		wantErr   bool
	}{
		{
			name: "SUCCESS: Refund found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.RefundResponse"),
					mock.AnythingOfType("string"),
					refundID,
					merchantID,
				).Run(func(args mock.Arguments) {
					ref := args.Get(1).(*refundModel.RefundResponse)
					*ref = mockRefundResponse
				}).Return(nil)
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "SUCCESS: Refund not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.RefundResponse"),
					mock.AnythingOfType("string"),
					refundID,
					merchantID,
				).Return(sql.ErrNoRows)
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.AnythingOfType("*refundModel.RefundResponse"),
					mock.AnythingOfType("string"),
					refundID,
					merchantID,
				).Return(errors.New("database error"))
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "refunds")
			result, err := repo.GetRefundByID(ctx, refundID, merchantID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, refundID, result.ID)
				assert.Equal(t, merchantID, result.MerchantID)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestListByPaymentID(t *testing.T) {
	now := time.Now().UTC()
	paymentID := uuid.NewString()

	db := mysqlMocks.NewIMySqlExt(t)

	mockRefundResponses := []refundModel.RefundResponse{
		{
			ID:                uuid.NewString(),
			ClientReferenceID: "client-ref-001",
			PaymentSessionID:  paymentID,
			ChargeID:          uuid.NewString(),
			Amount: commonModel.Amount{
				Value:    "10000.00", // NOSONAR
				Currency: "IDR",      // NOSONAR
			},
			CapturedAmount: commonModel.Amount{
				Value:    "100000.00", // NOSONAR
				Currency: "IDR",       // NOSONAR
			},
			IsFullAmount:    false,
			Status:          constant.RefundStatusSuccess,
			Reason:          "REQUESTED_BY_CUSTOMER", // NOSONAR
			Description:     "Refund issued",         // NOSONAR
			DestinationType: "ACCOUNT",               // NOSONAR
			Method:          "TRANSFER_ONLY",         // NOSONAR
			FailureCode:     "",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:                uuid.NewString(),
			ClientReferenceID: "client-ref-002", // NOSONAR
			PaymentSessionID:  paymentID,
			ChargeID:          uuid.NewString(),
			Amount: commonModel.Amount{
				Value:    "50000.00", // NOSONAR
				Currency: "IDR",      // NOSONAR
			},
			CapturedAmount: commonModel.Amount{
				Value:    "100000.00", // NOSONAR
				Currency: "IDR",       // NOSONAR
			},
			IsFullAmount:    false,
			Status:          constant.RefundStatusSuccess,
			Reason:          "DUPLICATE",         // NOSONAR
			Description:     "Duplicate payment", // NOSONAR
			DestinationType: "ACCOUNT",           // NOSONAR
			Method:          "TRANSFER_ONLY",     // NOSONAR
			FailureCode:     "",
			CreatedAt:       now.Add(-1 * time.Hour),
			UpdatedAt:       now.Add(-1 * time.Hour),
		},
	}

	testCases := []struct {
		name           string
		request        refundModel.ListByPaymentIDRequest
		mockSetup      func()
		expectedResult []refundModel.RefundResponse
		expectedError  error
	}{
		{
			name: "SUCCESS: Get refund list by payment ID",
			request: refundModel.ListByPaymentIDRequest{
				Status: "",
			},
			mockSetup: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentID,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]refundModel.RefundResponse) = mockRefundResponses
				}).Once().Return(nil)
			},
			expectedResult: mockRefundResponses,
		},
		{
			name: "SUCCESS: Get refund list by payment ID with status filter",
			request: refundModel.ListByPaymentIDRequest{
				Status: constant.RefundStatusSuccess,
			},
			mockSetup: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentID, constant.RefundStatusSuccess,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*[]refundModel.RefundResponse) = mockRefundResponses
				}).Return(nil)
			},
			expectedResult: mockRefundResponses,
		},
		{
			name: "SUCCESS: Empty refund list",
			request: refundModel.ListByPaymentIDRequest{
				Status: "",
			},
			mockSetup: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentID,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*[]refundModel.RefundResponse) = []refundModel.RefundResponse{}
				}).Once().Return(nil)
			},
			expectedResult: []refundModel.RefundResponse{},
		},
		{
			name: "ERROR: Database failure",
			request: refundModel.ListByPaymentIDRequest{
				Status: "",
			},
			mockSetup: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentID,
				).Once().Return(assert.AnError)
				// logger.On("Error", mock.Anything, "failed to get total refunded amount for payment_id:"+paymentID, mock.Anything).Once().Return()
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			repo := New(db, nil)
			result, err := repo.ListByPaymentID(t.Context(), paymentID, tc.request)
			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)

			db.AssertExpectations(t)
		})
	}
}

func TestGetTotalRefundedAmount(t *testing.T) {
	paymentID := uuid.NewString()

	logger := pdkLogMock.NewILogger(t)
	db := mysqlMocks.NewIMySqlExt(t)

	testCases := []struct {
		name           string
		mockSetup      func()
		expectedAmount float64
		expectedError  error
	}{
		{
			name: "SUCCESS: Get total refunded amount for payment",
			mockSetup: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, paymentID,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*float64) = 60000.00
				}).Once().Return(nil)
			},
			expectedAmount: 60000.00,
		},
		{
			name: "SUCCESS: No refunds for payment (zero amount)",
			mockSetup: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, paymentID,
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*float64) = 0.00
				}).Once().Return(nil)
			},
			expectedAmount: 0.00,
		},
		{
			name: "ERROR: Database failure",
			mockSetup: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, paymentID,
				).Once().Return(assert.AnError)
				logger.On("Error", mock.Anything, "failed to get total refunded amount for payment_id: "+paymentID, mock.Anything).Once().Return()
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			repo := New(db, logger)

			totalAmount, err := repo.GetTotalRefundedAmount(t.Context(), paymentID)
			assert.Equal(t, tc.expectedAmount, totalAmount)
			assert.Equal(t, tc.expectedError, err)

			db.AssertExpectations(t)
			logger.AssertExpectations(t)
		})
	}
}
