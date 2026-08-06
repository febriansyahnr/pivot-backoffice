package refundService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
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
	refundRepo := repositoryMock.NewIRefundRepository(t)
	cacheClient := redisMock.NewIRedisExt(t)
	gcsService := gcsMock.NewGCSService(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	svc := &RefundService{
		config:     cfg,
		logger:     log,
		refundRepo: refundRepo,
		redis:      cacheClient,
		gcs:        gcsService,
	}

	validRequest := &refundModel.GetRefundReceiptRequest{
		RefundID:   "refund-123",
		MerchantID: "merchant-123",
	}

	now := time.Now()

	testCases := []struct {
		name      string
		request   *refundModel.GetRefundReceiptRequest
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
			name:    "ERROR: GetRefundByID returns error",
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

				refundRepo.On("GetRefundByID", mock.Anything, validRequest.RefundID, validRequest.MerchantID).
					Return(nil, errors.New("database error")).Once()

				// Lock released
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				cacheClient.On("Del", mock.Anything, mock.Anything).Return(intCmd).Once()
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Refund not found",
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

				refundRepo.On("GetRefundByID", mock.Anything, validRequest.RefundID, validRequest.MerchantID).
					Return(nil, nil).Once()

				// Lock released
				intCmd := redis.NewIntCmd(context.Background())
				intCmd.SetVal(1)
				cacheClient.On("Del", mock.Anything, mock.Anything).Return(intCmd).Once()
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Refund status is not SUCCESS",
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

				refundRepo.On("GetRefundByID", mock.Anything, validRequest.RefundID, validRequest.MerchantID).
					Return(&refundModel.RefundResponse{
						ID:     "refund-123",
						Status: constant.RefundStatusPending,
						Amount: commonModel.Amount{Value: "50000", Currency: "IDR"},
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

				refundRepo.On("GetRefundByID", mock.Anything, validRequest.RefundID, validRequest.MerchantID).
					Return(&refundModel.RefundResponse{
						ID:                "refund-123",
						ClientReferenceID: "client-ref-123",
						PaymentSessionID:  "payment-session-123",
						Status:            constant.RefundStatusSuccess,
						Reason:            constant.RefundReasonDuplicate,
						Amount:            commonModel.Amount{Value: "50000", Currency: "IDR"},
						UpdatedAt:         now,
					}, nil).Once()

				gcsService.On("SetBucketWriter", mock.Anything, mock.Anything).Return(nil, errors.New("gcs error")).Once()

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
		})
	}
}

func TestBuildRefundDestinationDisplay(t *testing.T) {
	testCases := []struct {
		name     string
		refund   *refundModel.RefundResponse
		expected string
	}{
		{
			name: "CHANNEL: Credit card destination",
			refund: &refundModel.RefundResponse{
				DestinationType: "CHANNEL",
				ChannelDestination: &refundModel.ChannelDestination{
					PaymentMethod: "CREDIT_CARD",
					PaymentDetail: map[string]interface{}{
						"cardIssuing": "BRI",
						"cardBrand":   "Visa",
						"last4Digit":  "0002",
					},
				},
			},
			expected: "BRI (Visa) - *0002",
		},
		{
			name: "CHANNEL: QRIS destination",
			refund: &refundModel.RefundResponse{
				DestinationType: "CHANNEL",
				ChannelDestination: &refundModel.ChannelDestination{
					PaymentMethod:  "QRIS",
					PaymentChannel: "BNC",
					PaymentDetail: map[string]interface{}{
						"acquirer":     "BNC",
						"merchantName": "QM",
					},
				},
			},
			expected: "BNC (QM)",
		},
		{
			name: "CHANNEL: E-Wallet destination",
			refund: &refundModel.RefundResponse{
				DestinationType: "CHANNEL",
				ChannelDestination: &refundModel.ChannelDestination{
					PaymentMethod:  "EWALLET",
					PaymentChannel: "GOPAY",
					PaymentDetail: map[string]interface{}{
						"channel": "GOPAY",
					},
				},
			},
			expected: "GOPAY - GOPAY",
		},
		{
			name: "ACCOUNT: Bank transfer destination",
			refund: &refundModel.RefundResponse{
				DestinationType: "ACCOUNT",
				TransferDestination: &refundModel.TransferDestination{
					ChannelCode: "BCA",
					ChannelInformation: refundModel.ChannelInformation{
						AccountNumber: "1234567890",
						AccountName:   "John Doe",
					},
				},
			},
			expected: "BCA - *7890",
		},
		{
			name: "Unknown destination type",
			refund: &refundModel.RefundResponse{
				DestinationType: "UNKNOWN",
			},
			expected: "-",
		},
		{
			name: "CHANNEL but no channel destination",
			refund: &refundModel.RefundResponse{
				DestinationType: "CHANNEL",
			},
			expected: "-",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := buildRefundDestinationDisplay(tc.refund)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractRRN(t *testing.T) {
	testCases := []struct {
		name     string
		refund   *refundModel.RefundResponse
		expected string
	}{
		{
			name: "RRN present in channel destination",
			refund: &refundModel.RefundResponse{
				ChannelDestination: &refundModel.ChannelDestination{
					PaymentDetail: map[string]interface{}{
						"rrn": "0123456789",
					},
				},
			},
			expected: "0123456789",
		},
		{
			name: "No channel destination",
			refund: &refundModel.RefundResponse{},
			expected: "",
		},
		{
			name: "Channel destination without RRN",
			refund: &refundModel.RefundResponse{
				ChannelDestination: &refundModel.ChannelDestination{
					PaymentDetail: map[string]interface{}{
						"acquirer": "BNC",
					},
				},
			},
			expected: "",
		},
		{
			name: "Empty RRN",
			refund: &refundModel.RefundResponse{
				ChannelDestination: &refundModel.ChannelDestination{
					PaymentDetail: map[string]interface{}{
						"rrn": "",
					},
				},
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractRRN(tc.refund)
			assert.Equal(t, tc.expected, result)
		})
	}
}
