package paymentModel_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	snapCoreModelQr "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qr"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"strings"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestModelPaymentResponse(pt *testing.T) {
	merchantID := "123456"
	referenceID := "REF-0001"
	customerName := "John Wick"
	response := PaymentResponse{
		TransactionDate: &time.Time{},
	}

	payment := &PaymentDTO{
		UUID:        "09ab42a5-3c75-40bf-888c-2b20cae0fd3f",
		Status:      "OK",
		ReferenceID: &referenceID,
		CreatedAt:   time.Time{},
	}
	snapResp := &snapCoreModel.CreateVirtualAccountResponseData{
		Acquirer:         "ACCQUIRER",
		VirtualAccountNo: "VA-000001",
		AccountName:      customerName,
		TotalAmount:      snapCoreModel.Amount{Value: "21525"},
		MinAmount:        snapCoreModel.Amount{Value: "10000", Currency: "IDR"},
		MaxAmount:        snapCoreModel.Amount{Value: "100000", Currency: "IDR"},
	}
	paymentReq := &PaymentRequest{
		PaymentMethod:  "Transfer",
		TotalAmount:    Amount{Currency: "IDR"},
		VirtualAccount: &PaymentMetadataVirtualAccount{VirtualAccountTrxType: "VA-TYPE"},
	}
	firstName, lastName := customerModel.FullNameToFirstNameAndLastName(customerName)
	customer := &customerModel.Customer{
		UUID:        uuid.NewString(),
		FirstName:   firstName,
		LastName:    lastName,
		Email:       "email@example.id",
		PhoneNumber: "0881",
	}

	pt.Run("RUN:To payment response", func(t *testing.T) {
		response.ToPaymentResponse(
			payment, snapResp, paymentReq, customer, merchantID,
		)

		assert.Equal(t, PaymentResponse{
			UUID:          payment.UUID,
			MerchantID:    merchantID,
			Status:        payment.Status,
			PaymentMethod: paymentReq.PaymentMethod,
			ReferenceID:   referenceID,
			Customer: &PaymentRequestCustomer{
				CustomerID: customer.UUID,
				Name:       customerModel.FirstNameAndLastNameToFullName(customer.FirstName, customer.LastName),
				Email:      customer.Email,
				Phone:      customer.PhoneNumber,
			},
			TotalAmount: &Amount{
				Value:    decimal.RequireFromString(snapResp.TotalAmount.Value),
				Currency: paymentReq.TotalAmount.Currency,
			},
			TransactionDate: &time.Time{},
		}, response)
	})

	pt.Run("RUN:To virtual account response", func(t *testing.T) {
		response.ToVirtualAccountResponse(paymentReq, snapResp)

		assert.Equal(t, &PaymentVirtualAccountResponse{
			Issuer:                snapResp.Acquirer,
			VirtualAccountTrxType: paymentReq.VirtualAccount.VirtualAccountTrxType,
			VirtualAccountNumber:  snapResp.VirtualAccountNo,
			VirtualAccountName:    snapResp.AccountName,
			ExpiredDate:           &snapResp.ExpiredAt,
			MinAmount: &Amount{
				Value:    decimal.RequireFromString(snapResp.MinAmount.Value),
				Currency: snapResp.MinAmount.Currency,
			},
			MaxAmount: &Amount{
				Value:    decimal.RequireFromString(snapResp.MaxAmount.Value),
				Currency: snapResp.MaxAmount.Currency,
			},
		}, response.VirtualAccount)
	})
}

func TestModelPaymentItemRequest(t *testing.T) {
	req := PaymentItemRequest{
		ItemID:      "123456",
		Name:        "Item Name",
		Description: "Description",
		Qty:         1,
		Amount:      Amount{Value: decimal.New(12_000, 0)},
	}

	assert.Equal(t, snapCoreModel.BillDetail{
		BillName: req.Name,
		BillDescription: snapCoreModel.Description{
			English:   req.Description,
			Indonesia: req.Description,
		},
		BillAmount: snapCoreModel.Amount{
			Value:    req.Amount.Value.Mul(decimal.NewFromInt(int64(req.Qty))).StringFixed(2),
			Currency: req.Amount.Currency,
		},
		AdditionalInfo: req.Metadata,
	}, req.ToSnapCoreBillDetail())
}

