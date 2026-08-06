package user

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UserService) ValidateInvitationToken(ctx context.Context, token string) (*userModel.ValidateInvitationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/user/ValidateInvitationToken")
	defer segment.End()

	result := &userModel.ValidateInvitationResponse{}
	tokenKey := redisTokenKey("", constant.UserIdentifierUserInvitation, ":", constant.UserTokenNamespace, ":", token)

	err := s.redis.Get(ctx, tokenKey).Scan(result)
	if err == redis.Nil {
		s.logger.Error(ctx, "Key does not exist", logger.String("token", token))
		return nil, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
	} else if err != nil {
		s.logger.Error(ctx, "Error getting value from Redis", logger.String("token", token), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
	}

	merchantTNCStatus, err := s.tncSvc.GetTNCSigningStatus(ctx, result.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "failed to get merchant tnc status", logger.String("merchantID", result.MerchantID))
		return nil, err
	}

	if merchantTNCStatus != nil {
		result.TNCMetadata = merchantTNCStatus
	}

	return result, nil
}
