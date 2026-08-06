package unifiedPaymentService

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// GetChargeList retrieves a paginated list of charges based on the provided filter criteria.
// It starts an OpenTelemetry trace segment for performance monitoring and handles database errors
// by wrapping them with the appropriate HTTP error code.
func (s *UnifiedPaymentService) GetChargeList(ctx context.Context, request *unifiedPaymentModel.FilterChargeRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/GetChargeList")
	defer segment.End()

	charges, err := s.paymentRepo.GetChargeList(ctx, request)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	return charges, nil
}

// GetChargeDetail retrieves detailed information about a charge for a unified payment.
//
// It takes a context and a GetUnifiedPaymentChargeRequest containing ChargeID and MerchantID.
// The function fetches the charge from the payment repository using the ChargeID and validates
// that the charge belongs to the specified merchant.
func (s *UnifiedPaymentService) GetChargeDetail(ctx context.Context, request *unifiedPaymentModel.GetUnifiedPaymentChargeRequest) (*unifiedPaymentModel.ChargeResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v2/unifiedPayment/GetChargeDetail")
	defer segment.End()

	charge, err := s.paymentRepo.GetChargeByID(ctx, request.ChargeID)
	if err != nil {
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	if charge == nil {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrPaymentChargeNotFound)
	}

	if charge.MerchantID != request.MerchantID {
		return nil, pkgErr.New(response.HttpErrUnprocessableContent, constant.ErrMerchantIsNotMatch)
	}

	payment, err := s.paymentRepo.GetPaymentById(ctx, charge.PaymentSessionID)
	if err != nil {
		s.logger.Warn(ctx, "Failed to get payment for inquiry",
			logger.String("chargeID", request.ChargeID),
			logger.String("paymentSessionID", charge.PaymentSessionID),
			logger.Error(err),
		)
	}

	if payment != nil {
		inquiryResult, err := s.performPaymentInquiry(ctx, payment, unifiedPaymentModel.PerformInquiryRequest{
			LedgerID: request.ChargeID,
		})
		if err != nil {
			s.logger.Warn(ctx, "Payment inquiry failed, returning stored status",
				logger.String("paymentUUID", payment.UUID),
				logger.Error(err),
			)
		}

		if inquiryResult != nil && inquiryResult.UpdatedStatus {
			if err := s.processNotificationFromInquiry(ctx, payment, inquiryResult, charge.ID); err == nil {
				charge, _ = s.paymentRepo.GetChargeByID(ctx, request.ChargeID)
			}
		}

		if inquiryResult != nil && inquiryResult.LastInquiryAt != nil {
			charge.LastInquiryAt = inquiryResult.LastInquiryAt
		}
	}

	metadata := payment.ToUnifiedPaymentMetadata()
	if metadata != nil {
		charge.VirtualTerminal = metadata.VirtualTerminal
	}

	statusHistoryList, err := s.statusHistoriesRepo.GetByReference(ctx, constant.TypePayment, charge.PaymentSessionID)
	if err != nil {
		s.logger.Warn(ctx, "failed to get status history", logger.String("paymentId", charge.PaymentSessionID), logger.Error(err))
	}

	if len(statusHistoryList) == 0 {
		return charge, nil
	}

	charge.StatusHistory = make([]unifiedPaymentModel.ChargeStatusHistoryResponse, 0, len(statusHistoryList))
	seenLabel := make(map[string]bool)
	for _, statusHistory := range statusHistoryList {
		label := s.getPaymentStatusHistoryLabelsAndDescriptions(statusHistory.Status, statusInfoLabel)
		description := s.getPaymentStatusHistoryLabelsAndDescriptions(statusHistory.Status, statusInfoDescription)
		recommendation := s.getPaymentStatusHistoryLabelsAndDescriptions(statusHistory.Status, statusInfoRecommendation)
		if _, ok := seenLabel[label]; ok {
			continue
		}
		charge.StatusHistory = append(charge.StatusHistory, unifiedPaymentModel.ChargeStatusHistoryResponse{
			Status:         statusHistory.Status,
			Label:          label,
			Description:    description,
			Recommendation: recommendation,
			Timestamp:      &statusHistory.CreatedAt,
		})
		seenLabel[label] = true
	}

	return charge, nil
}

func (s *UnifiedPaymentService) ExportCharge(ctx context.Context, request *unifiedPaymentModel.FilterChargeRequest) (*commonModel.ExportResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/wallet/transaction/ExportMerchantTransactionHistoryList")
	defer segment.End()

	loc := util.GetTimeLocationFromContext(ctx)

	var (
		hashFilter = request.HashFilter(loc.String())
		result     = &commonModel.ExportResponse{}
		cacheKey   = fmt.Sprintf(constant.RedisKeyDownloadChargeHistoryFmt, hashFilter)
	)

	err := s.cache.Get(ctx, cacheKey).Scan(result)
	if err == nil {
		return result, nil
	}

	if !errors.Is(err, redisExt.ErrNil) {
		s.logger.Error(ctx, "Failed while retrieving download cache", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	charges, err := s.paymentRepo.GetCharges(ctx, request)
	if err != nil {
		s.logger.Error(ctx, "Failed while get merchant transaction history list", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrDatabase, err)
	}

	buf, err := s.ExportToExcelChargeHistories(ctx, request, charges)
	if err != nil {
		s.logger.Error(ctx, "Failed while export merchant transaction history list", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, err)
	}

	objectName := "downloads/unified-payment/charge-histories/" + hashFilter + ".xlsx"
	downloadFilename := fmt.Sprintf(`attachment; filename=merchant_charge_histories_%v.xlsx`, time.Now().Unix())

	if _, err := s.storage.UploadFile(ctx, objectName, buf, true, gcs.WriteContentDisposition(downloadFilename)); err != nil {
		s.logger.Error(ctx, "Failed while upload file to gcs", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, err)
	}
	result.ExpiresAt = time.Now().UTC().Add(chargeExportCacheExpiration)
	result.DownloadURL, err = s.storage.CreateSignedURL(ctx, objectName, chargeExportCacheExpiration)
	if err != nil {
		s.logger.Error(ctx, "Failed while create signed URL", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, err)
	}

	if err := s.cache.Set(ctx, cacheKey, result, chargeExportCacheExpiration).Err(); err != nil {
		s.logger.Error(ctx, "Failed while set signature URL", logger.Error(err))
		return nil, pkgErr.New(response.HttpErrInternal, err)
	}
	return result, nil
}
