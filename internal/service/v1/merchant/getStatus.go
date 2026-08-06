package merchant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) GetCachedMerchantStatus(ctx context.Context, id string) (*merchantModel.MerchantStatusResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetCachedMerchantStatus")
	defer segment.End()

	cacheKey := fmt.Sprintf(constant.MerchantStatusByIDCacheKey, id)

	cachedResp := merchantModel.MerchantStatusResponse{}
	if err := s.redis.Get(ctx, cacheKey).Scan(&cachedResp); err != nil && !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "error get merchant status cache", logger.String("uuid", id), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Only validate status, because reason_status can be empty
	if cachedResp.Status != "" {
		return &cachedResp, nil
	}

	// Get from DB if not exist
	merchantStatus, err := s.repo.FindStatusByID(ctx, id)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if merchantStatus == nil {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	// Set response to redis
	if err = s.redis.Set(ctx, cacheKey, merchantStatus, time.Duration(s.config.MerchantConfig.CacheStatusDurationInMinutes)*time.Minute).Err(); err != nil {
		s.logger.Error(ctx, "error set merchant status cache", logger.String("uuid", id), logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	return merchantStatus, nil
}
