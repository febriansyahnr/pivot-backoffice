package paymentService

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/xuri/excelize/v2"
)

var (
	investigationDownloadHeaders = []string{
		"Payment Reference ID",
		"Amount",
		"Currency",
		"Merchant Name",
		"Payment Method",
		"Payment Channel",
		"Payment Status",
		"Investigation Status",
		"Started At",
		"Last Updated At",
		"Completed At",
		"Notes",
	}

	investigationStatusDescription = map[string]string{
		"INVESTIGATION_IN_PROCESS": "In Process",
		"INVESTIGATION_SUCCESS":    "Success",
		"INVESTIGATION_FAILED":     "Failed",
	}
)

func (s *PaymentService) ExportInvestigatedPayments(ctx context.Context, request *paymentModel.InvestigationDownloadHistoryRequest) (*paymentModel.InvestigationDownloadHistoryResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/payment/ExportInvestigatedPayments")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	internalErr := pkgErrs.New(response.HttpErrInternal, fmt.Errorf(constant.InternalErrorFmt, traceId))

	// Convert request to filter
	filter := &paymentModel.GetInvestigatedPaymentsFilterRequest{
		MerchantID:          request.MerchantId,
		InvestigationStatus: request.InvestigationStatus,
		PaymentReferenceID:  request.PaymentReferenceID,
		PaymentMethod:       request.PaymentMethod,
		Channel:             request.Channel,
		FromDate:            request.FromDate,
		ToDate:              request.ToDate,
		Page:                1,
		Limit:               10000, // Get all records for export
	}

	// Get investigated payments
	result, err := s.GetInvestigatedPayments(ctx, filter)
	if err != nil {
		s.logger.Error(ctx, "Failed when getting investigated payments data", logger.Error(err))
		return nil, internalErr
	}

	payments, ok := result.Data.([]*paymentModel.InvestigatedPaymentResponse)
	if !ok {
		s.logger.Error(ctx, "Failed to cast investigated payments data")
		return nil, internalErr
	}

	// Create Excel file
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	if err := s.exportInvestigationToExcel(ctx, request, payments, buf); err != nil {
		s.logger.Error(ctx, "Failed when generate file excel", logger.Error(err))
		return nil, internalErr
	}

	// Upload to GCS
	var (
		objectName       = fmt.Sprintf("downloads/investigation-histories/%s-%d.xlsx", request.MerchantId, time.Now().Unix())
		downloadFilename = fmt.Sprintf("attachment; filename=investigation_histories_%d.xlsx", time.Now().In(loc).Unix())
	)

	if _, err := s.gcs.UploadFile(ctx, objectName, buf, true, gcs.WriteContentDisposition(downloadFilename)); err != nil {
		s.logger.Error(ctx, "Failed when upload file to gcs", logger.Error(err))
		return nil, internalErr
	}

	signedURL, err := s.gcs.CreateSignedURL(ctx, objectName, expires)
	if err != nil {
		s.logger.Error(ctx, "Failed when create signed URL", logger.Error(err))
		return nil, internalErr
	}

	return &paymentModel.InvestigationDownloadHistoryResponse{URL: signedURL}, nil
}

