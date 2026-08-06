package merchant

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func (s *MerchantService) DormantMerchant(ctx context.Context, date time.Time) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/DormantMerchant")
	defer segment.End()

	// Find all merchant
	merchantIDs, err := s.repo.GetAllActiveMerchantIDs(ctx)
	if err != nil {
		return fmt.Errorf("get all active merchant IDs: %w", err)
	}

	dormantNoTransactionsInDays := s.config.MerchantConfig.DormantCondition.NoTransactionInDays
	lastTransactionDateCondition := date.
		Add(-time.Duration(dormantNoTransactionsInDays) * (24 * time.Hour))

	for _, merchantId := range merchantIDs {
		// Find Merchant
		merchant, errFind := s.repo.FindMerchantByID(ctx, merchantId)
		if errFind != nil {
			return fmt.Errorf("find merchant: %w", err)
		} else if merchant == nil {
			continue
		}

		lastTransaction, err := s.accountTransactionRepo.FindLastMerchantTransactionDate(ctx, merchantId)
		if err != nil {
			return fmt.Errorf("find last merchant transaction: %w", err)
		} else if lastTransaction == nil {
			lastTransaction = &merchant.CreatedAt
		}

		if lastTransaction.Before(lastTransactionDateCondition) {
			// update merchant status to dormant & reason = no transaction
			if errUpdate := s.UpdateStatusByID(ctx,
				constant.MerchantStatusDormant,
				fmt.Sprintf("dormant merchant due to no transactions in %d days", dormantNoTransactionsInDays),
				merchantId,
			); errUpdate != nil {
				return errUpdate
			}
		}
	}

	return nil
}
