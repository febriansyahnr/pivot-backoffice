package unifiedPaymentService

import (
	"context"
	"encoding/json"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/outbound"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/ewallet"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *UnifiedPaymentService) InquiryEWalletPayment(ctx context.Context, payment *paymentModel.Payment) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/unifiedPayment/InquiryEWalletPayment")
	defer segment.End()

	var ewalletPayment = *payment

	paymentDataB, _ := json.Marshal(payment.Metadata)
	unifiedPaymentMetadata := unifiedPaymentModel.MetadataUnifiedPayment{
		IsUnifiedPaymentV2: false,
	}
	_ = json.Unmarshal(paymentDataB, &unifiedPaymentMetadata)

	if !slices.Contains([]string{ // Status is final (SUCCESS, FAILED, EXPIRED)
		constant.UnifiedPaymentSessionStatusRequireAction,
		constant.UnifiedPaymentSessionStatusProcessing,
	}, payment.Status) {
		s.logger.Info(ctx, "e-wallet payment status is not processing or require action. skip inquiry", logger.String("paymentID", payment.UUID), logger.String("status", payment.Status), logger.String("acquirer", unifiedPaymentMetadata.PaymentMethodOptions.Ewallet.Channel))
		return payment, nil
	}

	paymentLedger, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, constant.ErrUpdatePaymentLedger)
	}
	if paymentLedger == nil {
		s.logger.Error(ctx, "payment ledger not found", logger.String("paymentID", payment.UUID))
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrLedgerDetailNotFound)
	}

	if s.config.Environment != constant.EnvironmentProduction && s.isEWalletPaymentSimulationFlowEnabled(ctx, payment.MerchantID) {
		ctx = context.WithValue(ctx, constant.CtxPaymentSimulationMode, strconv.FormatBool(true))
		token := ""
		if parsedURL, errParsed := url.Parse(payment.PaymentURL); errParsed == nil {
			token = parsedURL.Query().Get("token")
		}
		ctx = context.WithValue(ctx, constant.CtxPaymentSimulationToken, token)
	}

	traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	ctx = context.WithValue(ctx, constant.CtxClientReqKey, &outbound.Client{
		RequestId:   traceID,
		OriginId:    payment.UUID,
		ReferenceId: payment.MerchantID,
	})
	snapResp, err := s.snapCoreRepo.InquiryStatusEWalletPayment(ctx, &ewallet.EWalletInquiryStatusRequest{
		TransactionID: paymentLedger.ProcessorReferenceId,
	})
	if err != nil {
		s.logger.Error(ctx, "error when inquiry ewallet payment status", logger.Error(err), logger.String("paymentID", payment.UUID))
		return nil, err
	}
	if snapResp != nil {
		ewalletPayment.InquiryDetail = &unifiedPaymentModel.InquiryDetail{}
		switch snapResp.LatestTransactionStatus {
		case constant.SnapLatestTransactionStatusCancelled,
			constant.SnapLatestTransactionStatusNotFound,
			constant.SnapLatestTransactionStatusFailed:
			ewalletPayment.InquiryDetail.Status = constant.StatusFailed
		case constant.SnapLatestTransactionStatusSuccess:
			ewalletPayment.InquiryDetail.Status = constant.StatusSuccess
		default:
			ewalletPayment.InquiryDetail.Status = constant.StatusPending
		}
	}

	return &ewalletPayment, nil
}

func (s *UnifiedPaymentService) UpdateEWalletPaymentSession(ctx context.Context, paymentID string) (*paymentModel.Payment, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/unifiedPayment/UpdateEWalletPaymentSession")
	defer segment.End()

	payment, err := s.paymentRepo.GetPaymentById(ctx, paymentID)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if payment == nil {
		s.logger.Info(ctx, constant.ErrPaymentNotFound.Error(), logger.String("paymentId", paymentID))
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, constant.ErrPaymentNotFound)
	}

	paymentLedger, err := s.accountTransactionRepo.FindByReference(ctx, payment.UUID, constant.TypePayment)
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, constant.ErrGetPaymentLedger)
	}

	if payment.Status != constant.UnifiedPaymentSessionStatusRequireAction {
		s.logger.Info(ctx, "payment status is not in require action status", logger.String("paymentId", payment.UUID), logger.String("status", payment.Status))
		return payment, nil
	}

	payment.Status = constant.UnifiedPaymentSessionStatusProcessing
	payment.UpdatedAt = time.Now().UTC()
	err = s.paymentRepo.UpdatePaymentStatus(ctx, payment.UUID, payment.MerchantID, payment.Status, payment.UpdatedAt)
	if err != nil {
		s.logger.Error(ctx, "failed to update payment status to processing",
			logger.String("paymentId", payment.UUID),
			logger.String("merchantId", payment.MerchantID),
			logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, constant.ErrUpdatePayment)
	}

	err = s.accountTransactionRepo.UpdatePaymentTransactionStatusAndMetadataByID(
		ctx,
		orchestrator_model.UpdatePaymentTransactionRequest{
			LedgerId:  paymentLedger.UUID.String(),
			Status:    constant.StatusPending,
			UpdatedAt: time.Now().UTC(),
		},
		orchestrator_model.MetadataPayment[any]{
			ChargeStatus: constant.ChargeStatusProcessing,
		},
	)
	if err != nil {
		s.logger.Error(ctx, "error updating payment ledger data", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, constant.ErrUpdatePayment)
	}

	s.SendCallback(ctx, payment)

	return payment, nil
}
