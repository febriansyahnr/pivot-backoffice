package paymentService

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (ps *PaymentService) FilterStaticQrisList(ctx context.Context, opt paymentModel.StaticQrisFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/FilterStaticQrisList")
	defer segment.End()

	opt.Validate()

	result, err := ps.paymentRepo.FilterStaticQrisList(ctx, opt)
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error filtering static QRIS list for merchant %s", opt.MerchantID), logger.Error(err))
		return nil, err
	}

	ps.logger.Info(ctx, fmt.Sprintf("successfully retrieved static QRIS list for merchant %s with %d items", opt.MerchantID, len(result.Data.([]paymentModel.StaticQrisListResponse))))

	return result, nil
}

func (ps *PaymentService) GetStaticQrisDetail(ctx context.Context, opt paymentModel.StaticQrisDetailRequest) (*paymentModel.StaticQrisDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetStaticQrisDetail")
	defer segment.End()
	if err := validateStaticQrisParams(opt.PaymentID, opt.MerchantID); err != nil {
		return nil, err
	}

	result, err := ps.paymentRepo.GetStaticQrisDetail(ctx, opt)
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error getting static QRIS detail for payment %s, merchant %s", opt.PaymentID, opt.MerchantID), logger.Error(err))
		return nil, err
	}

	return result, nil
}

func (ps *PaymentService) GetStaticQrisTransactions(ctx context.Context, opt paymentModel.StaticQrisTransactionFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetStaticQrisTransactions")
	defer segment.End()

	opt.Validate()

	if err := validateStaticQrisParams(opt.PaymentID, opt.MerchantID); err != nil {
		return nil, err
	}

	result, err := ps.paymentRepo.GetStaticQrisTransactions(ctx, opt)
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error getting static QRIS transactions for payment %s, merchant %s", opt.PaymentID, opt.MerchantID), logger.Error(err))
		return nil, err
	}

	ps.logger.Info(ctx, fmt.Sprintf("successfully retrieved static QRIS transactions for payment %s, merchant %s with %d items", opt.PaymentID, opt.MerchantID, len(result.Data.([]paymentModel.StaticQrisTransactionItem))))

	return result, nil
}

func (ps *PaymentService) DeactivateStaticQris(ctx context.Context, paymentID string, merchantID string, request paymentModel.StaticQrisUpdateStatusRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/DeactivateStaticQris")
	defer segment.End()

	if err := validateStaticQrisParams(paymentID, merchantID); err != nil {
		return err
	}

	// Get current payment to validate it's currently ACTIVE
	currentPayment, err := ps.paymentRepo.GetPaymentByIdAndMerchantId(ctx, paymentID, merchantID)
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error getting payment %s for merchant %s", paymentID, merchantID), logger.Error(err))
		return err
	}

	// Validate that payment is currently ACTIVE
	if currentPayment.Status != constant.QrStatusActive {
		return fmt.Errorf("static QRIS can only be deactivated if currently ACTIVE, current status: %s", currentPayment.Status)
	}

	err = ps.paymentRepo.UpdatePaymentStatus(ctx, paymentID, merchantID, request.Status, time.Now())
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error deactivating static QRIS for payment %s, merchant %s", paymentID, merchantID), logger.Error(err))
		return err
	}

	ps.logger.Info(ctx, fmt.Sprintf("successfully deactivated static QRIS for payment %s, merchant %s", paymentID, merchantID))

	return nil
}

func (ps *PaymentService) GetFirstActiveStaticQris(ctx context.Context, merchantID string, partnerReferenceNo string) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetFirstActiveStaticQris")
	defer segment.End()

	if merchantID == "" {
		return nil, fmt.Errorf("merchant ID is required")
	}

	payment, err := ps.paymentRepo.GetFirstActiveStaticQrisByMerchant(ctx, merchantID, partnerReferenceNo)
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error getting first active static QRIS for merchant %s", merchantID), logger.Error(err))
		return nil, err
	}

	if payment == nil {
		errorMsg := fmt.Sprintf("no active static QRIS found for merchant %s", merchantID)
		if partnerReferenceNo != "" {
			errorMsg = fmt.Sprintf("no active static QRIS found for merchant %s with partnerReferenceNo %s", merchantID, partnerReferenceNo)
		}
		return nil, fmt.Errorf("%s", errorMsg)
	}

	return payment, nil
}

func (ps *PaymentService) GetMaxActiveStaticQRPerMerchant() int {
	if ps.config == nil || ps.config.UnifiedPaymentConfig.QrConfig == nil {
		return 0
	}
	return ps.config.UnifiedPaymentConfig.QrConfig.MaxActiveStaticQRPerMerchant
}

func validateStaticQrisParams(paymentID, merchantID string) error {
	if paymentID == "" {
		return fmt.Errorf("payment ID is required")
	}
	if merchantID == "" {
		return fmt.Errorf("merchant ID is required")
	}
	return nil
}
