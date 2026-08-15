package disbursementService

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/paper-indonesia/pivot-backoffice/pkg/pdf"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

const (
	RedisTemplateDisbursementReceiptKey = "backend-portal:disbursements:%s:receipt"

	PDFFilenameTemplate          = "Disbursement_Receipt_%s"
	DisbursementReceiptBucketDir = "disbursements/receipt"
)

var (
	receiptExpireDuration = time.Hour
)

func (s *DisbursementService) GetReceiptByID(ctx context.Context, request *disbursementModel.GetDisbursementReceiptRequest) (*disbursementModel.GetDisbursementReceiptResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GetReceiptByID")
	defer segment.End()

	lockKeyValue := request.DisbursementID
	if request.DisbursementID == "" && request.ReferenceID != "" {
		lockKeyValue = request.ReferenceID
	}
	lockKey := fmt.Sprintf(RedisTemplateDisbursementReceiptKey, lockKeyValue) + ":lock"
	if can, err := s.redisExt.SetNX(ctx, lockKey, true, 10*time.Second).Result(); err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	} else if !can {
		return nil, pkgErrors.New(response.HttpErrUnprocessableContent, errors.New("the same request is in progress"))
	}
	defer func() {
		_ = s.redisExt.Del(ctx, lockKey)
	}()

	// Get from cache, return if exist
	tokenKey := fmt.Sprintf(RedisTemplateDisbursementReceiptKey, lockKeyValue)
	if receiptURL, err := s.redisExt.Get(ctx, tokenKey).Result(); err == nil {
		return &disbursementModel.GetDisbursementReceiptResponse{
			ReceiptURL: receiptURL,
		}, nil
	}

	var (
		disbursement *disbursementModel.DisbursementWithTransaction
		err          error
	)
	if request.DisbursementID != "" {
		disbursement, err = s.disbursementRepo.FindByID(ctx, request.DisbursementID)
	} else if request.ReferenceID != "" {
		// Background: to fulfill merchant request via ops to get receipt
		disbursement, err = s.disbursementRepo.FindByMerchantAndReference(ctx, request.MerchantID, request.ReferenceID)
	}
	if err != nil {
		return nil, pkgErrors.New(response.HttpErrDatabase, err)
	}
	if disbursement == nil {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrDisbursementNotFound)
	}
	if !(disbursement.TransactionStatus != nil && *disbursement.TransactionStatus == constant.StatusSuccess) {
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrDisbursementNotSuccessYet)
	}

	// Get merchant from disbursement merchantID
	merchantData, err := s.merchantSvc.FindMerchantByID(ctx, disbursement.MerchantID)
	if err != nil {
		return nil, err
	}

	// check disbursement with merchant
	if request.MerchantID != "" && disbursement.MerchantID != request.MerchantID {
		s.logger.Error(ctx, "Merchant not match", logger.String("disbursementMerchantID", disbursement.MerchantID), logger.String("merchantID", request.MerchantID))
		return nil, pkgErrors.New(response.HttpErrRequest, constant.ErrDisbursementNotFound)
	}

	// Build receipt data
	var (
		remark              = ""
		bankReferenceNo     = ""
		beneficiaryBankName = ""
	)
	if disbursement.Remark != nil {
		remark = *disbursement.Remark
	}
	if disbursement.BankReferenceNo != nil {
		bankReferenceNo = *disbursement.BankReferenceNo
	}
	if disbursement.BeneficiaryBankName != nil {
		beneficiaryBankName = *disbursement.BeneficiaryBankName
	}

	amount, _ := disbursement.Amount.Float64()

	receiptData := &disbursementModel.DisbursementReceiptData{
		CompletedAt:            util.ConvertToJakarta(disbursement.UpdatedAt).Format(util.ReceiptFormatLayout),
		ReferenceID:            disbursement.ReferenceID,
		Remark:                 remark,
		MerchantName:           merchantData.Name,
		SenderName:             s.config.MerchantPortalConfig.ReceiptSenderName,
		Amount:                 util.FormatRupiahWithoutDecimal(amount),
		Status:                 *disbursement.TransactionStatus,
		DisbursementID:         disbursement.UUID,
		BankReferenceNo:        bankReferenceNo,
		BeneficiaryBankName:    beneficiaryBankName,
		BeneficiaryAccountNo:   disbursement.BeneficiaryAccountNo,
		BeneficiaryAccountName: disbursement.BeneficiaryAccountName,

		ImageHeader: s.config.MerchantPortalConfig.LogoURL,
	}

	// Generate and upload receipt PDF
	gcsPath, err := s.generateAndUploadReceipt(ctx, receiptData)
	if err != nil {
		return nil, err
	}

	// Store gcsPath to redis
	_ = s.redisExt.Set(ctx, tokenKey, gcsPath, receiptExpireDuration)

	return &disbursementModel.GetDisbursementReceiptResponse{
		ReceiptURL: gcsPath,
	}, nil
}

func (s *DisbursementService) generateAndUploadReceipt(ctx context.Context, receiptData *disbursementModel.DisbursementReceiptData) (string, error) {
	// Get template path
	templatePath := "templates/disbursement/receipt.html"

	// Define PDF file name
	filename := fmt.Sprintf(PDFFilenameTemplate, receiptData.DisbursementID) + ".pdf"
	gcsObjectName := DisbursementReceiptBucketDir + "/" + filename

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
	signedURL, err := s.gcs.CreateSignedURL(ctx, gcsObjectName, receiptExpireDuration)
	if err != nil {
		s.logger.Error(ctx, "error generating signed URL", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateReceipt)
	}

	return signedURL, nil
}
