package creditcard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *CreditCardService) GetPaymentById(
	ctx context.Context,
	merchantID, uuid string,
) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/creditcard/GetPaymentById")
	defer segment.End()

	payment, err := c.paymentRepo.GetPaymentById(ctx, uuid)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}

	if payment == nil {
		key := fmt.Sprintf(constant.TemporaryPaymentRecordCacheKey, merchantID, uuid)
		result, err := c.redis.Get(ctx, key).Result()
		if err == nil {
			err = json.Unmarshal([]byte(result), &payment)
			if err != nil {
				c.logger.Error(ctx, "error when unmarshal payment data from redis", logger.Error(err), logger.String("key", key), logger.Any("result", result))
			}
		} else {
			if err != redisExt.ErrNil {
				c.logger.Error(ctx, "error when get payment from redis", logger.Error(err), logger.String("key", key))
			}
		}

		if payment == nil {
			err = constant.ErrCreditcardNotFound
			c.logger.Error(ctx, err.Error(), logger.Error(err))
			return nil, pkgErrors.New(response.HttpErrNotFound, err)
		}
	}

	// Validate merchant ownership (parent or derived)
	if payment.MerchantID != merchantID && payment.GetOnBehalfParentID() != merchantID {
		return nil, pkgErrors.New(response.HttpErrUnauthorized, errors.New("merchant not authorized"))
	}

	if payment.ExpiredAt.Before(time.Now().UTC()) &&
		payment.Status == constant.StatusPending {
		err = creditcardModel.UpdateCreditcardMetaData(payment.Metadata, nil, nil, nil, constant.CreditCardProcessorStatusExpired)
		if err != nil {
			c.logger.Error(ctx, constant.ErrWhenUpdateCreditcardMetaData.Error(), logger.Error(err))
			return nil, err
		}

		payment.Status = constant.StatusFailed
		payment.UpdatedAt = time.Now().UTC()
		err = c.paymentRepo.UpdatePaymentData(ctx,
			payment.ToDTO())
		if err != nil {
			return nil, pkgErrors.New(response.HttpErrDatabase, err)
		}
	}

	return payment, nil
}

func (c *CreditCardService) GetTransactionList(
	ctx context.Context,
	request *creditcardModel.GetTransactionListRequest,
) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/creditcard/GetTransactionList")
	defer segment.End()

	response, err := c.creditcardCoreProcessorRepo.GetTransactionList(ctx, request.ToCreditcardCoreGetTransactionListRequest())
	if err != nil {
		return nil, err
	}

	return creditcardModel.ToCreditcardCoreGetTransactionListResponse(response), nil
}

