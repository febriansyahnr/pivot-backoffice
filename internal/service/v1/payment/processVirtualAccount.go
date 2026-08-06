package paymentService

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
)

func (s *PaymentService) ProcessVirtualAccountPayment(ctx context.Context, request *paymentModel.VirtualAccountPaymentNotificationRequest) (err error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/ProcessVirtualAccountPayment")
	defer segment.End()

	s.logger.Info(ctx, "Process VA payment", logger.Any("request", request))

	// If not paid status then no need to create orchestrator transaction & send callback
	if strings.ToUpper(request.Status) != paymentConstant.VirtualAccountStatusPaid {
		return nil
	}

	ctxTrx, errCtx := s.paymentRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := s.paymentRepo.RollbackTransaction(ctxTrx); e != nil {
				s.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErrors.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	// If VA number found in payment then process payment
	payment, err := s.GetAndUpdateVirtualAccountPayment(ctxTrx, request)
	if err != nil {
		return err
	}

	vaType := paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC
	// marshal metadata to json
	if payment.Metadata != nil {
		snapCoreResp := snapCoreModel.CreateVirtualAccountResponseData{}
		jsonData, errMarshal := json.Marshal(payment.Metadata)
		if errMarshal != nil {
			s.logger.Error(ctx, "error when marshal payment metadata", logger.Error(errMarshal))
			return pkgErrors.New(response.HttpErrInternal, errMarshal)
		}

		// unmarshal metadata to snapCoreResp
		json.Unmarshal(jsonData, &struct {
			SnapCore interface{} `json:"snapCore"`
		}{
			SnapCore: &snapCoreResp,
		})

		vaType = snapCoreModel.FindVaTrxTypeByCriteria(snapCoreResp.IsClosedAmount, snapCoreResp.IsSingleUse)
	}

	if vaType == paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC {
		lockKey := fmt.Sprintf(constant.PaymentNotificationLockCacheKey, payment.UUID)
		s.logger.Info(ctx, "acquire lock payment notification", logger.String("key", lockKey))
		mutex := s.redis.NewMutex(lockKey, redsync.WithExpiry(constant.PaymentNotificationLockTTL))
		err = mutex.LockContext(ctxTrx)
		if err != nil {
			s.logger.Error(ctx, "error when acquire lock payment notification", logger.Error(err), logger.String("key", lockKey))
			return pkgErrors.New(response.HttpErrInternal, err)
		}
		s.logger.Info(ctx, "not release lock payment notification, because there isn't any trx state validation. let the lock expire naturally", logger.String("key", lockKey))
	}

	// Record payment processing start now that we have payment ID
	s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistoryProcessing)
	if err = s.internal.DeterminePaymentFee(&ctxTrx, payment); err != nil {
		return err
	}

	if !request.TrxDatetime.IsZero() {
		payment.TrxDatetime = &request.TrxDatetime
	}
	payment.ProcessorTransactionID = request.ProcessorTransactionID
	payment.Processor = request.Processor
	payment.ProcessorID = request.ProcessorID
	// Reference for va recon can only use va number and transaction timestamp
	payment.ReconReferenceNo = request.Number
	payment.TrxDatetime = &request.TrxDatetime
	payment.BankReferenceId = request.BankReferenceId

	// Create ledger transaction on paid VA
	paymentLedger, err := s.accountTransactionRepo.GetTransactionByReferenceIdAndProcessorId(ctxTrx, payment.UUID, request.ProcessorID)
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if paymentLedger == nil || vaType == paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC {
		if err := s.PostCreateLedger(ctxTrx, payment, &paymentModel.PostCreateLedgerRequest{
			Status:  constant.StatusSuccess,
			Channel: constant.ChannelVirtualAccount,
			Amount:  request.PaidAmount,
		}); err != nil {
			return err
		}

	} else {
		updateRequest := orchestrator_model.UpdatePaymentTransactionRequest{
			ProcessorTransactionId: request.ProcessorTransactionID,
			LedgerId:               paymentLedger.UUID.String(),
			UpdatedAt:              payment.UpdatedAt,
			TrxDatetime:            payment.TrxDatetime,
			Status:                 constant.StatusSuccess,
			Channel:                constant.ChannelVirtualAccount,
			Amount:                 request.PaidAmount,
		}
		if paymentLedger.SettlementModel.Valid {
			updateRequest.SettlementModel = util.ValueToPtr(paymentLedger.SettlementModel.String)
		}

		if err := s.UpdatePendingLedger(ctxTrx, payment, updateRequest); err != nil {
			return err
		}
	}

	if errCommit := s.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return pkgErrors.New(response.HttpErrDatabase, errCommit)
	}

	isCompleted = true

	// Record successful payment
	s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistoryPaid)

	// Send callback on paid VA
	s.sendVaCallbackOnPaidStatus(ctx, payment, request.PaidAmount)

	return nil
}

