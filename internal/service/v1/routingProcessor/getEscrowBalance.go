package routingprocessorService

import (
	"context"

	routingProcessorModelEscrowBalance "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/escrowBalance"
)

func (s *routingProcessorService) GetFlipEscrowBalance(ctx context.Context, processorReference string) (res *routingProcessorModelEscrowBalance.EscrowBalanceResponse, err error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/routingProcessor/GetFlipEscrowBalance")
	defer span.End()

	return s.flipProcessorRepository.GetEscrowBalance(ctx)
}

func (s *routingProcessorService) GetDanaEscrowBalance(ctx context.Context) (res *routingProcessorModelEscrowBalance.EscrowBalanceResponse, err error) {
	ctx, span := tracer.Start(ctx, "internal/service/v1/routingProcessor/GetDanaEscrowBalance")
	defer span.End()

	return s.danaProcessorRepository.GetEscrowBalance(ctx)
}
