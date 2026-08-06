package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
)

func (s *MerchantService) GetMerchantsByIDs(ctx context.Context, merchantIDs []string) ([]*merchant.Merchant, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/GetMerchantsByIDs")
	defer segment.End()

	return s.repo.GetMerchantsByIDs(ctx, merchantIDs)
}
