package withdrawalService_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/google/uuid"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/withdrawal"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIWithdrawalRepository(t)

	service := New(logger, repo, nil, nil, nil)

	response := &commonModel.PaginationResponse{
		Data: []withdrawal.WithdrawalHistoryResponse{{
			Id:                     "1",
			Amount:                 2,
			BeneficiaryBankName:    "3",
			BeneficiaryAccountName: "4",
			Status:                 "5",
		}},
		Meta: commonModel.Meta{},
	}
	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *commonModel.PaginationResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetList", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalHistoryRequest"), // NOSONAR
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetList", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalHistoryRequest"), // NOSONAR
				).Return(response, nil)
			},
			wantResult: response,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetList(context.Background(), &withdrawal.WithdrawalHistoryRequest{})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetById(t *testing.T) {
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	repo := repoMocks.NewIWithdrawalRepository(t)

	service := New(logger, repo, nil, nil, nil)

	result := withdrawal.WithdrawalDetailResponse{
		Id:                     "1",
		CreatedBy:              "2",
		Amount:                 3,
		Status:                 "4",
		BankReferenceNo:        "5",
		BeneficiaryBankName:    "6",
		BeneficiaryAccountNo:   "7",
		BeneficiaryAccountName: "8",
	}
	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *withdrawal.WithdrawalDetailResponse
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetById", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetById", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrDataNotFound),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetById", c.ValueCtxMockType(), mock.Anything,
				).Return(&result, nil)
			},
			wantResult: &result,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetById(context.Background(), &withdrawal.WithdrawalDetailRequest{})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestGetTodayWithdrawalInsight(t *testing.T) {
	var (
		ctx                = context.Background()
		mockWithdrawalRepo repoMocks.IWithdrawalRepository
		mockAccountRepo    repoMocks.IAccountRepository
		withdrawalService  = New(nil, &mockWithdrawalRepo, nil, nil, nil, WithAccountRepository(&mockAccountRepo))
		validMerchantID    = uuid.NewString()
		defaultInsight     = &withdrawal.WithdrawalInsightItem{
			Total: 0,
			TotalAmount: commonModel.Amount{
				Currency: "IDR",
				Value:    strconv.FormatFloat(0, 'f', 2, 64),
			},
		}
	)
	testCases := []struct {
		name      string
		payload   withdrawal.WithdrawalInsightRequest
		callMock  func()
		want      *withdrawal.WithdrawalInsightResponse
		wantErr   error
		shouldErr bool
	}{
		{
			name: "when failed to get withdrawal insight, then should return error",
			payload: withdrawal.WithdrawalInsightRequest{
				MerchantID: validMerchantID,
				Status:     c.StatusSuccess,
			},
			callMock: func() {
				mockWithdrawalRepo.On("GetTodayWithdrawalInsight", mock.Anything, withdrawal.WithdrawalInsightRequest{
					MerchantID: validMerchantID,
					Status:     c.StatusSuccess,
				}).Return(nil, errors.New("invalid arg of insight")).Once()
			},
			shouldErr: true,
			wantErr:   errors.New("invalid arg of insight"),
		},
		{
			name: "when withdrawal insight found, then return the insight with the currency",
			payload: withdrawal.WithdrawalInsightRequest{
				MerchantID: validMerchantID,
				Status:     c.StatusSuccess,
			},
			callMock: func() {
				mockWithdrawalRepo.On("GetTodayWithdrawalInsight", mock.Anything, withdrawal.WithdrawalInsightRequest{
					MerchantID: validMerchantID,
					Status:     c.StatusSuccess,
				}).Return(&withdrawal.WithdrawalInsightResponse{
					TodayTotalSuccess: defaultInsight,
					TodayTotalPending: defaultInsight,
					TodayTotalFailed:  defaultInsight,
				}, nil).Once()
			},
			want: &withdrawal.WithdrawalInsightResponse{
				TodayTotalSuccess: defaultInsight,
				TodayTotalPending: defaultInsight,
				TodayTotalFailed:  defaultInsight,
			},
		},
		{
			name: "when insight not found, then should return nil",
			payload: withdrawal.WithdrawalInsightRequest{
				MerchantID: validMerchantID,
				Status:     c.StatusSuccess,
			},
			callMock: func() {
				mockWithdrawalRepo.On("GetTodayWithdrawalInsight", mock.Anything, withdrawal.WithdrawalInsightRequest{
					MerchantID: validMerchantID,
					Status:     c.StatusSuccess,
				}).Return(nil, nil).Once()
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			insight, err := withdrawalService.GetTodayWithdrawalInsight(ctx, tc.payload)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, insight)
		})
	}
}
