package merchant

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *MerchantService) UpdateStatusByID(ctx context.Context, status, reasonStatus, id string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/UpdateStatusByID")
	defer segment.End()

	if err := s.repo.UpdateStatusByID(ctx, status, reasonStatus, id); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	// update merchant status cache
	cacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, id)
	_ = s.redis.Del(ctx, cacheKey)

	return nil
}
