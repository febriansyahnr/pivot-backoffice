package fdscommon

import (
	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/ruleEvaluations"
	"github.com/shopspring/decimal"
)

type CheckTransactionResponse struct {
	Status      string          `json:"status"`
	Score       decimal.Decimal `json:"score"`
	EvalResults *[]EvalResult   `json:"evalResults"`
}

type EvalResult struct {
	Success        bool                                  `json:"success,omitempty"`
	Code           *string                               `json:"code,omitempty"`
	Source         *string                               `json:"source,omitempty"`
	Message        any                                   `json:"message,omitempty"`
	RuleEvaluation *ruleevaluationsmodel.RuleEvaluations `json:"ruleEvaluation,omitempty"`
	Data           *CheckData                            `json:"-"` // no need to return the json
	Weight         decimal.Decimal                       `json:"-"` // no need to return the json
}

type CheckResponse struct {
	Success bool      `json:"success"`
	Code    *string   `json:"code,omitempty"`
	Source  *string   `json:"source,omitempty"`
	Message any       `json:"message,omitempty"`
	Data    CheckData `json:"data"`
}

type CheckData struct {
	ID        string      `json:"id"`
	Timer     int         `json:"timer"`
	RiskScore int         `json:"riskScore"` // range: 0 - 100
	RiskGroup string      `json:"riskGroup"` // values: very low, low, medium, high, very high
	Link      string      `json:"link"`
	Tags      []CheckTags `json:"tags"`
}

type CheckTags struct {
	ID        string  `json:"id"`
	Action    *string `json:"action"` // string or null
	Name      string  `json:"name"`
	Source    string  `json:"source"`    // values: rule, label
	Type      string  `json:"type"`      // values: label, queue, workflow
	State     *string `json:"state"`     // nullable
	Weight    *int    `json:"weight"`    // nullable
	RiskScore *int    `json:"riskScore"` // nullable
	RiskGroup *string `json:"riskGroup"` // values: very low, low, medium, high, very high
	Link      *string `json:"link"`      // nullable
}
