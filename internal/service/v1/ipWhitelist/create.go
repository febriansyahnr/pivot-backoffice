package ipwhitelistService

import (
	"context"
	"fmt"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *IPWhitelistService) Create(ctx context.Context, req *ipwhitelistModel.CreateIPWhitelistConfiguration) (*ipwhitelistModel.IPWhitelistConfiguration, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ipWhitelist/Create")
	defer segment.End()

	configuration, err := ipwhitelistModel.New(req)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}

	err = s.whitelistRepo.Create(ctx, configuration)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrCreateIPWhitelistConfiguration)
	}

	err = s.updateCache(ctx, req.MerchantID)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}

func (s *IPWhitelistService) updateCache(ctx context.Context, merchantId string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ipWhitelist/updateCache")
	defer segment.End()

	configList, _, err := s.whitelistRepo.List(ctx, &ipwhitelistModel.GetIPWhitelistConfiguration{
		MerchantID: merchantId,
		Status:     constant.StatusActive,
	})
	if err != nil {
		s.logger.Error(ctx, "error get config list for update ip whitelist cache", logger.Error(err))
		return pkgErrors.New(response.HttpErrInternal, constant.ErrGetIPWhitelistConfigurationList)
	}

	dataList := []string{}
	for _, config := range configList {
		data := config.IP + "," + config.Subnet + "," + config.Action
		dataList = append(dataList, data)
	}

	key := fmt.Sprintf(constant.CacheIPWhitelistKey, merchantId)
	value := strings.Join(dataList, "|")
	_, err = s.redisCache.Set(ctx, key, value, 0).Result()
	if err != nil {
		s.logger.Error(ctx, "error update ip whitelist configuration to cache", logger.Error(err), logger.Any("merchantId", merchantId))
		return pkgErrors.New(response.HttpErrInternal, constant.ErrCachedIPWhitelistConfiguration)
	}
	return nil

}
