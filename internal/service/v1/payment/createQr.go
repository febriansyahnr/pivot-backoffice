package paymentService

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *PaymentService) createPaymentUsingQrMpm(ctx context.Context, merchantID string, paymentRequest paymentModel.PaymentRequest) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/createPaymentUsingQrMpm")
	defer segment.End()

	var (
		pendingLedgerID    = uuid.New()
		paymentResponse    paymentModel.PaymentResponse
		qrMerchantId       = merchantID
		merchantExternalId string

		qrisAdditionalInfo = map[string]interface{}{
			constant.ProcessorExternalIDKey: pendingLedgerID,
		}
	)

	if paymentRequest.Qris.SubMerchantId != "" {
		qrMerchantId = paymentRequest.Qris.SubMerchantId
	}

	// Get merchant by merchant id (merchantId = subMerchantId if on behalf sub merchant)
	merchant, errFind := s.merchantRepo.FindMerchantByID(ctx, qrMerchantId)
	if errFind != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, errFind)
	} else if merchant == nil ||
		(paymentRequest.Qris.SubMerchantId != "" && merchant.ParentID.String != merchantID) {
		// validate wheter the submerchant belong to the merchant
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
	}
	// When sub-merchant create payment directly
	if _, ok := ctx.Value(constant.CtxParentMerchantId).(string); merchant.ParentID.String != "" && !ok {
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)
	}

	derivedMerchantId := merchantID
	merchantExternalId = merchant.ExternalId
	if merchant.ParentID.Valid && merchant.KYCStatus.String != constant.KYCStatusApproved {
		derivedMerchantId = merchant.ParentID.String

		if parentMerchant, errFindParent := s.merchantRepo.FindMerchantByID(ctx, derivedMerchantId); errFindParent != nil {
			return nil, pkgErrors.New(response.HttpErrDatabase, errFindParent)
		} else if parentMerchant != nil {
			merchantExternalId = parentMerchant.ExternalId
		}
	}

	// Get payment method by type
	paymentMethod, err := s.paymentMethodRepo.GetActivePaymentMethodByRequest(ctx, &paymentModel.GetPaymentMethodFilterRequest{
		MerchantID: derivedMerchantId,
		Category:   constantPayment.PAYMENT_METHOD_CATEGORY_PAYMENT,
		Type:       constantPayment.PAYMENT_METHOD_QRIS,
	})
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if paymentMethod == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodNotFound)
	}

	if err := s.ValidatePaymentExpiry(ctx, paymentModel.PaymentRequestExpiryValidation{
		MerchantID:    merchantID,
		Method:        constantPayment.PAYMENT_METHOD_QRIS,
		Request:       &paymentRequest,
		PaymentMethod: paymentMethod,
	}); err != nil {
		return nil, err
	}

	snapCoreRequestSubMerchantId := ""
	snapCoreRequestStoreId := ""
	// Get parent id and child id (the merchant itself) from qr_registration table
	qrRegistration, errFindQr := s.qrisSvc.FindQrRegistrationByExternalID(ctx, merchantExternalId)
	if errFindQr != nil {
		if errFindQr.Error() == pkgErrors.New(response.HttpErrNotFound, constant.ErrDataNotFound).Error() {
			return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrMerchantNotFound)
		}
		return nil, errFindQr
	}

	if qrRegistration.MerchantType == constant.QrMerchantTypeMerchant {
		snapCoreRequestSubMerchantId = *qrRegistration.AcquirerMerchantId
	} else {
		// If submerchant
		snapCoreRequestSubMerchantId = qrRegistration.AcquirerParentMerchantId
		snapCoreRequestStoreId = *qrRegistration.AcquirerMerchantId
	}

	// If paymentRequest.Qris.QrType == "STATIC", check if the qr already exist in payments table
	// Note that if one merchant one static qr policy is applied, then the qr will be reused
	// If exist, fetch the data and reuse it
	if paymentRequest.Qris.QrType == constant.QrTypeStatic {
		payment, err := s.paymentRepo.GetPaymentQrStaticByMerchantId(
			ctx,
			derivedMerchantId,
			paymentRequest.Qris.SubMerchantId,
			paymentMethod.UUID,
		)
		if err != nil {
			s.logger.Error(ctx, "error when get payment qr static by merchant id", logger.Error(err))
			return nil, err
		}
		if payment != nil {
			paymentResponse = s.processExistingPaymentQr(payment)
			return &paymentResponse, nil
		}
	}

	paymentUUID := uuid.NewString()
	if paymentRequest.UUID != "" {
		paymentUUID = paymentRequest.UUID
	}

	// Passing Account Transaction ID to Processor
	if _, ok := ctx.Value(constant.CtxChangePaymentMethod).(bool); ok {
		pendingLedger, err := s.orchestratorSvc.FindByReference(ctx, paymentUUID, constant.TypePayment)
		if err != nil {
			return nil, err
		} else if pendingLedger == nil {
			return nil, pkgErrors.New(response.HttpStatusErrorUnprocessableContent, constant.ErrPaymentNotFound)
		}

		qrisAdditionalInfo[constant.ProcessorExternalIDKey] = pendingLedger.UUID
	}

	// Request generate qr mpm to snap core
	snapCoreRequest := snapCoreModel.GenerateQrMpmRequest{
		PartnerReferenceNo: paymentRequest.ReferenceID,
		SubMerchantID:      snapCoreRequestSubMerchantId,
		StoreID:            snapCoreRequestStoreId,
		Amount: commonModel.Amount{
			Currency: "IDR",
			Value:    paymentRequest.Qris.Amount.Value.StringFixed(2),
		},
		QrType:         paymentRequest.Qris.QrType,
		ValidityPeriod: paymentRequest.Qris.ValidityPeriod,
		Acquirer:       paymentMethod.Acquirer,
		AdditionalInfo: qrisAdditionalInfo,
	}

	// Check if NEW flow (multi-acquirer routing) is enabled for this merchant
	if constant.IsQrMultiAcquirerRoutingEnabled(merchantID) {
		// NEW flow: Send merchant UUID to SNAP Core for priority-based routing
		snapCoreRequest.MerchantID = merchantID
		snapCoreRequest.Acquirer = "" // Clear acquirer - SNAP Core will route based on priority
	} else if qrRegistration.AcquirerMerchantId != nil && *qrRegistration.AcquirerMerchantId != "" {
		snapCoreRequest.MerchantID = *qrRegistration.AcquirerMerchantId
	}

	if qrRegistration.AcquirerTerminalId != nil && *qrRegistration.AcquirerTerminalId != "" {
		snapCoreRequest.TerminalID = *qrRegistration.AcquirerTerminalId
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		OriginId:    paymentUUID,
		ReferenceId: merchantID,
		From:        serviceName,
	})

	// Request to snap core
	snapCoreResp, err := s.snapCoreRepo.GenerateQrMpm(ctx, snapCoreRequest)
	if err != nil {
		s.logger.Error(ctx, "error when generate qr mpm", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrRequest, err)
	}
	// Truncate milisecond to omit the milisecond in the expired at
	snapCoreResp.ExpiredAt = snapCoreResp.ExpiredAt.Truncate(time.Second)

	// merchant descriptor using name from merchant short name
	snapCoreResp.MerchantName = qrRegistration.MerchantShortName

	// If multi-acquirer routing is enabled and SNAP Core returned a different acquirer,
	// find the correct payment method that matches the actual acquirer used
	actualPaymentMethodID := paymentMethod.UUID
	if constant.IsQrMultiAcquirerRoutingEnabled(merchantID) && snapCoreResp.Acquirer != "" {
		// Clean acquirer name by removing common suffixes (BNC_QRIS -> bnc, BNI_VA -> bni)
		responseAcquirer := util.CleanAcquirerName(snapCoreResp.Acquirer)

		if responseAcquirer != strings.ToLower(paymentMethod.Acquirer) {
			// Find payment method by category=PAYMENT, type=QRIS, acquirer from response
			matchedPaymentMethod, err := s.paymentMethodRepo.GetPaymentMethodByCategoryTypeAndAcquirer(
				ctx,
				constantPayment.PAYMENT_METHOD_CATEGORY_PAYMENT,
				constantPayment.PAYMENT_METHOD_QRIS,
				responseAcquirer,
			)
			if err != nil {
				s.logger.Error(ctx, "error finding payment method by acquirer from SNAP Core response",
					logger.Error(err),
					logger.String("rawAcquirer", snapCoreResp.Acquirer),
					logger.String("cleanedAcquirer", responseAcquirer),
					logger.String("originalAcquirer", paymentMethod.Acquirer))
				// Continue with original payment method if lookup fails
			} else if matchedPaymentMethod != nil {
				s.logger.Info(ctx, "using payment method matched from SNAP Core acquirer response",
					logger.String("originalPaymentMethodID", paymentMethod.UUID),
					logger.String("originalAcquirer", paymentMethod.Acquirer),
					logger.String("matchedPaymentMethodID", matchedPaymentMethod.UUID),
					logger.String("rawAcquirer", snapCoreResp.Acquirer),
					logger.String("cleanedAcquirer", responseAcquirer))
				actualPaymentMethodID = matchedPaymentMethod.UUID
			}
		}
	}

	amount := paymentRequest.Qris.Amount.Value // client input amount
	discount := decimal.NewFromFloat(0)
	feeInDecimal := decimal.NewFromFloat(0)
	totalAmount := amount.Add(feeInDecimal).Sub(discount)

	// Create metadata
	qrisPaymentMetadataB, _ := json.Marshal(paymentRequest.Qris)
	metadataMap := make(map[string]interface{})
	json.Unmarshal(qrisPaymentMetadataB, &metadataMap)
	metadataMap["snapCore"] = snapCoreResp
	metadataMap["isSnap"] = paymentRequest.IsSnap
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		metadataMap["onBehalf"] = &merchantModel.OnBehalfObject{
			ParentMerchantId: parentMerchantId,
		}
	}
	metadataMap["clientRedirectUrl"] = paymentRequest.ClientRedirectUrl
	metadataMap[constant.IsUnifiedPaymentKey] = paymentRequest.IsUnifiedPayment
	if paymentRequest.SplitRoutingConfigurations != nil && len(*paymentRequest.SplitRoutingConfigurations) > 0 {
		metadataMap[constant.SplitRoutingPaymentConfigKey] = *paymentRequest.SplitRoutingConfigurations
	}

	accountTrxMetadata := orchestratorModel.MetadataPayment[orchestratorModel.MetadataPaymentMethodQRIS]{
		ReconReferenceNo: snapCoreResp.ReferenceNo,
		ExpiredAt:        snapCoreResp.ExpiredAt,
		MethodDetail: orchestratorModel.MetadataPaymentMethodQRIS{
			QrType:             paymentRequest.Qris.QrType,
			QrMethodType:       paymentRequest.Qris.QrMethodType,
			PartnerReferenceNo: paymentRequest.ReferenceID,
			StoreID:            snapCoreResp.StoreID,
			MerchantID:         snapCoreResp.MerchantID,
			MerchantName:       snapCoreResp.MerchantName,
			QrUrl:              snapCoreResp.QrUrl,
			QrContent:          snapCoreResp.QrContent,
			AdditionalInfo:     snapCoreResp.AdditionalInfo,
		},
	}
	rawAccountTrxMetadata, _ := json.Marshal(accountTrxMetadata)

	metaDataB, _ := json.Marshal(metadataMap)
	metaDataString := string(metaDataB)

	// Begin Tx
	ctx, err = s.paymentRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	if paymentRequest.ReferenceID == "" {
		paymentRequest.ReferenceID = snapCoreResp.PartnerReferenceNo
	}

	// Create payment model to save to db
	var payment paymentModel.Payment
	paymentDTO := paymentModel.PaymentDTO{
		UUID:                     paymentUUID,
		ReferenceID:              &paymentRequest.ReferenceID,
		CustomerID:               paymentRequest.Customer.CustomerID,
		MerchantID:               merchantID,
		PaymentMethodID:          actualPaymentMethodID, // Use actual payment method from SNAP Core response
		ProcessorReferenceNumber: &snapCoreResp.ReferenceNo,
		Currency:                 paymentRequest.Qris.Amount.Currency,
		Amount:                   paymentRequest.Qris.Amount.Value.Round(2),
		Fee:                      &feeInDecimal,
		Discount:                 &discount,
		TotalAmount:              totalAmount,
		Status:                   paymentRequest.InitiateStatus,
		Metadata:                 &metaDataString,
		CreatedBy:                &paymentRequest.CreatedBy,
		CreatedAt:                time.Now().UTC(),
		UpdatedAt:                time.Now().UTC(),
		PaymentURL:               paymentRequest.PaymentUrl,
	}

	if paymentRequest.Qris.QrType == constant.QrTypeDynamic {
		paymentDTO.ExpiredAt = &snapCoreResp.ExpiredAt
	}

	if _, ok := ctx.Value(constant.CtxChangePaymentMethod).(bool); ok {

		currentPaymentMethod, _ := ctx.Value(constant.CtxCurrentPaymentMethod).(string)
		if err = s.paymentRepo.ChangePaymentMethod(ctx, &paymentDTO); err == nil && paymentRequest.Qris.QrType != constant.QrTypeStatic {
			trxRequest := orchestratorModel.UpdateTransactionWithPendingStatus{
				Channel:         constant.ChannelQris,
				Metadata:        rawAccountTrxMetadata,
				UpdatedAt:       time.Now().UTC(),
				Processor:       constant.SnapCoreProcessor,
				ProcessorID:     snapCoreResp.UUID,
				SettlementModel: paymentMethod.ChannelType,
			}
			err = s.accountTransactionRepo.UpdateTransactionWithPendingStatusByReferenceIdTypeAndChannel(
				ctx, paymentUUID, constant.TypePayment, currentPaymentMethod, trxRequest,
			)
		}

	} else {
		if err = s.paymentRepo.CreatePayment(ctx, &paymentDTO); err == nil && paymentRequest.Qris.QrType != constant.QrTypeStatic {
			// Create PENDING transaction with QRIS detail
			trxRequest := &orchestratorModel.CreateAccountTransactionRequest{
				UUID:                 uuid.New(),
				ReferenceID:          paymentUUID,
				Type:                 constant.TypePayment,
				MerchantID:           util.ParseUUID(merchantID),
				Currency:             paymentRequest.Qris.Amount.Currency,
				Credit:               paymentRequest.Qris.Amount.Value.Round(2).InexactFloat64(),
				Channel:              constant.ChannelQris,
				Status:               constant.StatusPending,
				SettlementStatus:     util.ValueToPtr(constant.StatusPending),
				TransactionTimestamp: paymentDTO.CreatedAt,
				Usecase:              constant.TypePayment,
				Processor:            constant.SnapCoreProcessor,
				ProcessorID:          snapCoreResp.UUID,
				AdditionalInfo: types.NullJSONText{
					Valid: true, JSONText: rawAccountTrxMetadata,
				},
				SettlementModel: util.ValueToPtr(paymentMethod.ChannelType),
			}
			// Action For Post Transaction
			err = s.orchestratorSvc.PostAccountTransaction(ctx, trxRequest)
		}
	}
	if err != nil {
		s.logger.Error(ctx, "error when create payment", logger.Error(err))

		// Rollback Tx
		if errRollback := s.paymentRepo.RollbackTransaction(ctx); errRollback != nil {
			return nil, errRollback
		}

		return nil, err
	}

	payment.PaymentFromDTO(&paymentDTO)
	s.PublishQRExpiryMessage(ctx, &paymentDTO)

	// Commit Tx
	if err := s.paymentRepo.CommitTransaction(ctx); err != nil {
		return nil, err
	}

	paymentResponse.ToQrisResponse(
		&paymentDTO,
		snapCoreResp,
		&paymentRequest,
	)

	// Add payment simulation only on staging
	if s.config.Environment == constant.EnvironmentStaging {
		paymentResponse.Qris.Metadata = map[string]any{
			constant.PaymentSimulatorKey: fmt.Sprintf(
				s.config.MerchantPortalConfig.PaymentSimulationPatternURL,
				base64.StdEncoding.EncodeToString([]byte(payment.UUID)),
			),
		}
	}

	return &paymentResponse, nil
}

