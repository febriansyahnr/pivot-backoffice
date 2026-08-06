package paymentConsumerController

import (
	"context"
	"testing"
	"time"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/qr_mpm"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/virtualAccount"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	slackMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/slackExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMain(m *testing.M) {
	_, _ = monitor.New("backend-portal-consumer", "0.0.0.0", "1234")

	m.Run()
}

func TestPaymentVirtualAccountNotification(t *testing.T) {
	vaNumber := "1234123412341234"
	vaPaymentNotificationRequest := &virtualAccount.PaymentNotificationPayload{
		Acquirer: constant.BANK_ACQUIRER_PERMATA,
		Number:   vaNumber,
		Status:   paymentConstant.VirtualAccountStatusPaid,
		PaidAmount: &virtualAccount.Amount{
			Currency: "IDR",
			Value:    "1000000.00",
		},
		ExpiredAt: timestamppb.New(time.Now().Add(24 * time.Hour)),
	}
	input, err := proto.Marshal(vaPaymentNotificationRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name       string
		input      []byte
		channel    string
		mocksSetup func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService)
		wantErr    bool
	}{
		{
			name:    "ERROR: Invalid channel",
			input:   input,
			channel: constant.ChannelBalance,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				// Do nothing
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Invalid input",
			input:   []byte{},
			channel: constant.ChannelVirtualAccount,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				// Do nothing
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Invalid proto input",
			input:   []byte("invalid-input"),
			channel: constant.ChannelVirtualAccount,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				// Do nothing
			},
			wantErr: true,
		},
		{
			name:    "ERROR: Process virtual account payment",
			input:   input,
			channel: constant.ChannelVirtualAccount,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				paymentSvc.On("ProcessVirtualAccountPayment",
					constant.ValueCtxMockType(),
					constant.PtrVANotificationRequestMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

				paymentSvc.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&paymentModel.Payment{}, nil)
			},
			wantErr: true,
		},
		{
			name:    "SUCCESS: Process virtual account payment",
			input:   input,
			channel: constant.ChannelVirtualAccount,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				paymentSvc.On("ProcessVirtualAccountPayment",
					constant.ValueCtxMockType(),
					constant.PtrVANotificationRequestMockType(),
				).Return(nil)

				paymentSvc.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&paymentModel.Payment{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: Process virtual account disbursement top up",
			input:   input,
			channel: constant.ChannelVirtualAccount,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				paymentSvc.On("ProcessVirtualAccountPayment",
					constant.ValueCtxMockType(),
					constant.PtrVANotificationRequestMockType(),
				).Return(constant.ErrPaymentNotFound)

				merchantTopUp.On("ProcessMerchantTopUpWithVirtualAccount",
					constant.ValueCtxMockType(),
					constant.PtrVANotificationRequestMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

				paymentSvc.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&paymentModel.Payment{}, nil)
			},
			wantErr: true,
		},
		{
			name:    "SUCCESS: Process virtual account disbursement top up",
			input:   input,
			channel: constant.ChannelVirtualAccount,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				paymentSvc.On("ProcessVirtualAccountPayment",
					constant.ValueCtxMockType(),
					constant.PtrVANotificationRequestMockType(),
				).Return(constant.ErrPaymentNotFound)

				merchantTopUp.On("ProcessMerchantTopUpWithVirtualAccount",
					constant.ValueCtxMockType(),
					constant.PtrVANotificationRequestMockType(),
				).Return(nil)

				paymentSvc.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&paymentModel.Payment{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: There is no VA processed",
			input:   input,
			channel: constant.ChannelVirtualAccount,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				paymentSvc.On("ProcessVirtualAccountPayment",
					constant.ValueCtxMockType(),
					constant.PtrVANotificationRequestMockType(),
				).Return(constant.ErrPaymentNotFound)

				merchantTopUp.On("ProcessMerchantTopUpWithVirtualAccount",
					constant.ValueCtxMockType(),
					constant.PtrVANotificationRequestMockType(),
				).Return(constant.ErrMerchantTopUpReferenceNotFound)

				paymentSvc.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&paymentModel.Payment{}, nil)
			},
			wantErr: true,
		},
	}

	conf := &config.Config{
		SlackConfig: config.SlackConfig{
			PaymentNotifWebhookURL: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentSvcMock := serviceMocks.NewIPaymentService(t)
			merchantTopUpSvc := serviceMocks.NewIMerchantTopUpService(t)
			orchSvcMock := serviceMocks.NewIOrchestratorService(t)
			rmqExtMock := rabbitMqMocks.NewRabbitMQExt(t)
			merchantSvcMock := serviceMocks.NewIMerchantService(t)
			slackMock := slackMocks.NewSlackNotifier(t)
			unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			logger := logger.NewSlogger(logger.Config{})

			tc.mocksSetup(paymentSvcMock, orchSvcMock, merchantTopUpSvc, rmqExtMock, merchantSvcMock)

			orchSvc := New(conf, logger, paymentSvcMock, merchantTopUpSvc, orchSvcMock, rmqExtMock, merchantSvcMock, unifiedPaymentSvc)
			ctx := context.Background()
			err := orchSvc.ProcessPaymentNotification(ctx, tc.input, tc.channel)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			orchSvcMock.AssertExpectations(t)
			merchantSvcMock.AssertExpectations(t)
			slackMock.AssertExpectations(t)
		})
	}

}

