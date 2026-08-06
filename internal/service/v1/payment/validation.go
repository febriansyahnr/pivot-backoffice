package paymentService

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PaymentService) ValidatePaymentExpiry(ctx context.Context, cmd paymentModel.PaymentRequestExpiryValidation) error {
	validationConfig := paymentModel.PaymentMethodExpiryConfig{}
	expiryRequest := time.Time{}
	if cmd.Method == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT && cmd.Request.VirtualAccount != nil {
		expiryRequest = util.ValueOfPtr(cmd.Request.VirtualAccount.ExpiredDate)
	}

	if cmd.Method == paymentConstant.PAYMENT_METHOD_QRIS && cmd.Request.Qris != nil {
		expiryRequest = time.Now().Add(time.Duration(cmd.Request.Qris.ValidityPeriod) * time.Second)
	}

	if expiryRequest.IsZero() {
		return nil
	}

	paymentMethodList := []string{
		paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		paymentConstant.PAYMENT_METHOD_QRIS,
	}

	if !slices.Contains(paymentMethodList, cmd.Method) {
		return nil
	}

	// set default config for virtual account
	if cmd.Method == paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT && s.config.UnifiedPaymentConfig.VirtualAccountConfig != nil {
		vaConfig := s.config.UnifiedPaymentConfig.VirtualAccountConfig
		validationConfig = paymentModel.PaymentMethodExpiryConfig{
			Duration: vaConfig.MaxExpiryDuration,
			Unit:     vaConfig.MaxExpiryDurationUnit,
		}
	}

	// set default config for qris
	if cmd.Method == paymentConstant.PAYMENT_METHOD_QRIS && s.config.UnifiedPaymentConfig.QrConfig != nil {
		qrConfig := s.config.UnifiedPaymentConfig.QrConfig
		validationConfig = paymentModel.PaymentMethodExpiryConfig{
			Duration: qrConfig.MaxExpiryDuration,
			Unit:     qrConfig.MaxExpiryDurationUnit,
		}
	}

	if s.config.UnifiedPaymentConfig.ExpiryConfig != nil && !s.config.UnifiedPaymentConfig.ExpiryConfig.ShouldValidateExpiry(cmd.MerchantID) {
		return nil
	}

	// replace config from database
	if cmd.PaymentMethod.PaymentMethod.ConfigObj != nil &&
		cmd.PaymentMethod.PaymentMethod.ConfigObj.ExpiryConfig.Unit != "" &&
		cmd.PaymentMethod.PaymentMethod.ConfigObj.ExpiryConfig.Duration > 0 {

		validationConfig = paymentModel.PaymentMethodExpiryConfig{
			Duration: cmd.PaymentMethod.PaymentMethod.ConfigObj.ExpiryConfig.Duration,
			Unit:     cmd.PaymentMethod.PaymentMethod.ConfigObj.ExpiryConfig.Unit,
		}
	}

	maxValidationTime := validationConfig.ToDateTime()
	// return error if expiry request is greater than max validation time
	if !maxValidationTime.IsZero() && expiryRequest.After(maxValidationTime) {
		if cmd.Request != nil && cmd.Request.IsSnap {
			expiryField := "expiredDate"
			if cmd.Request.PaymentMethod == paymentConstant.PAYMENT_METHOD_QRIS {
				expiryField = "validityPeriod"
			}
			return pkgErrors.New(response.SnapErrFieldFormat, fmt.Errorf(constant.InvalidMandatoryFieldSnapFmt, expiryField))
		}

		maxValidationTimeStr := fmt.Sprintf("%d %s", validationConfig.Duration, validationConfig.Unit)
		return pkgErrors.New(response.HttpErrRequest, fmt.Errorf(constant.ErrExceedMaxExpiryDate, cmd.Method, maxValidationTimeStr))
	}

	return nil
}
