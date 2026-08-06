package productService

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ProductService")

// restrictedNonKYCProduct is a map that defines products which are restricted for non-KYC (Know Your Customer) merchant.
// The keys in the map represent product identifiers.
var restrictedNonKYCProduct = map[string]bool{
	constant.ProductPlatform: true,
}

type ProductService struct {
	logger       logger.ILogger
	repo         repository.IProductRepository
	merchantRepo repository.IMerchantRepository
}

type optionalServiceProperty func(*ProductService)

func New(logger logger.ILogger, repo repository.IProductRepository, opts ...optionalServiceProperty) service.IProductService {
	svc := &ProductService{
		logger: logger,
		repo:   repo,
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc
}

func WithMerchantRepo(repo repository.IMerchantRepository) optionalServiceProperty {
	return func(s *ProductService) {
		s.merchantRepo = repo
	}
}
