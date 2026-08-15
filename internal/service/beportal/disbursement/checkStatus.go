package disbursementService

import (
	"context"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"golang.org/x/sync/errgroup"
)

func (s *DisbursementService) CheckTransactionStatus(ctx context.Context, request *disbursementModel.CheckDisbursementTransactionStatusRequest) ([]*disbursementModel.CheckDisbursementTransactionStatusResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/disbursement/CheckTransactionStatus")
	defer span.End()

	var (
		responses []*disbursementModel.CheckDisbursementTransactionStatusResponse
		eg        = new(errgroup.Group)
		mu        sync.Mutex
	)

	for _, disbursementID := range request.DisbursementIDs {
		eg.Go(func() error {
			response, err := s.processCheckTransactionStatus(ctx, disbursementID)
			if err != nil {
				s.logger.Error(ctx, "failed to process check transaction status", pdkLog.Error(err))

				response = &disbursementModel.CheckDisbursementTransactionStatusResponse{
					DisbursementID: disbursementID,
					Error:          true,
					ErrorMessage:   err.Error(),
				}
			}

			mu.Lock()
			responses = append(responses, response)
			mu.Unlock()

			return nil
		})
	}

	_ = eg.Wait()

	return responses, nil
}

func (s *DisbursementService) processCheckTransactionStatus(ctx context.Context, disbursementID string) (*disbursementModel.CheckDisbursementTransactionStatusResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/disbursement/processCheckTransactionStatus")
	defer span.End()

	var response = &disbursementModel.CheckDisbursementTransactionStatusResponse{}

	disbursement, err := s.disbursementRepo.FindByID(ctx, disbursementID)
	if err != nil {
		return nil, err

	} else if disbursement == nil {
		return nil, constant.ErrDisbursementNotFound
	}

	ledger, err := s.orchestratorSvc.FindByReference(ctx, disbursementID, constant.TypeDisbursement)
	if err != nil {
		return nil, err

	} else if ledger == nil {
		s.logger.Error(ctx, "Ledger not found", pdkLog.String("disbursementID", disbursementID))
		return nil, constant.ErrLedgerDetailNotFound
	}

	// Check status from processor
	if ledger.ProcessorReference == constant.SnapCoreProcessor {
		processorData, err := s.snapCoreRepo.CheckStatusByExternalId(ctx, ledger.UUID.String(), false)
		if err != nil {
			s.logger.Error(ctx, "Error check status to snap core bank transfer", pdkLog.String("disbursementID", disbursementID))
		}

		response.ProcessorData = processorData
	}

	response.DisbursementID = disbursementID
	response.TransactionStatus = ledger.Status
	response.TransactionUpdatedAt = ledger.UpdatedAt

	return response, nil
}
