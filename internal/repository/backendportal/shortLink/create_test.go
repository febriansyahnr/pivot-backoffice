package shortLinkRepository

import (
	"context"
	"errors"
	"testing"
	"time"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	shortLinkModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/shortLink"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShortLinkRepository_Create(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		shortLink *shortLinkModel.ShortLink
		mockSetup func(*mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "success",
			shortLink: &shortLinkModel.ShortLink{
				UUID:           "test-uuid",
				Reference:      "payment-ref-123",
				Code:           "ABC123",
				DestinationURL: "https://example.com/payment/123",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(24 * time.Hour),
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "database error",
			shortLink: &shortLinkModel.ShortLink{
				UUID:           "test-uuid",
				Reference:      "payment-ref-123",
				Code:           "ABC123",
				DestinationURL: "https://example.com/payment/123",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(24 * time.Hour),
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				dbErr := errors.New("duplicate key error")
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, dbErr)
			},
			wantErr: true,
		},
		{
			name: "empty uuid",
			shortLink: &shortLinkModel.ShortLink{
				UUID:           "",
				Reference:      "payment-ref-123",
				Code:           "ABC123",
				DestinationURL: "https://example.com/payment/123",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(24 * time.Hour),
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything).Return(false, nil)
			},
			wantErr: false,
		},
		{
			name: "minimal required fields",
			shortLink: &shortLinkModel.ShortLink{
				UUID:           "minimal-uuid",
				Reference:      "ref",
				Code:           "MIN",
				DestinationURL: "https://test.com",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(time.Hour),
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, nil)
			},
			wantErr: false,
		},
		{
			name: "long destination url",
			shortLink: &shortLinkModel.ShortLink{
				UUID:           "long-uuid",
				Reference:      "long-ref",
				Code:           "LONG123",
				DestinationURL: "https://example.com/very/long/payment/url/with/many/parameters?param1=value1&param2=value2&param3=value3&sessionId=abcd1234&token=xyz789&redirect=https://callback.com/success",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(72 * time.Hour),
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tt.mockSetup(mockDB)

			repo := &shortLinkRepo{
				db:  mockDB,
				log: mockLogger,
			}

			err := repo.Create(context.Background(), tt.shortLink)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
