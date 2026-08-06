package paymentService

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (ps *PaymentService) FilterStaticVaList(ctx context.Context, opt paymentModel.StaticVaFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/FilterStaticVaList")
	defer segment.End()

	opt.Validate()

	result, err := ps.paymentRepo.FilterStaticVaList(ctx, opt)
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error filtering static VA list for merchant %s", opt.MerchantID), logger.Error(err))
		return nil, err
	}

	ps.logger.Info(ctx, fmt.Sprintf("successfully retrieved static VA list for merchant %s with %d items", opt.MerchantID, len(result.Data.([]paymentModel.StaticVaListResponse))))

	return result, nil
}

func (ps *PaymentService) GetStaticVaDetail(ctx context.Context, opt paymentModel.StaticVaDetailRequest) (*paymentModel.StaticVaDetailResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetStaticVaDetail")
	defer segment.End()

	if err := validateStaticVaParams(opt.PaymentID, opt.MerchantID); err != nil {
		return nil, err
	}

	result, err := ps.paymentRepo.GetStaticVaDetail(ctx, opt)
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error getting static VA detail for payment %s, merchant %s", opt.PaymentID, opt.MerchantID), logger.Error(err))
		return nil, err
	}

	return result, nil
}

func (ps *PaymentService) GetStaticVaTransactions(ctx context.Context, opt paymentModel.StaticVaTransactionFilterRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetStaticVaTransactions")
	defer segment.End()

	opt.Validate()

	if err := validateStaticVaParams(opt.PaymentID, opt.MerchantID); err != nil {
		return nil, err
	}

	result, err := ps.paymentRepo.GetStaticVaTransactions(ctx, opt)
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error getting static VA transactions for payment %s, merchant %s", opt.PaymentID, opt.MerchantID), logger.Error(err))
		return nil, err
	}

	ps.logger.Info(ctx, fmt.Sprintf("successfully retrieved static VA transactions for payment %s, merchant %s with %d items", opt.PaymentID, opt.MerchantID, len(result.Data.([]paymentModel.StaticVaTransactionItem))))

	return result, nil
}

func (ps *PaymentService) DeactivateStaticVa(ctx context.Context, paymentID string, merchantID string, request paymentModel.StaticVaUpdateStatusRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/DeactivateStaticVa")
	defer segment.End()

	if err := validateStaticVaParams(paymentID, merchantID); err != nil {
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
		return fmt.Errorf("static VA can only be deactivated if currently ACTIVE, current status: %s", currentPayment.Status)
	}

	// get processorReferenceId from Metadata.methodDetail.processorReferenceId
	var (
		processorReferenceId   string
		unifiedPaymentMetadata unifiedPaymentModel.MetadataUnifiedPayment
	)

	metadataB, _ := json.Marshal(currentPayment.Metadata)
	_ = json.Unmarshal(metadataB, &unifiedPaymentMetadata)
	if unifiedPaymentMetadata.MethodDetail == nil || unifiedPaymentMetadata.MethodDetail.ProcessorReferenceID == "" {
		return fmt.Errorf("error asserting metadata for payment %s, merchant %s", paymentID, merchantID)
	}

	processorReferenceId = unifiedPaymentMetadata.MethodDetail.ProcessorReferenceID

	// Call snap core FIRST to delete VA - only proceed with DB update if this succeeds
	_, err = ps.snapCoreRepo.DeleteVirtualAccount(ctx, &snapCoreModel.DeleteVirtualAccountRequest{
		Number: *currentPayment.ProcessorReferenceNumber,
		UUID:   processorReferenceId,
	})

	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("snap-core error deleting virtual account for payment %s, merchant %s", paymentID, merchantID), logger.Error(err))
		return err
	}

	// Only update DB status if snap core succeeds
	err = ps.paymentRepo.UpdatePaymentStatus(ctx, paymentID, merchantID, request.Status, time.Now())
	if err != nil {
		ps.logger.Error(ctx, fmt.Sprintf("error deactivating static VA for payment %s, merchant %s", paymentID, merchantID), logger.Error(err))
		return err
	}

	ps.logger.Info(ctx, fmt.Sprintf("successfully deactivated static VA for payment %s, merchant %s", paymentID, merchantID))

	return nil
}

func validateStaticVaParams(paymentID, merchantID string) error {
	if paymentID == "" {
		return fmt.Errorf("payment ID is required")
	}
	if merchantID == "" {
		return fmt.Errorf("merchant ID is required")
	}
	return nil
}
