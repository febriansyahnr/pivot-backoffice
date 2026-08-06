package creditcard

import (
	"context"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

// RemoveCardTokenization is a function used to remove card tokenization from a customer
func (s *CreditCardService) RemoveCardTokenization(ctx context.Context, request unifiedPaymentModel.RemoveCardTokenizationRequest) error {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/creditcard/RemoveCardTokenization")
	defer span.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, request.PaymentID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get payment details based on payment id while removing card tokenization", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if payment == nil {
		return pkgErrors.New(response.HttpErrNotFound, constant.ErrPaymentNotFound)

	} else if payment.MerchantID != request.MerchantID || payment.CustomerID != request.CustomerID {
		return pkgErrors.New(response.HttpErrForbidden, errors.New("action not permitted"))
	}

	customer, err := s.customerRepo.FindCustomerById(ctx, payment.CustomerID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get customer details based on customer id while removing card tokenization", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	paymentMethods, err := util.ConvertToStruct[[]*unifiedPaymentModel.CustomerPaymentMethodResponse](customer.Metadata["paymentMethods"])
	if err != nil {
		s.logger.Error(ctx, "Failed while convert map value to struct", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}

	if err = s.customerRepo.RemovePaymentMethodFromCustomerByIDAndTokenID(ctx, payment.CustomerID, request.TokenID, paymentMethods); err != nil {
		if errors.Is(err, constant.ErrDataNotFound) {
			return pkgErrors.New(response.HttpErrNotFound, fmt.Errorf("payment method with token id %s not found", request.TokenID))
		}
		s.logger.Error(ctx, "Failed to remove payment method from customer", logger.Error(err))
		return pkgErrors.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)
	}
	return nil
}
