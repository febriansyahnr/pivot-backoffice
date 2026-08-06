package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) FindPaymentForSimulationByID(ctx context.Context, id string) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/FindPaymentForSimulationByID")
	defer segment.End()

	// Get payment by id
	payment, errFind := s.paymentRepo.GetPaymentById(ctx, id)
	if errFind != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, errFind)
	} else if payment == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentNotFound)
	}

	// Get payment_method by PaymentMethodID
	paymentMethod, errFindPaymentMethod := s.paymentMethodRepo.GetPaymentMethodById(ctx, payment.PaymentMethodID)
	if errFindPaymentMethod != nil && errors.Is(errFindPaymentMethod, constant.ErrPaymentMethodNotFound) {
		return nil, pkgErrors.New(response.HttpErrRequest, errFindPaymentMethod)
	} else if errFindPaymentMethod != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, errFindPaymentMethod)
	}

	// Get merchant by merchantID
	merchant, errFindMerchant := s.merchantRepo.FindMerchantByID(ctx, payment.MerchantID)
	if errFindMerchant != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, errFindMerchant)
	} else if merchant == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrMerchantNotFound)
	}

	var (
		resp *paymentModel.PaymentResponse
		err  error
	)

	if respUnifiedPaymentV2, ok := s.returnIfUnifiedPaymentV2(ctx, payment); ok {
		return respUnifiedPaymentV2, nil
	}

	switch paymentMethod.Type {
	case paymentConstant.PAYMENT_METHOD_QRIS:
		resp = s.buildPaymentForQris(ctx, payment)
	default:
		resp, err = s.FindPaymentById(ctx, id, payment.MerchantID)
		if err != nil {
			return nil, err
		}
	}

	// Inject merchantName and paymentMethodId
	resp.MerchantName = merchant.Name
	resp.PaymentMethodId = payment.PaymentMethodID

	// return response
	return resp, nil
}

func (s *PaymentService) returnIfUnifiedPaymentV2(ctx context.Context, payment *paymentModel.Payment) (resp *paymentModel.PaymentResponse, ok bool) {
	paymentMetadataB, _ := json.Marshal(payment.Metadata)
	unifiedPaymentMetadata := unifiedPaymentModel.MetadataUnifiedPayment{
		IsUnifiedPaymentV2: false,
	}
	_ = json.Unmarshal(paymentMetadataB, &unifiedPaymentMetadata)

	if !unifiedPaymentMetadata.IsUnifiedPaymentV2 {
		return nil, false
	}

	// Build base response
	resp, err := s.FindPaymentById(ctx, payment.UUID, payment.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "[FindForSimulation] error when find payment by id", pdkLogger.Error(err))
		return nil, false
	}

	chargeMethodDetails := unifiedPaymentMetadata.MethodDetail
	// Find charge
	charge, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		s.logger.Error(ctx, "[FindForSimulation] error when find charge by reference", pdkLogger.Error(err))
		return nil, false
	} else if charge == nil {
		s.logger.Warn(ctx, "[FindForSimulation] charge by reference is not found", pdkLogger.Error(err))
		return nil, false
	}

	chargeMethodDetails = &unifiedPaymentModel.ChargePaymentMethodDetails{}
	_ = json.Unmarshal(charge.AdditionalInfo.JSONText, &struct {
		MethodDetail interface{} `json:"methodDetail"`
	}{
		MethodDetail: chargeMethodDetails,
	})

	if resp.PaymentMethod == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT {
		resp.VirtualAccount = &paymentModel.PaymentVirtualAccountResponse{
			Issuer:                chargeMethodDetails.VirtualAccount.Channel,
			VirtualAccountTrxType: paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC, // Force dynamic for unified payment
			VirtualAccountNumber:  chargeMethodDetails.VirtualAccount.VirtualAccountNumber,
			VirtualAccountName:    chargeMethodDetails.VirtualAccount.VirtualAccountName,
			ExpiredDate:           &chargeMethodDetails.VirtualAccount.ExpiryAt,
		}

		if payment.Type == constant.UnifiedPaymentTypeMultiple {
			resp.VirtualAccount.VirtualAccountTrxType = paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC
			if payment.Amount.InexactFloat64() == 0 {
				resp.VirtualAccount.VirtualAccountTrxType = paymentConstant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC
			}
		}
	} else if resp.PaymentMethod == paymentConstant.PAYMENT_METHOD_QRIS {
		resp.Qris = &paymentModel.PaymentQrisResponse{
			ReferenceNo:  chargeMethodDetails.Qr.RetrievalReferenceNumber,
			QrContent:    chargeMethodDetails.Qr.QrContent,
			QrUrl:        chargeMethodDetails.Qr.QrUrl,
			MerchantName: chargeMethodDetails.Qr.MerchantName,
			QrType:       constant.QrTypeDynamic, // Force dynamic for unified payment
		}

		if chargeMethodDetails.Qr.QrType != "" {
			resp.Qris.QrType = chargeMethodDetails.Qr.QrType
		}
	}

	return resp, true
}
