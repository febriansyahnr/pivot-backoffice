package paymentService

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/shopspring/decimal"
)

func (s *PaymentService) GetAndUpdateVirtualAccountPayment(ctx context.Context, request *paymentModel.VirtualAccountPaymentNotificationRequest) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetAndUpdateVirtualAccountPayment")
	defer segment.End()

	switch strings.ToUpper(request.Status) {
	case paymentConstant.VirtualAccountStatusPaid:
		return s.payVirtualAccountPayment(ctx, request)

	case paymentConstant.VirtualAccountStatusExpired:
		return s.expireVirtualAccountPayment(ctx, request)

	default:
		return nil, errors.New("status not allowed")
	}
}

func (s *PaymentService) payVirtualAccountPayment(ctx context.Context, request *paymentModel.VirtualAccountPaymentNotificationRequest) (*paymentModel.Payment, error) {
	var (
		snapCoreResp snapCoreModel.CreateVirtualAccountResponseData
	)
	// Find payment first
	payment, err := s.paymentRepo.GetActivePaymentByProcessorReferenceNumber(ctx, &paymentModel.GetActivePaymentByProcessorReferenceNumberRequest{
		ProcessorReferenceNumber: request.Number,
		Amount:                   request.PaidAmount.ToDecimal(),
	})
	if err != nil {
		return nil, err

	} else if payment == nil {
		return nil, constant.ErrPaymentNotFound
	}

	// Validate acquirer
	if payment.PaymentMethod.Acquirer != request.Acquirer {
		return nil, constant.ErrPaymentNotFound
	}

	paidAmountValue, err := decimal.NewFromString(request.PaidAmount.Value)
	if err != nil {
		err = constant.ErrInvalidRequestPayload
		s.logger.Error(ctx, "error when parsing paid amount", logger.Error(err))
		return nil, err
	}

	// Get VA payment metadata
	// marshal metadata to json
	jsonData, errMarshal := json.Marshal(payment.Metadata)
	if errMarshal != nil {
		s.logger.Error(ctx, "error when marshal payment metadata", logger.Error(errMarshal))
		return nil, pkgError.New(httpResponse.HttpErrInternal, errMarshal)
	}

	// unmarshal metadata to snapCoreResp
	json.Unmarshal(jsonData, &struct {
		SnapCore interface{} `json:"snapCore"`
	}{
		SnapCore: &snapCoreResp,
	})

	// We assume this is full payment, for partial payment will be decided later for the logic
	// applied for closed amount va type
	if snapCoreResp.IsClosedAmount && paidAmountValue.Cmp(decimal.Zero) > 0 && !paidAmountValue.Equal(payment.Amount) {
		err = constant.ErrPaymentAmountNotMatch
		s.logger.Error(ctx, "error payment total amount not match with paid amount", logger.Error(err))
		return nil, err
	}

	// update status to success for dynamic va type
	if snapCoreResp.IsSingleUse {
		payment.Status = paymentConstant.PAYMENT_STATUS_SUCCESS
		payment.UpdatedAt = time.Now().UTC()
		err = s.paymentRepo.UpdatePaymentStatus(ctx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt)
		if err != nil {
			return nil, err
		}
	}

	return payment, nil
}

func (s *PaymentService) expireVirtualAccountPayment(ctx context.Context, request *paymentModel.VirtualAccountPaymentNotificationRequest) (*paymentModel.Payment, error) {
	var snapCoreResp snapCoreModel.CreateVirtualAccountResponseData

	// Find payment first
	payment, err := s.paymentRepo.GetActivePaymentByProcessorReferenceNumber(ctx, &paymentModel.GetActivePaymentByProcessorReferenceNumberRequest{
		ProcessorReferenceNumber: request.Number,
		Amount:                   request.PaidAmount.ToDecimal(),
	})
	if err != nil {
		return nil, err

	} else if payment == nil {
		return nil, constant.ErrPaymentNotFound
	}

	// Validate acquirer
	if payment.PaymentMethod.Acquirer != request.Acquirer {
		return nil, constant.ErrPaymentNotFound
	}

	// marshal metadata to json
	jsonData, errMarshal := json.Marshal(payment.Metadata)
	if errMarshal != nil {
		s.logger.Error(ctx, "error when marshal payment metadata", logger.Error(errMarshal))
		return nil, pkgError.New(httpResponse.HttpErrInternal, errMarshal)
	}

	// unmarshal metadata to snapCoreResp
	json.Unmarshal(jsonData, &struct {
		SnapCore interface{} `json:"snapCore"`
	}{
		SnapCore: &snapCoreResp,
	})

	// update snapCore resp
	snapCoreResp.ExpiredAt = time.Now().UTC()
	if request.ExpiredAt != nil {
		snapCoreResp.ExpiredAt = *request.ExpiredAt
	}
	snapCoreResp.Status = paymentConstant.VirtualAccountStatusExpired

	metadataMap := make(map[string]any)
	metadataMap["snapCore"] = snapCoreResp

	jsonMetadata, _ := json.Marshal(metadataMap)
	metaDataString := string(jsonMetadata)

	// get customer info
	customer, _ := s.customerRepo.FindCustomerById(ctx, payment.CustomerID)
	if customer == nil {
		customer = &customerModel.Customer{}
	}

	// update metadata
	payment.Metadata = &metadataMap
	err = s.paymentRepo.UpdatePayment(ctx, payment.UUID, payment.Amount, payment.TotalAmount, metaDataString, customer.UUID, *request.ExpiredAt)
	if err != nil {
		return nil, err
	}

	// update status
	payment.Status = paymentConstant.PAYMENT_STATUS_VOID
	payment.UpdatedAt = time.Now().UTC()
	err = s.paymentRepo.UpdatePaymentStatus(ctx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return payment, nil
}
