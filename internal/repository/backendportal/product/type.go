package productRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const (
	productTableName                 = "products"
	merchantSelectedProductTableName = "merchant_selected_products"
)

var otelTracer = otel.Tracer("ProductRepository")

type ProductRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IProductRepository {
	return &ProductRepository{
		db:     db,
		logger: logger,
	}
}
