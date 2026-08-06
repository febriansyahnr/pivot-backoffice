package merchant

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetNotificationConfig(t *testing.T) {
	mockRepo := repositoryMocks.NewIMerchantRepository(t)
	svc := &MerchantService{
		repo: mockRepo,
	}

	merchantID := "merchant-123"
	expectedConfig := &merchant.MerchantNotificationConfig{
		Transaction: &merchant.MerchantNotificationTransactionConfig{Active: true},
	}
	metadata := &merchant.MerchantMetadata{NotificationConfig: expectedConfig}
	metadataBytes, _ := json.Marshal(metadata)

	tests := []struct {
		name      string
		setupMock func()
		want      *merchant.MerchantNotificationConfig
		wantErr   error
	}{
		{
			name: "Success",
			setupMock: func() {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID:     merchantID,
					Metadata: types.NullJSONText{Valid: true, JSONText: metadataBytes},
				}, nil).Once()
			},
			want:    expectedConfig,
			wantErr: nil,
		},
		{
			name: "Merchant Not Found",
			setupMock: func() {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, nil).Once()
			},
			want:    nil,
			wantErr: constant.ErrMerchantNotFound,
		},
		{
			name: "Repository Error",
			setupMock: func() {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, errors.New("db error")).Once()
			},
			want:    nil,
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			got, err := svc.GetNotificationConfig(context.Background(), merchantID)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUpdateNotificationConfig(t *testing.T) {
	mockRepo := repositoryMocks.NewIMerchantRepository(t)
	svc := &MerchantService{
		repo: mockRepo,
	}

	merchantID := "merchant-123"
	req := &merchant.MerchantNotificationConfig{
		Transaction: &merchant.MerchantNotificationTransactionConfig{Active: true},
	}

	tests := []struct {
		name      string
		setupMock func()
		want      *merchant.MerchantNotificationConfig
		wantErr   error
	}{
		{
			name: "Success",
			setupMock: func() {
				m := &merchant.Merchant{UUID: merchantID}
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(m, nil).Once()
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(arg *merchant.Merchant) bool {
					meta, _ := arg.GetMetadata()
					return meta.NotificationConfig.Transaction.Active == true
				})).Return(nil).Once()

				// Mock for GetNotificationConfig call
				expectedConfig := &merchant.MerchantNotificationConfig{
					Transaction: &merchant.MerchantNotificationTransactionConfig{Active: true},
				}
				metadata := &merchant.MerchantMetadata{NotificationConfig: expectedConfig}
				metadataBytes, _ := json.Marshal(metadata)
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(&merchant.Merchant{
					UUID:     merchantID,
					Metadata: types.NullJSONText{Valid: true, JSONText: metadataBytes},
				}, nil).Once()
			},
			want:    req,
			wantErr: nil,
		},
		{
			name: "Merchant Not Found",
			setupMock: func() {
				mockRepo.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, nil).Once()
			},
			want:    nil,
			wantErr: constant.ErrMerchantNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			got, err := svc.UpdateNotificationConfig(context.Background(), merchantID, req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
