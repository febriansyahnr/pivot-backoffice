package fdsservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

func MapChannelToFraudNetPayment(channel string) string {
	switch channel {
	case "VIRTUAL_ACCOUNT", "BANK_TRANSFER", "MANUAL_TRANSFER", "BALANCE_TRANSFER", "TRANSFER":
		return constant.FRAUD_NET_PAYMENT_DIRECT_DEPOSIT
	case "CREDIT_CARD", "CARD":
		return constant.FRAUD_NET_PAYMENT_CREDIT_CARD
	case "QRIS", "QR":
		return constant.FRAUD_NET_PAYMENT_EWALLETS
	case "PPOB", "BILL":
		return constant.FRAUD_NET_PAYMENT_EPAYMENT
	case "BALANCE_ADJUSTMENT":
		return constant.FRAUD_NET_PAYMENT_INTERNAL_TRANSFER
	case "MANUAL_ACTION":
		return constant.FRAUD_NET_PAYMENT_OTHER
	case "MERCHANT_PAYMENT":
		return constant.FRAUD_NET_PAYMENT_GOODS_SERVICE
	case "TOP_UP":
		return constant.FRAUD_NET_PAYMENT_CASH
	case "ALTO":
		return constant.FRAUD_NET_PAYMENT_ATM
	case "XB":
		return constant.FRAUD_NET_PAYMENT_THIRD_PARTY_PROCESSOR
	default:
		return constant.FRAUD_NET_PAYMENT_OTHER
	}
}

func MapCardTypeToFraudNet(cardType string) string {
	switch strings.ToUpper(cardType) {
	case "MASTERCARD":
		return constant.FRAUD_NET_CARD_TYPE_MC
	case "VISA":
		return constant.FRAUD_NET_CARD_TYPE_VISA
	case "AMEX", "AMERICAN_EXPRESS":
		return constant.FRAUD_NET_CARD_TYPE_AMEX
	case "DISCOVER":
		return constant.FRAUD_NET_CARD_TYPE_DISCOVER
	case "DINERS", "DINERS_CLUB":
		return constant.FRAUD_NET_CARD_TYPE_DINERS_CLUB
	default:
		return constant.FRAUD_NET_CARD_TYPE_OTHER
	}
}

func MapTransactionStatusToFraudNet(status string) string {
	switch strings.ToUpper(status) {
	case "SUCCESS":
		return constant.FRAUD_NET_TRX_STATUS_FULFILLED
	case "FAILED":
		return constant.FRAUD_NET_TRX_STATUS_CANCELLED
	case "PENDING":
		return constant.FRAUD_NET_TRX_STATUS_QUEUED
	default:
		return constant.FRAUD_NET_TRX_STATUS_NEW
	}
}

func MapPaymentStatusToFraudNetStatus(status string) string {
	switch strings.ToUpper(status) {
	case "PAID", "SUCCESS":
		return constant.FRAUD_NET_PAYMENT_STATUS_PAID
	case "REFUNDED":
		return constant.FRAUD_NET_PAYMENT_STATUS_REFUNDED
	case "PENDING", "WAITING_FOR_PAYMENT":
		return constant.FRAUD_NET_PAYMENT_STATUS_AUTH
	case "REQUIRE_CONFIRMATION":
		return constant.FRAUD_NET_PAYMENT_STATUS_INVOICED
	case "REQUIRE_ACTION":
		return constant.FRAUD_NET_PAYMENT_STATUS_PARTIAL_DEFAULT
	case "FAILED", "BLOCKED", "CANCELLED":
		return constant.FRAUD_NET_PAYMENT_STATUS_DECLINED
	case "EXPIRED":
		return constant.FRAUD_NET_PAYMENT_STATUS_DEFAULT
	case "VOID":
		return constant.FRAUD_NET_PAYMENT_STATUS_VOID
	default:
		return constant.FRAUD_NET_PAYMENT_STATUS_AUTH
	}
}

