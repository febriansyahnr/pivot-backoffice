package disbursementDashboardModel

import commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

type DisbursementDashboardResponse struct {
	AvailableBalance                      commonModel.Amount `json:"availableBalance"`
	PendingBalance                        commonModel.Amount `json:"pendingBalance"`
	AllTodayTransaction                   SummaryTransaction `json:"allTodayTransaction"`
	SuccessTodayTransaction               SummaryTransaction `json:"successTodayTransaction"`
	WaitingTodayTransaction               SummaryTransaction `json:"waitingTodayTransaction"`
	WaitingTodaySingleTransaction         SummaryTransaction `json:"waitingTodaySingleTransaction"`
	WaitingTodayBulkTransaction           SummaryTransaction `json:"waitingTodayBulkTransaction"`
	WaitingForTopUpTodayTransaction       SummaryTransaction `json:"waitingForTopUpTodayTransaction"`
	WaitingForTopUpTodaySingleTransaction SummaryTransaction `json:"waitingForTopUpTodaySingleTransaction"`
	WaitingForTopUpTodayBulkTransaction   SummaryTransaction `json:"waitingForTopUpTodayBulkTransaction"`
	PendingTodayTransaction               SummaryTransaction `json:"pendingTodayTransaction"`
	RejectedTodayTransaction              SummaryTransaction `json:"rejectedTodayTransaction"`
	ApprovedTodayTransaction              SummaryTransaction `json:"approvedTodayTransaction"`
	FailedTodayTransaction                SummaryTransaction `json:"failedTodayTransaction"`

	// this is for new ability of the date filter
	// should remove prev response after release
	AllTransaction      SummaryTransaction `json:"allTransaction"`
	SuccessTransaction  SummaryTransaction `json:"successTransaction"`
	PendingTransaction  SummaryTransaction `json:"pendingTransaction"`
	RejectedTransaction SummaryTransaction `json:"rejectedTransaction"`
	ApprovedTransaction SummaryTransaction `json:"approvedTransaction"`
	FailedTransaction   SummaryTransaction `json:"failedTransaction"`
}

type SummaryTransaction struct {
	Count int                `json:"count" example:"10"`
	Sum   commonModel.Amount `json:"sum"`
}

type DisbursementApprovalDashboardResponse struct {
	WaitingSingleDisbursement SummaryTransaction `json:"waitingSingleDisbursement"`
	WaitingBulkDisbursement   SummaryTransaction `json:"waitingBulkDisbursement"`
	PendingSingleDisbursement SummaryTransaction `json:"pendingSingleDisbursement"`
	PendingBulkDisbursement   SummaryTransaction `json:"pendingBulkDisbursement"`
}
