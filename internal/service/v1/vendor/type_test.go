package vendor

import (
	"testing"

	mockRepositories "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	vendorRepo := mockRepositories.NewIVendorRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	service := New(vendorRepo, logger)

	assert.NotNil(t, service)
}
