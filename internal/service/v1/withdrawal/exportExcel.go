package withdrawalService

import (
	"context"
	"fmt"
	"io"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"

	"github.com/xuri/excelize/v2"
)

func (s *withdrawalService) ExportToExcel(ctx context.Context, req *withdrawal.WithdrawalListRequest, transactions []withdrawal.WithdrawalHistoryResponse, wr io.Writer) error {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/withdrawal/ExportToExcel")
	defer segment.End()

	f := excelize.NewFile()
	defer f.Close()

	sw, _ := f.NewStreamWriter("Sheet1")

	_ = sw.SetColWidth(1, 3, 20)
	_ = sw.SetColWidth(4, 4, 20)
	_ = sw.SetColWidth(5, 5, 15)
	_ = sw.SetColWidth(6, 6, 35)
	_ = sw.SetColWidth(7, 10, 25)

	_ = sw.SetRow("A1", []interface{}{
		excelize.Cell{StyleID: xlsx.StyleTitle(f), Value: "Payment Balance Withdrawal History"},
	}, xlsx.DefaultRowOpt())

	_ = sw.SetRow("A2", []interface{}{
		excelize.Cell{StyleID: xlsx.StyleSubTitle(f), Value: util.DateStrMonthYear(req.StartDate) + " - " + util.DateStrMonthYear(req.EndDate)},
	}, xlsx.DefaultRowOpt())

	headers := make([]interface{}, len(downloadWithdrawalHeaders))
	for i, name := range downloadWithdrawalHeaders {
		headers[i] = excelize.Cell{StyleID: xlsx.StyleHeaderColumn(f), Value: name}
	}
	_ = sw.SetRow("A4", headers, xlsx.DefaultRowOpt())

	styleNormalRow := xlsx.StyleNormalRow(f)
	styleCurrencyRow := xlsx.StyleCurrencyRow(f)
	styleDatetimeRow := xlsx.StyleDatetime12HoursRow(f)
	for i, trx := range transactions {
		_ = sw.SetRow(fmt.Sprintf("A%d", 5+i), []interface{}{
			excelize.Cell{StyleID: styleDatetimeRow, Value: trx.Date.In(loc)},         // Created Date
			excelize.Cell{StyleID: styleDatetimeRow, Value: trx.UpdatedAt.In(loc)},    // Last Update
			excelize.Cell{StyleID: styleNormalRow, Value: trx.CreatedBy},              // Created By
			excelize.Cell{StyleID: styleCurrencyRow, Value: trx.Amount},               // Amount
			excelize.Cell{StyleID: styleNormalRow, Value: util.ToTitle(trx.Status)},   // Status
			excelize.Cell{StyleID: styleNormalRow, Value: trx.Id},                     // Transaction ID
			excelize.Cell{StyleID: styleNormalRow, Value: trx.BankReference},          // Bank Reference
			excelize.Cell{StyleID: styleNormalRow, Value: trx.BeneficiaryBankName},    // Bank Name
			excelize.Cell{StyleID: styleNormalRow, Value: trx.BeneficiaryAccountNo},   // Account Number
			excelize.Cell{StyleID: styleNormalRow, Value: trx.BeneficiaryAccountName}, // Beneficiary Name
		})
	}

	if err := sw.Flush(); err != nil {
		return pkgErrs.New(response.HttpErrInternal, fmt.Errorf("flush stream writer: %w", err))
	}

	return f.Write(wr)
}
