package commService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CommService")

type communication struct {
	paperComm repository.IPaperCommunicationRepository
}

func New(paperComm repository.IPaperCommunicationRepository) service.ICommService {
	return &communication{paperComm}
}
