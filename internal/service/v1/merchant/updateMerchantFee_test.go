package merchant

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	mockJWT "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/repository/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMerchantServiceUpdateMerchantFee(t *testing.T) {
	maxFeeAmount := float64(1000)

	merchantFeeRequest := &merchantModel.UpdateMerchantFeeRequest{
		MerchantID: "merchant-id", // NOSONAR
		Amount:     1000,          // NOSONAR
		AmountType: "AMOUNT",      // NOSONAR
	}

	// payoutMerchantFeeRequest := &merchantModel.UpdateMerchantFeeRequest{
	// 	MerchantID:    "2b23f236-d51b-4900-898b-f1b658ff7213",
	// 	AmountType:    constant.MerchantFeeAmountType,
	// 	DeductionType: constant.MerchantFeeDeductionTypeDirect,
	// }

	merchantFee := &merchantModel.MerchantFee{
		UUID:       "merchant-fee-id", // NOSONAR
		MerchantID: "merchant-id",     // NOSONAR
		Amount:     1000,              // NOSONAR
		AmountType: "AMOUNT",          // NOSONAR
		Reference:  "ACCOUNT_INQUIRY", // NOSONAR
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	payoutMerchantFee := *merchantFee
	payoutMerchantFee.Channel = util.ValueToPtr("PERMATA") // NOSONAR
	payoutMerchantFee.Reference = constant.ReferenceDisbursement

	platformTransferMerchantFee := *merchantFee
	platformTransferMerchantFee.Reference = constant.ReferencePlatformTransfer

	expectedMerchant := &merchantModel.Merchant{
		UUID:      "merchant-id",                               // NOSONAR
		Name:      "test",                                      // NOSONAR
		Logo:      "https://paper.id/test.jpg",                 // NOSONAR
		MID:       sql.NullString{String: "0000", Valid: true}, // NOSONAR
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rdb := redisMock.NewIRedisExt(t)

	testCases := []struct {
		name             string
		input            *merchantModel.UpdateMerchantFeeRequest
		expectedMerchant *merchantModel.Merchant
		expectedError    string
		mocksSetup       func(merchantRepo *mockMerchant.IMerchantRepository)
		wantErr          bool
	}{
		{
			name:             "SUCCESS: successfully update merchant fee",
			input:            merchantFeeRequest,
			expectedMerchant: expectedMerchant,
			mocksSetup: func(merchantRepo *mockMerchant.IMerchantRepository) {
				merchantRepo.On(
					"GetMerchantFeeByID",
					mock.Anything,
					constant.StringMockType(),
				).Return(merchantFee, nil)

				merchantRepo.On(
					"UpdateMerchantFee",
					mock.Anything,
					constant.PtrMerchantFeeMockType(),
				).Return(nil)

				rdb.On("Del", mock.Anything, constant.StringMockType()).Once().Return(&redis.IntCmd{})
			},
			wantErr: false,
		},
		{
			name:          "ERROR: failed to find merchant",
			input:         merchantFeeRequest,
			expectedError: "failed to find merchant",
			mocksSetup: func(merchantRepo *mockMerchant.IMerchantRepository) {
				merchantRepo.On(
					"GetMerchantFeeByID",
					mock.Anything,
					constant.StringMockType(),
				).Return(nil, nil)

			},
			wantErr: true,
		},
		{
			name:  "Error: existed merchant fee not found",
			input: merchantFeeRequest,
			mocksSetup: func(merchantRepo *mockMerchant.IMerchantRepository) {
				merchantRepo.On(
					"GetMerchantFeeByID",
					mock.Anything,
					constant.StringMockType(),
				).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: failed to update merchant fee",
			input: merchantFeeRequest,
			mocksSetup: func(merchantRepo *mockMerchant.IMerchantRepository) {
				merchantRepo.On(
					"GetMerchantFeeByID",
					mock.Anything,
					constant.StringMockType(),
				).Return(merchantFee, nil)

				merchantRepo.On(
					"UpdateMerchantFee",
					mock.Anything,
					constant.PtrMerchantFeeMockType(),
				).Return(assert.AnError)

			},
			wantErr: true,
		},
		{
			name: "ERROR: invalid amount type",
			input: &merchantModel.UpdateMerchantFeeRequest{
				MerchantID:   "merchant-id", // NOSONAR
				ID:           "fee-id",      // NOSONAR
				Amount:       1000,          // NOSONAR
				AmountType:   "INVALID",     // NOSONAR
				MaxFeeAmount: maxFeeAmount,  // NOSONAR
			},
			mocksSetup: func(merchantRepo *mockMerchant.IMerchantRepository) {
				merchantRepo.On(
					"GetMerchantFeeByID",
					mock.Anything,
					constant.StringMockType(),
				).Return(&platformTransferMerchantFee, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR: failed redis del",
			input: merchantFeeRequest,
			mocksSetup: func(merchantRepo *mockMerchant.IMerchantRepository) {
				merchantRepo.On(
					"GetMerchantFeeByID",
					mock.Anything,
					constant.StringMockType(),
				).Return(merchantFee, nil)

				merchantRepo.On(
					"UpdateMerchantFee",
					mock.Anything,
					constant.PtrMerchantFeeMockType(),
				).Return(nil)

				rdb.On("Del", mock.Anything, mock.Anything).Return(redis.NewIntResult(0, assert.AnError))
			},
			wantErr: true,
		},
		{
			name:  "ERROR: failed redis del",
			input: merchantFeeRequest,
			mocksSetup: func(merchantRepo *mockMerchant.IMerchantRepository) {
				merchantRepo.On("GetMerchantFeeByID", mock.Anything, constant.StringMockType()).Return(&payoutMerchantFee, nil)
				merchantRepo.On("UpdateMerchantFee", mock.Anything, constant.PtrMerchantFeeMockType()).Return(nil)

				rdb.On("Del", mock.Anything, mock.Anything).Return(redis.NewIntResult(0, assert.AnError))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := mockMerchant.NewIMerchantRepository(t)
			userRepo := mockUser.NewIUserRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			rabbitMqMock := mockRabbitMq.NewRabbitMQExt(t)
			jwtMock := mockJWT.NewIJwt(t)
			encryptMock := mockEncrypt.NewICrypto(t)
			accountSvc := mocks.NewIAccountService(t)

			tc.mocksSetup(merchantRepo)

			trxSvc := New(merchantRepo, loggerMock, userRepo, jwtMock, rabbitMqMock, encryptMock, WithAccountService(accountSvc), WithRedisClient(rdb))

			if err := trxSvc.UpdateMerchantFee(context.Background(), tc.input); tc.wantErr {
				require.Error(t, err)

			} else {
				assert.NoError(t, err)
			}

			merchantRepo.AssertExpectations(t)
		})
	}
}
