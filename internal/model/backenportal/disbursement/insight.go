package disbursementModel

import (
	"encoding/json"
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

type GetDisbursementInsightFilter struct {
	MerchantID            string    `json:"merchantId"`
	InsightStartDate      time.Time `json:"insightStartDate"`
	InsightEndDate        time.Time `json:"insightEndDate"`
	IsXbPayout            bool      `json:"isXbPayout"`
	IncludePreviousPeriod bool      `json:"includePreviousPeriod"`
}

type DisbursementInsightResponse struct {
	WaitingForApproval SummaryTransaction           `json:"waitingForApproval"`
	Delayed            SummaryTransaction           `json:"delayed"`
	AllStatus          AllStatusSummary             `json:"allStatus"`
	FailureReasons     []SummaryTransactionByReason `json:"failureReasons"`
	SuccessMetrics     SuccessRateMetrics           `json:"successMetrics"`
	SLAMetrics         SLAMetrics                   `json:"slaMetrics"`
	SuccessRate        *ComparisonData              `json:"successRate,omitempty"`
	SLA                *ComparisonData              `json:"sla,omitempty"`
}

type SummaryTransaction struct {
	Count int                `json:"count" example:"10"`
	Sum   commonModel.Amount `json:"sum"`
}

type AllStatusSummary struct {
	Success SummaryTransaction `json:"success"`
	Pending SummaryTransaction `json:"pending"`
	Failed  SummaryTransaction `json:"failed"`
}

type SummaryTransactionByReason struct {
	ReasonType string             `json:"reasonType"`
	Count      int                `json:"count"`
	Sum        commonModel.Amount `json:"sum"`
}

// BigQuery-based metrics for success rate analysis
type SuccessRateMetrics struct {
	OverallSuccessRate float64                  `json:"overallSuccessRate" example:"95.5"`
	AverageSuccessRate float64                  `json:"averageSuccessRate" example:"94.2"`
	DailyBreakdown     []DailySuccessRateMetric `json:"dailyBreakdown"`
}

type DailySuccessRateMetric struct {
	Date               string  `json:"date" example:"2024-01-15"`
	SuccessfulCount    int64   `json:"successfulCount" example:"42"`
	TotalCount         int64   `json:"totalCount" example:"45"`
	SuccessRatePercent float64 `json:"successRatePercent" example:"93.33"`
}

// BigQuery-based metrics for SLA analysis
type SLAMetrics struct {
	AverageProcessingTimeMinutes float64          `json:"averageProcessingTimeMinutes" example:"2.5"`
	DailyBreakdown               []DailySLAMetric `json:"dailyBreakdown"`
}

type DailySLAMetric struct {
	Date                         string  `json:"date" example:"2024-01-15"`
	AverageProcessingTimeMinutes float64 `json:"averageProcessingTimeMinutes" example:"2.1"`
}

type ComparisonData struct {
	Previous   json.Number `json:"previous" example:"94.2"`
	Current    json.Number `json:"current" example:"95.5"`
	Difference json.Number `json:"difference" example:"+1.3"`
}

type DateRange struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type QueryDisbursementSuccessRateComparisonRequest struct {
	MerchantId   string    `json:"merchantId"`
	PrevRange    DateRange `json:"prevRange"`
	CurrentRange DateRange `json:"currentRange"`
}

type QueryDisbursementSLAComparisonRequest struct {
	MerchantId   string    `json:"merchantId"`
	PrevRange    DateRange `json:"prevRange"`
	CurrentRange DateRange `json:"currentRange"`
}

type DisbursementSuccessRateComparison struct {
	MerchantId          string      `json:"merchantId"`
	PreviousSuccessRate interface{} `json:"previousSuccessRate"`
	CurrentSuccessRate  interface{} `json:"currentSuccessRate"`
	DifferenceRate      interface{} `json:"differenceRate"`
}

type DisbursementSLAComparison struct {
	MerchantId               string      `json:"merchantId"`
	PreviousProcessingTime   interface{} `json:"previousProcessingTime"`
	CurrentProcessingTime    interface{} `json:"currentProcessingTime"`
	ProcessingTimeDifference interface{} `json:"processingTimeDifference"`
}
