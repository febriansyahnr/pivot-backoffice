package refundService

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const (
	RedisTemplateRefundReceiptKey = "backend-portal:refunds:%s:receipt"

	RefundReceiptFilenameTemplate = "Refund_Receipt_%s"
	RefundReceiptBucketDir        = "refunds/receipt"
)

var refundReceiptExpireDuration = time.Hour

// GetReceipt generates and returns a signed URL for a refund receipt PDF.
// Only generates receipt for refunds with SUCCESS status.
func (s *RefundService) GetReceipt(
	ctx context.Context,
	request *refundModel.GetRefundReceiptRequest,
) (*refundModel.GetRefundReceiptResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/refund/GetReceipt")
	defer span.End()

	// Use distributed lock to prevent concurrent generation for the same refund
	lockKey := fmt.Sprintf(RedisTemplateRefundReceiptKey, request.RefundID) + ":lock"
	if can, err := s.redis.SetNX(ctx, lockKey, true, 10*time.Second).Result(); err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if !can {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("the same request is in progress"))
	}
	defer func() {
		_ = s.redis.Del(ctx, lockKey)
	}()

	// Get from cache, return if exist
	tokenKey := fmt.Sprintf(RedisTemplateRefundReceiptKey, request.RefundID)
	if receiptURL, err := s.redis.Get(ctx, tokenKey).Result(); err == nil {
		return &refundModel.GetRefundReceiptResponse{
			ReceiptURL: receiptURL,
		}, nil
	}

	// Get refund detail
	refundDetail, err := s.refundRepo.GetRefundByID(ctx, request.RefundID, request.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "error getting refund detail for receipt", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}
	if refundDetail == nil {
		return nil, pkgErrors.New(response.HttpErrNotFound, constant.ErrRefundNotFound)
	}

	// Validate transaction status must be SUCCESS
	if refundDetail.Status != constant.RefundStatusSuccess {
		return nil, pkgErrors.New(response.HttpErrRequest, errors.New("receipt is only available for successful refunds"))
	}

	// Build receipt data
	amount, _ := strconv.ParseFloat(refundDetail.Amount.Value, 64)
	receiptData := &refundModel.RefundReceiptData{
		Amount:             util.FormatRupiahWithoutDecimal(amount),
		CompletedAt:        util.ConvertToJakarta(refundDetail.UpdatedAt).Format(util.ReceiptFormatLayout),
		RefundReferenceID:  refundDetail.ClientReferenceID,
		PaymentReferenceID: refundDetail.PaymentSessionID,
		RefundReason:       refundDetail.Reason,
		RefundDestination:  buildRefundDestinationDisplay(refundDetail),
		RRN:                extractRRN(refundDetail),
		ImageHeader:        s.config.MerchantPortalConfig.LogoURL,
		ImageBackground:    s.config.MerchantPortalConfig.PaymentReceiptBackgroundURL,
	}

	// Generate and upload receipt PDF to GCS
	gcsPath, err := s.generateAndUploadRefundReceipt(ctx, request.RefundID, receiptData)
	if err != nil {
		return nil, err
	}

	// Store gcsPath to redis
	_ = s.redis.Set(ctx, tokenKey, gcsPath, refundReceiptExpireDuration)

	return &refundModel.GetRefundReceiptResponse{
		ReceiptURL: gcsPath,
	}, nil
}

func (s *RefundService) generateAndUploadRefundReceipt(
	ctx context.Context,
	refundID string,
	receiptData *refundModel.RefundReceiptData,
) (string, error) {
	templatePath := "templates/refund/receipt.html"

	filename := fmt.Sprintf(RefundReceiptFilenameTemplate, refundID) + ".pdf"
	gcsObjectName := RefundReceiptBucketDir + "/" + filename

	// Set GCS Writer
	gcsWriter, err := s.gcs.SetBucketWriter(ctx, gcsObjectName)
	if err != nil {
		s.logger.Error(ctx, "error init bucket writer for refund receipt", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateReceipt)
	}

	// Generate & Write PDF into GCS
	r := pdf.NewRequestPdf(wkhtmltopdf.PageSizeA4, wkhtmltopdf.OrientationPortrait, pdf.WithOutput(gcsWriter))
	if err := r.GeneratePDF(ctx, templatePath, receiptData); err != nil {
		s.logger.Error(ctx, "error generate refund receipt pdf", logger.Error(err))

		if err := gcsWriter.Close(); err != nil {
			s.logger.Error(ctx, "error closing GCS writer", logger.Error(err))
		}

		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateReceipt)
	}

	// Close the GCS writer to flush data to GCS
	if err := gcsWriter.Close(); err != nil {
		s.logger.Error(ctx, "error closing GCS writer", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateReceipt)
	}

	// Generate Signed URL
	signedURL, err := s.gcs.CreateSignedURL(ctx, gcsObjectName, refundReceiptExpireDuration)
	if err != nil {
		s.logger.Error(ctx, "error generating signed URL for refund receipt", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateReceipt)
	}

	return signedURL, nil
}

// buildRefundDestinationDisplay builds a human-readable destination string for the receipt.
func buildRefundDestinationDisplay(refund *refundModel.RefundResponse) string {
	if refund.DestinationType == "CHANNEL" && refund.ChannelDestination != nil {
		cd := refund.ChannelDestination
		switch strings.ToUpper(cd.PaymentMethod) {
		case "CREDIT_CARD":
			last4 := ""
			brand := ""
			issuing := ""
			if v, ok := cd.PaymentDetail["last4Digit"].(string); ok {
				last4 = v
			}
			if v, ok := cd.PaymentDetail["cardBrand"].(string); ok {
				brand = v
			}
			if v, ok := cd.PaymentDetail["cardIssuing"].(string); ok {
				issuing = v
			}
			if issuing != "" && brand != "" && last4 != "" {
				return fmt.Sprintf("%s (%s) - *%s", issuing, brand, last4)
			}
		case "EWALLET":
			channel := cd.PaymentChannel
			detail := ""
			if v, ok := cd.PaymentDetail["channel"].(string); ok {
				detail = v
			}
			if channel != "" && detail != "" {
				return fmt.Sprintf("%s - %s", channel, detail)
			}
			if channel != "" {
				return channel
			}
		case "QRIS":
			acquirer := ""
			merchantName := ""
			if v, ok := cd.PaymentDetail["acquirer"].(string); ok {
				acquirer = v
			}
			if v, ok := cd.PaymentDetail["merchantName"].(string); ok {
				merchantName = v
			}
			if acquirer != "" && merchantName != "" {
				return fmt.Sprintf("%s (%s)", acquirer, merchantName)
			}
		}
	}

	if refund.DestinationType == "ACCOUNT" && refund.TransferDestination != nil {
		td := refund.TransferDestination
		maskedAccount := ""
		if len(td.ChannelInformation.AccountNumber) >= 4 {
			maskedAccount = "*" + td.ChannelInformation.AccountNumber[len(td.ChannelInformation.AccountNumber)-4:]
		}
		if td.ChannelCode != "" && maskedAccount != "" {
			return fmt.Sprintf("%s - %s", td.ChannelCode, maskedAccount)
		}
	}

	return "-"
}

// extractRRN extracts the RRN from channel destination payment detail if available.
func extractRRN(refund *refundModel.RefundResponse) string {
	if refund.ChannelDestination != nil && refund.ChannelDestination.PaymentDetail != nil {
		if rrn, ok := refund.ChannelDestination.PaymentDetail["rrn"].(string); ok && rrn != "" {
			return rrn
		}
	}
	return ""
}
