package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/ewallet"
	snapQrisModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	snapCoreTopUpSimulationModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/topUpSimulation"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

func (s *PaymentService) ProcessPaymentForSimulationByID(ctx context.Context, id string, paymentAmount commonModel.Amount, status string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/ProcessPaymentForSimulationByID")
	defer segment.End()

	// Get payment by id
	payment, errFind := s.paymentRepo.GetPaymentById(ctx, id)
	if errFind != nil {
		return pkgErrors.New(response.HttpErrDatabase, errFind)
	} else if payment == nil {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentNotFound)
	}

	switch status {
	case constant.ChargeStatusExpired:
		s.logger.Info(ctx, "[SimulatePayment] Payment is EXPIRED")
		if err := s.rabbitMqExt.PublishWithDelay(
			ctx,
			rabbitMqExt.PaymentExpirationRoutingKey,
			&paymentModel.ExpiringPayment{
				UUID:       payment.UUID,
				MerchantID: payment.MerchantID,
				ExpiredAt:  time.Now(),
			},
			10*time.Second,
		); err != nil {
			s.logger.Error(ctx, "[SimulatePayment] error publish payment expiration message", logger.Error(err))
			return err
		}
		return nil
	case constant.ChargeStatusProcessing:
		s.logger.Info(ctx, "[SimulatePayment] Payment is PROCESSING")
		if err := s.rabbitMqExt.PublishWithDelay(
			ctx,
			rabbitMqExt.PaymentExpirationRoutingKey,
			&paymentModel.ExpiringPayment{
				UUID:         payment.UUID,
				MerchantID:   payment.MerchantID,
				ChargeStatus: constant.ChargeStatusProcessing,
			},
			10*time.Second,
		); err != nil {
			s.logger.Error(ctx, "[SimulatePayment] error publish payment processing message", logger.Error(err))
			return err
		}
	}

	// Record payment processing for simulation
	s.RecordPaymentStatusHistory(ctx, id, constant.StatusHistoryActorSystem, constant.PaymentStatusHistoryProcessing)

	// Get payment_method by PaymentMethodID
	paymentMethod, errFindPaymentMethod := s.paymentMethodRepo.GetPaymentMethodById(ctx, payment.PaymentMethodID)
	if errFindPaymentMethod != nil && errors.Is(errFindPaymentMethod, constant.ErrPaymentMethodNotFound) {
		return pkgErrors.New(response.HttpErrRequest, errFindPaymentMethod)
	} else if errFindPaymentMethod != nil {
		return pkgErrors.New(response.HttpErrDatabase, errFindPaymentMethod)
	}

	partnerReferenceNo := ""
	if payment.ReferenceID != nil {
		partnerReferenceNo = *payment.ReferenceID
	}

	processorReferenceNumber := ""
	if payment.ProcessorReferenceNumber != nil {
		processorReferenceNumber = *payment.ProcessorReferenceNumber
	}

	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		From:        serviceName,
		OriginId:    payment.UUID,
		ReferenceId: payment.MerchantID,
	})

	switch paymentMethod.Type {
	case paymentConstant.PAYMENT_METHOD_QRIS:
		return s.processQrisForSimulation(ctx, payment, &snapQrisModel.QrMpmPaymentSimulationRequest{
			PartnerReferenceNo: partnerReferenceNo,
			Status:             constant.StatusSuccess,
			Amount:             paymentAmount,
		})
	case paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT:
		return s.processVAForSimulation(ctx, payment, &snapCoreTopUpSimulationModel.TopupSimulationRequest{
			VANumber: processorReferenceNumber,
			TotalAmount: snapCoreTopUpSimulationModel.Amount{
				Value:    paymentAmount.Value,
				Currency: paymentAmount.Currency,
			},
		})
	case paymentConstant.PAYMENT_METHOD_EWALLET:
		// validate payment
		charge, err := s.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.TypePayment)
		if err != nil {
			return err
		}

		if charge == nil {
			s.logger.Warn(ctx, "payment charge not found", logger.String("paymentId", payment.UUID))
			return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentChargeNotFound)
		}
		// end of payment validation

		// call processor ewallet simulation
		return s.processEWalletForSimulation(ctx, payment, &ewallet.EWalletPaymentSimulationRequest{
			Acquirer:            payment.PaymentMethod.Acquirer,
			OriginalReferenceID: charge.UUID.String(),
			Status:              status,
			Amount:              paymentAmount,
		})
	default:
		return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodNotFound)
	}
}

