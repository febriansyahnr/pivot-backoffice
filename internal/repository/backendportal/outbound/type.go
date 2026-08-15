package outbound

import (
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("OutboundRepository")

type repository struct {
	db mySqlExt.IMySqlExt
}

func New(db mySqlExt.IMySqlExt) port.IOutboundRepository {
	return &repository{db}
}
