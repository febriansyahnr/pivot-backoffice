package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) GetInvitationURL(ctx context.Context, merchantId, email string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/GetInvitationURL")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	user, err := s.userRepo.FindUserByEmail(ctx, email)
	if err != nil {
		s.logger.Error(ctx, "find user by email", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("FU: "+constant.InternalErrorFmt, traceId))

	} else if user == nil || user.MerchantId != merchantId {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, constant.ErrDataNotFound)
	}

	feature, keys := constant.UserIdentifierUserInvitation, []string{}

	prefix := redisTokenKey("", feature, ":", constant.UserTokenNamespace, ":")
	hashed := encryption.GenerateHMAC(string(feature), email)

	redisCmd := s.redis.Keys(ctx, prefix+hashed+"*")
	if err := redisCmd.Err(); err != nil {
		s.logger.Error(ctx, "redis scan keys", logger.Error(err))
		return "", pkgErrs.New(response.HttpErrDatabase, fmt.Errorf("RD: "+constant.InternalErrorFmt, traceId))
	}

	if _ = redisCmd.ScanSlice(&keys); len(keys) == 0 {
		return "", pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("invitation url not found"))
	}
	return s.config.MerchantPortalConfig.UserInvitationURL + "?token=" + strings.Replace(keys[0], prefix, "", 1), nil
}
