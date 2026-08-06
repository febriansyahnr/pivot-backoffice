package merchant

import (
	"context"
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	rabbitmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRedisExt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	logger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateKYC(t *testing.T) {
	mockMerchantRepo := repositoryMock.NewIMerchantRepository(t)
	mockLogger, _ := logger.NewZapLogger(logger.Config{})
	validMerchantID := "valid-update-merchant-id"
	mockRedis := mockRedisExt.NewIRedisExt(t)
	userSvc := serviceMock.NewIUserService(t)
	rabbitmq := rabbitmqMock.NewRabbitMQExt(t)

	ctx := context.Background()
	mid := "1001"

	rabbitmq.On("PublishMerchantCallback", mock.Anything, mock.Anything).Return(nil)

	service := MerchantService{
		repo:        mockMerchantRepo,
		logger:      mockLogger,
		redis:       mockRedis,
		rabbitMqExt: rabbitmq,
		config: &config.Config{
			MerchantConfig: config.MerchantConfig{
				CacheStatusDurationInMinutes: 10,
			},
		},
		UserSvc: userSvc,
	}

	testCases := []struct {
		name      string
		payload   merchant.UpdateMerchantKYCRequest
		setupMock func()
		shouldErr bool
		wantErr   error
		want      *merchant.UpdateMerchantKYCResponse
	}{
		{
			name: "when failed to get merchant info, then should return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID: validMerchantID,
				KYCStatus:  constant.KYCStatusApproved,
			},
			setupMock: func() {
				mockMerchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), validMerchantID).
					Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr: true,
			wantErr:   constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "when the merchant was not found, then should return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID: validMerchantID,
				KYCStatus:  constant.KYCStatusApproved,
			},
			setupMock: func() {
				mockMerchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), validMerchantID).
					Return(nil, nil).Once()
			},
			shouldErr: true,
			wantErr:   constant.ErrMerchantNotFound,
		},
		{
			name: "when the merchant kyc status is waiting for submission but it want to change to approve, then should return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID: validMerchantID,
				KYCStatus:  constant.KYCStatusApproved,
			},
			setupMock: func() {
				mockMerchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), validMerchantID).
					Return(&merchant.Merchant{
						UUID: validMerchantID,
						KYCStatus: sql.NullString{
							String: constant.KYCStatusWaitingForDocument,
							Valid:  true,
						},
					}, nil).Once()
			},
			shouldErr: true,
			wantErr:   pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenChangeKYCStatus),
		},
		{
			name: "when failed to update kyc, then should return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID: validMerchantID,
				KYCStatus:  constant.KYCStatusApproved,
			},
			setupMock: func() {
				mockMerchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), validMerchantID).
					Return(&merchant.Merchant{
						UUID: validMerchantID,
						KYCStatus: sql.NullString{
							String: constant.KYCStatusInReview,
							Valid:  true,
						},
						Status: constant.MerchantStatusCreated,
					}, nil).Once()

				mockMerchantRepo.On("GenerateNewMID", mock.Anything).Return(mid, nil).Once()
				mockMerchantRepo.On("UpdateKYC", mock.Anything, merchant.UpdateMerchantKYCRequest{
					MerchantID:     validMerchantID,
					KYCStatus:      constant.KYCStatusApproved,
					MerchantStatus: constant.MerchantStatusActive,
					MID:            &mid,
				}).Return(constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr: true,
			wantErr:   constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "when update kyc (approved) succeeded, then should not return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID: validMerchantID,
				KYCStatus:  constant.KYCStatusApproved,
			},
			setupMock: func() {
				mockMerchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), validMerchantID).
					Return(&merchant.Merchant{
						UUID: validMerchantID,
						KYCStatus: sql.NullString{
							String: constant.KYCStatusInReview,
							Valid:  true,
						},
						Status: constant.MerchantStatusCreated,
					}, nil).Once()
				mockMerchantRepo.On("GenerateNewMID", mock.Anything).Return(mid, nil).Once()
				mockMerchantRepo.On("UpdateKYC", mock.Anything, merchant.UpdateMerchantKYCRequest{
					MerchantID:     validMerchantID,
					KYCStatus:      constant.KYCStatusApproved,
					MerchantStatus: constant.MerchantStatusActive,
					MID:            &mid,
				}).Return(nil).Once()
				userSvc.On("FindUserByEmail", mock.Anything, mock.Anything).Once().Return(&user.User{
					MerchantId: validMerchantID,
				}, nil)
				userSvc.On("SendGeneratedInvitationURL", mock.Anything, mock.Anything).Once().Return(constant.ErrSomeErrorForUnitTest)
				mockRedis.On("Del", mock.Anything, constant.StringMockType()).Return(redis.NewIntResult(0, nil)).Once()
			},
			shouldErr: false,
			want: &merchant.UpdateMerchantKYCResponse{
				UUID:   validMerchantID,
				Status: constant.KYCStatusApproved,
			},
		},
		{
			name: "when update kyc (rejected) succeeded, then should not return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID: validMerchantID,
				KYCStatus:  constant.KYCStatusRejected,
			},
			setupMock: func() {
				mockMerchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), validMerchantID).
					Return(&merchant.Merchant{
						UUID: validMerchantID,
						KYCStatus: sql.NullString{
							Valid: true, String: constant.KYCStatusInReview,
						},
						Status: constant.MerchantStatusCreated,
					}, nil).Once()
				mockMerchantRepo.On("UpdateKYC", mock.Anything, merchant.UpdateMerchantKYCRequest{
					MerchantID:     validMerchantID,
					KYCStatus:      constant.KYCStatusRejected,
					MerchantStatus: constant.MerchantStatusCreated,
				}).Return(nil).Once()
				mockRedis.On("Del", mock.Anything, constant.StringMockType()).Return(redis.NewIntResult(0, nil)).Once()
			},
			shouldErr: false,
			want: &merchant.UpdateMerchantKYCResponse{
				UUID:   validMerchantID,
				Status: constant.KYCStatusRejected,
			},
		},
		{
			name: "when failed to generate MID, then should return error",
			payload: merchant.UpdateMerchantKYCRequest{
				MerchantID: validMerchantID,
				KYCStatus:  constant.KYCStatusApproved,
			},
			setupMock: func() {
				mockMerchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), validMerchantID).
					Return(&merchant.Merchant{
						UUID: validMerchantID,
						KYCStatus: sql.NullString{
							String: constant.KYCStatusInReview,
							Valid:  true,
						},
						Status: constant.MerchantStatusCreated,
					}, nil).Once()

				mockMerchantRepo.On("GenerateNewMID", mock.Anything).Return(mid, constant.ErrSomeErrorForUnitTest).Once()
			},
			shouldErr: true,
			wantErr:   constant.ErrSomeErrorForUnitTest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()
			res, err := service.UpdateKYC(ctx, tc.payload)

			if tc.shouldErr {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, res)
		})
	}
}

