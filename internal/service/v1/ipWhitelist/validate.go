package ipwhitelistService

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
)

func (s *IPWhitelistService) ValidateIP(ctx context.Context, merchantId, ip string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/ipWhitelist/ValidateIP")
	defer segment.End()

	// Validate that the provided IP is a valid IPv4 or IPv6 address
	if net.ParseIP(ip) == nil {
		s.logger.Error(ctx, "invalid IP address format", logger.Any("ip", ip), logger.Any("merchantId", merchantId))
		return pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidIPAddress)
	}

	if s.config != nil && util.Contains(s.config.WhitelistedIPs, ip) {
		s.logger.Info(ctx, "IP is whitelisted in config as internal ip", logger.Any("ip", ip))
		return nil
	}

	key := fmt.Sprintf(constant.CacheIPWhitelistKey, merchantId)
	value, err := s.redisCache.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		err = s.updateCache(ctx, merchantId)
		if err != nil {
			s.logger.Error(ctx, "error load configuration", logger.Error(err), logger.Any("merchantId", merchantId))
			return err
		}
		value, err = s.redisCache.Get(ctx, key).Result()
	}
	if err != nil {
		s.logger.Error(ctx, "error when retrieve ip whitelist configuration from cache", logger.Error(err), logger.Any("merchantId", merchantId), logger.Any("ip", ip))
		return pkgErrors.New(response.HttpErrInternal, constant.ErrGetIPWhitelistConfigurationList)
	}
	if value == "" {
		return nil
	}

	configList := strings.Split(value, "|")
	for _, configData := range configList {
		data := strings.Split(configData, ",")
		configIP := data[0]
		subnet := data[1]
		action := data[2]

		isMatch := util.IsIPMatch(configIP, subnet, ip)
		if !isMatch {
			continue
		}
		if action == constant.ActionAllow {
			return nil
		}
		if action == constant.ActionBlock {
			return pkgErrors.New(response.HttpErrForbidden, constant.ErrForbiddenIPAddress)
		}
	}

	s.logger.Info(ctx, "no configuration matched with IP request", logger.Any("ip", ip), logger.Any("merchantId", merchantId))
	err = constant.ErrForbiddenIPAddress

	if s.config.Environment == constant.EnvironmentStaging {
		err = fmt.Errorf("%s, please check ip %s in your whitelist config", constant.ErrForbiddenIPAddress, ip)
	}

	return pkgErrors.New(response.HttpErrForbidden, err)
}
