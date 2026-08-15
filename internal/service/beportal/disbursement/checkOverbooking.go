package disbursementService

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankConfig"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
)

func (s *DisbursementService) IsBankcodeOverbookingChannelAllowed(ctx context.Context, bankcode, merchantId string) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/isBankcodeOverbookingChannelAllowed")
	defer segment.End()

	// handling if bankcode is empty
	if bankcode == "" {
		return false
	}

	// Read from cache
	cachedValue, err := s.redisExt.Get(ctx, constant.ListOverbookingBankCacheKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		s.logger.Error(ctx, "Error when get cache list of overbooking bank", logger.Error(err))
		return false
	} else if cachedValue != "" {
		var result []string
		if err = json.Unmarshal([]byte(cachedValue), &result); err != nil {
			s.logger.Error(ctx, "Error when unmarshal cache overbooking bank", logger.Error(err))
			return false
		} else if len(result) > 0 {
			return slices.Contains(result, bankcode)
		}
	}

	var overbookingBankCodes []string

	// Get overbooking bank from consul
	overbookingBankCodes = s.getOverbookingBankCodeListViaFlip(ctx, merchantId)

	// Append with overbooking bank from snapCore
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		From: "Bank-Config-Service",
	})
	bankCodeList, err := s.snapCoreRepo.GetBankCodeList(ctx, &snapCoreModel.GetBankCodeListRequest{
		TransferType: constant.DisbursementBankTransferTypeOverbooking,
		IsActive:     1,
	})
	if err == nil && bankCodeList != nil && bankCodeList.BankCodes != nil && len(*bankCodeList.BankCodes) > 0 {
		overbookingBankCodes = append(overbookingBankCodes, *bankCodeList.BankCodes...)
	}

	cacheValue, _ := json.Marshal(overbookingBankCodes)
	_ = s.redisExt.Set(ctx, constant.ListOverbookingBankCacheKey, cacheValue, time.Hour)

	return slices.Contains(overbookingBankCodes, bankcode)
}