func TestModelPaymentMethod(t *testing.T) {
	pm := &PaymentMethod{
		UUID:     uuid.NewString(),
		Type:     "TYPE",
		Category: "CATEGORY",
		Name:     "NAME",
		Acquirer: "ACQUIRER",
	}
	assert.Equal(t, &PaymentMethodResponse{
		UUID:     pm.UUID,
		Type:     pm.Type,
		Category: pm.Category,
		Name:     pm.Name,
		Acquirer: pm.Acquirer,
	}, pm.ToResponse())
}

func TestToSnapVAPaymentResponse(t *testing.T) {
	// Setup common test data
	customerName := "John Wick"
	now := time.Now()
	expiredDate := now.Add(24 * time.Hour)
	zeroTime := time.Time{}

	testCases := []struct {
		name                string
		response            *PaymentResponse
		expectedVaStatus    string
		expectedExpiredDate string
	}{
		{
			name: "With valid ExpiredDate and Open VA type",
			response: &PaymentResponse{
				UUID:        uuid.NewString(),
				MerchantID:  "123456",
				ReferenceID: "REF-0001",
				Customer: &PaymentRequestCustomer{
					CustomerID: uuid.NewString(),
					Name:       customerName,
					Email:      "john@example.com",
					Phone:      "123456789",
				},
				Status: "ACTIVE",
				PaidAmount: &commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				TotalAmount: &Amount{
					Currency: "IDR",
					Value:    decimal.New(10000, 0),
				},
				PaymentMethod: "VIRTUAL_ACCOUNT",
				VirtualAccount: &PaymentVirtualAccountResponse{
					Issuer:                "ISSUER",
					VirtualAccountTrxType: constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC,
					VirtualAccountNumber:  "VA-000001",
					VirtualAccountName:    customerName,
					ExpiredDate:           &expiredDate,
					MinAmount: &Amount{
						Currency: "IDR",
						Value:    decimal.New(5000, 0),
					},
					MaxAmount: &Amount{
						Currency: "IDR",
						Value:    decimal.New(20000, 0),
					},
				},
				TransactionDate: &now,
			},
			expectedVaStatus:    "ACTIVE",
			expectedExpiredDate: util.SnapCompatible(expiredDate),
		},
		{
			name: "With Closed Dynamic VA type",
			response: &PaymentResponse{
				UUID:        uuid.NewString(),
				MerchantID:  "123456",
				ReferenceID: "REF-0001",
				Customer: &PaymentRequestCustomer{
					CustomerID: uuid.NewString(),
					Name:       customerName,
				},
				Status: "ACTIVE",
				PaidAmount: &commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				TotalAmount: &Amount{
					Currency: "IDR",
					Value:    decimal.New(10000, 0),
				},
				PaymentMethod: "VIRTUAL_ACCOUNT",
				VirtualAccount: &PaymentVirtualAccountResponse{
					Issuer:                "ISSUER",
					VirtualAccountTrxType: constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
					VirtualAccountNumber:  "VA-000001",
					VirtualAccountName:    customerName,
					ExpiredDate:           &expiredDate,
				},
				TransactionDate: &now,
			},
			expectedVaStatus:    "INACTIVE",
			expectedExpiredDate: util.SnapCompatible(expiredDate),
		},
		{
			name: "With nil ExpiredDate",
			response: &PaymentResponse{
				UUID:        uuid.NewString(),
				MerchantID:  "123456",
				ReferenceID: "REF-0001",
				Customer: &PaymentRequestCustomer{
					CustomerID: uuid.NewString(),
					Name:       customerName,
				},
				Status: "ACTIVE",
				PaidAmount: &commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				TotalAmount: &Amount{
					Currency: "IDR",
					Value:    decimal.New(10000, 0),
				},
				PaymentMethod: "VIRTUAL_ACCOUNT",
				VirtualAccount: &PaymentVirtualAccountResponse{
					Issuer:                "ISSUER",
					VirtualAccountTrxType: constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC,
					VirtualAccountNumber:  "VA-000001",
					VirtualAccountName:    customerName,
					ExpiredDate:           nil, // Nil ExpiredDate
				},
				TransactionDate: &now,
			},
			expectedVaStatus:    "ACTIVE",
			expectedExpiredDate: "",
		},
		{
			name: "With zero ExpiredDate",
			response: &PaymentResponse{
				UUID:        uuid.NewString(),
				MerchantID:  "123456",
				ReferenceID: "REF-0001",
				Customer: &PaymentRequestCustomer{
					CustomerID: uuid.NewString(),
					Name:       customerName,
				},
				Status: "ACTIVE",
				PaidAmount: &commonModel.Amount{
					Currency: "IDR",
					Value:    "10000",
				},
				TotalAmount: &Amount{
					Currency: "IDR",
					Value:    decimal.New(10000, 0),
				},
				PaymentMethod: "VIRTUAL_ACCOUNT",
				VirtualAccount: &PaymentVirtualAccountResponse{
					Issuer:                "ISSUER",
					VirtualAccountTrxType: constantPayment.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC,
					VirtualAccountNumber:  "VA-000001",
					VirtualAccountName:    customerName,
					ExpiredDate:           &zeroTime, // Zero ExpiredDate
				},
				TransactionDate: &now,
			},
			expectedVaStatus:    "ACTIVE",
			expectedExpiredDate: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the method under test
			result := tc.response.ToSnapVAPaymentResponse()

			// Verify the results
			assert.Equal(t, tc.response.UUID, result.TrxId, "TrxId should match")
			assert.Equal(t, "", result.VirtualAccountEmail, "VirtualAccountEmail should be empty")
			assert.Equal(t, "", result.VirtualAccountPhone, "VirtualAccountPhone should be empty")
			assert.Equal(t, tc.response.VirtualAccount.VirtualAccountNumber, result.VirtualAccountNo, "VirtualAccountNo should match")
			assert.Equal(t, tc.response.VirtualAccount.VirtualAccountName, result.VirtualAccountName, "VirtualAccountName should match")

			// Check AdditionalInfo fields
			assert.NotNil(t, result.AdditionalInfo, "AdditionalInfo should not be nil")
			assert.Equal(t, tc.response.ReferenceID, result.AdditionalInfo.ReferenceId, "ReferenceId should match")
			assert.Equal(t, strings.ToUpper(tc.response.VirtualAccount.Issuer), result.AdditionalInfo.Issuer, "Issuer should match and be uppercase")
			assert.Equal(t, tc.response.VirtualAccount.VirtualAccountTrxType, result.AdditionalInfo.VirtualAccountTrxType, "VirtualAccountTrxType should match")

			// Check the conditional fields we're specifically testing
			assert.Equal(t, tc.expectedVaStatus, result.AdditionalInfo.VaStatus, "VaStatus should match expected value")
			assert.Equal(t, tc.expectedExpiredDate, result.AdditionalInfo.ExpiredDate, "ExpiredDate should match expected value")

			// Check that PaymentStatus is always SUCCESS
			assert.Equal(t, constantPayment.PAYMENT_STATUS_SUCCESS, result.AdditionalInfo.PaymentStatus, "PaymentStatus should be SUCCESS")
		})
	}
}

