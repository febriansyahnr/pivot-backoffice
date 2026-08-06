package qris

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validation"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("QrisService")

type qrisService struct {
	logger       logger.ILogger
	repository   repository.IQrisRepository
	merchantRepo repository.IMerchantRepository
	validator    validation.Validator
	snapRepo     repository.ISnapCoreRepository
	gcs          gcs.IGCSService
	pdf          pdf.PDFGenerator
	config       *config.Config
}

type DependFunc func(*qrisService)

func New(log logger.ILogger, repo repository.IQrisRepository, merchantRepo repository.IMerchantRepository, snapRepo repository.ISnapCoreRepository, depends ...DependFunc) service.IQrisService {
	q := &qrisService{
		logger:       log,
		repository:   repo,
		merchantRepo: merchantRepo,
		validator:    validation.New(),
		snapRepo:     snapRepo,
	}
	for _, f := range depends {
		f(q)
	}
	return q
}

func WithGCSService(gcs gcs.IGCSService) DependFunc {
	return func(qs *qrisService) {
		qs.gcs = gcs
	}
}

func WithPDFGenerator(pdf pdf.PDFGenerator) DependFunc {
	return func(qs *qrisService) {
		qs.pdf = pdf
	}
}

func WithServiceConfig(cfg *config.Config) DependFunc {
	return func(qs *qrisService) {
		qs.config = cfg
	}
}

const (
	certificateEstablishmentType = "CertificateEstablishment"
)

var documents = map[string]map[string]bool{
	"Enterprise": {
		"NationalIdentityCard":       true,
		"BusinessLicense":            true,
		"TaxIdentification":          true,
		"BusinessRegistration":       true,
		"CertificateIncorporation":   true,
		"CertificateNo40":            true,
		"CertificateLastAmendment":   true,
		"CertificateDeedAmendment":   true,
		"CertificateAmendmentAct":    true,
		"CertificateEstablishment":   true,
		"CertificateTaxRegistration": true,
		"BusinessEnvironmentPhoto":   true,
	},
}
