package paymentMethodService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *PaymentMethodService) Activate(ctx context.Context, request *paymentModel.PaymentMethodWithPivot) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/Activate")
	defer segment.End()

	merchant, err := s.merchantSvc.FindMerchantByID(ctx, request.MerchantID)
	if err != nil {
		return err

	} else if merchant == nil {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantNotFound)

	} else if merchant.ParentID.Valid && merchant.KYCStatus.String != constant.KYCStatusApproved {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrMerchantShouldKYC)
	}

	paymentMethod, err := s.FindPaymentMethodByIdAndMerchant(ctx, request.UUID, request.MerchantID)
	if err != nil {
		return err

	} else if paymentMethod.IsActive {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentMethodAlreadyActive)
	}

	switch paymentMethod.Type {
	case paymentConstant.PAYMENT_METHOD_QRIS:
		return s.activateQris(ctx, merchant, paymentMethod)

	case paymentConstant.PAYMENT_METHOD_INSTALLMENT:
		return s.activateInstallment(ctx, merchant, paymentMethod)

	default:
		return s.activateDefault(ctx, paymentMethod)
	}
}

func (s *PaymentMethodService) activateDefault(ctx context.Context, paymentMethodMerchant *paymentModel.PaymentMethodWithPivot) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/activateDefault")
	defer segment.End()

	paymentMethodMerchant.IsActive = true
	if err := s.paymentMethodRepo.UpsertPaymentMethodMerchantByIdAndMerchant(ctx, paymentMethodMerchant); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	return nil
}

func (s *PaymentMethodService) activateQris(ctx context.Context, merchant *merchant.Merchant, paymentMethodMerchant *paymentModel.PaymentMethodWithPivot) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/activateQris")
	defer segment.End()

	// Find status in QR registration, if status SUCCESS then continue, else throw error.
	qrRegistration, errFindQr := s.qrisSvc.FindQrRegistrationByExternalIDAndAcquirer(ctx, merchant.ExternalId, paymentMethodMerchant.Acquirer)
	if errFindQr != nil {
		return errFindQr
	}
	if qrRegistration.Status != constant.QrRegistrationStatusSuccess {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrQrRegistrationIsNotCompleted)
	}

	// Upsert QRIS
	paymentMethodMerchant.IsActive = true
	if err := s.paymentMethodRepo.UpsertPaymentMethodMerchantByIdAndMerchant(ctx, paymentMethodMerchant); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	// Sync to SnapCore after activation
	err := s.syncQrisRegistration(ctx, qrRegistration, snapCoreModel.SyncRegistrationOption{
		MerchantID: merchant.UUID,
		StoreIDs:   []string{}, // Empty for now, can be populated if needed
		IsActive:   &paymentMethodMerchant.IsActive,
	})
	if err != nil {
		// Just log the error, don't fail activation
		s.logger.Error(ctx, "[activateQris] Failed to sync QRIS registration to SnapCore", logger.Error(err))
	}

	return nil
}

func (s *PaymentMethodService) activateInstallment(ctx context.Context, merchant *merchant.Merchant, paymentMethodMerchant *paymentModel.PaymentMethodWithPivot) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/paymentMethod/activateInstallment")
	defer segment.End()

	if paymentMethodMerchant.Subtype == constant.InstallmentPlanPaymentMethodCard {
		// Check if card payment method is active first
		cardPaymentMethods, err := s.paymentMethodRepo.GetListPaymentMethodMerchant(ctx, &paymentModel.GetPaymentMethodFilterRequest{
			MerchantID: merchant.UUID,
			Type:       paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			Status:     constant.PaymentMethodGeneralStatusActive,
		})
		if err != nil {
			return pkgErrors.New(response.HttpErrDatabase, err)
		}
		if len(cardPaymentMethods) < 1 {
			return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrDependentCardPaymentMethodNotActive)
		}
	}
	if paymentMethodMerchant.MerchantConfigObj.PartnerConfig.Installment == nil ||
		(paymentMethodMerchant.MerchantConfigObj.PartnerConfig.Installment != nil && len(paymentMethodMerchant.MerchantConfigObj.PartnerConfig.Installment.InstallmentPlanIDs) <= 0) {
		return pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrCardInstallmentNotConfigured)
	}

	// Notes: Assumes installment plan is already validate during config process. Simplify process.
	paymentMethodMerchant.IsActive = true
	if err := s.paymentMethodRepo.UpsertPaymentMethodMerchantByIdAndMerchant(ctx, paymentMethodMerchant); err != nil {
		return pkgErrors.New(response.HttpErrDatabase, err)
	}

	return nil
}