func TestToQrisResponse(t *testing.T) {
	referenceId := "REF-0001"
	now := time.Now()
	expiredAt := now.Add(time.Hour)

	// Test cases
	testCases := []struct {
		name                    string
		payment                 *PaymentDTO
		snapCoreResp            *snapCoreModelQr.GenerateQrMpmResponseData
		paymentRequest          *PaymentRequest
		expectedQrType          string
		expectedQrExpiredDate   string
		expectedTransactionDate string
	}{
		{
			name: "Static QR",
			payment: &PaymentDTO{
				ReferenceID: &referenceId,
				Status:      constantPayment.PAYMENT_STATUS_PENDING,
				UpdatedAt:   now,
				ExpiredAt:   &expiredAt,
			},
			snapCoreResp: &snapCoreModelQr.GenerateQrMpmResponseData{
				Amount: commonModel.Amount{
					Value:    "10000",
					Currency: "IDR",
				},
			},
			paymentRequest: &PaymentRequest{
				Qris: &PaymentMetadataQris{
					QrType: constant.QrTypeStatic,
				},
			},
			expectedQrType:          constant.QrTypeStatic,
			expectedQrExpiredDate:   "", // Static QR should have empty QrExpiredDate
			expectedTransactionDate: "", // Pending status should have empty TransactionDate
		},
		{
			name: "Dynamic QR with Pending Status",
			payment: &PaymentDTO{
				ReferenceID: &referenceId,
				Status:      constantPayment.PAYMENT_STATUS_PENDING,
				UpdatedAt:   now,
				ExpiredAt:   &expiredAt,
			},
			snapCoreResp: &snapCoreModelQr.GenerateQrMpmResponseData{
				Amount: commonModel.Amount{
					Value:    "10000",
					Currency: "IDR",
				},
			},
			paymentRequest: &PaymentRequest{
				Qris: &PaymentMetadataQris{
					QrType: constant.QrTypeDynamic,
				},
			},
			expectedQrType:          constant.QrTypeDynamic,
			expectedQrExpiredDate:   util.SnapCompatible(expiredAt), // Dynamic QR should have QrExpiredDate
			expectedTransactionDate: "",                             // Pending status should have empty TransactionDate
		},
		{
			name: "Dynamic QR with Success Status",
			payment: &PaymentDTO{
				ReferenceID: &referenceId,
				Status:      constantPayment.PAYMENT_STATUS_SUCCESS,
				UpdatedAt:   now,
				ExpiredAt:   &expiredAt,
			},
			snapCoreResp: &snapCoreModelQr.GenerateQrMpmResponseData{
				Amount: commonModel.Amount{
					Value:    "10000",
					Currency: "IDR",
				},
			},
			paymentRequest: &PaymentRequest{
				Qris: &PaymentMetadataQris{
					QrType: constant.QrTypeDynamic,
				},
			},
			expectedQrType:          constant.QrTypeDynamic,
			expectedQrExpiredDate:   util.SnapCompatible(expiredAt), // Dynamic QR should have QrExpiredDate
			expectedTransactionDate: util.SnapCompatible(now),       // Success status should have TransactionDate
		},
		{
			name: "Dynamic QR with nil ExpiredAt",
			payment: &PaymentDTO{
				ReferenceID: &referenceId,
				Status:      constantPayment.PAYMENT_STATUS_PENDING,
				UpdatedAt:   now,
				ExpiredAt:   nil, // Nil ExpiredAt
			},
			snapCoreResp: &snapCoreModelQr.GenerateQrMpmResponseData{
				Amount: commonModel.Amount{
					Value:    "10000",
					Currency: "IDR",
				},
			},
			paymentRequest: &PaymentRequest{
				Qris: &PaymentMetadataQris{
					QrType: constant.QrTypeDynamic,
				},
			},
			expectedQrType:          constant.QrTypeDynamic,
			expectedQrExpiredDate:   "", // Nil ExpiredAt should result in empty QrExpiredDate
			expectedTransactionDate: "", // Pending status should have empty TransactionDate
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new PaymentResponse
			response := &PaymentResponse{
				Qris: &PaymentQrisResponse{},
			}

			// Call the method under test
			response.ToQrisResponse(tc.payment, tc.snapCoreResp, tc.paymentRequest)

			// Verify the results
			assert.Equal(t, tc.expectedQrType, response.Qris.QrType, "QrType should match expected value")
			assert.Equal(t, tc.expectedQrExpiredDate, response.Qris.QrExpiredDate, "QrExpiredDate should match expected value")
			assert.Equal(t, tc.expectedTransactionDate, response.Qris.TransactionDate, "TransactionDate should match expected value")
		})
	}
}

