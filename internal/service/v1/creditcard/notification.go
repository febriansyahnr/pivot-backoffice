package creditcard

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/v2/logger"
	"google.golang.org/protobuf/types/known/anypb"
)

func (c *CreditCardService) PaymentNotification(ctx context.Context, request creditcardModel.CardPaymentNotificationRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/creditcard/PaymentNotification")
	defer segment.End()

	initiatorMerchantID := request.Data.MerchantID.String()

	// Ignore refunded status
	if request.Data.PaymentStatus == constant.CreditCardStatusRefunded {
		c.logger.Info(ctx, "[CardNotification] Ignore refunded status")
		return nil
	}

	var isUnifiedPayment = false

	payment, err := c.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, request.Data.MerchantID.String(), request.Data.ReferenceID)
	if err != nil {
		return err
	}

	if payment == nil {
		// Create payment first if not exist
		_, err = c.CreatePayment(ctx, creditcardModel.CreateCardPaymentRequest{
			PaymentUUID:          request.Data.PaymentUUID,
			ReferenceID:          request.Data.ReferenceID,
			BankMerchantID:       request.Data.BankMerchantID,
			Amount:               request.Data.Amount,
			Currency:             request.Data.Currency,
			AuthenticationMethod: request.Data.AuthenticationMethod,
			MerchantID:           request.Data.MerchantID,
			RedirectUrl: creditcardModel.CreditcardRedirectUrlRequest{
				SuccessUrl: request.Data.RedirectUrl.SuccessUrl,
				FailedUrl:  request.Data.RedirectUrl.FailedUrl,
			},
		})
		if err != nil {
			return err
		}

		payment, err = c.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, request.Data.MerchantID.String(), request.Data.ReferenceID)
		if err != nil {
			return err
		}

		if payment == nil {
			err = constant.ErrCreditcardReferenceIdNotFound
			c.logger.Error(ctx, err.Error(), logger.Error(err))
			return err
		}
	}
	if payment.Metadata != nil {
		if isUnifiedPaymentOk, ok := (*payment.Metadata)[constant.IsUnifiedPaymentKey].(bool); ok {
			isUnifiedPayment = isUnifiedPaymentOk
		}

	}
	if payment.CreatedBy != nil {
		initiatorMerchantID = *payment.CreatedBy
	}

	err = creditcardModel.UpdateCreditcardMetaData(payment.Metadata, request.Data.CardData, request.Data.AuthorizationData, request.Data.AuthenticationData, request.Data.PaymentStatus)
	if err != nil {
		c.logger.Error(ctx, constant.ErrWhenUpdateCreditcardMetaData.Error(), logger.Error(err))
		return err
	}

	merchant, err := c.merchantRepo.FindMerchantByID(ctx, payment.MerchantID)
	if err != nil {
		c.logger.Error(ctx, "Failed while find merchant by id", logger.Error(err))
		return err
	}

	channel := "FOREIGN_"
	if merchant.BusinessCountry.String == request.Data.CardData.IssuingCountry {
		channel = "LOCAL_"
	}
	channel += strings.ToUpper(request.Data.CardData.CardBrand)

	payment.PaymentMethod.Acquirer = channel // For calculation payment fee per channel
	if err = c.paymentLedgerSvc.DeterminePaymentFee(&ctx, payment); err != nil {
		c.logger.Error(ctx, "Failed while determine payment fee", logger.Error(err))
		return err
	}

	if err := payment.UpdatePaymentFromCreditcardNotification(&request.Data); err != nil {
		c.logger.Error(ctx, "error when updating payment from creditcard notification", logger.Error(err))
		return err
	}
	payment.Processor = constant.CreditCardCoreProcessor
	payment.ProcessorID = request.Data.AcquirerTransactionID

	ctxTrx, errCtx := c.paymentRepo.BeginTransaction(ctx)
	if errCtx != nil {
		return errCtx
	}

	isCompleted := false
	defer func() {
		if !isCompleted {
			if e := c.paymentRepo.RollbackTransaction(ctxTrx); e != nil {
				c.logger.Error(ctxTrx, "error when execute rollback transaction", logger.Error(pkgErrors.New(response.HttpErrDatabase, e)))
			}
		}
	}()

	// Update payment data
	err = c.paymentRepo.UpdatePaymentData(ctxTrx, payment.ToDTO())
	if err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	if request.Data.PaymentStatus == constant.CreditCardStatusVoid {
		// Record payment void/cancellation status history
		if c.paymentSvc != nil {
			c.paymentSvc.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistoryVoid)
		}

		reasonType := constant.ReasonTypeOtherReason
		reasonDescription := constant.ReasonDescTransactionVoidByProcessor

		// Fee
		accountTransactionFee, err := c.orchestratorSvc.FindByReference(ctxTrx, payment.UUID, constant.TypeFee)
		if err != nil {
			c.logger.Error(ctx, constant.ErrWhenFindAccountTransaction.Error(), logger.Error(err))
			return err
		}

		if accountTransactionFee != nil {
			// Fee metadata
			feemetadataObj := orchestratorModel.FeeTransactionMetadataObject{}
			if accountTransactionFee.AdditionalInfo.Valid {
				_ = json.Unmarshal(accountTransactionFee.AdditionalInfo.JSONText, &feemetadataObj)
			}

			if feemetadataObj.SettlementStatus == constant.StatusPending {
				feemetadataObj.SettlementStatus = constant.SettlementStatusCancelled
				accountTransactionFee.Status = constant.StatusFailed

				feemetadataJson, _ := json.Marshal(feemetadataObj)

				errUpdate := c.orchestratorSvc.UpdateStatusAccountAndAdditionalInfoTransaction(ctxTrx,
					accountTransactionFee.UUID.String(), accountTransactionFee.Status, reasonType, reasonDescription,
					feemetadataJson)
				if errUpdate != nil {
					c.logger.Error(ctx, constant.ErrWhenUpdateAccountTransaction.Error(),
						logger.Error(errUpdate))
					return errUpdate
				}
			} else if feemetadataObj.SettlementStatus == constant.StatusSuccess {
				feemetadataObj.LinkedTransactionId = accountTransactionFee.UUID.String()

				rawReversalMetadata, _ := json.Marshal(feemetadataObj)

				if errLedger := c.orchestratorSvc.PostAccountTransaction(ctxTrx, &orchestratorModel.CreateAccountTransactionRequest{
					UUID:                 uuid.New(),
					ReferenceID:          accountTransactionFee.ReferenceID,
					Type:                 accountTransactionFee.Type,
					MerchantID:           accountTransactionFee.MerchantID,
					Currency:             accountTransactionFee.Currency,
					Credit:               accountTransactionFee.Debit,
					Debit:                0,
					Channel:              constant.ChannelCreditCard,
					Status:               constant.StatusSuccess,
					ReasonType:           &reasonType,
					ReasonDescription:    &reasonDescription,
					TransactionTimestamp: request.Data.Updated,
					Usecase:              orchestratorModel.TypePayment,
					Processor:            constant.CreditCardCoreProcessor,
					ProcessorID:          request.Data.AcquirerTransactionID,
					AdditionalInfo: types.NullJSONText{
						Valid: true, JSONText: rawReversalMetadata,
					},
				}); errLedger != nil {
					return pkgErrors.New(response.HttpErrDatabase, errLedger)
				}
			}
		}

		// Payment
		accountTransactionPayments, err := c.orchestratorSvc.FindByReference(ctxTrx, payment.UUID, constant.TypePayment)
		if err != nil {
			c.logger.Error(ctx, constant.ErrWhenFindAccountTransaction.Error(), logger.Error(err))
			return err
		}

		if accountTransactionPayments != nil {
			var settlementStatus string
			if accountTransactionPayments.SettlementStatus.Valid {
				settlementStatus = accountTransactionPayments.SettlementStatus.String
			}
			if settlementStatus == constant.StatusPending {
				errUpdate := c.orchestratorSvc.VoidCreditcardTransaction(ctx, &orchestratorModel.VoidTransactionRequest{
					TrxID:             accountTransactionPayments.UUID.String(),
					Status:            constant.StatusFailed,
					ReasonType:        reasonType,
					ReasonDescription: reasonDescription,
					SettlementStatus:  constant.SettlementStatusCancelled,
				})
				if errUpdate != nil {
					c.logger.Error(ctx, constant.ErrWhenUpdateAccountTransaction.Error(),
						logger.Error(errUpdate))
					return errUpdate
				}
			} else if settlementStatus == constant.StatusSuccess {
				if errLedger := c.orchestratorSvc.PostAccountTransaction(ctxTrx, &orchestratorModel.CreateAccountTransactionRequest{
					UUID:                 uuid.New(),
					ReferenceID:          accountTransactionPayments.ReferenceID,
					Type:                 accountTransactionPayments.Type,
					MerchantID:           accountTransactionPayments.MerchantID,
					Currency:             accountTransactionPayments.Currency,
					Credit:               0.00,
					Debit:                accountTransactionPayments.Credit,
					Channel:              accountTransactionPayments.Channel,
					Status:               constant.StatusSuccess,
					ReasonType:           &reasonType,
					ReasonDescription:    &reasonDescription,
					TransactionTimestamp: request.Data.Updated,
					Usecase:              orchestratorModel.TypePayment,
					Processor:            constant.CreditCardCoreProcessor,
					ProcessorID:          request.Data.AcquirerTransactionID,
				}); errLedger != nil {
					return pkgErrors.New(response.HttpErrDatabase, errLedger)
				}
			}
		}

	} else {
		// get payment ledger not using GetTransactionByReferenceIdAndProcessorId because processorReference is not filled once create payment.
		paymentLedger, err := c.orchestratorSvc.FindByReference(ctxTrx, payment.UUID, constant.TypePayment)
		if err != nil {
			return pkgErrors.New(response.HttpErrDatabase, err)
		}

		requestAmount := commonModel.Amount{
			Currency: request.Data.Currency,
			Value:    request.Data.Amount.StringFixed(2),
		}
		transactionStatus := payment.GetTransactionStatus()

		if paymentLedger == nil {
			// Create ledger transaction on paid CC
			if errLedger := c.paymentLedgerSvc.PostCreateLedger(ctxTrx, payment, &paymentModel.PostCreateLedgerRequest{
				Status:  transactionStatus,
				Channel: constant.ChannelCreditCard,
				Amount:  requestAmount,
			}); errLedger != nil {
				return errLedger
			}
		} else if paymentLedger.Status == constant.StatusPending {
			methodDetail := orchestratorModel.MetadataPaymentMethodCC{
				AuthenticationMethod: request.Data.AuthenticationMethod,
				BankMerchantID:       request.Data.BankMerchantID,
				ProcessorStatus:      request.Data.PaymentStatus,
				CardData: &orchestratorModel.MetadataPaymentMethodCCCard{
					First8Digit:    request.Data.CardData.First8Digit,
					Last4Digit:     request.Data.CardData.Last4Digit,
					CardType:       request.Data.CardData.CardType,
					CardBrand:      request.Data.CardData.CardBrand,
					CardIssuing:    request.Data.CardData.CardIssuing,
					CountryCode:    request.Data.CardData.CountryCode,
					IssuingCountry: request.Data.CardData.IssuingCountry,
					Fingerprint:    request.Data.CardData.Fingerprint,
					ExpiryMonth:    request.Data.CardData.ExpiryMonth,
					ExpiryYear:     request.Data.CardData.ExpiryYear,
				},
			}

			if request.Data.AuthorizationData != nil {
				methodDetail.AuthorizationData = &orchestratorModel.MetadataPaymentMethodCCAuthorization{
					AuthorizationResult:   request.Data.AuthorizationData.AuthorizationResult,
					OrderID:               request.Data.AuthorizationData.OrderID,
					TransactionStaus:      request.Data.AuthorizationData.TransactionStaus,
					AuthorizationID:       request.Data.AuthorizationData.AuthorizationID,
					ApprovalCode:          request.Data.AuthorizationData.ApprovalCode,
					BankMerchantID:        request.Data.AuthorizationData.BankMerchantID,
					AcquirerTransactionID: request.Data.AuthorizationData.AcquirerTransactionID,
					TransactionReference:  request.Data.AuthorizationData.TransactionReference,
					CvvResult:             request.Data.AuthorizationData.CvvResult,
					AcquirerResponseCode:  request.Data.AuthorizationData.AcquirerResponseCode,
					Stan:                  request.Data.AuthorizationData.Stan,
					AvsResult:             request.Data.AuthorizationData.AvsResult,
				}
			}

			if request.Data.AuthenticationData != nil {
				methodDetail.AuthenticationData = &orchestratorModel.MetadataPaymentMethodCCAuthentication{
					AuthenticationResult: request.Data.AuthenticationData.AuthenticationResult,
					AuthenticationID:     request.Data.AuthenticationData.AuthenticationID,
					PaRes:                request.Data.AuthenticationData.PaRes,
					VeRes:                request.Data.AuthenticationData.VeRes,
					XID:                  request.Data.AuthenticationData.XID,
					CAVV:                 request.Data.AuthenticationData.CAVV,
					EciCode:              request.Data.AuthenticationData.EciCode,
					ThreeDsVer:           request.Data.AuthenticationData.ThreeDsVer,
					ChallengeCode:        request.Data.AuthenticationData.ChallengeCode,
				}
			}

			updateRequest := orchestratorModel.UpdatePaymentTransactionRequest{
				ProcessorReferenceId:   request.Data.AcquirerTransactionID,
				ProcessorTransactionId: request.Data.TransactionID.String(),
				LedgerId:               paymentLedger.UUID.String(),
				UpdatedAt:              payment.UpdatedAt,
				TrxDatetime:            payment.TrxDatetime,
				Status:                 transactionStatus,
				Channel:                constant.ChannelCreditCard,
				Amount:                 requestAmount,
				MethodDetail:           methodDetail,
			}
			if paymentLedger.SettlementModel.Valid {
				updateRequest.SettlementModel = util.ValueToPtr(paymentLedger.SettlementModel.String)
			}
			if err := c.paymentLedgerSvc.UpdatePendingLedger(ctxTrx, payment, updateRequest); err != nil {
				return err
			}
		}
	}

	if errCommit := c.paymentRepo.CommitTransaction(ctxTrx); errCommit != nil {
		return pkgErrors.New(response.HttpErrDatabase, errCommit)
	}
	isCompleted = true

	// Record payment status history based on payment status after transaction commit
	if c.paymentSvc != nil && request.Data.PaymentStatus != constant.CreditCardStatusVoid {
		transactionStatus := payment.GetTransactionStatus()
		if transactionStatus == constant.StatusSuccess {
			c.paymentSvc.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistorySuccess)
		} else if transactionStatus == constant.StatusFailed {
			c.paymentSvc.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistoryFailed)
		}
	}

	var (
		paymentCallbackRequestWrapper *anypb.Any
		paymentCallbackEvent          string
	)

	if isUnifiedPayment {
		paymentResult, err := c.FindPaymentResponseById(ctx, payment.UUID, payment.MerchantID)
		if err != nil {
			c.logger.Error(ctx, "Error when get payment response by ID", logger.Error(err))
			return err
		}

		paymentCallbackRequestWrapper, err = anypb.New(paymentResult.ToPbUnifiedPaymentCallbackRequest(payment))
		if err != nil {
			c.logger.Error(ctx, "generate anypb.New", logger.Error(err))
			return fmt.Errorf("generate anypb.New: %w", err)
		}

		paymentStatus, err := payment.GetCreditcardPaymentStatus()
		if err != nil {
			c.logger.Error(ctx, "get payment status error", logger.Error(err))
			return err
		}
		if paymentStatus == constant.CreditCardStatusPAID {
			paymentStatus = paymentConstant.PAYMENT_STATUS_SUCCESS
		}
		paymentCallbackEvent = fmt.Sprintf(constant.CallbackEventUnifiedPaymentPattern, paymentStatus)

	} else {
		callbackDataRequest, err := payment.ToPaymentCreditCardCallbackRequest()
		if err != nil {
			c.logger.Error(ctx, constant.ErrWhenBuildCallbackDataRequest, logger.Error(err))
			return err
		}

		paymentCallbackRequestWrapper, err = anypb.New(callbackDataRequest)
		if err != nil {
			c.logger.Error(ctx, "generate anypb.New", logger.Error(err))
			return fmt.Errorf("generate anypb.New: %w", err)
		}

		paymentCallbackEvent = constant.CallbackEventPaymentCreditcardPaid
	}

	if payment.Type == constant.TypeVirtualTerminal {
		c.logger.Info(ctx, "Payment callback notifications are disabled for virtual terminal transactions", logger.Any("payment", payment))
		return nil
	}

	// Send callback on every payment changes to partner
	callbackRequest := &pb.ProcessCallbackRequest{
		Name:       constant.CallbackNamePayment,
		Event:      paymentCallbackEvent,
		MerchantId: initiatorMerchantID,
		Request:    paymentCallbackRequestWrapper,
	}

	if err = c.rabbitMqExt.PublishMerchantCallback(ctx, callbackRequest); err != nil {
		c.logger.Error(ctx, constant.ErrWhenPublishCreditcardData, logger.Error(err))
		return err
	}

	subject, message := constant.GetNotificationMessage(payment.UUID, payment.Status)

	err = c.rabbitMqExt.PushNotification(ctx, &notification.PushNotification{
		RoutingKey: fmt.Sprintf(constant.NotificationRoutingKeyFmt, payment.UUID),
		Payload: notification.PushNotificationPayload{
			ID:        uuid.NewString(),
			Subject:   subject,
			Type:      constant.CreateVAPaymentNotifType,
			Message:   message,
			CreatedAt: time.Now().UTC(),
			Status:    payment.Status,
		},
	})
	if err != nil {
		c.logger.Error(ctx, "push notification for payment "+payment.UUID, logger.Error(err))
	}

	return nil
}

