package paymentMethodService

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PaymentMethodService")

type PaymentMethodService struct {
	logger            logger.ILogger
	paymentMethodRepo repository.IPaymentMethodRepository
	snapCoreRepo      repository.ISnapCoreRepository
	creditCardRepo    repository.ICreditcardCoreProcessorRepository
	merchantRepo      repository.IMerchantRepository
	paymentRepo       repository.IPaymentRepository

	qrisSvc            service.IQrisService
	merchantSvc        service.IMerchantService
	installmentPlanSvc service.IInstallmentPlanService
	config             *config.Config
	redis              redisExt.IRedisExt

	// internal config for payment checkout validation
	paymentMethodValidationConfig map[string]*PaymentMethodValidationConfig
}

type PaymentMethodServiceFunc func(*PaymentMethodService)
type PaymentMethodValidationConfig struct {
	MinAmount             float64
	MaxAmount             float64
	MaxExpiryDurationUnit string
	MaxExpiryDuration     int
}

func New(
	logger logger.ILogger,
	paymentMethodRepo repository.IPaymentMethodRepository,
	snapCoreRepo repository.ISnapCoreRepository,
	creditCardRepo repository.ICreditcardCoreProcessorRepository,
	depends ...PaymentMethodServiceFunc,
) service.IPaymentMethodService {
	s := &PaymentMethodService{
		logger:                        logger,
		paymentMethodRepo:             paymentMethodRepo,
		snapCoreRepo:                  snapCoreRepo,
		creditCardRepo:                creditCardRepo,
		paymentMethodValidationConfig: map[string]*PaymentMethodValidationConfig{},
	}

	for _, fn := range depends {
		fn(s)
	}

	return s
}

func WithQrisService(svc service.IQrisService) PaymentMethodServiceFunc {
	return func(ds *PaymentMethodService) {
		ds.qrisSvc = svc
	}
}

func WithMerchantService(svc service.IMerchantService) PaymentMethodServiceFunc {
	return func(ds *PaymentMethodService) {
		ds.merchantSvc = svc
	}
}

func WithConfig(cfg *config.Config) PaymentMethodServiceFunc {
	return func(ds *PaymentMethodService) {
		ds.config = cfg

		// the key should be equal with payment_methods.type in backend_portal database
		if cfg.UnifiedPaymentConfig.VirtualAccountConfig != nil {
			ds.paymentMethodValidationConfig[paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT] = &PaymentMethodValidationConfig{
				MinAmount:             util.ValueOfPtr(cfg.UnifiedPaymentConfig.VirtualAccountConfig.MinAmount),
				MaxAmount:             util.ValueOfPtr(cfg.UnifiedPaymentConfig.VirtualAccountConfig.MaxAmount),
				MaxExpiryDurationUnit: cfg.UnifiedPaymentConfig.VirtualAccountConfig.MaxExpiryDurationUnit,
				MaxExpiryDuration:     cfg.UnifiedPaymentConfig.VirtualAccountConfig.MaxExpiryDuration,
			}
		}

		if cfg.UnifiedPaymentConfig.EwalletConfig != nil {
			ds.paymentMethodValidationConfig[paymentConstant.PAYMENT_METHOD_EWALLET] = &PaymentMethodValidationConfig{
				MinAmount: util.ValueOfPtr(cfg.UnifiedPaymentConfig.EwalletConfig.MinAmount),
				MaxAmount: util.ValueOfPtr(cfg.UnifiedPaymentConfig.EwalletConfig.MaxAmount),
			}
		}

		if cfg.UnifiedPaymentConfig.QrConfig != nil {
			ds.paymentMethodValidationConfig[paymentConstant.PAYMENT_METHOD_QRIS] = &PaymentMethodValidationConfig{
				MinAmount:             util.ValueOfPtr(cfg.UnifiedPaymentConfig.QrConfig.MinAmount),
				MaxAmount:             util.ValueOfPtr(cfg.UnifiedPaymentConfig.QrConfig.MaxAmount),
				MaxExpiryDurationUnit: cfg.UnifiedPaymentConfig.QrConfig.MaxExpiryDurationUnit,
				MaxExpiryDuration:     cfg.UnifiedPaymentConfig.QrConfig.MaxExpiryDuration,
			}
		}

		if cfg.UnifiedPaymentConfig.CardConfig != nil {
			ds.paymentMethodValidationConfig[paymentConstant.PAYMENT_METHOD_CREDIT_CARD] = &PaymentMethodValidationConfig{
				MinAmount: util.ValueOfPtr(cfg.UnifiedPaymentConfig.CardConfig.MinAmount),
				MaxAmount: util.ValueOfPtr(cfg.UnifiedPaymentConfig.CardConfig.MaxAmount),
			}
		}
	}
}

func WithRedisClient(redis redisExt.IRedisExt) PaymentMethodServiceFunc {
	return func(ds *PaymentMethodService) {
		ds.redis = redis
	}
}

func WithMerchantRepository(svc repository.IMerchantRepository) PaymentMethodServiceFunc {
	return func(ds *PaymentMethodService) {
		ds.merchantRepo = svc
	}
}

func WithPaymentRepository(repo repository.IPaymentRepository) PaymentMethodServiceFunc {
	return func(ds *PaymentMethodService) {
		ds.paymentRepo = repo
	}
}

func WithInstallmentPlanService(svc service.IPaymentMethodService, installmentSvc service.IInstallmentPlanService) {
	svc.(*PaymentMethodService).installmentPlanSvc = installmentSvc
}
