package activityService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ActivityService")

type ActivityService struct {
	repo repository.IActivityRepository
}

func New(repo repository.IActivityRepository) service.IActivityService {
	return &ActivityService{
		repo: repo,
	}
}
