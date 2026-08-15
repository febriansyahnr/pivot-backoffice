package inboundRepository

import (
	"go.opentelemetry.io/otel"
	
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
)

var tracer = otel.Tracer("InboundRepository")

type repository struct {
	db mySqlExt.IMySqlExt
}

func New(db mySqlExt.IMySqlExt) port.IInboundRepository {
	return &repository{db}
}
