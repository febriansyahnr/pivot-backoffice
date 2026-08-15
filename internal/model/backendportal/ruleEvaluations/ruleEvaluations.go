package ruleevaluationsmodel

import (
	"time"

	"github.com/shopspring/decimal"
)

type RuleEvaluations struct {
	UUID        string          `json:"uuid" db:"uuid"`
	ReferenceID string          `json:"referenceId" db:"reference_id"`
	RuleID      string          `json:"ruleId" db:"rule_id"`
	Result      string          `json:"result" db:"result"`
	Score       decimal.Decimal `json:"score" db:"score"`
	Reason      string          `json:"reason" db:"reason"`
	EvaluatedAt time.Time       `json:"evaluedAt" db:"evaluated_at"`
	Provider    string          `db:"-" json:"provider,omitempty"`
}
