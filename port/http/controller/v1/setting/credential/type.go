package credential

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CredentialSettingController")

type handler struct {
	validate       *validatorExt.Validate
	securitySecret *config.SecuritySecret
	service        service.ICredentialService
}

func New(validate *validatorExt.Validate, securitySecret *config.SecuritySecret, service service.ICredentialService) controller.V1CredentialSettingController {
	return &handler{validate, securitySecret, service}
}
