package disbursementService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	mocks_repository "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/test"
)

func TestDisbursementService_GetPayoutStatusAndRouting(t *testing.T) {
	type Mockers struct {
		disbursementRepo       *mocks_repository.IDisbursementRepository
		snapCoreRepo           *mocks_repository.ISnapCoreRepository
		accountTransactionRepo *mocks_repository.IAccountTransactionRepository
	}

	referenceID := "test-reference-id"
	disbursementUUID := "disbursement-uuid-123"
	processorReferenceID := "processor-ref-123"
	beneficiaryBankName := "Bank ABC"
	amount := decimal.NewFromFloat(100000.00)

	mockDisbursement := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:                   disbursementUUID,
			ReferenceID:            referenceID,
			Status:                 "APPROVED",
			BeneficiaryBankName:    &beneficiaryBankName,
			BeneficiaryAccountNo:   "1234567890",
			BeneficiaryAccountName: "John Doe",
			Amount:                 amount,
			ProcessorReferenceID:   &processorReferenceID,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		},
	}

	mockAccountTransaction := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:   uuid.New(),
		Status: "SUCCESS",
	}

	mockStatusData := &snapCoreModel.BankTransferCheckStatusResponseData{
		UUID:         disbursementUUID,
		BankAcquirer: "BANK_ABC",
		TransferType: "INTERBANK",
		Status:       "PENDING",
		TransferLogs: []snapCoreModel.TransferLog{
			{
				UUID:   "log-1",
				Bank:   "Bank ABC",
				Action: "TRANSFER_REQUEST",
				Status: "PENDING",
				ResponsePayload: snapCoreModel.ResponsePayload{
					ResponseCode:    "2001",
					ResponseMessage: "Transaction pending",
				},
				CreatedAt: "2024-01-01T10:00:00Z",
			},
		},
		Outbounds: []snapCoreModel.Outbound{
			{
				UUID:      "outbound-1",
				Title:     "Bank Transfer Request",
				Acquirer:  "Bank ABC",
				CreatedAt: "2024-01-01T10:00:00Z",
			},
		},
	}

	testCases := []struct {
		desc      string
		request   *disbursementModel.CRMSinglePayoutStatusRequest
		wantError bool
		setupMock func(mockers Mockers)
	}{
		{
			desc: "success - found by disbursement UUID with processor reference ID",
			request: &disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: disbursementUUID,
			},
			wantError: false,
			setupMock: func(mockers Mockers) {
				mockers.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					disbursementUUID,
				).Return(mockDisbursement, nil)

				mockers.accountTransactionRepo.On(
					"FindByReference",
					mock.Anything,
					disbursementUUID,
					constant.TypeDisbursement,
				).Return(mockAccountTransaction, nil)

				mockers.snapCoreRepo.On(
					"CheckStatusByExternalId",
					mock.Anything,
					mockAccountTransaction.UUID.String(),
					false,
				).Return(mockStatusData, nil)
			},
		},
		{
			desc: "success - found by processor reference ID",
			request: &disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: processorReferenceID,
			},
			wantError: false,
			setupMock: func(mockers Mockers) {
				mockers.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					processorReferenceID,
				).Return(nil, errors.New("not found"))

				mockers.disbursementRepo.On(
					"FindByReference",
					mock.Anything,
					processorReferenceID,
				).Return(mockDisbursement, nil)

				mockers.accountTransactionRepo.On(
					"FindByReference",
					mock.Anything,
					disbursementUUID,
					constant.TypeDisbursement,
				).Return(mockAccountTransaction, nil)

				mockers.snapCoreRepo.On(
					"CheckStatusByExternalId",
					mock.Anything,
					mockAccountTransaction.UUID.String(),
					false,
				).Return(mockStatusData, nil)
			},
		},
		{
			desc: "success - no processor reference ID, basic routing history",
			request: &disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: disbursementUUID,
			},
			wantError: false,
			setupMock: func(mockers Mockers) {
				mockDisbursementWithoutProcessor := *mockDisbursement
				mockDisbursementWithoutProcessor.ProcessorReferenceID = nil

				mockers.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					disbursementUUID,
				).Return(&mockDisbursementWithoutProcessor, nil)
			},
		},
		{
			desc: "success - snap-core error, fallback to basic routing history",
			request: &disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: disbursementUUID,
			},
			wantError: false,
			setupMock: func(mockers Mockers) {
				mockers.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					disbursementUUID,
				).Return(mockDisbursement, nil)

				mockers.accountTransactionRepo.On(
					"FindByReference",
					mock.Anything,
					disbursementUUID,
					constant.TypeDisbursement,
				).Return(mockAccountTransaction, nil)

				mockers.snapCoreRepo.On(
					"CheckStatusByExternalId",
					mock.Anything,
					mockAccountTransaction.UUID.String(),
					false,
				).Return(nil, errors.New("snap-core error"))
			},
		},
		{
			desc: "error - transaction not found",
			request: &disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: "non-existent-id",
			},
			wantError: true,
			setupMock: func(mockers Mockers) {
				mockers.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					"non-existent-id",
				).Return(nil, errors.New("not found"))

				mockers.disbursementRepo.On(
					"FindByReference",
					mock.Anything,
					"non-existent-id",
				).Return(nil, errors.New("not found"))
			},
		},
		{
			desc: "error - database error on FindByID",
			request: &disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: disbursementUUID,
			},
			wantError: true,
			setupMock: func(mockers Mockers) {
				mockers.disbursementRepo.On(
					"FindByID",
					mock.Anything,
					disbursementUUID,
				).Return(nil, errors.New("database connection error"))

				mockers.disbursementRepo.On(
					"FindByReference",
					mock.Anything,
					disbursementUUID,
				).Return(nil, errors.New("database connection error"))

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockers := Mockers{
				disbursementRepo:       new(mocks_repository.IDisbursementRepository),
				snapCoreRepo:           new(mocks_repository.ISnapCoreRepository),
				accountTransactionRepo: new(mocks_repository.IAccountTransactionRepository),
			}
			tc.setupMock(mockers)

			_, pdkLogger, err := test.SetupLogger()
			assert.NoError(t, err)

			service := &DisbursementService{
				disbursementRepo:       mockers.disbursementRepo,
				snapCoreRepo:           mockers.snapCoreRepo,
				accountTransactionRepo: mockers.accountTransactionRepo,
				logger:                 pdkLogger,
			}

			result, err := service.GetPayoutStatusAndRouting(context.Background(), tc.request)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "00", result.Code)
				assert.NotNil(t, result.Data)
				assert.Equal(t, disbursementUUID, result.Data.DisbursementUUID)
				assert.Equal(t, referenceID, result.Data.ReferenceID)
				assert.NotNil(t, result.Data.TransferLogs)
			}

			mockers.disbursementRepo.AssertExpectations(t)
			mockers.snapCoreRepo.AssertExpectations(t)
			mockers.accountTransactionRepo.AssertExpectations(t)
		})
	}
}

