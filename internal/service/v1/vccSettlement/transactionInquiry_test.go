package vccsettlement

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cimbProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cimbProcessor"
	merchantRcnModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantRcn"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRcnTransactionInquiry(t *testing.T) {
	validRequest := &vccSettlement.VccTransactionInquiryRequest{
		RcnId:        "rcn-123",
		MerchantId:   "merchant-456",
		RecordType:   "ST",
		BillingCycle: "02",
		PostingDate:  "20250201",
	}

	invalidRequest := &vccSettlement.VccTransactionInquiryRequest{
		RcnId:        "rcn-123",
		MerchantId:   "merchant-456",
		RecordType:   "ST",
		BillingCycle: "99", // Invalid billing cycle
		PostingDate:  "20250201",
	}

	validMerchantRcn := &merchantRcnModel.MerchantRcnDetail{
		CardNumber: "1234567890123456",
	}

	testCases := []struct {
		name      string
		request   *vccSettlement.VccTransactionInquiryRequest
		mockSetup func(
			rabbitMq *rabbitMqMocks.RabbitMQExt,
			merchantRcnSvc *serviceMocks.IMerchantRcnService,
		)
		wantErr    bool
		assertResp func(t *testing.T, resp *vccSettlement.VccTransactionInquiryResponse)
	}{
		{
			name:    "Success",
			request: validRequest,
			mockSetup: func(rabbitMq *rabbitMqMocks.RabbitMQExt, merchantRcnSvc *serviceMocks.IMerchantRcnService) {
				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(validMerchantRcn, nil).Once()

				rabbitMq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*vccSettlement.VccTransactionInquiryRequest")).
					Return(nil).Once()
			},
			wantErr: false,
			assertResp: func(t *testing.T, resp *vccSettlement.VccTransactionInquiryResponse) {
				assert.NotEmpty(t, resp.PartnerReferenceNo)
			},
		},
		{
			name:    "Failure - Validation error",
			request: invalidRequest,
			mockSetup: func(rabbitMq *rabbitMqMocks.RabbitMQExt, merchantRcnSvc *serviceMocks.IMerchantRcnService) {
				// No mocks needed - validation happens before any service calls
			},
			wantErr: true,
			assertResp: func(t *testing.T, resp *vccSettlement.VccTransactionInquiryResponse) {
				assert.Nil(t, resp)
			},
		},
		{
			name:    "Failure - GetRcnDetail error",
			request: validRequest,
			mockSetup: func(rabbitMq *rabbitMqMocks.RabbitMQExt, merchantRcnSvc *serviceMocks.IMerchantRcnService) {
				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(nil, errors.New("rcn not found")).Once()
			},
			wantErr: true,
			assertResp: func(t *testing.T, resp *vccSettlement.VccTransactionInquiryResponse) {
				assert.Nil(t, resp)
			},
		},
		{
			name:    "Failure - RabbitMQ publish error",
			request: validRequest,
			mockSetup: func(rabbitMq *rabbitMqMocks.RabbitMQExt, merchantRcnSvc *serviceMocks.IMerchantRcnService) {
				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(validMerchantRcn, nil).Once()

				rabbitMq.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.AnythingOfType("*vccSettlement.VccTransactionInquiryRequest")).
					Return(errors.New("connection failed")).Once()
			},
			wantErr: true,
			assertResp: func(t *testing.T, resp *vccSettlement.VccTransactionInquiryResponse) {
				assert.Nil(t, resp)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCache := redisMocks.NewIRedisExt(t)
			mockRabbitMq := rabbitMqMocks.NewRabbitMQExt(t)
			mockMerchantRcnSvc := serviceMocks.NewIMerchantRcnService(t)
			mockNotificationSvc := serviceMocks.NewINotificationService(t)
			mockCimbProcessor := repositoryMocks.NewICimbProcessorRepository(t)
			mockVccRepo := repositoryMocks.NewIVCCSettlementRepository(t)

			tc.mockSetup(mockRabbitMq, mockMerchantRcnSvc)

			// Setup service
			cfg := config.Config{
				ServiceName: "test-service",
			}
			svc := New(cfg, mockLogger, mockMerchantRcnSvc, mockNotificationSvc, mockCimbProcessor, mockVccRepo, mockCache, mockRabbitMq)

			// Execute
			resp, err := svc.RcnTransactionInquiry(context.Background(), tc.request)

			// Assert
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.assertResp != nil {
				tc.assertResp(t, resp)
			}
		})
	}
}

