package tnc

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("TNCService")

const (
	TNCTemplatePath = "templates/common/tnc_agreement.html"
	TNCSignedURLTTL = 15 * time.Minute
)

// TNCServiceConfig carries the per-feature display info used while rendering
// the TNC PDF and the GCS signed-URL TTL.
type TNCServiceConfig struct {
	TemplatePath       string
	SignedURLTTL       time.Duration
	DocumentFolderName string
}

type TNCService struct {
	repo         repository.ITNCRepository
	merchantRepo repository.IMerchantRepository
	logger       logger.ILogger
	cfg          *config.Config

	gcs      gcs.IGCSService
	activity service.IActivityService
}

// TNCServiceFunc mutates the TNCService during construction.
type TNCServiceFunc func(*TNCService)

// WithGCSService wires the GCS uploader used to store generated PDFs.
func WithGCSService(g gcs.IGCSService) TNCServiceFunc {
	return func(s *TNCService) { s.gcs = g }
}

// WithActivityService wires the activity log writer.
func WithActivityService(a service.IActivityService) TNCServiceFunc {
	return func(s *TNCService) { s.activity = a }
}

// WithTNCConfig wires the per-feature config block.
func WithConfig(c *config.Config) TNCServiceFunc {
	return func(s *TNCService) {
		s.cfg = c
	}
}

func New(
	repo repository.ITNCRepository,
	merchantRepo repository.IMerchantRepository,
	logger logger.ILogger,
	opts ...TNCServiceFunc,
) service.ITNCService {
	s := &TNCService{
		repo:         repo,
		merchantRepo: merchantRepo,
		logger:       logger,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
