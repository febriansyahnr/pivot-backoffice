package disbursementService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetBulkDisbursementForOpenApiByID(t *testing.T) {
	validFilter := &disbursementModel.GetDisbursementFilterRequest{
		BulkID: uuid.NewString(),
	}
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	data := make([]*disbursementModel.DisbursementWithTransactionResponse, 0)
	remark := "Remark"
	data = append(data, &disbursementModel.DisbursementWithTransactionResponse{
		DisbursementWithTransaction: disbursementModel.DisbursementWithTransaction{
			Disbursement: disbursementModel.Disbursement{
				UUID:                   uuid.NewString(),
				BeneficiaryAccountNo:   "999966660007",
				BeneficiaryAccountName: "Test",
				BeneficiaryBankCode:    "013",
				Remark:                 &remark,
			},
		},
	})
	expectedResponse := commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	dataDelayed := make([]*disbursementModel.DisbursementWithTransactionResponse, 0)
	dataDelayed = append(dataDelayed, &disbursementModel.DisbursementWithTransactionResponse{
		DisbursementWithTransaction: disbursementModel.DisbursementWithTransaction{
			Disbursement: disbursementModel.Disbursement{
				UUID:                   uuid.NewString(),
				BeneficiaryAccountNo:   "999966660007",
				BeneficiaryAccountName: "Test",
				BeneficiaryBankCode:    "013",
				Remark:                 &remark,
			},
			TransactionReasonType: stringPtr(constant.ReasonTypePayoutDelayed),
		},
	})
	expectedDelayedResponse := commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	dataCancelled := make([]*disbursementModel.DisbursementWithTransactionResponse, 0)
	dataCancelled = append(dataCancelled, &disbursementModel.DisbursementWithTransactionResponse{
		DisbursementWithTransaction: disbursementModel.DisbursementWithTransaction{
			Disbursement: disbursementModel.Disbursement{
				UUID:                   uuid.NewString(),
				BeneficiaryAccountNo:   "999966660007",
				BeneficiaryAccountName: "Test",
				BeneficiaryBankCode:    "013",
				Remark:                 &remark,
				ReasonType:             stringPtr(constant.DisbursementReasonTypeCancelled),
				ReasonDescription:      stringPtr(constant.DisbursementReasonTypeCancelled),
			},
		},
	})
	expectedCancelledResponse := commonModel.PaginationResponse{
		Data: dataCancelled,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.IDisbursementRepository)
		filter    *disbursementModel.GetDisbursementFilterRequest
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.
					On(
						"FindBulkDisbursementByID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.BulkDisbursement{UUID: validFilter.BulkID}, nil)

				mockRepo.
					On(
						"GetList",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"),
						mock.AnythingOfType("int64"),
						mock.AnythingOfType("int64"),
					).Return(&expectedResponse, nil)

				mockRepo.
					On(
						"SummaryPendingByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				mockRepo.
					On(
						"SummarySuccessByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				mockRepo.
					On(
						"SummaryFailedByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				mockRepo.
					On(
						"SummaryCancelledByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})
			},
			filter: validFilter,
		},
		{
			name:    "SUCCESS: Contains delayed status",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.
					On(
						"FindBulkDisbursementByID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.BulkDisbursement{UUID: validFilter.BulkID}, nil)

				mockRepo.
					On(
						"GetList",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"),
						mock.AnythingOfType("int64"),
						mock.AnythingOfType("int64"),
					).Return(&expectedDelayedResponse, nil)

				mockRepo.
					On(
						"SummaryPendingByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				mockRepo.
					On(
						"SummarySuccessByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				mockRepo.
					On(
						"SummaryFailedByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				mockRepo.
					On(
						"SummaryCancelledByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

			},
			filter: validFilter,
		},
		{
			name:    "SUCCESS: Contains cancelled status",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.
					On(
						"FindBulkDisbursementByID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.BulkDisbursement{UUID: validFilter.BulkID}, nil)

				mockRepo.
					On(
						"GetList",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*disbursementModel.GetDisbursementFilterRequest"),
						mock.AnythingOfType("int64"),
						mock.AnythingOfType("int64"),
					).Return(&expectedCancelledResponse, nil)

				mockRepo.
					On(
						"SummaryPendingByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				mockRepo.
					On(
						"SummarySuccessByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				mockRepo.
					On(
						"SummaryFailedByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				mockRepo.
					On(
						"SummaryCancelledByBulkID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(disbursementDashboardModel.SummaryTransactionDTO{})

			},
			filter: validFilter,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockRepo := repositoryMocks.NewIDisbursementRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			ctx := context.Background()
			tc.mockSetup(mockRepo)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			svc := New(&conf, loggerMock, merchantRepo, mockRepo, nil, nil,
				WithStatusHistoriesRepository(statusHistoriesRepo),
			)
			response, err := svc.GetBulkDisbursementForOpenApiByID(ctx, tc.filter, 1, 20)
			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, response)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetBulkDisbursementForOpenApiByReferenceID(t *testing.T) {
	bulkID := uuid.NewString()
	merchantID := uuid.NewString()
	remark := "Remark"
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.IDisbursementRepository)
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.
					On(
						"FindBulkDisbursementByID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.BulkDisbursement{UUID: bulkID, MerchantID: merchantID}, nil)

				mockRepo.
					On(
						"FindByMerchantAndReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						BulkID:                 &bulkID,
						UUID:                   uuid.NewString(),
						BeneficiaryAccountNo:   "1234",
						BeneficiaryAccountName: "Test",
						BeneficiaryBankCode:    "013",
						Remark:                 &remark,
					},
				}, nil)
			},
		},
		{
			name:    "SUCCESS: Contains delayed payout",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.
					On(
						"FindBulkDisbursementByID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.BulkDisbursement{UUID: bulkID, MerchantID: merchantID}, nil)

				mockRepo.
					On(
						"FindByMerchantAndReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						BulkID:                 &bulkID,
						UUID:                   uuid.NewString(),
						BeneficiaryAccountNo:   "1234",
						BeneficiaryAccountName: "Test",
						BeneficiaryBankCode:    "013",
						Remark:                 &remark,
					},
					TransactionReasonType: stringPtr(constant.ReasonTypePayoutDelayed),
				}, nil)
			},
		},
		{
			name:    "SUCCESS: Contains cancelled payout",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.
					On(
						"FindBulkDisbursementByID",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.BulkDisbursement{UUID: bulkID, MerchantID: merchantID}, nil)

				mockRepo.
					On(
						"FindByMerchantAndReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						BulkID:                 &bulkID,
						UUID:                   uuid.NewString(),
						BeneficiaryAccountNo:   "1234",
						BeneficiaryAccountName: "Test",
						BeneficiaryBankCode:    "013",
						Remark:                 &remark,
						ReasonType:             stringPtr(constant.DisbursementReasonTypeCancelled),
						ReasonDescription:      stringPtr(constant.DisbursementReasonTypeCancelled),
					},
				}, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockRepo := repositoryMocks.NewIDisbursementRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			ctx := context.Background()
			tc.mockSetup(mockRepo)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			svc := New(&conf, loggerMock, merchantRepo, mockRepo, nil, nil,
				WithStatusHistoriesRepository(statusHistoriesRepo),
			)
			response, err := svc.GetBulkDisbursementForOpenApiByReferenceID(ctx, bulkID, "reference", merchantID)
			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, response)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBuildReason(t *testing.T) {
	reason := "test reason"
	transactionReason := "transaction reason"

	testCases := []struct {
		name         string
		disbursement disbursementModel.DisbursementWithTransaction
		want         string
	}{
		{
			name: "insufficient balance reason",
			disbursement: disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{
					ReasonType:        stringPtr(constant.DisbursementReasonTypeInsufficientBalance),
					ReasonDescription: &reason,
				},
			},
			want: reason,
		},
		{
			name: "invalid account reason",
			disbursement: disbursementModel.DisbursementWithTransaction{
				Disbursement:                 disbursementModel.Disbursement{},
				TransactionReasonType:        stringPtr(constant.ReasonTypeBeneficiaryAccountReason),
				TransactionReasonDescription: stringPtr("invalid account"),
			},
			want: "Invalid Account",
		},
		{
			name: "dormant account reason",
			disbursement: disbursementModel.DisbursementWithTransaction{
				Disbursement:                 disbursementModel.Disbursement{},
				TransactionReasonType:        stringPtr(constant.ReasonTypeBeneficiaryAccountReason),
				TransactionReasonDescription: stringPtr("dormant account"),
			},
			want: "Dormant Account",
		},
		{
			name: "blocked by harsya reason",
			disbursement: disbursementModel.DisbursementWithTransaction{
				Disbursement:                 disbursementModel.Disbursement{},
				TransactionReasonType:        stringPtr(constant.ReasonTypeBlockedByHarsya),
				TransactionReasonDescription: &transactionReason,
			},
			want: "Feature not allowed at this time",
		},
		{
			name: "bank network error record not found",
			disbursement: disbursementModel.DisbursementWithTransaction{
				Disbursement:                 disbursementModel.Disbursement{},
				TransactionReasonType:        stringPtr(constant.ReasonTypeBankNetworkError),
				TransactionReasonDescription: stringPtr("record not found"),
			},
			want: "Failed to process by Bank Network",
		},
		{
			name: "beneficiary account reason with description",
			disbursement: disbursementModel.DisbursementWithTransaction{
				Disbursement:                 disbursementModel.Disbursement{},
				TransactionReasonType:        stringPtr(constant.ReasonTypeBeneficiaryAccountReason),
				TransactionReasonDescription: &transactionReason,
			},
			want: transactionReason,
		},
		{
			name: "transaction failed",
			disbursement: disbursementModel.DisbursementWithTransaction{
				Disbursement:                 disbursementModel.Disbursement{},
				TransactionReasonType:        stringPtr(constant.ReasonTypeOtherReason),
				TransactionReasonDescription: stringPtr("Transaction Failed"),
			},
			want: "Failed to process by Bank Network",
		},
		{
			name: "no reason type",
			disbursement: disbursementModel.DisbursementWithTransaction{
				Disbursement:      disbursementModel.Disbursement{},
				TransactionStatus: stringPtr(constant.StatusFailed),
			},
			want: "Failed to process by Bank Network",
		},
		{
			name: "no reason type and empty transaction status",
			disbursement: disbursementModel.DisbursementWithTransaction{
				Disbursement: disbursementModel.Disbursement{},
			},
			want: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildReason(tc.disbursement)
			assert.Equal(t, tc.want, got)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestGetDisbursementByReferenceID(t *testing.T) {
	merchantID := uuid.NewString()
	referenceID := "test-reference-id"
	remark := "Test Remark"
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	testCases := []struct {
		name      string
		wantErr   bool
		mockSetup func(mockRepo *repositoryMocks.IDisbursementRepository)
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.
					On(
						"FindByMerchantAndReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(&disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:                   uuid.NewString(),
						ReferenceID:            referenceID,
						BeneficiaryAccountNo:   "1234567890",
						BeneficiaryAccountName: "Test Account",
						BeneficiaryBankCode:    "013",
						Remark:                 &remark,
						Status:                 constant.StatusSuccess,
					},
				}, nil)
			},
		},
		{
			name:    "ERROR_DISBURSEMENT_NOT_FOUND",
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.
					On(
						"FindByMerchantAndReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(nil, nil)
			},
		},
		{
			name:    "ERROR_REPOSITORY",
			wantErr: true,
			mockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.
					On(
						"FindByMerchantAndReference",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("string"),
						mock.AnythingOfType("string"),
					).Return(nil, errors.New("repository error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockRepo := repositoryMocks.NewIDisbursementRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			ctx := context.Background()
			tc.mockSetup(mockRepo)
			statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			svc := New(&conf, loggerMock, merchantRepo, mockRepo, nil, nil,
				WithStatusHistoriesRepository(statusHistoriesRepo),
			)
			response, err := svc.GetDisbursementByReferenceID(ctx, referenceID, merchantID)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, referenceID, response.ReferenceID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