func TestProcessRcnTransactionInquiry(t *testing.T) {
	validRequest := &vccSettlement.VccTransactionInquiryRequest{
		RcnId:              "rcn-123",
		MerchantId:         "merchant-456",
		RecordType:         "A",
		BillingCycle:       "202502",
		PostingDate:        "20250201",
		PartnerReferenceNo: "partner-ref-789",
	}

	validMerchantRcn := &merchantRcnModel.MerchantRcnDetail{
		CardNumber: "1234567890123456",
	}

	singlePageResponse := &vccSettlement.ProcessorVccTransactionInquiryResponse{
		HasNextPage:     false,
		TransactionData: []vccSettlement.VccTransactionInquiryTrxData{{}, {}}, // 2 transactions
	}

	multiPageResponse1 := &vccSettlement.ProcessorVccTransactionInquiryResponse{
		HasNextPage:     true,
		TransactionData: []vccSettlement.VccTransactionInquiryTrxData{{}, {}},
	}

	multiPageResponse2 := &vccSettlement.ProcessorVccTransactionInquiryResponse{
		HasNextPage:     false,
		TransactionData: []vccSettlement.VccTransactionInquiryTrxData{{}}, // Last page with 1 transaction
	}

	testCases := []struct {
		name      string
		request   *vccSettlement.VccTransactionInquiryRequest
		mockSetup func(
			cache *redisMocks.IRedisExt,
			merchantRcnSvc *serviceMocks.IMerchantRcnService,
			notificationSvc *serviceMocks.INotificationService,
			cimbProcessor *repositoryMocks.ICimbProcessorRepository,
			vccRepo *repositoryMocks.IVCCSettlementRepository,
		)
		wantErr  bool
		assertFn func(t *testing.T, err error, mocks struct {
			cache           *redisMocks.IRedisExt
			merchantRcnSvc  *serviceMocks.IMerchantRcnService
			notificationSvc *serviceMocks.INotificationService
			cimbProcessor   *repositoryMocks.ICimbProcessorRepository
			vccRepo         *repositoryMocks.IVCCSettlementRepository
		})
	}{
		{
			name:    "Success - Single page, no previous state",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				// Lock acquisition succeeds
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(true, nil)).Once()

				// No previous state
				cache.On("Get", mock.Anything, mock.Anything).
					Return(redis.NewStringResult("", redis.Nil)).Once()

				// Get merchant RCN
				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(validMerchantRcn, nil).Once()

				// Soft delete existing data
				vccRepo.On("Delete", mock.Anything, validRequest.RcnId, mock.AnythingOfType("time.Time")).
					Return(nil).Once()

				// Fetch first page
				cimbProcessor.On("InquiryTransactionCorporateCreditCard", mock.Anything, mock.MatchedBy(func(req *cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest) bool {
					return req.Page == 1
				})).Return(singlePageResponse, nil).Once()

				// Insert data
				vccRepo.On("BulkInsert", mock.Anything, mock.AnythingOfType("[]*vccSettlement.VccSettlement")).
					Return(nil).Once()

				// Store state
				cache.On("Set", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewStatusResult("OK", nil)).Once()

				// Delete state at end
				cache.On("Del", mock.Anything, mock.Anything).
					Return(redis.NewIntResult(1, nil)).Twice() // stateKey + lockKey
			},
			wantErr: false,
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				// Verify BulkInsert was called once
				mocks.vccRepo.AssertNumberOfCalls(t, "BulkInsert", 1)
				// Verify Delete (soft delete) was called
				mocks.vccRepo.AssertCalled(t, "Delete", mock.Anything, validRequest.RcnId, mock.AnythingOfType("time.Time"))
			},
		},
		{
			name:    "Success - Multiple pages",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(true, nil)).Once()

				cache.On("Get", mock.Anything, mock.Anything).
					Return(redis.NewStringResult("", redis.Nil)).Once()

				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(validMerchantRcn, nil).Once()

				vccRepo.On("Delete", mock.Anything, validRequest.RcnId, mock.AnythingOfType("time.Time")).
					Return(nil).Once()

				// Page 1
				cimbProcessor.On("InquiryTransactionCorporateCreditCard", mock.Anything, mock.MatchedBy(func(req *cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest) bool {
					return req.Page == 1
				})).Return(multiPageResponse1, nil).Once()

				vccRepo.On("BulkInsert", mock.Anything, mock.AnythingOfType("[]*vccSettlement.VccSettlement")).
					Return(nil).Once()

				cache.On("Set", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewStatusResult("OK", nil)).Once()

				// Page 2
				cimbProcessor.On("InquiryTransactionCorporateCreditCard", mock.Anything, mock.MatchedBy(func(req *cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest) bool {
					return req.Page == 2
				})).Return(multiPageResponse2, nil).Once()

				vccRepo.On("BulkInsert", mock.Anything, mock.AnythingOfType("[]*vccSettlement.VccSettlement")).
					Return(nil).Once()

				cache.On("Set", mock.Anything, mock.Anything, "2", mock.Anything).
					Return(redis.NewStatusResult("OK", nil)).Once()

				cache.On("Del", mock.Anything, mock.Anything).
					Return(redis.NewIntResult(1, nil)).Twice()
			},
			wantErr: false,
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				// Verify BulkInsert was called twice (2 pages)
				mocks.vccRepo.AssertNumberOfCalls(t, "BulkInsert", 2)
				// Verify last page data was inserted
				mocks.cimbProcessor.AssertCalled(t, "InquiryTransactionCorporateCreditCard", mock.Anything, mock.MatchedBy(func(req *cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest) bool {
					return req.Page == 2
				}))
			},
		},
		{
			name:    "Success - Resume from saved state (page 3)",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(true, nil)).Once()

				// Previous state: page 2 was last processed, so resume from page 2
				cache.On("Get", mock.Anything, mock.Anything).
					Return(redis.NewStringResult("2", nil)).Once()

				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(validMerchantRcn, nil).Once()

				// No soft delete since we're resuming (page != 1)

				// Start from page 2
				cimbProcessor.On("InquiryTransactionCorporateCreditCard", mock.Anything, mock.MatchedBy(func(req *cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest) bool {
					return req.Page == 2
				})).Return(singlePageResponse, nil).Once()

				vccRepo.On("BulkInsert", mock.Anything, mock.AnythingOfType("[]*vccSettlement.VccSettlement")).
					Return(nil).Once()

				cache.On("Set", mock.Anything, mock.Anything, "2", mock.Anything).
					Return(redis.NewStatusResult("OK", nil)).Once()

				cache.On("Del", mock.Anything, mock.Anything).
					Return(redis.NewIntResult(1, nil)).Twice()
			},
			wantErr: false,
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				// Verify Delete was NOT called (no soft delete on resume)
				mocks.vccRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
				// Verify started from page 2
				mocks.cimbProcessor.AssertCalled(t, "InquiryTransactionCorporateCreditCard", mock.Anything, mock.MatchedBy(func(req *cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest) bool {
					return req.Page == 2
				}))
			},
		},
		{
			name:    "Failure - Lock already acquired (concurrent processing)",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				// Lock acquisition fails (another process is running)
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(false, nil)).Once()
			},
			wantErr: false, // Returns nil when lock can't be acquired
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				// Should not call any other services
				mocks.merchantRcnSvc.AssertNotCalled(t, "GetRcnDetail", mock.Anything, mock.Anything, mock.Anything)
				mocks.cimbProcessor.AssertNotCalled(t, "InquiryTransactionCorporateCreditCard", mock.Anything, mock.Anything)
			},
		},
		{
			name:    "Failure - Lock acquisition error",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				// Lock acquisition fails with error
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(false, errors.New("redis connection error"))).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				assert.Equal(t, constant.ErrAcquireTransactionInquiryLock, err)
			},
		},
		{
			name:    "Failure - Merchant RCN not found",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(true, nil)).Once()

				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(nil, errors.New("rcn not found")).Once()

				cache.On("Del", mock.Anything, mock.Anything).
					Return(redis.NewIntResult(1, nil)).Once() // lockKey cleanup
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				assert.Contains(t, err.Error(), "rcn not found")
				// Should not proceed to API call
				mocks.cimbProcessor.AssertNotCalled(t, "InquiryTransactionCorporateCreditCard", mock.Anything, mock.Anything)
			},
		},
		{
			name:    "Failure - Soft delete fails",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(true, nil)).Once()

				cache.On("Get", mock.Anything, mock.Anything).
					Return(redis.NewStringResult("", redis.Nil)).Once()

				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(validMerchantRcn, nil).Once()

				// Soft delete fails
				vccRepo.On("Delete", mock.Anything, validRequest.RcnId, mock.AnythingOfType("time.Time")).
					Return(errors.New("database error")).Once()

				cache.On("Del", mock.Anything, mock.Anything).
					Return(redis.NewIntResult(1, nil)).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				assert.Equal(t, constant.ErrProcessSettlementTransactionInquiry, err)
				// Should not proceed to API call
				mocks.cimbProcessor.AssertNotCalled(t, "InquiryTransactionCorporateCreditCard", mock.Anything, mock.Anything)
			},
		},
		{
			name:    "Failure - API error on first page",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(true, nil)).Once()

				cache.On("Get", mock.Anything, mock.Anything).
					Return(redis.NewStringResult("", redis.Nil)).Once()

				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(validMerchantRcn, nil).Once()

				vccRepo.On("Delete", mock.Anything, validRequest.RcnId, mock.AnythingOfType("time.Time")).
					Return(nil).Once()

				// API error
				cimbProcessor.On("InquiryTransactionCorporateCreditCard", mock.Anything, mock.Anything).
					Return(nil, errors.New("API connection failed")).Once()

				// Notification should be sent
				notificationSvc.On("SendVccSettlementTransactionAlert", mock.Anything, mock.AnythingOfType("*vccSettlement.VccTransactionInquiryAlert")).
					Return(nil).Once()

				cache.On("Del", mock.Anything, mock.Anything).
					Return(redis.NewIntResult(1, nil)).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				assert.Contains(t, err.Error(), "API connection failed")
				// Should not call BulkInsert
				mocks.vccRepo.AssertNotCalled(t, "BulkInsert", mock.Anything, mock.Anything)
			},
		},
		{
			name:    "Failure - Database insert error on first page",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(true, nil)).Once()

				cache.On("Get", mock.Anything, mock.Anything).
					Return(redis.NewStringResult("", redis.Nil)).Once()

				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(validMerchantRcn, nil).Once()

				vccRepo.On("Delete", mock.Anything, validRequest.RcnId, mock.AnythingOfType("time.Time")).
					Return(nil).Once()

				cimbProcessor.On("InquiryTransactionCorporateCreditCard", mock.Anything, mock.Anything).
					Return(singlePageResponse, nil).Once()

				// BulkInsert fails
				vccRepo.On("BulkInsert", mock.Anything, mock.AnythingOfType("[]*vccSettlement.VccSettlement")).
					Return(errors.New("duplicate key violation")).Once()

				// Notification should be sent
				notificationSvc.On("SendVccSettlementTransactionAlert", mock.Anything, mock.AnythingOfType("*vccSettlement.VccTransactionInquiryAlert")).
					Return(nil).Once()

				cache.On("Del", mock.Anything, mock.Anything).
					Return(redis.NewIntResult(1, nil)).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				assert.Equal(t, constant.ErrProcessSettlementTransactionInquiry, err)
				// State should NOT be saved on error
				mocks.cache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:    "Failure - API error on second page (partial success)",
			request: validRequest,
			mockSetup: func(cache *redisMocks.IRedisExt, merchantRcnSvc *serviceMocks.IMerchantRcnService, notificationSvc *serviceMocks.INotificationService, cimbProcessor *repositoryMocks.ICimbProcessorRepository, vccRepo *repositoryMocks.IVCCSettlementRepository) {
				cache.On("SetNX", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewBoolResult(true, nil)).Once()

				cache.On("Get", mock.Anything, mock.Anything).
					Return(redis.NewStringResult("", redis.Nil)).Once()

				merchantRcnSvc.On("GetRcnDetail", mock.Anything, validRequest.RcnId, validRequest.MerchantId).
					Return(validMerchantRcn, nil).Once()

				vccRepo.On("Delete", mock.Anything, validRequest.RcnId, mock.AnythingOfType("time.Time")).
					Return(nil).Once()

				// Page 1 succeeds
				cimbProcessor.On("InquiryTransactionCorporateCreditCard", mock.Anything, mock.MatchedBy(func(req *cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest) bool {
					return req.Page == 1
				})).Return(multiPageResponse1, nil).Once()

				vccRepo.On("BulkInsert", mock.Anything, mock.AnythingOfType("[]*vccSettlement.VccSettlement")).
					Return(nil).Once()

				cache.On("Set", mock.Anything, mock.Anything, "1", mock.Anything).
					Return(redis.NewStatusResult("OK", nil)).Once()

				// Page 2 fails
				cimbProcessor.On("InquiryTransactionCorporateCreditCard", mock.Anything, mock.MatchedBy(func(req *cimbProcessorModel.InquiryTransactionCorporateCreditCardRequest) bool {
					return req.Page == 2
				})).Return(nil, errors.New("timeout")).Once()

				// Notification should be sent
				notificationSvc.On("SendVccSettlementTransactionAlert", mock.Anything, mock.AnythingOfType("*vccSettlement.VccTransactionInquiryAlert")).
					Return(nil).Once()

				cache.On("Del", mock.Anything, mock.Anything).
					Return(redis.NewIntResult(1, nil)).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error, mocks struct {
				cache           *redisMocks.IRedisExt
				merchantRcnSvc  *serviceMocks.IMerchantRcnService
				notificationSvc *serviceMocks.INotificationService
				cimbProcessor   *repositoryMocks.ICimbProcessorRepository
				vccRepo         *repositoryMocks.IVCCSettlementRepository
			}) {
				assert.Contains(t, err.Error(), "timeout")
				// Page 1 should have been inserted
				mocks.vccRepo.AssertNumberOfCalls(t, "BulkInsert", 1)
				// Page 1 state should have been saved (can resume from page 2)
				mocks.cache.AssertCalled(t, "Set", mock.Anything, mock.Anything, "1", mock.Anything)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCache := redisMocks.NewIRedisExt(t)
			mockRabbitMq := rabbitMqMocks.NewRabbitMQExt(t)
			mockMerchantRcnSvc := serviceMocks.NewIMerchantRcnService(t)
			mockNotificationSvc := serviceMocks.NewINotificationService(t)
			mockCimbProcessor := repositoryMocks.NewICimbProcessorRepository(t)
			mockVccRepo := repositoryMocks.NewIVCCSettlementRepository(t)

			tc.mockSetup(mockCache, mockMerchantRcnSvc, mockNotificationSvc, mockCimbProcessor, mockVccRepo)

			// Setup service
			cfg := config.Config{
				ServiceName: "test-service",
			}
			svc := New(cfg, mockLogger, mockMerchantRcnSvc, mockNotificationSvc, mockCimbProcessor, mockVccRepo, mockCache, mockRabbitMq)

			// Execute
			err := svc.ProcessRcnTransactionInquiry(context.Background(), tc.request)

			// Assert
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.assertFn != nil {
				tc.assertFn(t, err, struct {
					cache           *redisMocks.IRedisExt
					merchantRcnSvc  *serviceMocks.IMerchantRcnService
					notificationSvc *serviceMocks.INotificationService
					cimbProcessor   *repositoryMocks.ICimbProcessorRepository
					vccRepo         *repositoryMocks.IVCCSettlementRepository
				}{
					cache:           mockCache,
					merchantRcnSvc:  mockMerchantRcnSvc,
					notificationSvc: mockNotificationSvc,
					cimbProcessor:   mockCimbProcessor,
					vccRepo:         mockVccRepo,
				})
			}
		})
	}
}
