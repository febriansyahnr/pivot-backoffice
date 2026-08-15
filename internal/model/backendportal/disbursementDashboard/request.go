package disbursementDashboardModel

import "time"

type GetDisbursementDashboardFilter struct {
	MerchantID       string
	IsXbPayout       bool
	InsightStartDate time.Time
	InsightEndDate   time.Time
}