func (s *PaymentService) processQrisForSimulation(ctx context.Context, payment *paymentModel.Payment, data *snapQrisModel.QrMpmPaymentSimulationRequest) error {
	paymentMetadataB, _ := json.Marshal(payment.Metadata)
	unifiedPaymentMetadata := unifiedPaymentModel.MetadataUnifiedPayment{
		IsUnifiedPaymentV2: false,
	}
	_ = json.Unmarshal(paymentMetadataB, &unifiedPaymentMetadata)

	if !unifiedPaymentMetadata.IsUnifiedPaymentV2 {
		qrisMetadata, errQrisMetadata := buildQrisMetadata(payment.Metadata)
		if errQrisMetadata != nil {
			s.logger.Error(ctx, "Error build QRIS metadata", logger.Error(errQrisMetadata))
			return errQrisMetadata
		}

		// Validate expired date
		expiredAt := payment.CreatedAt.Add(time.Duration(qrisMetadata.ValidityPeriod) * time.Second)
		if expiredAt.Before(time.Now().UTC()) && qrisMetadata.QrType == constant.QrTypeDynamic {
			return pkgErrors.New(response.HttpErrRequest, errors.New("expiredAt is not allowed to be less than current time"))
		}

		paidAmountValue, _ := decimal.NewFromString(data.Amount.Value)
		if qrisMetadata.QrType == constant.QrTypeDynamic &&
			paidAmountValue.Cmp(decimal.Zero) > 0 && !paidAmountValue.Equal(payment.Amount) {

			s.logger.Error(ctx, "Error payment total amount not match with paid amount", logger.Error(constant.ErrPaymentAmountNotMatch))
			return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAmountNotMatch)
		}
	}

	if err := s.snapCoreRepo.QrMpmPaymentSimulation(ctx, data); err != nil {
		s.logger.Error(ctx, "Error process payment simulation for QRIS", logger.Any("partnerReferenceNo", data.PartnerReferenceNo), logger.Error(err))

		// Record payment failed for simulation
		s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistoryFailed)

		return pkgErrors.New(response.HttpErrInternal, constant.ErrPartnerInGeneral)
	}

	// Record payment success for simulation
	s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistoryPaid)

	return nil
}

func (s *PaymentService) processVAForSimulation(
	ctx context.Context,
	payment *paymentModel.Payment, data *snapCoreTopUpSimulationModel.TopupSimulationRequest) error {

	paymentMetadataB, _ := json.Marshal(payment.Metadata)
	unifiedPaymentMetadata := unifiedPaymentModel.MetadataUnifiedPayment{
		IsUnifiedPaymentV2: false,
	}
	_ = json.Unmarshal(paymentMetadataB, &unifiedPaymentMetadata)

	if !unifiedPaymentMetadata.IsUnifiedPaymentV2 {
		vaMetadata, errVAMetadata := buildVAMetadataSimulation(payment.Metadata)
		if errVAMetadata != nil {
			s.logger.Error(ctx, "Error build VA metadata", logger.Error(errVAMetadata))
			return errVAMetadata
		}

		// Validate amount and expired date
		vaTrxType := snapCoreModel.FindVaTrxTypeByCriteria(vaMetadata.IsClosedAmount, vaMetadata.IsSingleUse)
		if vaTrxType == paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC && vaMetadata.ExpiredAt.Before(time.Now()) {
			return pkgErrors.New(response.HttpErrInternal, errors.New("expiredDate is not allowed to be less than current time"))
		}
		if (vaTrxType == paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC || vaTrxType == paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC) && vaMetadata.TotalAmount.Value.Cmp(decimal.NewFromInt(paymentConstant.VIRTUAL_ACCOUNT_MINIMUM_AMOUNT)) < 0 {
			return pkgErrors.New(response.HttpErrInternal, errors.New("totalAmount is not allowed to be less than 10000 for type CLOSED_DYNAMIC payment"))
		}
	}

	_, err := s.snapCoreRepo.TopUpSimulation(ctx, *data)
	if err != nil {
		s.logger.Error(ctx, "failed to create top up simulation VA", logger.Error(err))

		// Record payment failed for simulation
		s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistoryFailed)

		return pkgErrors.New(response.HttpErrInternal, err)
	}

	// Record payment success for simulation
	s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistoryPaid)

	return nil
}

func buildVAMetadataSimulation(paymentMetadata *map[string]any) (*paymentModel.PaymentSimulationMetadataVA, error) {
	if paymentMetadata == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentNotFound)
	}

	// Check if "snapCore" exists in the map
	snapCore, ok := (*paymentMetadata)["snapCore"]
	if !ok {
		return nil, pkgErrors.New(response.HttpErrRequest, errors.New("snapCore data not found"))
	}

	// Marshal only the "snapCore" data
	jsonData, err := json.Marshal(snapCore)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, errors.New("error marshaling snapCore data"))
	}

	resp := paymentModel.PaymentSimulationMetadataVA{}
	err = json.Unmarshal(jsonData, &resp)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrRequest, errors.New("error unmarshaling data to struct"))
	}

	return &resp, nil
}

func (s *PaymentService) processEWalletForSimulation(
	ctx context.Context,
	payment *paymentModel.Payment, data *ewallet.EWalletPaymentSimulationRequest) error {

	paymentMetadataB, _ := json.Marshal(payment.Metadata)
	unifiedPaymentMetadata := unifiedPaymentModel.MetadataUnifiedPayment{
		IsUnifiedPaymentV2: false,
	}
	_ = json.Unmarshal(paymentMetadataB, &unifiedPaymentMetadata)

	if !unifiedPaymentMetadata.IsUnifiedPaymentV2 {
		err := errors.New("payment is not using unified payment v2")
		s.logger.Warn(ctx, "unified payment validation", logger.Error(err))
		return err
	}

	_, err := s.snapCoreRepo.EWalletPaymentSimulation(ctx, data)
	if err != nil {
		s.logger.Error(ctx, "failed to request payment ewallet simulation", logger.Error(err))

		// Record payment failed for simulation
		s.RecordPaymentStatusHistory(ctx, payment.UUID, constant.StatusHistoryActorProcessor, constant.PaymentStatusHistoryFailed)

		return pkgErrors.New(response.HttpErrInternal, err)
	}

	return nil
}
