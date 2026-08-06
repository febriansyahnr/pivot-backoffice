package unifiedPaymentService

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) PrepareRecurringPaymentRequest(ctx context.Context, request *model.CreateUnifiedPaymentSessionRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/PrepareRecurringPaymentRequest")
	defer segment.End()

	if request.RecurringID == "" {
		return nil
	}

	recurringContract, err := s.recurringContractRepo.GetDetailByID(ctx, request.MerchantID, request.RecurringID)
	if err != nil {
		s.logger.Error(ctx, "Failed when retrieving recurring payment contract", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if recurringContract == nil {
		return constant.NewErrResourceNotFound("recurring payment contract", request.RecurringID)

	} else if recurringContract.IsFirstAuthorization() && !request.InitiateFirstAuthorization {
		return pkgErrs.New(response.HttpErrRequest, fmt.Errorf("%s", "The first authorization must be completed before the next subsequent transaction"))

	} else if recurringContract.Status == constant.RecurringContractStatusInactive {
		return pkgErrs.New(response.HttpErrRequest, fmt.Errorf("%s", "The recurring payment contract is no longer active"))
	}

	// For subsequent recurring payments, the payment method must use the reference specified in the contract.
	if recurringContract.Status == constant.RecurringContractStatusActive && !request.InitiateFirstAuthorization {
		request.PaymentMethod = &model.PaymentMethod{
			Type: recurringContract.GetUnifiedPaymentMethodType(),
			CardPaymentMethodDetail: &model.CardPaymentMethodDetail{ // Currently, recurring payments are only available for card payment methods.
				Token: util.ValueOfPtr(recurringContract.PaymentTokenID),
			},
		}
		request.PaymentMethodOptions = model.PaymentMethodOptions{
			Card: &model.PaymentMethodOptionCard{
				ThreeDsMethod: constant.CardThreeDsMethodNever,
			},
		}

	} else if request.InitiateFirstAuthorization {
		if request.PaymentMethodOptions.Card == nil {
			request.PaymentMethodOptions.Card = &model.PaymentMethodOptionCard{}
		}
		request.SaveForFutureUse = util.ValueToPtr(true)
		request.PaymentMethodOptions.Card.ThreeDsMethod = constant.CardThreeDsMethodChallenge
	}

	// When the payment method is changed, the authorization will always use the ONE_DOLLAR authorization method.
	if recurringContract.Status == constant.RecurringContractStatusActive && request.InitiateFirstAuthorization {
		recurringContract.AuthMethod = constant.RecurringContractAuthMethodOneDollar
	}
	minAmountPerPayment, maxAmountPerPayment := recurringContract.GetMinMaxAmountPerPayment()
	billingCycleAmount := recurringContract.GetRecurringAmountForBillingCycle(request.InitiateFirstAuthorization)

	if request.Amount.Value != 0 && request.Amount.Value != billingCycleAmount {
		return pkgErrs.New(response.HttpErrRequest, fmt.Errorf("%s", "The transaction amount does not match the billing cycle calculation"))
	}

	now := time.Now().UTC()

	maxProcessDuration := constant.RecurringPaymentMaxProcessDuration
	if !request.ExpiryAt.IsZero() && now.Before(request.ExpiryAt) {
		maxProcessDuration = request.ExpiryAt.Sub(now)
	}
	processKey := fmt.Sprintf(constant.RecurringPaymentMutualExclusionKey, request.RecurringPaymentType(), request.RecurringID)

	if ok, err := s.redis.SetNX(ctx, processKey, true, maxProcessDuration).Result(); err != nil {
		s.logger.Error(ctx, "Failed to acquire recurring payment mutex lock", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser)

	} else if !ok {
		return constant.NewErrStringRequest(response.HttpErrConflict, constant.ErrCodeDuplicateError, "The recurring payment is currently being processed")
	}
	s.logger.Info(ctx, "Exclusive lock acquired for recurring ID "+request.RecurringID+" type "+request.RecurringPaymentType())

	request.CleanupPreparedRecurringPaymentLock = func(ctx context.Context) {
		s.logger.Info(ctx, "Cleanup prepared recurring payment for recurring ID "+request.RecurringID+" type "+request.RecurringPaymentType())

		if delErr := s.redis.Del(ctx, processKey).Err(); delErr != nil {
			s.logger.Error(ctx, "Failed to cleanup prepared recurring payment lock", logger.Error(delErr))
		}
	}

	if request.Amount.Value == 0 {
		request.Amount = model.Amount{
			Currency: constant.CurrencyIDR,
			Value:    billingCycleAmount,
		}
	}
	request.CustomerID = recurringContract.CustomerID
	request.RecurringBillingCycle = model.RecurringBillingCycle{
		Interval:               recurringContract.Billing.Interval,
		IntervalUnit:           recurringContract.Billing.IntervalUnit,
		ExpiryDate:             recurringContract.EndDate.In(asiaJakartaLoc).Format(time.DateOnly),
		MinDaysBetweenPayments: constant.RecurringMinDaysBetweenPayments(recurringContract.Billing.IntervalUnit),
		MinAmountPerPayment:    minAmountPerPayment,
		MaxAmountPerPayment:    maxAmountPerPayment,
	}
	// The billing cycle count is incremented only during the first or subsequent recurring payments.
	if !request.InitiateFirstAuthorization || recurringContract.AuthMethod == constant.RecurringContractAuthMethodFirstPayment {
		request.RecurringBillingCycle.Count = recurringContract.Billing.Count + 1
	}
	request.RecurringStatus = recurringContract.Status
	if request.InitiateFirstAuthorization {
		request.FirstAuthorizationMethod = recurringContract.AuthMethod
	}
	request.FirstAuthorizationOrderID = recurringContract.ProcessorOrderID

	// When the first authorization uses the ONE_DOLLAR method, the process will follow the Auth & Capture flow.
	if request.FirstAuthorizationMethod == constant.RecurringContractAuthMethodOneDollar {
		request.PaymentMethodOptions.Card.CaptureMethod = constant.UnifiedPaymentCardCaptureMethodManual
	}
	return nil
}
