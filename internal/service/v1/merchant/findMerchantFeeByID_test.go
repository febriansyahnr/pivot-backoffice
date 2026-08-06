package merchant

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
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

func TestMerchantServiceFindMerchantFeeByID(t *testing.T) {
	expectedMerchantFee := &merchantModel.MerchantFee{
		UUID:       "merchant-fee-id",
		MerchantID: "merchant-id",
		Amount:     1000,
		Reference:  "DISBURSEMENT",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	testCases := []struct {
		Name           string
		MerchantFeeID  string
		ExpectedResult *merchantModel.MerchantFee
		MockSetup      func(mockRepo *mockMerchant.IMerchantRepository)
		WantErr        bool
	}{
		{
			Name:           "SUCCESS: find merchant by id",
			MerchantFeeID:  expectedMerchantFee.UUID,
			ExpectedResult: expectedMerchantFee,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"GetMerchantFeeByID",
					mock.Anything,
					constant.StringMockType(),
				).Return(expectedMerchantFee, nil)
			},
			WantErr: false,
		},
		{
			Name:          "ERROR: error find merchant",
			MerchantFeeID: expectedMerchantFee.UUID,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"GetMerchantFeeByID",
					mock.Anything,
					constant.StringMockType(),
				).Return(nil, errors.New("error when finding merchant by id"))

			},
			WantErr: true,
		},
		{
			Name:          "ERROR: merchant not found",
			MerchantFeeID: expectedMerchantFee.UUID,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"GetMerchantFeeByID",
					mock.Anything,
					constant.StringMockType(),
				).Return(nil, nil)
			},
			WantErr: true,
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

			response, err := svc.FindMerchantFeeByID(context.Background(), tc.MerchantFeeID)
			if tc.WantErr {
				require.Error(t, err)
				require.Empty(t, response)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			}

			merchantRepo.AssertExpectations(t)
		})
	}
}