func (s *CreditCardService) FindPaymentResponseById(ctx context.Context, id, merchantID string) (*paymentModel.PaymentResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/creditcard/FindPaymentResponseById")
	defer segment.End()

	var (
		paymentResponse        *paymentModel.PaymentResponse
		paymentItemsResponse   []paymentModel.PaymentResponseItem
		customer               *customerModel.Customer
		paymentRequestCustomer paymentModel.PaymentRequestCustomer
		paymentReferenceId     = ""
		isUnifiedPayment       bool
	)

	// Get Payment Method by category
	payment, err := s.paymentRepo.GetPaymentById(ctx, id)
	if err != nil {
		s.logger.Error(ctx, "error when get payment data by id", logger.Error(err))
		return nil, err
	}

	if payment == nil {
		err := pkgErrors.New(response.HttpErrNotFound, constant.ErrPaymentNotFound)
		s.logger.Error(ctx, "payment not found", logger.String("id", id))
		return nil, err
	}

	// Get payment_method by id
	paymentMethod, err := s.paymentMethodRepo.GetPaymentMethodById(ctx, payment.PaymentMethodID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment method data by id", logger.Error(err))
		return nil, err
	}

	// Get customer by id
	if payment.CustomerID != "" {
		customer, err = s.customerRepo.FindCustomerById(ctx, payment.CustomerID)
		if err != nil {
			s.logger.Error(ctx, "error when get customer data by id", logger.Error(err))
			return nil, err
		}

		// convert paymentRequestCustomer
		paymentRequestCustomer = paymentModel.PaymentRequestCustomer{
			CustomerID: customer.UUID,
			Name:       customerModel.FirstNameAndLastNameToFullName(customer.FirstName, customer.LastName),
			Email:      customer.Email,
			Phone:      customer.PhoneNumber,
			Metadata:   nil,
		}
	}

	// check if customer merchant id is same with request merchant id
	if payment.MerchantID != merchantID {
		s.logger.Error(ctx, "merchant id not match", logger.Error(fmt.Errorf("payment not found, merchant id not match")))
		return nil, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("payment not found"))
	}

	// Get payment items
	paymentItems, err := s.paymentRepo.GetPaymentItemsByPaymentId(ctx, payment.UUID)
	if err != nil {
		s.logger.Error(ctx, "error when get payment items data by payment id", logger.Error(err))
		return nil, err
	}

	for _, item := range paymentItems {
		paymentItemsResponse = append(paymentItemsResponse, *item.ToPaymentResponseItem())
	}

	// create amount
	amount := paymentModel.Amount{
		Value:    payment.Amount,
		Currency: payment.Currency,
	}

	// unmarshal payment.Metadata map[string]any to createVirtualAccountResponseData
	if payment.Metadata != nil {
		jsonData, errMarshal := json.Marshal(payment.Metadata)
		if errMarshal != nil {
			s.logger.Error(ctx, "error when marshal payment metadata", logger.Error(errMarshal))
			return nil, errMarshal
		}

		if err := json.Unmarshal(jsonData, &struct {
			IsUnifiedPayment *bool `json:"isUnifiedPayment"`
		}{
			IsUnifiedPayment: &isUnifiedPayment,
		}); err != nil {
			s.logger.Error(ctx, "error when unmarshal payment metadata", logger.Error(err))
		}
	}

	if payment.ReferenceID != nil {
		paymentReferenceId = *payment.ReferenceID
	}

	paymentResponse = &paymentModel.PaymentResponse{
		UUID:             payment.UUID,
		MerchantID:       payment.MerchantID,
		ReferenceID:      paymentReferenceId,
		Customer:         &paymentRequestCustomer,
		Status:           payment.Status,
		TotalAmount:      &amount,
		PaymentMethod:    paymentMethod.Type,
		PaymentItems:     &paymentItemsResponse,
		LastUpdateDate:   &payment.UpdatedAt,
		CreatedAt:        payment.CreatedAt,
		ExpiredAt:        payment.ExpiredAt,
		PaymentURL:       payment.PaymentURL,
		IsUnifiedPayment: isUnifiedPayment,
	}

	return paymentResponse, nil
}

func (s *CreditCardService) GetStoredCardByCustomerID(
	ctx context.Context,
	merchantID, customerID string,
) ([]*unifiedPaymentModel.CustomerPaymentMethodResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/creditcard/GetStoredCardByCustomerID")
	defer span.End()

	customer, err := s.customerRepo.GetCustomerById(ctx, customerID, merchantID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, constant.ErrDatabaseGetCustomer)
	}
	if customer == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrCustomerNotFound)
	}
	customerResp := customer.ToUnifiedPaymentCustomerResponse()
	cardPaymentMethods := []*unifiedPaymentModel.CustomerPaymentMethodResponse{}
	for _, method := range customerResp.StoredPaymentMethods {
		if method.PaymentMethod == constant.UnifiedPaymentMethodCard {
			cardPaymentMethods = append(cardPaymentMethods, method)
		}
	}

	return cardPaymentMethods, nil
}

func (s *CreditCardService) GetCardEncryptionPublicKey(ctx context.Context, merchantID string) ([]byte, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/creditcard/GetCardEncryptionPublicKey")
	defer span.End()

	return s.creditcardCoreProcessorRepo.GetCardEncryptionPublicKey(ctx, merchantID)
}
