package merchantCronHandler_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/cron/handler/merchant"
	"github.com/stretchr/testify/assert"
)

func TestDormantMerchant(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	service := serviceMocks.NewIMerchantService(t)

	handler := New(logger, service)

	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "ERROR: Some error",
			setupMock: func() {
				service.On("DormantMerchant", c.ValueCtxMockType(), c.TimeMockType()).Once().Return(c.ErrSomeErrorForUnitTest)

				logger.On(
					"Fatal", c.ValueCtxMockType(), "An error occurred while dormant merchant", c.ZapFieldMockType(),
				).Once().Return()
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				service.On("DormantMerchant", c.ValueCtxMockType(), c.TimeMockType()).Once().Return(nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger.On(
				"Info", c.ValueCtxMockType(), "Start to find and update dormant merchant",
			).Times(1).Return()
			logger.On(
				"Info", c.ValueCtxMockType(), "Dormant merchant completed", c.ZapFieldMockType(), c.ZapFieldMockType(), c.ZapFieldMockType(),
			).Times(1).Return()

			test.setupMock()
			handler.DormantMerchant(context.Background(), "")

			assert.True(t, logger.AssertExpectations(t))
			assert.True(t, service.AssertExpectations(t))
		})
	}
}
