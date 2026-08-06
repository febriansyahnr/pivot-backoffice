package xbPayoutService

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	statusHistoryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/shopspring/decimal"
)

func (s *xbPayoutService) UpdateStatusFromProcessor(ctx context.Context, request *xbModel.ConsumePayoutStatusChangeRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/UpdateStatusFromProcessor")
	defer segment.End()

	var currentTransactionStatus, currentFeeStatus string

	s.logger.Info(ctx, "Process XB status change", logger.Any("request", request))

	disbursement, err := s.disbursementRepo.FindByProcessorReferenceID(ctx, request.AcquirerTransactionId)
	if err != nil {
		s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to find disbursement by processorReferenceID", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	} else if disbursement == nil {
		s.logger.Error(ctx, "UpdateStatusFromProcessor - Disbursement not found by processorReferenceID", logger.Any("request", request))
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrPayoutIsNotFound)
	}

	disbursementStatus, disbursementReasonType, transactionStatus := constant.MapXbProcessorStatusToCoreStatus(request.Status)
	disbursementReasonDesc := constant.MapXbReasonTypeToDesc(disbursementReasonType)

	// Send slack notification
	if !s.isBeneficiaryNoSendSlackAlert(ctx, disbursement.MetadataObj.XbDetail.BeneficiaryData.Name) {
		slackMessage := &slackPb.PostWebhookCmd{
			URL:   s.config.SlackConfig.XBPayoutStatusUpdateWebhookURL,
			Color: slackPb.Color_GOOD,
			Title: "<!subteam^S074BUU9W3A> :earth_asia: XB Transaction Alert :earth_asia:",
			Fields: []*slackPb.AttachmentField{
				{Title: "Status Update", Value: fmt.Sprintf("Payout %s", util.ToTitle(request.Status)), Short: true},
				{Title: "Description", Value: fmt.Sprintf("%s - %s", util.ToTitle(disbursementReasonType), disbursementReasonDesc), Short: false},
				{Title: "Transaction ID", Value: disbursement.UUID, Short: true},
				{Title: "Reference ID", Value: request.PartnerTransactionId, Short: true},
				{Title: "Beneficiary Name", Value: disbursement.MetadataObj.XbDetail.BeneficiaryData.Name, Short: true},
				{Title: "Country", Value: disbursement.MetadataObj.XbDetail.BeneficiaryData.CountryName, Short: true},
				{Title: "Currency", Value: disbursement.MetadataObj.XbDetail.DestinationCurrency, Short: true},
				{Title: "Amount", Value: disbursement.MetadataObj.XbDetail.DestinationAmount.String(), Short: true},
				{Title: "Transaction Time", Value: util.SnapCompatible(request.Timestamp), Short: false},
			},
		}

		rawSlackMessage, _ := proto.Marshal(slackMessage)
		_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawSlackMessage)
	}

	accountTransaction, err := s.orchestratorSvc.FindByReference(ctx, disbursement.UUID, constant.TypeDisbursement)
	if err != nil {
		s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to find account transaction", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if accountTransaction != nil && slices.Contains([]string{constant.StatusSuccess, constant.StatusFailed}, accountTransaction.Status) {
		if request.Status != constant.XbStatusReturned {
			s.logger.Error(ctx, "UpdateStatusFromProcessor - Transaction already in final status", logger.Any("request", request))
			return pkgErrors.New(response.HttpErrRequest, constant.ErrTransactionAlreadyInFinalStatus)
		}
		s.logger.Info(ctx, "UpdateStatusFromProcessor - Allowing RETURNED status to update final transaction", logger.Any("request", request))
	}

	accountTransactionFee, err := s.orchestratorSvc.FindByReference(ctx, disbursement.UUID, constant.TypeFee)
	if err != nil {
		s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to find account transaction fee", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Save current statuses BEFORE any modification (for RETURNED handling)
	// Handle accountTransaction and accountTransactionFee independently
	if request.Status == constant.XbStatusReturned {
		if accountTransaction != nil {
			currentTransactionStatus = accountTransaction.Status
		}
		if accountTransactionFee != nil {
			currentFeeStatus = accountTransactionFee.Status
		}
	}

	ctxTrx, errTrx := s.disbursementRepo.BeginTransaction(ctx)
	if errTrx != nil {
		s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to begin transaction", logger.Error(errTrx))
		return errTrx
	}
	isCompleted := false
	defer func() {
		if !isCompleted {
			if errRollback := s.disbursementRepo.RollbackTransaction(ctxTrx); errRollback != nil {
				s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to rollback transaction", logger.Error(errRollback))
			}
		}
	}()

	if errUpd := s.disbursementRepo.UpdateStatusAndReasonByID(ctxTrx, disbursement.UUID, disbursementStatus, &disbursementReasonType, &disbursementReasonDesc); errUpd != nil {
		s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to update disbursement status", logger.Error(errUpd))
		return pkgErrors.New(response.HttpErrDatabase, errUpd)
	}

	// Update account transaction
	// For RETURNED after SUCCESS, only update reason_type without changing status or updated_at
	if accountTransaction != nil {
		isReturnedAfterSuccess := request.Status == constant.XbStatusReturned &&
			currentTransactionStatus == constant.StatusSuccess

		if isReturnedAfterSuccess {
			// Update reason_type to REFUNDED without changing status or updated_at
			if errUpd := s.orchestratorSvc.UpdateReasonOnly(ctxTrx, accountTransaction.UUID.String(), &disbursementReasonType, &disbursementReasonDesc); errUpd != nil {
				s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to update account transaction reason", logger.Error(errUpd))
				return pkgErrors.New(response.HttpErrDatabase, errUpd)
			}
			s.logger.Info(ctx, "UpdateStatusFromProcessor - Updated reason_type to REFUNDED without changing updated_at",
				logger.String("disbursementID", disbursement.UUID),
				logger.String("currentStatus", currentTransactionStatus))
		} else if accountTransaction.Status != constant.StatusSuccess {
			if errUpd := s.orchestratorSvc.UpdateStatusAccountTransaction(ctxTrx, accountTransaction.UUID.String(), transactionStatus, &disbursementReasonType, &disbursementReasonDesc); errUpd != nil {
				s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to update account transaction status", logger.Error(errUpd))
				return pkgErrors.New(response.HttpErrDatabase, errUpd)
			}
		}
	}

	// Update account transaction fee
	// Same logic: for RETURNED after SUCCESS, only update reason_type
	if accountTransactionFee != nil {
		isReturnedAfterSuccess := request.Status == constant.XbStatusReturned &&
			currentFeeStatus == constant.StatusSuccess

		if isReturnedAfterSuccess {
			// Update reason_type to REFUNDED without changing status or updated_at
			if errUpd := s.orchestratorSvc.UpdateReasonOnly(ctxTrx, accountTransactionFee.UUID.String(), &disbursementReasonType, &disbursementReasonDesc); errUpd != nil {
				s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to update account transaction fee reason", logger.Error(errUpd))
				return pkgErrors.New(response.HttpErrDatabase, errUpd)
			}
			s.logger.Info(ctx, "UpdateStatusFromProcessor - Updated fee reason_type to REFUNDED without changing updated_at",
				logger.String("disbursementID", disbursement.UUID),
				logger.String("currentFeeStatus", currentFeeStatus))
		} else if accountTransactionFee.Status != constant.StatusSuccess {
			if errUpd := s.orchestratorSvc.UpdateStatusAccountTransaction(ctxTrx, accountTransactionFee.UUID.String(), transactionStatus, &disbursementReasonType, &disbursementReasonDesc); errUpd != nil {
				s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to update account transaction fee status", logger.Error(errUpd))
				return pkgErrors.New(response.HttpErrDatabase, errUpd)
			}
		}
	}

	// Record status history
	_ = s.RecordStatusHistory(ctx, &statusHistoryModel.RecordDisbursementStatusHistoryRequest{
		DisbursementID: disbursement.UUID,
		Status:         request.Status,
		Actor:          constant.UserSystemType,
	})

	if errTrx = s.disbursementRepo.CommitTransaction(ctxTrx); errTrx != nil {
		s.logger.Error(ctx, "UpdateStatusFromProcessor - Failed to commit transaction", logger.Error(errTrx))
		return errTrx
	}
	isCompleted = true

	s.sendCallbackToClient(ctx, disbursement.UUID)

	return nil
}

