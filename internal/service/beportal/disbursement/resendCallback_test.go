package disbursementService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	rabbitMqExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResendDisbursementCallback(t *testing.T) {
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	merchantID := uuid.NewString()
	disbursementID := uuid.NewString()
	clientReferenceID := "client-ref-" + uuid.NewString()
	bulkID := uuid.NewString()
	databaseError := errors.New("database error")
	createdFrom := constant.DisbursementCreatedFromOpenApi

	tests := []struct {
		name          string
		request       *callbackModel.ResendCallbackRequest
		setupMocks    func(*repositoryMocks.IDisbursementRepository, *rabbitMqExtMocks.RabbitMQExt, *redisExtMocks.IRedisExt)
		expectedError string
	}{
		{
			name: "SUCCESS: Resend callback with ReferenceID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						ReferenceID: clientReferenceID,
						BulkID:      &bulkID,
						CreatedFrom: &createdFrom,
						Amount:      decimal.NewFromInt(100000),
					},
				}
				bulkDisbursement := &disbursementModel.BulkDisbursement{
					UUID:       bulkID,
					MerchantID: merchantID,
					Status:     constant.BulkDisbursementStatusDone,
				}

				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
				disbursementRepo.On("FindBulkDisbursementByID", mock.Anything, bulkID).Return(bulkDisbursement, nil)

				disbursementRepo.On(
					"CountByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(2)

				disbursementRepo.On(
					"SummarySuccessByBulkID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				disbursementRepo.On(
					"SummaryFailedByBulkID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				disbursementRepo.On(
					"SummaryPendingByBulkID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				disbursementRepo.On(
					"SummaryCancelledByBulkID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				disbursementRepo.On(
					"GetMerchantIDsForPayoutCallback", mock.Anything, mock.Anything,
				).Once().Return(nil, nil)

				rmq.On(
					"PublishMerchantCallback", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)

				boolResult := redis.BoolCmd{}
				boolResult.SetVal(true)
				redisMock.On(
					"SetNX", constant.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(&boolResult)
			},
			expectedError: "",
		},
		{
			name: "SUCCESS: Resend callback with ClientReferenceID, Error delete lock",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:        merchantID,
				ClientReferenceID: clientReferenceID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						ReferenceID: clientReferenceID,
						BulkID:      &bulkID,
						CreatedFrom: &createdFrom,
						Amount:      decimal.NewFromInt(100000),
					},
				}
				bulkDisbursement := &disbursementModel.BulkDisbursement{
					UUID:       bulkID,
					MerchantID: merchantID,
					Status:     constant.BulkDisbursementStatusDone,
				}

				disbursementRepo.On("FindByMerchantAndReference", mock.Anything, merchantID, clientReferenceID).Return(disbursement, nil)
				disbursementRepo.On("FindBulkDisbursementByID", mock.Anything, bulkID).Return(bulkDisbursement, nil)

				disbursementRepo.On(
					"CountByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Once().Return(2)

				disbursementRepo.On(
					"SummarySuccessByBulkID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				disbursementRepo.On(
					"SummaryFailedByBulkID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				disbursementRepo.On(
					"SummaryPendingByBulkID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				disbursementRepo.On(
					"SummaryCancelledByBulkID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(disbursementDashboardModel.SummaryTransactionDTO{})

				disbursementRepo.On(
					"GetMerchantIDsForPayoutCallback", mock.Anything, mock.Anything,
				).Once().Return(nil, nil)

				boolResult := redis.BoolCmd{}
				boolResult.SetVal(true)
				redisMock.On(
					"SetNX", constant.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(&boolResult)

				rmq.On(
					"PublishMerchantCallback", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)
			},
			expectedError: "",
		},
		{
			name: "ERROR: Database error when getting disbursement by ID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(nil, databaseError)
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, databaseError).Error(),
		},
		{
			name: "ERROR: Database error when getting disbursement by ClientReferenceID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:        merchantID,
				ClientReferenceID: clientReferenceID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursementRepo.On("FindByMerchantAndReference", mock.Anything, merchantID, clientReferenceID).Return(nil, databaseError)
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, databaseError).Error(),
		},
		{
			name: "ERROR: Disbursement not found by ID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(nil, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrNotFound, constant.ErrDisbursementNotFound).Error(),
		},
		{
			name: "ERROR: Disbursement not found by ClientReferenceID",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:        merchantID,
				ClientReferenceID: clientReferenceID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursementRepo.On("FindByMerchantAndReference", mock.Anything, merchantID, clientReferenceID).Return(nil, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrNotFound, constant.ErrDisbursementNotFound).Error(),
		},
		{
			name: "ERROR: Merchant ID does not match",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:       disbursementID,
						MerchantID: "different-merchant-" + uuid.NewString(),
					},
				}
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
			},
			expectedError: pkgErrors.New(response.HttpErrRequest, constant.ErrMerchantIsNotMatch).Error(),
		},
		{
			name: "ERROR: Disbursement not created from OPEN_API (nil CreatedFrom)",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						CreatedFrom: nil,
					},
				}
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
			},
			expectedError: "disbursement was not created from OPEN_API",
		},
		{
			name: "ERROR: Disbursement not created from OPEN_API (wrong value)",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				wrongCreatedFrom := "MERCHANT_PORTAL"
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						CreatedFrom: &wrongCreatedFrom,
					},
				}
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
			},
			expectedError: "disbursement was not created from OPEN_API",
		},
		{
			name: "ERROR: Disbursement does not have a bulk ID (nil BulkID)",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						CreatedFrom: &createdFrom,
						BulkID:      nil,
					},
				}
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
			},
			expectedError: "disbursement does not have a bulk ID",
		},
		{
			name: "ERROR: Disbursement does not have a bulk ID (empty string)",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				emptyBulkID := ""
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						CreatedFrom: &createdFrom,
						BulkID:      &emptyBulkID,
					},
				}
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
			},
			expectedError: "disbursement does not have a bulk ID",
		},
		{
			name: "ERROR: Database error when getting bulk disbursement",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						CreatedFrom: &createdFrom,
						BulkID:      &bulkID,
					},
				}
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
				disbursementRepo.On("FindBulkDisbursementByID", mock.Anything, bulkID).Return(nil, databaseError)
			},
			expectedError: pkgErrors.New(response.HttpErrDatabase, databaseError).Error(),
		},
		{
			name: "ERROR: Bulk disbursement not found",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						CreatedFrom: &createdFrom,
						BulkID:      &bulkID,
					},
				}
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
				disbursementRepo.On("FindBulkDisbursementByID", mock.Anything, bulkID).Return(nil, nil)
			},
			expectedError: "bulk disbursement not found",
		},
		{
			name: "ERROR: Bulk disbursement status is not DONE (status PENDING)",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						CreatedFrom: &createdFrom,
						BulkID:      &bulkID,
					},
				}
				bulkDisbursement := &disbursementModel.BulkDisbursement{
					UUID:       bulkID,
					MerchantID: merchantID,
					Status:     "PENDING",
				}
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
				disbursementRepo.On("FindBulkDisbursementByID", mock.Anything, bulkID).Return(bulkDisbursement, nil)
			},
			expectedError: "bulk disbursement status is not DONE (current: PENDING)",
		},
		{
			name: "ERROR: Bulk disbursement status is not DONE (status PROCESSING)",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						CreatedFrom: &createdFrom,
						BulkID:      &bulkID,
					},
				}
				bulkDisbursement := &disbursementModel.BulkDisbursement{
					UUID:       bulkID,
					MerchantID: merchantID,
					Status:     "PROCESSING",
				}
				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
				disbursementRepo.On("FindBulkDisbursementByID", mock.Anything, bulkID).Return(bulkDisbursement, nil)
			},
			expectedError: "bulk disbursement status is not DONE (current: PROCESSING)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Setup mocks
			disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
			bankAccountRepo := repositoryMocks.NewIBankAccountRepository(t)
			rabbitMq := rabbitMqExtMocks.NewRabbitMQExt(t)
			redisClient := redisExtMocks.NewIRedisExt(t)

			tt.setupMocks(disbursementRepo, rabbitMq, redisClient)

			// Create service
			svc := New(
				&conf,
				pdkLoggerMock,
				merchantRepo,
				disbursementRepo,
				snapCoreRepo,
				bankAccountRepo,
				WithRabbitMQClient(rabbitMq),
				WithRedisClient(redisClient),
			)

			// Execute
			err := svc.ResendDisbursementCallback(ctx, tt.request)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			// Verify all expectations were met
			disbursementRepo.AssertExpectations(t)
		})
	}
}

