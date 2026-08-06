package disbursementService

import (
	"context"
)

func (s *DisbursementService) IsExistReferenceID(ctx context.Context, merchantID, referenceID string) bool {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/IsExistReferenceID")
	defer segment.End()

	return s.disbursementRepo.CountByMerchantAndReference(ctx, merchantID, referenceID) > 0
}
