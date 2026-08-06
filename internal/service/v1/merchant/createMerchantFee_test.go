package merchant

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMerchantServiceCreateMerchantFee(t *testing.T) {

	rdb := redisMock.NewIRedisExt(t)
	rdb.On("Del", mock.Anything, mock.Anything).Return(redis.NewIntResult(0, nil))

	newMerchantFeeRequest := &merchantModel.NewMerchantFeeRequest{
		MerchantID: "merchant-id",
		Amount:     1000,
		Reference:  constant.ReferenceDisbursement,
		Channel:    "PERMATA", // NOSONAR
	}

	merchantPaymentFee := &merchantModel.NewMerchantFeeRequest{
		MerchantID:    "c8186469-c1f0-42f2-a0bb-f259b6c44d22",
		Reference:     constant.ReferencePayment,
		PaymentMethod: constant.ChannelVirtualAccount,
		Channel:       "BSI", // NOSONAR
	}
	paymentMethodRepo := repoMocks.NewIPaymentMethodRepository(t)

	expectedMerchantFee := &merchantModel.MerchantFee{
		UUID:       "merchant-fee-id",
		MerchantID: "merchant-id",
		Amount:     1000,
		Reference:  "DISBURSEMENT",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	expectedMerchant := &merchantModel.Merchant{
		UUID:      "merchant-id",
		Name:      "test",
		Logo:      "https://paper.id/test.jpg",
		MID:       sql.NullString{String: "0000", Valid: true},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	testCases := []struct {
		name             string
		input            *merchantModel.NewMerchantFeeRequest
		expectedMerchant *merchantModel.Merchant
		mocksSetup       func(merchantRepo *repoMocks.IMerchantRepository)
		wantErr          bool
	}{
		{
			name:             "SUCCESS:Successfully create merchant fee",
			input:            newMerchantFeeRequest,
			expectedMerchant: expectedMerchant,
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, constant.StringMockType(),
				).Return(expectedMerchant, nil)

				merchantRepo.On(
					"GetMerchantFeeByRequest", mock.Anything, mock.Anything,
				).Return(nil, nil)

				merchantRepo.On(
					"CreateMerchantFee", mock.Anything, constant.PtrMerchantFeeMockType(),
				).Return(nil)

			},
			wantErr: false,
		},
		{
			name:  "ERROR:Merchant not found",
			input: newMerchantFeeRequest,
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, constant.StringMockType()).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR:Failed to find merchant",
			input: newMerchantFeeRequest,
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On("FindMerchantByID", mock.Anything, constant.StringMockType()).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:  "ERROR:Non KYC sub merchant",
			input: newMerchantFeeRequest,
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, constant.StringMockType(),
				).Return(&merchantModel.Merchant{ParentID: sql.NullString{Valid: true}}, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Set platform transaction fee",
			input: &merchantModel.NewMerchantFeeRequest{
				Reference: constant.ReferencePlatformTransaction,
			},
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, constant.StringMockType(),
				).Return(&merchantModel.Merchant{
					ParentID:  sql.NullString{Valid: true},
					KYCStatus: sql.NullString{String: constant.KYCStatusApproved},
				}, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR:Get payment method by category type and acquirer",
			input: merchantPaymentFee,
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, constant.StringMockType(),
				).Return(expectedMerchant, nil)

				paymentMethodRepo.On(
					"GetPaymentMethodByCategoryTypeAndAcquirer", mock.Anything, constant.ReferencePayment, constant.ChannelVirtualAccount, "BSI", // NOSONAR
				).Once().Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:  "ERROR:Payment method not found",
			input: merchantPaymentFee,
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, constant.StringMockType(),
				).Return(expectedMerchant, nil)

				paymentMethodRepo.On(
					"GetPaymentMethodByCategoryTypeAndAcquirer", mock.Anything, constant.ReferencePayment, constant.ChannelVirtualAccount, "BSI", // NOSONAR
				).Once().Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Reference is not intended for sub-merchant",
			input: &merchantModel.NewMerchantFeeRequest{
				Reference: constant.ReferencePlatformTransfer,
			},
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				subMerchantData := *expectedMerchant

				subMerchantData.ParentID = sql.NullString{Valid: true, String: uuid.NewString()}
				merchantRepo.On("FindMerchantByID", mock.Anything, constant.StringMockType()).Return(&subMerchantData, nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR:Non payment references fill channel attributes",
			input: &merchantModel.NewMerchantFeeRequest{
				Reference: constant.ReferenceDisbursement,
				Channel:   "ABC",
			},
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, constant.StringMockType(),
				).Return(expectedMerchant, nil)
			},
			wantErr: true,
		},
		{
			name:  "ERROR:Merchant fee already existed",
			input: newMerchantFeeRequest,
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, constant.StringMockType(),
				).Return(expectedMerchant, nil)

				merchantRepo.On(
					"GetMerchantFeeByRequest", mock.Anything, mock.Anything,
				).Return(expectedMerchantFee, nil)

			},
			wantErr: true,
		},
		{
			name:  "ERROR:Failed to create merchant fee",
			input: newMerchantFeeRequest,
			mocksSetup: func(merchantRepo *repoMocks.IMerchantRepository) {
				merchantRepo.On(
					"FindMerchantByID", mock.Anything, constant.StringMockType(),
				).Return(expectedMerchant, nil)

				merchantRepo.On(
					"GetMerchantFeeByRequest", mock.Anything, mock.Anything,
				).Return(nil, nil)

				merchantRepo.On(
					"CreateMerchantFee", mock.Anything, mock.Anything,
				).Return(assert.AnError)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repoMocks.NewIMerchantRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mocksSetup(merchantRepo)
			trxSvc := New(merchantRepo, loggerMock, nil, nil, nil, nil, WithPaymentMethodRepository(paymentMethodRepo), WithRedisClient(rdb))

			_, err := trxSvc.CreateMerchantFee(context.Background(), tc.input)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			merchantRepo.AssertExpectations(t)
		})
	}
}
