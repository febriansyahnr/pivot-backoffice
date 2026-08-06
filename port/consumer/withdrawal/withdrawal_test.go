package withdrawalConsumer_test

import (
	"context"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/withdrawal"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/withdrawal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestWithdrawalProcess(t *testing.T) {
	logger := pdkLoggerMock.NewILogger(t)
	service := serviceMock.NewIWithdrawalService(t)

	consumer := New(logger, service)

	withdrawalRequest := &pb.WithdrawalRequest{
		MerchantId:           "264f941d-2234-4748-9930-25d6f43528ee",
		AccountName:          c.TypePayment,
		BeneficiaryBankCode:  "002",
		BeneficiaryAccountNo: "1111111111",
		Amount:               10_000,
	}
	rawRequest, _ := proto.Marshal(withdrawalRequest)
	requestMockType := mock.AnythingOfType("*withdrawal.WithdrawalRequest")

	logger.On(
		"Info", c.ValueCtxMockType(), "Proccess automatic withdrawal", c.LoggerFieldMockType(), c.LoggerFieldMockType(), c.LoggerFieldMockType(), c.LoggerFieldMockType(),
	).Return()

	tests := []struct {
		name      string
		body      []byte
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Invalid payload",
			body: []byte("A"),
			setupMock: func() {
				logger.On(
					"Error", c.ValueCtxMockType(), "Failed while unmarshal proto", c.LoggerFieldMockType(),
				).Once().Return()
			},
			wantErr: c.ErrUnmarshalProto,
		},
		{
			name: "ERROR:Some error",
			body: rawRequest,
			setupMock: func() {
				service.On(
					"Create", c.ValueCtxMockType(), requestMockType,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "ERROR:Insufficient",
			body: rawRequest,
			setupMock: func() {
				service.On(
					"Create", c.ValueCtxMockType(), requestMockType,
				).Once().Return(nil, c.ErrInsufficientBalance)
			},
		},
		{
			name: "SUCCESS",
			body: rawRequest,
			setupMock: func() {
				service.On("Create", c.ValueCtxMockType(), requestMockType).Return(&withdrawal.WithdrawalProcessResponse{}, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, consumer.WithdrawalProcess(context.Background(), test.body, ""))
		})
	}
}
