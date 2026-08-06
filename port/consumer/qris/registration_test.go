package qris_test

import (
	"context"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/qr_mpm"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/qris"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	logger := pdkLoggerMock.NewILogger(t)
	logger.On(
		"Info", c.ValueCtxMockType(), c.StringMockType(), c.LoggerFieldMockType(), c.LoggerFieldMockType(), c.LoggerFieldMockType(),
	).Return()

	service := serviceMocks.NewIQrisService(t)

	handler := New(logger, service)

	req := qr_mpm.RegistrationCallbackRequest{
		ApplicationCode: "1234567890",
		ApplymentCode:   "1234567890",
		MID:             "1234567890",
		AuditStatus:     "APPROVED",
		ResultCode:      "00",
		DateTime:        timestamppb.New(time.Now()),
	}

	body, _ := proto.Marshal(&req)

	tests := []struct {
		name      string
		request   []byte
		setupMock func()
		wantErr   string
	}{
		{
			name:    "ERROR:Invalid body format",
			request: []byte("B"),
			wantErr: "unmarshal proto",
		},
		{
			name:    "ERROR:Some error",
			request: body,
			setupMock: func() {
				service.On(
					"RegistrationCallback", c.ValueCtxMockType(), mock.AnythingOfType("*qris.RegistrationCallback"),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:    "SUCCESS",
			request: body,
			setupMock: func() {
				service.On(
					"RegistrationCallback", c.ValueCtxMockType(), mock.AnythingOfType("*qris.RegistrationCallback"),
				).Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			if err := handler.Process(context.Background(), test.request, ""); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
