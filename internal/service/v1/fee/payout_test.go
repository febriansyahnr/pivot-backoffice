package feeService_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fee"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/go-redis/redismock/v9"
)

func TestGetPayoutTransactionFee(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	rdb, clientMock := redismock.NewClientMock()
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)

	service := New(logger, nil, merchantRepo, WithRedisClient(redisExt.WrapRedisClient(rdb, nil)))

	payoutFeeRequest := feeModel.GetPayoutTrxFeeRequest{
		MerchantId:   "56ba3e7b-4b54-41d2-8528-ea071638e23b",
		MerchantType: constant.MerchantTypeMerchant,
		BankCode:     "013", // NOSONAR
	}
	payoutFeeCacheKey := fmt.Sprintf(constant.CacheKeyFmtPayoutTransactionFee, payoutFeeRequest.MerchantId, "permata") // NOSONAR

	var (
		nilFeeMetadataObject      *feeModel.FeeMetadataObject
		nilTrxFeeOnBehalfMetadata *feeModel.TrxFeeOnBehalfMetadata
	)

	tests := []struct {
		name       string
		request    feeModel.GetPayoutTrxFeeRequest
		setupMock  func()
		wantErr    error
		wantResult feeModel.FeeResponseder
	}{
		{
			name: "ERROR:Get transaction fee on behalf",
			request: feeModel.GetPayoutTrxFeeRequest{
				MerchantType: constant.MerchantTypeSubMerchant,
			},
			setupMock: func() {
				merchantRepo.On(
					"GetTransactionFeeOnBehalf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantErr:    assert.AnError,
			wantResult: nilTrxFeeOnBehalfMetadata,
		},
		{
			name:    "ERROR:Get fee calculation and detail",
			request: payoutFeeRequest,
			setupMock: func() {
				clientMock.ExpectGet(payoutFeeCacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"DeterminePayoutFeeByMerchantIdAndChannel", mock.Anything, payoutFeeRequest.MerchantId, "PERMATA", constant.ReferenceDisbursement, // NOSONAR
				).Once().Return(nil, assert.AnError)
				logger.On(
					"Error", mock.Anything, "Failed while determine payout fee by merchant and channel", mock.Anything,
				).Once().Return()
			},
			wantErr:    pkgErrs.New(response.HttpErrDatabase, assert.AnError),
			wantResult: nilFeeMetadataObject,
		},
		{
			name: "SUCCESS:Get transaction fee on behalf",
			request: feeModel.GetPayoutTrxFeeRequest{
				MerchantType: constant.MerchantTypeSubMerchant,
			},
			setupMock: func() {
				merchantRepo.On(
					"GetTransactionFeeOnBehalf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(nil, nil)
			},
			wantResult: &feeModel.TrxFeeOnBehalfMetadata{
				Type: constant.FeeOnBehalfTypeNotSet, AmountType: "AMOUNT",
			},
		},
		{
			name:    "SUCCESS:Get fee calculation and detail",
			request: payoutFeeRequest,
			setupMock: func() {
				clientMock.ExpectGet(payoutFeeCacheKey).SetVal("{}")
			},
			wantResult: &feeModel.FeeMetadataObject{Type: constant.ReferenceDisbursement, Channel: "PERMATA"}, // NOSONAR
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientMock.ClearExpect()

			test.setupMock()

			result, err := service.GetPayoutTransactionFee(context.Background(), test.request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
