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

func TestMerchantServiceFindMerchantFeeByMerchantIDAndType(t *testing.T) {
	expectedMerchantFee := &merchantModel.MerchantFee{
		UUID:       "merchant-fee-id",
		MerchantID: "merchant-id",
		Amount:     1000,
		Reference:  constant.TypeDisbursement,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	testCases := []struct {
		Name           string
		MerchantID     string
		FeeType        string
		ExpectedResult *merchantModel.MerchantFee
		MockSetup      func(mockRepo *mockMerchant.IMerchantRepository)
		WantErr        bool
	}{
		{
			Name:           "SUCCESS: find merchant by merchant id and type",
			MerchantID:     expectedMerchantFee.MerchantID,
			FeeType:        constant.TypeDisbursement,
			ExpectedResult: expectedMerchantFee,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"GetMerchantFeeByMerchantIDAndType",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(expectedMerchantFee, nil)
			},
			WantErr: false,
		},
		{
			Name:       "ERROR: error find merchant fee",
			MerchantID: expectedMerchantFee.MerchantID,
			FeeType:    constant.TypeDisbursement,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"GetMerchantFeeByMerchantIDAndType",
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil, errors.New("error when finding merchant by merchant id and reference"))

			},
			WantErr: true,
		},
		{
			Name:       "ERROR: merchant fee not found",
			MerchantID: expectedMerchantFee.MerchantID,
			FeeType:    constant.TypeDisbursement,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"GetMerchantFeeByMerchantIDAndType",
					mock.Anything,
					constant.StringMockType(),
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

			response, err := svc.FindMerchantFeeByMerchantIDAndType(context.Background(), tc.MerchantID, tc.FeeType)
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
