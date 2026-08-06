package merchantHandler_test

import (
	"testing"

	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/console/handler/merchant"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMigrateMerchantSecretsToEncryption(t *testing.T) {
	service := serviceMocks.NewIMerchantService(t)

	handler := New(nil, service, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				service.On("MigrateMerchantSecretsToEncryption", mock.Anything).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				service.On("MigrateMerchantSecretsToEncryption", mock.Anything).Once().Return(nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, handler.MigrateMerchantSecretsToEncryption(t.Context()))

			service.AssertExpectations(t)
		})
	}
}
