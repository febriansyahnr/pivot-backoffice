package disbursementService

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/mock"
)

func TestIsExistReferenceID(t *testing.T) {
	conf := config.Config{
		Environment: constant.EnvironmentStaging,
	}
	testCases := []struct {
		Name      string
		WantErr   bool
		MockSetup func(mockRepo *repositoryMocks.IDisbursementRepository)
	}{
		{
			Name:    "SUCCESS",
			WantErr: false,
			MockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.On(
					"CountByMerchantAndReference",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(1)
			},
		},
		{
			Name:    "SUCCESS: Is not exist",
			WantErr: false,
			MockSetup: func(mockRepo *repositoryMocks.IDisbursementRepository) {
				mockRepo.On(
					"CountByMerchantAndReference",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(0)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			mockRepo := repositoryMocks.NewIDisbursementRepository(t)
			ctx := context.Background()
			tc.MockSetup(mockRepo)

			svc := New(&conf, pdkLoggerMock, merchantRepo, mockRepo, nil, nil)
			_ = svc.IsExistReferenceID(ctx, uuid.NewString(), "reference")
			mockRepo.AssertExpectations(t)
		})
	}
}
