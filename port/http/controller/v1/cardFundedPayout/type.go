package cardFundedPayoutController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CardFundedPayoutController")

type handler struct {
	config   *config.Config
	validate *validatorExt.Validate
	logger   logger.ILogger

	// Service
	cardFundedPayoutService service.ICardFundedPayoutService
	feeService              service.IFeeService
	merchantService         service.IMerchantService
	vendorService           service.IVendorService
}

type Option func(*handler)

func New(
	config *config.Config,
	validate *validatorExt.Validate,
	cardFundedPayoutService service.ICardFundedPayoutService,
	depends ...Option,
) controller.V1CardFundedPayoutController {
	c := &handler{
		config:                  config,
		validate:                validate,
		cardFundedPayoutService: cardFundedPayoutService,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithLogger(log logger.ILogger) Option {
	return func(c *handler) {
		c.logger = log
	}
}

func WithVendorService(vendorService service.IVendorService) Option {
	return func(c *handler) {
		c.vendorService = vendorService
	}
}

func WithMerchantService(svc service.IMerchantService) Option {
	return func(c *handler) {
		c.merchantService = svc
	}
}

func WithFeeService(svc service.IFeeService) Option {
	return func(c *handler) {
		c.feeService = svc
	}
}
