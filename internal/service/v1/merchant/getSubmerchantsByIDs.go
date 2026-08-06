package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
)

func (s *MerchantService) GetSubmerchantsByIDs(ctx context.Context, parentMerchantID string, submerchantIDs []string) ([]*merchant.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetSubmerchantsByIDs")
	defer segment.End()

	return s.repo.GetSubmerchantsByIDs(ctx, parentMerchantID, submerchantIDs)
}
