package disbursementDashboardService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDisbursementDashboardService_Get(t *testing.T) {
	var (
		validUUID = uuid.New()
	)
	testCases := []struct {
		name       string
		payload    disbursementDashboardModel.GetDisbursementDashboardFilter
		mocksSetup func(
			disbursementRepo *mocks.IDisbursementRepository,
			accTrxRepo *mocks.IAccountTransactionRepository,
			balanceRepo *mocks.IAccountRepository,
			orchSvc *serviceMocks.IOrchestratorService)
		wantErr bool
	}{
		{
			name: "SUCCESS: Get disbursement dashboard",
			payload: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: validUUID.String(),
			},
			mocksSetup: func(
				disbursementRepo *mocks.IDisbursementRepository,
				accTrxRepo *mocks.IAccountTransactionRepository,
				balanceRepo *mocks.IAccountRepository,
				orchSvc *serviceMocks.IOrchestratorService,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(&account_model.Account{Name: constant.TypeDisbursement}, nil)


				orchSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(0), nil)

				orchSvc.On(
					"GetPendingBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(0), nil)

				for _, method := range []string{"SummaryWaitingToday", "SummarySingleWaitingToday", "SummaryBulkWaitingToday", "GetSummaryAll", "GetSummaryInProgress", "GetSummaryFailed", "GetSummarySuccess", "GetSummaryRejected", "SummaryWaitingForTopUpToday", "SummarySingleWaitingForTopUpToday", "SummaryBulkWaitingForTopUpToday", "GetSummaryRejected", "GetSummaryApproved"} {
					disbursementRepo.On(
						method,
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("disbursementDashboardModel.GetDisbursementDashboardFilter"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{}, nil)
				}
			},
			wantErr: false,
		},
		{
			name: "ERROR: FindMerchantAccountByName",
			payload: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: validUUID.String(),
			},
			mocksSetup: func(
				disbursementRepo *mocks.IDisbursementRepository,
				accTrxRepo *mocks.IAccountTransactionRepository,
				balanceRepo *mocks.IAccountRepository,
				orchSvc *serviceMocks.IOrchestratorService,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("some error"))
			},
			wantErr: false,
		},
		{
			name: "ERROR: Invalid Merchant ID",
			payload: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: "invalid-uuid",
			},
			mocksSetup: func(
				disbursementRepo *mocks.IDisbursementRepository,
				accTrxRepo *mocks.IAccountTransactionRepository,
				balanceRepo *mocks.IAccountRepository,
				orchSvc *serviceMocks.IOrchestratorService,

			) {
			},
			wantErr: true,
		},
		{
			name: "ERROR: Merchant balance not found",
			payload: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: validUUID.String(),
			},
			mocksSetup: func(
				disbursementRepo *mocks.IDisbursementRepository,
				accTrxRepo *mocks.IAccountTransactionRepository,
				balanceRepo *mocks.IAccountRepository,
				orchSvc *serviceMocks.IOrchestratorService,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(nil, nil)

			},
			wantErr: false,
		},
		{
			name: "ERROR: GetAggregateTransactions",
			payload: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: validUUID.String(),
			},
			mocksSetup: func(
				disbursementRepo *mocks.IDisbursementRepository,
				accTrxRepo *mocks.IAccountTransactionRepository,
				balanceRepo *mocks.IAccountRepository,
				orchSvc *serviceMocks.IOrchestratorService,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(&account_model.Account{}, nil)


				orchSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(0), nil)

				orchSvc.On(
					"GetPendingBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(0), nil)

				for _, method := range []string{"SummaryWaitingToday", "SummarySingleWaitingToday", "SummaryBulkWaitingToday", "GetSummaryAll", "GetSummaryInProgress", "GetSummaryFailed", "GetSummarySuccess", "GetSummaryRejected", "SummaryWaitingForTopUpToday", "SummarySingleWaitingForTopUpToday", "SummaryBulkWaitingForTopUpToday", "GetSummaryRejected", "GetSummaryApproved"} {
					disbursementRepo.On(
						method,
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("disbursementDashboardModel.GetDisbursementDashboardFilter"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{}, nil)
				}
			},
			wantErr: false,
		},
		{
			name: "ERROR: GetAvailableMerchantBalance",
			payload: disbursementDashboardModel.GetDisbursementDashboardFilter{
				MerchantID: validUUID.String(),
			},
			mocksSetup: func(
				disbursementRepo *mocks.IDisbursementRepository,
				accTrxRepo *mocks.IAccountTransactionRepository,
				balanceRepo *mocks.IAccountRepository,
				orchSvc *serviceMocks.IOrchestratorService,

			) {
				balanceRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					mock.AnythingOfType("string"),
				).Return(&account_model.Account{}, nil)


				orchSvc.On(
					"GetAvailableMerchantBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(0), constant.ErrSomeErrorForUnitTest)

				orchSvc.On(
					"GetPendingBalance",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(float64(0), nil)

				for _, method := range []string{"SummaryWaitingToday", "SummarySingleWaitingToday", "SummaryBulkWaitingToday", "GetSummaryAll", "GetSummaryInProgress", "GetSummaryFailed", "GetSummarySuccess", "GetSummaryRejected", "SummaryWaitingForTopUpToday", "SummarySingleWaitingForTopUpToday", "SummaryBulkWaitingForTopUpToday", "GetSummaryRejected", "GetSummaryApproved"} {
					disbursementRepo.On(
						method,
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("disbursementDashboardModel.GetDisbursementDashboardFilter"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{}, nil)
				}
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			disbursementRepoMock := mocks.NewIDisbursementRepository(t)
			accTrxRepoMock := mocks.NewIAccountTransactionRepository(t)
			balanceRepoMock := mocks.NewIAccountRepository(t)
			orchSvc := serviceMocks.NewIOrchestratorService(t)

			tc.mocksSetup(disbursementRepoMock, accTrxRepoMock, balanceRepoMock, orchSvc)

			trxSvc := New(loggerMock, disbursementRepoMock, accTrxRepoMock, balanceRepoMock, orchSvc)

			_, err := trxSvc.Get(context.Background(), tc.payload)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			balanceRepoMock.AssertExpectations(t)
		})
	}
}
