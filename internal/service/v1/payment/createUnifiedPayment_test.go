package paymentService

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	jwtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/jwt"
	rabbitMqExtMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	paymentIntMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUnifiedPayment(t *testing.T) {
	jwt := jwtMocks.NewIJwt(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	paymentIntFunc := paymentIntMocks.NewIPaymentInternalDirectFunc(t)
	rmq := rabbitMqExtMocks.NewRabbitMQExt(t)
	rdb, rdbMocks := redismock.NewClientMock()

	config := &config.Config{
		PaymentUIConfig: config.PaymentUIConfig{
			PaymentLinkURL: "https://",
		},
	}
	service := New(
		paymentRepo, logger, nil, nil, nil, nil, nil,
		WithJWTExt(jwt),
		WithConfig(config),
		WithRabbitMQClient(rmq),
		WithInternalDirectFunc(paymentIntFunc),
		WithRedisClient(redisExt.WrapRedisClient(rdb, nil)),
	)

	setupRedisMocks := []func(){
		func() { rdbMocks.ClearExpect() },
	}
	tokenStr := "new-token"
	tokenKey := fmt.Sprintf(c.PaymentTokenCacheKey, util.HashString(tokenStr))

	tests := []struct {
		name       string
		request    *paymentModel.CreateUnifiedPaymentRequest
		setupMock  func()
		wantErr    error
		wantResult *paymentModel.CreateUnifiedPaymentResponse
	}{
		{
			name: "ERROR:Get payment by merchant and reference",
			setupMock: func() {
				paymentRepo.On(
					"GetPaymentByMerchantAndReferenceId", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErr.New(response.HttpErrDatabase, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Payment already exists",
			setupMock: func() {
				paymentRepo.On(
					"GetPaymentByMerchantAndReferenceId", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(&paymentModel.Payment{}, nil)
			},
			wantErr: pkgErr.New(response.HttpErrUnprocessableContent, c.ErrClientReferenceIDAlreadyExist),
		},
		{
			name: "ERROR:Generate payment token",
			setupMock: func() {
				paymentRepo.On(
					"GetPaymentByMerchantAndReferenceId", c.ValueCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Return(nil, nil)

				jwt.On(
					"GeneratePaymentToken", c.StringMockType(), c.TimeMockType(),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantErr: pkgErr.New(response.HttpErrInternal, c.ErrSomeErrorForUnitTest),
		},
		{
			name: "ERROR:Set payment token",
			setupMock: func() {
				jwt.On("GeneratePaymentToken", c.StringMockType(), c.TimeMockType()).Return(tokenStr, nil)

				rdbMocks.ExpectSet(tokenKey, true, 0).SetErr(c.ErrSomeErrorForUnitTest)

				setupRedisMocks = append(setupRedisMocks, func() { rdbMocks.Regexp().ExpectSet("", true, 0).SetVal("true") })
			},
			wantErr: pkgErr.New(response.HttpErrInternal, c.ErrSomeErrorForUnitTest),
		},
		{
			name:      "ERROR:Invalid payment method",
			setupMock: func() { /* No Body Func */ },
			wantErr:   pkgErr.New(response.HttpErrUnprocessableContent, errors.New("invalid payment method: ")),
		},
		{
			name: "ERROR:Create payment virtual account",
			request: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				PaymentMethodOptions: &paymentModel.UnifiedPaymentMethodOption{
					VirtualAccount: &paymentModel.UnifiedPaymentMethodOptionVirtualAccount{},
				},
			},
			setupMock: func() {
				paymentIntFunc.On(
					"CreatePayment", c.ValueCtxMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Create payment QRIS",
			request: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
			},
			setupMock: func() {
				paymentIntFunc.On(
					"CreatePayment", c.ValueCtxMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Error publish expiry message",
			request: &paymentModel.CreateUnifiedPaymentRequest{
				PaymentID:     "ad2c405f-0dab-4b11-a812-b9f3ac0e1650",
				PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
			},
			setupMock: func() {
				paymentIntFunc.On(
					"CreatePayment", c.ValueCtxMockType(), c.StringMockType(), mock.Anything,
				).Return(nil, nil)

				rmq.On(
					"PublishWithDelay", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantResult: &paymentModel.CreateUnifiedPaymentResponse{
				ID:            "ad2c405f-0dab-4b11-a812-b9f3ac0e1650",
				PaymentMethod: paymentConstant.PAYMENT_METHOD_QRIS,
				PaymentLink:   "https://%!(EXTRA string=new-token)",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, redisMock := range setupRedisMocks {
				redisMock()
			}
			test.setupMock()

			if test.request == nil {
				test.request = &paymentModel.CreateUnifiedPaymentRequest{}
			}

			result, err := service.CreateUnifiedPayment(context.Background(), test.request)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
