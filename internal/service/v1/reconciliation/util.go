package reconciliation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	reconModel "github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/xlsx"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

const (
	sheetNameToUpload  = "To upload"
	maxRowDataToUpload = 1000

	columnTransactionDatetime   = 0
	columnTransactionReference  = 1
	columnTransactionReference2 = 2
	columnAmount                = 3
	columnBank                  = 4
	columnChannel               = 5
)

var (
	bulkUploadHeaders = map[int]string{
		columnTransactionDatetime:   "Transaction date & time",
		columnTransactionReference:  "Transaction reference used for recon",
		columnTransactionReference2: "Transaction reference used for recon_2",
		columnAmount:                "Transaction amount",
		columnBank:                  "Partner",
		columnChannel:               "Channel",
	}
)

// getRowsAndValidateBulkUpload retrieves the rows from the specified sheet in the provided XLSX file,
// validates the headers, and ensures the number of rows is within the allowed limit.
// It returns the rows as a 2D slice of strings, or an error if any validation fails.
func (*ReconciliationService) getRowsAndValidateBulkUpload(f xlsx.Filer) ([][]string, error) {
	rows, err := f.GetRows(sheetNameToUpload, xlsx.Options{RawCellValue: true})
	if err != nil {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("sheet to upload not found"))

	} else if len(rows) < 2 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("empty data to upload"))

	} else if len(rows) > maxRowDataToUpload+1 {
		return nil, pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("max row data is 1000"))
	}
	// Headers validation
	for idx, name := range rows[0] {
		if _, ok := bulkUploadHeaders[idx]; !ok {
			continue

		} else if bulkUploadHeaders[idx] != strings.TrimSpace(name) {
			return nil, constant.ErrHeaderColumnDoesNotMatchWithTemplate
		}
	}
	return rows, nil
}

// convertRowsToUploadedTransaction converts a row of data from a bulk upload sheet into a Transaction model.
// It parses the transaction date and time from the provided string, and returns the populated Transaction struct.
// If there is an error parsing the date and time, the error is returned.
func convertRowsToUploadedTransaction(col []string, f xlsx.Filer, i int) (reconModel.Transaction, error) {
	reference := util.SliceExtractOrDefault(col, columnTransactionReference, "")
	reference2 := util.SliceExtractOrDefault(col, columnTransactionReference2, "")
	amount := util.SliceExtractOrDefault(col, columnAmount, "")
	bank := util.SliceExtractOrDefault(col, columnBank, "")
	channel := util.SliceExtractOrDefault(col, columnChannel, "")

	if channel == "" {
		return reconModel.Transaction{}, errors.New("invalid channel")
	}

	// Handle empty or invalid amount values to prevent panic
	if amount == "" {
		amount = "0.00"
	}

	parsedAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return reconModel.Transaction{}, fmt.Errorf("invalid amount value '%s': %w", amount, err)
	}

	uploadedTransactions := reconModel.Transaction{
		Reference:  reference,
		Reference2: reference2,
		Amount:     parsedAmount,
		Bank:       bank,
		Channel:    convertChannel(channel),
	}

	dateTime, err := extractDatetime(f, i)
	if err != nil {
		return uploadedTransactions, err
	}

	uploadedTransactions.TransactionDate = dateTime

	return uploadedTransactions, nil
}

func extractDatetime(f xlsx.Filer, i int) (time.Time, error) {
	timeLayout := constant.ReconTimeFormat + "-07:00"
	loc, _ := time.LoadLocation("Asia/Jakarta")
	excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, loc)

	strDatetime, err := f.GetCellValue(sheetNameToUpload, fmt.Sprintf("A%d", i+2), excelize.Options{
		RawCellValue: true,
	})
	if err != nil {
		return excelEpoch, err
	}

	epoch, err := strconv.ParseFloat(strDatetime, 64)
	if err == nil {
		days := int(epoch)
		fractionalDays := epoch - float64(days)
		seconds := int(fractionalDays * 86400)

		datetime := excelEpoch.AddDate(0, 0, days).Add(time.Second * time.Duration(seconds))
		return datetime.UTC(), err
	}

	//parse string to time.Time
	datetime, err := time.Parse(timeLayout, strDatetime+"+07:00")
	if err != nil {
		return excelEpoch, errors.New("invalid datetime format. format should be November 1, 2024 5:22 AM")
	}
	return datetime, err
}

func convertChannel(rowVal string) string {
	switch strings.ToUpper(rowVal) {
	case "QRIS":
		return constant.ChannelQris
	case "VA", "VIRTUAL_ACCOUNT":
		return constant.ChannelVirtualAccount
	case "CC", "CARDS", "CARD", "CREDIT_CARD":
		return constant.ChannelCard
	case "DISBURSEMENT", "BANK_TRANSFER":
		return constant.ChannelBankTransfer

	}
	return rowVal
}
