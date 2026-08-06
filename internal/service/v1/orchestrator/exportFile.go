package orchestrator_service

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/xuri/excelize/v2"
)

var trxHistoryColumns = []string{
	"Reference ID", "Balance Type", "Transaction Type", "Channel", "Created Date", "Settlement Date", "Created By", "Transaction ID", "Amount", "Fee", "Bank Reference", "Transaction Status", "Reason Type", "Reason Description", "Remarks", "Destination Bank", "Destination Account", "Beneficiary Name",
}
var customCurrencyFmt = "[$Rp -421]#,##0.00"

const downloadedFileName = "balance_history.xlsx"

func (s *OrchestratorService) GenExcelForTransactionHistories(ctx context.Context, w *orchestrator_model.FileWriter, req *orchestrator_model.TransactionHistoryFilterRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/orchestrator/GenExcelForTransactionHistories")
	defer segment.End()

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)

	var (
		err          error
		transactions []orchestrator_model.TransactionHistory
	)
	if constant.IsEnableBalanceHistoryViaDataReporting(req.MerchantID) {
		transactions, err = s.reportingRepo.ExportBalanceHistory(ctx, req)
	} else {
		transactions, err = s.accountTransactionRepo.GetListTransactionHistories(ctx, req)
	}
	if err != nil {
		s.logger.Error(ctx, "get list transaction histories", logger.Error(err))
		return pkgErrs.New(response.HttpErrDatabase, fmt.Errorf(constant.InternalErrorFmt, traceId))
	}

	f := excelize.NewFile()
	defer f.Close()

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return pkgErrs.New(response.HttpErrInternal, fmt.Errorf("write stream writer: %w", err))
	}
	// Set column widths for all the new columns
	_ = sw.SetColWidth(1, 8, 20)   // Reference ID to Transaction ID
	_ = sw.SetColWidth(9, 10, 15)  // Amount and Fee
	_ = sw.SetColWidth(11, 11, 20) // Bank Reference
	_ = sw.SetColWidth(12, 12, 15) // Transaction Status
	_ = sw.SetColWidth(13, 15, 25) // Reason Type, Description, Remarks
	_ = sw.SetColWidth(16, 18, 23) // Destination Bank, Account, Beneficiary

	stTitle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true, Family: "Arial", Size: 12,
		},
	})
	stSubTitle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true, Family: "Arial", Size: 10,
		},
	})
	stHdColumn, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true, Family: "Arial", Size: 10,
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill: excelize.Fill{
			Type: "pattern", Pattern: 1, Color: []string{"#e4e4e4"},
		},
	})
	stColumnData, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: false, Family: "Arial", Size: 10,
		},
	})
	stColumnDataCurrency, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: false, Family: "Arial", Size: 10,
		},
		CustomNumFmt: &customCurrencyFmt,
	})
	_ = sw.SetRow("A1", []interface{}{
		excelize.Cell{StyleID: stTitle, Value: "Settlement Report - Harsya"},
	})
	_ = sw.SetRow("A2", []interface{}{
		excelize.Cell{StyleID: stSubTitle, Value: util.DateStrMonthYear(req.StartDate) + " - " + util.DateStrMonthYear(req.EndDate)},
	})

	headerColumns := make([]interface{}, len(trxHistoryColumns))
	for i, name := range trxHistoryColumns {
		headerColumns[i] = excelize.Cell{StyleID: stHdColumn, Value: name}
	}
	_ = sw.SetRow("A4", headerColumns)

	for i, t := range transactions {
		// Format dates
		settlementAt := "-"
		if t.SettlementAt != nil {
			settlementAt = util.DateStrMonthYearHour(*t.SettlementAt)
		}

		// Format created date - CreatedAt is not a pointer in the model
		createdAt := util.DateStrMonthYearHour(t.CreatedAt)

		_ = sw.SetRow(fmt.Sprintf("A%d", 5+i), []interface{}{
			excelize.Cell{StyleID: stColumnData, Value: t.MerchantReferenceID},
			excelize.Cell{StyleID: stColumnData, Value: t.BalanceType},
			excelize.Cell{StyleID: stColumnData, Value: orchestrator_model.TransactionTypeForUser(t.Type, t.Channel)},
			excelize.Cell{StyleID: stColumnData, Value: orchestrator_model.FormatChannelName(t.Channel)},
			excelize.Cell{StyleID: stColumnData, Value: createdAt},
			excelize.Cell{StyleID: stColumnData, Value: settlementAt},
			excelize.Cell{StyleID: stColumnData, Value: t.CreatedBy},
			excelize.Cell{StyleID: stColumnData, Value: t.Id},
			excelize.Cell{StyleID: stColumnDataCurrency, Value: t.Amount},
			excelize.Cell{StyleID: stColumnDataCurrency, Value: t.Fee},
			excelize.Cell{StyleID: stColumnData, Value: t.BankReference},
			excelize.Cell{StyleID: stColumnData, Value: t.Status},
			excelize.Cell{StyleID: stColumnData, Value: t.ReasonType},
			excelize.Cell{StyleID: stColumnData, Value: t.ReasonDescription},
			excelize.Cell{StyleID: stColumnData, Value: t.Remarks},
			excelize.Cell{StyleID: stColumnData, Value: t.BeneficiaryBankName},
			excelize.Cell{StyleID: stColumnData, Value: t.BeneficiaryAccountNo},
			excelize.Cell{StyleID: stColumnData, Value: t.BeneficiaryAccountName},
		})
	}
	if err := sw.Flush(); err != nil {
		return pkgErrs.New(response.HttpErrInternal, fmt.Errorf("flush stream writer: %w", err))
	}

	w.WriteHeader(downloadedFileName)
	return f.Write(w)
}
