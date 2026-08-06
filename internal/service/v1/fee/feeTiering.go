package feeService

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"

	ants "github.com/panjf2000/ants/v2"
)

func (s *FeeService) DetermineFeeTierLvlFromMonthlyTPV(ctx context.Context, date time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/fee/DetermineFeeTierLvlFromMonthlyTPV")
	defer segment.End()

	if date.Day() != 1 {
		s.logger.Error(ctx, "Fee determination can only be done on the 1st of each month", logger.String("inputDate", date.Format(time.DateOnly)))
		return nil
	}

	startDate := time.Date(date.Year(), date.Month()-1, 1, 0, 0, 0, 0, tz).UTC()
	endDate := time.Date(date.Year(), date.Month(), 0, 23, 59, 59, 0, tz).UTC()

	merchantFeeThatUseTiers, err := s.merchantRepo.GetListMerchantFeeThatUseTiers(ctx)
	if err != nil {
		return fmt.Errorf("get list merchant fee that use tiers: %w", err)

	} else if merchantFeeThatUseTiers == nil {
		s.logger.Info(ctx, "There is no merchant fee configuration using tiers")
		return nil
	}

	var (
		chanErr = make(chan error, 1)

		totalProcessed, numberOfWorkers = int64(0), 10
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	workers, err := ants.NewPoolWithFunc(numberOfWorkers, func(merchantFees interface{}) {

		atomic.AddInt64(&totalProcessed, 1)
		request, _ := merchantFees.([]merchant.MerchantFeeThatUseTier)

		chanErr <- s.DetermineFeeTierLvlFromMonthlyTPVPerMerchant(
			ctx, request, startDate, endDate,
		)
	})
	if err != nil {
		s.logger.Error(ctx, "failed to create worker pool", logger.Error(err))
		return err
	}
	defer workers.Release()

	go func() {
		for _, merchantFee := range merchantFeeThatUseTiers {
			if ctx.Err() != nil {
				return
			}
			workers.Invoke(merchantFee)
		}
	}()

	for range len(merchantFeeThatUseTiers) {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-chanErr:
			if err != nil {
				return err
			}
		}
	}

	s.logger.Info(ctx,
		"Data successfully processed with details",
		logger.Int("totalMerchant", len(merchantFeeThatUseTiers)),
		logger.Int64("totalProcessed", totalProcessed),
	)
	return nil
}

func (s *FeeService) DetermineFeeTierLvlFromMonthlyTPVPerMerchant(ctx context.Context, merchantFees []merchant.MerchantFeeThatUseTier, startDate, endDate time.Time) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/fee/DetermineFeeTierLvlFromMonthlyTPVPerMerchant")
	defer segment.End()

	if len(merchantFees) == 0 {
		return errors.New("merchant fee list not found")
	}

	var tpvSummaries map[string]orchestratorModel.CalculatingMerchantTPVSummary

	merchantId := merchantFees[0].MerchantId
	if merchantFees[0].Reference == constant.ReferencePlatformActivity {
		subMerchantIds, errSub := s.merchantRepo.GetSubMerchantIdListByParentId(ctx, merchantId)
		if errSub != nil {
			return fmt.Errorf("get sub merchant id list by parent id: %w", errSub)
		}
		tpvSummaries, err = s.accountTransactionRepo.CalculatingTPVForPlatformActivitiesToDetermineFeeTierLevel(ctx, subMerchantIds, startDate, endDate)

	} else {
		tpvSummaries, err = s.accountTransactionRepo.CalculatingMerchantTPVToDetermineFeeTierLevel(ctx, merchantId, startDate, endDate)
	}
	if err != nil {
		return fmt.Errorf("calculating merchant tpv: %w", err)
	}

	for _, merchantFee := range merchantFees {
		if merchantFee.Reference == constant.ReferenceDisbursement {
			merchantFee.PaymentMethod = util.ValueToPtr(constant.ChannelBankTransfer)
		}

		key := merchantFee.Reference
		if merchantFee.PaymentMethod != nil {
			key += "_" + *merchantFee.PaymentMethod // Ex: DISBURSEMENT_BANK_TRANSFER
		}

		var (
			monthly = tpvSummaries[key]
			value   = monthly.Volume
		)
		if merchantFee.TieringType == constant.FrequencyTieringType {
			value = monthly.Frequency
		}
		determineFeeTierLvl := s.DetermineFeeTierLevel(ctx, value, merchantFee.TieringConfigs)
		if determineFeeTierLvl == nil {
			determineFeeTierLvl = &merchantFee.TieringConfigs[0]
		}
		if err = s.merchantRepo.AppliedFeeFromTiers(ctx, merchantFee.Id, determineFeeTierLvl); err != nil {
			return fmt.Errorf("applied fee from tiers: %w", err)
		}
	}
	return nil
}

func (s *FeeService) DetermineFeeTierLevel(ctx context.Context, value float64, tiers []merchant.FeeTieringConfig) *merchant.FeeTieringConfig {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/fee/DetermineFeeTierLevel")
	defer segment.End()

	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].Tier < tiers[j].Tier
	})

	for _, tier := range tiers {
		if value >= tier.Min && value <= tier.Max {
			return &tier
		}
	}
	return nil
}
