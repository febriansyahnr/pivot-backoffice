package paymentModel

import (
	"encoding/json"

	"github.com/jmoiron/sqlx/types"
)

type PaymentDashboardInsights struct {
	WaitingForCaptureCount uint                                   `db:"waiting_for_capture_count" json:"waitingForCaptureCount"`
	PaidCount              uint                                   `db:"paid_count" json:"paidCount"`
	PaidTotal              float32                                `db:"paid_total" json:"paidTotal"`
	RefundedCount          uint                                   `db:"refunded_count" json:"refundedCount"`
	RefundedTotal          float32                                `db:"refunded_total" json:"refundedTotal"`
	FailedCount            uint                                   `db:"failed_count" json:"failedCount"`
	FailedTotal            float32                                `db:"failed_total" json:"failedTotal"`
	FailedRefundCount      uint                                   `db:"failed_refund_count" json:"failedRefundCount"`
	SuccessRate            *PaymentSuccessRateComparison          `db:"-" json:"successRate"`
	RawFailureReasons      types.NullJSONText                     `db:"failure_reasons" json:"-"`
	FailureReasons         []PaymentDashboardInsightFailureReason `db:"-" json:"failureReasons"`
}

type PaymentDashboardInsightFailureReason struct {
	Count       uint        `json:"count"`
	Percentage  json.Number `json:"percentage"`
	FailureCode string      `json:"failureCode"`
}

type PaymentSuccessRateComparison struct {
	PreviousSuccessRate json.Number `json:"previousSuccessRate"`
	CurrentSuccessRate  json.Number `json:"currentSuccessRate"`
	DifferenceRate      json.Number `json:"differenceRate"`
}
