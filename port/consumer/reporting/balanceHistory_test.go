package reportingConsumer_test

import (
	"encoding/json"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/reporting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBalanceHistory(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	service := serviceMocks.NewIReportingService(t)

	consumer := New(logger, service)

	tests := []struct {
		name      string
		data      []byte
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR: Invalid json",
			data: []byte(`{invalid json}`),
			setupMock: func() {
				logger.On("Error", mock.Anything, "Failed to parse event from the account_transactions table", mock.Anything, mock.Anything).Once().Return()
			},
			wantError: func() error {
				return pkgErrs.NewNonRetryableError(json.Unmarshal([]byte(`{invalid json}`), &struct{}{}))
			}(),
		},
		{
			name: "ERROR: Empty event payload",
			data: []byte(`{"before": null, "after": null, "op": "c", "ts_ms": 1234567890}`),
			setupMock: func() {
				logger.On("Warn", mock.Anything, "Missing before and after data in the event from the account_transactions table", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.NewNonRetryableError(constant.ErrMalformedRequestBodyPayload),
		},
		{
			name: "ERROR: Service returns error",
			data: []byte(`{"before": null, "after": {"uuid": "uuid-1", "merchant_id": "m1"}, "op": "c", "ts_ms": 1234567890, "source": {"table": "account_transactions"}}`),
			setupMock: func() {
				logger.On("Info", mock.Anything, "Process balance history change events", mock.Anything, mock.Anything).Return()
				service.On("UpsertBalanceHistory", mock.Anything, mock.Anything).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS: Create event",
			data: []byte(`{"before": null, "after": {"uuid": "new-uuid", "merchant_id": "m1"}, "op": "c", "ts_ms": 1234567890, "source": {"table": "account_transactions"}}`),
			setupMock: func() {
				service.On("UpsertBalanceHistory", mock.Anything, mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.setupMock()

			assert.Equal(t, tt.wantError, consumer.BalanceHistory(t.Context(), tt.data))

			logger.AssertExpectations(t)
			service.AssertExpectations(t)
		})
	}
}
