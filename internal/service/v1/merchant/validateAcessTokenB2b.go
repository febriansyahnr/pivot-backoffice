package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) ValidateAccessTokenB2b(ctx context.Context, request *merchantModel.ValidateAccessTokenB2bRequest) (*merchantModel.MerchantAuthTokenClaims, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetAccessTokenB2b")
	defer segment.End()

	claims, err := s.JWT.VerifyMerchantToken(ctx, request.AccessToken)
	if err != nil {
		if err == constant.ErrExpiredMerchantAuth {
			return nil, errPkg.New(responseHttp.HttpErrUnauthorized, constant.ErrExpiredMerchantAuth)
		}

		s.logger.Error(ctx, "error validate merchant token", logger.Error(err))
		return nil, errPkg.New(responseHttp.HttpErrInternal, constant.ErrValidateMerchantAuth)
	}

	if !request.IsSnapRequest && claims.MerchantId != request.MerchantId {
		return nil, errPkg.New(responseHttp.HttpErrUnauthorized, constant.ErrMismatchMerchantAuth)
	}

	return claims, nil
}
