package paymentService

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"

	"github.com/xuri/excelize/v2"
)

var (
	methodDescription = map[string]string{
		"QRIS":            "QRIS",
		"VIRTUAL_ACCOUNT": "Virtual Account",
		"CREDIT_CARD":     "Cards",
	}

	methodTypeDescription = map[string]string{
		"DYNAMIC":        "Dynamic",
		"STATIC":         "Static",
		"OPEN_STATIC":    "Open Static",
		"CLOSED_STATIC":  "Closed Static",
		"CLOSED_DYNAMIC": "Closed Dynamic",
		"DEBIT":          "Debit",
		"CREDIT":         "Credit",
		"PREPAID":        "Prepaid",
	}
)

// Tech Debt:
// If the date parameters on the payment history page have been adjusted. Then this section also needs to be adjusted.
func (s *PaymentService) ExportToExcel(ctx context.Context, request *paymentModel.PaymentDownloadHistoryRequest, transactions []paymentModel.PaymentHistoryItem, wr io.Writer) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/payment/ExportToExcel")
	defer segment.End()

	f := excelize.NewFile()
	defer f.Close()

	sw, _ := f.NewStreamWriter("Sheet1")

	_ = sw.SetColWidth(1, 1, 20)
	_ = sw.SetColWidth(2, 5, 16)
	_ = sw.SetColWidth(6, 8, 20)
	_ = sw.SetColWidth(9, 9, 16)
	_ = sw.SetColWidth(10, 12, 30)
	_ = sw.SetColWidth(13, 17, 23)
	_ = sw.SetColWidth(18, 19, 15)
	_ = sw.SetColWidth(20, 20, 20)
	_ = sw.SetColWidth(21, 23, 20)

	startDate, _ := time.ParseInLocation(time.DateTime, request.StartDate+" 00:00:00", loc)
	endDate, _ := time.ParseInLocation(time.DateTime, request.EndDate+" 23:59:59", loc)

	_ = sw.SetRow("A1", []interface{}{
		excelize.Cell{StyleID: xlsx.StyleTitle(f), Value: "Payment History"},
	}, xlsx.DefaultRowOpt())

	_ = sw.SetRow("A2", []interface{}{
		excelize.Cell{StyleID: xlsx.StyleSubTitle(f), Value: util.DateStrMonthYear(startDate) + " - " + util.DateStrMonthYear(endDate)},
	}, xlsx.DefaultRowOpt())

	headers := make([]interface{}, len(downloadHeaders))
	for i, name := range downloadHeaders {
		headers[i] = excelize.Cell{StyleID: xlsx.StyleHeaderColumn(f), Value: name}
	}
	_ = sw.SetRow("A4", headers, xlsx.DefaultRowOpt())

	styleNormalRow := xlsx.StyleNormalRow(f)
	styleCurrencyRow := xlsx.StyleCurrencyRow(f)
	styleDatetimeRow := xlsx.StyleDatetime12HoursRow(f)
	for i, trx := range transactions {
		amount, _ := strconv.ParseFloat(trx.Amount, 64)

		rowPaidAt := excelize.Cell{StyleID: styleDatetimeRow}
		if trx.PaidAt != nil {
			rowPaidAt.Value = trx.PaidAt.In(loc)
		}
		rowExpiredAt := excelize.Cell{StyleID: styleDatetimeRow}
		if trx.ExpiredAt != nil {
			rowExpiredAt.Value = trx.ExpiredAt.In(loc)
		}
		rowPaidAmount := excelize.Cell{StyleID: styleCurrencyRow}
		if trx.AmountPaid != nil {
			rowPaidAmount.Value, _ = strconv.ParseFloat(*trx.AmountPaid, 64)
		}
		rowCardExpiryDate := excelize.Cell{StyleID: styleDatetimeRow, Value: "  -"}
		if trx.Method == constant.ChannelCreditCard {
			trx.MethodType = trx.CardType
			rowCardExpiryDate.Value = trx.CardExpiry
		}

		// Refund info
		rowRefundDate := excelize.Cell{StyleID: styleDatetimeRow, Value: "  -"}
		if trx.RefundDate != nil {
			rowRefundDate.Value = trx.RefundDate.In(loc)
		}
		rowRefundAmount := excelize.Cell{StyleID: styleCurrencyRow, Value: "  -"}
		if trx.RefundAmount != nil {
			if amount, err := strconv.ParseFloat(*trx.RefundAmount, 64); err == nil {
				rowRefundAmount.Value = amount
			}
		}
		rowRefundStatus := excelize.Cell{StyleID: styleNormalRow, Value: "  -"}
		if trx.RefundStatus != nil {
			rowRefundStatus.Value = util.ToTitle(*trx.RefundStatus)
		}

		// Acquiring Bank and MID - only show for Bank Account settlement + Direct PSP
		acquiringBank := ""
		mid := ""

		if trx.MIDType == constant.PaymentMethodChannelTypeDirect {
			acquiringBank = trx.MIDAcquirer
			mid = trx.MID
		}

		_ = sw.SetRow(fmt.Sprintf("A%d", 5+i), []interface{}{
			excelize.Cell{StyleID: styleDatetimeRow, Value: trx.CreatedAt.In(loc)},               // "Created Date"
			excelize.Cell{StyleID: styleNormalRow, Value: methodDescription[trx.Method]},         // "Method"
			excelize.Cell{StyleID: styleNormalRow, Value: methodTypeDescription[trx.MethodType]}, // "Type"
			excelize.Cell{StyleID: styleCurrencyRow, Value: amount},                              // "Bill Amount"
			excelize.Cell{StyleID: styleNormalRow, Value: util.ToTitle(trx.Status)},              // "Payment Status"
			rowPaidAt, // "Payment Date"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.Channel}, // "Payment Channel"
			rowExpiredAt,  // "Expiry Time"
			rowPaidAmount, // "Paid Amount"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.CustomerId},                   // "Customer No"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.ReferenceID},                  // "Reference ID"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.UUID},                         // "Transaction ID"
			excelize.Cell{StyleID: styleNormalRow, Value: util.ValueOfPtr(trx.RecurringID)}, // "Recurring Contract ID"
			excelize.Cell{StyleID: styleNormalRow, Value: "  -"},                            // "Bank Reference"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.VirtualAccountName},           // "VA Name"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.VirtualAccountNo},             // "VA Number"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.QrisMerchantName},             // "QR Merchant Name"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.QrisURL},                      // "QR URL"
			excelize.Cell{StyleID: styleNormalRow, Value: acquiringBank},                    // "Acquiring Bank"
			excelize.Cell{StyleID: styleNormalRow, Value: mid},                              // "MID"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.CardIssuerBank},               // "Card Issuer Bank"
			excelize.Cell{StyleID: styleNormalRow, Value: trx.CardNumber},                   // "Card Number"
			rowCardExpiryDate, // "Card Expiry Date"
			rowRefundDate,     // "Refund Date"
			rowRefundAmount,   // "Refund Amount"
			rowRefundStatus,   // "Refund Status"
		})
	}

	if err := sw.Flush(); err != nil {
		return pkgErrs.New(response.HttpErrInternal, fmt.Errorf("flush stream writer: %w", err))
	}

	return f.Write(wr)
}
