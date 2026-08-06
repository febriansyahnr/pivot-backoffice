package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateSubMerchantParent(t *testing.T) {
	merchant := &merchant.Merchant{
		ParentID: sql.NullString{
			String: uuid.Max.String(),
			Valid:  true,
		},
	}
	parentMerchantID := uuid.Max.String()

	tests := []struct {
		name             string
		setupMocks       func(repo *mockRepo.IMerchantRepository)
		expectedError    bool
		parentMerchantID string
	}{
		{
			name:          "SUCCESS: validate sub merchant parent",
			expectedError: false,
			setupMocks: func(repo *mockRepo.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(merchant, nil)
			},
			parentMerchantID: parentMerchantID,
		},
		{
			name:          "ERROR: parent not matched",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(merchant, nil)
			},
			parentMerchantID: uuid.NewString(),
		},
		{
			name:          "ERROR: submerchant not found",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
			parentMerchantID: uuid.NewString(),
		},
		{
			name:          "ERROR: error find submerchant",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
			parentMerchantID: uuid.NewString(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockRepo.NewIMerchantRepository(t)
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tt.setupMocks(repo)

			s := New(repo, logger, nil, nil, nil, nil)
			err := s.ValidateSubMerchantParent(context.Background(), tt.parentMerchantID, uuid.NewString())

			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

		})
	}

}
