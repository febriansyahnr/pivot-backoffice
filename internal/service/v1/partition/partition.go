package partitionService

import (
	"context"

	partitionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/partition"
)

func (s *service) ReorganizeMonthlyRangePartition(ctx context.Context, request partitionModel.ReorganizeRangePartitionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/partition/ReorganizeMonthlyRangePartition")
	defer segment.End()

	return s.repository.ReorganizeMonthlyRangePartition(ctx, request)
}