func (s *PaymentService) sendVaCallbackOnPaidStatus(ctx context.Context, payment *paymentModel.Payment, requestAmount commonModel.Amount) {
	var (
		paymentCallbackRequestWrapper *anypb.Any
		paymentCallbackEvent          string
	)

	paymentResult, err := s.FindPaymentById(ctx, payment.UUID, payment.MerchantID)
	if err != nil {
		return
	} else if paymentResult == nil {
		return
	}

	if payment.TrxDatetime != nil {
		paymentResult.TransactionDate = payment.TrxDatetime
	}
	paymentResult.AdditionalInfo = &map[string]any{
		"bankReferenceId": payment.BankReferenceId,
	}

	isSnap := paymentResult.VirtualAccount.IsSnap
	callbackName := constant.CallbackNamePayment
	if isSnap {
		callbackName = constant.CallbackMasterPaymentSNAPVA
	}

	if paymentResult.IsUnifiedPayment {
		paymentCallbackRequestWrapper, err = anypb.New(paymentResult.ToPbUnifiedPaymentCallbackRequest(payment))
		if err != nil {
			s.logger.Error(ctx, "Generate anypb.New ToPbUnifiedPaymentCallbackRequest", logger.Error(err))
			return
		}

		paymentCallbackEvent = fmt.Sprintf(constant.CallbackEventUnifiedPaymentPattern, payment.Status)
	} else {
		paymentRequest := buildCallbackRequest(paymentResult, requestAmount)
		paymentCallbackRequestWrapper, err = anypb.New(paymentRequest)
		if err != nil {
			s.logger.Error(ctx, "Generate anypb.New buildCallbackRequest", logger.Error(err))
			return
		}

		paymentCallbackEvent = constant.CallbackEventPaymentVirtualAccountPaid
	}

	merchantID := payment.MerchantID
	// Send callback to merchant that initiated the transaction
	if payment.CreatedBy != nil {
		merchantID = *payment.CreatedBy
	}

	callbackRequest := &pb.ProcessCallbackRequest{
		Name:       callbackName,
		Event:      paymentCallbackEvent,
		MerchantId: merchantID,
		Request:    paymentCallbackRequestWrapper,
		IsSnap:     isSnap,
	}
	_ = s.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest)

	subject, message := constant.GetNotificationMessage(payment.UUID, paymentResult.Status)

	err = s.rabbitMqExt.PushNotification(ctx, &notification.PushNotification{
		RoutingKey: fmt.Sprintf(constant.NotificationRoutingKeyFmt, payment.UUID),
		Payload: notification.PushNotificationPayload{
			ID:        uuid.NewString(),
			Subject:   subject,
			Type:      constant.CreateVAPaymentNotifType,
			Message:   message,
			CreatedAt: time.Now().UTC(),
			Status:    paymentResult.Status,
		},
	})
	if err != nil {
		s.logger.Error(ctx, "push notification for payment "+payment.UUID, logger.Error(err))
	}
}

func buildCallbackRequest(paymentResult *paymentModel.PaymentResponse, requestAmount commonModel.Amount) *pb.PaymentCallbackRequest {
	if paymentResult.VirtualAccount != nil && paymentResult.VirtualAccount.IsSnap {
		paymentResult.PaidAmount = &requestAmount
	}

	// Send callback on every payment changes
	paymentItems := make([]*pb.PaymentRequestItem, len(*paymentResult.PaymentItems))
	for i, item := range *paymentResult.PaymentItems {
		paymentItems[i] = &pb.PaymentRequestItem{
			ItemId:      item.ItemID,
			Name:        item.Name,
			Description: item.Description,
			Amount:      item.Amount.ProtoAmount(),
			Qty:         float64(item.Qty),
		}
	}
	// Payment callback request
	paymentRequest := &pb.PaymentCallbackRequest{
		Uuid:        paymentResult.UUID,
		MerchantId:  paymentResult.MerchantID,
		ReferenceId: paymentResult.ReferenceID,
		Customer: &pb.PaymentRequestCustomer{
			CustomerId: paymentResult.Customer.CustomerID,
			Name:       paymentResult.Customer.Name,
			Email:      paymentResult.Customer.Email,
			Phone:      paymentResult.Customer.Phone,
			Metadata:   nil,
		},
		Status:          paymentResult.Status,
		PaidAmount:      paymentResult.PaidAmount.ProtoAmount(),
		TotalAmount:     paymentResult.TotalAmount.ProtoAmount(),
		PaymentMethodId: paymentResult.PaymentMethodId,
		PaymentMethod:   paymentResult.PaymentMethod,
		PaymentItems:    paymentItems,
		LastUpdateDate:  timestamppb.New(*paymentResult.LastUpdateDate),
		TransactionDate: timestamppb.New(*paymentResult.TransactionDate),
	}
	if paymentResult.VirtualAccount != nil {
		paymentRequest.VirtualAccount = &pb.PaymentVirtualAccountRequest{
			Issuer:                paymentResult.VirtualAccount.Issuer,
			VirtualAccountTrxType: paymentResult.VirtualAccount.VirtualAccountTrxType,
			VirtualAccountNumber:  paymentResult.VirtualAccount.VirtualAccountNumber,
			VirtualAccountName:    paymentResult.VirtualAccount.VirtualAccountName,
			MinAmount:             paymentResult.VirtualAccount.MinAmount.ProtoAmount(),
			MaxAmount:             paymentResult.VirtualAccount.MaxAmount.ProtoAmount(),
			ExpiredDate:           timestamppb.New(*paymentResult.VirtualAccount.ExpiredDate),
			IsSnap:                paymentResult.VirtualAccount.IsSnap,
			Metadata:              nil,
		}
	}

	if paymentResult.AdditionalInfo != nil {
		pbAdditionalInfo := &anypb.Any{}
		structVal, err := structpb.NewStruct(*paymentResult.AdditionalInfo)
		if err == nil {
			pbAdditionalInfo, _ = anypb.New(structVal)
		}

		paymentRequest.AdditionalInfo = pbAdditionalInfo
	}

	return paymentRequest
}