func (c *CreditCardService) PaymentNotificationFDS(
	ctx context.Context,
	request creditcardModel.CardPaymentNotificationRequest,
) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/creditcard/PaymentNotificationFDS")
	defer segment.End()

	c.logger.Info(ctx, "PaymentNotificationFDS", logger.Any("request", request))

	// get payment
	payment, err := c.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, request.Data.MerchantID.String(), request.Data.ReferenceID)
	if err != nil {
		c.logger.Error(ctx, "error when get payment by merchant and reference id", logger.Error(err))
		return err
	}

	// only update metadata if payment status is waiting for external FDS
	// will update CardData, AuthorizationData, AuthenticationData without affecting payment status
	err = creditcardModel.UpdateCreditcardMetaData(payment.Metadata, request.Data.CardData, request.Data.AuthorizationData, request.Data.AuthenticationData, payment.Status)
	if err != nil {
		c.logger.Error(ctx, "error when update creditcard meta data", logger.Error(err))
		return err
	}

	c.logger.Info(ctx, "payment", logger.Any("payment", payment))

	// Update payment
	err = c.paymentRepo.UpdatePaymentData(ctx, payment.ToDTO())
	if err != nil {
		c.logger.Error(ctx, "error when update payment data", logger.Error(err))
		return err
	}

	// also update account_transaction additional info
	paymentLedger, err := c.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		c.logger.Error(ctx, "error when find payment ledger by reference", logger.Error(err))
		return err
	}

	if paymentLedger == nil {
		c.logger.Info(ctx, "payment ledger is nil", logger.Any("payment", payment))
		return nil
	}

	// Parse existing additionalInfo if it exists
	additionalInfo := make(map[string]interface{})
	if paymentLedger.AdditionalInfo.Valid {
		if err := json.Unmarshal(paymentLedger.AdditionalInfo.JSONText, &additionalInfo); err != nil {
			c.logger.Error(ctx, "error when unmarshal existing additional info", logger.Error(err))
			// additionalInfo is already initialized above, so no need to reassign
		}
	}

	// Check if methodDetail exists, if not create it
	methodDetail, ok := additionalInfo["methodDetail"].(map[string]interface{})
	if !ok {
		methodDetail = make(map[string]interface{})
	}

	// Update methodDetail with CardData
	if request.Data.CardData != nil {
		methodDetail["cardData"] = request.Data.CardData
		additionalInfo["methodDetail"] = methodDetail
	}

	// Marshal the updated additionalInfo
	updatedJson, err := json.Marshal(additionalInfo)
	if err != nil {
		c.logger.Error(ctx, "error when marshal updated additional info", logger.Error(err))
		return err
	}

	c.logger.Info(ctx, "updated additional info", logger.Any("updated additional info", additionalInfo))

	// Update additionalInfo in paymentLedger
	paymentLedger.AdditionalInfo = types.NullJSONText{
		Valid:    true,
		JSONText: updatedJson,
	}

	err = c.orchestratorSvc.UpdateAdditionalInfoByID(ctx, paymentLedger.UUID.String(), paymentLedger.AdditionalInfo.JSONText)
	if err != nil {
		c.logger.Error(ctx, "error when update additional info for pending card", logger.Error(err))
		return err
	}

	return nil
}
