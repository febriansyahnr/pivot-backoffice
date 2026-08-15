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

func TestShortLinkRepository_Update(t *testing.T) {
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
				Reference:      "updated-ref-123",
				Code:           "UPD123",
				DestinationURL: "https://example.com/updated/payment/123",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(48 * time.Hour),
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
				Reference:      "updated-ref-123",
				Code:           "UPD123",
				DestinationURL: "https://example.com/updated/payment/123",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(48 * time.Hour),
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				dbErr := errors.New("foreign key constraint error")
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
			name: "record not found (no rows affected)",
			shortLink: &shortLinkModel.ShortLink{
				UUID:           "nonexistent-uuid",
				Reference:      "updated-ref-123",
				Code:           "UPD123",
				DestinationURL: "https://example.com/updated/payment/123",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(48 * time.Hour),
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
			name: "update with extended expiration",
			shortLink: &shortLinkModel.ShortLink{
				UUID:           "extend-uuid",
				Reference:      "extend-ref",
				Code:           "EXT123",
				DestinationURL: "https://example.com/extended",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(7 * 24 * time.Hour), // 7 days
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
			name: "update destination url only",
			shortLink: &shortLinkModel.ShortLink{
				UUID:           "url-update-uuid",
				Reference:      "same-ref",
				Code:           "SAME123",
				DestinationURL: "https://newdomain.com/payment/redirect",
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
				).Return(false, nil)
			},
			wantErr: false,
		},
		{
			name: "connection timeout error",
			shortLink: &shortLinkModel.ShortLink{
				UUID:           "timeout-uuid",
				Reference:      "timeout-ref",
				Code:           "TIMEOUT",
				DestinationURL: "https://example.com/timeout",
				CreatedAt:      now,
				UpdatedAt:      now,
				ExpiredAt:      now.Add(24 * time.Hour),
			},
			mockSetup: func(mockDB *mysqlMocks.IMySqlExt) {
				dbErr := errors.New("connection timeout")
				mockDB.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, dbErr)
			},
			wantErr: true,
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

			err := repo.Update(context.Background(), tt.shortLink)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
