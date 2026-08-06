package paymentCapture

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const tableName = "payment_captures"

var otelTracer = otel.Tracer("PaymentCaptureRepository")

type paymentCaptureRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IPaymentCaptureRepository {
	return &paymentCaptureRepository{db: db, logger: logger}
}