func TestResendDisbursementForSinglePayoutCallback(t *testing.T) {
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}

	merchantID := "targeted-merchant-id"
	disbursementID := uuid.NewString()
	clientReferenceID := "client-ref-" + uuid.NewString()
	bulkID := uuid.NewString()
	createdFrom := constant.DisbursementCreatedFromOpenApi

	tests := []struct {
		name          string
		request       *callbackModel.ResendCallbackRequest
		setupMocks    func(*repositoryMocks.IDisbursementRepository, *rabbitMqExtMocks.RabbitMQExt, *redisExtMocks.IRedisExt)
		expectedError string
	}{
		{
			name: "SUCCESS: Resend callback",
			request: &callbackModel.ResendCallbackRequest{
				MerchantID:  merchantID,
				ReferenceID: disbursementID,
			},
			setupMocks: func(disbursementRepo *repositoryMocks.IDisbursementRepository, rmq *rabbitMqExtMocks.RabbitMQExt, redisMock *redisExtMocks.IRedisExt) {
				disbursement := &disbursementModel.DisbursementWithTransaction{
					Disbursement: disbursementModel.Disbursement{
						UUID:        disbursementID,
						MerchantID:  merchantID,
						ReferenceID: clientReferenceID,
						BulkID:      &bulkID,
						CreatedFrom: &createdFrom,
						Amount:      decimal.NewFromInt(100000),
					},
					TransactionStatus: util.ValueToPtr(constant.StatusFailed),
				}
				bulkDisbursement := &disbursementModel.BulkDisbursement{
					UUID:       bulkID,
					MerchantID: merchantID,
					Status:     constant.BulkDisbursementStatusDone,
				}

				disbursementRepo.On("FindByID", mock.Anything, disbursementID).Return(disbursement, nil)
				disbursementRepo.On("FindBulkDisbursementByID", mock.Anything, bulkID).Return(bulkDisbursement, nil)

				disbursementRepo.On(
					"CountByBulkID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(1)

				disbursementRepo.On(
					"GetAllDisbursementByBulkID",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return([]*disbursementModel.DisbursementWithTransaction{
					{
						Disbursement: disbursementModel.Disbursement{
							UUID:                 uuid.NewString(),
							ReferenceID:          "ref-001",
							BulkID:               &bulkID,
							MerchantID:           merchantID,
							Status:               constant.DisbursementStatusApproved,
							BeneficiaryBankCode:  "002",
							BeneficiaryAccountNo: "9999999666660001",
							Amount:               decimal.NewFromInt(100000),
						},
						TransactionStatus:            util.ValueToPtr(constant.StatusFailed),
						TransactionReasonType:        util.ValueToPtr(constant.DisbursementReasonTypeInsufficientBalance),
						TransactionReasonDescription: util.ValueToPtr(constant.DisbursementReasonTypeInsufficientBalance),
					},
				}, nil)

				disbursementRepo.On(
					"GetMerchantIDsForPayoutCallback", mock.Anything, mock.Anything,
				).Once().Return(nil, nil)

				boolResult := redis.BoolCmd{}
				boolResult.SetVal(true)
				redisMock.On(
					"SetNX", constant.ValueCtxMockType(), mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(&boolResult)

				rmq.On(
					"PublishMerchantCallback", constant.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil)
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Setup mocks
			disbursementRepo := repositoryMocks.NewIDisbursementRepository(t)
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
			bankAccountRepo := repositoryMocks.NewIBankAccountRepository(t)
			rabbitMq := rabbitMqExtMocks.NewRabbitMQExt(t)
			redisClient := redisExtMocks.NewIRedisExt(t)

			tt.setupMocks(disbursementRepo, rabbitMq, redisClient)

			// Create service
			svc := New(
				&conf,
				pdkLoggerMock,
				merchantRepo,
				disbursementRepo,
				snapCoreRepo,
				bankAccountRepo,
				WithRabbitMQClient(rabbitMq),
				WithRedisClient(redisClient),
			)

			// Execute
			err := svc.ResendDisbursementCallback(ctx, tt.request)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			// Verify all expectations were met
			disbursementRepo.AssertExpectations(t)
		})
	}
}
