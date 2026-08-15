package disbursementService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	gcsMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReject(t *testing.T) {
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}
	disbursementID := uuid.NewString()
	validInput := &disbursementModel.RejectRequest{
		RejectAction: []disbursementModel.RejectActionObject{
			{
				DisbursementID: disbursementID,
			},
		},
		MerchantID: uuid.NewString(),
		RejectedBy: uuid.NewString(),
	}

	validRejectedDisbursement := []*disbursementModel.Disbursement{
		{
			UUID: disbursementID,
		},
	}

	testCases := []struct {
		name       string
		mocksSetup func(disbursementRepo *repositoryMocks.IDisbursementRepository,
			gcsSvc *gcsMocks.GCSService,
		)
		input   *disbursementModel.RejectRequest
		wantErr bool
	}{
		{
			name: "SUCCESS: Reject action",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
			) {
				disbursementRepo.On(
					"CountByIDsAndMerchantID",
					mock.Anything,
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
				).Return(1)

				disbursementRepo.On(
					"Reject",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				disbursementRepo.On("CommitTransaction", mock.Anything).
					Return(nil)

				disbursementRepo.On(
					"GetByIDs",
					mock.Anything,
					constant.ArrayStringMockType(),
				).Return(validRejectedDisbursement, nil)
			},
			input:   validInput,
			wantErr: false,
		},
		{
			name: "SUCCESS: Without any rejection due to empty disbursement ID",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
			) {
			},
			input: &disbursementModel.RejectRequest{
				RejectAction: []disbursementModel.RejectActionObject{},
				MerchantID:   uuid.NewString(),
				RejectedBy:   uuid.NewString(),
			},
			wantErr: false,
		},
		{
			name: "ERROR: Count disbursement not match with request",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
			) {
				disbursementRepo.On(
					"CountByIDsAndMerchantID",
					mock.Anything,
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
				).Return(0)
			},
			input:   validInput,
			wantErr: true,
		},
		{
			name: "ERROR: BeginTransaction error",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
			) {
				disbursementRepo.On(
					"CountByIDsAndMerchantID",
					mock.Anything,
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
				).Return(1)

				disbursementRepo.On(
					"GetByIDs",
					mock.Anything,
					constant.ArrayStringMockType(),
				).Return(validRejectedDisbursement, nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), constant.ErrSomeErrorForUnitTest)
			},
			input:   validInput,
			wantErr: true,
		},
		{
			name: "ERROR: Reject action got error",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
			) {
				disbursementRepo.On(
					"CountByIDsAndMerchantID",
					mock.Anything,
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
				).Return(1)

				disbursementRepo.On(
					"GetByIDs",
					mock.Anything,
					constant.ArrayStringMockType(),
				).Return(validRejectedDisbursement, nil)

				disbursementRepo.On(
					"Reject",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				disbursementRepo.On("RollbackTransaction", mock.Anything).Return(nil)
			},
			input:   validInput,
			wantErr: true,
		},
		{
			name: "ERROR: Reject action and rollback function got error",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
			) {
				disbursementRepo.On(
					"CountByIDsAndMerchantID",
					mock.Anything,
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
				).Return(1)

				disbursementRepo.On(
					"GetByIDs",
					mock.Anything,
					constant.ArrayStringMockType(),
				).Return(validRejectedDisbursement, nil)

				disbursementRepo.On(
					"Reject",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				disbursementRepo.On("RollbackTransaction", mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
			input:   validInput,
			wantErr: true,
		},
		{
			name: "ERROR: Reject action got error on CommitTransaction",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
			) {
				disbursementRepo.On(
					"CountByIDsAndMerchantID",
					mock.Anything,
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
				).Return(1)

				disbursementRepo.On(
					"GetByIDs",
					mock.Anything,
					constant.ArrayStringMockType(),
				).Return(validRejectedDisbursement, nil)

				disbursementRepo.On(
					"Reject",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				disbursementRepo.On("RollbackTransaction", mock.Anything).Return(nil)
				disbursementRepo.On("CommitTransaction", mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
			input:   validInput,
			wantErr: true,
		},
		{
			name: "ERROR: Rollback transaction",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
			) {
				disbursementRepo.On(
					"CountByIDsAndMerchantID",
					mock.Anything,
					constant.ArrayStringMockType(),
					mock.AnythingOfType("string"),
				).Return(1)

				disbursementRepo.On(
					"GetByIDs",
					mock.Anything,
					constant.ArrayStringMockType(),
				).Return(validRejectedDisbursement, nil)

				disbursementRepo.On(
					"Reject",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).
					Return(context.Background(), nil)
				disbursementRepo.On("RollbackTransaction", mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
				disbursementRepo.On("CommitTransaction", mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
			input:   validInput,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			disbursementRepoMock := repositoryMocks.NewIDisbursementRepository(t)
			snapCoreRepoMock := repositoryMocks.NewISnapCoreRepository(t)
			orchSvcMock := serviceMocks.NewIOrchestratorService(t)
			beneficiaryAccSvcMock := serviceMocks.NewIBeneficiaryAccountService(t)
			mockGcs := gcsMocks.NewGCSService(t)

			statusHistoryRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
			statusHistoryRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

			tc.mocksSetup(disbursementRepoMock, mockGcs)

			svc := New(
				&conf, pdkLoggerMock, merchantRepo, disbursementRepoMock, snapCoreRepoMock, nil,
				WithOrchestratorService(orchSvcMock), WithBeneficiaryAccService(beneficiaryAccSvcMock), WithGoogleCloudStorage(mockGcs),
				WithStatusHistoriesRepository(statusHistoryRepo),
			)
			ctx := context.Background()
			_, err := svc.Reject(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			disbursementRepoMock.AssertExpectations(t)
			snapCoreRepoMock.AssertExpectations(t)
			orchSvcMock.AssertExpectations(t)
			beneficiaryAccSvcMock.AssertExpectations(t)
		})
	}
}
