package location

import (
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("LocationRepository")

type repository struct {
	db mySqlExt.IMySqlExt
}

func New(db mySqlExt.IMySqlExt) port.IAddrLocationRepository {
	return &repository{db}
}

const (
	provinceTableName = "provinces"
	cityTableName     = "cities"
	districtTableName = "districts"
)
