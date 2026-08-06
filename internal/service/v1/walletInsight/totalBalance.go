package walletInsightService

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/walletInsights"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
)

func (s *WalletInsightService) TotalBalance(ctx context.Context, merchantId string, hardRefresh bool) (*walletInsights.MerchantTotalBalance, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/walletInsight/TotalBalance")
	defer segment.End()

	var (
		cacheKey      = fmt.Sprintf("backend-portal:wallet:insights:total-balance:%s", merchantId)
		expiration    = time.Minute * 10
		lastUpdatedAt = time.Now().UTC()
	)

	if hardRefresh {
		result, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err != nil && err != redis.Nil {
			s.logger.Error(ctx, "error when retrieve total balance from redis", logger.Error(err), logger.String("key", cacheKey))
		}
		if result != "" {
			var totalBalance walletInsights.MerchantTotalBalance
			err = json.Unmarshal([]byte(result), &totalBalance)
			if err != nil {
				return nil, err
			}
			return &totalBalance, nil
		}
	}

	totalBalance, err := s.orchestratorSvc.GetWalletCustomersTotalBalance(ctx, &orchestrator_model.GetWalletTotalBalanceRequest{
		MerchantID: merchantId,
		Status: []string{
			constant.StatusSuccess,
			constant.StatusPending,
		},
		IncludeIndirectFee: false,
	})
	if err != nil {
		return nil, err
	}

	response := &walletInsights.MerchantTotalBalance{
		TotalBalance:  totalBalance,
		LastUpdatedAt: lastUpdatedAt,
	}
	b, _ := json.Marshal(response)
	_, err = s.redisClient.Set(ctx, cacheKey, string(b), expiration).Result()
	if err != nil {
		s.logger.Error(ctx, "error when store total balance into redis cache.", logger.Error(err), logger.String("key", cacheKey))
	}

	return response, nil
}
