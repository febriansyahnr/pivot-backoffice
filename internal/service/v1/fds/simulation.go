package fdsservice

import (
	"context"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"

	"github.com/shopspring/decimal"
)

func (s *FdsService) ModifyResponseForCardSimulation(ctx context.Context, resp *fdscommon.CheckTransactionResponse) {
	number, ok := ctx.Value(constant.CtxTestCardNumber).(string)
	if !ok || number == "" {
		return
	}

	var (
		status, ruleEvalResult string
		score                  float64
		useCardSimulation      = false
	)
	if slices.Contains(s.cfg.PaymentSimulationConfig.Cards.BlockedByFds, number) {
		// Blocked By FDS
		useCardSimulation, status, score, ruleEvalResult = true, constant.FDS_STATUS_REJECTED, 99, "very high"

	} else if slices.Contains(s.cfg.PaymentSimulationConfig.Cards.RequireReviewByFds, number) {
		// Require Review
		useCardSimulation, status, score, ruleEvalResult = true, constant.FDS_STATUS_REVIEW, 85, "high"
	}

	if !useCardSimulation {
		return
	}
	resp.Status = status
	resp.Score = decimal.NewFromFloat(score)

	if resp.EvalResults == nil || len(*resp.EvalResults) == 0 {
		return
	}

	for i := range *resp.EvalResults {
		if (*resp.EvalResults)[i].RuleEvaluation == nil {
			continue
		}
		(*resp.EvalResults)[i].RuleEvaluation.Score = resp.Score
		(*resp.EvalResults)[i].RuleEvaluation.Result = ruleEvalResult
		(*resp.EvalResults)[i].RuleEvaluation.Reason = "simulation"
	}
}
