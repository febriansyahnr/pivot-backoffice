package merchant_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	feeService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fee"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateFeeTieringConfig(t *testing.T) {

	rdb, clientMock := redismock.NewClientMock()

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := New(merchantRepo, logger, nil, nil, nil, nil, WithFeeCalculation(&feeService.FeeService{}), WithRedisClient(redisExt.WrapRedisClient(rdb, nil)))

	merchantId := uuid.NewString()
	cacheKey := fmt.Sprintf(
		c.NonPaymentFeeConfigsFmt, merchantId, strings.ToLower(c.ReferenceDisbursement),
	)
	payoutCacheKeyPattern := fmt.Sprintf(
		c.CacheKeyFmtPayoutTransactionFee, merchantId, "*",
	)
	merchantFee := &merchant.MerchantFee{
		MerchantID:    merchantId,
		Reference:     c.ReferenceDisbursement,
		AmountType:    c.MerchantFeeAmountType,
		Amount:        4_000,
		DeductionType: c.MerchantFeeDeductionTypeDirect,
		TaxType:       c.MerchantTaxTypeNonPKP,
		TaxPercentage: 0,
	}
	feeTieringConfigs := []merchant.FeeTieringConfig{
		{Tier: 1, Min: 0, Max: 10_000, AmountType: "AMOUNT", Amount: 5_000},
		{Tier: 2, Min: 10_001, Max: 20_000, AmountType: "AMOUNT", Amount: 4_500},
		{Tier: 3, Min: 20_001, Max: 30_000, AmountType: "AMOUNT", Amount: 4_000},
		{Tier: 4, Min: 30_001, Max: 40_000, AmountType: "AMOUNT", Amount: 3_500},
		{Tier: 5, Min: 40_001, Max: 50_000, AmountType: "AMOUNT", Amount: 3_000}, // The max value will be overwritten with 18446744073709551615
	}

	tests := []struct {
		name       string
		request    *merchant.FeeTieringRequest
		setupMock  func()
		wantErr    error
		wantResult *merchant.FeeTieringResponse
	}{
		{
			name:    "ERROR:Nil fee tiering config",
			request: &merchant.FeeTieringRequest{},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("make sure the data sent is not empty")), // NOSONAR
		},
		{
			name: "ERROR:Get merchant fee by id",
			setupMock: func() {
				merchantRepo.On(
					"GetMerchantFeeByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Merchant fee ID is not found",
			setupMock: func() {
				merchantRepo.On(
					"GetMerchantFeeByID", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrDataNotFound),
		},
		{
			name: "ERROR:Minimum value of tier 1 must be 0",
			request: &merchant.FeeTieringRequest{
				MerchantId: merchantId,
				Configs:    []merchant.FeeTieringConfig{{Tier: 1, Min: 1, Max: 2}},
			},
			setupMock: func() {
				merchantRepo.On(
					"GetMerchantFeeByID", c.ValueCtxMockType(), c.StringMockType(),
				).Return(merchantFee, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("minimum value of tier 1 must be 0")), // NOSONAR
		},
		{
			name: "ERROR:Invalid fee tiering sequence",
			request: &merchant.FeeTieringRequest{
				MerchantId: merchantId,
				Configs:    []merchant.FeeTieringConfig{{}},
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrFeeTieringSequence),
		},
		{
			name: "ERROR:Invalid max fee amount",
			request: &merchant.FeeTieringRequest{
				MerchantId: merchantId,
				Configs: []merchant.FeeTieringConfig{
					{Tier: 1, Min: 0, Max: 10, MaxFeeAmount: util.ValueToPtr(1.00)},
				},
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("not allowed to have maximum fee for requested reference & referenceType")), // NOSONAR
		},
		{
			name: "ERROR:Invalid fee tiering range",
			request: &merchant.FeeTieringRequest{
				MerchantId: merchantId,
				Configs: []merchant.FeeTieringConfig{
					{Tier: 1, Min: 0, Max: 10},
					{Tier: 2, Min: 1, Max: 10},
				},
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, c.ErrFeeTieringRange), // NOSONAR
		},
		{
			name: "ERROR:Fee tiering",
			request: &merchant.FeeTieringRequest{
				MerchantId: merchantId,
				Configs: []merchant.FeeTieringConfig{
					{Tier: 1, Min: 0, Max: 10, AmountType: "AMOUNT", Amount: 4_000},
					{Tier: 2, Min: 11, Max: 99, AmountType: "AMOUNT", Amount: 4_000}, // The max value will be overwritten with 18446744073709551615
				},
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("fee amount tier 2 is greater than or equal tier 1")),
		},
		{
			name: "ERROR:Tier to be applied was not found",
			request: &merchant.FeeTieringRequest{
				MerchantId:  merchantId,
				AppliedTier: 3,
				Configs: []merchant.FeeTieringConfig{
					{Tier: 1, Min: 0, Max: 10, AmountType: "AMOUNT", Amount: 4_000},
					{Tier: 2, Min: 11, Max: 99, AmountType: "AMOUNT", Amount: 3_000}, // The max value will be overwritten with 18446744073709551615
				},
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("tier to be applied was not found")),
		},
		{
			name: "ERROR:Update fee tiering config",
			request: &merchant.FeeTieringRequest{
				MerchantId: merchantId,
				Configs:    feeTieringConfigs,
			},
			setupMock: func() {
				merchantRepo.On(
					"UpdateFeeTieringConfig", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.FeeTieringRequest"),
				).Once().Return(c.ErrSomeErrorForUnitTest)

			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Clear cache",
			request: &merchant.FeeTieringRequest{
				MerchantId:  merchantId,
				AppliedTier: 3,
				Configs:     feeTieringConfigs,
			},
			setupMock: func() {
				merchantRepo.On(
					"UpdateFeeTieringConfig", c.ValueCtxMockType(), mock.AnythingOfType("*merchant.FeeTieringRequest"),
				).Return(nil)

				clientMock.ClearExpect()
				clientMock.ExpectKeys(payoutCacheKeyPattern).SetVal([]string{})
				clientMock.ExpectDel(cacheKey).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:LADDER with AppliedTier",
			request: &merchant.FeeTieringRequest{
				MerchantId:  merchantId,
				Model:       c.LadderTieringModel,
				AppliedTier: 2,
				Configs: []merchant.FeeTieringConfig{
					{Tier: 1, Min: 0, Max: 10, AmountType: "AMOUNT", Amount: 4_000},
					{Tier: 2, Min: 11, Max: 99, AmountType: "AMOUNT", Amount: 3_000},
				},
			},
			wantErr: pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("appliedTier is not supported for LADDER tiering model")),
		},
		{
			name: "SUCCESS:MONTHLY_ASSESSED (default)",
			request: &merchant.FeeTieringRequest{
				MerchantId: merchantId,
				Configs:    feeTieringConfigs,
			},
			setupMock: func() {
				clientMock.ClearExpect()
				clientMock.ExpectKeys(payoutCacheKeyPattern).SetVal([]string{})
				clientMock.ExpectDel(cacheKey).SetVal(1)
			},
			wantResult: &merchant.FeeTieringResponse{
				MerchantId:    merchantFee.MerchantID,
				Reference:     merchantFee.Reference,
				PaymentMethod: merchantFee.PaymentMethod,
				DeductionType: merchantFee.DeductionType,
				Model:         c.MonthlyAssessedTieringModel,
				Configs:       feeTieringConfigs,
			},
		},
		{
			name: "SUCCESS:LADDER model",
			request: &merchant.FeeTieringRequest{
				MerchantId: merchantId,
				Model:      c.LadderTieringModel,
				Configs:    feeTieringConfigs,
			},
			setupMock: func() {
				clientMock.ClearExpect()
				clientMock.ExpectKeys(payoutCacheKeyPattern).SetVal([]string{})
				clientMock.ExpectDel(cacheKey).SetVal(1)
			},
			wantResult: &merchant.FeeTieringResponse{
				MerchantId:    merchantFee.MerchantID,
				Reference:     merchantFee.Reference,
				PaymentMethod: merchantFee.PaymentMethod,
				DeductionType: merchantFee.DeductionType,
				Model:         c.LadderTieringModel,
				Configs:       feeTieringConfigs,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.request == nil {
				test.request = &merchant.FeeTieringRequest{
					Configs: []merchant.FeeTieringConfig{{Tier: 1, Min: 0, Max: 1}},
				}
			}
			if test.setupMock != nil {
				test.setupMock()
			}

			result, err := service.UpdateFeeTieringConfig(context.Background(), test.request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
