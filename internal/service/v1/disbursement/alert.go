package disbursementService

import (
	"context"
	"fmt"
	"sort"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

func (s *DisbursementService) ProcessPayoutAlert(ctx context.Context, request *disbursementModel.PayoutTransactionAlertRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ProcessPayoutAlert")
	defer segment.End()

	transaction, err := s.orchestratorSvc.FindByReference(ctx, request.DisbursementID, constant.TypeDisbursement)
	if err != nil {
		s.logger.Error(ctx, "[ProcessPayoutAlert] Find transaction error", logger.Error(err))
		return err
	} else if transaction == nil {
		s.logger.Warn(ctx, "[ProcessPayoutAlert] Find transaction not found")
		return pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrDataNotFound)
	} else if transaction.Status == constant.StatusSuccess {
		s.logger.Info(ctx, "[ProcessPayoutAlert] Success transaction no need to send alert")
		return nil
	}

	// process to get transfer data from processor
	// if status is not success, need to send alert

	s.sendPayoutTransactionAlert(ctx, request)
	return nil
}

func (s *DisbursementService) sendPayoutTransactionAlert(ctx context.Context, request *disbursementModel.PayoutTransactionAlertRequest) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/sendPayoutTransactionAlert")
	defer segment.End()

	s.logger.Info(ctx, "[sendPayoutTransactionAlert] Start send payout transaction alert", logger.String("disbursementID", request.DisbursementID))
	defer func() {
		s.logger.Info(ctx, "[sendPayoutTransactionAlert] Finish send payout transaction alert", logger.String("disbursementID", request.DisbursementID))
	}()

	disbursement, err := s.disbursementRepo.FindByID(ctx, request.DisbursementID)
	if err != nil {
		s.logger.Error(ctx, "[sendPayoutTransactionAlert] Find disbursement error", logger.Error(err))
		return
	} else if disbursement == nil {
		s.logger.Warn(ctx, "[sendPayoutTransactionAlert] Find disbursement not found")
		return
	}

	transaction, err := s.orchestratorSvc.FindByReference(ctx, disbursement.UUID, constant.TypeDisbursement)
	if err != nil {
		s.logger.Error(ctx, "[sendPayoutTransactionAlert] Find transaction error", logger.Error(err))
		return
	} else if transaction == nil {
		s.logger.Warn(ctx, "[sendPayoutTransactionAlert] Find transaction not found")
		return
	} else if transaction.Status == constant.StatusSuccess {
		s.logger.Info(ctx, "[sendPayoutTransactionAlert] Success transaction no need to send alert")
		return
	}

	transferLogsStatus := ""
	transferLogs, err := s.snapCoreRepo.CheckStatusByExternalId(
		ctx,
		transaction.UUID.String(),
		true, // check bank statement
	)
	if err == nil && transferLogs != nil {
		// build transaction history
		request.History = s.buildTransactionHistoryFromCheckStatus(transferLogs)
		transferLogsStatus = transferLogs.Status
	}

	alertTitle := "Pending Transaction Alert"
	if transaction.Status == constant.StatusFailed {
		alertTitle = "Failed Transaction Alert"
	} else if transferLogsStatus == constant.StatusSuccess {
		// get transaction fee id
		transactionFee, errSuccessTrx := s.orchestratorSvc.FindByReference(ctx, disbursement.UUID, constant.TypeFee)
		if errSuccessTrx == nil {
			// no need to sent alert, but update the payout transaction
			transactionAndFee := &orchestrator_model.TransactionAndFeeObject{
				TransactionID: transaction.UUID.String(),
				FeeID:         transactionFee.UUID.String(),
				MerchantID:    disbursement.MerchantID,
			}

			errSuccessTrx = s.updateTransactionStatusWithHistory(
				ctx,
				transactionAndFee,
				transferLogs.Status,
				&transaction.ReasonType.String,
				&transaction.ReasonDescription.String,
				disbursement.UUID,
			)
		}

		if errSuccessTrx == nil {
			s.logger.Info(ctx, "[sendPayoutTransactionAlert] Success transaction no need to send alert after check status success")
			return
		}

		s.logger.Error(ctx, "[sendPayoutTransactionAlert] Error when update transaction status after check status success", logger.Error(errSuccessTrx))
	}

	remark, bankName := "", ""
	if disbursement.Remark != nil {
		remark = *disbursement.Remark
	}
	if disbursement.BeneficiaryBankName != nil {
		bankName = *disbursement.BeneficiaryBankName
	}

	merchantName := ""
	merchant, err := s.merchantRepo.FindMerchantByID(ctx, disbursement.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "[sendPayoutTransactionAlert] Error when get merchant by id", logger.Error(err))
		return
	} else if merchant != nil {
		merchantName = merchant.Name
	}

	fields := []*slackPb.AttachmentField{
		{Title: "Disbursement ID", Value: disbursement.UUID, Short: true},
		{Title: "Transaction Time", Value: util.SnapCompatible(transaction.TransactionTimestamp), Short: true},
		{Title: "Amount", Value: fmt.Sprintf("Rp %s", util.ConvertFloatToCurrency(disbursement.Amount.InexactFloat64())), Short: true},
		{Title: "Merchant Ref", Value: disbursement.ReferenceID, Short: true},
		{Title: "Merchant Remarks", Value: remark, Short: true},
		{Title: "Bank Ref No", Value: request.BankRefNo, Short: true},
		{Title: "Beneficiary Bank No", Value: disbursement.BeneficiaryAccountNo, Short: true},
		{Title: "Beneficiary Bank Name", Value: bankName, Short: true},
		{Title: "Beneficiary Name", Value: disbursement.BeneficiaryAccountName, Short: true},
		{Title: "Source Account Bank", Value: request.BankProcessor, Short: true},
		{Title: "Transfer Type", Value: request.TransferType, Short: true},
		{Title: "Failure Reason", Value: transaction.ReasonDescription.String, Short: true},
		{Title: "Merchant Name", Value: merchantName, Short: true},
	}

	if len(request.History) > 0 {
		historyText := ""
		for _, history := range request.History {
			historyText += fmt.Sprintf("%d. %s - %s - %s\n",
				history.Order, history.Acquirer, history.Action, history.Status)

			if history.Reason != "" {
				historyText += fmt.Sprintf(" - Reason: %s\n", history.Reason)
			}
		}
		fields = append(fields, &slackPb.AttachmentField{
			Title: "Transfer History",
			Value: historyText,
			Short: false,
		})
	}

	slackMessage := &slackPb.PostWebhookCmd{
		URL:    s.config.SlackConfig.PayoutAlertWebHookURL,
		Color:  slackPb.Color_GOOD,
		Title:  "<!subteam^S095QH85GUE> " + alertTitle,
		Fields: fields,
	}
	rawSlackMessage, _ := proto.Marshal(slackMessage)
	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawSlackMessage)
}

func (s *DisbursementService) buildTransactionHistoryFromCheckStatus(transferLogs *snapCoreModel.BankTransferCheckStatusResponseData) []*disbursementModel.PayoutTransactionAlertHistory {

	var histories []*disbursementModel.PayoutTransactionAlertHistory
	for _, log := range transferLogs.TransferLogs {
		history := &disbursementModel.PayoutTransactionAlertHistory{
			Status:   log.Status,
			Acquirer: log.Bank,
			Order:    log.Order,
			Action:   log.Action,
			Reason:   getPayoutHistoryReason(log.AdditionalInfo),
		}
		histories = append(histories, history)
	}

	// sort by order asc
	sort.Slice(histories, func(i, j int) bool {
		return histories[i].Order < histories[j].Order
	})

	return histories
}

func getPayoutHistoryReason(historyAdditionalInfo *snapCoreModel.TransferLogAdditionalInfo) string {
	if historyAdditionalInfo == nil || historyAdditionalInfo.FailedReason == nil || historyAdditionalInfo.FailedReason.Data == nil {
		return ""
	}

	return historyAdditionalInfo.FailedReason.Data.ResponseMessage
}
