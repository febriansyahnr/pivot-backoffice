package fraudruleservice

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("FdsService")

type FraudRuleServiceFunc func(*FraudRuleService)

type FraudRuleService struct {
	logger               logger.ILogger
	fraudRulesRepository repository.IFraudRulesRepository
}

func New(
	log logger.ILogger,
	fraudrulesrepository repository.IFraudRulesRepository,
	depends ...FraudRuleServiceFunc,
) service.IFraudRuleService {
	service := &FraudRuleService{
		logger:               log,
		fraudRulesRepository: fraudrulesrepository,
	}

	for _, fn := range depends {
		fn(service)
	}

	return service
}
