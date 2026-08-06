package commService_test

import (
	"context"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/commService"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/commService"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestPostEmailHandler(t *testing.T) {
	logger := pdkLoggerMock.NewILogger(t)
	service := serviceMocks.NewICommService(t)

	handler := New(logger, service)

	payload := &pb.EmailRequest{
		Event: "event",
		To:    "email@example.com",
	}
	body, _ := proto.Marshal(payload)

	logger.On(
		"Info", c.ValueCtxMockType(), "Email sending process", loggerMock.Any("details", map[string]string{"merchantId": "", "to": payload.To, "event": payload.Event}),
	).Return()

	tests := []struct {
		name      string
		body      []byte
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Unmarshal proto",
			body: []byte(`A`),
			setupMock: func() {
				logger.On(
					"Error", c.ValueCtxMockType(), "Failed while unmarshal proto", c.LoggerFieldMockType(),
				).Once().Return()
			},
			wantErr: c.ErrUnmarshalProto,
		},
		{
			name: "ERROR:Some error",
			body: body,
			setupMock: func() {
				service.On(
					"PostEmailService", c.ValueCtxMockType(), c.StringMockType(), c.PtrPaperCommEmailRequestMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			body: body,
			setupMock: func() {
				service.On(
					"PostEmailService", c.ValueCtxMockType(), c.StringMockType(), c.PtrPaperCommEmailRequestMockType(),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, handler.PostEmailHandler(context.Background(), test.body, ""))
		})
	}
}
