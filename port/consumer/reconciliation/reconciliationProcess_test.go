package reconciliation

import (
	"context"
	"errors"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestReconciliationProcess(t *testing.T) {
	tests := []struct {
		name      string
		body      []byte
		setupMock func(*serviceMocks.IReconciliationService)
		wantErr   bool
	}{
		{
			name: "SUCCESS:Successfully process reconciliation",
			body: []byte(`{"uuid": "1234567890"}`),
			setupMock: func(svc *serviceMocks.IReconciliationService) {
				svc.On("ProcessFile", mock.Anything, "1234567890").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR:Failed to process reconciliation",
			body: []byte(`{"uuid": "1234567890"}`),
			setupMock: func(svc *serviceMocks.IReconciliationService) {
				svc.On("ProcessFile", mock.Anything, "1234567890").Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR:Failed to unmarshal body",
			body: []byte(`invalid`),
			setupMock: func(svc *serviceMocks.IReconciliationService) {
				// empty
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := serviceMocks.NewIReconciliationService(t)
			logger := pdkLoggerMock.NewILogger(t)
			controller := New(logger, mockSvc)

			tc.setupMock(mockSvc)

			err := controller.ReconciliationProcess(context.Background(), tc.body, "")

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockSvc.AssertExpectations(t)
		})
	}
}
