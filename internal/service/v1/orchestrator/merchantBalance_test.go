package orchestrator_service_test

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	accountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/orchestrator"
	logMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMerchantBalance(t *testing.T) {
	logger := logMock.NewILogger(t)
	accountRepo := repositoryMocks.NewIAccountRepository(t)
	accountTransactionRepo := repositoryMocks.NewIAccountTransactionRepository(t)

	service := New(logger, nil, accountTransactionRepo, accountRepo)

	merchantID := "12f513ca-d538-412a-92a2-6a02344d9b6c"
	merchantUUID, _ := uuid.Parse(merchantID)
	now := time.Now().UTC()

	paymentAccount := &accountModel.Account{
		UUID:                uuid.New(),
		ReferenceID:         merchantUUID,
		Name:                constant.AccountNamePayment,
		EODBalance:          100_000.00,
		Currency:            constant.CurrencyIDR,
		LastUpdateBalanceAt: now.Add(-24 * time.Hour),
	}
	disbursementAccount := &accountModel.Account{
		UUID:                uuid.New(),
		ReferenceID:         merchantUUID,
		Name:                constant.AccountNameDisbursement,
		EODBalance:          200_000.00,
		Currency:            constant.CurrencyIDR,
		LastUpdateBalanceAt: now.Add(-24 * time.Hour),
	}
	validAggregateResponse := &orchestratorModel.AggregateResponse{
		SumOfCredit:     50_000.00,
		SumOfDebit:      10_000.00,
		SumOfPendCredit: 20_000.00,
		SumOfPendDebit:  5_000.00,
	}

	tests := []struct {
		name       string
		request    orchestratorModel.GetMerchantBalanceRequest
		setupMock  func()
		wantErr    error
		wantResult *orchestratorModel.GetMerchantBalanceResponse
	}{
		{
			name: "ERROR: invalid merchant ID format",
			request: orchestratorModel.GetMerchantBalanceRequest{
				MerchantID: "invalid-uuid",
			},
			setupMock: func() {
				logger.On("Error", mock.Anything, "Failed while parsing merchant ID to UUID format", mock.Anything).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid),
		},
		{
			name: "ERROR: FindMerchantAccountByName returns error",
			request: orchestratorModel.GetMerchantBalanceRequest{
				MerchantID:  merchantID,
				BalanceName: constant.AccountNamePayment,
				Date:        now,
			},
			setupMock: func() {
				accountRepo.On(
					"FindMerchantAccountByName", mock.Anything, merchantUUID, constant.AccountNamePayment,
				).Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to find merchant account by name", mock.Anything).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS: account not found returns zero balance",
			request: orchestratorModel.GetMerchantBalanceRequest{
				MerchantID:  merchantID,
				BalanceName: constant.AccountNamePayment,
				Date:        now,
			},
			setupMock: func() {
				accountRepo.On(
					"FindMerchantAccountByName", mock.Anything, merchantUUID, constant.AccountNamePayment,
				).Once().Return(nil, nil)
			},
			wantResult: &orchestratorModel.GetMerchantBalanceResponse{
				AvailableBalance: commonModel.Amount{Currency: constant.CurrencyIDR, Value: "0.00"},
				PendingBalance:   commonModel.Amount{Currency: constant.CurrencyIDR, Value: "0.00"},
			},
		},
		{
			name: "ERROR: GetAggregateTransactions returns error",
			request: orchestratorModel.GetMerchantBalanceRequest{
				MerchantID:  merchantID,
				BalanceName: constant.AccountNameDisbursement,
				Date:        now,
			},
			setupMock: func() {
				accountRepo.On(
					"FindMerchantAccountByName", mock.Anything, merchantUUID, constant.AccountNameDisbursement,
				).Once().Return(disbursementAccount, nil)
				accountTransactionRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed while get aggregate transactions", mock.Anything).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR: CalculatePendingBalance returns error",
			request: orchestratorModel.GetMerchantBalanceRequest{
				MerchantID:  merchantID,
				BalanceName: constant.AccountNamePayment,
				Date:        now,
			},
			setupMock: func() {
				accountRepo.On(
					"FindMerchantAccountByName", mock.Anything, merchantUUID, constant.AccountNamePayment,
				).Once().Return(paymentAccount, nil)
				accountTransactionRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).Once().Return(validAggregateResponse, nil)
				accountTransactionRepo.On("CalculatePendingBalance", mock.Anything, mock.Anything).Once().Return(0.0, assert.AnError)
				logger.On("Error", mock.Anything, "Failed while calculate pending balance", mock.Anything).Once().Return()
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS: balance without pending calculation (disbursement account)",
			request: orchestratorModel.GetMerchantBalanceRequest{
				MerchantID:  merchantID,
				BalanceName: constant.AccountNameDisbursement,
				Date:        now,
			},
			setupMock: func() {
				accountRepo.On(
					"FindMerchantAccountByName", mock.Anything, merchantUUID, constant.AccountNameDisbursement,
				).Once().Return(disbursementAccount, nil)
				accountTransactionRepo.On("GetAggregateTransactions", mock.Anything, mock.Anything).Once().Return(validAggregateResponse, nil)
			},
			wantResult: &orchestratorModel.GetMerchantBalanceResponse{
				AvailableBalance: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    "240000.00",
				},
				PendingBalance: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    "15000.00",
				},
				TotalBalance: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    "255000.00",
				},
			},
		},
		{
			name: "SUCCESS: balance with pending calculation (payment account)",
			request: orchestratorModel.GetMerchantBalanceRequest{
				MerchantID:  merchantID,
				BalanceName: constant.AccountNamePayment,
				Date:        now,
			},
			setupMock: func() {
				accountRepo.On(
					"FindMerchantAccountByName", mock.Anything, merchantUUID, constant.AccountNamePayment,
				).Once().Return(paymentAccount, nil)
				accountTransactionRepo.On(
					"GetAggregateTransactions", mock.Anything, mock.Anything,
				).Once().Return(validAggregateResponse, nil)
				accountTransactionRepo.On(
					"CalculatePendingBalance", mock.Anything, mock.Anything,
				).Once().Return(50_000.00, nil)
			},
			wantResult: &orchestratorModel.GetMerchantBalanceResponse{
				AvailableBalance: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    "140000.00",
				},
				PendingBalance: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    "50000.00",
				},
				TotalBalance: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    "190000.00",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.setupMock()

			result, err := service.GetMerchantBalance(t.Context(), test.request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)

			logger.AssertExpectations(t)
			accountRepo.AssertExpectations(t)
			accountTransactionRepo.AssertExpectations(t)
		})
	}
}
