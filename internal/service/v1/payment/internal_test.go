package paymentService

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"os"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestGetCacheDownloadHistory(t *testing.T) {
	rdb, rdbMock := redismock.NewClientMock()

	service := &PaymentService{
		redis: redisExt.WrapRedisClient(rdb, nil),
	}

	hash := "682448b94f95589c9ae13e6d2872d527f8e30c8c642df94800129d43f667bdca"
	redisKey := fmt.Sprintf(c.RedisKeyDownloadPaymentHistoryFmt, hash)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult string
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				rdbMock.ExpectGet(redisKey).SetErr(c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS:Cache not found",
			setupMock: func() {
				rdbMock.ExpectGet(redisKey).SetErr(redisExt.ErrNil)
			},
		},
		{
			name: "SUCCESS:Cache found",
			setupMock: func() {
				rdbMock.ExpectGet(redisKey).SetVal("https://cache-found")
			},
			wantResult: "https://cache-found",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rdbMock.ClearExpect()

			test.setupMock()
			result, err := service.GetCacheDownloadHistory(context.Background(), hash)
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestExportToExcel(t *testing.T) {
	service := &PaymentService{}

	request := &paymentModel.PaymentDownloadHistoryRequest{
		StartDate: "2024-11-01",
		EndDate:   "2024-11-17",
	}
	sheetName := "Sheet1"
	headers := map[string]string{
		"A4": "Created Date",
		"B4": "Method",
		"C4": "Type",
		"D4": "Bill Amount",
		"E4": "Payment Status",
		"F4": "Payment Date",
		"G4": "Payment Channel",
		"H4": "Expiry Time",
		"I4": "Paid Amount",
		"J4": "Customer No",
		"K4": "Reference ID",
		"L4": "Transaction ID",
		"M4": "Recurring ID",
		"N4": "Bank Reference",
		"O4": "VA Name",
		"P4": "VA Number",
		"Q4": "QR Merchant Name",
		"R4": "QR URL",
		"S4": "Acquiring Bank",
		"T4": "Bank Merchant ID (MID)",
		"U4": "Card Issuer Bank",
		"V4": "Card Number",
		"W4": "Card Expiry Date",
		"X4": "Refund Date",
		"Y4": "Refund Amount",
		"Z4": "Refund Status",
	}

	tests := []struct {
		name             string
		transactions     []paymentModel.PaymentHistoryItem
		wantExcelContent map[string]string
	}{
		{
			name:         "No transactions",
			transactions: []paymentModel.PaymentHistoryItem{},
			wantExcelContent: map[string]string{
				"A1": "Payment History",
				"A2": "01 November 2024 - 17 November 2024",
				"A3": "",
			},
		},
		{
			name: "Transactions found",
			transactions: []paymentModel.PaymentHistoryItem{
				{
					UUID:           "f378a8ba-ef6e-41eb-8081-7ccd62bb9ede",
					ReferenceID:    "1731622233KYDWX",
					RecurringID:    util.ValueToPtr("f378a8ba-ef6e-41eb-8081-7ccd62bb9afa"),
					Method:         "CREDIT_CARD",
					MethodType:     "CLOSED_DYNAMIC",
					Channel:        "MASTERCARD",
					Amount:         "101000",
					AmountPaid:     util.ValueToPtr("101000"),
					Status:         "SUCCESS",
					CreatedAt:      time.Date(2024, 11, 14, 22, 10, 34, 0, time.UTC),
					PaidAt:         util.ValueToPtr(time.Date(2024, 11, 14, 22, 12, 00, 0, time.UTC)),
					ExpiredAt:      util.ValueToPtr(time.Date(2024, 11, 15, 22, 10, 34, 0, time.UTC)),
					CustomerId:     "ec796f22-fd96-4b33-92c0-4ba1f93c19e8",
					CardType:       "DEBIT",
					CardIssuerBank: "BRI_S2I",
					CardNumber:     "0008",
				},
			},
			wantExcelContent: map[string]string{
				"A1": "Payment History",
				"A2": "01 November 2024 - 17 November 2024",
				"A3": "",
				"A5": "15 Nov 2024, 05:10 AM",
				"B5": "Cards",
				"C5": "Debit",
				"D5": "Rp 101,000",
				"E5": "Success",
				"F5": "15 Nov 2024, 05:12 AM",
				"G5": "MASTERCARD",
				"H5": "16 Nov 2024, 05:10 AM",
				"I5": "Rp 101,000",
				"J5": "ec796f22-fd96-4b33-92c0-4ba1f93c19e8",
				"K5": "1731622233KYDWX",
				"L5": "f378a8ba-ef6e-41eb-8081-7ccd62bb9ede",
				"M5": "f378a8ba-ef6e-41eb-8081-7ccd62bb9afa",
				"N5": "  -",
				"O5": "",
				"P5": "",
				"Q5": "",
				"R5": "",
				"S5": "",
				"T5": "",
				"U5": "BRI_S2I",
				"V5": "0008",
				"W5": "",
				"X5": "  -",
				"Y5": "  -",
				"Z5": "  -",
			},
		},
		{
			name: "Transaction with refund",
			transactions: []paymentModel.PaymentHistoryItem{
				{
					UUID:           "f378a8ba-ef6e-41eb-8081-7ccd62bb9ede",
					ReferenceID:    "1731622233KYDWX",
					Method:         "CREDIT_CARD",
					MethodType:     "CLOSED_DYNAMIC",
					Channel:        "MASTERCARD",
					Amount:         "101000",
					AmountPaid:     util.ValueToPtr("101000"),
					Status:         "SUCCESS",
					CreatedAt:      time.Date(2024, 11, 14, 22, 10, 34, 0, time.UTC),
					PaidAt:         util.ValueToPtr(time.Date(2024, 11, 14, 22, 12, 00, 0, time.UTC)),
					ExpiredAt:      util.ValueToPtr(time.Date(2024, 11, 15, 22, 10, 34, 0, time.UTC)),
					CustomerId:     "ec796f22-fd96-4b33-92c0-4ba1f93c19e8",
					CardType:       "DEBIT",
					CardIssuerBank: "BRI_S2I",
					CardNumber:     "0008",
					RefundDate:     util.ValueToPtr(time.Date(2024, 11, 16, 10, 30, 0, 0, time.UTC)),
					RefundAmount:   util.ValueToPtr("50000"),
					RefundStatus:   util.ValueToPtr("SUCCESS"),
				},
			},
			wantExcelContent: map[string]string{
				"A1": "Payment History",
				"A2": "01 November 2024 - 17 November 2024",
				"A3": "",
				"A5": "15 Nov 2024, 05:10 AM",
				"B5": "Cards",
				"C5": "Debit",
				"D5": "Rp 101,000",
				"E5": "Success",
				"F5": "15 Nov 2024, 05:12 AM",
				"G5": "MASTERCARD",
				"H5": "16 Nov 2024, 05:10 AM",
				"I5": "Rp 101,000",
				"J5": "ec796f22-fd96-4b33-92c0-4ba1f93c19e8",
				"K5": "1731622233KYDWX",
				"L5": "f378a8ba-ef6e-41eb-8081-7ccd62bb9ede",
				"M5": "",
				"N5": "  -",
				"O5": "",
				"P5": "",
				"Q5": "",
				"R5": "",
				"S5": "",
				"T5": "",
				"U5": "BRI_S2I",
				"V5": "0008",
				"W5": "",
				"X5": "16 Nov 2024, 05:30 PM",
				"Y5": "Rp 50,000",
				"Z5": "Success",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			assert.NoError(t, service.ExportToExcel(context.Background(), request, test.transactions, buf))

			f, err := excelize.OpenReader(buf)
			require.NoError(t, err)

			idx, err := f.GetSheetIndex(sheetName)
			require.NoError(t, err)
			require.Greater(t, idx, -1)

			maps.Copy(test.wantExcelContent, headers)

			for cell, want := range test.wantExcelContent {

				val, err := f.GetCellValue(sheetName, cell)
				assert.NoError(t, err)
				assert.Equal(t, want, val)
			}
		})
	}
}

func TestDefineDefaultSettlementConfig(t *testing.T) {

	fileConfig, err := os.CreateTemp(os.TempDir(), "settlement-config-*.yml")
	require.NoError(t, err)
	require.NotNil(t, fileConfig)

	defer func() {
		_ = fileConfig.Close()
		_ = os.Remove(fileConfig.Name())
	}()

	_, _ = fileConfig.WriteString(`
PAYMENT_SETTLEMENT:
  CREDIT_CARD:
    OTHER_CHANNEL:
      TYPE: "T+5"
    LOCAL_VISA:
      TYPE: "T+3"
  VIRTUAL_ACCOUNT:
    OTHER_CHANNEL:
      TYPE: "INSTANT"
    BCA:
      TYPE: "T+1"
  QRIS:
    OTHER_CHANNEL:
      TYPE: "T+1"
    BNC:
      TYPE: "T+2"
  EWALLET:
    OTHER_CHANNEL:
      TYPE: "T+2"
    DANA:
      TYPE: "T+4"
`)

	config, _, err := config.LoadConfig(fileConfig.Name(), fileConfig.Name())
	require.NoError(t, err)
	require.NotNil(t, config)

	payment := &PaymentService{
		config: config,
	}

	tests := []struct {
		name               string
		method             string
		channel            string
		wantSettlementTime *merchantModel.SettlementConfig
	}{
		{
			name:    "Local visa",
			method:  paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			channel: "LOCAL_VISA",
			wantSettlementTime: &merchantModel.SettlementConfig{
				Type: "T+3",
			},
		},
		{
			name:    "Foreign Mastercard",
			method:  paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
			channel: "FOREIGN_MASTERCARD",
			wantSettlementTime: &merchantModel.SettlementConfig{
				Type: "T+5",
			},
		},
		{
			name:    "VA BRI",
			method:  paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			channel: "BRI",
			wantSettlementTime: &merchantModel.SettlementConfig{
				Type: "INSTANT",
			},
		},
		{
			name:    "VA BCA",
			method:  paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			channel: "BCA",
			wantSettlementTime: &merchantModel.SettlementConfig{
				Type: "T+1",
			},
		},
		{
			name:    "QRIS BRI",
			method:  paymentConstant.PAYMENT_METHOD_QRIS,
			channel: "BRI",
			wantSettlementTime: &merchantModel.SettlementConfig{
				Type: "T+1",
			},
		},
		{
			name:    "QRIS BNC",
			method:  paymentConstant.PAYMENT_METHOD_QRIS,
			channel: "BNC",
			wantSettlementTime: &merchantModel.SettlementConfig{
				Type: "T+2",
			},
		},
		{
			name:    "EWallet DANA",
			method:  paymentConstant.PAYMENT_METHOD_EWALLET,
			channel: "DANA",
			wantSettlementTime: &merchantModel.SettlementConfig{
				Type: "T+4",
			},
		},
		{
			name:    "EWallet OVO",
			method:  paymentConstant.PAYMENT_METHOD_EWALLET,
			channel: "OVO",
			wantSettlementTime: &merchantModel.SettlementConfig{
				Type: "T+2",
			},
		},
		{
			name:    "Others",
			method:  "OTHERS",
			channel: "TESTING",
			wantSettlementTime: &merchantModel.SettlementConfig{
				Type: constant.SettlementTypeInstant,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			settlementTime := payment.defineDefaultSettlementConfig(test.method, test.channel)

			assert.Equal(t, test.wantSettlementTime, settlementTime)
		})
	}
}
