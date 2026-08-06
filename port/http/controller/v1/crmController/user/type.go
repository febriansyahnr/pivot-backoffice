package user

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMUserController")

type CRMUserController struct {
	config      *config.Config
	secret      *config.Secret
	userSvc     service.IUserService
	roleSvc     service.IRoleService
	userRoleSvc service.IUserRoleService
	merchantSvc service.IMerchantService
	JWT         jwt.IJwt
	validate    *validator.Validate
	rabbitMqExt rabbitMqExt.IRabbitMQExt
}

type CRMUserControllerFunc func(*CRMUserController)

func WithJWT(jwt jwt.IJwt) CRMUserControllerFunc {
	return func(cc *CRMUserController) {
		cc.JWT = jwt
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) CRMUserControllerFunc {
	return func(cc *CRMUserController) {
		cc.rabbitMqExt = rmq
	}
}

func WithValidator(v *validator.Validate) CRMUserControllerFunc {
	return func(cc *CRMUserController) {
		cc.validate = v
	}
}

func New(
	config *config.Config,
	secret *config.Secret,
	userSvc service.IUserService,
	roleSvc service.IRoleService,
	userRoleSvc service.IUserRoleService,
	merchantSvc service.IMerchantService,
	depends ...CRMUserControllerFunc,
) *CRMUserController {
	cc := &CRMUserController{
		config:      config,
		secret:      secret,
		userSvc:     userSvc,
		roleSvc:     roleSvc,
		userRoleSvc: userRoleSvc,
		merchantSvc: merchantSvc,
	}
	for _, fn := range depends {
		fn(cc)
	}
	return cc
}
