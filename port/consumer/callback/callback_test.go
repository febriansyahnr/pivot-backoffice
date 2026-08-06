package callback_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/callback"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestProcessCallback(t *testing.T) {
	logger := pdkLoggerMock.NewILogger(t)
	rabbitmq := rmqMock.NewRabbitMQExt(t)
	service := serviceMocks.NewICallbackService(t)

	handler := New(logger, service, rabbitmq)

	logger.On(
		"Info", mock.Anything, "Callback request body", mock.Anything,
	).Return()
	logger.On(
		"Warn", mock.Anything, "Failed while sending metrics on merchant callback process", mock.Anything,
	).Return()

	invalidPayoutAnyType, err := anypb.New(&pb.DisbursementInvalidCallbackTest{Status: true})
	require.NoError(t, err)

	tests := []struct {
		name      string
		body      []byte
		directReq *pb.ProcessCallbackRequest
		setupMock func()
		wantErr   error
	}{
		{
			name:    "ERROR:Unmarshal",
			body:    []byte(`ABC`),
			wantErr: c.ErrUnmarshalProto,
		},
		{
			name:      "ERROR:Mapping request data not found",
			directReq: &pb.ProcessCallbackRequest{},
			setupMock: func() {
				logger.On("Info", mock.Anything, "Publish for resubmission via workflow", mock.Anything, mock.Anything).Once().Return()
				rabbitmq.On("Publish", mock.Anything, rabbitMqExt.WorkflowCallbackRoutingKey, (*string)(nil), mock.Anything).Once().Return(nil)
			},
			wantErr: errors.New("mapping request data not found"),
		},
		{
			name: "ERROR:Invalid disbursement data type",
			directReq: &pb.ProcessCallbackRequest{
				Name:    c.CallbackNameDisbursement,
				Request: invalidPayoutAnyType,
			},
			setupMock: func() {
				logger.On("Info", mock.Anything, "Publish for resubmission via workflow", mock.Anything, mock.Anything).Once().Return()
				rabbitmq.On("Publish", mock.Anything, rabbitMqExt.WorkflowCallbackRoutingKey, (*string)(nil), mock.Anything).Once().Return(nil)
			},
			wantErr: errors.New("json unmarshal: json: cannot unmarshal bool into Go struct field DisbursementDataCallbackRequest.status of type string"),
		},
		{
			name: "ERROR:Some error",
			directReq: &pb.ProcessCallbackRequest{
				Name: c.CallbackMasterPaymentSNAPQRIS,
			},
			setupMock: func() {
				service.On("ProcessCallback", c.ValueCtxMockType(), mock.Anything).Once().Return(c.ErrSomeErrorForUnitTest)
				logger.On("Info", mock.Anything, "Publish for resubmission via workflow", mock.Anything, mock.Anything).Once().Return()
				rabbitmq.On("Publish", mock.Anything, rabbitMqExt.WorkflowCallbackRoutingKey, (*string)(nil), mock.Anything).Once().Return(nil)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name:      "SUCCESS:Callback sub account registration",
			directReq: &pb.ProcessCallbackRequest{Name: c.CallbackNameSubAccountRegistration},
			setupMock: func() { service.On("ProcessCallback", c.ValueCtxMockType(), mock.Anything).Return(nil) },
		},
		{
			name:      "SUCCESS:Callback XB",
			directReq: &pb.ProcessCallbackRequest{Name: c.CallbackNameXB},
			setupMock: func() { service.On("ProcessCallback", c.ValueCtxMockType(), mock.Anything).Return(nil) },
		},
		{
			name: "SUCCESS:Payment credit card",
			directReq: &pb.ProcessCallbackRequest{
				Name:  c.CallbackNamePayment,
				Event: c.CallbackEventPaymentCreditcardPaid,
			},
		},
		{
			name: "SUCCESS:Payment virtual account",
			directReq: &pb.ProcessCallbackRequest{
				Name:  c.CallbackNamePayment,
				Event: c.CallbackEventPaymentVirtualAccountPaid,
			},
		},
		{
			name: "SUCCESS:Virtual Card",
			directReq: &pb.ProcessCallbackRequest{
				Name:  c.CallbackNameVirtualCard,
				Event: c.CallbackEventVirtualCardNotification,
			},
		},
		{
			name: "SUCCESS:Wallet Backend Top Up",
			directReq: &pb.ProcessCallbackRequest{
				Name:  c.CallbackNameWalletTopup,
				Event: c.CallbackEventWalletTopUp,
			},
		},
		{
			name: "SUCCESS:Wallet Backend User Activations",
			directReq: &pb.ProcessCallbackRequest{
				Name:  c.CallbackNameWalletUserActivationName,
				Event: c.CallbackEventWalletUserActivation,
			},
		},
		{
			name: "SUCCESS:Withdrawal status callback",
			directReq: &pb.ProcessCallbackRequest{
				Name:  c.CallbackNameWithdrawal,
				Event: fmt.Sprintf(c.CallbackEventWithdrawPattern, "SUCCESS"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.directReq != nil {
				test.body, _ = proto.Marshal(test.directReq)
			}
			assert.Equal(t, test.wantErr, handler.ProcessCallback(context.Background(), test.body, ""))
		})
	}
}
