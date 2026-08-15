package withdrawalRepository_test

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/withdrawal"
	mysqlMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db)

	request := &withdrawal.WithdrawalHistoryRequest{
		WithdrawalListRequest: &withdrawal.WithdrawalListRequest{
			Status: "PENDING", Id: uuid.NewString(), Sort: "-date",
		},
		Page:    1,
		PerPage: 10,
	}
	response := withdrawal.WithdrawalHistoryResponse{
		Id:                     "1",
		Amount:                 2,
		BeneficiaryBankName:    "3",
		BeneficiaryAccountName: "4",
		Status:                 "5",
	}
	resultMockType := mock.AnythingOfType("*[]withdrawal.WithdrawalHistoryResponse")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *commonModel.PaginationResponse
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"SelectContext", c.CancelCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				db.On(
					"GetContext", c.CancelCtxMockType(), mock.AnythingOfType("*int64"), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Data not found",
			setupMock: func() {
				db.On(
					"SelectContext", c.CancelCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
				db.On(
					"GetContext", c.CancelCtxMockType(), mock.AnythingOfType("*int64"), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil)
			},
			wantResult: &commonModel.PaginationResponse{
				Data: []withdrawal.WithdrawalHistoryResponse{},
				Meta: commonModel.Meta{
					Page: 1, PerPage: 10,
				},
			},
		},
		{
			name: "SUCCESS:Data found",
			setupMock: func() {
				db.On(
					"SelectContext", c.CancelCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]withdrawal.WithdrawalHistoryResponse)) = []withdrawal.WithdrawalHistoryResponse{response}
				}).Return(nil)
				db.On(
					"GetContext", c.CancelCtxMockType(), mock.AnythingOfType("*int64"), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.TimeMockType(), c.TimeMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 1
				}).Return(nil)
			},
			wantResult: &commonModel.PaginationResponse{
				Data: []withdrawal.WithdrawalHistoryResponse{response},
				Meta: commonModel.Meta{
					Page: 1, PerPage: 10, TotalItems: 1, TotalPages: 1,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetList(context.Background(), request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetById(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db)

	response := withdrawal.WithdrawalDetailResponse{
		Id:                     "1",
		CreatedBy:              "2",
		Amount:                 3,
		Status:                 "4",
		BankReferenceNo:        "5",
		BeneficiaryBankName:    "6",
		BeneficiaryAccountNo:   "7",
		BeneficiaryAccountName: "8",
	}
	resultMockType := mock.AnythingOfType("*withdrawal.WithdrawalDetailResponse")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *withdrawal.WithdrawalDetailResponse
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr:    c.ErrSomeErrorForUnitTest,
			wantResult: &withdrawal.WithdrawalDetailResponse{},
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*withdrawal.WithdrawalDetailResponse)) = response
				}).Return(nil)
			},
			wantResult: &response,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetById(context.Background(), &withdrawal.WithdrawalDetailRequest{})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetByReferenceId(t *testing.T) {
	db := mysqlMock.NewIMySqlExt(t)

	repo := New(db)

	response := withdrawal.WithdrawalDetailResponse{
		Id:                     "1",
		CreatedBy:              "2",
		Amount:                 3,
		Status:                 "4",
		BankReferenceNo:        "5",
		BeneficiaryBankName:    "6",
		BeneficiaryAccountNo:   "7",
		BeneficiaryAccountName: "8",
	}
	resultMockType := mock.AnythingOfType("*withdrawal.WithdrawalDetailResponse")

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *withdrawal.WithdrawalDetailResponse
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr:    c.ErrSomeErrorForUnitTest,
			wantResult: &withdrawal.WithdrawalDetailResponse{},
		},
		{
			name: "ERROR:Data not found",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(sql.ErrNoRows)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				db.On(
					"GetContext", c.ValueCtxMockType(), resultMockType, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*withdrawal.WithdrawalDetailResponse)) = response
				}).Return(nil)
			},
			wantResult: &response,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetByReferenceId(context.Background(), uuid.NewString(), uuid.NewString())
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetTodayWithdrawalStatusInsight(t *testing.T) {
	var (
		ctx               = context.Background()
		mockDB            = mysqlMock.IMySqlExt{}
		withdrawalRepo    = New(&mockDB)
		validMerchantID   = "valid-merchant-id"
		invalidMerchantID = "invalid-merchant-id"

		loc, _         = time.LoadLocation(c.TimeLoc)
		now            = time.Now().In(loc)
		startOfDay     = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
		defaultInsight = &withdrawal.WithdrawalInsightItem{
			Total: 0,
			TotalAmount: commonModel.Amount{
				Currency: c.CurrencyIDR,
				Value:    strconv.FormatFloat(0, 'f', 2, 64),
			},
		}
	)
	testCases := []struct {
		name      string
		payload   withdrawal.WithdrawalInsightRequest
		callMock  func(t *testing.T)
		want      *withdrawal.WithdrawalInsightResponse
		wantErr   error
		shouldErr bool
	}{
		{
			name: "when merchant is exist, then should return correct data",
			payload: withdrawal.WithdrawalInsightRequest{
				MerchantID: validMerchantID,
				Status:     c.StatusSuccess,
			},
			callMock: func(t *testing.T) {
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, validMerchantID, startOfDay).
					Return(nil).
					Run(func(args mock.Arguments) {
						*args.Get(1).(*[]withdrawal.WithdrawalInsightQuery) = []withdrawal.WithdrawalInsightQuery{
							{
								Total:       2,
								TotalAmount: 200000,
								Status:      c.StatusSuccess,
								Currency:    c.CurrencyIDR,
							},
							{
								Total:       0,
								TotalAmount: 0,
								Status:      c.StatusPending,
								Currency:    c.CurrencyIDR,
							},
							{
								Total:       0,
								TotalAmount: 0,
								Status:      c.StatusFailed,
								Currency:    c.CurrencyIDR,
							},
						}

					}).Once()
			},
			want: &withdrawal.WithdrawalInsightResponse{
				TodayTotalSuccess: &withdrawal.WithdrawalInsightItem{
					Total: 2,
					TotalAmount: commonModel.Amount{
						Currency: c.CurrencyIDR,
						Value:    strconv.FormatFloat(200000, 'f', 2, 64),
					},
				},
				TodayTotalPending: defaultInsight,
				TodayTotalFailed:  defaultInsight,
			},
		},
		{
			name: "when database was down, then should return error",
			payload: withdrawal.WithdrawalInsightRequest{
				MerchantID: invalidMerchantID,
				Status:     c.StatusSuccess,
			},
			callMock: func(t *testing.T) {
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, invalidMerchantID, startOfDay).
					Return(errors.New("database down")).Once()
			},
			shouldErr: true,
			wantErr:   errors.New("database down"),
		},
		{
			name: "when no merchant withdrawal exist, then should not return error",
			payload: withdrawal.WithdrawalInsightRequest{
				MerchantID: validMerchantID,
				Status:     c.StatusPending,
			},
			callMock: func(t *testing.T) {
				mockDB.On("SelectContext", mock.Anything, mock.Anything, mock.Anything, validMerchantID, startOfDay).
					Return(sql.ErrNoRows).Once()
			},
			want: &withdrawal.WithdrawalInsightResponse{
				TodayTotalSuccess: defaultInsight,
				TodayTotalPending: defaultInsight,
				TodayTotalFailed:  defaultInsight,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock(t)

			insight, err := withdrawalRepo.GetTodayWithdrawalInsight(ctx, tc.payload)

			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, insight)
		})
	}
}
