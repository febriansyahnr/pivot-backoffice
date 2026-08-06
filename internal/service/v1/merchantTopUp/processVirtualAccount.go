package merchantTopUp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	common "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchantTopUp"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/proto"
)

func (s *merchantTopUpService) ProcessMerchantTopUpWithVirtualAccount(ctx context.Context, request *paymentModel.VirtualAccountPaymentNotificationRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursementTopUp/ProcessVirtualAccountDisbursementTopUp")
	defer segment.End()

	s.logger.Info(ctx, "Process Merchant Top Up VA", logger.Any("request", request))

	// If not paid status then no need to create orchestrator transaction
	if strings.ToUpper(request.Status) != paymentConstant.VirtualAccountStatusPaid {
		return nil
	}

	// If VA number found in merchant_top_up_references then process top up
	topUp, err := s.FindByReferenceNumber(ctx, request.Number)
	if err != nil {
		return err
	}

	currentBalance, err := s.orchestratorSvc.GetAvailableMerchantBalance(ctx, topUp.MerchantID, topUp.AccountName)
	if err != nil {
		s.logger.Error(ctx, "Failed while getting available merchant balance", logger.Error(err))
		return err
	}

	merchant, err := s.merchantService.FindMerchantByID(ctx, topUp.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "Failed while find merchant by id", logger.Error(err))
		return err

	} else if merchant == nil {
		return constant.ErrMerchantNotFound
	}

	transactionAmount, err := strconv.ParseFloat(request.PaidAmount.Value, 64)
	if err != nil {
		return err
	}

	transactionType := orchestrator_model.TypeMerchantTopUp
	if topUp.AccountName == constant.TypeDisbursement {
		transactionType = orchestrator_model.TypeDisbursementTopUp
	}
	transactionTime := time.Now().UTC()
	if !request.TrxDatetime.IsZero() {
		transactionTime = request.TrxDatetime
	}

	trxMetadata := orchestratorModel.MetadataPayment[any]{
		ProcessorTransactionId: request.ProcessorTransactionID,
		ChargeStatus:           constant.StatusSuccess,
	}

	// get merchant top up fee
	feeRequest := feeModel.GetFeeRequest{
		MerchantID:      merchant.UUID,
		Channel:         request.Acquirer,
		Reference:       constant.ReferenceTopUp,
		ReferenceType:   transactionType,
		PaymentMethod:   paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		ReferenceAmount: transactionAmount,
	}

	merchantTopUpFee, feeDetail, err := s.feeSvc.GetFeeCalculationAndDetail(ctx, &feeRequest)
	if err != nil {
		s.logger.Error(ctx, "Failed while getting merchant top up fee", logger.Error(err),
			logger.String("topUpID", topUp.ID),
			logger.String("merchantID", merchant.UUID),
			logger.String("number", request.Number),
		)
		return err
	}
	trxMetadata.FeeDetail = feeDetail

	transactionRequest := &orchestrator_model.CreateAccountTransactionRequest{
		UUID:                   uuid.New(),
		ReferenceID:            topUp.ID,
		Type:                   transactionType,
		MerchantID:             util.ParseUUID(topUp.MerchantID),
		Currency:               request.PaidAmount.Currency,
		Credit:                 transactionAmount,
		Debit:                  0.00,
		Channel:                constant.ChannelVirtualAccount,
		Status:                 constant.StatusSuccess,
		Remarks:                "",
		TransactionTimestamp:   transactionTime,
		Usecase:                topUp.AccountName,
		Processor:              request.Processor,
		ProcessorID:            request.ProcessorID,
		ProcessorTransactionID: request.ProcessorTransactionID,
	}

	transactionRequest.AdditionalInfo.Valid = true
	transactionRequest.AdditionalInfo.JSONText, _ = json.Marshal(trxMetadata)

	if err = s.orchestratorSvc.CreateAccountTransaction(ctx, transactionRequest); err != nil {
		s.logger.Error(ctx, "Failed while create account transaction", logger.Error(err))
		return err
	}

	if merchantTopUpFee > 0 {
		// only create fee if merchant top up fee is above 0
		feeTransactionRequest := &orchestrator_model.CreateAccountTransactionRequest{
			UUID:                   util.GenerateUUID(),
			ReferenceID:            topUp.ID,
			Type:                   orchestrator_model.TypeFee,
			MerchantID:             util.ParseUUID(topUp.MerchantID),
			Currency:               request.PaidAmount.Currency,
			Credit:                 0.00,
			Debit:                  merchantTopUpFee,
			Channel:                constant.ChannelVirtualAccount,
			Status:                 constant.StatusSuccess,
			Remarks:                "",
			TransactionTimestamp:   transactionTime,
			Usecase:                topUp.AccountName,
			Processor:              request.Processor,
			ProcessorID:            request.ProcessorID,
			ProcessorTransactionID: request.ProcessorTransactionID,
		}
		if err = s.orchestratorSvc.CreateAccountTransaction(ctx, feeTransactionRequest); err != nil {
			s.logger.Error(ctx, "Failed while create fee account transaction", logger.Error(err),
				logger.String("topUpID", topUp.ID),
				logger.String("merchantID", merchant.UUID),
				logger.String("number", request.Number),
				logger.Float64("merchantTopUpFee", merchantTopUpFee),
			)
			return err
		}
	}
	// Send Merchant Callback
	requestCallback := &merchantTopUp.MerchantTopUpCallbackRequest{
		UUID:         transactionRequest.UUID.String(),
		MerchantID:   topUp.MerchantID,
		MerchantName: merchant.Name,
		AccountName:  topUp.AccountName,
		Amount: common.Amount{
			Currency: request.PaidAmount.Currency,
			Value:    fmt.Sprintf("%.2f", transactionAmount),
		},
		BalanceBefore: common.Amount{
			Currency: request.PaidAmount.Currency,
			Value:    fmt.Sprintf("%.2f", currentBalance),
		},
		BalanceAfter: common.Amount{
			Currency: request.PaidAmount.Currency,
			Value:    fmt.Sprintf("%.2f", currentBalance+transactionAmount),
		},
		PaymentMethod: merchantTopUp.MerchantTopUpCallbackPaymentMethodObject{
			Type: constant.ChannelVirtualAccount,
		},
		PaymentMethodOptions: merchantTopUp.MerchantTopUpCallbackPaymentMethodOptionsObject{
			VirtualAccount: &merchantTopUp.MerchantTopUpCallbackPaymentMethodOptionVAObject{
				Channel:              strings.ToUpper(request.Acquirer),
				VirtualAccountNumber: request.Number,
				VirtualAccountName:   merchant.ShortName,
			},
		},
		TransactionTime:  transactionTime,
		ParentMerchantID: merchant.ParentID.String,
	}
	if errCallback := s.internal.SendCallback(ctx, constant.CallbackEventMerchantTopUpSuccess, requestCallback); errCallback != nil {
		s.logger.Warn(ctx, "Failed while send callback", logger.Error(err))
	}

	// Send Slack Notification
	slackMessage := &slackPb.PostWebhookCmd{
		URL:   s.config.SlackConfig.TopUpNotifWebhookURL,
		Color: slackPb.Color_GOOD,
		Title: "<!subteam^S074BUU9W3A> Incoming Top UP",
		Fields: []*slackPb.AttachmentField{
			{Title: "Merchant Name", Value: merchant.Name, Short: true},
			{Title: "Balance Name", Value: topUp.AccountName, Short: true},
			{Title: "Amount", Value: util.FormatRupiah(transactionAmount), Short: true},
			{Title: "Fee", Value: util.FormatRupiah(merchantTopUpFee), Short: true},
			{Title: "Bank", Value: strings.ToUpper(request.Acquirer), Short: true},
			{Title: "VA Number", Value: request.Number, Short: true},
			{Title: "Balance Before", Value: util.FormatRupiah(currentBalance), Short: true},
			{Title: "Balance After", Value: util.FormatRupiah(currentBalance + transactionAmount), Short: true},
			{Title: "Transaction Time", Value: util.SnapCompatible(transactionTime), Short: true},
		},
	}
	rawSlackMessage, _ := proto.Marshal(slackMessage)

	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawSlackMessage)

	return nil
}
