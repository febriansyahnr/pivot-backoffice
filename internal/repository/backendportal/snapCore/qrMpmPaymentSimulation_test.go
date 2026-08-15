package snapCoreRepository

import (
	"context"
	"testing"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapQrisModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/qris"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQrMpmPaymentSimulation(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	httpRequest := httpMocks.NewIHTTPRequest(t)

	testCases := []struct {
		name      string
		wantError bool
		setupMock func()
	}{
		{
			name:      "ERROR: Call QR MPM Payment simulation to snap-core",
			wantError: true,
			setupMock: func() {
				httpRequest.On("POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.PtrQrMpmPaymentSimulationRequestMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Once().Return([]byte(`{}`), 400, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:      "ERROR: Invalid QR MPM Payment simulation response",
			wantError: true,
			setupMock: func() {
				httpRequest.On("POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.PtrQrMpmPaymentSimulationRequestMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Once().Return([]byte(`{invalid}`), 400, nil)
			},
		},
		{
			name:      "ERROR: QR MPM Payment simulation status code 500",
			wantError: true,
			setupMock: func() {
				httpRequest.On("POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.PtrQrMpmPaymentSimulationRequestMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Once().Return([]byte(`{"test":"OK"}`), 500, nil)
			},
		},
		{
			name:      "SUCCESS",
			wantError: false,
			setupMock: func() {
				httpRequest.On("POST",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.PtrQrMpmPaymentSimulationRequestMockType(),
					mock.AnythingOfType(constant.MockTypeMapStringStringReference),
				).Return([]byte(`{"test":"OK"}`), 200, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			conf := &config.Config{SnapCoreConfig: config.SnapCoreConfig{BaseUrl: ""}}
			secret := &config.Secret{SnapCoreSecret: struct {
				InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
			}{InternalServiceKey: ""}}

			repo := New(conf, secret, logger, httpRequest)
			err := repo.QrMpmPaymentSimulation(context.Background(), &snapQrisModel.QrMpmPaymentSimulationRequest{})
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