func (s *xbPayoutService) sendCallbackToClient(ctx context.Context, disbursementID string) {
	disbursement, err := s.disbursementRepo.FindByID(ctx, disbursementID)
	if err != nil {
		s.logger.Error(ctx, "sendCallbackToClient - Failed to find disbursement", logger.String("disbursementID", disbursementID), logger.Error(err))
		return
	} else if disbursement == nil {
		s.logger.Error(ctx, "sendCallbackToClient - Disbursement not found", logger.String("disbursementID", disbursementID))
		return
	}

	if util.ValueOfPtr(disbursement.CreatedFrom) != constant.DisbursementCreatedFromOpenApi {
		s.logger.Info(ctx, "skip send callback due to not open api transaction", logger.String("disbursementID", disbursementID))
		return
	}

	// Use stored final fee amount (already calculated with FX conversion during creation)
	finalFeeAmount := disbursement.MetadataObj.FeeDetail.FinalAmount
	fee := decimal.NewFromFloat(finalFeeAmount)

	sourceAmount := disbursement.MetadataObj.XbDetail.SourceAmount
	totalAmount := disbursement.MetadataObj.XbDetail.TotalAmount

	status := ""
	if disbursement.ReasonType != nil {
		status = *disbursement.ReasonType
	}
	statusDescription := ""
	if disbursement.ReasonDescription != nil {
		statusDescription = *disbursement.ReasonDescription
	}
	remark := ""
	if disbursement.Remark != nil {
		remark = *disbursement.Remark
	}

	xbPayoutRequest := &pb.XbPayoutCallbackRequest{
		Uuid:                disbursement.UUID,
		MerchantId:          disbursement.MerchantID,
		ReferenceId:         disbursement.ReferenceID,
		SourceCurrency:      disbursement.MetadataObj.XbDetail.SourceCurrency,
		DestinationCurrency: disbursement.Currency,
		DestinationAmount:   disbursement.Amount.String(),
		FxRate:              disbursement.MetadataObj.XbDetail.FxRate.String(),
		DestinationFxRate:   disbursement.MetadataObj.XbDetail.DestinationFxRate.String(),
		SourceAmount:        sourceAmount.String(),
		Fee:                 fee.String(),
		TotalAmount:         totalAmount.String(),
		CreatedAt:           timestamppb.New(disbursement.CreatedAt),
		UpdatedAt:           timestamppb.New(disbursement.UpdatedAt),
		ExpiredAt:           timestamppb.New(disbursement.MetadataObj.XbDetail.ExpiredAt),
		Status:              status,
		StatusDescription:   statusDescription,
		PurposeCode:         disbursement.MetadataObj.XbDetail.PurposeCode,
		Remark:              remark,
		SenderData:          disbursement.MetadataObj.XbDetail.SenderData.ToProtoSenderDataCallback(),
		BeneficiaryId:       disbursement.MetadataObj.XbDetail.BeneficiaryId,
		BeneficiaryData:     disbursement.MetadataObj.XbDetail.BeneficiaryData.ToProtoBeneficiaryXbDataCallback(),
		RoutingCode:         disbursement.MetadataObj.XbDetail.RoutingCode,
		RoutingValue:        disbursement.MetadataObj.XbDetail.RoutingValue,
	}

	anyWrapper, err := anypb.New(xbPayoutRequest)
	if err != nil {
		s.logger.Error(ctx, "sendCallbackToClient - Failed to generate anypb.New", logger.Error(err))
		return
	}

	callbackRequest := &pb.ProcessCallbackRequest{
		Name:       constant.CallbackNameXB,
		Event:      constant.CallbackNameXB + "." + status,
		MerchantId: disbursement.MerchantID,
		Request:    anyWrapper,
	}

	if errPublish := s.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest); errPublish != nil {
		s.logger.Error(ctx, "sendCallbackToClient - Failed to publish callback message", logger.Error(errPublish))
	}
}

func (s *xbPayoutService) isBeneficiaryNoSendSlackAlert(ctx context.Context, beneficiaryName string) bool {
	beneff := ffcontext.NewEvaluationContext(strings.ToUpper(beneficiaryName))
	getFF, err := ffclient.BoolVariation("backend-portal-xb-beneficiary-no-send-slack-alert", beneff, false)
	if err != nil {
		return false
	}

	return getFF
}
