package disbursementService

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	gcsMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApprovalAction(t *testing.T) {
	disbursementUUID := uuid.NewString()

	rejectedRequest := &disbursementModel.ApprovalActionsRequest{
		BulkID:     uuid.NewString(),
		UserID:     uuid.NewString(),
		MerchantID: uuid.NewString(),
		Approve:    []disbursementModel.ApproveActionObject{},
		Reject: []disbursementModel.RejectActionObject{
			{DisbursementID: disbursementUUID},
		},
	}

	approvedRequest := &disbursementModel.ApprovalActionsRequest{
		BulkID:     uuid.NewString(),
		UserID:     uuid.NewString(),
		MerchantID: uuid.NewString(),
		Approve: []disbursementModel.ApproveActionObject{
			{DisbursementID: uuid.NewString()},
		},
		Reject: []disbursementModel.RejectActionObject{},
	}

	conf := config.Config{
		Environment: c.EnvironmentStaging,
	}

	validRejectedDisbursement := []*disbursementModel.Disbursement{
		{
			UUID:                disbursementUUID,
			BeneficiaryBankName: util.ValueToPtr("BRI"),
		},
	}

	trxCloser := serviceMocks.NewITransactionCloser(t)
	trxCloser.On("Close", mock.Anything, c.BoolMockType()).Return(nil)

	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	parentId := "dd915031-db2d-4245-b6a5-910b643d8794"
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	statusHistoriesRepo := repositoryMocks.NewIStatusHistoriesRepository(t)
	statusHistoriesRepo.On("Insert", mock.Anything, mock.Anything).Return(nil).Maybe()

	testCases := []struct {
		name       string
		mocksSetup func(disbursementRepo *repositoryMocks.IDisbursementRepository, gcsSvc *gcsMocks.GCSService, internal *serviceMocks.IDisbursementInternalService)
		input      *disbursementModel.ApprovalActionsRequest
		wantErr    bool
	}{
		{
			name: "SUCCESS:Reject Action",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
				internal *serviceMocks.IDisbursementInternalService,
			) {
				internal.On("Approve", c.ValueCtxMockType(), mock.Anything).Return(nil)
				disbursementRepo.On(
					"CountByIDsAndMerchantID", mock.Anything, c.ArrayStringMockType(), c.StringMockType(),
				).Return(1)

				disbursementRepo.On(
					"GetByIDs", mock.Anything, c.ArrayStringMockType(),
				).Return(validRejectedDisbursement, nil)

				disbursementRepo.On(
					"Reject", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				disbursementRepo.On("CommitTransaction", mock.Anything).Return(nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Return(nil)

				gcsSvc.On(
					"UploadFile", c.ValueCtxMockType(), c.StringMockType(), c.PtrBytesBufferMockType(), c.BoolMockType(),
				).Return(&gcs.UploadMultipart{PublicURL: "publicURL"}, nil)
				gcsSvc.On(
					"CreateSignedURL", c.ValueCtxMockType(), c.StringMockType(), c.DurationMockType(),
				).Return("signedURL", nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementRejectedFileByID", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Return(nil)
			},
			input: rejectedRequest, wantErr: false,
		},
		{
			name: "SUCCESS:Approve Action",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, _ *gcsMocks.GCSService, internal *serviceMocks.IDisbursementInternalService) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Once().Return(&merchantModel.Merchant{
					ParentID: sql.NullString{String: parentId},
				}, nil)
				disbursementRepo.On(
					"GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType(),
				).Return(&disbursementModel.ActionTransactionSummary{Total: 1}, nil)

				internal.On(
					"ValidateDailyTransactionLimit", c.ValueCtxMockType(), c.StringMockType(), c.Float64MockType(),
				).Return(trxCloser, nil)

				internal.On("Approve", c.ValueCtxMockType(), mock.Anything).Return(nil)
			},
			input: approvedRequest, wantErr: false,
		},
		{
			name: "ERROR:Find merchant by id",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, _ *gcsMocks.GCSService, _ *serviceMocks.IDisbursementInternalService) {
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input: approvedRequest, wantErr: true,
		},
		{
			name: "ERROR:Merchant not found",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, _ *gcsMocks.GCSService, _ *serviceMocks.IDisbursementInternalService) {
				merchantRepo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			input: approvedRequest, wantErr: true,
		},
		{
			name: "ERROR:Get Action Transaction Summary",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, _ *gcsMocks.GCSService, _ *serviceMocks.IDisbursementInternalService) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, mock.Anything,
				).Return(&merchantModel.Merchant{
					ParentID: sql.NullString{String: parentId},
				}, nil)
				disbursementRepo.On(
					"GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType(),
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input: approvedRequest, wantErr: true,
		},
		{
			name: "ERROR:Transaction Not Permitted",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, _ *gcsMocks.GCSService, _ *serviceMocks.IDisbursementInternalService) {
				disbursementRepo.On(
					"GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType(),
				).Return(nil, nil)
			},
			input: approvedRequest, wantErr: true,
		},
		{
			name: "ERROR:Validate Daily Transaction Limit",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository, _ *gcsMocks.GCSService, internal *serviceMocks.IDisbursementInternalService) {
				disbursementRepo.On(
					"GetActionTransactionSummary", mock.Anything, c.StringMockType(), c.ArrayStringMockType(),
				).Return(&disbursementModel.ActionTransactionSummary{Total: 1}, nil)

				internal.On(
					"ValidateDailyTransactionLimit", c.ValueCtxMockType(), c.StringMockType(), c.Float64MockType(),
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input: approvedRequest, wantErr: true,
		},
		{
			name: "ERROR:Update Disbursement Status By ID",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository,
				gcsSvc *gcsMocks.GCSService,
				internal *serviceMocks.IDisbursementInternalService,
			) {
				internal.On("Approve", c.ValueCtxMockType(), mock.Anything).Return(nil)
				disbursementRepo.On(
					"CountByIDsAndMerchantID", mock.Anything, c.ArrayStringMockType(), c.StringMockType(),
				).Return(1)

				disbursementRepo.On(
					"GetByIDs", mock.Anything, c.ArrayStringMockType(),
				).Return(validRejectedDisbursement, nil)

				disbursementRepo.On(
					"Reject", mock.Anything, c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil)

				disbursementRepo.On("BeginTransaction", mock.Anything).Return(context.Background(), nil)
				disbursementRepo.On("CommitTransaction", mock.Anything).Return(nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementStatusByID", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Return(c.ErrSomeErrorForUnitTest)

				gcsSvc.On(
					"UploadFile", c.ValueCtxMockType(), c.StringMockType(), c.PtrBytesBufferMockType(), c.BoolMockType(),
				).Return(&gcs.UploadMultipart{PublicURL: "publicURL"}, nil)
				gcsSvc.On(
					"CreateSignedURL", c.ValueCtxMockType(), c.StringMockType(), c.DurationMockType(),
				).Return("signedURL", nil)

				disbursementRepo.On(
					"UpdateBulkDisbursementRejectedFileByID", mock.Anything, c.StringMockType(), c.StringMockType(),
				).Return(nil)
			},
			input: rejectedRequest, wantErr: true,
		},
	}

	type testKey string

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			disbursementRepoMock := repositoryMocks.NewIDisbursementRepository(t)
			mockGcs := gcsMocks.NewGCSService(t)
			disbursementIntSvc := serviceMocks.NewIDisbursementInternalService(t)

			tc.mocksSetup(disbursementRepoMock, mockGcs, disbursementIntSvc)
			svc := New(
				&conf, logger, merchantRepo, disbursementRepoMock, nil, nil,
				WithStatusHistoriesRepository(statusHistoriesRepo), WithGoogleCloudStorage(mockGcs), WithDisbursementInternalService(disbursementIntSvc),
			)

			_, err := svc.ApprovalAction(context.WithValue(context.Background(), testKey("KEY"), "VALUE"), tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockGcs.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			disbursementIntSvc.AssertExpectations(t)
			disbursementRepoMock.AssertExpectations(t)

			os.RemoveAll(c.ExportTempDir)
		})
	}
}

