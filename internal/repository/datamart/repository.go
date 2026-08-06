package datamart

import (
	"context"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
)

type IDatamartPaymentMetrics interface {
	QueryPaymentSuccessRateComparison(ctx context.Context, request paymentModel.QueryPaymentSuccessRateComparisonRequest) (*paymentModel.PaymentSuccessRateComparison, error)
}

type IDatamartDisbursementMetrics interface {
	GetSuccessRateMetrics(ctx context.Context, filter disbursementModel.GetDisbursementInsightFilter) (*disbursementModel.SuccessRateMetrics, error)
	GetSLAMetrics(ctx context.Context, filter disbursementModel.GetDisbursementInsightFilter) (*disbursementModel.SLAMetrics, error)
	QueryDisbursementSuccessRateComparison(ctx context.Context, request disbursementModel.QueryDisbursementSuccessRateComparisonRequest) (*disbursementModel.ComparisonData, error)
	QueryDisbursementSLAComparison(ctx context.Context, request disbursementModel.QueryDisbursementSLAComparisonRequest) (*disbursementModel.ComparisonData, error)
}
