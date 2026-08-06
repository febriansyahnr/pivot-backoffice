package feeService

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetFeeCalculationAndDetail(t *testing.T) {
	rdb, clientMock := redismock.NewClientMock()

	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	feeRepo := repositoryMocks.NewIFeeRepository(t)
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)

	merchantId := uuid.NewString()
	cacheKey := fmt.Sprintf(
		c.NonPaymentFeeConfigsFmt, merchantId, strings.ToLower(c.ReferenceAccountInquiry),
	)
	payoutTrxFeeCacheKey := fmt.Sprintf(c.CacheKeyFmtPayoutTransactionFee, merchantId, "bri")
	payoutFeeRequest := &feeModel.GetFeeRequest{
		MerchantID: merchantId,
		Reference:  c.ReferenceDisbursement,
		Channel:    "BRI", // NOSONAR
	}

	tests := []struct {
		name      string
		request   *feeModel.GetFeeRequest
		setupMock func()
		wantErr   string
	}{
		{
			name: "SUCCESS:Merchant fee config in cache",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetVal(`{}`)
			},
		},
		{
			name:    "SUCCESS:Merchant payout transaction fee in cache",
			request: payoutFeeRequest,
			setupMock: func() {
				clientMock.ExpectGet(payoutTrxFeeCacheKey).SetVal(`{}`)
			},
		},
		{
			name: "SUCCESS:Wallet merchant fee config",
			request: &feeModel.GetFeeRequest{
				MerchantID: merchantId,
				Reference:  c.TypeWallet,
			},
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByRequest", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)
			},
		},
		{
			name: "ERROR:Getting merchant config from cache",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:GetMerchantFeeByMerchantIDAndType error",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:Determine payment fee error",
			request: &feeModel.GetFeeRequest{
				Reference: c.TypePayment,
			},
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name: "ERROR:Fee amount minus",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					AmountType:    c.MerchantFeeAmountType,
					Amount:        -1_000,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeNonPKP,
				}, nil)
			},
			wantErr: pkgErrs.New(response.HttpErrRequest, errors.New("fee amount cannot be negative")).Error(),
		},
		{
			name:    "ERROR:Determine payout transaction fee",
			request: payoutFeeRequest,
			setupMock: func() {
				clientMock.ExpectGet(payoutTrxFeeCacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"DeterminePayoutFeeByMerchantIdAndChannel", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, assert.AnError)
			},
			wantErr: pkgErrs.New(response.HttpErrDatabase, assert.AnError).Error(),
		},
		{
			name:    "SUCCESS:Determine payout transaction fee",
			request: payoutFeeRequest,
			setupMock: func() {
				clientMock.ExpectGet(payoutTrxFeeCacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"DeterminePayoutFeeByMerchantIdAndChannel", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{}, nil)
				clientMock.ExpectSet(payoutTrxFeeCacheKey, &merchantModel.MerchantFee{}, 15*time.Minute).SetVal("true") // NOSONAR
			},
		},
		{
			name: "SUCCESS:Type amount and inclusive tax",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        1000,
					AmountType:    c.MerchantFeeAmountType,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeInclusive,
					TaxPercentage: 10,
				}, nil)
			},
		},
		{
			name: "SUCCESS:Type amount and exclusive tax",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        1000,
					AmountType:    c.MerchantFeeAmountType,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeExclusive,
					TaxPercentage: 10,
				}, nil)
			},
		},
		{
			name: "SUCCESS:Type percentage and inclusive tax",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        0,
					AmountType:    c.MerchantFeePercentageType,
					Percentage:    10,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeInclusive,
					TaxPercentage: 10,
				}, nil)
			},
		},
		{
			name: "SUCCESS:Type percentage and inclusive tax",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        0,
					AmountType:    c.MerchantFeePercentageType,
					Percentage:    10,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeExclusive,
					TaxPercentage: 10,
				}, nil)
			},
		},
		{
			name: "SUCCESS:Type amount_percentage and inclusive tax",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        1000,
					AmountType:    c.MerchantFeeAmountPercentageType,
					Percentage:    10,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeInclusive,
					TaxPercentage: 10,
				}, nil)
			},
		},
		{
			name: "SUCCESS:Type amount_percentage and inclusive tax",
			setupMock: func() {
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:            1000,
					AmountType:        c.MerchantFeeAmountPercentageType,
					Percentage:        10,
					MaxFeeAmount:      util.ValueToPtr(5_000.00),
					DeductionType:     c.MerchantFeeDeductionTypeAutomated,
					DeductionDay:      util.ValueToPtr(int16(21)),
					DeductionLastDate: util.ValueToPtr(time.Now().UTC().Add(30 * (24 * time.Hour))),
					TaxType:           c.MerchantTaxTypeExclusive,
					TaxPercentage:     10,
				}, nil)
			},
		},
		{
			name: "SUCCESS:Default platform transaction amount_percentage and no tax",
			setupMock: func() {
				cacheKey := fmt.Sprintf(
					c.NonPaymentFeeConfigsFmt, merchantId, strings.ToLower(c.ReferencePlatformTransaction),
				)

				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On(
					"GetMerchantFeeByMerchantIDAndType", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID: merchantId,
				Reference:  c.ReferencePlatformTransaction,
			},
		},
		{
			name: "SUCCESS:Credit card with default custom channel",
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:    merchantId,
				Reference:     c.ReferencePayment,
				PaymentMethod: c.ChannelCreditCard,
				Channel:       "LOCAL_VISA", // NOSONAR
			},
		},
		{
			name: "SUCCESS:Credit card with default other channel",
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:    merchantId,
				Reference:     c.ReferencePayment,
				PaymentMethod: c.ChannelCreditCard,
				Channel:       "LOCAL_AMEX", // NOSONAR
			},
		},
		{
			name: "SUCCESS:Payment funded payout with default fee - standard settlement",
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:       merchantId,
				Reference:        c.ReferencePaymentFundedPayout,
				PaymentMethod:    c.ChannelCreditCard,
				Channel:          "VISA",
				ReferenceAmount:  100000,
				SettlementMethod: c.SettlementMethodStandard,
			},
		},
		{
			name: "SUCCESS:Payment funded payout with instant settlement",
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:       merchantId,
				Reference:        c.ReferencePaymentFundedPayout,
				PaymentMethod:    c.ChannelCreditCard,
				Channel:          "VISA",
				ReferenceAmount:  100000,
				SettlementMethod: c.SettlementMethodStandard,
			},
		},
		{
			name: "SUCCESS:Payment funded payout with merchant fee config",
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        1500,
					AmountType:    c.MerchantFeeAmountType,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeNonPKP,
				}, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.ReferencePaymentFundedPayout,
				PaymentMethod:   c.ChannelCreditCard,
				Channel:         "VISA",
				ReferenceAmount: 100000,
			},
		},
		{
			name: "SUCCESS:Payment funded payout with percentage fee",
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        0,
					AmountType:    c.MerchantFeePercentageType,
					Percentage:    1.5,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeNonPKP,
				}, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.ReferencePaymentFundedPayout,
				PaymentMethod:   c.ChannelCreditCard,
				Channel:         "VISA",
				ReferenceAmount: 100000,
			},
		},
		{
			name: "SUCCESS:Payment funded payout with amount percentage fee",
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        1500,
					AmountType:    c.MerchantFeeAmountPercentageType,
					Percentage:    1.5,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeNonPKP,
				}, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.ReferencePaymentFundedPayout,
				PaymentMethod:   c.ChannelCreditCard,
				Channel:         "VISA",
				ReferenceAmount: 100000,
			},
		},
		{
			name: "SUCCESS:Payment funded payout with inclusive tax",
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        1500,
					AmountType:    c.MerchantFeeAmountType,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeInclusive,
					TaxPercentage: 10,
				}, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.ReferencePaymentFundedPayout,
				PaymentMethod:   c.ChannelCreditCard,
				Channel:         "VISA",
				ReferenceAmount: 100000,
			},
		},
		{
			name: "SUCCESS:Payment funded payout with exclusive tax",
			setupMock: func() {
				merchantRepo.On(
					"DeterminePaymentFundedPayoutFeeByMerchantIdMethodAndChannel", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&merchantModel.MerchantFee{
					Amount:        1500,
					AmountType:    c.MerchantFeeAmountType,
					DeductionType: c.MerchantFeeDeductionTypeDirect,
					TaxType:       c.MerchantTaxTypeExclusive,
					TaxPercentage: 10,
				}, nil)
			},
			request: &feeModel.GetFeeRequest{
				MerchantID:      merchantId,
				Reference:       c.ReferencePaymentFundedPayout,
				PaymentMethod:   c.ChannelCreditCard,
				Channel:         "VISA",
				ReferenceAmount: 100000,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientMock.ClearExpect()

			service := New(logger, feeRepo, merchantRepo, WithRedisClient(redisExt.WrapRedisClient(rdb, nil)))

			test.setupMock()
			if test.request == nil {
				test.request = &feeModel.GetFeeRequest{
					MerchantID: merchantId,
					Reference:  c.ReferenceAccountInquiry,
				}
			}
			if _, _, err := service.GetFeeCalculationAndDetail(context.Background(), test.request); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestDefaultFeeAmount(t *testing.T) {
	tests := []struct {
		reference  string
		wantResult float64
	}{
		{
			reference:  c.TypeDisbursement,
			wantResult: 4_000,
		},
		{
			reference:  c.ReferenceAccountInquiry,
			wantResult: 450,
		},
		{
			reference:  c.TypePayment,
			wantResult: 4_000,
		},
		{
			reference: c.TypeXB,
		},
		{
			reference:  c.ReferencePlatformActivity,
			wantResult: 10_000,
		},
		{
			reference:  c.ReferencePlatformTransfer,
			wantResult: 0,
		},
		{
			reference:  c.ReferencePlatformTransaction,
			wantResult: 0,
		},
		{
			reference:  c.ReferenceWallet,
			wantResult: 0,
		},
		{
			reference:  "OTHERS",
			wantResult: 4_000,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantResult, defaultFeeAmount(test.reference))
	}
}

func TestDefaultFeeAmountForPaymentUsecase(t *testing.T) {

	configContent := `
CREDIT_CARD_REFERENCES:
  DEFAULT_FEE:
    OTHER_CHANNEL:
      AMOUNT: 2000
      PERCENTAGE: 2.5
    CUSTOM_CHANNEL:
      LOCAL_VISA:
        AMOUNT: 1500
        PERCENTAGE: 2.75
PAYMENT_FEE_DEFAULTS:
  EWALLET:
    DANA:
      TYPE: PERCENTAGE
      AMOUNT: 0
      PERCENTAGE: 1.3
    SHOPEEPAY:
      TYPE: PERCENTAGE
      AMOUNT: 0
      PERCENTAGE: 4
  OTHER:
    TYPE: AMOUNT
    AMOUNT: 4000
    PERCENTAGE: 0
INSTALLMENT_FEE:
  DEFAULT:
    AMOUNT: 2000
    PERCENTAGE: 2.5
  CHANNEL:
    BCA:
      TENOR: [3,6,12]
      AMOUNT: [2000,2000,2000]
      PERCENTAGE: [2.5,2.5,2.5]
`

	f, err := os.CreateTemp(os.TempDir(), "*.yml")
	require.NoError(t, err)
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()
	require.NoError(t, os.WriteFile(f.Name(), []byte(configContent), 0777))

	_, _, err = config.LoadConfig(f.Name(), f.Name())
	require.NoError(t, err)

	tests := []struct {
		method         string
		channel        string
		amount         float64
		wantFeeAmount  float64
		wantAmountType string
		wantAmount     float64
		wantPercentage float64
	}{
		{
			method:         paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			amount:         100_000, // NOSONAR
			wantFeeAmount:  4_000,   // NOSONAR
			wantAmountType: c.MerchantFeeAmountType,
			wantAmount:     4_000, // NOSONAR
		},
		{
			method:         paymentConstant.PAYMENT_METHOD_QRIS,
			amount:         100_000, // NOSONAR
			wantFeeAmount:  700,     // NOSONAR
			wantAmountType: c.MerchantFeePercentageType,
			wantPercentage: 0.7, // NOSONAR
		},
		{
			method:         paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			amount:         100_000, // NOSONAR
			wantAmountType: c.MerchantFeeAmountPercentageType,
			wantAmount:     2_000, // NOSONAR
			wantPercentage: 2.5,   // NOSONAR
			wantFeeAmount:  4_500, // NOSONAR
		},
		{
			method:         paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			channel:        "LOCAL_VISA", // NOSONAR
			amount:         100_000,      // NOSONAR
			wantAmountType: c.MerchantFeeAmountPercentageType,
			wantAmount:     1_500, // NOSONAR
			wantPercentage: 2.75,  // NOSONAR
			wantFeeAmount:  4_250, // NOSONAR
		},
		{
			method:         paymentConstant.PAYMENT_METHOD_INSTALLMENT,
			channel:        "BCA_12M", // NOSONAR
			amount:         100_000,   // NOSONAR
			wantAmountType: c.MerchantFeeAmountPercentageType,
			wantAmount:     2_000, // NOSONAR
			wantPercentage: 2.5,   // NOSONAR
			wantFeeAmount:  4_500, // NOSONAR
		},
		{
			method:         "OTHERS", // NOSONAR
			amount:         100_000,  // NOSONAR
			wantFeeAmount:  4_000,    // NOSONAR
			wantAmountType: c.MerchantFeeAmountType,
			wantAmount:     4_000, // NOSONAR
		},
		{
			method:         paymentConstant.PAYMENT_METHOD_EWALLET,
			channel:        c.UnifiedPaymentEWalletDanaAcquirer,
			amount:         100_000, // NOSONAR
			wantAmountType: c.MerchantFeePercentageType,
			wantPercentage: 1.3,   // NOSONAR
			wantFeeAmount:  1_300, // NOSONAR
		},
		{
			method:         paymentConstant.PAYMENT_METHOD_EWALLET,
			channel:        c.UnifiedPaymentEWalletShopeePayAcquirer,
			amount:         100_000, // NOSONAR
			wantAmountType: c.MerchantFeePercentageType,
			wantPercentage: 4,     // NOSONAR
			wantFeeAmount:  4_000, // NOSONAR
		},
		{
			method:         paymentConstant.PAYMENT_METHOD_EWALLET,
			channel:        "GOPAY", // NOSONAR
			amount:         100_000, // NOSONAR
			wantAmountType: c.MerchantFeeAmountType,
			wantAmount:     4_000, // NOSONAR
			wantPercentage: 0,     // NOSONAR
			wantFeeAmount:  4_000, // NOSONAR
		},
	}
	for _, test := range tests {
		feeAmount, amountType, amount, percentage := defaultFeeAmountForPaymentUsecase(&feeModel.GetFeeRequest{
			PaymentMethod:   test.method,
			Channel:         test.channel,
			ReferenceAmount: test.amount,
		})

		assert.Equal(t, fmt.Sprintf("%.6f", test.wantFeeAmount), fmt.Sprintf("%.6f", feeAmount))
		assert.Equal(t, test.wantAmountType, amountType)
		assert.Equal(t, test.wantAmount, amount)
		assert.Equal(t, test.wantPercentage, percentage)
	}
}

func TestGetMerchantFeeXB(t *testing.T) {
	rdb, clientMock := redismock.NewClientMock()
	logger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)

	merchantID := uuid.NewString()
	reference := c.TypeXB
	channel := "LOCAL"

	tests := []struct {
		name       string
		request    *feeModel.GetFeeRequest
		setupMock  func()
		wantResult *merchantModel.MerchantFee
		wantErr    string
	}{
		{
			name: "SUCCESS:Fee found in cache",
			request: &feeModel.GetFeeRequest{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    channel,
			},
			setupMock: func() {
				cacheKey := fmt.Sprintf(c.CacheKeyFmtPayoutTransactionFee, merchantID, strings.ToLower(reference+"-"+channel))
				clientMock.ExpectGet(cacheKey).SetVal(`{}`)
			},
			wantResult: &merchantModel.MerchantFee{},
		},
		{
			name: "SUCCESS:Fee from database",
			request: &feeModel.GetFeeRequest{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    channel,
			},
			setupMock: func() {
				cacheKey := fmt.Sprintf(c.CacheKeyFmtPayoutTransactionFee, merchantID, strings.ToLower(reference+"-"+channel))
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On("GetMerchantFeeXB", mock.Anything, mock.MatchedBy(func(query *merchantModel.MerchantFeeXBQuery) bool {
					return query.MerchantID == merchantID && query.Reference == reference && query.Channel == channel
				})).Once().Return(&merchantModel.MerchantFee{
					MerchantID: merchantID,
					Reference:  reference,
					Channel:    &channel,
					Amount:     15.0,
				}, nil)
				clientMock.ExpectSet(cacheKey, mock.Anything, 15*time.Minute).SetVal("OK")
			},
			wantResult: &merchantModel.MerchantFee{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    &channel,
				Amount:     15.0,
			},
		},
		{
			name: "SUCCESS:Default LOCAL fee when no config found",
			request: &feeModel.GetFeeRequest{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    "LOCAL",
			},
			setupMock: func() {
				cacheKey := fmt.Sprintf(c.CacheKeyFmtPayoutTransactionFee, merchantID, strings.ToLower(reference+"-LOCAL"))
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On("GetMerchantFeeXB", mock.Anything, mock.MatchedBy(func(query *merchantModel.MerchantFeeXBQuery) bool {
					return query.MerchantID == merchantID && query.Reference == reference && query.Channel == "LOCAL"
				})).Once().Return(nil, nil)
			},
			wantResult: &merchantModel.MerchantFee{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    util.ValueToPtr("LOCAL"),
				Percentage: 0.0,
				AmountType: c.MerchantFeeAmountType,
				Amount:     c.XBLocalFee,
				TaxType:    c.MerchantTaxTypeNonPKP,
			},
		},
		{
			name: "SUCCESS:Default SWIFT fee when no config found",
			request: &feeModel.GetFeeRequest{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    "SWIFT",
			},
			setupMock: func() {
				cacheKey := fmt.Sprintf(c.CacheKeyFmtPayoutTransactionFee, merchantID, strings.ToLower(reference+"-SWIFT"))
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On("GetMerchantFeeXB", mock.Anything, mock.MatchedBy(func(query *merchantModel.MerchantFeeXBQuery) bool {
					return query.MerchantID == merchantID && query.Reference == reference && query.Channel == "SWIFT"
				})).Once().Return(nil, nil)
			},
			wantResult: &merchantModel.MerchantFee{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    util.ValueToPtr("SWIFT"),
				Percentage: 0.0,
				AmountType: c.MerchantFeeAmountType,
				Amount:     c.XBSwiftFee,
				TaxType:    c.MerchantTaxTypeNonPKP,
			},
		},
		{
			name: "SUCCESS:Default zero fee for unknown channel",
			request: &feeModel.GetFeeRequest{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    "UNKNOWN",
			},
			setupMock: func() {
				cacheKey := fmt.Sprintf(c.CacheKeyFmtPayoutTransactionFee, merchantID, strings.ToLower(reference+"-UNKNOWN"))
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On("GetMerchantFeeXB", mock.Anything, mock.MatchedBy(func(query *merchantModel.MerchantFeeXBQuery) bool {
					return query.MerchantID == merchantID && query.Reference == reference && query.Channel == "UNKNOWN"
				})).Once().Return(nil, nil)
			},
			wantResult: &merchantModel.MerchantFee{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    util.ValueToPtr("UNKNOWN"),
				Percentage: 0.0,
				AmountType: c.MerchantFeeAmountType,
				Amount:     0.0,
				TaxType:    c.MerchantTaxTypeNonPKP,
			},
		},
		{
			name: "ERROR:Cache error other than nil",
			request: &feeModel.GetFeeRequest{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    channel,
			},
			setupMock: func() {
				cacheKey := fmt.Sprintf(c.CacheKeyFmtPayoutTransactionFee, merchantID, strings.ToLower(reference+"-"+channel))
				clientMock.ExpectGet(cacheKey).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
		{
			name: "ERROR:Database error",
			request: &feeModel.GetFeeRequest{
				MerchantID: merchantID,
				Reference:  reference,
				Channel:    channel,
			},
			setupMock: func() {
				cacheKey := fmt.Sprintf(c.CacheKeyFmtPayoutTransactionFee, merchantID, strings.ToLower(reference+"-"+channel))
				clientMock.ExpectGet(cacheKey).SetErr(redisExt.ErrNil)
				merchantRepo.On("GetMerchantFeeXB", mock.Anything, mock.MatchedBy(func(query *merchantModel.MerchantFeeXBQuery) bool {
					return query.MerchantID == merchantID && query.Reference == reference && query.Channel == channel
				})).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: "some error",
		},
	}
	conf := &config.Config{
		XbCoreProcessorConfig: config.XbCoreProcessorConfig{
			DefaultLocalFee: 10.0,
			DefaultSwiftFee: 25.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			clientMock.ClearExpect()

			service := New(logger, nil, merchantRepo, WithRedisClient(redisExt.WrapRedisClient(rdb, nil)), WithConfig(conf))

			test.setupMock()

			feeService := service.(*FeeService)
			result, err := feeService.getMerchantFeeXB(context.Background(), test.request)

			if test.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.wantResult, result)
			}
		})
	}
}

func TestGetTransactionFeeOnBehalf(t *testing.T) {
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)

	service := New(nil, nil, merchantRepo)

	parentMerchantId := "19d47f58-4af4-4d3b-be93-b3811319db12"

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *feeModel.TrxFeeOnBehalfMetadata
	}{
		{
			name: "ERROR:Get transaction fee on-behalf",
			setupMock: func() {
				merchantRepo.On(
					"GetTransactionFeeOnBehalf", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Config not found",
			setupMock: func() {
				merchantRepo.On(
					"GetTransactionFeeOnBehalf", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantResult: &feeModel.TrxFeeOnBehalfMetadata{
				Type: c.FeeOnBehalfTypeNotSet, AmountType: "AMOUNT",
			},
		},
		{
			name: "SUCCESS:Config found",
			setupMock: func() {
				merchantRepo.On(
					"GetTransactionFeeOnBehalf", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(), c.StringMockType(),
				).Return(&merchantModel.TransactionFeeOnBehalf{
					Type:       "DEFAULT",
					AmountType: "AMOUNT",
					Amount:     2_000,
				}, nil)
			},
			wantResult: &feeModel.TrxFeeOnBehalfMetadata{
				Type:        "DEFAULT",
				AmountType:  "AMOUNT",
				Amount:      2_000,
				FinalAmount: 2_000,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetTransactionFeeOnBehalf(context.Background(), &feeModel.GetTrxFeeOnBehalfRequest{
				MerchantId: parentMerchantId,
			})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
