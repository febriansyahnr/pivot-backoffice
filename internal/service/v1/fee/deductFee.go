package feeService

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *FeeService) DeductBalanceForIndirectFeeType(ctx context.Context, date time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/fee/DeductBalanceForIndirectFeeType")
	defer segment.End()

	merchantFees, err := s.merchantRepo.GetMerchantFeeListForBalanceDeduction(ctx)
	if err != nil {
		return fmt.Errorf("get merchant list: %v", err)

	} else if len(merchantFees) == 0 {
		s.logger.Info(ctx, "List of Merchants / Sub-Merchants with deduction type automated is not found")
		return nil
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, tz)
	endOfMonth := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, tz)

	for _, merchantFee := range merchantFees {

		deductionDate := time.Date(date.Year(), date.Month(), merchantFee.DeductionDay, 0, 0, 0, 0, tz)
		if date.Month() != deductionDate.Month() {
			deductionDate = endOfMonth
		}

		if date.Day() != deductionDate.Day() {
			s.logger.Info(ctx,
				"There is no schedule for deduct balance", logger.String("merchantId", merchantFee.MerchantId),
			)
			continue
		}

		if merchantFee.LastDeductDate == nil {
			merchantFee.LastDeductDate = &merchantFee.CreatedAt
		}
		startDate := merchantFee.LastDeductDate.Add(time.Second) // UTC
		endDate := date.Add(-time.Second).UTC()
		period := date.Format("2006-01")

		accumulateTrxFees, err := s.accountTransactionRepo.
			GetAccumulateTransactionFees(ctx, merchantFee.MerchantId, merchantFee.Reference, merchantFee.Method, startDate, endDate)
		if err != nil {
			return fmt.Errorf("get accumulate transaction fees: %v", err)
		}

		s.logger.Info(ctx,
			"Balance deduction for merchant id "+merchantFee.MerchantId,
			logger.String("period", period),
			logger.String("reference", merchantFee.Reference),
			logger.String("method", merchantFee.Method),
			logger.Uint64("totalRows", accumulateTrxFees.TotalRows),
			logger.Float64("totalFees", accumulateTrxFees.TotalFees),
			logger.Float64("totalTaxes", accumulateTrxFees.TotalTaxes),
			logger.String("accountName", accumulateTrxFees.AccountName),
		)

		if accumulateTrxFees.TotalFees > 0 {
			availableBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, merchantFee.MerchantId, accumulateTrxFees.AccountName)
			if err != nil {
				s.logger.Error(ctx, "Failed to get available merchant balance", logger.Error(err))
				return err

			} else if accumulateTrxFees.TotalFees > availableBalance {
				s.logger.Warn(ctx,
					"Deduct balance for indirect fee type -> insufficient balance",
					logger.String("merchantId", merchantFee.MerchantId),
					logger.Float64("availableBalance", availableBalance),
					logger.Float64("totalFees", accumulateTrxFees.TotalFees),
				)

				// Continue next merchant.
				continue
			}

			// Sufficient balance for fee deduction
			err = s.accountTransactionRepo.
				DeductBalanceForIndirectFeeType(ctx, merchantFee.MerchantId, accumulateTrxFees.TransactionIds)
			if err != nil {
				return err
			}
		}

		if err = s.merchantRepo.UpdateMerchantFeeLastDeductionDate(ctx, merchantFee.MerchantId, merchantFee.Reference, endDate); err != nil {
			return fmt.Errorf("update last deduct date: %v", err)
		}
	}

	return nil
}
