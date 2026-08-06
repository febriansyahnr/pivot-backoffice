package walletTransaction

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/wallet/transaction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestExportExcelMerchantTransactionHistoryList(t *testing.T) {
	request := model.MerchantTransactionHistoryListReq{
		Type:      "MERCHANT_TOP_UP",
		Status:    "SUCCESS",
		Id:        "123456",
		StartDate: time.Date(2025, 3, 15, 17, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 3, 17, 16, 59, 59, 0, time.UTC),
	}
	transactions := []model.MerchantTransactionHistoryListResp{{
		Id:               "123456",
		Type:             "MERCHANT_TOP_UP",
		Channel:          "BANK_TRANSFER",
		CreatedAt:        time.Date(2025, 3, 17, 7, 40, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2025, 3, 17, 7, 41, 0, 0, time.UTC),
		Amount:           10_000,
		Status:           "",
		SettlementStatus: "SUCCESS",
		CreatedBy:        "John Doe",
		ReferenceId:      "-",
	}}

	service := &service{}

	buff, err := service.ExportExcelMerchantTransactionHistoryList(context.Background(), request, transactions)
	require.NoError(t, err)
	require.NotNil(t, buff)

	f, err := excelize.OpenReader(buff)
	require.NoError(t, err)
	defer func() { require.NoError(t, f.Close()) }()

	_, err = f.GetSheetIndex(constant.DefaultSheetName)
	require.NoError(t, err)

	values := map[string]string{
		"A2": "Wallet Transaction History",
		"A3": "16 March 2025 - 17 March 2025",
		"A4": "Transaction Type:",
		"B4": "Merchant Top Up",
		"A5": "Transaction Status:",
		"B5": "Success",
		"A6": "Transaction Id:",
		"B6": "123456",
		"A7": "", "B7": "", "C7": "",
		"A8": "Created Date",
		"B8": "Last Update",
		"C8": "Transaction Type",
		"D8": "Created By",
		"E8": "Transaction ID",
		"F8": "Reference ID",
		"G8": "Amount",
		"H8": "Bank Name",
		"I8": "Account Number",
		"J8": "Status",
		"A9": "17 Mar 2025, 02:40 PM",
		"B9": "17 Mar 2025, 02:41 PM",
		"C9": "Merchant Top Up",
		"D9": "John Doe",
		"E9": "123456",
		"F9": "-",
		"G9": "+ Rp 10,000",
		"H9": "",
		"I9": "",
		"J9": "Success",
	}
	for cell, want := range values {
		result, err := f.GetCellValue(constant.DefaultSheetName, cell)
		assert.NoError(t, err)
		assert.Equal(t, want, result)
	}
}
