package menuService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MenuService")

type MenuService struct {
	repo   repository.IMenuRepository
	logger logger.ILogger

	productService service.IProductService
}

type MenuServiceFunc func(menuService *MenuService)

func New(repo repository.IMenuRepository, logger logger.ILogger, depends ...MenuServiceFunc) service.IMenuService {
	s := &MenuService{
		repo:   repo,
		logger: logger,
	}

	for _, fn := range depends {
		fn(s)
	}

	return s
}

func WithProductService(svc service.IProductService) MenuServiceFunc {
	return func(ds *MenuService) {
		ds.productService = svc
	}
}
