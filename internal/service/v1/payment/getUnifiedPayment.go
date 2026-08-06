package paymentService

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentService) GetPaymentByReferenceId(ctx context.Context, referenceId string, merchantID string) (*paymentModel.UnifiedPaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetPaymentByReferenceId")
	defer segment.End()

	var (
		paymentResponse        *paymentModel.UnifiedPaymentResponse
		paymentItemsResponse   []paymentModel.PaymentResponseItem
		customer               *customerModel.Customer
		paymentRequestCustomer paymentModel.PaymentRequestCustomer
		creditAmount           float64
		creditCurrency         = constant.CurrencyIDR
	)

	payment, err := s.paymentRepo.GetPaymentByMerchantAndReferenceId(ctx, merchantID, referenceId)
	if err != nil {
		s.logger.Error(ctx, "error when get payment data by reference id", logger.Error(err))
		return nil, err
	}

	if payment == nil {
		err := pkgErrors.New(response.HttpErrNotFound, errors.New("payment not found"))
		s.logger.Error(ctx, "payment not found", logger.String("referenceId", referenceId))
		return nil, err
	}

	if payment.ExpiredAt != nil && time.Now().After(*payment.ExpiredAt) &&
		(payment.Status == paymentConstant.UnifiedPaymentStatusWaitingForPayment || payment.Status == constant.UnifiedPaymentSessionStatusRequireAction) {
		err := s.paymentRepo.UpdatePaymentStatus(ctx, payment.UUID, payment.MerchantID, paymentConstant.UnifiedPaymentStatusExpired, time.Now())
		if err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}
		payment.Status = paymentConstant.UnifiedPaymentStatusExpired

		// Record status history
		s.recordPaymentExpired(ctx, payment.UUID, constant.StatusHistoryActorSystem)
	}

	paymentMethod, err := s.paymentMethodRepo.GetPaymentMethodById(ctx, payment.PaymentMethodID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment method data by id", logger.Error(err))
		return nil, err
	}

	if payment.CustomerID != "" {
		customer, err = s.customerRepo.FindCustomerById(ctx, payment.CustomerID)
		if err != nil {
			s.logger.Error(ctx, "error when get customer data by id", logger.Error(err))
			return nil, err
		}

		paymentRequestCustomer = paymentModel.PaymentRequestCustomer{
			CustomerID: customer.UUID,
			Name:       customerModel.FirstNameAndLastNameToFullName(customer.FirstName, customer.LastName),
			Email:      customer.Email,
			Phone:      customer.PhoneNumber,
			Metadata:   nil,
		}
	}

	if payment.MerchantID != merchantID {
		s.logger.Error(ctx, "merchant id not match", logger.Error(fmt.Errorf("payment not found, merchant id not match")))
		return nil, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("payment not found"))
	}

	paymentItems, err := s.paymentRepo.GetPaymentItemsByPaymentId(ctx, payment.UUID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment items data by payment id", logger.Error(err))
		return nil, err
	}

	for _, item := range paymentItems {
		paymentItemsResponse = append(paymentItemsResponse, *item.ToPaymentResponseItem())
	}

	accountTransaction, err := s.orchestratorSvc.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		return nil, err
	}

	if accountTransaction != nil && accountTransaction.Status == constant.StatusSuccess {
		creditAmount = accountTransaction.Credit
		creditCurrency = accountTransaction.Currency
	}

	paymentAmount, _ := payment.Amount.Float64()

	ledgerMetadata := []byte{}
	if accountTransaction != nil && accountTransaction.AdditionalInfo.Valid {
		ledgerMetadata = accountTransaction.AdditionalInfo.JSONText
	}

	paymentResponse = &paymentModel.UnifiedPaymentResponse{
		UUID:        payment.UUID,
		MerchantID:  payment.MerchantID,
		ReferenceID: referenceId,
		Customer:    &paymentRequestCustomer,
		Status:      constant.MapUnifiedPaymentStatusToV1(payment.Status),
		Amount: commonModel.Amount{
			Currency: payment.Currency,
			Value:    strconv.FormatFloat(paymentAmount, 'f', 2, 64),
		},
		PaidAmount: commonModel.Amount{
			Currency: creditCurrency,
			Value:    strconv.FormatFloat(creditAmount, 'f', 2, 64),
		},
		PaymentMethod:     paymentMethod.Name,
		PaymentMethodType: s.GetPaymentSubType(ctx, payment.PaymentMethod.Type, payment.Metadata),
		TypeDetail:        s.GetPaymentTypeDetail(ctx, payment.PaymentMethod.Type, payment.Metadata, ledgerMetadata),
		PaymentItems:      &paymentItemsResponse,
		LastUpdateDate:    &payment.UpdatedAt,
		ExpiryAt:          payment.ExpiredAt,
	}

	return paymentResponse, nil
}
