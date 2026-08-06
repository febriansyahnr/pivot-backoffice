package shortLink

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortLinkService_Create(t *testing.T) {
	now := time.Now().UTC()
	expiredAt := now.Add(24 * time.Hour)

	testCases := []struct {
		name       string
		request    *shortLinkModel.CreateShortLink
		mockSetup  func(*repoMocks.IShortLinkRepository)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "SUCCESS: create short link",
			request: &shortLinkModel.CreateShortLink{
				Reference:      "payment-ref-123",
				DestinationURL: "https://example.com/payment/123",
				UniqueID:       "unique-123",
				ExpiredAt:      expiredAt,
			},
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
				mockRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*shortLinkModel.ShortLink"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: create short link without unique ID",
			request: &shortLinkModel.CreateShortLink{
				Reference:      "payment-ref-456",
				DestinationURL: "https://example.com/payment/456",
				UniqueID:       "",
				ExpiredAt:      expiredAt,
			},
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
				mockRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*shortLinkModel.ShortLink"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: invalid destination URL",
			request: &shortLinkModel.CreateShortLink{
				Reference:      "payment-ref-789",
				DestinationURL: "",
				UniqueID:       "unique-789",
				ExpiredAt:      expiredAt,
			},
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
			},
			wantErr:    true,
			wantErrMsg: constant.ErrShortLinkDestinationRequired.Error(),
		},
		{
			name: "ERROR: repository create failed",
			request: &shortLinkModel.CreateShortLink{
				Reference:      "payment-ref-789",
				DestinationURL: "https://example.com/payment/789",
				UniqueID:       "unique-789",
				ExpiredAt:      expiredAt,
			},
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
				mockRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*shortLinkModel.ShortLink"),
				).Return(errors.New("database connection failed"))
			},
			wantErr:    true,
			wantErrMsg: constant.ErrCreateShortLink.Error(),
		},
		{
			name: "ERROR: duplicate key constraint",
			request: &shortLinkModel.CreateShortLink{
				Reference:      "payment-ref-duplicate",
				DestinationURL: "https://example.com/payment/duplicate",
				UniqueID:       "duplicate-unique-id",
				ExpiredAt:      expiredAt,
			},
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
				mockRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*shortLinkModel.ShortLink"),
				).Return(errors.New("duplicate key error"))
			},
			wantErr:    true,
			wantErrMsg: constant.ErrCreateShortLink.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewIShortLinkRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			config := config.Config{
				ShortLinkRedirection: config.ShortLinkRedirectionConfig{
					URLFormat: "http://host/s/%s",
				},
			}

			tc.mockSetup(mockRepo)

			service := NewShortLinkService(mockLogger, mockRepo)
			WithConfig(service, &config)
			ctx := context.Background()

			result, err := service.Create(ctx, tc.request)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.request.Reference, result.Reference)
				assert.Equal(t, tc.request.DestinationURL, result.DestinationURL)
				assert.NotEmpty(t, result.UUID)
				assert.NotEmpty(t, result.Code)
				assert.Equal(t, tc.request.ExpiredAt, result.ExpiredAt)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
