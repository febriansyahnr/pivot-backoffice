package user

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("UserController")

type UserController struct {
	config      *config.Config
	secret      *config.Secret
	userSvc     service.IUserService
	roleSvc     service.IRoleService
	userRoleSvc service.IUserRoleService
	merchantSvc service.IMerchantService
	JWT         jwt.IJwt
	validate    *validator.Validate
	rabbitMqExt rabbitMqExt.IRabbitMQExt

	logger logger.ILogger
}

func New(
	validate *validator.Validate,
	userSvc service.IUserService,
	roleSvc service.IRoleService,
	userRoleSvc service.IUserRoleService,
	merchantSvc service.IMerchantService,
	JWT jwt.IJwt,
	config *config.Config,
	secret *config.Secret,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	logger logger.ILogger) *UserController {
	return &UserController{
		config:      config,
		secret:      secret,
		validate:    validate,
		userSvc:     userSvc,
		roleSvc:     roleSvc,
		userRoleSvc: userRoleSvc,
		merchantSvc: merchantSvc,
		JWT:         JWT,
		rabbitMqExt: rabbitMqExt,
		logger:      logger,
	}
}
