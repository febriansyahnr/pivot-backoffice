package callbackService

import (
	"context"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	callbackPartnerSvc "github.com/paper-indonesia/pivot-backoffice/pkg/callback"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CallbackService")

type DependFunc func(*CallbackService)

type CallbackService struct {
	logger             logger.ILogger
	cacheService       redisExt.IRedisExt
	callbackRepo       repository.ICallbackRepository
	callbackPartnerSvc callbackPartnerSvc.ICallbackPartner
	merchantSvc        service.IMerchantService
	rmq                rabbitMqExt.IRabbitMQExt
	userSvc            service.IUserService
	merchantRepo       repository.IMerchantRepository
	encryption         vault.IVaultTransit
}

func New(
	logger logger.ILogger,
	cacheService redisExt.IRedisExt,
	callbackRepo repository.ICallbackRepository,
	callbackPartnerSvc callbackPartnerSvc.ICallbackPartner,
	merchantSvc service.IMerchantService,
	depends ...DependFunc,
) service.ICallbackService {
	c := &CallbackService{
		logger:             logger,
		cacheService:       cacheService,
		callbackRepo:       callbackRepo,
		callbackPartnerSvc: callbackPartnerSvc,
		merchantSvc:        merchantSvc,
	}
	for _, f := range depends {
		f(c)
	}
	return c
}

func (s *CallbackService) ActivityLog(ctx context.Context, merchantID, userID *string, info *http.Request, activity string, params map[string]string) {
	if params == nil {
		params = map[string]string{}
	}
	params["referer"] = info.Referer()
	params["user_agent"] = info.UserAgent()
	params["remote_addr"] = info.RemoteAddr
	_ = s.rmq.PublishActivity(ctx, merchantID, userID, constant.TagCallbackSetting, activity, params)
}

func WithRabbitMQExt(rmq rabbitMqExt.IRabbitMQExt) DependFunc {
	return func(cs *CallbackService) {
		cs.rmq = rmq
	}
}

func WithUserService(service service.IUserService) DependFunc {
	return func(cs *CallbackService) {
		cs.userSvc = service
	}
}

func WithMerchantRepository(repo repository.IMerchantRepository) DependFunc {
	return func(cs *CallbackService) {
		cs.merchantRepo = repo
	}
}

func WithVaultTransit(transit vault.IVaultTransit) DependFunc {
	return func(cs *CallbackService) {
		cs.encryption = transit
	}
}