func (s *PaymentService) exportInvestigationToExcel(ctx context.Context, request *paymentModel.InvestigationDownloadHistoryRequest, payments []*paymentModel.InvestigatedPaymentResponse, wr io.Writer) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/exportInvestigationToExcel")
	defer segment.End()

	f := excelize.NewFile()
	defer f.Close()

	sw, _ := f.NewStreamWriter("Sheet1")

	// Set column widths
	_ = sw.SetColWidth(1, 1, 30)  // Payment Reference ID
	_ = sw.SetColWidth(2, 2, 15)  // Amount
	_ = sw.SetColWidth(3, 3, 10)  // Currency
	_ = sw.SetColWidth(4, 4, 25)  // Merchant Name
	_ = sw.SetColWidth(5, 5, 18)  // Payment Method
	_ = sw.SetColWidth(6, 6, 20)  // Payment Channel
	_ = sw.SetColWidth(7, 7, 18)  // Payment Status
	_ = sw.SetColWidth(8, 8, 20)  // Investigation Status
	_ = sw.SetColWidth(9, 9, 20)  // Started At
	_ = sw.SetColWidth(10, 10, 20) // Last Updated At
	_ = sw.SetColWidth(11, 11, 20) // Completed At
	_ = sw.SetColWidth(12, 12, 40) // Notes

	// Title
	_ = sw.SetRow("A1", []interface{}{
		excelize.Cell{StyleID: xlsx.StyleTitle(f), Value: "Investigation History"},
	}, xlsx.DefaultRowOpt())

	// Date range subtitle
	dateRange := "All Time"
	if request.FromDate != nil && request.ToDate != nil {
		dateRange = util.DateStrMonthYear(*request.FromDate) + " - " + util.DateStrMonthYear(*request.ToDate)
	} else if request.FromDate != nil {
		dateRange = "From " + util.DateStrMonthYear(*request.FromDate)
	} else if request.ToDate != nil {
		dateRange = "Until " + util.DateStrMonthYear(*request.ToDate)
	}

	_ = sw.SetRow("A2", []interface{}{
		excelize.Cell{StyleID: xlsx.StyleSubTitle(f), Value: dateRange},
	}, xlsx.DefaultRowOpt())

	// Headers
	headers := make([]interface{}, len(investigationDownloadHeaders))
	for i, name := range investigationDownloadHeaders {
		headers[i] = excelize.Cell{StyleID: xlsx.StyleHeaderColumn(f), Value: name}
	}
	_ = sw.SetRow("A4", headers, xlsx.DefaultRowOpt())

	// Data rows
	styleNormalRow := xlsx.StyleNormalRow(f)
	styleCurrencyRow := xlsx.StyleCurrencyRow(f)
	styleDatetimeRow := xlsx.StyleDatetime12HoursRow(f)

	for i, payment := range payments {
		amount, _ := payment.Amount.Float64()

		rowStartedAt := excelize.Cell{StyleID: styleDatetimeRow, Value: "  -"}
		if payment.StartedAt != nil {
			rowStartedAt.Value = payment.StartedAt.In(loc)
		}

		rowCompletedAt := excelize.Cell{StyleID: styleDatetimeRow, Value: "  -"}
		if payment.CompletedAt != nil {
			rowCompletedAt.Value = payment.CompletedAt.In(loc)
		}

		rowNotes := excelize.Cell{StyleID: styleNormalRow, Value: "  -"}
		if payment.Notes != nil && *payment.Notes != "" {
			rowNotes.Value = *payment.Notes
		}

		investigationStatusLabel := investigationStatusDescription[payment.InvestigationStatus]
		if investigationStatusLabel == "" {
			investigationStatusLabel = payment.InvestigationStatus
		}

		_ = sw.SetRow(fmt.Sprintf("A%d", 5+i), []interface{}{
			excelize.Cell{StyleID: styleNormalRow, Value: payment.PaymentReferenceID},
			excelize.Cell{StyleID: styleCurrencyRow, Value: amount},
			excelize.Cell{StyleID: styleNormalRow, Value: payment.Currency},
			excelize.Cell{StyleID: styleNormalRow, Value: payment.MerchantName},
			excelize.Cell{StyleID: styleNormalRow, Value: methodDescription[payment.PaymentMethod]},
			excelize.Cell{StyleID: styleNormalRow, Value: payment.PaymentChannel},
			excelize.Cell{StyleID: styleNormalRow, Value: util.ToTitle(payment.PaymentStatus)},
			excelize.Cell{StyleID: styleNormalRow, Value: investigationStatusLabel},
			rowStartedAt,
			excelize.Cell{StyleID: styleDatetimeRow, Value: payment.LastUpdatedAt.In(loc)},
			rowCompletedAt,
			rowNotes,
		})
	}

	if err := sw.Flush(); err != nil {
		return pkgErrs.New(response.HttpErrInternal, fmt.Errorf("flush stream writer: %w", err))
	}

	return f.Write(wr)
}
