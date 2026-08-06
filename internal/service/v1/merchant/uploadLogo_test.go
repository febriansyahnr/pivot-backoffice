package merchant

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	mockGcs "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestMerchantService_UploadMerchantLogo(t *testing.T) {
	redisClient, redisClientMock := redismock.NewClientMock()
	merchantID := "47d69568-c811-4e3c-84da-d64af033df45"

	testCases := []struct {
		name      string
		file      *multipart.FileHeader
		mockSetup func(
			mockRepo *mocks.IMerchantRepository,
			mockGcs *mockGcs.GCSService,
			redisMock redismock.ClientMock,
		)
		wantErr bool
		errMsg  string
	}{
		{
			name: "SUCCESS: upload logo successfully",
			file: &multipart.FileHeader{
				Filename: "logo.png",
				Size:     1024 * 1024, // 1MB
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockGcs *mockGcs.GCSService,
				redisMock redismock.ClientMock,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID:      merchantID,
					Name:      "Test Merchant",
					Logo:      "https://old-logo.com/logo.png",
					UpdatedAt: time.Now(),
				}, nil)

				mockGcs.On("UploadFileFromMultipartToBucket", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*multipart.FileHeader"), true).
					Return(&gcs.UploadMultipart{
						PublicURL:  "https://storage.googleapis.com",
						Bucket:     "merchant-logos",
						ObjectName: merchantID + "_1234567890.png",
					}, nil)

				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(nil)

				redisMock.ExpectDel("key-1").SetVal(1)
			},
			wantErr: false,
		},
		{
			name: "ERROR: merchant not found",
			file: &multipart.FileHeader{
				Filename: "logo.png",
				Size:     1024 * 1024,
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockGcs *mockGcs.GCSService,
				redisMock redismock.ClientMock,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, nil)
			},
			wantErr: true,
			errMsg:  constant.ErrMerchantNotFound.Error(),
		},
		{
			name: "ERROR: database error when finding merchant",
			file: &multipart.FileHeader{
				Filename: "logo.png",
				Size:     1024 * 1024,
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockGcs *mockGcs.GCSService,
				redisMock redismock.ClientMock,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: file is nil",
			file: nil,
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockGcs *mockGcs.GCSService,
				redisMock redismock.ClientMock,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: file size exceeds limit",
			file: &multipart.FileHeader{
				Filename: "large-logo.png",
				Size:     10 * 1024 * 1024, // 10MB - exceeds 5MB limit
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockGcs *mockGcs.GCSService,
				redisMock redismock.ClientMock,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: invalid file type",
			file: &multipart.FileHeader{
				Filename: "logo.pdf",
				Size:     1024 * 1024,
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockGcs *mockGcs.GCSService,
				redisMock redismock.ClientMock,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: GCS upload fails",
			file: &multipart.FileHeader{
				Filename: "logo.png",
				Size:     1024 * 1024,
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockGcs *mockGcs.GCSService,
				redisMock redismock.ClientMock,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)

				mockGcs.On("UploadFileFromMultipartToBucket", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*multipart.FileHeader"), true).
					Return(nil, errors.New("GCS upload error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: database update fails",
			file: &multipart.FileHeader{
				Filename: "logo.jpg",
				Size:     1024 * 1024,
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockGcs *mockGcs.GCSService,
				redisMock redismock.ClientMock,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)

				mockGcs.On("UploadFileFromMultipartToBucket", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*multipart.FileHeader"), true).
					Return(&gcs.UploadMultipart{
						PublicURL:  "https://storage.googleapis.com",
						Bucket:     "merchant-logos",
						ObjectName: merchantID + "_1234567890.jpg",
					}, nil)

				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).
					Return(errors.New("database update error"))
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: upload JPEG file",
			file: &multipart.FileHeader{
				Filename: "logo.jpeg",
				Size:     2 * 1024 * 1024, // 2MB
			},
			mockSetup: func(
				mockRepo *mocks.IMerchantRepository,
				mockGcs *mockGcs.GCSService,
				redisMock redismock.ClientMock,
			) {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)

				mockGcs.On("UploadFileFromMultipartToBucket", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*multipart.FileHeader"), true).
					Return(&gcs.UploadMultipart{
						PublicURL:  "https://storage.googleapis.com",
						Bucket:     "merchant-logos",
						ObjectName: merchantID + "_1234567890.jpeg",
					}, nil)

				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*merchant.Merchant")).Return(nil)

				redisMock.ExpectDel("key-1").SetVal(1)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewIMerchantRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRabbitMq := rabbitMqMocks.NewRabbitMQExt(t)
			mockCrypto := mockEncrypt.NewICrypto(t)
			mockGcsService := mockGcs.NewGCSService(t)

			tc.mockSetup(mockRepo, mockGcsService, redisClientMock)

			svc := New(mockRepo, mockLogger, nil, nil, mockRabbitMq, mockCrypto,
				WithGCSService(mockGcsService),
				WithRedisClient(redisExt.WrapRedisClient(redisClient, nil)),
			)

			logoURL, err := svc.UploadMerchantLogo(context.Background(), merchantID, tc.file)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
				assert.Empty(t, logoURL)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, logoURL)
				assert.Contains(t, logoURL, "storage.googleapis.com")
				assert.Contains(t, logoURL, "merchant-logos")
			}

			mockRepo.AssertExpectations(t)
			mockGcsService.AssertExpectations(t)
		})
	}
}

func TestMerchantService_validateLogoFile(t *testing.T) {
	mockRepo := mocks.NewIMerchantRepository(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockRabbitMq := rabbitMqMocks.NewRabbitMQExt(t)
	mockCrypto := mockEncrypt.NewICrypto(t)

	svc := New(mockRepo, mockLogger, nil, nil, mockRabbitMq, mockCrypto).(*MerchantService)

	testCases := []struct {
		name    string
		file    *multipart.FileHeader
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ERROR: nil file",
			file:    nil,
			wantErr: true,
			errMsg:  "file is required",
		},
		{
			name: "ERROR: file size exceeds limit",
			file: &multipart.FileHeader{
				Filename: "large.png",
				Size:     6 * 1024 * 1024, // 6MB
			},
			wantErr: true,
			errMsg:  "file size exceeds maximum limit",
		},
		{
			name: "ERROR: invalid extension - PDF",
			file: &multipart.FileHeader{
				Filename: "document.pdf",
				Size:     1024,
			},
			wantErr: true,
			errMsg:  "invalid file type",
		},
		{
			name: "ERROR: invalid extension - SVG",
			file: &multipart.FileHeader{
				Filename: "logo.svg",
				Size:     1024,
			},
			wantErr: true,
			errMsg:  "invalid file type",
		},
		{
			name: "SUCCESS: valid PNG file",
			file: &multipart.FileHeader{
				Filename: "logo.png",
				Size:     1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: valid JPG file",
			file: &multipart.FileHeader{
				Filename: "logo.jpg",
				Size:     2 * 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: valid JPEG file",
			file: &multipart.FileHeader{
				Filename: "logo.jpeg",
				Size:     3 * 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: uppercase extension",
			file: &multipart.FileHeader{
				Filename: "LOGO.PNG",
				Size:     1024 * 1024,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.validateLogoFile(tc.file)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