func (s *PaymentService) processExistingPaymentQr(payment *paymentModel.Payment) paymentModel.PaymentResponse {
	// Parse metadata
	metadata := *payment.Metadata

	// Parse snap core response
	var snapCoreResp snapCoreModel.GenerateQrMpmResponseData
	snapCoreB, _ := json.Marshal(metadata["snapCore"])
	json.Unmarshal(snapCoreB, &snapCoreResp)

	// Parse payment request from metadata
	paymentRequest := &paymentModel.PaymentRequest{
		PaymentMethod: constantPayment.PAYMENT_METHOD_QRIS,
		Qris:          &paymentModel.PaymentMetadataQris{},
	}
	paymentRequestB, _ := json.Marshal(metadata)
	json.Unmarshal(paymentRequestB, &paymentRequest.Qris)

	// Create paymentDTO
	paymentDTO := payment.ToDTO()

	var paymentResponse paymentModel.PaymentResponse
	paymentResponse.ToQrisResponse(
		paymentDTO,
		&snapCoreResp,
		paymentRequest,
	)

	// Add payment simulation only on staging
	if s.config.Environment == constant.EnvironmentStaging {
		paymentResponse.Qris.Metadata = map[string]any{
			constant.PaymentSimulatorKey: fmt.Sprintf(
				s.config.MerchantPortalConfig.PaymentSimulationPatternURL,
				base64.StdEncoding.EncodeToString([]byte(payment.UUID)),
			),
		}
	}

	return paymentResponse
}