func TestPaymentQrisNotification(t *testing.T) {
	qrisNumber := "QR1234"
	qrisPaymentNotificationRequest := &qr_mpm.QrisPaymentNotificationRequest{
		Acquirer:    constant.BANK_ACQUIRER_BNC,
		ReferenceNo: qrisNumber,
		Status:      paymentConstant.VirtualAccountStatusPaid,
		PaidAmount: &qr_mpm.QrisPaymentNotificationRequest_Amount{
			Currency: "IDR",
			Value:    "1000000.00",
		},
		ExpiredAt: timestamppb.New(time.Now().Add(time.Minute * 5)),
	}

	input, err := proto.Marshal(qrisPaymentNotificationRequest)
	assert.NoError(t, err)

	testCases := []struct {
		name       string
		input      []byte
		channel    string
		mocksSetup func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService)
		wantErr    bool
	}{
		{
			name:    "SUCCESS: successfully process payment notification for payment case",
			input:   input,
			channel: constant.ChannelQris,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				paymentSvc.On(
					"ProcessQrisPayment",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*paymentModel.QrisPaymentNotificationRequest"),
				).Return(nil)

				paymentSvc.On(
					"GetActivePaymentByProcessorReferenceNumber",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&paymentModel.Payment{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "ERROR: when unmarshal data request",
			input:   []byte("invalid input"),
			channel: constant.ChannelQris,
			mocksSetup: func(paymentSvc *serviceMocks.IPaymentService, orchSvc *serviceMocks.IOrchestratorService, merchantTopUp *serviceMocks.IMerchantTopUpService, rmqExt *rabbitMqMocks.RabbitMQExt, merchantSvc *serviceMocks.IMerchantService) {
				// do nothing
			},
			wantErr: true,
		},
	}

	conf := &config.Config{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentSvcMock := serviceMocks.NewIPaymentService(t)
			merchantTopUpSvc := serviceMocks.NewIMerchantTopUpService(t)
			orchSvcMock := serviceMocks.NewIOrchestratorService(t)
			rmqExtMock := rabbitMqMocks.NewRabbitMQExt(t)
			merchantSvcMock := serviceMocks.NewIMerchantService(t)
			slackMock := slackMocks.NewSlackNotifier(t)
			unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			logger := logger.NewSlogger(logger.Config{})

			tc.mocksSetup(paymentSvcMock, orchSvcMock, merchantTopUpSvc, rmqExtMock, merchantSvcMock)

			orchSvc := New(conf, logger, paymentSvcMock, merchantTopUpSvc, orchSvcMock, rmqExtMock, merchantSvcMock, unifiedPaymentSvc)
			ctx := context.Background()
			err := orchSvc.ProcessPaymentNotification(ctx, tc.input, tc.channel)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			orchSvcMock.AssertExpectations(t)
			merchantSvcMock.AssertExpectations(t)
			slackMock.AssertExpectations(t)
		})
	}

}
