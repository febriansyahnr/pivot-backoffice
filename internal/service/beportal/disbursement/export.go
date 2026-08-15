package disbursementService

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/tealeg/xlsx"
)

func (s *DisbursementService) ExportToExcel(ctx context.Context, filter *disbursementModel.GetDisbursementFilterRequest) (*disbursementModel.ExportDisbursementListResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/ExportToExcel")
	defer segment.End()

	var (
		disbursements   []*disbursementModel.DisbursementWithTransactionResponse
		page            int64 = 1
		perPage         int64 = 1000
		isNextPageExist       = true
		errGenerate     error
	)

	// Looping if the data in pagination still exist, export data. (Avoid to query without limitation)
	for {
		if list, err := s.disbursementRepo.GetList(ctx, filter, page, perPage); err != nil {
			return nil, pkgErrors.New(response.HttpErrDatabase, err)
		} else if len(list.Data.([]*disbursementModel.DisbursementWithTransactionResponse)) == 0 {
			isNextPageExist = false
		} else {
			disbursements = append(disbursements, list.Data.([]*disbursementModel.DisbursementWithTransactionResponse)...)

			if page < list.Meta.TotalPages {
				page += 1
			} else {
				isNextPageExist = false
			}
		}

		// Check the condition to exit the loop
		if !isNextPageExist {
			break
		}
	}

	// define filename
	filename := fmt.Sprintf(constant.DefaultFilenameDisbursementHistory, util.GetCurrentTimeWithMillisFormatted())

	// Generate excel file
	excelFile := ""
	if filter.BulkID == "" {
		if excelFile, errGenerate = s.generateDefaultExcelFile(ctx, filename, disbursements); errGenerate != nil {
			return nil, errGenerate
		}
	} else {
		// Generate excel with upload format
		if excelFile, errGenerate = s.generateBulkDisbursementExcelFile(ctx, filename, disbursements); errGenerate != nil {
			return nil, errGenerate
		}
	}

	signedUrl, err := s.uploadExcelFileToGCS(ctx, filename, excelFile)
	if err != nil {
		return nil, err
	}

	// Remove the Excel file after successful upload
	defer func() {
		if err = os.Remove(excelFile); err != nil {
			s.logger.Error(ctx, "failed to remove temporary Excel file", logger.Error(err))
		}
	}()

	return &disbursementModel.ExportDisbursementListResponse{
		Url: signedUrl,
	}, nil
}

