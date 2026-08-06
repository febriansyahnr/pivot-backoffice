package cardFundedPayoutService

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/tealeg/xlsx"
)

func (s *service) ExportPayoutList(ctx context.Context, filter *cardFundedPayoutModel.FilterGetPayoutList) (*cardFundedPayoutModel.ExportPayoutListResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/ExportPayoutList")
	defer segment.End()

	var (
		payouts         []*cardFundedPayoutModel.GetPayoutListResponse
		page            int64 = 1
		perPage         int64 = 1000
		isNextPageExist       = true
	)

	// Loop through all pages to get all data
	for {
		filter.Page = page
		filter.PerPage = perPage

		list, err := s.disbursementRepo.GetCardFundedPayoutList(ctx, filter)
		if err != nil {
			return nil, pkgErrors.New(response.HttpErrDatabase, err)
		}

		data, ok := list.Data.([]*cardFundedPayoutModel.GetPayoutListResponse)
		if !ok || len(data) == 0 {
			isNextPageExist = false
		} else {
			payouts = append(payouts, data...)

			if page >= list.Meta.TotalPages {
				isNextPageExist = false
			} else {
				page += 1
			}
		}

		if !isNextPageExist {
			break
		}
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf(constant.DefaultFilenameCardFundedPayoutHistory, util.GetCurrentTimeWithMillisFormatted())

	// Generate Excel file
	excelFile, err := s.generateExcelFile(ctx, filename, payouts)
	if err != nil {
		return nil, err
	}

	// Upload to GCS
	signedUrl, err := s.uploadExcelFileToGCS(ctx, filename, excelFile)
	if err != nil {
		return nil, err
	}

	// Clean up temporary file
	defer func() {
		if err := os.Remove(excelFile); err != nil {
			s.logger.Error(ctx, "failed to remove temporary Excel file", logger.Error(err))
		}
	}()

	return &cardFundedPayoutModel.ExportPayoutListResponse{
		Url: signedUrl,
	}, nil
}

func (s *service) uploadExcelFileToGCS(ctx context.Context, filename, srcFile string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/uploadExcelFileToGCS")
	defer segment.End()

	gcsObjectName := constant.CardFundedPayoutHistoryBucketDir + "/" + filename + constant.DefaultExtXlsx

	resp, err := s.gcs.UploadFileToGCS(ctx, gcsObjectName, srcFile, true, nil)
	if err != nil {
		s.logger.Error(ctx, "failed to upload to GCS", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	return resp.SignedUrl, nil
}

func (s *service) generateExcelFile(ctx context.Context, filename string, payouts []*cardFundedPayoutModel.GetPayoutListResponse) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/cardFundedPayout/generateExcelFile")
	defer segment.End()

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Card Funded Payout History")
	if err != nil {
		s.logger.Error(ctx, "failed to add sheet", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	tz, err := util.GetTimeZoneFromContext(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to get timezone from context", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrInvalidTimeZone)
	}

	// Add style for headers
	style := xlsx.NewStyle()
	style.Font.Bold = true
	style.Fill = *xlsx.NewFill("solid", "E4E4E4", "E4E4E4")
	style.ApplyFill = true

	// Add headers
	headers := []string{
		"Payout ID",
		"Created Date",
		"Reference ID",
		"Amount",
		"Payout Status",
		"Approval",
		"Vendor Name",
		"Card",
	}
	row := sheet.AddRow()
	for _, header := range headers {
		cell := row.AddCell()
		cell.Value = header
		cell.SetStyle(style)
	}

	// Add data rows
	for _, payout := range payouts {
		row := sheet.AddRow()

		row.AddCell().Value = payout.UUID
		row.AddCell().Value = payout.CreatedAt.In(tz).Format(util.ReceiptFormatLayout)
		row.AddCell().Value = payout.ReferenceID
		amount, _ := strconv.ParseFloat(payout.Amount, 64)
		row.AddCell().Value = util.ConvertFloatToCurrency(amount)
		row.AddCell().Value = payout.TransactionStatus
		row.AddCell().Value = payout.ApprovalStatus
		row.AddCell().Value = payout.VendorName
		row.AddCell().Value = fmt.Sprintf("*%s %s (%s)", payout.Card.LastFour, payout.Card.Brand, payout.Card.Channel)
	}

	// Auto-adjust column widths
	maxWidths := make([]float64, len(headers))
	for i, header := range headers {
		maxWidths[i] = float64(len(header)) * 1.5
	}

	for _, row := range sheet.Rows {
		for colIndex, cell := range row.Cells {
			textLength := float64(len(cell.Value)) * 1.25
			if colIndex < len(maxWidths) && maxWidths[colIndex] < textLength {
				maxWidths[colIndex] = textLength
			}
		}
	}

	padding := 2.0
	for colIndex, width := range maxWidths {
		sheet.SetColWidth(colIndex, colIndex, width*0.75+padding)
	}

	// Ensure temp directory exists
	if err := util.EnsureTmpDir(constant.ExportTempDir); err != nil {
		s.logger.Error(ctx, "failed to ensure /tmp directory", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	// Save file
	tempFilePath := filepath.Join(constant.ExportTempDir, filename+constant.DefaultExtXlsx)
	if err = file.Save(tempFilePath); err != nil {
		s.logger.Error(ctx, "failed to save file", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	return tempFilePath, nil
}
