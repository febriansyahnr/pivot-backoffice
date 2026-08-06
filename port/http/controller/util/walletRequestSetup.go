package httpControllerUtil

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

type InternalWalletRequestSetup struct {
	config *config.Config
	secret *config.Secret
	logger logger.ILogger
}

func NewInternalWalletRequestSetup(config *config.Config, secret *config.Secret, logger logger.ILogger) *InternalWalletRequestSetup {
	return &InternalWalletRequestSetup{config: config, secret: secret, logger: logger}
}

func (s *InternalWalletRequestSetup) PrepareInternalWalletRequest(r *http.Request) error {
	ctx := r.Context()
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		return pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound)
	}

	r.Header.Set(constant.HeaderXInternalServiceKey, s.secret.WalletBackendSecret.InternalServiceKey)
	r.Header.Set(constant.HeaderXMerchantId, user.MerchantId)

	paths := strings.Split(r.URL.Path, "/")
	path := s.config.WalletBackendConfig.InternalPrefixUrl + strings.Join(paths[4:], "/")
	queryParams := r.URL.Query().Encode()
	if queryParams != "" {
		path += "?" + queryParams
	}
	reqUrl, err := url.Parse(s.config.WalletBackendConfig.Host + path)
	if err != nil {
		s.logger.Error(ctx, "Error parsing wallet backend request URL", logger.Error(err))
		return pkgErr.New(response.HttpErrInternal, errors.New("error parse url"))
	}
	r.URL = reqUrl
	r.Host = reqUrl.Host

	return nil
}
