package slackConsumerController_test

import (
	"context"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	slackMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/slackExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/slack"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestBatchCreateDisbursement(t *testing.T) {

	logger := pdkLoggerMock.NewILogger(t)
	slack := slackMock.NewSlackNotifier(t)

	handler := New(slack, logger)

	payload := &pb.PostWebhookCmd{
		Color:  pb.Color_GOOD,
		URL:    "http://",
		Title:  "Test",
		Fields: []*pb.AttachmentField{{}},
	}
	rawPayload, err := proto.Marshal(payload)
	require.NoError(t, err)

	requestMockType := mock.AnythingOfType("*slackExt.PostWebhookCmd")

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
				slack.On("PostWebhook", c.ValueCtxMockType(), requestMockType).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS", // NOSONAR
			body: rawPayload,
			setupMock: func() {
				slack.On("PostWebhook", c.ValueCtxMockType(), requestMockType).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			assert.Equal(t, test.wantErr, handler.ProcessSlackPostWebhook(context.Background(), test.body, ""))
		})
	}
}
