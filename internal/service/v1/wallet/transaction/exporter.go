package walletTransaction

import (
	"bytes"
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/wallet/transaction"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"

	"github.com/xuri/excelize/v2"
)

func (service) ExportExcelMerchantTransactionHistoryList(ctx context.Context, request model.MerchantTransactionHistoryListReq, transactions []model.MerchantTransactionHistoryListResp) (*bytes.Buffer, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/wallet/transaction/exportExcelMerchantTransactionHistoryList")
	defer segment.End()

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sw, err := f.NewStreamWriter(constant.DefaultSheetName)
	if err != nil {
		return nil, err
	}
	loc := util.GetTimeLocationFromContext(ctx)

	startDate := request.StartDate.In(loc)
	endDate := request.EndDate.In(loc)

	_ = sw.MergeCell("A2", "C2")
	_ = sw.MergeCell("A3", "C3")
	_ = sw.SetColWidth(1, 2, 20)
	_ = sw.SetColWidth(3, 3, 23)
	_ = sw.SetColWidth(4, 4, 18)
	_ = sw.SetColWidth(5, 6, 36)
	_ = sw.SetColWidth(7, 7, 15)
	_ = sw.SetColWidth(8, 9, 20)
	_ = sw.SetColWidth(10, 10, 15)

	styleSubTitle := xlsx.StyleSubTitle(f)
	styleHeaderColumn := xlsx.StyleHeaderColumn(f)
	styleValDatetime12Hour := xlsx.StyleDatetime12HoursRow(f)
	styleValString := xlsx.StyleNormalRow(f)
	styleValCurrency := xlsx.StyleCurrencyRowWithCustomFormat(f, xlsx.CustomNormalCurrencyWithColorFmt)

	_ = sw.SetRow("A2", []any{excelize.Cell{StyleID: xlsx.StyleTitleWithAlignment(f), Value: "Wallet Transaction History"}}, excelize.RowOpts{
		Height: 20,
	})
	_ = sw.SetRow("A3", []any{
		excelize.Cell{StyleID: styleSubTitle, Value: util.DateStrMonthYear(startDate) + " - " + util.DateStrMonthYear(endDate)},
	}, xlsx.DefaultRowOpt())

	rows := 3
	if request.Type != "" {
		rows++
		_ = sw.SetRow(fmt.Sprintf("A%d", rows), []any{
			excelize.Cell{StyleID: styleSubTitle, Value: "Transaction Type:"},
			excelize.Cell{StyleID: styleValString, Value: model.TransactionTypeToTitle(request.Type)},
		}, xlsx.DefaultRowOpt())
	}
	if request.Status != "" {
		rows++
		_ = sw.SetRow(fmt.Sprintf("A%d", rows), []any{
			excelize.Cell{StyleID: styleSubTitle, Value: "Transaction Status:"},
			excelize.Cell{StyleID: styleValString, Value: util.ToTitle(request.Status)},
		}, xlsx.DefaultRowOpt())
	}
	if request.Id != "" {
		rows++
		_ = sw.SetRow(fmt.Sprintf("A%d", rows), []any{
			excelize.Cell{StyleID: styleSubTitle, Value: "Transaction Id:"},
			excelize.Cell{StyleID: styleValString, Value: request.Id},
		}, xlsx.DefaultRowOpt())
	}

	headerNames := []string{
		"Created Date", "Last Update", "Transaction Type", "Created By", "Transaction ID", "Reference ID", "Amount", "Bank Name", "Account Number", "Status",
	}

	headers := make([]any, len(headerNames))
	for i := range headerNames {
		headers[i] = excelize.Cell{StyleID: styleHeaderColumn, Value: headerNames[i]}
	}
	rows += 2
	_ = sw.SetRow(fmt.Sprintf("A%d", rows), headers)

	for i, trx := range transactions {
		_ = sw.SetRow(
			fmt.Sprintf("A%d", (rows+1)+i), []any{
				excelize.Cell{StyleID: styleValDatetime12Hour, Value: trx.CreatedAt.In(loc)},
				excelize.Cell{StyleID: styleValDatetime12Hour, Value: trx.UpdatedAt.In(loc)},
				excelize.Cell{StyleID: styleValString, Value: trx.TransactionTypeToTitle()},
				excelize.Cell{StyleID: styleValString, Value: trx.CreatedBy},
				excelize.Cell{StyleID: styleValString, Value: trx.Id},
				excelize.Cell{StyleID: styleValString, Value: trx.GetReferenceID()},
				excelize.Cell{StyleID: styleValCurrency, Value: trx.Amount},
				excelize.Cell{StyleID: styleValString, Value: trx.BankAccountName},
				excelize.Cell{StyleID: styleValString, Value: trx.BankAccountNumber},
				excelize.Cell{StyleID: styleValString, Value: util.ToTitle(trx.SettlementStatus)},
			},
		)
	}
	if err := sw.Flush(); err != nil {
		return nil, err
	}
	return f.WriteToBuffer()
}
