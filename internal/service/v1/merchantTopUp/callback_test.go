package merchantTopUp_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	common "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchantTopUp"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendCallback(t *testing.T) {
	log := loggerMock.NewILogger(t)
	repo := repoMocks.NewIMerchantTopUpRepository(t)
	rmq := rmqMock.NewRabbitMQExt(t)

	service := New(&config.Config{}, log, nil, repo, nil,
		WithRabbitMQClient(rmq),
	)

	validRequest := &merchantTopUp.MerchantTopUpCallbackRequest{
		MerchantID:   "merchant_12345",
		MerchantName: "PT MAJU JAYA",
		AccountName:  c.AccountNamePayment,
		Amount: common.Amount{
			Currency: "IDR",
			Value:    fmt.Sprintf("%.2f", 100000.00), // e.g., Rp 100,000.00
		},
		BalanceBefore: common.Amount{
			Currency: "IDR",
			Value:    fmt.Sprintf("%.2f", 500000.00), // e.g., previous balance Rp 500,000.00
		},
		BalanceAfter: common.Amount{
			Currency: "IDR",
			Value:    fmt.Sprintf("%.2f", 600000.00), // 500000 + 100000
		},
		PaymentMethod: merchantTopUp.MerchantTopUpCallbackPaymentMethodObject{
			Type: c.ChannelVirtualAccount,
		},
		PaymentMethodOptions: merchantTopUp.MerchantTopUpCallbackPaymentMethodOptionsObject{
			VirtualAccount: &merchantTopUp.MerchantTopUpCallbackPaymentMethodOptionVAObject{
				Channel:              "BCA",          // e.g., from request.Acquirer
				VirtualAccountNumber: "123456789012", // request.Number
				VirtualAccountName:   "MAJUJAYA",     // merchant.ShortName
			},
		},
		TransactionTime:  time.Date(2025, 5, 14, 15, 30, 0, 0, time.UTC),
		ParentMerchantID: "parent_merchant_67890",
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
		request   *merchantTopUp.MerchantTopUpCallbackRequest
	}{
		{
			name: "ERROR: Callback publish for sub/main account",
			setupMock: func() {
				log.On("Info", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rmq.On("PublishMerchantCallback", c.ValueCtxMockType(), mock.Anything).
					Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Error", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
			request: validRequest,
		},
		{
			name: "ERROR: Callback publish for main account",
			setupMock: func() {
				log.On("Info", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rmq.On("PublishMerchantCallback", c.ValueCtxMockType(), mock.Anything).
					Once().Return(nil)
				rmq.On("PublishMerchantCallback", c.ValueCtxMockType(), mock.Anything).
					Once().Return(c.ErrSomeErrorForUnitTest)
				log.On("Info", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				log.On("Error", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
			request: validRequest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				log.On("Info", mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil)
				rmq.On("PublishMerchantCallback", c.ValueCtxMockType(), mock.Anything).
					Twice().Return(nil)
			},
			wantErr: nil,
			request: validRequest,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			assert.Equal(t, test.wantErr, service.SendCallback(context.Background(), c.CallbackEventMerchantTopUpSuccess, test.request))
		})
	}
}
