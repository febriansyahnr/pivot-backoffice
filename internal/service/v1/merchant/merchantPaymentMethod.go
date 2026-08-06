package merchant

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// EnableAllPaymentMethod enables all payment methods for a given merchant.
// This function is intended to be used in the staging environment only.
// It retrieves all payment methods of a specific category and enables them
// for the provided merchant. If the payment method is of type QRIS, it also
// duplicates the QR registration for the merchant.
// in this func also set all payment method to be approved even though its manual activation
func (s *MerchantService) EnableAllPaymentMethod(ctx context.Context, merchant *merchantModel.Merchant) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/EnableAllPaymentMethod")
	defer segment.End()

	if s.config.Environment != constant.EnvironmentStaging {
		return nil
	}

	var (
		qrisAlreadyEnabled bool
	)

	paymentMethods, err := s.paymentMethod.GetAllPaymentMethodByCategory(ctx, constant.TypePayment)
	if err != nil {
		return err
	}

	for _, pm := range paymentMethods {
		if pm.Type == paymentConstant.PAYMENT_METHOD_QRIS {
			// avoid duplicate registration
			if qrisAlreadyEnabled {
				continue
			}

			_, err := s.qrisSvc.DuplicateRegistration(ctx, &qris.DuplicateRegistrationReq{
				SourceMerchantId: s.config.UnifiedPaymentConfig.MasterQRDuplicationID,
				TargetMerchantId: merchant.UUID,
			})
			if err != nil {
				s.logger.Error(ctx, "failed to duplicate the qr registration", logger.Error(err))
				continue
			}

			qrisAlreadyEnabled = true
		}
		err := s.paymentMethod.UpsertPaymentMethodMerchantByIdAndMerchant(ctx, &paymentModel.PaymentMethodWithPivot{
			MerchantID:       merchant.UUID,
			IsActive:         true,
			PaymentMethod:    *pm,
			ChannelType:      constant.PaymentMethodChannelTypeAggregator,
			ActivationStatus: constant.PaymentMethodActivationStatusApproved,
		})
		if err != nil {
			s.logger.Error(ctx, "failed to enable payment method", logger.Error(err), logger.String("payment_method_id", pm.UUID), logger.String("merchant_id", merchant.UUID))
		}
	}

	return nil
}
