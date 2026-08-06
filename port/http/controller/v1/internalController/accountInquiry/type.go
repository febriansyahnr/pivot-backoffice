package internalAccountInquiry

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalAccountInquiryController")

type AccountInquiryController struct {
	validator *validatorExt.Validate
	service   service.IAccountInquiryService
}

func New(
	service service.IAccountInquiryService,
) controller.V1InternalAccountInquiryController {
	return &AccountInquiryController{
		validator: validatorExt.New(),
		service:   service,
	}
}
