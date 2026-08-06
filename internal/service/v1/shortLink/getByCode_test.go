package shortLink

import (
	"context"
	"errors"
	"testing"
	"time"

	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/shortLink"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortLinkService_GetByCode(t *testing.T) {
	now := time.Now().UTC()
	validShortLink := &shortLinkModel.ShortLink{
		UUID:           "test-uuid",
		Reference:      "payment-ref-123",
		Code:           "ABC123",
		DestinationURL: "https://example.com/payment/123",
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiredAt:      now.Add(24 * time.Hour), // Future expiration
	}

	expiredShortLink := &shortLinkModel.ShortLink{
		UUID:           "expired-uuid",
		Reference:      "payment-ref-expired",
		Code:           "EXPIRED",
		DestinationURL: "https://example.com/payment/expired",
		CreatedAt:      now.Add(-48 * time.Hour),
		UpdatedAt:      now.Add(-48 * time.Hour),
		ExpiredAt:      now.Add(-24 * time.Hour), // Past expiration
	}

	testCases := []struct {
		name       string
		code       string
		mockSetup  func(*repoMocks.IShortLinkRepository)
		wantResult *shortLinkModel.ShortLink
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "SUCCESS: valid code with future expiration",
			code: "ABC123",
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
				mockRepo.On(
					"GetByCode",
					mock.Anything,
					"ABC123",
				).Return(validShortLink, nil)
			},
			wantResult: validShortLink,
			wantErr:    false,
		},
		{
			name: "ERROR: repository error",
			code: "ERROR123",
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
				mockRepo.On(
					"GetByCode",
					mock.Anything,
					"ERROR123",
				).Return(nil, errors.New("database connection failed"))
			},
			wantResult: nil,
			wantErr:    true,
			wantErrMsg: "error when retrieve link",
		},
		{
			name: "ERROR: short link not found (nil result)",
			code: "NOTFOUND",
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
				mockRepo.On(
					"GetByCode",
					mock.Anything,
					"NOTFOUND",
				).Return(nil, nil)
			},
			wantResult: nil,
			wantErr:    true,
			wantErrMsg: "link not found",
		},
		{
			name: "ERROR: short link expired",
			code: "EXPIRED",
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
				mockRepo.On(
					"GetByCode",
					mock.Anything,
					"EXPIRED",
				).Return(expiredShortLink, nil)
			},
			wantResult: nil,
			wantErr:    true,
			wantErrMsg: "link expired",
		},
		{
			name: "SUCCESS: empty code handled by repository",
			code: "",
			mockSetup: func(mockRepo *repoMocks.IShortLinkRepository) {
				mockRepo.On(
					"GetByCode",
					mock.Anything,
					"",
				).Return(nil, nil)
			},
			wantResult: nil,
			wantErr:    true,
			wantErrMsg: "link not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repoMocks.NewIShortLinkRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockRepo)

			service := NewShortLinkService(mockLogger, mockRepo)
			ctx := context.Background()

			result, err := service.GetByCode(ctx, tc.code)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.wantResult.UUID, result.UUID)
				assert.Equal(t, tc.wantResult.Reference, result.Reference)
				assert.Equal(t, tc.wantResult.Code, result.Code)
				assert.Equal(t, tc.wantResult.DestinationURL, result.DestinationURL)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}