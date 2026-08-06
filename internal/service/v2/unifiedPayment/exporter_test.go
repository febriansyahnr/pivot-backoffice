package unifiedPaymentService

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestUnifiedPaymentServiceExportToExcelChargeHistories(t *testing.T) {
	tests := []struct {
		name           string
		request        *unifiedPaymentModel.FilterChargeRequest
		charges        []unifiedPaymentModel.ChargeResponse
		expectedValues map[string]string
		expectError    bool
	}{
		{
			name: "successful export with card payment",
			request: &unifiedPaymentModel.FilterChargeRequest{
				Status:         "SUCCESS",
				UUID:           "charge-123",
				StartCreatedAt: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
				EndCreatedAt:   time.Date(2025, 3, 17, 23, 59, 59, 0, time.UTC),
			},
			charges: []unifiedPaymentModel.ChargeResponse{
				{
					ID:                              "charge-123",
					PaymentSessionID:                "payment-session-123",
					PaymentSessionClientReferenceID: "ref-123",
					Status:                          "success",
					CreatedAt:                       time.Date(2025, 3, 16, 10, 30, 0, 0, time.UTC),
					Amount:                          unifiedPaymentModel.Amount{Value: 100000.00, Currency: "IDR"},
					AuthorizedAmount:                &unifiedPaymentModel.Amount{Value: 100000.00, Currency: "IDR"},
					CapturedAmount:                  &unifiedPaymentModel.Amount{Value: 100000.00, Currency: "IDR"},
					StatementDescriptor:             "Test Merchant",
					MerchantName:                    "Test Merchant Ltd",
					ExpiredAt:                       &[]time.Time{time.Date(2025, 3, 16, 11, 30, 0, 0, time.UTC)}[0],
					PaidAt:                          &[]time.Time{time.Date(2025, 3, 16, 10, 45, 0, 0, time.UTC)}[0],
					ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
						Card: &unifiedPaymentModel.ChargePaymentMethodDetailCard{
							First6: "411111",
							Last4:  "1111",
							BinInformations: unifiedPaymentModel.ChargePaymentMethodDetailBinInformation{
								Brand:       "VISA",
								IssuingBank: "Bank ABC",
							},
							AuthorizationResult: &unifiedPaymentModel.ChargePaymentMethodDetailCardAuthorizationResult{
								AcquirerReferenceNumber: "RRN123456",
								IssuerAuthorizationCode: "AUTH123",
							},
						},
					},
				},
			},
			expectedValues: map[string]string{
				"A2": "Charge History",
				"A3": "15 March 2025 - 18 March 2025",
				"A4": "Charge Status:",
				"B4": "Success",
				"A5": "Charge Id:",
				"B5": "charge-123",
				"A7": "Created Date",
				"B7": "Method",
				"C7": "Channel",
				"D7": "Payment ID",
				"E7": "Reference ID",
				"F7": "Amount",
				"G7": "Charge ID",
				"H7": "Status",
				"I7": "Failure Reason",
				"J7": "Payment Date",
				"K7": "Bank Reference ID",
				"L7": "Expiry Time",
				"M7": "Total Authorized Amount",
				"N7": "Total Captured Amount",
				"O7": "Statement Descriptor",
				"P7": "Network Response Code",
				"Q7": "Acquiring Bank",
				"R7": "Bank Merchant ID (MID)",
				"S7": "Issuer Bank",
				"T7": "Card Number",
				"U7": "VA Name",
				"V7": "VA Number",
				"W7": "Merchant Name",
				"A8": "16 Mar 2025, 05:30 PM",
				"B8": "Card",
				"C8": "VISA",
				"D8": "payment-session-123",
				"E8": "ref-123",
				"F8": "Rp 100,000",
				"G8": "charge-123",
				"H8": "Success",
				"I8": "",
				"J8": "2025-03-16 17:45:00",
				"K8": "RRN123456",
				"L8": "2025-03-16 18:30:00",
				"M8": "Rp 100,000",
				"N8": "Rp 100,000",
				"O8": "Test Merchant",
				"P8": "AUTH123",
				"Q8": "",
				"R8": "",
				"S8": "Bank ABC",
				"T8": "411111****1111",
				"U8": "",
				"V8": "",
				"W8": "Test Merchant Ltd",
			},
			expectError: false,
		},
		{
			name: "successful export with virtual account payment",
			request: &unifiedPaymentModel.FilterChargeRequest{
				StartCreatedAt: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
				EndCreatedAt:   time.Date(2025, 3, 17, 23, 59, 59, 0, time.UTC),
			},
			charges: []unifiedPaymentModel.ChargeResponse{
				{
					ID:                              "va-charge-123",
					PaymentSessionID:                "va-payment-session-123",
					PaymentSessionClientReferenceID: "va-ref-123",
					Status:                          "pending",
					CreatedAt:                       time.Date(2025, 3, 16, 14, 20, 0, 0, time.UTC),
					Amount:                          unifiedPaymentModel.Amount{Value: 250000.00, Currency: "IDR"},
					StatementDescriptor:             "VA Test Merchant",
					MerchantName:                    "VA Test Merchant Ltd",
					ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
						VirtualAccount: &unifiedPaymentModel.ChargePaymentMethodDetailVirtualAccount{
							Channel:              "BCA",
							VirtualAccountName:   "Test User",
							VirtualAccountNumber: "123456789012345",
							ExpiryAt:             time.Date(2025, 3, 17, 14, 20, 0, 0, time.UTC),
						},
					},
				},
			},
			expectedValues: map[string]string{
				"A2": "Charge History",
				"A3": "15 March 2025 - 18 March 2025",
				"A5": "Created Date",
				"B5": "Method",
				"C5": "Channel",
				"A6": "16 Mar 2025, 09:20 PM",
				"B6": "Virtual Account",
				"C6": "BCA",
				"D6": "va-payment-session-123",
				"E6": "va-ref-123",
				"F6": "Rp 250,000",
				"G6": "va-charge-123",
				"H6": "Pending",
				"I6": "",
				"J6": "",
				"K6": "",
				"L6": "2025-03-17 21:20:00",
				"M6": "Rp 0",
				"N6": "Rp 0",
				"O6": "VA Test Merchant",
				"U6": "Test User",
				"V6": "123456789012345",
				"W6": "VA Test Merchant Ltd",
			},
			expectError: false,
		},
		{
			name: "successful export with ewallet payment",
			request: &unifiedPaymentModel.FilterChargeRequest{
				StartCreatedAt: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
				EndCreatedAt:   time.Date(2025, 3, 17, 23, 59, 59, 0, time.UTC),
			},
			charges: []unifiedPaymentModel.ChargeResponse{
				{
					ID:                              "ew-charge-123",
					PaymentSessionID:                "ew-payment-session-123",
					PaymentSessionClientReferenceID: "ew-ref-123",
					Status:                          "success",
					CreatedAt:                       time.Date(2025, 3, 16, 16, 45, 0, 0, time.UTC),
					Amount:                          unifiedPaymentModel.Amount{Value: 75000.00, Currency: "IDR"},
					StatementDescriptor:             "Ewallet Test Merchant",
					MerchantName:                    "Ewallet Test Merchant Ltd",
					PaidAt:                          &[]time.Time{time.Date(2025, 3, 16, 16, 50, 0, 0, time.UTC)}[0],
					ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
						Ewallet: &unifiedPaymentModel.ChargePaymentMethodDetailEwallet{
							Channel:     "GOPAY",
							ReferenceNo: "GOPAY-REF-123456",
						},
					},
				},
			},
			expectedValues: map[string]string{
				"A2": "Charge History",
				"A3": "15 March 2025 - 18 March 2025",
				"A6": "16 Mar 2025, 11:45 PM",
				"B6": "E-WALLET",
				"C6": "GOPAY",
				"D6": "ew-payment-session-123",
				"E6": "ew-ref-123",
				"F6": "Rp 75,000",
				"G6": "ew-charge-123",
				"H6": "Success",
				"I6": "",
				"J6": "2025-03-16 23:50:00",
				"K6": "GOPAY-REF-123456",
				"L6": "",
				"O6": "Ewallet Test Merchant",
				"W6": "Ewallet Test Merchant Ltd",
			},
			expectError: false,
		},
		{
			name: "successful export with QRIS payment",
			request: &unifiedPaymentModel.FilterChargeRequest{
				StartCreatedAt: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
				EndCreatedAt:   time.Date(2025, 3, 17, 23, 59, 59, 0, time.UTC),
			},
			charges: []unifiedPaymentModel.ChargeResponse{
				{
					ID:                              "qris-charge-123",
					PaymentSessionID:                "qris-payment-session-123",
					PaymentSessionClientReferenceID: "qris-ref-123",
					Status:                          "success",
					CreatedAt:                       time.Date(2025, 3, 16, 18, 15, 0, 0, time.UTC),
					Amount:                          unifiedPaymentModel.Amount{Value: 50000.00, Currency: "IDR"},
					StatementDescriptor:             "QRIS Merchant",
					MerchantName:                    "",
					PaidAt:                          &[]time.Time{time.Date(2025, 3, 16, 18, 20, 0, 0, time.UTC)}[0],
					ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{
						Qr: &unifiedPaymentModel.ChargePaymentMethodDetailQr{
							Acquirer:                 "NOBU",
							RetrievalReferenceNumber: "QRIS-RRN-789",
							ExpiryAt:                 time.Date(2025, 3, 16, 19, 15, 0, 0, time.UTC),
						},
					},
				},
			},
			expectedValues: map[string]string{
				"A2": "Charge History",
				"A3": "15 March 2025 - 18 March 2025",
				"A6": "17 Mar 2025, 01:15 AM",
				"B6": "QRIS",
				"C6": "NOBU",
				"D6": "qris-payment-session-123",
				"E6": "qris-ref-123",
				"F6": "Rp 50,000",
				"G6": "qris-charge-123",
				"H6": "Success",
				"I6": "",
				"J6": "2025-03-17 01:20:00",
				"K6": "QRIS-RRN-789",
				"L6": "2025-03-17 02:15:00",
				"O6": "QRIS Merchant",
				"W6": "QRIS Merchant", // Uses statement descriptor when merchant name is empty
			},
			expectError: false,
		},
		{
			name: "successful export with empty charge data",
			request: &unifiedPaymentModel.FilterChargeRequest{
				StartCreatedAt: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
				EndCreatedAt:   time.Date(2025, 3, 17, 23, 59, 59, 0, time.UTC),
			},
			charges: []unifiedPaymentModel.ChargeResponse{},
			expectedValues: map[string]string{
				"A2": "Charge History",
				"A3": "15 March 2025 - 18 March 2025",
			},
			expectError: false,
		},
		{
			name: "successful export with charge without payment method details",
			request: &unifiedPaymentModel.FilterChargeRequest{
				StartCreatedAt: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
				EndCreatedAt:   time.Date(2025, 3, 17, 23, 59, 59, 0, time.UTC),
			},
			charges: []unifiedPaymentModel.ChargeResponse{
				{
					ID:                              "charge-no-method",
					PaymentSessionID:                "payment-session-no-method",
					PaymentSessionClientReferenceID: "ref-no-method",
					Status:                          "pending",
					CreatedAt:                       time.Date(2025, 3, 16, 10, 30, 0, 0, time.UTC),
					Amount:                          unifiedPaymentModel.Amount{Value: 50000.00, Currency: "IDR"},
					StatementDescriptor:             "No Method Merchant",
					MerchantName:                    "No Method Merchant Ltd",
					ChargePaymentMethodDetails:      &unifiedPaymentModel.ChargePaymentMethodDetails{},
				},
			},
			expectedValues: map[string]string{
				"A2": "Charge History",
				"A3": "15 March 2025 - 18 March 2025",
				"A6": "16 Mar 2025, 05:30 PM",
				"B6": "", // No payment method
				"C6": "", // No channel
				"D6": "payment-session-no-method",
				"E6": "ref-no-method",
				"F6": "Rp 50,000",
				"G6": "charge-no-method",
				"H6": "Pending",
				"I6": "", // No failure reason
				"J6": "", // No payment date
				"K6": "", // No bank reference
				"L6": "", // No expiry time
				"O6": "No Method Merchant",
				"W6": "No Method Merchant Ltd",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &UnifiedPaymentService{}

			buff, err := service.ExportToExcelChargeHistories(context.Background(), tt.request, tt.charges)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, buff)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, buff)

			f, err := excelize.OpenReader(buff)
			require.NoError(t, err)
			defer func() { require.NoError(t, f.Close()) }()

			_, err = f.GetSheetIndex(constant.DefaultSheetName)
			require.NoError(t, err)

			for cell, want := range tt.expectedValues {
				result, err := f.GetCellValue(constant.DefaultSheetName, cell)
				assert.NoError(t, err)
				assert.Equal(t, want, result, "cell %s should have value %s but got %s", cell, want, result)
			}
		})
	}
}
