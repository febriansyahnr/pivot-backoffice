package fdsservice

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("FdsService")

type FdsServiceFunc func(*FdsService)

type FdsService struct {
	cfg                           *config.Config
	logger                        logger.ILogger
	fraudRulesRepository          repository.IFraudRulesRepository
	ruleEvaluationsRepository     repository.IRuleEvaluationsRepository
	accountTransactionsRepository repository.IAccountTransactionRepository
	paymentRepository             repository.IPaymentRepository
	merchantRepository            repository.IMerchantRepository
	thirdPartyProcessor           map[string]repository.IFdsProcessorRepository // mapping for common functionalities at third party
	outboundRepository            repository.IOutboundRepository
	rabbitMq                      rabbitMqExt.IRabbitMQExt
	customerRepository            repository.ICustomerRepository
	paymentMethodRepository       repository.IPaymentMethodRepository
}

func New(
	cfg *config.Config,
	log logger.ILogger,
	fraudrulesrepository repository.IFraudRulesRepository,
	ruleEvaluationsRepository repository.IRuleEvaluationsRepository,
	accountTransactionsRepository repository.IAccountTransactionRepository,
	paymentRepository repository.IPaymentRepository,
	merchantRepository repository.IMerchantRepository,
	thirdPartyProcessor map[string]repository.IFdsProcessorRepository,
	depends ...FdsServiceFunc,
) service.IFdsService {
	service := &FdsService{
		cfg:                           cfg,
		logger:                        log,
		fraudRulesRepository:          fraudrulesrepository,
		ruleEvaluationsRepository:     ruleEvaluationsRepository,
		accountTransactionsRepository: accountTransactionsRepository,
		paymentRepository:             paymentRepository,
		merchantRepository:            merchantRepository,
		thirdPartyProcessor:           thirdPartyProcessor,
	}

	for _, fn := range depends {
		fn(service)
	}

	return service
}

func WithRabbitMqExt(rabbitMq rabbitMqExt.IRabbitMQExt) FdsServiceFunc {
	return func(rs *FdsService) {
		rs.rabbitMq = rabbitMq
	}
}

func WithOutboundRepository(outboundRepository repository.IOutboundRepository) FdsServiceFunc {
	return func(rs *FdsService) {
		rs.outboundRepository = outboundRepository
	}
}

func WithCustomerRepository(customerRepository repository.ICustomerRepository) FdsServiceFunc {
	return func(rs *FdsService) {
		rs.customerRepository = customerRepository
	}
}

func WithPaymentMethodRepository(paymentMethodRepository repository.IPaymentMethodRepository) FdsServiceFunc {
	return func(rs *FdsService) {
		rs.paymentMethodRepository = paymentMethodRepository
	}
}

func (s *FdsService) GetScoreThreshold() int64 {
	fdsFF := constant.GetFdsFeatureFlag(s.cfg.Environment)
	if fdsFF == nil {
		return s.cfg.FdsConfig.ScoreThreshold
	}

	return fdsFF.ScoreThreshold
}

func (s *FdsService) GetBinLength() int64 {
	fdsFF := constant.GetFdsFeatureFlag(s.cfg.Environment)
	if fdsFF == nil {
		return s.cfg.FdsConfig.BinLength
	}

	return fdsFF.BinLength
}