func MapAcquirerResponseCodeToFraudNetCardStatus(responseCode string) string {
	switch responseCode {
	case "01", "03", "05", "06", "12", "13", "22", "40", "57", "61", "62", "63", "64", "65", "6P", "70", "82", "92", "93", "100", "109", "110", "115", "N7", "51", "116", "121", "19", "80", "90", "91", "96", "911":
		return constant.FRAUD_NET_CARD_STATUS_DECLINE
	case "54", "101":
		return constant.FRAUD_NET_CARD_STATUS_EXPIRED
	case "14", "15", "21", "46", "52", "53", "78", "79", "111":
		return constant.FRAUD_NET_CARD_STATUS_INACTIVE
	case "04", "07", "41", "43", "200":
		return constant.FRAUD_NET_CARD_STATUS_STOLEN
	case "34", "59", "83":
		return constant.FRAUD_NET_CARD_STATUS_SUSPENDED
	default:
		return constant.FRAUD_NET_CARD_STATUS_DECLINE
	}
}

func (s *FdsService) SendFdsSlackAlert(ctx context.Context, transactionID string, cardID string, payment *paymentModel.Payment, merchant *merchant.Merchant, resp fdscommon.CheckTransactionResponse) error {
	if len(*resp.EvalResults) == 0 {
		return nil
	}

	var fields []*slackPb.AttachmentField

	fields = append(fields,
		&slackPb.AttachmentField{Title: "Transaction ID", Value: transactionID, Short: true},
		&slackPb.AttachmentField{Title: "Card ID", Value: cardID, Short: true},
		&slackPb.AttachmentField{Title: "Merchant ID", Value: merchant.UUID, Short: true},
		&slackPb.AttachmentField{Title: "Merchant Name", Value: merchant.Name, Short: true},
		&slackPb.AttachmentField{Title: "Amount", Value: fmt.Sprintf("%s %s", payment.Amount.String(), payment.Currency), Short: true},
		&slackPb.AttachmentField{Title: "Risk Score", Value: resp.Score.String(), Short: true},
	)

	// EvalResults is based from FDS Check from each provider
	// if there are more that one provider it needs to
	// display each risk score from the given provider
	for _, result := range *resp.EvalResults {
		if result.RuleEvaluation != nil && result.RuleEvaluation.Provider != "" {
			fields = append(fields, &slackPb.AttachmentField{
				Title: fmt.Sprintf("%s Provider", strings.ToUpper(result.RuleEvaluation.Provider)),
				Value: fmt.Sprintf("Weight: %s | Score: %s | Result: %s", result.Weight.String(), result.RuleEvaluation.Score.String(), result.RuleEvaluation.Result),
				Short: false,
			})

			// FDS Check might trigger more than one rule from the provider
			// alert needs to display each of the rule triggered
			if result.Data.Tags != nil {
				ruleNames := s.extractRuleNamesFromTags(result.Data.Tags)

				// make sure there are rules for the field
				if len(ruleNames) > 0 {
					// iterate through each rules
					for i, ruleName := range ruleNames {
						fields = append(fields, &slackPb.AttachmentField{
							Title: fmt.Sprintf("Rule %d: %s", i+1, ruleName),
							Value: "Triggered",
							Short: true,
						})
					}
				}
			}
		}
	}

	slackMsg := &slackPb.PostWebhookCmd{
		Color:  slackPb.Color_GOOD,
		URL:    s.cfg.SlackConfig.FDSAlertWebhookURL,
		Title:  "FDS Alert - Transaction REJECTED :alert:",
		Fields: fields,
	}

	msgBytes, err := proto.Marshal(slackMsg)
	if err != nil {
		s.logger.Error(ctx, "failed to marshal slack message", logger.Error(err))
		return err
	}

	return s.rabbitMq.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, msgBytes)
}

