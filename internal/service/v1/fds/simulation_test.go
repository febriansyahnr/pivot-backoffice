package fdsservice

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	ruleevaluationsmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/ruleEvaluations"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestModifyResponseForCardSimulation(t *testing.T) {
	ctx := context.Background()
	ctxWithCardBlocked := context.WithValue(ctx, constant.CtxTestCardNumber, "4440000334400008")      // NOSONAR
	ctxWithTrxRequireReview := context.WithValue(ctx, constant.CtxTestCardNumber, "4440000334400009") // NOSONAR
	ctxWithOtherCardNumber := context.WithValue(ctx, constant.CtxTestCardNumber, "5550000112200001")  // NOSONAR

	notModifyResponse := &fdscommon.CheckTransactionResponse{
		Status: constant.FDS_STATUS_PASSED,
		Score:  decimal.NewFromInt(10), // NOSONAR
		EvalResults: &[]fdscommon.EvalResult{
			{
				Success: true,
				RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
					UUID:        "2e614593-7919-44dc-97d2-89fee4100384",
					ReferenceID: "878577d0-9973-4e83-8e71-dfb93dd1297e",
					RuleID:      "40a604d9-1b9d-4c58-a984-98fc57a20ee8",
					Result:      "low",
					Score:       decimal.NewFromInt(10), // NOSONAR
					Reason:      "ABC",
				},
			},
		},
	}
	tests := []struct {
		name         string
		ctx          context.Context
		response     *fdscommon.CheckTransactionResponse
		wantResponse *fdscommon.CheckTransactionResponse
	}{
		{
			name: "Not Payment Simulation",
			ctx:  ctx,
		},
		{
			name: "Blocked Transaction By FDS",
			ctx:  ctxWithCardBlocked,
			response: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(10), // NOSONAR
			},
			wantResponse: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_REJECTED,
				Score:  decimal.NewFromFloat(99), // NOSONAR
			},
		},
		{
			name: "Transaction Require Review",
			ctx:  ctxWithTrxRequireReview,
			response: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(10), // NOSONAR
			},
			wantResponse: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_REVIEW,
				Score:  decimal.NewFromFloat(85), // NOSONAR
			},
		},
		{
			name: "Transaction Require Review With Empty Rule Evaluation",
			ctx:  ctxWithTrxRequireReview,
			response: &fdscommon.CheckTransactionResponse{
				Status:      constant.FDS_STATUS_PASSED,
				Score:       decimal.NewFromInt(10), // NOSONAR
				EvalResults: &[]fdscommon.EvalResult{{Success: true}},
			},
			wantResponse: &fdscommon.CheckTransactionResponse{
				Status:      constant.FDS_STATUS_REVIEW,
				Score:       decimal.NewFromFloat(85), // NOSONAR
				EvalResults: &[]fdscommon.EvalResult{{Success: true}},
			},
		},
		{
			name: "Blocked By FDS With Evaluation Rule",
			ctx:  ctxWithCardBlocked,
			response: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_PASSED,
				Score:  decimal.NewFromInt(10), // NOSONAR
				EvalResults: &[]fdscommon.EvalResult{
					{
						Success:        true,
						RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{},
					},
				},
			},
			wantResponse: &fdscommon.CheckTransactionResponse{
				Status: constant.FDS_STATUS_REJECTED,
				Score:  decimal.NewFromFloat(99), // NOSONAR
				EvalResults: &[]fdscommon.EvalResult{
					{
						Success: true,
						RuleEvaluation: &ruleevaluationsmodel.RuleEvaluations{
							Score:  decimal.NewFromFloat(99), // NOSONAR
							Result: "very high",              // NOSONAR
							Reason: "simulation",
						},
					},
				},
			},
		},
		{
			name:         "Not Modify Response",
			ctx:          ctxWithOtherCardNumber,
			response:     notModifyResponse,
			wantResponse: notModifyResponse,
		},
	}

	service := &FdsService{
		cfg: &config.Config{PaymentSimulationConfig: config.PaymentSimulationConfig{
			Cards: config.PaymentSimulationCardConfig{
				BlockedByFds:       []string{"4440000334400008", "5550000334400008", "3330000334400008"},
				RequireReviewByFds: []string{"4440000334400009", "5550000334400009", "3330000334400009"},
			},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			service.ModifyResponseForCardSimulation(test.ctx, test.response)

			assert.Equal(t, test.wantResponse, test.response)
		})
	}
}
