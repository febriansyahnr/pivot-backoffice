package cardFundedPayoutService

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const (
	RedisTemplateCardFundedPayoutReceiptKey = "backend-portal:card-funded-payouts:%s:receipt"

	CardFundedPayoutReceiptFilenameTemplate = "Card_Funded_Payout_Receipt_%s"
	CardFundedPayoutReceiptBucketDir        = "card-funded-payouts/receipt"
)

var cardFundedPayoutReceiptExpireDuration = time.Hour

func (s *service) GetReceipt(
	ctx context.Context,
	request *cardFundedPayoutModel.GetReceiptRequest,
) (*cardFundedPayoutModel.GetReceiptResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/GetReceipt")
	defer span.End()

	// Use distributed lock to prevent concurrent generation for the same payout
	lockKey := fmt.Sprintf(RedisTemplateCardFundedPayoutReceiptKey, request.PayoutID) + ":lock"
	if can, err := s.cacheClient.SetNX(ctx, lockKey, true, 10*time.Second).Result(); err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if !can {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("the same request is in progress"))
	}
	defer func() {
		_ = s.cacheClient.Del(ctx, lockKey)
	}()

	// Get from cache, return if exist
	tokenKey := fmt.Sprintf(RedisTemplateCardFundedPayoutReceiptKey, request.PayoutID)
	if receiptURL, err := s.cacheClient.Get(ctx, tokenKey).Result(); err == nil {
		return &cardFundedPayoutModel.GetReceiptResponse{
			ReceiptURL: receiptURL,
		}, nil
	}

	// Get payout detail
	payoutDetail, err := s.disbursementRepo.GetCardFundedPayoutDetail(ctx, &cardFundedPayoutModel.GetPayoutDetailRequest{
		PayoutID:   request.PayoutID,
		MerchantID: request.MerchantID,
	})
	if err != nil {
		s.logger.Error(ctx, "error getting payout detail for receipt", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}
	if payoutDetail == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrDisbursementNotFound)
	}

	// Validate transaction status must be SUCCESS
	if payoutDetail.TransactionStatus != constant.StatusSuccess {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrDisbursementNotSuccessYet)
	}

	amount, _ := strconv.ParseFloat(payoutDetail.Amount, 64)
	fee, _ := strconv.ParseFloat(payoutDetail.Fee, 64)
	totalAmount, _ := strconv.ParseFloat(payoutDetail.TotalAmount, 64)

	// Build receipt data
	receiptData := &cardFundedPayoutModel.ReceiptData{
		CreatedAt:       util.ConvertToJakarta(payoutDetail.CreatedAt).Format(util.ReceiptFormatLayout),
		ReferenceID:     payoutDetail.ReferenceID,
		PayoutID:        payoutDetail.UUID,
		Amount:          util.FormatRupiahWithoutDecimal(amount),
		Fee:             util.FormatRupiahWithoutDecimal(fee),
		TotalAmount:     util.FormatRupiahWithoutDecimal(totalAmount),
		VendorName:      payoutDetail.VendorName,
		Remarks:         payoutDetail.Remarks,
		BankName:        payoutDetail.BankName,
		AccountNo:       payoutDetail.AccountNumber,
		AccountName:     payoutDetail.AccountName,
		ImageHeader:     s.config.MerchantPortalConfig.LogoURL,
		ImageBackground: s.config.MerchantPortalConfig.PaymentReceiptBackgroundURL,
	}

	// Generate and upload receipt PDF to GCS
	gcsPath, err := s.generateAndUploadReceipt(ctx, receiptData)
	if err != nil {
		return nil, err
	}

	// Store gcsPath to redis
	_ = s.cacheClient.Set(ctx, tokenKey, gcsPath, cardFundedPayoutReceiptExpireDuration)

	return &cardFundedPayoutModel.GetReceiptResponse{
		ReceiptURL: gcsPath,
	}, nil
}

func (s *service) generateAndUploadReceipt(
	ctx context.Context,
	receiptData *cardFundedPayoutModel.ReceiptData,
) (string, error) {
	// Get template path
	templatePath := "templates/cardFundedPayout/receipt.html"

	// Define PDF file name
	filename := fmt.Sprintf(CardFundedPayoutReceiptFilenameTemplate, receiptData.PayoutID) + ".pdf"
	gcsObjectName := CardFundedPayoutReceiptBucketDir + "/" + filename

	// Set GCS Writer
	gcsWriter, err := s.gcs.SetBucketWriter(ctx, gcsObjectName)
	if err != nil {
		s.logger.Error(ctx, "error init bucket writer", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateReceipt)
	}

	// Generate & Write PDF into GCS
	r := pdf.NewRequestPdf(wkhtmltopdf.PageSizeA4, wkhtmltopdf.OrientationPortrait, pdf.WithOutput(gcsWriter))
	if err := r.GeneratePDF(ctx, templatePath, receiptData); err != nil {
		s.logger.Error(ctx, "error generate pdf", logger.Error(err))

		// close gcs writer if error
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
	signedURL, err := s.gcs.CreateSignedURL(ctx, gcsObjectName, cardFundedPayoutReceiptExpireDuration)
	if err != nil {
		s.logger.Error(ctx, "error generating signed URL", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateReceipt)
	}

	return signedURL, nil
}