// PublishQRExpiryMessage publishes a payment expiration message to RabbitMQ if the payment
// expires before the next day's cutoff time (18:00 UTC).
//
// The function calculates the cutoff time (lastPublishExpiryAt) which is set to 18:00 UTC
// (01:00 JKT time) of the current day. If the current time is between 18:00-24:00 UTC,
// the cutoff is set to 18:00 UTC of the next day.
//
// If the payment's expiration time is after the cutoff time, no message is published.
// Otherwise, a message with the payment details is published to the RabbitMQ with a TTL
// equal to the time remaining until the payment expires.
//
// If there's an error during publishing, it's logged but not returned to the caller.
func (s *PaymentService) PublishQRExpiryMessage(ctx context.Context, payment *paymentModel.PaymentDTO) {
	now := time.Now().UTC()

	if payment.ExpiredAt == nil {
		return
	}

	// Set lastPublishExpiryAt, equals to 01.00 JKT time
	lastPublishExpiryAt := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, now.Location())
	if now.Hour() >= 18 && now.Hour() < 24 {
		lastPublishExpiryAt = lastPublishExpiryAt.Add(24 * time.Hour)
	}

	if payment.ExpiredAt.After(lastPublishExpiryAt) {
		return
	}

	err := s.rabbitMqExt.PublishWithDelay(
		ctx,
		rabbitMqExt.PaymentExpirationRoutingKey,
		&paymentModel.ExpiringPayment{
			UUID:       payment.UUID,
			MerchantID: payment.MerchantID,
			ExpiredAt:  util.ValueOfPtr(payment.ExpiredAt),
		},
		payment.ExpiredAt.Sub(now),
	)
	if err != nil {
		s.logger.Error(ctx, "error publish QR payment expiration message", logger.Error(err))
	}
}
