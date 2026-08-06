package installmentplan

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InstallmentPlanService")

type InstallmentPlanService struct {
	logger           logger.ILogger
	repo             repository.IInstallmentPlanRepository
	creditCardSvc    service.ICreditCardService
	merchantSvc      service.IMerchantService
	paymentMethodSvc service.IPaymentMethodService
}

func NewInstallmentPlanService(logger logger.ILogger, repo repository.IInstallmentPlanRepository, creditCardSvc service.ICreditCardService, merchantSvc service.IMerchantService) *InstallmentPlanService {
	return &InstallmentPlanService{
		logger:        logger,
		repo:          repo,
		creditCardSvc: creditCardSvc,
		merchantSvc:   merchantSvc,
	}
}

func WithPaymentMethodService(svc service.IInstallmentPlanService, paymentMethodSvc service.IPaymentMethodService) {
	svc.(*InstallmentPlanService).paymentMethodSvc = paymentMethodSvc
}