func TestAllowedChangeKYCStatus(t *testing.T) {
	service := &MerchantService{}

	testCases := []struct {
		name   string
		before string
		after  string
		want   bool
	}{
		{
			name:   "valid transition from WaitingForDocumentSubmission to DocumentInReview",
			before: constant.KYCStatusWaitingForDocument,
			after:  constant.KYCStatusInReview,
			want:   true,
		},
		{
			name:   "invalid transition from WaitingForDocumentSubmission to Approved",
			before: constant.KYCStatusWaitingForDocument,
			after:  constant.KYCStatusApproved,
			want:   false,
		},
		{
			name:   "valid transition from DocumentInReview to Approved",
			before: constant.KYCStatusInReview,
			after:  constant.KYCStatusApproved,
			want:   true,
		},
		{
			name:   "valid transition from DocumentInReview to Rejected",
			before: constant.KYCStatusInReview,
			after:  constant.KYCStatusRejected,
			want:   true,
		},
		{
			name:   "valid transition from DocumentInReview to RequireChange",
			before: constant.KYCStatusInReview,
			after:  constant.KYCStatusNeedResubmission,
			want:   true,
		},
		{
			name:   "invalid transition from DocumentInReview to WaitingForDocumentSubmission",
			before: constant.KYCStatusInReview,
			after:  constant.KYCStatusWaitingForDocument,
			want:   false,
		},
		{
			name:   "valid transition from RequireChange to DocumentInReview",
			before: constant.KYCStatusNeedResubmission,
			after:  constant.KYCStatusInReview,
			want:   true,
		},
		{
			name:   "invalid transition from RequireChange to Approved",
			before: constant.KYCStatusNeedResubmission,
			after:  constant.KYCStatusApproved,
			want:   false,
		},
		{
			name:   "invalid transition from NotRequired to WaitingForDocumentSubmission",
			before: constant.KYCStatusNotRequired,
			after:  constant.KYCStatusWaitingForDocument,
			want:   false,
		},
		{
			name:   "invalid transition from NotRequired to DocumentInReview",
			before: constant.KYCStatusNotRequired,
			after:  constant.KYCStatusInReview,
			want:   false,
		},
		{
			name:   "invalid transition from NotRequired to Approved",
			before: constant.KYCStatusNotRequired,
			after:  constant.KYCStatusApproved,
			want:   false,
		},
		{
			name:   "when the status is not recognized, then should return false",
			before: "",
			after:  constant.KYCStatusApproved,
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := service.AllowedChangeKYCStatus(tc.before, tc.after)
			assert.Equal(t, tc.want, got)
		})
	}
}
