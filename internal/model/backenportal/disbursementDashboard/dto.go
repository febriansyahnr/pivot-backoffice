package disbursementDashboardModel

type SummaryTransactionDTO struct {
	Count int     `db:"count"`
	Sum   float64 `db:"sum"`
}

type SummaryTransactionByReasonType struct {
	ReasonType string  `db:"reason_type" json:"reasonType"`
	Count      int     `db:"count" json:"count"`
	Sum        float64 `db:"sum" json:"sum"`
}
