package disbursementDashboardService

import (
	"context"
	"strconv"
	"sync"

	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementDashboardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursementDashboard"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *DisbursementDashboardService) Get(ctx context.Context, filterRequest disbursementDashboardModel.GetDisbursementDashboardFilter) (*disbursementDashboardModel.DisbursementDashboardResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursementDashboard/Get")
	defer segment.End()

	var (
		defaultAmount = commonModel.Amount{
			Currency: "IDR",
			Value:    "0.00",
		}
		defaultSummaryTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: 0,
			Sum:   defaultAmount,
		}

		availableBalance                      commonModel.Amount
		pendingBalance                        commonModel.Amount
		allTransaction                        disbursementDashboardModel.SummaryTransaction
		successTransaction                    disbursementDashboardModel.SummaryTransaction
		failedTransaction                     disbursementDashboardModel.SummaryTransaction
		inProgressTransaction                 disbursementDashboardModel.SummaryTransaction
		waitingTodayTransaction               disbursementDashboardModel.SummaryTransaction
		waitingForTopUpTodayTransaction       disbursementDashboardModel.SummaryTransaction
		waitingTodaySingleTransaction         disbursementDashboardModel.SummaryTransaction
		waitingTodayBulkTransaction           disbursementDashboardModel.SummaryTransaction
		waitingForTopUpTodaySingleTransaction disbursementDashboardModel.SummaryTransaction
		waitingForTopUpTodayBulkTransaction   disbursementDashboardModel.SummaryTransaction
		approvedTransaction                   disbursementDashboardModel.SummaryTransaction
		rejectedTransaction                   disbursementDashboardModel.SummaryTransaction

		response = &disbursementDashboardModel.DisbursementDashboardResponse{
			AvailableBalance:                      defaultAmount,
			PendingBalance:                        defaultAmount,
			AllTodayTransaction:                   defaultSummaryTransaction,
			SuccessTodayTransaction:               defaultSummaryTransaction,
			WaitingTodayTransaction:               defaultSummaryTransaction,
			WaitingForTopUpTodayTransaction:       defaultSummaryTransaction,
			PendingTodayTransaction:               defaultSummaryTransaction,
			FailedTodayTransaction:                defaultSummaryTransaction,
			RejectedTodayTransaction:              defaultSummaryTransaction,
			WaitingTodaySingleTransaction:         defaultSummaryTransaction,
			WaitingTodayBulkTransaction:           defaultSummaryTransaction,
			WaitingForTopUpTodaySingleTransaction: defaultSummaryTransaction,
			WaitingForTopUpTodayBulkTransaction:   defaultSummaryTransaction,
			ApprovedTodayTransaction:              defaultSummaryTransaction,
			AllTransaction:                        defaultSummaryTransaction,
			SuccessTransaction:                    defaultSummaryTransaction,
			PendingTransaction:                    defaultSummaryTransaction,
			RejectedTransaction:                   defaultSummaryTransaction,
			ApprovedTransaction:                   defaultSummaryTransaction,
			FailedTransaction:                     defaultSummaryTransaction,
		}
	)
	wg := sync.WaitGroup{}

	uuidOfMerchantID, err := uuid.Parse(filterRequest.MerchantID)
	if err != nil {
		return nil, err
	}

	// Find disbursement balance by merchantID
	disbursementAccount, err := s.accountRepo.FindMerchantAccountByName(ctx, uuidOfMerchantID, constant.TypeDisbursement)
	if err != nil {
		return response, nil
	}
	if disbursementAccount == nil {
		err = constant.ErrMerchantBalanceNotFound
		s.logger.Error(ctx, err.Error(), logger.Error(err))
		return response, nil
	}

	// get AvailableBalance
	wg.Add(1)
	go func() {
		defer wg.Done()

		availableBalanceFromSvc, err := s.orchestratorSvc.GetAvailableMerchantBalance(
			ctx, filterRequest.MerchantID, disbursementAccount.Name,
		)
		if err != nil {
			return
		}

		availableBalance = commonModel.Amount{
			Value:    strconv.FormatFloat(availableBalanceFromSvc, 'f', 2, 64),
			Currency: disbursementAccount.Currency,
		}
	}()

	// get PendingBalance
	wg.Add(1)
	go func() {
		defer wg.Done()

		pendingBalanceInFloat, err := s.orchestratorSvc.GetPendingBalance(ctx, uuidOfMerchantID.String(), disbursementAccount.Name)
		if err != nil {
			return
		}

		pendingBalance = commonModel.Amount{
			Value:    strconv.FormatFloat(pendingBalanceInFloat, 'f', 2, 64),
			Currency: disbursementAccount.Currency,
		}
	}()

	// get AllTodayTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummaryAll(ctx, filterRequest)
		allTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get SuccessTodayTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummarySuccess(ctx, filterRequest)
		successTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get FailedTodayTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummaryFailed(ctx, filterRequest)
		failedTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get WaitingTodayTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.SummaryWaitingToday(ctx, filterRequest)
		waitingTodayTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get WaitingTodaySingleTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.SummarySingleWaitingToday(ctx, filterRequest)
		waitingTodaySingleTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get WaitingTodayBulkTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.SummaryBulkWaitingToday(ctx, filterRequest)
		waitingTodayBulkTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get WaitingForTopUpTodayTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.SummaryWaitingForTopUpToday(ctx, filterRequest)
		waitingForTopUpTodayTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get WaitingForTopUpTodaySingleTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.SummarySingleWaitingForTopUpToday(ctx, filterRequest)
		waitingForTopUpTodaySingleTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get WaitingForTopUpTodayBulkTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.SummaryBulkWaitingForTopUpToday(ctx, filterRequest)
		waitingForTopUpTodayBulkTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get InProgressTodayTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummaryInProgress(ctx, filterRequest)
		inProgressTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get RejectedTodayTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummaryRejected(ctx, filterRequest)
		rejectedTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	// get ApprovedTodayTransaction
	wg.Add(1)
	go func() {
		defer wg.Done()

		summary := s.disbursementRepo.GetSummaryApproved(ctx, filterRequest)
		approvedTransaction = disbursementDashboardModel.SummaryTransaction{
			Count: summary.Count,
			Sum: commonModel.Amount{
				Currency: disbursementAccount.Currency,
				Value:    strconv.FormatFloat(summary.Sum, 'f', 2, 64),
			},
		}
	}()

	wg.Wait()

	response = &disbursementDashboardModel.DisbursementDashboardResponse{
		AvailableBalance:                      availableBalance,
		PendingBalance:                        pendingBalance,
		AllTodayTransaction:                   allTransaction, // TODO: remove the property to avoid ambiguity
		SuccessTodayTransaction:               successTransaction,
		FailedTodayTransaction:                failedTransaction,
		PendingTodayTransaction:               inProgressTransaction,
		WaitingTodayTransaction:               waitingTodayTransaction,
		WaitingTodaySingleTransaction:         waitingTodaySingleTransaction,
		WaitingTodayBulkTransaction:           waitingTodayBulkTransaction,
		WaitingForTopUpTodayTransaction:       waitingForTopUpTodayTransaction,
		WaitingForTopUpTodaySingleTransaction: waitingForTopUpTodaySingleTransaction,
		WaitingForTopUpTodayBulkTransaction:   waitingForTopUpTodayBulkTransaction,
		RejectedTodayTransaction:              rejectedTransaction,
		ApprovedTodayTransaction:              approvedTransaction,

		AllTransaction:      allTransaction,
		SuccessTransaction:  successTransaction,
		FailedTransaction:   failedTransaction,
		PendingTransaction:  inProgressTransaction,
		RejectedTransaction: rejectedTransaction,
		ApprovedTransaction: approvedTransaction,
	}
	return response, nil
}
