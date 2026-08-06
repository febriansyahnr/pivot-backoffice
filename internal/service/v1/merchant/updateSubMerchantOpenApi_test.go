package merchant

import (
	"context"
	"errors"
	"testing"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateSubMerchantOpenApi(t *testing.T) {
	ctx := context.Background()

	Name := "test-1"
	Description := "description-1"
	Logo := "logo-1"
	MerchantEmail := "email-1"
	MerchantPhone := "phone-1"

	merchant := merchantModel.Merchant{
		Name:          "existing-name",
		Description:   "existing-description",
		Logo:          "existing-logo",
		MerchantEmail: "existing-email",
		MerchantPhone: "existing-phone",
	}

	request := &merchantModel.UpdateMerchantOpenApiRequest{
		Name:          Name,
		Description:   Description,
		Logo:          Logo,
		MerchantEmail: MerchantEmail,
		MerchantPhone: MerchantPhone,
	}

	tests := []struct {
		name           string
		setupMocks     func(repo *mockRepo.IMerchantRepository)
		expectedError  bool
		expectDBUpdate bool
		request        *merchantModel.UpdateMerchantOpenApiRequest
	}{
		{
			name:           "SUCCESS: Update merchant details",
			expectedError:  false,
			expectDBUpdate: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			request: request,
		},
		{
			name:           "ERROR: Merchant not found",
			expectedError:  true,
			expectDBUpdate: false,
			setupMocks: func(repo *mockRepo.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
			request: request,
		},
		{
			name:           "ERROR: Failed to find merchant",
			expectedError:  true,
			expectDBUpdate: false,
			setupMocks: func(repo *mockRepo.IMerchantRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, errors.New("error finding merchant"))
			},
			request: request,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockRepo.NewIMerchantRepository(t)
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tt.setupMocks(repo)

			s := New(repo, logger, nil, nil, nil, nil)

			_, err := s.UpdateSubMerchantOpenApi(ctx, tt.request)

			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			if tt.expectDBUpdate {
				repo.AssertCalled(t, "Update", mock.Anything, mock.Anything)
			} else {
				repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			}
		})
	}
}
