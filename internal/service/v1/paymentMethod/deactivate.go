package paymentMethodService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentMethodService) Deactivate(ctx context.Context, request *paymentModel.PaymentMethodWithPivot) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/Deactivate")
	defer segment.End()

	merchant, err := s.merchantSvc.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return err
	}

	if merchant == nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)
	}

	if merchant.ParentID.Valid && merchant.KYCStatus.String != constant.KYCStatusApproved {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantShouldKYC)
	}

	paymentMethod, err := s.FindPaymentMethodByIdAndMerchant(ctx, request.UUID, request.MerchantID)
	if err != nil {
		return err
	}

	paymentMethod.IsActive = false
	if err := s.paymentMethodRepo.UpsertPaymentMethodMerchantByIdAndMerchant(ctx, paymentMethod); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Sync to SnapCore for QRIS payment methods
	if paymentMethod.Type == paymentConstant.PAYMENT_METHOD_QRIS {
		qrRegistration, errFindQr := s.qrisSvc.FindQrRegistrationByExternalIDAndAcquirer(ctx, merchant.ExternalId, paymentMethod.Acquirer)
		if errFindQr != nil {
			// Just log the error, don't fail deactivation
			s.logger.Error(ctx, "[Deactivate] Failed to find QRIS registration",
				logger.Error(errFindQr),
				logger.String("acquirer", paymentMethod.Acquirer))
		} else {
			err := s.syncQrisRegistration(ctx, qrRegistration, snapCoreModel.SyncRegistrationOption{
				MerchantID: merchant.UUID,
				StoreIDs:   []string{},
				IsActive:   &paymentMethod.IsActive,
			})
			if err != nil {
				// Just log the error, don't fail deactivation
				s.logger.Error(ctx, "[Deactivate] Failed to sync QRIS deactivation to SnapCore",
					logger.Error(err),
					logger.String("acquirer", paymentMethod.Acquirer))
			}
		}
	}

	return nil
}
