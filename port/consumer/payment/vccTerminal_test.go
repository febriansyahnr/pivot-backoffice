package paymentConsumerController_test

import (
	"encoding/json"
	"testing"

	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/payment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVCCTerminalSubmitCharge(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	paymentSvc := serviceMocks.NewIPaymentService(t)

	handler := New(nil, logger, paymentSvc, nil, nil, nil, nil, nil)
	messageBody := []byte(`{"merchantId": "aec6636d-7a02-4d93-a4c5-006b9c235068","batchId": "019c8ede-4414-7576-807c-62b56a2c9652","paymentId": "019c8ede-4414-762e-b7d7-e9866ad61276","encryptedPayload": "encryptedPayload"}`)

	tests := []struct {
		name      string
		body      []byte
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR: JSON unmarshal body", // NOSONAR
			body: []byte(`invalid`),
			setupMock: func() {
				logger.On("Info", mock.Anything, "Charging VCC terminal transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: func() error {
				return pkgErrs.NewNonRetryableError(json.Unmarshal([]byte(`invalid`), &struct{}{}))
			}(),
		},
		{
			name: "ERROR: Some error", // NOSONAR
			body: messageBody,
			setupMock: func() {
				paymentSvc.On("VCCTerminalSubmitCharge", mock.Anything, mock.Anything).Once().Return(assert.AnError)
				logger.On("Info", mock.Anything, "Charging VCC terminal transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			body: messageBody,
			setupMock: func() {
				paymentSvc.On("VCCTerminalSubmitCharge", mock.Anything, mock.Anything).Once().Return(nil)
				logger.On("Info", mock.Anything, "Charging VCC terminal transaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return()
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, handler.VCCTerminalSubmitCharge(t.Context(), test.body, ""))
		})
	}
}
