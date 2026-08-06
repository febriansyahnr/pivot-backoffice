package merchant

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindMerchantByID(t *testing.T) {
	expectedMerchant := &merchantModel.Merchant{
		UUID:      "uuid-uuid-uuid",
		Name:      "test",
		Logo:      "https://paper.id/test.jpg",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	testCases := []struct {
		Name           string
		IsSuccess      bool
		MerchantID     string
		ExpectedResult *merchantModel.MerchantResponse
		ExpectedError  string
		MockSetup      func(mockRepo *mockMerchant.IMerchantRepository)
	}{
		{
			Name:           "SUCCESS: find merchant by id",
			IsSuccess:      true,
			MerchantID:     "uuid-uuid-uuid",
			ExpectedResult: expectedMerchant.ToResponse(),
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedMerchant, nil)
			},
		},
		{
			Name:          "ERROR: error find merchant",
			IsSuccess:     false,
			MerchantID:    "merchant-error",
			ExpectedError: "error find merchant",
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("error find merchant"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			merchantRepo := mockMerchant.NewIMerchantRepository(t)
			userRepo := mockUser.NewIUserRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			jwtMock := mockJWT.NewIJwt(t)
			accountSvc := mocks.NewIAccountService(t)

			tc.MockSetup(merchantRepo)
			svc := New(merchantRepo, loggerMock, userRepo, jwtMock, nil, nil, WithAccountService(accountSvc))

			response, err := svc.FindMerchantByID(context.Background(), tc.MerchantID)
			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			} else {
				require.Error(t, err)
				require.Empty(t, response)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			}

			merchantRepo.AssertExpectations(t)
		})
	}
}
