package disbursementDashboardService

import (
	"context"
	"strconv"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
)

func (s *DisbursementDashboardService) GetApprovalDashboard(ctx context.Context, merchantID uuid.UUID) (*disbursementDashboardModel.DisbursementApprovalDashboardResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursementDashboard/GetApprovalDashboard")
	defer segment.End()

	var (
		wg = sync.WaitGroup{}

		defaultSummaryTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: 0,
			Sum: commonModel.Amount{
				Currency: "IDR",
				Value:    "0.00",
			},
		}

		waitingSingleDisbursement disbursementDashboardModel.SummaryTransaction
		waitingBulkDisbursement   disbursementDashboardModel.SummaryTransaction
		pendingSingleDisbursement disbursementDashboardModel.SummaryTransaction
		pendingBulkDisbursement   disbursementDashboardModel.SummaryTransaction

		response = &disbursementDashboardModel.DisbursementApprovalDashboardResponse{
			WaitingSingleDisbursement: defaultSummaryTransaction,
			WaitingBulkDisbursement:   defaultSummaryTransaction,
			PendingSingleDisbursement: defaultSummaryTransaction,
			PendingBulkDisbursement:   defaultSummaryTransaction,
		}

		filter = disbursementDashboardModel.GetDisbursementDashboardFilter{
			MerchantID: merchantID.String(),
			IsXbPayout: false,
		}
	)

	disbursementAccount, err := s.accountRepo.FindMerchantAccountByName(ctx, merchantID, constant.TypeDisbursement)
	if err != nil {
		return nil, err
	} else if disbursementAccount == nil {
		s.logger.Error(ctx, constant.ErrMerchantBalanceNotFound.Error(), logger.Error(constant.ErrMerchantBalanceNotFound))
		return response, nil
	}

	errorCh := make(chan error, 4)

	fetchData := func(fetch func(context.Context, disbursementDashboardModel.GetDisbursementDashboardFilter) (disbursementDashboardModel.SummaryTransactionDTO, error), result *disbursementDashboardModel.SummaryTransaction) {
		defer wg.Done()
		dto, err := fetch(ctx, filter)
		if err != nil {
			errorCh <- err
			return
		}
		*result = convertDTOToSummaryTransaction(dto, disbursementAccount.Currency)
	}

	wg.Add(1)
	go fetchData(s.disbursementRepo.CountWaitingSingleDisbursement, &waitingSingleDisbursement)

	wg.Add(1)
	go fetchData(s.disbursementRepo.CountWaitingBulkDisbursement, &waitingBulkDisbursement)

	wg.Add(1)
	go fetchData(s.disbursementRepo.CountPendingSingleDisbursement, &pendingSingleDisbursement)

	wg.Add(1)
	go fetchData(s.disbursementRepo.CountPendingBulkDisbursement, &pendingBulkDisbursement)

	wg.Wait()
	close(errorCh)

	for err := range errorCh {
		if err != nil {
			return nil, err
		}
	}

	response = &disbursementDashboardModel.DisbursementApprovalDashboardResponse{
		WaitingSingleDisbursement: waitingSingleDisbursement,
		WaitingBulkDisbursement:   waitingBulkDisbursement,
		PendingSingleDisbursement: pendingSingleDisbursement,
		PendingBulkDisbursement:   pendingBulkDisbursement,
	}
	return response, nil
}

func convertDTOToSummaryTransaction(dto disbursementDashboardModel.SummaryTransactionDTO, currency string) disbursementDashboardModel.SummaryTransaction {
	return disbursementDashboardModel.SummaryTransaction{
		Count: dto.Count,
		Sum: commonModel.Amount{
			Currency: currency,
			Value:    strconv.FormatFloat(dto.Sum, 'f', 2, 64),
		},
	}
}