func TestToQueryQrMpmStaticResponse(t *testing.T) {
	referenceId := "REF-0001"
	payment := &Payment{
		ReferenceID: &referenceId,
		Metadata:    &map[string]interface{}{},
		Status:      constantPayment.PAYMENT_STATUS_VOID,
	}
	snapCoreResp := &snapCoreModelQr.GenerateQrMpmResponseData{
		Amount: commonModel.Amount{
			Value:    "10000",
			Currency: "IDR",
		},
	}
	snapCoreQueryResp := &snapCoreModelQr.QueryQrMpmStaticResponseData{
		DetailData: []snapCoreModelQr.TransactionHistoryListResponseDetailData{{}},
	}

	resp := &PaymentResponse{}
	resp.ToQueryQrMpmStaticResponse(payment, snapCoreResp, snapCoreQueryResp)

	expectedResp := &PaymentResponse{
		Qris: &PaymentQrisResponse{
			QrType: constant.QrTypeStatic,
		},
	}

	assert.Equal(t, expectedResp.Qris.QrType, resp.Qris.QrType)
}

func TestLoadPaymentV2CustomerOrderInformation(t *testing.T) {
	tests := []struct {
		name             string
		payment          *Payment
		customer         *customerModel.Customer
		expectedOrder    *unifiedPaymentModel.PaymentOrder
		expectedCustomer *unifiedPaymentModel.CustomerInformation
	}{
		{
			name: "Both payment and customer provided",
			payment: &Payment{
				Metadata: &map[string]any{
					"orderInformation": map[string]interface{}{
						"OrderID": "order-123",
					},
				},
			},
			customer: &customerModel.Customer{
				UUID:             "cust-001",
				FirstName:        "Ali",
				LastName:         "Achmad",
				Email:            "ali@example.com",
				PhoneCountryCode: "+62",
				PhoneNumber:      "81234567890",
				Metadata: map[string]interface{}{
					"refundPreference": map[string]interface{}{
						"BankAccount": "123456789",
					},
					"paymentMethods": []interface{}{
						map[string]interface{}{
							"token": "sample-token",
						},
					},
				},
			},
			expectedOrder: &unifiedPaymentModel.PaymentOrder{},
			expectedCustomer: &unifiedPaymentModel.CustomerInformation{
				CustomerID: "cust-001",
				GivenName:  "Ali",
				Surname:    util.ValueToPtr("Achmad"),
				Email:      "ali@example.com",
				PhoneNumber: &unifiedPaymentModel.UnifiedPaymentPhoneNumber{
					CountryCode: "+62",
					Number:      "81234567890",
				},
				RefundPreference: &unifiedPaymentModel.UnifiedPaymentRefundPreference{
					Method: "TRANSFER_ONLY",
					TransferDestination: &unifiedPaymentModel.RefundTransferDestination{
						ChannelCode: "014",
						ChannelInformation: &unifiedPaymentModel.RefundChannelInformation{
							AccountNumber: "01234",
							AccountName:   "TEST",
						},
					},
				},
				StoredPaymentMethods: []*unifiedPaymentModel.CustomerPaymentMethod{
					{
						Token: "sample-token",
					},
				},
			},
		},
		{
			name: "Only payment provided",
			payment: &Payment{
				Metadata: &map[string]any{
					"orderInformation": map[string]interface{}{
						"billingInfo": map[string]interface{}{
							"givenName": "john",
						},
					},
				},
			},
			customer: nil,
			expectedOrder: &unifiedPaymentModel.PaymentOrder{BillingInformation: &unifiedPaymentModel.BillingInformation{
				GivenName: "john",
			}},
			expectedCustomer: nil,
		},
		{
			name:             "Only customer provided",
			payment:          nil,
			customer:         &customerModel.Customer{UUID: "cust-002", Metadata: map[string]interface{}{}},
			expectedOrder:    nil,
			expectedCustomer: &unifiedPaymentModel.CustomerInformation{CustomerID: "cust-002"},
		},
		{
			name:             "Both nil",
			payment:          nil,
			customer:         nil,
			expectedOrder:    nil,
			expectedCustomer: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp PaymentHistoryDetailResponse
			resp.LoadPaymentV2CustomerOrderInformation(tt.payment, tt.customer)

			if !reflect.DeepEqual(resp.OrderInfo, tt.expectedOrder) {
				t.Errorf("OrderInfo mismatch:\nGot:      %+v\nExpected: %+v", resp.OrderInfo, tt.expectedOrder)
			}

			if tt.expectedCustomer != nil {
				if resp.CustomerInfo == nil {
					t.Errorf("Expected CustomerInfo, got nil")
					return
				}
				if resp.CustomerInfo.CustomerID != tt.expectedCustomer.CustomerID {
					t.Errorf("CustomerID mismatch: got %v, want %v", resp.CustomerInfo.CustomerID, tt.expectedCustomer.CustomerID)
				}
			} else if resp.CustomerInfo != nil {
				t.Errorf("Expected nil CustomerInfo, got %+v", resp.CustomerInfo)
			}
		})
	}
}

