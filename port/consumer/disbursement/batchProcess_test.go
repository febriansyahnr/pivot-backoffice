package disbursementConsumerController_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/disbursement"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/disbursement"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestBatchProcessDisbursement(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	service := serviceMocks.NewIDisbursementService(t)

	handler := New(logger, service, nil)

	logger.On(
		"Info", c.ValueCtxMockType(), "Incoming message on BatchProcessDisbursement", mock.Anything, mock.Anything,
	).Return()
	payload := &pb.BatchProcessDisbursementRequest{
		BulkId:          "123456",
		DisbursementIds: []string{"1", "2", "3"},
	}
	rawPayload, err := proto.Marshal(payload)
	require.NoError(t, err)

	requestMockType := mock.AnythingOfType("*disbursementModel.BatchProcessDisbursementRequest")
	tests := []struct {
		name      string
		body      []byte
		setupMock func()
		wantErr   error
	}{
		{
			name:    "ERROR:Unmarshal proto",
			body:    []byte(`BLABLA`),
			wantErr: c.ErrUnmarshalProto,
		},
		{
			name: "ERROR:Some error", // NOSONAR
			body: rawPayload,
			setupMock: func() {
				service.On(
					"BatchProcessDisbursement", c.ValueCtxMockType(), requestMockType,
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			body: rawPayload,
			setupMock: func() {
				service.On("BatchProcessDisbursement", c.ValueCtxMockType(), requestMockType).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			assert.Equal(t, test.wantErr, handler.BatchProcessDisbursement(context.Background(), test.body, ""))
		})
	}
}
