package crmCreditcardController

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMCreditcardController")

type handler struct {
	config        *config.Config
	secret        *config.Secret
	validator     *validatorExt.Validate
	creditcardSvc service.ICreditCardService
}

func New(cfg *config.Config, secret *config.Secret, creditcardSvc service.ICreditCardService) controller.V1CRMCreditcardController {
	return &handler{
		config:        cfg,
		secret:        secret,
		validator:     validatorExt.New(),
		creditcardSvc: creditcardSvc,
	}
}
