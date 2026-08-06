package v1InternalUnifiedPaymentController

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V2InternalPaymentController")

type paymentController struct {
	config   *config.Config
	validate *validatorExt.Validate
	monitor  *monitoring.Monitor
	logger   logger.ILogger

	paymentSvc        services.IPaymentService
	unifiedPaymentSvc services.IUnifiedPaymentService
	customerSvc       services.ICustomerService
}

type ControllerFunc func(*paymentController)

func New(
	config *config.Config,
	monitor *monitoring.Monitor,
	options ...ControllerFunc,
) controller.V1InternalUnifiedPaymentController {
	c := &paymentController{
		config:   config,
		validate: validatorExt.New(),
		monitor:  monitor,
	}
	for _, fn := range options {
		fn(c)
	}
	return c
}

func WithLogger(log logger.ILogger) ControllerFunc {
	return func(c *paymentController) {
		c.logger = log
	}
}

func WithPaymentService(svc services.IPaymentService) ControllerFunc {
	return func(c *paymentController) {
		c.paymentSvc = svc
	}
}

func WithUnifiedPaymentService(svc services.IUnifiedPaymentService) ControllerFunc {
	return func(c *paymentController) {
		c.unifiedPaymentSvc = svc
	}
}

func WithCustomerService(svc services.ICustomerService) ControllerFunc {
	return func(c *paymentController) {
		c.customerSvc = svc
	}
}

func (c *paymentController) isPaymentMigrationV1ToV2Enabled(ctx context.Context, merchantId string) bool {
	_, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/unifiedPayment/runIfPaymentMigrationV1ToV2Enabled")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(c.config.Environment)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, c.config.Environment)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagPaymentMigrationV1toV2Enabled, attr, false)
	return enabled
}