func (s *DisbursementService) uploadExcelFileToGCS(ctx context.Context, filename, srcFile string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/uploadExcelFileToGCS")
	defer segment.End()

	// Upload to GCS
	gcsObjectName := constant.DisbursementHistoryBucketDir + "/" + filename + constant.DefaultExtXlsx

	// upload file to GCS
	resp, err := s.gcs.UploadFileToGCS(ctx, gcsObjectName, srcFile, true, nil)
	if err != nil {
		s.logger.Error(ctx, "failed to upload to GCS", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	return resp.SignedUrl, nil
}

func (s *DisbursementService) generateDefaultExcelFile(ctx context.Context, filename string, disbursements []*disbursementModel.DisbursementWithTransactionResponse) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/generateDefaultExcelFile")
	defer segment.End()

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Disbursement History")
	if err != nil {
		s.logger.Error(ctx, "failed to add sheet", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	tz, err := util.GetTimeZoneFromContext(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to get timezone from context", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrInvalidTimeZone)
	}

	// Add Style
	style := xlsx.NewStyle()
	style.Font.Bold = true
	style.Fill = *xlsx.NewFill("solid", "E4E4E4", "E4E4E4") // Background color
	style.ApplyFill = true

	// Add headers
	headers := []string{"Type", "Created Date", "Destination Bank", "Destination Account", "Beneficiary Name",
		"Amount", "Fee", "Total Amount", "Remarks", "Reference ID", "Approval Status", "Approval Date", "Transaction Status",
		"Failed Reason", "Reject Reason", "Last Updated", "Transaction ID", "Bank Reference", "Created By", "Approval By"}
	row := sheet.AddRow()
	for _, header := range headers {
		cell := row.AddCell()
		cell.Value = header
		cell.SetStyle(style)
	}

	for _, dt := range disbursements {
		row := sheet.AddRow()

		approvedAt := "-"
		if dt.ApprovedAt != nil {
			approvedAt = dt.ApprovedAt.In(tz).Format(util.ReceiptFormatLayout)
		}

		disbursementType := constant.DisbursementTypeSingleTitle
		if dt.BulkID != nil {
			sliceBulkID := strings.Split(*dt.BulkID, "-")
			disbursementType = constant.DisbursementTypeBulkTitle + " " + sliceBulkID[len(sliceBulkID)-1]
		}

		beneficiaryBankName := ""
		if dt.BeneficiaryBankName != nil {
			beneficiaryBankName = *dt.BeneficiaryBankName
		}

		amount, _ := dt.Amount.Float64()
		fee, _ := dt.Fee.Float64()
		totalAmount, _ := dt.TotalAmount.Float64()

		remarks := ""
		if dt.Remark != nil {
			remarks = *dt.Remark
		}

		transactionStatus := ""
		if dt.TransactionStatus != nil {
			transactionStatus = *dt.TransactionStatus
		}

		bankReference := ""
		if dt.BankReferenceNo != nil {
			bankReference = *dt.BankReferenceNo
		}

		createdBy := ""
		if dt.CreatedBy != nil {
			createdBy = *dt.CreatedBy
		}

		approvedBy := ""
		if dt.ApprovedBy != nil {
			approvedBy = *dt.ApprovedBy
		}

		failedReason := ""
		if dt.FailedReason != nil {
			failedReason = *dt.FailedReason
		}

		rejectReason := ""
		if dt.RejectReason != nil {
			rejectReason = *dt.RejectReason
		}

		row.AddCell().Value = disbursementType                                     // TrxType
		row.AddCell().Value = dt.CreatedAt.In(tz).Format(util.ReceiptFormatLayout) // CreatedAt
		row.AddCell().Value = beneficiaryBankName                                  // Destination Bank
		row.AddCell().Value = dt.BeneficiaryAccountNo                              // Destination Account
		row.AddCell().Value = dt.BeneficiaryAccountName                            // Beneficiary Name
		row.AddCell().Value = util.ConvertFloatToCurrency(amount)                  // Amount
		row.AddCell().Value = util.ConvertFloatToCurrency(fee)                     // Fee
		row.AddCell().Value = util.ConvertFloatToCurrency(totalAmount)             // Total amount
		row.AddCell().Value = remarks                                              // Remark
		row.AddCell().Value = dt.ReferenceID                                       // Reference ID
		row.AddCell().Value = dt.Status                                            // Approval Status
		row.AddCell().Value = approvedAt                                           // Approval Date
		row.AddCell().Value = transactionStatus                                    // Transaction Status
		row.AddCell().Value = failedReason                                         // Failed Reason
		row.AddCell().Value = rejectReason                                         // Reject Reason
		row.AddCell().Value = dt.UpdatedAt.In(tz).Format(util.ReceiptFormatLayout) // Last Updated
		row.AddCell().Value = dt.UUID                                              // Transaction ID
		row.AddCell().Value = bankReference                                        // Bank reference
		row.AddCell().Value = createdBy                                            // Created By
		row.AddCell().Value = approvedBy                                           // Approved By
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

	if err := util.EnsureTmpDir(constant.ExportTempDir); err != nil {
		s.logger.Error(ctx, "failed to ensure /tmp directory", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	tempFilePath := filepath.Join(constant.ExportTempDir, filename+constant.DefaultExtXlsx)
	if err = file.Save(tempFilePath); err != nil {
		s.logger.Error(ctx, "failed to save file", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	return tempFilePath, nil
}

func (s *DisbursementService) generateBulkDisbursementExcelFile(ctx context.Context, filename string, disbursements []*disbursementModel.DisbursementWithTransactionResponse) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/generateBulkDisbursementExcelFile")
	defer segment.End()

	bankDB := bankTransfer.NewBankDB()
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Template")
	if err != nil {
		s.logger.Error(ctx, "failed to add sheet", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	// Add Style
	style := xlsx.NewStyle()
	style.Font.Bold = true
	style.Fill = *xlsx.NewFill("solid", "E4E4E4", "E4E4E4") // Background color
	style.ApplyFill = true

	// Add headers
	headers := []string{"Reference ID", "Amount", "Channel Code", "Account Number", "Account Name", "Remarks", "Status", "Failed Reason", "Reject Reason"}
	row := sheet.AddRow()
	for _, header := range headers {
		cell := row.AddCell()
		cell.Value = header
		cell.SetStyle(style)
	}

	for _, dt := range disbursements {
		row := sheet.AddRow()

		bank := bankDB.FindByCode(dt.BeneficiaryBankCode)

		remark := ""
		if dt.Remark != nil {
			remark = *dt.Remark
		}

		failedReason := ""
		if dt.FailedReason != nil {
			failedReason = *dt.FailedReason
		}

		rejectReason := ""
		if dt.RejectReason != nil {
			rejectReason = *dt.RejectReason
		}

		row.AddCell().Value = dt.ReferenceID
		row.AddCell().Value = dt.Amount.String()
		row.AddCell().Value = bank.ChannelCode
		row.AddCell().Value = dt.BeneficiaryAccountNo
		row.AddCell().Value = dt.BeneficiaryAccountName
		row.AddCell().Value = remark
		row.AddCell().Value = dt.Status
		row.AddCell().Value = failedReason
		row.AddCell().Value = rejectReason
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

	if err := util.EnsureTmpDir(constant.ExportTempDir); err != nil {
		s.logger.Error(ctx, "failed to ensure /tmp directory", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	tempFilePath := filepath.Join(constant.ExportTempDir, filename+constant.DefaultExtXlsx)
	if err = file.Save(tempFilePath); err != nil {
		s.logger.Error(ctx, "failed to save file", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	return tempFilePath, nil
}
