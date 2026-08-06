package vendor

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("VendorService")

type VendorService struct {
	vendorRepo repository.IVendorRepository
	logger     logger.ILogger
}

func New(vendorRepo repository.IVendorRepository, logger logger.ILogger) service.IVendorService {
	return &VendorService{
		vendorRepo: vendorRepo,
		logger:     logger,
	}
}
