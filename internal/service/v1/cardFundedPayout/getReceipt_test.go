package cardFundedPayoutService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	"github.com/redis/go-redis/v9"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetReceipt(t *testing.T) {
	cfg := &config.Config{
		MerchantPortalConfig: config.MerchantPortalConfig{
			LogoURL:                     "https://example.com/logo.png",
			PaymentReceiptBackgroundURL: "https://example.com/bg.png",
		},
	}
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	cacheClient := redisMock.NewIRedisExt(t)
	gcs := gcsMock.NewGCSService(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	svc := New(cfg, log,
		WithDisbursementRepository(disbursementRepo),
		WithCacheClient(cacheClient),
		WithGCS(gcs),
	).(*service)

	validRequest := &cardFundedPayoutModel.GetReceiptRequest{
		PayoutID:   "payout-123",
		MerchantID: "merchant-123",
	}

	now := time.Now()

	testCases := []struct {
		name      string
		request   *cardFundedPayoutModel.GetReceiptRequest
		setupMock func()
		wantErr   bool
	}{
		{
			name:    "SUCCESS: Get receipt from cache",
			request: validRequest,
			setupMock: func() {
				// Lock acquired
				boolCmd := redis.NewBoolCmd(context.Background())
				boolCmd.SetVal(true)
				cacheClient.On("SetNX", mock.Anything, mock.Anything, true, 10*time.Second).Return(boolCmd).Once()

				// Return cached URL
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetVal("https://storage.googleapis.com/bucket/signed-url")
				cacheClient.On("Get", mock.Anything, mock.Anything).Return(cmd).Once()

				// Lock released
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				cacheClient.On("Del", mock.Anything, mock.Anything).Return(intCmd).Once()
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Failed to acquire lock",
			request: validRequest,
			setupMock: func() {
				boolCmd := redis.NewBoolCmd(context.Background())
				boolCmd.SetErr(errors.New("redis error"))
				cacheClient.On("SetNX", mock.Anything, mock.Anything, true, 10*time.Second).Return(boolCmd).Once()
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Lock already acquired (request in progress)",
			request: validRequest,
			setupMock: func() {
				boolCmd := redis.NewBoolCmd(context.Background())
				boolCmd.SetVal(false)
				cacheClient.On("SetNX", mock.Anything, mock.Anything, true, 10*time.Second).Return(boolCmd).Once()
			},
			wantErr: true,
		},
		{
			name:    "ERROR: GetCardFundedPayoutDetail returns error",
			request: validRequest,
			setupMock: func() {
				// Lock acquired
				boolCmd := redis.NewBoolCmd(context.Background())
				boolCmd.SetVal(true)
				cacheClient.On("SetNX", mock.Anything, mock.Anything, true, 10*time.Second).Return(boolCmd).Once()

				// Cache miss
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(errors.New("redis error"))
				cacheClient.On("Get", mock.Anything, mock.Anything).Return(cmd).Once()

				disbursementRepo.On("GetCardFundedPayoutDetail", mock.Anything, mock.Anything).
					Return(nil, errors.New("database error")).Once()

				// Lock released
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				cacheClient.On("Del", mock.Anything, mock.Anything).Return(intCmd).Once()
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Payout detail is nil",
			request: validRequest,
			setupMock: func() {
				// Lock acquired
				boolCmd := redis.NewBoolCmd(context.Background())
				boolCmd.SetVal(true)
				cacheClient.On("SetNX", mock.Anything, mock.Anything, true, 10*time.Second).Return(boolCmd).Once()

				// Cache miss
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(errors.New("redis error"))
				cacheClient.On("Get", mock.Anything, mock.Anything).Return(cmd).Once()

				disbursementRepo.On("GetCardFundedPayoutDetail", mock.Anything, mock.Anything).
					Return(nil, nil).Once()

				// Lock released
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				cacheClient.On("Del", mock.Anything, mock.Anything).Return(intCmd).Once()
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Transaction status is not SUCCESS",
			request: validRequest,
			setupMock: func() {
				// Lock acquired
				boolCmd := redis.NewBoolCmd(context.Background())
				boolCmd.SetVal(true)
				cacheClient.On("SetNX", mock.Anything, mock.Anything, true, 10*time.Second).Return(boolCmd).Once()

				// Cache miss
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(errors.New("redis error"))
				cacheClient.On("Get", mock.Anything, mock.Anything).Return(cmd).Once()

				disbursementRepo.On("GetCardFundedPayoutDetail", mock.Anything, mock.Anything).
					Return(&cardFundedPayoutModel.GetPayoutDetailResponse{
						UUID:              "payout-123",
						TransactionStatus: "PROCESSING",
						CreatedAt:         now,
					}, nil).Once()

				// Lock released
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				cacheClient.On("Del", mock.Anything, mock.Anything).Return(intCmd).Once()
			},
			wantErr: true,
		},
		{
			name:    "ERROR: GCS SetBucketWriter fails",
			request: validRequest,
			setupMock: func() {
				// Lock acquired
				boolCmd := redis.NewBoolCmd(context.Background())
				boolCmd.SetVal(true)
				cacheClient.On("SetNX", mock.Anything, mock.Anything, true, 10*time.Second).Return(boolCmd).Once()

				// Cache miss
				cmd := redis.NewStringCmd(context.Background())
				cmd.SetErr(errors.New("redis error"))
				cacheClient.On("Get", mock.Anything, mock.Anything).Return(cmd).Once()

				disbursementRepo.On("GetCardFundedPayoutDetail", mock.Anything, mock.Anything).
					Return(&cardFundedPayoutModel.GetPayoutDetailResponse{
						UUID:              "payout-123",
						TransactionStatus: constant.StatusSuccess,
						CreatedAt:         now,
						ReferenceID:       "ref-123",
						Amount:            "10000",
						Fee:               "1000",
						TotalAmount:       "11000",
						VendorName:        "Vendor A",
						Remarks:           "Test remarks",
						BankName:          "BCA",
						AccountNumber:     "1234567890",
						AccountName:       "John Doe",
					}, nil).Once()

				gcs.On("SetBucketWriter", mock.Anything, mock.Anything).Return(nil, errors.New("gcs error")).Once()

				// Lock released
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				cacheClient.On("Del", mock.Anything, mock.Anything).Return(intCmd).Once()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, err := svc.GetReceipt(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.ReceiptURL)
			}

			disbursementRepo.AssertExpectations(t)
			cacheClient.AssertExpectations(t)
			gcs.AssertExpectations(t)
		})
	}
}
