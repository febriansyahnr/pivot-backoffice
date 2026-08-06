package credential

import (
	"context"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	port "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CredentialService")

type service struct {
	log        logger.ILogger
	repo       repository.ICredentialRepository
	rmq        rabbitMqExt.IRabbitMQExt
	userSvc    port.IUserService
	encryption vault.IVaultTransit
}

type DependFunc func(*service)

func New(log logger.ILogger, repo repository.ICredentialRepository, rmq rabbitMqExt.IRabbitMQExt, depends ...DependFunc) port.ICredentialService {
	s := &service{
		log: log, repo: repo, rmq: rmq,
	}

	for _, f := range depends {
		f(s)
	}
	return s
}

func WithUserService(userSvc port.IUserService) DependFunc {
	return func(s *service) {
		s.userSvc = userSvc
	}
}

func WithVaultTransit(transit vault.IVaultTransit) DependFunc {
	return func(s *service) {
		s.encryption = transit
	}
}

func (s *service) ActivityLog(ctx context.Context, merchantID, userID *string, info *http.Request, activity string, params map[string]string) {
	if params == nil {
		params = map[string]string{}
	}
	params["referer"] = info.Referer()
	params["user_agent"] = info.UserAgent()
	params["remote_addr"] = info.RemoteAddr
	_ = s.rmq.PublishActivity(ctx, merchantID, userID, constant.TagCredentialSettings, activity, params)
}