func (s *FdsService) extractRuleNamesFromTags(tags []fdscommon.CheckTags) []string {
	var ruleNames []string

	for _, tag := range tags {
		if tag.Source == "rule" && tag.Type == "rule" {
			var ruleName string

			if tag.State != nil && *tag.State != "" {
				ruleName = *tag.State
				if tag.Name != "" {
					ruleName = fmt.Sprintf("%s %s", *tag.State, tag.Name)
				}
			} else if tag.Name != "" {
				ruleName = tag.Name
			}

			if ruleName != "" {
				ruleNames = append(ruleNames, ruleName)
			}
		}
	}

	return ruleNames
}

// saveFdsRiskAssessmentToLedger saves the FDS risk assessment results to the ledger's additional_info field
func (s *FdsService) saveFdsRiskAssessmentToLedger(ctx context.Context, trx interface{}, fdsResp *fdscommon.CheckTransactionResponse, evalResults []fdscommon.EvalResult) error {
	// Type assert to AccountTransactionWithUseCase from orchestrator model
	accountTrx, ok := trx.(*orchestrator_model.AccountTransactionWithUseCase)
	if !ok {
		s.logger.Error(ctx, "failed to cast transaction to AccountTransactionWithUseCase type")
		return nil // Skip if type assertion fails
	}

	// Create risk assessment data
	riskAssessment := &fdscommon.FdsRiskAssessment{
		Score:          fdsResp.Score,
		Status:         fdsResp.Status,
		EvaluatedAt:    time.Now().UTC(),
		Recommendation: "Reject", // Default to Reject
	}

	// Determine recommendation and level from the latest successful evaluation
	if fdsResp.Status == constant.FDS_STATUS_PASSED {
		riskAssessment.Recommendation = "Approve"
	} else if fdsResp.Status == constant.FDS_STATUS_REVIEW {
		riskAssessment.Recommendation = util.ToTitle(constant.FDS_STATUS_REVIEW)
	}

	// Extract level from the most recent successful rule evaluation
	for i := len(evalResults) - 1; i >= 0; i-- {
		if evalResults[i].Success && evalResults[i].RuleEvaluation != nil {
			riskAssessment.Level = evalResults[i].RuleEvaluation.Result
			break
		}
	}

	// If no level found, set a default based on status
	if riskAssessment.Level == "" {
		if fdsResp.Status == constant.FDS_STATUS_PASSED {
			riskAssessment.Level = "low"
		} else {
			riskAssessment.Level = "high"
		}
	}

	// Get existing additional_info or create new one
	additionalInfo := make(map[string]interface{})
	if accountTrx.AdditionalInfo.Valid && len(accountTrx.AdditionalInfo.JSONText) > 0 {
		// Parse existing additional_info
		if err := json.Unmarshal(accountTrx.AdditionalInfo.JSONText, &additionalInfo); err != nil {
			s.logger.Error(ctx, "failed to unmarshal existing additional_info", logger.Error(err))
			// additionalInfo is already initialized above, so no need to reassign
		}
	}

	// Add FDS risk assessment to additional_info
	additionalInfo["fdsRiskAssessment"] = riskAssessment

	switch fdsResp.Status {
	case constant.FDS_STATUS_REJECTED:
		additionalInfo["failureCode"] = constant.FailureCodeBlockedByFDS
	case constant.FDS_STATUS_REVIEW:
		additionalInfo["failureCode"] = constant.FailureCodeRequireReview
	}

	// Convert to JSON for storage
	additionalInfoBytes, err := json.Marshal(additionalInfo)
	if err != nil {
		s.logger.Error(ctx, "failed to marshal additional_info", logger.Error(err))
		return err
	}

	// Update the transaction in the repository using the existing method
	nullJSONText := types.NullJSONText{
		JSONText: additionalInfoBytes,
		Valid:    true,
	}

	return s.accountTransactionsRepository.UpdateAdditionalInfoByID(ctx, accountTrx.UUID.String(), nullJSONText)
}