func TestDisbursementService_buildTransferLogsFromStatusData(t *testing.T) {
	service := &DisbursementService{}

	statusData := &snapCoreModel.BankTransferCheckStatusResponseData{
		TransferLogs: []snapCoreModel.TransferLog{
			{
				UUID:   "log-1",
				Bank:   "Bank ABC",
				Action: "TRANSFER_REQUEST",
				Status: "FAILED",
				ResponsePayload: snapCoreModel.ResponsePayload{
					ResponseCode:    "4001",
					ResponseMessage: "Insufficient funds",
				},
				CreatedAt: "2024-01-01T10:00:00Z",
			},
			{
				UUID:   "log-2",
				Bank:   "Bank DEF",
				Action: "TRANSFER_RETRY",
				Status: "SUCCESS",
				ResponsePayload: snapCoreModel.ResponsePayload{
					ResponseCode:    "2000",
					ResponseMessage: "Transaction successful",
				},
				CreatedAt: "2024-01-01T10:05:00Z",
			},
		},
	}

	result := service.buildTransferLogsFromStatusData(statusData)

	assert.Len(t, result, 2) // 2 transfer logs only

	// First log - FAILED status should show responseMessage
	assert.Equal(t, 1, result[0].Order)
	assert.Equal(t, "Bank ABC", result[0].BankName)
	assert.Equal(t, "FAILED", result[0].Status)
	assert.Equal(t, "Insufficient funds", result[0].ResponseMsg)

	// Second log - SUCCESS status should have empty responseMessage
	assert.Equal(t, 2, result[1].Order)
	assert.Equal(t, "Bank DEF", result[1].BankName)
	assert.Equal(t, "SUCCESS", result[1].Status)
	assert.Equal(t, "", result[1].ResponseMsg)
}

