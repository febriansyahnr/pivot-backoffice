package merchantRcn

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MerchantRcnService")

type MerchantRcnService struct {
	repo          repository.IMerchantRcnRepository
	cimbProcessor repository.ICimbProcessorRepository
	gcsEncryption encryption.GCSClient
	logger        logger.ILogger
}

type OptionFunc func(*MerchantRcnService)

func New(
	repo repository.IMerchantRcnRepository,
	cimbProcessor repository.ICimbProcessorRepository,
	gcsEncryption encryption.GCSClient,
	logger logger.ILogger,
) service.IMerchantRcnService {
	svc := &MerchantRcnService{
		repo:          repo,
		cimbProcessor: cimbProcessor,
		gcsEncryption: gcsEncryption,
		logger:        logger,
	}

	return svc
}
