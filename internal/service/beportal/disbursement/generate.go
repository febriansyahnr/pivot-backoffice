package disbursementService

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/tealeg/xlsx"
)

func (s *DisbursementService) GenerateExcelAndUpdateInvalidBulkDisbursement(ctx context.Context, bulkID string, rows []*disbursementModel.BulkPreviewResponse) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GenerateExcelAndGetInvalidPath")
	defer segment.End()

	// generate excel and get the path
	gcsResponse, err := s.generateExcel(ctx, rows, constant.DefaultFilenameInvalidBulkDisbursement)
	if err != nil {
		return "", err
	}

	// Update bulk disbursement
	if err = s.disbursementRepo.UpdateBulkDisbursementFailedFileByID(ctx, bulkID, gcsResponse.PublicUrl); err != nil {
		return "", err
	}

	return gcsResponse.SignedUrl, nil
}

func (s *DisbursementService) GenerateExcelAndUpdateRejectedBulkDisbursement(ctx context.Context, bulkID string, rows []*disbursementModel.BulkPreviewResponse) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/GenerateExcelAndUpdateRejectedBulkDisbursement")
	defer segment.End()

	// generate excel and get the path
	gcsResponse, err := s.generateExcel(ctx, rows, constant.DefaultFilenameRejectedBulkDisbursement)
	if err != nil {
		return "", err
	}

	// Update bulk disbursement (rejected file)
	if err = s.disbursementRepo.UpdateBulkDisbursementRejectedFileByID(ctx, bulkID, gcsResponse.PublicUrl); err != nil {
		return "", err
	}

	return gcsResponse.SignedUrl, nil
}

var bufPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func (s *DisbursementService) generateExcel(ctx context.Context, rows []*disbursementModel.BulkPreviewResponse, filenameTemplate string) (*gcs.Response, error) {

	file := xlsx.NewFile()

	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	sheet, err := file.AddSheet("Template")
	if err != nil {
		s.logger.Error(ctx, "failed to add sheet", logger.Error(err))
		return nil, err
	}

	// Add Style
	style := xlsx.NewStyle()
	style.Font.Bold = true
	style.Fill = *xlsx.NewFill("solid", "E4E4E4", "E4E4E4") // Background color
	style.ApplyFill = true

	// Add headers
	headers := []string{"Reference ID", "Amount", "Channel Code", "Account Number", "Account Name", "Remarks", "Result", "Error"}
	row := sheet.AddRow()
	for _, header := range headers {
		cell := row.AddCell()
		cell.Value = header
		cell.SetStyle(style)
	}

	for _, dt := range rows {
		row := sheet.AddRow()

		row.AddCell().Value = dt.ReferenceID
		row.AddCell().Value = dt.Amount
		row.AddCell().Value = dt.ChannelCode
		row.AddCell().Value = dt.BeneficiaryAccountNo
		row.AddCell().Value = dt.BeneficiaryAccountName
		row.AddCell().Value = dt.Remark
		row.AddCell().Value = dt.Result
		row.AddCell().Value = dt.Error
	}

	maxWidths := make([]float64, len(headers))
	for i, header := range headers {
		maxWidths[i] = float64(len(header)) * 1.5
	}

	for _, row := range sheet.Rows {
		for colIndex, cell := range row.Cells {
			textLength := float64(len(cell.Value)) * 1.25
			if maxWidths[colIndex] < textLength {
				maxWidths[colIndex] = textLength
			}
		}
	}

	padding := 2.0
	for colIndex, width := range maxWidths {
		sheet.SetColWidth(colIndex, colIndex, width*0.75+padding)
	}

	if err = file.Write(buf); err != nil {
		s.logger.Error(ctx, "failed to write buffer file", logger.Error(err))
		return nil, err
	}

	objectName := filepath.Join(
		"disbursements/bulk-transactions", constant.ExportBulkDisbursementFailedDir, fmt.Sprintf(
			filenameTemplate, util.GetCurrentTimeWithMillisFormatted(),
		),
	) + constant.DefaultExtXlsx

	gcsResponse, err := s.gcs.UploadFile(ctx, objectName, buf, true)
	if err != nil {
		s.logger.Error(ctx, "failed when upload file", logger.Error(err))
		return nil, err
	}

	signedURL, _ := s.gcs.CreateSignedURL(ctx, objectName, 15*time.Minute)
	return &gcs.Response{
		SignedUrl: signedURL,
		PublicUrl: gcsResponse.PublicURL,
	}, nil
}