func TestDisbursementService_GetBatchPayoutStatusAndRouting(t *testing.T) {
	type Mockers struct {
		disbursementRepo       *mocks_repository.IDisbursementRepository
		snapCoreRepo           *mocks_repository.ISnapCoreRepository
		accountTransactionRepo *mocks_repository.IAccountTransactionRepository
	}

	referenceID1 := "test-reference-id-1"
	referenceID2 := "test-reference-id-2"
	referenceID3 := "non-existent-id"
	disbursementUUID1 := "disbursement-uuid-1"
	disbursementUUID2 := "disbursement-uuid-2"
	processorReferenceID1 := "processor-ref-1"
	beneficiaryBankName := "Bank ABC"
	amount := decimal.NewFromFloat(100000.00)

	mockDisbursement1 := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:                   disbursementUUID1,
			ReferenceID:            referenceID1,
			Status:                 "SUCCESS",
			BeneficiaryBankName:    &beneficiaryBankName,
			BeneficiaryAccountNo:   "1234567890",
			BeneficiaryAccountName: "John Doe",
			Amount:                 amount,
			ProcessorReferenceID:   &processorReferenceID1,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		},
	}

	mockDisbursement2 := &disbursementModel.DisbursementWithTransaction{
		Disbursement: disbursementModel.Disbursement{
			UUID:                   disbursementUUID2,
			ReferenceID:            referenceID2,
			Status:                 "PENDING",
			BeneficiaryBankName:    &beneficiaryBankName,
			BeneficiaryAccountNo:   "9876543210",
			BeneficiaryAccountName: "Jane Smith",
			Amount:                 amount,
			ProcessorReferenceID:   nil, // No processor reference ID
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		},
	}

	mockAccountTransaction := &orchestratorModel.AccountTransactionWithUseCase{
		UUID:   uuid.New(),
		Status: "SUCCESS",
	}

	testCases := []struct {
		desc             string
		request          *disbursementModel.CRMBatchPayoutStatusRequest
		wantError        bool
		expectedResults  int
		expectedSuccess  int
		expectedFailed   int
		expectedNotFound int
		setupMock        func(mockers Mockers)
	}{
		{
			desc: "success - multiple reference IDs with mixed results",
			request: &disbursementModel.CRMBatchPayoutStatusRequest{
				ReferenceIDs: []string{referenceID1, referenceID2, referenceID3},
			},
			wantError:        false,
			expectedResults:  3,
			expectedSuccess:  2,
			expectedFailed:   1,
			expectedNotFound: 0,
			setupMock: func(mockers Mockers) {
				// Mock for first reference ID - success with snap-core data
				mockers.disbursementRepo.On("FindByID", mock.Anything, referenceID1).Return(mockDisbursement1, nil)
				mockers.accountTransactionRepo.On("FindByReference", mock.Anything, disbursementUUID1, constant.TypeDisbursement).Return(mockAccountTransaction, nil)
				mockers.snapCoreRepo.On("CheckStatusByExternalId", mock.Anything, mockAccountTransaction.UUID.String(), false).Return(&snapCoreModel.BankTransferCheckStatusResponseData{
					UUID:   disbursementUUID1,
					Status: "SUCCESS",
					TransferLogs: []snapCoreModel.TransferLog{
						{
							UUID:   "log-1",
							Bank:   "Bank ABC",
							Status: "SUCCESS",
							ResponsePayload: snapCoreModel.ResponsePayload{
								ResponseMessage: "Transaction successful",
							},
							CreatedAt: "2024-01-01T10:00:00Z",
						},
					},
				}, nil)

				// Mock for second reference ID - success without snap-core data
				mockers.disbursementRepo.On("FindByID", mock.Anything, referenceID2).Return(mockDisbursement2, nil)

				// Mock for third reference ID - not found
				mockers.disbursementRepo.On("FindByID", mock.Anything, referenceID3).Return(nil, errors.New("not found"))
				mockers.disbursementRepo.On("FindByReference", mock.Anything, referenceID3).Return(nil, errors.New("not found"))
			},
		},
		{
			desc: "success - single reference ID in array",
			request: &disbursementModel.CRMBatchPayoutStatusRequest{
				ReferenceIDs: []string{referenceID1},
			},
			wantError:        false,
			expectedResults:  1,
			expectedSuccess:  1,
			expectedFailed:   0,
			expectedNotFound: 0,
			setupMock: func(mockers Mockers) {
				mockers.disbursementRepo.On("FindByID", mock.Anything, referenceID1).Return(mockDisbursement1, nil)
				mockers.accountTransactionRepo.On("FindByReference", mock.Anything, disbursementUUID1, constant.TypeDisbursement).Return(mockAccountTransaction, nil)
				mockers.snapCoreRepo.On("CheckStatusByExternalId", mock.Anything, mockAccountTransaction.UUID.String(), false).Return(&snapCoreModel.BankTransferCheckStatusResponseData{
					UUID:   disbursementUUID1,
					Status: "SUCCESS",
				}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockers := Mockers{
				disbursementRepo:       new(mocks_repository.IDisbursementRepository),
				snapCoreRepo:           new(mocks_repository.ISnapCoreRepository),
				accountTransactionRepo: new(mocks_repository.IAccountTransactionRepository),
			}
			tc.setupMock(mockers)

			_, pdkLogger, err := test.SetupLogger()
			assert.NoError(t, err)

			service := &DisbursementService{
				disbursementRepo:       mockers.disbursementRepo,
				snapCoreRepo:           mockers.snapCoreRepo,
				accountTransactionRepo: mockers.accountTransactionRepo,
				logger:                 pdkLogger,
			}

			result, err := service.GetBatchPayoutStatusAndRouting(context.Background(), tc.request)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Data, tc.expectedResults)
				assert.Equal(t, "00", result.Code)

				// Count success and failed results
				successCount := 0
				failedCount := 0
				for _, res := range result.Data {
					if res.Success {
						successCount++
					} else {
						failedCount++
					}
				}
				assert.Equal(t, tc.expectedSuccess, successCount)
				assert.Equal(t, tc.expectedFailed, failedCount)

				// Verify individual results
				for _, res := range result.Data {
					assert.NotEmpty(t, res.ReferenceID)
					if res.Success {
						assert.NotNil(t, res.Data)
						assert.Nil(t, res.Error)
					} else {
						assert.Nil(t, res.Data)
						assert.NotNil(t, res.Error)
					}
				}
			}

			mockers.disbursementRepo.AssertExpectations(t)
			mockers.snapCoreRepo.AssertExpectations(t)
			mockers.accountTransactionRepo.AssertExpectations(t)
		})
	}
}
