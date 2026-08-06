package xbPayoutService

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/xuri/excelize/v2"
)

func (s *xbPayoutService) ExportToExcel(ctx context.Context, request *xbModel.ExportXbPayoutRequest) (*xbModel.ExportXbPayoutResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/ExportToExcel")
	defer segment.End()

	var (
		disbursements   []*disbursementModel.DisbursementWithTransactionResponse
		page            int64 = 1
		perPage         int64 = 1000
		isNextPageExist       = true
	)

	disbursementFilter := &disbursementModel.GetDisbursementFilterRequest{
		MerchantID:     request.MerchantID,
		UUID:           request.UUID,
		StartCreatedAt: request.StartAt,
		EndCreatedAt:   request.EndAt,
		ReasonType:     request.Status,
		IsXbPayout:     true,
	}

	for {
		if list, err := s.disbursementRepo.GetList(ctx, disbursementFilter, page, perPage); err != nil {
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

		if !isNextPageExist {
			break
		}
	}

	filename := fmt.Sprintf("xb-payout-history_%s", util.GetCurrentTimeWithMillisFormatted())

	buffer, err := s.generateXbPayoutExcelFile(ctx, disbursements)
	if err != nil {
		return nil, err
	}

	signedUrl, err := s.uploadExcelFileToGCS(ctx, filename, buffer)
	if err != nil {
		return nil, err
	}

	return &xbModel.ExportXbPayoutResponse{
		Url: signedUrl,
	}, nil
}

func (s *xbPayoutService) uploadExcelFileToGCS(ctx context.Context, filename string, buffer *bytes.Buffer) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/uploadExcelFileToGCS")
	defer segment.End()

	gcsObjectName := "xb-payout-history/" + filename + constant.DefaultExtXlsx

	_, err := s.gcs.UploadFile(ctx, gcsObjectName, buffer, true)
	if err != nil {
		s.logger.Error(ctx, "failed to upload to GCS", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	signedUrl, err := s.gcs.CreateSignedURL(ctx, gcsObjectName, 15*time.Minute)
	if err != nil {
		s.logger.Error(ctx, "failed to create signed URL", logger.Error(err))
		return "", pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	return signedUrl, nil
}

func (s *xbPayoutService) generateXbPayoutExcelFile(ctx context.Context, disbursements []*disbursementModel.DisbursementWithTransactionResponse) (*bytes.Buffer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/xbPayout/generateXbPayoutExcelFile")
	defer segment.End()

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sw, err := f.NewStreamWriter(constant.DefaultSheetName)
	if err != nil {
		s.logger.Error(ctx, "failed to create stream writer", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	tz, err := util.GetTimeZoneFromContext(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to get timezone from context", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrInvalidTimeZone)
	}

	// Create header style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"E4E4E4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		s.logger.Error(ctx, "failed to create header style", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	// Create normal cell style
	cellStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		s.logger.Error(ctx, "failed to create cell style", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	headers := []string{
		"Reference ID", "Created Date", "Updated Date", "Status", "Destination Amount", "Conversion Rate",
		"Source Amount", "Transfer Fee", "Total Amount", "Purpose of Transfer", "Remarks",
		"Beneficiary Name", "Beneficiary Bank Name", "Beneficiary Account Number",
		"Beneficiary Local/SWIFT Code", "Beneficiary Account Type", "Beneficiary Country",
		"Beneficiary State", "Beneficiary City", "Beneficiary Address", "Beneficiary Postal Code",
		"Beneficiary Email", "Beneficiary Phone Number", "Sender Name", "Sender Identification Type",
		"Sender Identification Number", "Sender Account Type", "Sender Country", "Sender State",
		"Sender City", "Sender Address", "Sender Postal Code", "Sender Date of Birth",
		"Sender Bank Account Number", "Sender Source of Income",
	}

	// Set column widths
	for i := 1; i <= len(headers); i++ {
		_ = sw.SetColWidth(i, i, 20)
	}

	// Write header row
	headerCells := make([]any, len(headers))
	for i, header := range headers {
		headerCells[i] = excelize.Cell{StyleID: headerStyle, Value: header}
	}
	_ = sw.SetRow("A1", headerCells)

	// Write data rows
	for i, dt := range disbursements {
		xbDetail := dt.MetadataObj.XbDetail
		senderData := xbDetail.SenderData
		beneficiaryData := xbDetail.BeneficiaryData

		status := ""
		if dt.ReasonType != nil {
			status = *dt.ReasonType
		}

		remark := ""
		if dt.Remark != nil {
			remark = *dt.Remark
		}

		purposeCode := xbDetail.PurposeCode

		destinationAmount := dt.Amount.String()
		sourceAmount := xbDetail.SourceAmount.String()
		fxRate := xbDetail.FxRate.String()
		fee, _ := dt.Fee.Float64()
		totalAmount := xbDetail.TotalAmount.String()

		beneficiaryPhone := beneficiaryData.ContactCountryCode + beneficiaryData.ContactNumber

		rowData := []any{
			excelize.Cell{StyleID: cellStyle, Value: dt.ReferenceID},
			excelize.Cell{StyleID: cellStyle, Value: dt.CreatedAt.In(tz).Format(util.ReceiptFormatLayout)},
			excelize.Cell{StyleID: cellStyle, Value: dt.UpdatedAt.In(tz).Format(util.ReceiptFormatLayout)},
			excelize.Cell{StyleID: cellStyle, Value: status},
			excelize.Cell{StyleID: cellStyle, Value: destinationAmount},
			excelize.Cell{StyleID: cellStyle, Value: fxRate},
			excelize.Cell{StyleID: cellStyle, Value: sourceAmount},
			excelize.Cell{StyleID: cellStyle, Value: util.ConvertFloatToCurrency(fee)},
			excelize.Cell{StyleID: cellStyle, Value: totalAmount},
			excelize.Cell{StyleID: cellStyle, Value: purposeCode},
			excelize.Cell{StyleID: cellStyle, Value: remark},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.Name},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.BankName},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.AccountNumber},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.BankCode},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.AccountType},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.CountryName},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.State},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.City},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.Address},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.Postcode},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryData.Email},
			excelize.Cell{StyleID: cellStyle, Value: beneficiaryPhone},
			excelize.Cell{StyleID: cellStyle, Value: senderData.Name},
			excelize.Cell{StyleID: cellStyle, Value: senderData.IdentificationType},
			excelize.Cell{StyleID: cellStyle, Value: senderData.IdentificationNumber},
			excelize.Cell{StyleID: cellStyle, Value: senderData.AccountType},
			excelize.Cell{StyleID: cellStyle, Value: senderData.CountryName},
			excelize.Cell{StyleID: cellStyle, Value: senderData.State},
			excelize.Cell{StyleID: cellStyle, Value: senderData.City},
			excelize.Cell{StyleID: cellStyle, Value: senderData.Address},
			excelize.Cell{StyleID: cellStyle, Value: senderData.Postcode},
			excelize.Cell{StyleID: cellStyle, Value: senderData.Dob},
			excelize.Cell{StyleID: cellStyle, Value: senderData.BankAccountNumber},
			excelize.Cell{StyleID: cellStyle, Value: senderData.SourceOfIncome},
		}

		_ = sw.SetRow(fmt.Sprintf("A%d", i+2), rowData)
	}

	if err := sw.Flush(); err != nil {
		s.logger.Error(ctx, "failed to flush stream writer", logger.Error(err))
		return nil, pkgErrors.New(response.HttpErrInternal, constant.ErrFailedToGenerateExcel)
	}

	return f.WriteToBuffer()
}
