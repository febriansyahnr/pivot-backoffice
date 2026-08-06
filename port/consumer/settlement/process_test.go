package settlementConsumerController_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/settlement"
)

func TestMain(m *testing.M) {
	_, _ = monitor.New("backend-portal-consumer", "0.0.0.0", "1234")

	m.Run()
}

func TestProcessPaymentSettlement(t *testing.T) {
	logger := logger.NewSlogger(logger.Config{})
	service := serviceMocks.NewISettlementService(t)

	handler := New(logger, service)

	req := settlementModel.ProcessSettlementRequest{
		TransactionID:    uuid.NewString(),
		FeeTransactionID: uuid.NewString(),
		MerchantID:       uuid.NewString(),
	}

	body, _ := json.Marshal(req)

	tests := []struct {
		name      string
		request   []byte
		setupMock func()
		wantErr   string
	}{
		{
			name:    "ERROR:Invalid body format",
			request: []byte("B"),
			wantErr: "invalid character 'B' looking for beginning of value",
		},
		{
			name:    "ERROR:Some error",
			request: body,
			setupMock: func() {
				service.On(
					"ProcessSettlement", c.ValueCtxMockType(), mock.AnythingOfType("*settlementModel.ProcessSettlementRequest"),
				).Once().Return(c.ErrSomeErrorForUnitTest)

			},
			wantErr: c.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:    "SUCCESS",
			request: body,
			setupMock: func() {
				service.On(
					"ProcessSettlement", c.ValueCtxMockType(), mock.AnythingOfType("*settlementModel.ProcessSettlementRequest"),
				).Once().Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			if err := handler.ProcessPaymentSettlement(context.Background(), test.request, ""); test.wantErr == "" {
				require.NoError(t, err)

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
