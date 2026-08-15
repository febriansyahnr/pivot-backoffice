package fdscommon

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestFdsRiskAssessmentUpdate(t *testing.T) {
	now := time.Now()
	score := decimal.NewFromInt(85)

	tests := []struct {
		name     string
		original *FdsRiskAssessment
		update   *FdsRiskAssessment
		expected *FdsRiskAssessment
	}{
		{
			name: "update all fields",
			original: &FdsRiskAssessment{
				Score:          decimal.NewFromInt(10),
				Level:          "low",
				Recommendation: "Approve",
				Status:         "PASSED",
				EvaluatedAt:    time.Time{},
				IsFraud:        nil,
			},
			update: &FdsRiskAssessment{
				Score:          score,
				Level:          "high",
				Recommendation: "Reject",
				Status:         "REJECTED",
				EvaluatedAt:    now,
				IsFraud:        util.ValueToPtr(true),
			},
			expected: &FdsRiskAssessment{
				Score:          score,
				Level:          "high",
				Recommendation: "Reject",
				Status:         "REJECTED",
				EvaluatedAt:    now,
				IsFraud:        util.ValueToPtr(true),
			},
		},
		{
			name: "update only score and level - other fields remain unchanged",
			original: &FdsRiskAssessment{
				Score:          decimal.NewFromInt(10),
				Level:          "low",
				Recommendation: "Approve",
				Status:         "PASSED",
				EvaluatedAt:    now,
				IsFraud:        util.ValueToPtr(false),
			},
			update: &FdsRiskAssessment{
				Score: score,
				Level: "medium",
			},
			expected: &FdsRiskAssessment{
				Score:          score,
				Level:          "medium",
				Recommendation: "Approve",
				Status:         "PASSED",
				EvaluatedAt:    now,
				IsFraud:        util.ValueToPtr(false),
			},
		},
		{
			name: "no fields updated - zero values skipped",
			original: &FdsRiskAssessment{
				Score:          decimal.NewFromInt(50),
				Level:          "medium",
				Recommendation: "Approve",
				Status:         "PASSED",
				EvaluatedAt:    now,
				IsFraud:        util.ValueToPtr(false),
			},
			update: &FdsRiskAssessment{},
			expected: &FdsRiskAssessment{
				Score:          decimal.NewFromInt(50),
				Level:          "medium",
				Recommendation: "Approve",
				Status:         "PASSED",
				EvaluatedAt:    now,
				IsFraud:        util.ValueToPtr(false),
			},
		},
		{
			name: "update chargeback status",
			original: &FdsRiskAssessment{
				Score:            decimal.NewFromInt(50),
				Level:            "medium",
				ChargebackStatus: "",
			},
			update: &FdsRiskAssessment{
				ChargebackStatus: "opened",
			},
			expected: &FdsRiskAssessment{
				Score:            decimal.NewFromInt(50),
				Level:            "medium",
				ChargebackStatus: "opened",
			},
		},
		{
			name: "update IsFraud to nil does not overwrite existing value",
			original: &FdsRiskAssessment{
				IsFraud: util.ValueToPtr(true),
			},
			update: &FdsRiskAssessment{
				IsFraud: nil,
			},
			expected: &FdsRiskAssessment{
				IsFraud: util.ValueToPtr(true),
			},
		},
		{
			name: "update IsFraud to false",
			original: &FdsRiskAssessment{
				IsFraud: util.ValueToPtr(true),
			},
			update: &FdsRiskAssessment{
				IsFraud: util.ValueToPtr(false),
			},
			expected: &FdsRiskAssessment{
				IsFraud: util.ValueToPtr(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.original.Update(tt.update)
			assert.Equal(t, tt.expected, tt.original)
		})
	}
}
