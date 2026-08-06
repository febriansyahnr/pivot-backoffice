package reconciliation

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const MAX_FILENAME_LENGTH = 100

var otelTracer = otel.Tracer("ReconciliationService")

type ReconciliationService struct {
	config                 *config.Config
	logger                 logger.ILogger
	reconRepo              repository.IReconciliationRepository
	accountTransactionRepo repository.IAccountTransactionRepository
	snapCoreRepo           repository.ISnapCoreRepository
	creditCardCoreProcessorRepo repository.ICreditcardCoreProcessorRepository

	// ext
	gcs         gcs.IGCSService
	excel       xlsx.Exceler
	rabbitMqExt rabbitMqExt.IRabbitMQExt
}

type DependFunc func(*ReconciliationService)

var _ service.IReconciliationService = (*ReconciliationService)(nil)

func New(conf *config.Config, log logger.ILogger, repo repository.IReconciliationRepository, depends ...DependFunc) *ReconciliationService {
	rc := &ReconciliationService{
		config:    conf,
		logger:    log,
		reconRepo: repo,
	}
	for _, f := range depends {
		f(rc)
	}
	return rc
}

func WithGCSService(gcs gcs.IGCSService) DependFunc {
	return func(rc *ReconciliationService) {
		rc.gcs = gcs
	}
}

func WithExcelService(excel xlsx.Exceler) DependFunc {
	return func(rc *ReconciliationService) {
		rc.excel = excel
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) DependFunc {
	return func(rc *ReconciliationService) {
		rc.rabbitMqExt = rmq
	}
}

func WithAccountTransactionRepository(repo repository.IAccountTransactionRepository) DependFunc {
	return func(rc *ReconciliationService) {
		rc.accountTransactionRepo = repo
	}
}

func WithSnapCoreRepository(repo repository.ISnapCoreRepository) DependFunc {
	return func(rc *ReconciliationService) {
		rc.snapCoreRepo = repo
	}
}

func WithCreditCardCoreProcessorRepository(repo repository.ICreditcardCoreProcessorRepository) DependFunc {
	return func(rc *ReconciliationService) {
		rc.creditCardCoreProcessorRepo = repo
	}
}