func TestValidateBatchPayoutItems(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	testCases := []struct {
		name       string
		mocksSetup func(disbursementRepo *repositoryMocks.IDisbursementRepository)
		input      *disbursementModel.ApprovalActionsRequest
		expected   *disbursementModel.ApprovalActionsRequest
		wantErr    bool
	}{
		{
			name: "SUCCESS: Empty BulkID returns request unchanged",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				// No mocks needed as function returns early
			},
			input: &disbursementModel.ApprovalActionsRequest{
				BulkID:     "",
				UserID:     uuid.NewString(),
				MerchantID: uuid.NewString(),
				Approve: []disbursementModel.ApproveActionObject{
					{DisbursementID: uuid.NewString()},
				},
				Reject: []disbursementModel.RejectActionObject{
					{DisbursementID: uuid.NewString()},
				},
			},
			expected: &disbursementModel.ApprovalActionsRequest{
				BulkID:     "",
				UserID:     "", // Will be set by input
				MerchantID: "", // Will be set by input
				Approve: []disbursementModel.ApproveActionObject{
					{DisbursementID: ""}, // Will be set by input
				},
				Reject: []disbursementModel.RejectActionObject{
					{DisbursementID: ""}, // Will be set by input
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: All items are valid - no cleanup needed",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				validPayouts := []*disbursementModel.DisbursementWithTransaction{
					{Disbursement: disbursementModel.Disbursement{UUID: "valid-approve-id"}},
					{Disbursement: disbursementModel.Disbursement{UUID: "valid-reject-id"}},
				}
				disbursementRepo.On(
					"GetAllDisbursementByBulkID", mock.Anything, "valid-bulk-id",
				).Return(validPayouts, nil)
			},
			input: &disbursementModel.ApprovalActionsRequest{
				BulkID:     "valid-bulk-id",
				UserID:     uuid.NewString(),
				MerchantID: uuid.NewString(),
				Approve: []disbursementModel.ApproveActionObject{
					{DisbursementID: "valid-approve-id"},
				},
				Reject: []disbursementModel.RejectActionObject{
					{DisbursementID: "valid-reject-id"},
				},
			},
			expected: &disbursementModel.ApprovalActionsRequest{
				BulkID:     "valid-bulk-id",
				UserID:     "",
				MerchantID: "",
				Approve: []disbursementModel.ApproveActionObject{
					{DisbursementID: "valid-approve-id"},
				},
				Reject: []disbursementModel.RejectActionObject{
					{DisbursementID: "valid-reject-id"},
				},
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Reject unrelated approve items",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				validPayouts := []*disbursementModel.DisbursementWithTransaction{
					{Disbursement: disbursementModel.Disbursement{UUID: "valid-approve-id"}},
				}
				disbursementRepo.On(
					"GetAllDisbursementByBulkID", mock.Anything, "valid-bulk-id",
				).Return(validPayouts, nil)
			},
			input: &disbursementModel.ApprovalActionsRequest{
				BulkID:     "valid-bulk-id",
				UserID:     uuid.NewString(),
				MerchantID: uuid.NewString(),
				Approve: []disbursementModel.ApproveActionObject{
					{DisbursementID: "valid-approve-id"},
					{DisbursementID: "invalid-approve-id"},
				},
				Reject: []disbursementModel.RejectActionObject{},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "SUCCESS: Reject unrelated reject items",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				validPayouts := []*disbursementModel.DisbursementWithTransaction{
					{Disbursement: disbursementModel.Disbursement{UUID: "valid-reject-id"}},
				}
				disbursementRepo.On(
					"GetAllDisbursementByBulkID", mock.Anything, "valid-bulk-id",
				).Return(validPayouts, nil)
			},
			input: &disbursementModel.ApprovalActionsRequest{
				BulkID:     "valid-bulk-id",
				UserID:     uuid.NewString(),
				MerchantID: uuid.NewString(),
				Approve:    []disbursementModel.ApproveActionObject{},
				Reject: []disbursementModel.RejectActionObject{
					{DisbursementID: "valid-reject-id"},
					{DisbursementID: "invalid-reject-id"},
				},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "SUCCESS: Reject unrelated items from both approve and reject",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				validPayouts := []*disbursementModel.DisbursementWithTransaction{
					{Disbursement: disbursementModel.Disbursement{UUID: "valid-approve-id"}},
					{Disbursement: disbursementModel.Disbursement{UUID: "valid-reject-id"}},
				}
				disbursementRepo.On(
					"GetAllDisbursementByBulkID", mock.Anything, "valid-bulk-id",
				).Return(validPayouts, nil)
			},
			input: &disbursementModel.ApprovalActionsRequest{
				BulkID:     "valid-bulk-id",
				UserID:     uuid.NewString(),
				MerchantID: uuid.NewString(),
				Approve: []disbursementModel.ApproveActionObject{
					{DisbursementID: "valid-approve-id"},
					{DisbursementID: "invalid-approve-id"},
				},
				Reject: []disbursementModel.RejectActionObject{
					{DisbursementID: "valid-reject-id"},
					{DisbursementID: "invalid-reject-id"},
				},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "SUCCESS: Empty valid payouts reject all items",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				validPayouts := []*disbursementModel.DisbursementWithTransaction{}
				disbursementRepo.On(
					"GetAllDisbursementByBulkID", mock.Anything, "valid-bulk-id",
				).Return(validPayouts, nil)
			},
			input: &disbursementModel.ApprovalActionsRequest{
				BulkID:     "valid-bulk-id",
				UserID:     uuid.NewString(),
				MerchantID: uuid.NewString(),
				Approve: []disbursementModel.ApproveActionObject{
					{DisbursementID: "invalid-approve-id"},
				},
				Reject: []disbursementModel.RejectActionObject{
					{DisbursementID: "invalid-reject-id"},
				},
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "ERROR: GetAllDisbursementByBulkID returns error",
			mocksSetup: func(disbursementRepo *repositoryMocks.IDisbursementRepository) {
				disbursementRepo.On(
					"GetAllDisbursementByBulkID", mock.Anything, "invalid-bulk-id",
				).Return(nil, c.ErrSomeErrorForUnitTest)
			},
			input: &disbursementModel.ApprovalActionsRequest{
				BulkID:     "invalid-bulk-id",
				UserID:     uuid.NewString(),
				MerchantID: uuid.NewString(),
				Approve: []disbursementModel.ApproveActionObject{
					{DisbursementID: uuid.NewString()},
				},
				Reject: []disbursementModel.RejectActionObject{
					{DisbursementID: uuid.NewString()},
				},
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			disbursementRepoMock := repositoryMocks.NewIDisbursementRepository(t)
			tc.mocksSetup(disbursementRepoMock)

			svc := New(
				&config.Config{}, logger, nil, disbursementRepoMock, nil, nil,
			)

			result, err := svc.ValidateBatchPayoutItems(context.Background(), tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Set expected values from input for comparison
				if tc.expected.BulkID == "" && tc.input.BulkID == "" {
					tc.expected.UserID = tc.input.UserID
					tc.expected.MerchantID = tc.input.MerchantID
					tc.expected.Approve = tc.input.Approve
					tc.expected.Reject = tc.input.Reject
				} else {
					tc.expected.UserID = tc.input.UserID
					tc.expected.MerchantID = tc.input.MerchantID
				}

				assert.Equal(t, tc.expected.BulkID, result.BulkID)
				assert.Equal(t, tc.expected.UserID, result.UserID)
				assert.Equal(t, tc.expected.MerchantID, result.MerchantID)
				assert.Equal(t, len(tc.expected.Approve), len(result.Approve))
				assert.Equal(t, len(tc.expected.Reject), len(result.Reject))

				// Compare individual items
				for i, expectedApprove := range tc.expected.Approve {
					assert.Equal(t, expectedApprove.DisbursementID, result.Approve[i].DisbursementID)
				}

				for i, expectedReject := range tc.expected.Reject {
					assert.Equal(t, expectedReject.DisbursementID, result.Reject[i].DisbursementID)
				}
			}

			disbursementRepoMock.AssertExpectations(t)
		})
	}
}
