package paymentService

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

const (
	RedisTemplatePaymentReceiptKey = "backend-portal:payments:%s:%s:receipt"
	PaymentReceiptFilenameTemplate = "Payment_Receipt_%s"
)

var (
	paymentReceiptExpireDuration = time.Minute * 15
)

func (s *PaymentService) GetReceiptByID(ctx context.Context, request *paymentModel.GetPaymentReceiptRequest) (*paymentModel.GetPaymentReceiptResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/GetReceiptByID")
	defer segment.End()

	// Determine key value for caching and singleflight
	keyValue := request.PaymentID
	if request.PaymentID == "" && request.ReferenceID != "" {
		keyValue = request.ReferenceID
	}

	// Use singleflight to deduplicate concurrent requests
	// Multiple requests for the same key will wait and share the result
	result, err, _ := s.receiptSf.Do(keyValue, func() (interface{}, error) {
		return s.generateReceipt(ctx, request, keyValue)
	})
	if err != nil {
		return nil, err
	}

	return result.(*paymentModel.GetPaymentReceiptResponse), nil
}

func (s *PaymentService) generateReceipt(ctx context.Context, request *paymentModel.GetPaymentReceiptRequest, keyValue string) (*paymentModel.GetPaymentReceiptResponse, error) {
	// Check cache for existing receipt
	tokenKey := fmt.Sprintf(RedisTemplatePaymentReceiptKey, request.MerchantID, keyValue)
	if cachedReceipt, err := s.redis.Get(ctx, tokenKey).Result(); err == nil && cachedReceipt != "" {
		return &paymentModel.GetPaymentReceiptResponse{
			Filename: fmt.Sprintf(PaymentReceiptFilenameTemplate, keyValue) + ".pdf",
			PDF:      []byte(cachedReceipt),
		}, nil
	}

	// Get payment data with merchant info in single query (reduces DB round-trips)
	receiptDTO, err := s.paymentRepo.GetPaymentReceiptData(ctx, request.PaymentID, request.ReferenceID, request.MerchantID)
	if err != nil {
		s.logger.Error(ctx, "error getting payment receipt data", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}
	if receiptDTO == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentNotFound)
	}

	// Validate payment status must be PAID
	if receiptDTO.Status != constant.UnifiedPaymentSessionStatusPaid {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentNotSuccessYet)
	}

	// Validate merchant
	if request.MerchantID != "" && receiptDTO.MerchantID != request.MerchantID {
		s.logger.Error(ctx, "merchant not match", logger.String("paymentMerchantID", receiptDTO.MerchantID), logger.String("merchantID", request.MerchantID))
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentNotFound)
	}

	// Build receipt data
	amount, _ := receiptDTO.TotalAmount.Float64()
	paymentMethod := receiptDTO.PaymentMethod.String

	receiptData := &paymentModel.PaymentReceiptData{
		TransactionDate: util.ConvertToJakarta(receiptDTO.CreatedAt).Format(util.ReceiptFormatLayout),
		ReferenceID:     util.ValueOfPtr(receiptDTO.ReferenceID),
		MerchantName:    receiptDTO.MerchantName.String,
		Amount:          util.FormatRupiahWithoutDecimal(amount),
		PaymentID:       receiptDTO.UUID,
		PaymentMethod:   paymentMethod,
		ImageHeader:     s.config.MerchantPortalConfig.LogoURL,
		ImageBackground: s.config.MerchantPortalConfig.PaymentReceiptBackgroundURL,
	}

	// Generate PDF
	pdfBytes, filename, err := s.generateReceiptPDF(ctx, receiptData, receiptDTO.UUID)
	if err != nil {
		s.logger.Error(ctx, "error generating receipt PDF", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateReceipt)
	}

	// Cache PDF bytes (base64 encoded for redis storage)
	if err := s.redis.Set(ctx, tokenKey, string(pdfBytes), paymentReceiptExpireDuration).Err(); err != nil {
		// Don't fail the request if caching fails, just log the error
		s.logger.Error(ctx, "error caching receipt PDF", logger.Error(err))
	}

	return &paymentModel.GetPaymentReceiptResponse{
		Filename: filename,
		PDF:      pdfBytes,
	}, nil
}

func (s *PaymentService) generateReceiptPDF(ctx context.Context, receiptData *paymentModel.PaymentReceiptData, paymentID string) ([]byte, string, error) {
	// Get template path
	templatePath := "templates/payment/receipt.html"

	// Define PDF file name
	filename := fmt.Sprintf(PaymentReceiptFilenameTemplate, paymentID) + ".pdf"

	// Generate PDF to buffer
	var buf bytes.Buffer
	r := pdf.NewRequestPdf(wkhtmltopdf.PageSizeA4, wkhtmltopdf.OrientationPortrait, pdf.WithOutput(&buf))
	if err := r.GeneratePDF(ctx, templatePath, receiptData); err != nil {
		s.logger.Error(ctx, "error generate pdf", logger.Error(err))
		return nil, "", err
	}

	return buf.Bytes(), filename, nil
}
