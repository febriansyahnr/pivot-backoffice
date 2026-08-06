package disbursementDashboardService

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetApprovalDashboard(t *testing.T) {
	loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name             string
		setupMocks       func(disbursementRepo *mocks.IDisbursementRepository, accountRepo *mocks.IAccountRepository)
		expectedResponse *disbursementDashboardModel.DisbursementApprovalDashboardResponse
		wantErr          bool
		expectedError    error
	}{
		{
			name: "SUCCESS: Get disbursement approval dashboard",
			setupMocks: func(disbursementRepo *mocks.IDisbursementRepository, accountRepo *mocks.IAccountRepository) {
				accountRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(&account_model.Account{Name: constant.TypeDisbursement, Currency: "IDR"}, nil)
				mockDTO := disbursementDashboardModel.SummaryTransactionDTO{
					Count: 3,
					Sum:   61041111.00,
				}

				uuidStr := constant.EmptyUUID

				methods := []struct {
					methodName string
					arg1       context.Context
					arg2       string
					dto        disbursementDashboardModel.SummaryTransactionDTO
				}{
					{"CountWaitingSingleDisbursement", context.Background(), uuidStr, mockDTO},
					{"CountWaitingBulkDisbursement", context.Background(), uuidStr, mockDTO},
					{"CountPendingSingleDisbursement", context.Background(), uuidStr, mockDTO},
					{"CountPendingBulkDisbursement", context.Background(), uuidStr, mockDTO},
				}

				for _, m := range methods {
					disbursementRepo.On(
						m.methodName,
						constant.ValueCtxMockType(),
						mock.AnythingOfType("disbursementDashboardModel.GetDisbursementDashboardFilter"),
					).Return(m.dto, nil).Once()
				}
			},
			expectedResponse: &disbursementDashboardModel.DisbursementApprovalDashboardResponse{
				WaitingSingleDisbursement: disbursementDashboardModel.SummaryTransaction{
					Count: 3,
					Sum: commonModel.Amount{
						Currency: "IDR",
						Value:    "61041111.00",
					},
				},
				WaitingBulkDisbursement: disbursementDashboardModel.SummaryTransaction{
					Count: 3,
					Sum: commonModel.Amount{
						Currency: "IDR",
						Value:    "61041111.00",
					},
				},
				PendingSingleDisbursement: disbursementDashboardModel.SummaryTransaction{
					Count: 3,
					Sum: commonModel.Amount{
						Currency: "IDR",
						Value:    "61041111.00",
					},
				},
				PendingBulkDisbursement: disbursementDashboardModel.SummaryTransaction{
					Count: 3,
					Sum: commonModel.Amount{
						Currency: "IDR",
						Value:    "61041111.00",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "ERROR: Get error on FindMerchantAccountByName repo",
			setupMocks: func(disbursementRepo *mocks.IDisbursementRepository, accountRepo *mocks.IAccountRepository) {
				accountRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			expectedResponse: nil,
			wantErr:          true,
			expectedError:    constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR: FindMerchantAccountByName not found",
			setupMocks: func(disbursementRepo *mocks.IDisbursementRepository, accountRepo *mocks.IAccountRepository) {
				accountRepo.On(
					"FindMerchantAccountByName",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.UuidMockType(),
					constant.StringMockType(),
				).Return(nil, nil)
			},
			expectedResponse: &disbursementDashboardModel.DisbursementApprovalDashboardResponse{
				WaitingSingleDisbursement: disbursementDashboardModel.SummaryTransaction{
					Count: 0,
					Sum: commonModel.Amount{
						Currency: "IDR",
						Value:    "0.00",
					},
				},
				WaitingBulkDisbursement: disbursementDashboardModel.SummaryTransaction{
					Count: 0,
					Sum: commonModel.Amount{
						Currency: "IDR",
						Value:    "0.00",
					},
				},
				PendingSingleDisbursement: disbursementDashboardModel.SummaryTransaction{
					Count: 0,
					Sum: commonModel.Amount{
						Currency: "IDR",
						Value:    "0.00",
					},
				},
				PendingBulkDisbursement: disbursementDashboardModel.SummaryTransaction{
					Count: 0,
					Sum: commonModel.Amount{
						Currency: "IDR",
						Value:    "0.00",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			disbursementRepoMock := mocks.NewIDisbursementRepository(t)
			accountRepoMock := mocks.NewIAccountRepository(t)

			tc.setupMocks(disbursementRepoMock, accountRepoMock)

			trxSvc := New(loggerMock, disbursementRepoMock, nil, accountRepoMock, nil)

			response, err := trxSvc.GetApprovalDashboard(context.Background(), uuid.New())

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, response)
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResponse, response)
			}
			disbursementRepoMock.AssertExpectations(t)
			accountRepoMock.AssertExpectations(t)
		})
	}
}