func TestUnifiedPaymentResponse_ToSnapPayment(t *testing.T) {
	testCases := []struct {
		name                 string
		unifiedResponse      *UnifiedPaymentResponse
		expectedType         string
		expectedResultType   interface{}
		shouldReturnOriginal bool
	}{
		{
			name: "SUCCESS: VA with PaymentMethod set and TypeDetail filled",
			unifiedResponse: &UnifiedPaymentResponse{
				Status:        constant.StatusSuccess,
				PaymentMethod: constant.ChannelVirtualAccount,
				TypeDetail: PaymentTypeDetail{
					VirtualAccountNumber: func() *string { s := "1234567890"; return &s }(),
					VirtualAccountName:   func() *string { s := "Test Account"; return &s }(),
				},
			},
			expectedType:       "VA",
			expectedResultType: SnapVACallbackResponse{},
		},
		{
			name: "SUCCESS: QRIS with PaymentMethod set and QrContent",
			unifiedResponse: &UnifiedPaymentResponse{
				Status:        constant.StatusPending,
				PaymentMethod: constant.ChannelQris,
				TypeDetail: PaymentTypeDetail{
					QrContent: func() *string { s := "qr-content-data"; return &s }(),
				},
			},
			expectedType:       "QRIS",
			expectedResultType: SnapQrisNotificationCallbackResponse{},
		},
		{
			name: "SUCCESS: VA fallback detection without PaymentMethod",
			unifiedResponse: &UnifiedPaymentResponse{
				Status: constant.StatusFailed,
				TypeDetail: PaymentTypeDetail{
					VirtualAccountNumber: func() *string { s := "9876543210"; return &s }(),
					VirtualAccountName:   func() *string { s := "Fallback Account"; return &s }(),
				},
			},
			expectedType:       "VA",
			expectedResultType: SnapVACallbackResponse{},
		},
		{
			name: "SUCCESS: QRIS fallback detection with PaymentMethod but no QrContent",
			unifiedResponse: &UnifiedPaymentResponse{
				Status:        constant.StatusActive,
				PaymentMethod: constant.ChannelQris,
				TypeDetail:    PaymentTypeDetail{},
			},
			expectedType:       "QRIS",
			expectedResultType: SnapQrisNotificationCallbackResponse{},
		},
		{
			name: "SUCCESS: QRIS fallback detection with QrContent but no PaymentMethod",
			unifiedResponse: &UnifiedPaymentResponse{
				Status: constant.StatusPending,
				TypeDetail: PaymentTypeDetail{
					QrContent: func() *string { s := "fallback-qr-content"; return &s }(),
				},
			},
			expectedType:       "QRIS",
			expectedResultType: SnapQrisNotificationCallbackResponse{},
		},
		{
			name: "SUCCESS: Return original when no conversion criteria met",
			unifiedResponse: &UnifiedPaymentResponse{
				Status:        "UNKNOWN",
				PaymentMethod: "OTHER_METHOD",
				TypeDetail:    PaymentTypeDetail{},
			},
			expectedType:         "original",
			shouldReturnOriginal: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.unifiedResponse.ToSnapPayment()

			if tc.shouldReturnOriginal {
				assert.Equal(t, tc.unifiedResponse, result, "Should return original UnifiedPaymentResponse when no conversion criteria met")
			} else {
				assert.IsType(t, tc.expectedResultType, result, "Should return correct type")

				switch tc.expectedType {
				case "VA":
					vaResult := result.(SnapVACallbackResponse)
					assert.NotEmpty(t, vaResult.ResponseCode, "VA response should have ResponseCode")
					assert.NotEmpty(t, vaResult.ResponseMessage, "VA response should have ResponseMessage")
					assert.NotNil(t, vaResult.VirtualAccountData, "VA response should have VirtualAccountData")

					// Verify status mapping
					switch tc.unifiedResponse.Status {
					case constant.StatusSuccess, "COMPLETED":
						assert.Equal(t, "2005200", vaResult.ResponseCode)
						assert.Equal(t, "Successful", vaResult.ResponseMessage)
					case constant.StatusFailed:
						assert.Equal(t, "4015200", vaResult.ResponseCode)
						assert.Equal(t, "Transaction Failed", vaResult.ResponseMessage)
					case constant.StatusPending:
						assert.Equal(t, "2005201", vaResult.ResponseCode)
						assert.Equal(t, "Transaction Pending", vaResult.ResponseMessage)
					}

				case "QRIS":
					qrisResult := result.(SnapQrisNotificationCallbackResponse)
					assert.NotEmpty(t, qrisResult.ResponseCode, "QRIS response should have ResponseCode")
					assert.NotEmpty(t, qrisResult.ResponseMessage, "QRIS response should have ResponseMessage")
					assert.NotNil(t, qrisResult.AdditionalInfo, "QRIS response should have AdditionalInfo")
					assert.Equal(t, constant.StatusActive, qrisResult.AdditionalInfo.QrStatus)
					assert.Equal(t, tc.unifiedResponse.Status, qrisResult.AdditionalInfo.PaymentStatus)

					// Verify status mapping
					switch tc.unifiedResponse.Status {
					case constant.StatusSuccess, "COMPLETED":
						assert.Equal(t, "2005200", qrisResult.ResponseCode)
						assert.Equal(t, "Successful", qrisResult.ResponseMessage)
					case constant.StatusFailed:
						assert.Equal(t, "4015200", qrisResult.ResponseCode)
						assert.Equal(t, "Transaction Failed", qrisResult.ResponseMessage)
					case constant.StatusPending:
						assert.Equal(t, "2005201", qrisResult.ResponseCode)
						assert.Equal(t, "Transaction Pending", qrisResult.ResponseMessage)
					}
				}
			}
		})
	}
}
