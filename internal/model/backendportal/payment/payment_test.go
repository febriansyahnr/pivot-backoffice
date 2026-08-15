package paymentModel

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pdk/v2/encrypt"
	"github.com/paper-indonesia/pdk/v2/gcp"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/creditcard"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/paymentCapture"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/proto/messages/callback"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/unifiedPayment"
	gcpMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/gcp"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPaymentPaymentFromDTO(t *testing.T) {
	amount := decimal.NewFromFloat(1000)
	totalAmount := decimal.NewFromFloat(1000)
	now := time.Now()
	newUuid := uuid.NewString()
	referenceId := "reference-id"
	processorReferenceNumber := "processor-reference-number"
	fee := decimal.NewFromFloat(1000)
	discount := decimal.NewFromFloat(1000)
	metadata := "{\"testing\":\"testing\"}"
	metadataMap := map[string]any{"testing": "testing"}

	paymentMethod := PaymentMethod{
		UUID:      newUuid,
		Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
		Name:      "VA Permata",
		Acquirer:  constant.BANK_ACQUIRER_PERMATA,
		CreatedAt: now,
		UpdatedAt: now,
	}

	payment := &Payment{
		UUID:                     "uuid-uuid-uuid",
		ReferenceID:              &referenceId,
		MerchantID:               "merchant-id",
		CustomerID:               "customer-id",
		PaymentMethodID:          "payment-method-id",
		ProcessorReferenceNumber: &processorReferenceNumber,
		Currency:                 "IDR",
		Amount:                   amount,
		Fee:                      &fee,
		Discount:                 &discount,
		TotalAmount:              totalAmount,
		Status:                   "pending",
		Metadata:                 nil,
		CreatedAt:                now,
		UpdatedAt:                now,
		PaymentMethod:            paymentMethod,
	}

	paymentDTO := &PaymentDTO{
		UUID:                     newUuid,
		ReferenceID:              &referenceId,
		MerchantID:               "merchant-id",
		CustomerID:               "customer-id",
		PaymentMethodID:          "payment-method-id",
		ProcessorReferenceNumber: &processorReferenceNumber,
		Currency:                 "IDR",
		Amount:                   amount,
		Fee:                      &fee,
		Discount:                 &discount,
		TotalAmount:              totalAmount,
		Status:                   "pending",
		Metadata:                 nil,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	testCases := []struct {
		Name     string
		Input    *Payment
		Param    *PaymentDTO
		Expected *Payment
	}{
		{
			Name: "it should return payment with metadata",
			Input: &Payment{
				UUID:                     newUuid,
				ReferenceID:              &referenceId,
				MerchantID:               "merchant-id",
				CustomerID:               "customer-id",
				PaymentMethodID:          "payment-method-id",
				ProcessorReferenceNumber: &processorReferenceNumber,
				Currency:                 "IDR",
				Amount:                   amount,
				Fee:                      &fee,
				Discount:                 &discount,
				TotalAmount:              totalAmount,
				Status:                   "pending",
				Metadata:                 &metadataMap,
				CreatedAt:                now,
				UpdatedAt:                now,
				PaymentMethod:            paymentMethod,
			},
			Param: &PaymentDTO{
				UUID:                     newUuid,
				ReferenceID:              &referenceId,
				MerchantID:               "merchant-id",
				CustomerID:               "customer-id",
				PaymentMethodID:          "payment-method-id",
				ProcessorReferenceNumber: &processorReferenceNumber,
				Currency:                 "IDR",
				Amount:                   amount,
				Fee:                      &fee,
				Discount:                 &discount,
				TotalAmount:              totalAmount,
				Status:                   "pending",
				Metadata:                 &metadata,
				CreatedAt:                now,
				UpdatedAt:                now,
			},
			Expected: &Payment{
				UUID:                     newUuid,
				ReferenceID:              &referenceId,
				MerchantID:               "merchant-id",
				CustomerID:               "customer-id",
				PaymentMethodID:          "payment-method-id",
				ProcessorReferenceNumber: &processorReferenceNumber,
				Currency:                 "IDR",
				Amount:                   amount,
				Fee:                      &fee,
				Discount:                 &discount,
				TotalAmount:              totalAmount,
				Status:                   "pending",
				Metadata:                 &metadataMap,
				CreatedAt:                now,
				UpdatedAt:                now,
				PaymentMethod:            paymentMethod,
			},
		},
		{
			Name:     "it should return payment with nil metadata",
			Input:    payment,
			Param:    paymentDTO,
			Expected: payment,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Expected.PaymentFromDTO(tc.Param)
			require.Equal(t, tc.Expected, tc.Input)
		})
	}
}

func TestPaymentToDTO(t *testing.T) {
	amount := decimal.NewFromFloat(1000)
	totalAmount := decimal.NewFromFloat(1000)
	now := time.Now()
	newUuid := uuid.NewString()
	fee := decimal.NewFromFloat(1000)
	discount := decimal.NewFromFloat(1000)
	metadata := "{\"testing\":\"testing\"}"
	metadataMap := map[string]any{"testing": "testing"}

	paymentMethod := PaymentMethod{
		UUID:      newUuid,
		Type:      paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
		Category:  paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
		Name:      "VA Permata",
		Acquirer:  constant.BANK_ACQUIRER_PERMATA,
		CreatedAt: now,
		UpdatedAt: now,
	}

	payment := &Payment{
		UUID:            newUuid,
		MerchantID:      "merchant-id",
		CustomerID:      "customer-id",
		PaymentMethodID: "payment-method-id",
		Currency:        "IDR",
		Amount:          amount,
		Fee:             &fee,
		Discount:        &discount,
		TotalAmount:     totalAmount,
		Status:          "pending",
		Metadata:        &metadataMap,
		CreatedAt:       now,
		UpdatedAt:       now,
		PaymentMethod:   paymentMethod,
	}

	paymentDTO := &PaymentDTO{
		UUID:            newUuid,
		MerchantID:      "merchant-id",
		CustomerID:      "customer-id",
		PaymentMethodID: "payment-method-id",
		Currency:        "IDR",
		Amount:          amount,
		Fee:             &fee,
		Discount:        &discount,
		TotalAmount:     totalAmount,
		Status:          "pending",
		Metadata:        &metadata,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	testCases := []struct {
		Name     string
		Metadata *map[string]any
		Input    *Payment
		Expected *PaymentDTO
	}{
		{
			Name: "it should return payment dto",
			Input: &Payment{
				UUID:            newUuid,
				MerchantID:      "merchant-id",
				CustomerID:      "customer-id",
				PaymentMethodID: "payment-method-id",
				Currency:        "IDR",
				Amount:          amount,
				Fee:             &fee,
				Discount:        &discount,
				TotalAmount:     totalAmount,
				Status:          "pending",
				Metadata:        &metadataMap,
				CreatedAt:       now,
				UpdatedAt:       now,
				PaymentMethod:   paymentMethod,
			},
			Expected: &PaymentDTO{
				UUID:            newUuid,
				MerchantID:      "merchant-id",
				CustomerID:      "customer-id",
				PaymentMethodID: "payment-method-id",
				Currency:        "IDR",
				Amount:          amount,
				Fee:             &fee,
				Discount:        &discount,
				TotalAmount:     totalAmount,
				Status:          "pending",
				Metadata:        &metadata,
				CreatedAt:       now,
				UpdatedAt:       now,
			},
		},
		{
			Name:     "it should return payment dto with nil metadata",
			Input:    payment,
			Expected: paymentDTO,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			require.Equal(t, tc.Expected, tc.Input.ToDTO())
		})
	}
}

func TestPaymentPaymentFromPaymentWithPaymentMethodDTO(t *testing.T) {
	amount := decimal.NewFromFloat(1000)
	now := time.Now()
	newUuid := uuid.NewString()
	fee := decimal.NewFromFloat(1000)
	discount := decimal.NewFromFloat(1000)
	metadata := "{\"testing\":\"testing\"}"
	metadataMap := map[string]any{"testing": "testing"}
	emptyString := ""

	paymentMethodDTO := PaymentWithPaymentMethodDTO{
		UUID:                  newUuid,
		MerchantID:            "merchant-id",
		CustomerID:            "customer-id",
		PaymentMethodID:       "payment-method-id",
		Currency:              "IDR",
		Amount:                amount,
		Fee:                   &fee,
		Discount:              &discount,
		TotalAmount:           amount,
		Status:                "pending",
		Metadata:              &metadata,
		CreatedAt:             now,
		UpdatedAt:             now,
		PaymentMethodType:     sql.NullString{String: "VA", Valid: true},
		PaymentMethodName:     sql.NullString{String: "VA Permata", Valid: true},
		PaymentMethodAcquirer: sql.NullString{String: "Permata", Valid: true},
		PaymentMethodLogo:     sql.NullString{String: "", Valid: true},
		PaymentMethodBankName: sql.NullString{String: "", Valid: true},
	}

	payment := &Payment{
		UUID:            newUuid,
		MerchantID:      "merchant-id",
		CustomerID:      "customer-id",
		PaymentMethodID: "payment-method-id",
		Currency:        "IDR",
		Amount:          amount,
		Fee:             &fee,
		Discount:        &discount,
		TotalAmount:     amount,
		Status:          "pending",
		Metadata:        &metadataMap,
		CreatedAt:       now,
		UpdatedAt:       now,
		PaymentMethod: PaymentMethod{
			Type:     "VA",
			Name:     "VA Permata",
			Acquirer: "Permata",
			BankName: &emptyString,
			Logo:     &emptyString,
		},
	}

	testCases := []struct {
		Name     string
		Input    *Payment
		Param    *PaymentWithPaymentMethodDTO
		Expected *Payment
	}{
		{
			Name:     "it should return payment with metadata",
			Input:    &Payment{},
			Param:    &paymentMethodDTO,
			Expected: payment,
		},
		{
			Name:  "if deletedAt is valid",
			Input: &Payment{},
			Param: &PaymentWithPaymentMethodDTO{
				UUID:                  newUuid,
				MerchantID:            "merchant-id",
				CustomerID:            "customer-id",
				PaymentMethodID:       "payment-method-id",
				Currency:              "IDR",
				Amount:                amount,
				Fee:                   &fee,
				Discount:              &discount,
				TotalAmount:           amount,
				Status:                "pending",
				Metadata:              &metadata,
				CreatedAt:             now,
				UpdatedAt:             now,
				PaymentMethodType:     sql.NullString{String: "VA", Valid: true},
				PaymentMethodName:     sql.NullString{String: "VA Permata", Valid: true},
				PaymentMethodAcquirer: sql.NullString{String: "Permata", Valid: true},
				DeletedAt:             sql.NullTime{Time: now, Valid: true},
				PaymentSnapCoreId:     util.ValueToPtr("12345"),
			},
			Expected: &Payment{
				UUID:            newUuid,
				MerchantID:      "merchant-id",
				CustomerID:      "customer-id",
				PaymentMethodID: "payment-method-id",
				Currency:        "IDR",
				Amount:          amount,
				Fee:             &fee,
				Discount:        &discount,
				TotalAmount:     amount,
				Status:          "pending",
				Metadata:        &metadataMap,
				CreatedAt:       now,
				UpdatedAt:       now,
				DeletedAt:       &now,
				PaymentMethod: PaymentMethod{
					Type:     "VA",
					Name:     "VA Permata",
					Acquirer: "Permata",
					BankName: &emptyString,
					Logo:     &emptyString,
				},
				SnapCoreId: util.ValueToPtr("12345"),
			},
		},
		{
			Name:  "if expiredAt is valid",
			Input: &Payment{},
			Param: &PaymentWithPaymentMethodDTO{
				UUID:                  newUuid,
				MerchantID:            "merchant-id",
				CustomerID:            "customer-id",
				PaymentMethodID:       "payment-method-id",
				Currency:              "IDR",
				Amount:                amount,
				Fee:                   &fee,
				Discount:              &discount,
				TotalAmount:           amount,
				Status:                "pending",
				Metadata:              &metadata,
				CreatedAt:             now,
				UpdatedAt:             now,
				PaymentMethodType:     sql.NullString{String: "VA", Valid: true},
				PaymentMethodName:     sql.NullString{String: "VA Permata", Valid: true},
				PaymentMethodAcquirer: sql.NullString{String: "Permata", Valid: true},
				DeletedAt:             sql.NullTime{Time: now, Valid: true},
				ExpiredAt:             sql.NullTime{Time: now, Valid: true},
				PaymentSnapCoreId:     util.ValueToPtr("12345"),
			},
			Expected: &Payment{
				UUID:            newUuid,
				MerchantID:      "merchant-id",
				CustomerID:      "customer-id",
				PaymentMethodID: "payment-method-id",
				Currency:        "IDR",
				Amount:          amount,
				Fee:             &fee,
				Discount:        &discount,
				TotalAmount:     amount,
				Status:          "pending",
				Metadata:        &metadataMap,
				CreatedAt:       now,
				UpdatedAt:       now,
				DeletedAt:       &now,
				ExpiredAt:       &now,
				PaymentMethod: PaymentMethod{
					Type:     "VA",
					Name:     "VA Permata",
					Acquirer: "Permata",
					BankName: &emptyString,
					Logo:     &emptyString,
				},
				SnapCoreId: util.ValueToPtr("12345"),
			},
		},
		{
			Name:  "with valid payment captures - single capture",
			Input: &Payment{},
			Param: &PaymentWithPaymentMethodDTO{
				UUID:                  newUuid,
				MerchantID:            "merchant-id",
				CustomerID:            "customer-id",
				PaymentMethodID:       "payment-method-id",
				Currency:              "IDR",
				Amount:                amount,
				Fee:                   &fee,
				Discount:              &discount,
				TotalAmount:           amount,
				Status:                "pending",
				Metadata:              &metadata,
				CreatedAt:             now,
				UpdatedAt:             now,
				PaymentMethodType:     sql.NullString{String: "CARD", Valid: true},
				PaymentMethodName:     sql.NullString{String: "Credit Card", Valid: true},
				PaymentMethodAcquirer: sql.NullString{String: "VISA", Valid: true},
				PaymentCapturesRaw: types.NullJSONText{
					Valid: true,
					JSONText: []byte(`[{
						"id": "capture-1",
						"paymentId": "payment-1",
						"processorReferenceId": "ref-1",
						"status": "SUCCESS",
						"releaseRemainingAmount": true,
						"currency": "IDR",
						"amount": 1000
					}]`),
				},
			},
			Expected: &Payment{
				UUID:            newUuid,
				MerchantID:      "merchant-id",
				CustomerID:      "customer-id",
				PaymentMethodID: "payment-method-id",
				Currency:        "IDR",
				Amount:          amount,
				Fee:             &fee,
				Discount:        &discount,
				TotalAmount:     amount,
				Status:          "pending",
				Metadata:        &metadataMap,
				CreatedAt:       now,
				UpdatedAt:       now,
				PaymentMethod: PaymentMethod{
					Type:     "CARD",
					Name:     "Credit Card",
					Acquirer: "VISA",
					BankName: &emptyString,
					Logo:     &emptyString,
				},
				PaymentCaptures: []*paymentCaptureModel.PaymentCapture{
					{
						ID:                     "capture-1",
						PaymentID:              "payment-1",
						ProcessorReferenceID:   util.ValueToPtr("ref-1"),
						Status:                 "SUCCESS",
						ReleaseRemainingAmount: true,
						Currency:               "IDR",
						Amount:                 1000,
					},
				},
			},
		},
		{
			Name:  "with empty payment captures JSON array",
			Input: &Payment{},
			Param: &PaymentWithPaymentMethodDTO{
				UUID:                  newUuid,
				MerchantID:            "merchant-id",
				CustomerID:            "customer-id",
				PaymentMethodID:       "payment-method-id",
				Currency:              "IDR",
				Amount:                amount,
				Fee:                   &fee,
				Discount:              &discount,
				TotalAmount:           amount,
				Status:                "pending",
				Metadata:              &metadata,
				CreatedAt:             now,
				UpdatedAt:             now,
				PaymentMethodType:     sql.NullString{String: "CARD", Valid: true},
				PaymentMethodName:     sql.NullString{String: "Credit Card", Valid: true},
				PaymentMethodAcquirer: sql.NullString{String: "VISA", Valid: true},
				PaymentCapturesRaw: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`[]`),
				},
			},
			Expected: &Payment{
				UUID:            newUuid,
				MerchantID:      "merchant-id",
				CustomerID:      "customer-id",
				PaymentMethodID: "payment-method-id",
				Currency:        "IDR",
				Amount:          amount,
				Fee:             &fee,
				Discount:        &discount,
				TotalAmount:     amount,
				Status:          "pending",
				Metadata:        &metadataMap,
				CreatedAt:       now,
				UpdatedAt:       now,
				PaymentMethod: PaymentMethod{
					Type:     "CARD",
					Name:     "Credit Card",
					Acquirer: "VISA",
					BankName: &emptyString,
					Logo:     &emptyString,
				},
				PaymentCaptures: []*paymentCaptureModel.PaymentCapture{},
			},
		},
		{
			Name:  "with null payment captures - not valid",
			Input: &Payment{},
			Param: &PaymentWithPaymentMethodDTO{
				UUID:                  newUuid,
				MerchantID:            "merchant-id",
				CustomerID:            "customer-id",
				PaymentMethodID:       "payment-method-id",
				Currency:              "IDR",
				Amount:                amount,
				Fee:                   &fee,
				Discount:              &discount,
				TotalAmount:           amount,
				Status:                "pending",
				Metadata:              &metadata,
				CreatedAt:             now,
				UpdatedAt:             now,
				PaymentMethodType:     sql.NullString{String: "CARD", Valid: true},
				PaymentMethodName:     sql.NullString{String: "Credit Card", Valid: true},
				PaymentMethodAcquirer: sql.NullString{String: "VISA", Valid: true},
				PaymentCapturesRaw: types.NullJSONText{
					Valid: false,
				},
			},
			Expected: &Payment{
				UUID:            newUuid,
				MerchantID:      "merchant-id",
				CustomerID:      "customer-id",
				PaymentMethodID: "payment-method-id",
				Currency:        "IDR",
				Amount:          amount,
				Fee:             &fee,
				Discount:        &discount,
				TotalAmount:     amount,
				Status:          "pending",
				Metadata:        &metadataMap,
				CreatedAt:       now,
				UpdatedAt:       now,
				PaymentMethod: PaymentMethod{
					Type:     "CARD",
					Name:     "Credit Card",
					Acquirer: "VISA",
					BankName: &emptyString,
					Logo:     &emptyString,
				},
				PaymentCaptures: nil,
			},
		},
		{
			Name:  "with invalid payment captures JSON - should not parse",
			Input: &Payment{},
			Param: &PaymentWithPaymentMethodDTO{
				UUID:                  newUuid,
				MerchantID:            "merchant-id",
				CustomerID:            "customer-id",
				PaymentMethodID:       "payment-method-id",
				Currency:              "IDR",
				Amount:                amount,
				Fee:                   &fee,
				Discount:              &discount,
				TotalAmount:           amount,
				Status:                "pending",
				Metadata:              &metadata,
				CreatedAt:             now,
				UpdatedAt:             now,
				PaymentMethodType:     sql.NullString{String: "CARD", Valid: true},
				PaymentMethodName:     sql.NullString{String: "Credit Card", Valid: true},
				PaymentMethodAcquirer: sql.NullString{String: "VISA", Valid: true},
				PaymentCapturesRaw: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`invalid json`),
				},
			},
			Expected: &Payment{
				UUID:            newUuid,
				MerchantID:      "merchant-id",
				CustomerID:      "customer-id",
				PaymentMethodID: "payment-method-id",
				Currency:        "IDR",
				Amount:          amount,
				Fee:             &fee,
				Discount:        &discount,
				TotalAmount:     amount,
				Status:          "pending",
				Metadata:        &metadataMap,
				CreatedAt:       now,
				UpdatedAt:       now,
				PaymentMethod: PaymentMethod{
					Type:     "CARD",
					Name:     "Credit Card",
					Acquirer: "VISA",
					BankName: &emptyString,
					Logo:     &emptyString,
				},
				PaymentCaptures: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Input.PaymentFromPaymentWithPaymentMethodDTO(tc.Param)
			require.Equal(t, tc.Expected, tc.Input)
		})
	}
}

func TestToPaymentCreditCardCallbackRequest(t *testing.T) {

	payment := Payment{}

	tests := []struct {
		input      *Payment
		wantErr    error
		wantResult *pb.PaymentCreditCardCallbackRequest
	}{
		{
			input: &Payment{
				Metadata: &map[string]any{"chan": make(chan bool, 1)},
			},
			wantErr: errors.New("json.Marshal: json: unsupported type: chan bool"),
		},
		{
			input: &Payment{
				Metadata: &map[string]any{"authenticationMethod": false},
			},
			wantErr: errors.New("json.Unmarshal: json: cannot unmarshal bool into Go struct field CreditcardMetadata.authenticationMethod of type string"),
		},
		{
			input: &Payment{
				ReferenceID: util.ValueToPtr("ABC"),
				Metadata: &map[string]any{
					"authenticationMethod": "OTP",
					"processorStatus":      "SUCCESS",
					"cardData":             map[string]any{},
				},
			},
			wantResult: &pb.PaymentCreditCardCallbackRequest{
				ReferenceId:   "ABC",
				Created:       payment.CreatedAt.Format(time.RFC3339),
				PaymentStatus: "PAID",
				Updated:       payment.UpdatedAt.Format(time.RFC3339),
				CardData:      &pb.PaymentCreditCardData{},
				Amount:        "0",
			},
		},
	}
	for _, test := range tests {
		result, err := test.input.ToPaymentCreditCardCallbackRequest()
		assert.Equal(t, test.wantErr, err)
		assert.Equal(t, test.wantResult, result)
	}
}

func TestPaymentStatus(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "SUCCESS: Pending Status",
			input:   paymentConstant.PAYMENT_STATUS_PENDING,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Success Status",
			input:   paymentConstant.PAYMENT_STATUS_SUCCESS,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Void Status",
			input:   paymentConstant.PAYMENT_STATUS_VOID,
			wantErr: false,
		},
		{
			name:    "ERROR: Incorrect Status",
			input:   "MIXED",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePaymentStatus(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidatePaymentHistorySortColumn(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "SUCCESS: createdAt Column",
			input:   "createdAt",
			wantErr: false,
		},
		{
			name:    "SUCCESS: amount Column",
			input:   "amount",
			wantErr: false,
		},
		{
			name:    "SUCCESS: amountPaid Column",
			input:   "amountPaid",
			wantErr: false,
		},
		{
			name:    "ERROR: Incorrect Sort Column",
			input:   "MIXED",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePaymentHistorySortColumn(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateUnifiedPaymentStatus(t *testing.T) {
	testCases := []struct {
		name          string
		status        string
		expectedError error
	}{
		{
			name:          "SUCCESS: Pending Status",
			status:        paymentConstant.PAYMENT_STATUS_PENDING,
			expectedError: nil,
		},
		{
			name:          "SUCCESS: Success Status",
			status:        paymentConstant.PAYMENT_STATUS_SUCCESS,
			expectedError: nil,
		},
		{
			name:          "SUCCESS: Void Status",
			status:        paymentConstant.PAYMENT_STATUS_VOID,
			expectedError: nil,
		},
		{
			name:          "Valid status - REQUIRE_PAYMENT_METHOD",
			status:        constant.UnifiedPaymentSessionStatusRequirePaymentMethod,
			expectedError: nil,
		},
		{
			name:          "Valid status - REQUIRE_CONFIRMATION",
			status:        constant.UnifiedPaymentSessionStatusRequireConfirmation,
			expectedError: nil,
		},
		{
			name:          "Valid status - REQUIRE_ACTION",
			status:        constant.UnifiedPaymentSessionStatusRequireAction,
			expectedError: nil,
		},
		{
			name:          "Valid status - PROCESSING",
			status:        constant.UnifiedPaymentSessionStatusProcessing,
			expectedError: nil,
		},
		{
			name:          "Valid status - CANCELLED",
			status:        constant.UnifiedPaymentSessionStatusCancelled,
			expectedError: nil,
		},
		{
			name:          "Valid status - EXPIRED",
			status:        constant.UnifiedPaymentSessionStatusExpired,
			expectedError: nil,
		},
		{
			name:          "Valid status - PAID",
			status:        constant.UnifiedPaymentSessionStatusPaid,
			expectedError: nil,
		},
		{
			name:          "Invalid status",
			status:        "INVALID_STATUS",
			expectedError: constant.ErrInvalidPaymentStatus,
		},
		{
			name:          "Case insensitive - lower case",
			status:        "paid",
			expectedError: nil,
		},
		{
			name:          "Case insensitive - mixed case",
			status:        "Paid",
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUnifiedPaymentStatus(tc.status)
			if tc.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tc.expectedError, err)
			}
		})
	}
}
func TestToUnifiedPaymentResponse(t *testing.T) {
	now := time.Now()
	metadata := map[string]any{
		"autoConfirm": false,
		"mode":        "test",
		"clientRedirectUrl": map[string]any{
			"successReturnUrl":    "https://success.url",
			"failureReturnUrl":    "https://failure.url",
			"expirationReturnUrl": "https://expiration.url",
		},
		"paymentMethod": map[string]any{
			"type": "CARD",
		},
		"statementDescriptor": "Test Descriptor",
		"clientMetadata": map[string]any{
			"key": "value",
		},
		"encryptedEncryptionKey": map[string]any{
			"publicKey":     `yp7rgxAudc9lgMYDdThABgvZP9UXr8pkn5G8mQW58/oTXbfTA/Jcluz6sRKWOy+5Xpqtd6+mJgaHu0x/5f/nYpD7QDXFKgymTIGnUURqdL+1cLIha96AbeNGaqdrC6RA7igbKckJHc6Q56+VS9g0SOddrLI6pClyAunYQSsBIXhSzkdbFCmZlAJpyYtMx2QEHvNe3cNSDxe2yjzduTo7X3p9dG05f7Jkx3AtO4THohqdByl+onPjovM8WBsGtT5bUE9OnlzWnvzemIHG1gZAWusv2zAnqTocKBORymPZODXu6QV22eMzRlLES9UhuJgEM7mxlXxeJI8He2yKPWtfjZ2zwTUaE7fM1gm5yvQTGRqimJFCNb7t+iixqh+s57fmm7ruzrAkFBZ9J8catnlwWr9o/HJprPzQCyj+9F3oTAhmV3CiaTbqhYVJixkaSZPugfjsmcvi7ADEAui/Gp0V2UQ1VN+dsEhy7BoCmb6seGxQiJ3mm/qIIf5Lou+vUToEGwa3Ua0kW/mubbAYRS6oNBDEZbycd95T0m9UL3eeTr8ma5cn`,
			"privateKey":    `qw4HeMoDg3AS1CY9io9AO+8HN3HZr6UsRJ+tc8p+M7GZmhkmQzh5/wt4uIDNDFrNi6TbwfHbrfeNw/Jn8T7xIR2wLySo3uIGujGRht0M/7muF0wJSslLJtyhUJvZtjkTZ47HpP8vzt3dds7WRG9dQFKEs2IhKHMFLYfByGB4OljQKtbf1zT+Mcg3TzflV+zqhxXCG5GX/4n170YDjc++5c24rdyHiZlq11Wrqe5Jdm5TMOUsSRQNl1V9EQMirYh0xHkUu+NPQ7NSzCMvKkFyo4gxrNndH0YqsJyvcQojfonHNpUPEIHTK4Aykrfs58V2xXuZq1gYDTYb4EkkAB5ssMMqQDkb4m/BeVmWIq/jBUZPFrfcHBLD7JYZKb6mXp9+oInjjkPHCVkGhnpciXSxBA1zO/nftu4oWsKYKpsS524YmQ50owaI3b2W/2njtIW5BYsHDhEBzuPXTKk3tPltjnrE/Jx1Yrmww+vvrvutK8IBAPLF5iKetWp9F3qFpJqX/BGf2jYruputY71fNyw7SLRqMEZYeySF3O32tY/zL0D478gXCbLBX1VQK2Y3pTys7diLDREOl2L+APiohyPvUtMwCaSPpiJ0N7Ek7Ainp2eJfFg4LRRDyftxm8okM8bqeERSjeCgsKDi6llvOdybvOhjUILvCUKvmA36lTmnXk3x2JqRjTqlJwBfPNSXPWShjMimvAbAH/VW53XbvCp/zZEC3riQRK3HesmurQyYfkK+cNGoj042LlGuk6LRiAqWietUAq8ybGYfpijfW2mTN7lcz6ocZlwikjJHf/FJ6OTdTF+mBTCVEsGCgkgtJJwGvubntjzTnTY5Wrr7XwF4cV4QTrGodzFMQZsPZhFQoqYltXcklcYy/l3UTQ4oTfYOjU68/vmEsb2XjWHi0B8ZBLxUa1WNVk5O6aMm7lqLfhOvm3B5K4hsN2NAHQbISpORh+Q1e+fcNbTtE+bQY6mlrYVUuCuJzqoVrOsAyPWTkXPZ96pSuS8j3bTnltT9NwdhqARZGjjnPztCDfmhd0nWx5YbObcHiB2os8bGi9VqPcEgwISX+vFYST4YIcojV++G1dOj6Q77YP7CXmbH25EnDWEdyj0YtMRtgqZ0UpMojGRgOJEzB6Gz9wytJixYePnvMe7YDESjiw1NROMhj+LXwlYY74xtT1jfvZwPoz+9HO+qxzn6LxlA47rV0p0/iNwK8JDom2+rf9V8iwRg637AoSPQfS1kJ5cHgaQiNtC7AKsW5t3GGuKpaaCOUJkYtbBqb5Kz1SeZMyycVM/zfBJ9bcbnEKmK3PzK2a9LzOti4CaFc+9zR1epUXfUUnP+qWT6Zp1UCvdRLWMsTkE2xJw2xNHxMTNG3O5R2Vz/Jso3kutII4nNxQRk4SIHlMQO8MtGJ1Zgq4XXnwUL+SjPRxbLIVXqnBzXey9f36/9Xe8o3f3gjY/1JngnAqrYzLT8Mx12iSGq/AZPzHZXCrKbZLJTDtr+Uxr6BuS3+5O8n8RgiDCpO3+sNhp+VzVOiFIqnEejqBWk/qJsA+f23Qc240A+OxdOSRmDmOxdHQNmbFgbvByc3QqMUJcO9If/eqoA2bveTJb/yYneiObcrU1OaEaWJMCwdeFKWBmF0vZyuMApWWtNZePCOK0wkxUUWP2CLuAjUsAkQ2egnZk7wT9n3N/oHSo8x5SgpX2j9/xLuUFX/m/Q+3FMjH3lPRDJLXTVrE6jIhVsBmu6yV3DYD5/t1+fRv5lmJGJod4drzj5kXA7yT7VL72A36G+fony69eIypkfl3tkaXfsZRzmSN6IPwQ55B6X6OYwTdYOHerkI6GZq1uJ0lAYF2X7X/YppEwwfSPt8RjFHBuJ1+4wRLObW0UuMRooS5gx8iyTarAehhNniJKAeR63MYXdAWJuK/LFFjyNtuj4lEhl6FfEV7WcuCkMUbqdbxPr8b0x9AZn3Saz3A6VQtfsH1FefSCOgFnwaSbmGDBhgRS/4v6c6TXKUl4XHn+iYIksi+mXGa/lVBDTTDTdeVcjCmC8A4vvaTlvGHLe55G/RKoNRuBAnza6vuLqYHVAhWzIalWsi3AqtZr2/0maT80RpbAx1GuC/PoD+mUKQYFrwsHJOn8EwStVwwHKE9dToMX3OD8dik+BjDEEPf6qjKzm24Zvn6dE7jK2QEZz09Ry8EXnBqQfwrLWLbdkp2X/8zo=`,
			"secretVersion": 3,
		},
	}

	secretManagerClient := gcpMock.NewIGCPSecretManager(t)

	secretManagerClient.On(
		"GetSecretValueJSON", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Once().Run(func(args mock.Arguments) {
		*args.Get(4).(*commonModel.EncryptionSecret) = commonModel.EncryptionSecret{
			Payment: commonModel.EncryptionSecretPayment{
				KeyEncryptionKey: "MORBApyDC5lpOtcogh4dPbZ9rGtby2g4", // NOSONAR
			},
		}
	}).Return(nil)

	payment := &Payment{
		UUID:        "test-uuid",
		ReferenceID: util.ValueToPtr("test-reference-id"),
		MerchantID:  "test-merchant-id",
		Currency:    "IDR",
		Amount:      decimal.NewFromFloat(1000),
		Status:      "pending",
		Metadata:    &metadata,
		PaymentURL:  "https://payment.url",
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiredAt:   &now,
	}

	expected1 := unifiedPaymentModel.UnifiedPaymentSessionResponse{
		ID:                "test-uuid",
		ClientReferenceID: "test-reference-id",
		Amount: unifiedPaymentModel.Amount{
			Currency: "IDR",
			Value:    1000,
		},
		AutoConfirm: false,
		Mode:        "test",
		RedirectUrl: unifiedPaymentModel.RedirectUrl{
			SuccessReturnUrl:    "https://success.url",
			FailureReturnUrl:    "https://failure.url",
			ExpirationReturnUrl: "https://expiration.url",
		},
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{
			Type: "CARD",
		},
		StatementDescriptor: "Test Descriptor",
		ExpiryAt:            &now,
		Status:              "pending",
		CreatedAt:           now,
		UpdatedAt:           now,
		PaymentUrl:          "https://payment.url",
		Metadata: map[string]any{
			"key": "value",
		},
	}
	gcp.SetGlobalSecretManagerClient(nil)

	resultRaw1, _ := json.Marshal(payment.ToUnifiedPaymentResponse())
	expectedRaw1, _ := json.Marshal(expected1)
	assert.JSONEq(t, string(expectedRaw1), string(resultRaw1))

	expected2 := expected1
	gcp.SetGlobalSecretManagerClient(secretManagerClient)

	expected2.EncryptionKey = &encrypt.RSAEncryptedKey{}
	expected2.EncryptionKey.SetPublicKey(`MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyjG5dkSoh/v2lQruMKxXKnpvZUEx+oH/KRFBYHzgEV55ooEVViTJHgjoVlt34xpLRJOM89Gw036aF6it9yCKKUJDZbVnxZpWSuRNSv4IibQGGJ+9t9KL4bn2WyIq/19eABfLOp95Iq60weOkGslqWHVWm1KvopZ3QtShr16LHVo2q5CV4+W2FkQ8zKLMb79U20cc5YDQhuJUgwGH7r5QXH85Wc2jpvXm5a/VEELoGs4kh1gJt/gb6TdYjfTEA+gUalsQVCVG6V7GroND36FCqaChx9ONu8oS6+fRzloPNPXRpdyy2yO2obTlWE1Nh7BVVG3EXkDFxYiow/uHv+vL1QIDAQAB`)

	resultRaw2, _ := json.Marshal(payment.ToUnifiedPaymentResponse())
	expectedRaw2, _ := json.Marshal(expected2)
	assert.JSONEq(t, string(expectedRaw2), string(resultRaw2))

	secretManagerClient.AssertExpectations(t)
}

func TestToUnifiedPaymentAndChargeResponse(t *testing.T) {
	now := time.Now()
	chargeUUID := uuid.New()
	chargeAdditionalInfo := `{
		"chargeStatus": "SUCCESS",
		"statementDescriptor": "Test Descriptor",
		"methodDetail": {
			"key": "value"
		}
	}`

	testCases := []struct {
		name     string
		payment  *Payment
		charge   *orchestrator_model.AccountTransactionWithUseCase
		expected *unifiedPaymentModel.UnifiedPaymentSessionResponse
	}{
		{
			name: "SUCCESS: With charge details",
			payment: &Payment{
				UUID:        "payment-uuid",
				ReferenceID: util.ValueToPtr("reference-id"),
				MerchantID:  "merchant-id",
				Amount:      decimal.NewFromFloat(1000),
				Currency:    "IDR",
				Status:      constant.StatusSuccess,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			charge: &orchestrator_model.AccountTransactionWithUseCase{
				UUID:                 chargeUUID,
				Currency:             "IDR",
				Credit:               1000,
				CreatedAt:            now,
				UpdatedAt:            now,
				TransactionTimestamp: now,
				Status:               constant.ChargeStatusSuccess,
				AdditionalInfo:       types.NullJSONText{JSONText: []byte(chargeAdditionalInfo), Valid: true},
			},
			expected: &unifiedPaymentModel.UnifiedPaymentSessionResponse{
				ID:                "payment-uuid",
				ClientReferenceID: "reference-id",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    1000,
				},
				Status:    constant.StatusSuccess,
				CreatedAt: now,
				UpdatedAt: now,

				ChargeDetails: []*unifiedPaymentModel.ChargeResponse{
					{
						ID:                              chargeUUID.String(),
						PaymentSessionID:                "payment-uuid",
						PaymentSessionClientReferenceID: "reference-id",
						Amount: unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    1000,
						},
						StatementDescriptor: "Test Descriptor",
						Status:              constant.ChargeStatusSuccess,
						IsCaptured:          true,
						AuthorizedAmount: &unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    1000,
						},
						CapturedAmount: &unifiedPaymentModel.Amount{
							Currency: "IDR",
							Value:    1000,
						},

						ChargePaymentMethodDetails: &unifiedPaymentModel.ChargePaymentMethodDetails{},
						CreatedAt:                  now,
						UpdatedAt:                  now,
						PaidAt:                     &now,
					},
				},
			},
		},
		{
			name: "SUCCESS: Without charge details",
			payment: &Payment{
				UUID:        "payment-uuid",
				ReferenceID: util.ValueToPtr("reference-id"),
				MerchantID:  "merchant-id",
				Amount:      decimal.NewFromFloat(1000),
				Currency:    "IDR",
				Status:      constant.StatusSuccess,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			charge: nil,
			expected: &unifiedPaymentModel.UnifiedPaymentSessionResponse{
				ID:                "payment-uuid",
				ClientReferenceID: "reference-id",
				Amount: unifiedPaymentModel.Amount{
					Currency: "IDR",
					Value:    1000,
				},
				Status:    constant.StatusSuccess,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.payment.ToUnifiedPaymentAndChargeResponse(tc.charge)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToPbUnifiedPaymentV2CallbackRequestWithFdsRiskAssessment(t *testing.T) {
	now := time.Now()
	chargeUUID := uuid.New()

	// Create charge additional info with FDS risk assessment
	chargeAdditionalInfo := `{
		"chargeStatus": "SUCCESS",
		"statementDescriptor": "Test Descriptor",
		"methodDetail": {
			"ewallet": {"channel":"DANA"}
		},
		"fdsRiskAssessment": {
			"score": 75.5,
			"level": "MEDIUM",
			"recommendation": "APPROVE",
			"status": "PASSED",
			"evaluatedAt": "2023-01-01T12:00:00Z"
		}
	}`

	payment := &Payment{
		UUID:        "payment-uuid",
		ReferenceID: util.ValueToPtr("reference-id"),
		MerchantID:  "merchant-id",
		Amount:      decimal.NewFromFloat(1000),
		Currency:    "IDR",
		Status:      constant.StatusSuccess,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata: &map[string]any{
			"autoConfirm": true,
			"mode":        "REDIRECT",
			"clientRedirectUrl": map[string]any{
				"successReturnUrl":    "https://success.url",
				"failureReturnUrl":    "https://failure.url",
				"expirationReturnUrl": "https://expiration.url",
			},
			"paymentMethod": map[string]any{
				"type": "EWALLET",
			},
			"statementDescriptor": "Test Statement",
		},
	}

	charge := &orchestrator_model.AccountTransactionWithUseCase{
		UUID:                 chargeUUID,
		ReferenceID:          payment.UUID,
		ClientReferenceID:    util.ValueOfPtr(payment.ReferenceID),
		Type:                 constant.TypePayment,
		Currency:             "IDR",
		Credit:               1000,
		CreatedAt:            now,
		UpdatedAt:            now,
		TransactionTimestamp: now,
		Status:               constant.ChargeStatusSuccess,
		AdditionalInfo:       types.NullJSONText{JSONText: []byte(chargeAdditionalInfo), Valid: true},
	}

	customer := &unifiedPaymentModel.CustomerInformationResponse{
		CustomerID: "test-customer-id",
		GivenName:  "John",
		SureName:   "Doe",
		Surname:    util.ValueToPtr("Doe"),
		Email:      "john.doe@example.com",
		PhoneNumber: &unifiedPaymentModel.UnifiedPaymentPhoneNumber{
			CountryCode: "+62",
			Number:      "81234567890",
		},
		RefundPreference: &unifiedPaymentModel.UnifiedPaymentRefundPreference{
			Method: "AUTO",
			TransferDestination: &unifiedPaymentModel.RefundTransferDestination{
				ChannelCode: "BCA",
				ChannelInformation: &unifiedPaymentModel.RefundChannelInformation{
					AccountNumber: "1234567890",
					AccountName:   "John Doe",
				},
			},
		},
		StoredPaymentMethods: []*unifiedPaymentModel.CustomerPaymentMethodResponse{
			{
				Token:          "token-123",
				PaymentMethod:  "CARD",
				PaymentChannel: "VISA",
				Status:         "ACTIVE",
				CreatedAt:      now,
				Card: &unifiedPaymentModel.CustomerPaymentMethodCardResponse{
					Fingerprint:         "fp-123",
					Network:             "VISA",
					First6:              "424242",
					First8:              "42424242",
					Last4:               "4242",
					ExpMonth:            "12",
					ExpYear:             "2025",
					CardHolderFirstName: "John",
					CardHolderLastName:  "Doe",
					CardHolderEmail:     "john.doe@example.com",
					CardHolderPhone:     "+6281234567890",
				},
			},
		},
	}

	result := payment.ToPbUnifiedPaymentV2CallbackRequest(charge, customer)

	// Verify the callback request structure
	assert.NotNil(t, result)
	assert.Equal(t, "payment-uuid", result.Id)
	assert.Equal(t, "reference-id", result.ClientReferenceId)
	assert.Equal(t, "IDR", result.Amount.Currency)
	assert.Equal(t, float64(1000), result.Amount.Value)
	assert.True(t, result.AutoConfirm)
	assert.Equal(t, "REDIRECT", result.Mode)
	assert.Equal(t, "EWALLET", result.PaymentMethod.Type)

	// Verify redirect URLs
	assert.Equal(t, "https://success.url", result.RedirectUrl.SuccessReturnUrl)
	assert.Equal(t, "https://failure.url", result.RedirectUrl.FailureReturnUrl)
	assert.Equal(t, "https://expiration.url", result.RedirectUrl.ExpirationReturnUrl)

	// Verify charge details
	assert.Len(t, result.ChargeDetails, 1)
	chargeDetail := result.ChargeDetails[0]

	assert.Equal(t, chargeUUID.String(), chargeDetail.Id)
	assert.Equal(t, "payment-uuid", chargeDetail.PaymentSessionId)
	assert.Equal(t, "reference-id", chargeDetail.PaymentSessionClientReferenceId)
	assert.Equal(t, "IDR", chargeDetail.Amount.Currency)
	assert.Equal(t, float64(1000), chargeDetail.Amount.Value)
	assert.Equal(t, "Test Descriptor", chargeDetail.StatementDescriptor)
	assert.Equal(t, constant.ChargeStatusSuccess, chargeDetail.Status)
	assert.Equal(t, "DANA", chargeDetail.Ewallet.Channel)

	// Verify customer information
	assert.NotNil(t, result.Customer)
	assert.Equal(t, "test-customer-id", result.Customer.CustomerId)
	assert.Equal(t, "John", result.Customer.GivenName)
	assert.Equal(t, "Doe", result.Customer.SureName)
	if result.Customer.Surname != nil {
		assert.Equal(t, "Doe", *result.Customer.Surname)
	}
	assert.Equal(t, "john.doe@example.com", result.Customer.Email)

	// Verify phone number
	assert.NotNil(t, result.Customer.PhoneNumber)
	assert.Equal(t, "+62", result.Customer.PhoneNumber.CountryCode)
	assert.Equal(t, "81234567890", result.Customer.PhoneNumber.Number)

	// Verify refund preference
	assert.NotNil(t, result.Customer.RefundPreference)
	assert.Equal(t, "AUTO", result.Customer.RefundPreference.Method)
	assert.NotNil(t, result.Customer.RefundPreference.TransferDestination)
	assert.Equal(t, "BCA", result.Customer.RefundPreference.TransferDestination.ChannelCode)
	assert.NotNil(t, result.Customer.RefundPreference.TransferDestination.ChannelInformation)
	assert.Equal(t, "1234567890", result.Customer.RefundPreference.TransferDestination.ChannelInformation.AccountNumber)
	assert.Equal(t, "John Doe", result.Customer.RefundPreference.TransferDestination.ChannelInformation.AccountName)

	// Verify stored payment methods - note: current implementation creates slice with length and then appends, so we get double the items
	assert.Len(t, result.Customer.StoredPaymentMethods, 1)
	storedMethod := result.Customer.StoredPaymentMethods[0]
	assert.Equal(t, "token-123", storedMethod.Token)
	assert.Equal(t, "CARD", storedMethod.PaymentMethod)
	assert.Equal(t, "VISA", storedMethod.PaymentChannel)
	assert.Equal(t, "ACTIVE", storedMethod.Status)

	// Verify card details
	assert.NotNil(t, storedMethod.Card)
	assert.Equal(t, "VISA", storedMethod.Card.Network)
	assert.Equal(t, "424242", storedMethod.Card.First6)
	assert.Equal(t, "42424242", storedMethod.Card.First8)
	assert.Equal(t, "4242", storedMethod.Card.Last4)
	assert.Equal(t, "12", storedMethod.Card.ExpMonth)
	assert.Equal(t, "2025", storedMethod.Card.ExpYear)
	assert.Equal(t, "John", storedMethod.Card.CardHolderFirstName)
	assert.Equal(t, "Doe", storedMethod.Card.CardHolderLastName)
	assert.Equal(t, "john.doe@example.com", storedMethod.Card.CardHolderEmail)
	assert.Equal(t, "+6281234567890", storedMethod.Card.CardHolderPhone)

	// Note: FdsRiskAssessment is not included in the protobuf structure.
	// It will be available when the protobuf is converted to JSON and unmarshaled
	// into UnifiedPaymentSessionResponse which has the FdsRiskAssessment field.
}

func TestToPbUnifiedPaymentV2CallbackRequest_WithoutFdsRiskAssessment(t *testing.T) {
	now := time.Now()
	chargeUUID := uuid.New()

	// Create charge additional info without FDS risk assessment
	chargeAdditionalInfo := `{
		"chargeStatus": "SUCCESS",
		"statementDescriptor": "Test Descriptor",
		"methodDetail": {
			"key": "value"
		}
	}`

	payment := &Payment{
		UUID:        "payment-uuid",
		ReferenceID: util.ValueToPtr("reference-id"),
		MerchantID:  "merchant-id",
		Amount:      decimal.NewFromFloat(1000),
		Currency:    "IDR",
		Status:      constant.StatusSuccess,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata: &map[string]any{
			"autoConfirm": true,
			"mode":        "REDIRECT",
			"paymentMethod": map[string]any{
				"type": "CARD",
			},
		},
	}

	charge := &orchestrator_model.AccountTransactionWithUseCase{
		UUID:                 chargeUUID,
		Currency:             "IDR",
		Credit:               1000,
		CreatedAt:            now,
		UpdatedAt:            now,
		TransactionTimestamp: now,
		Status:               constant.ChargeStatusSuccess,
		AdditionalInfo:       types.NullJSONText{JSONText: []byte(chargeAdditionalInfo), Valid: true},
	}

	// Test with nil customer
	result := payment.ToPbUnifiedPaymentV2CallbackRequest(charge, nil)

	// Verify charge details exist
	assert.Len(t, result.ChargeDetails, 1)

	// Verify customer information is nil when not provided
	assert.Nil(t, result.Customer)

	// Note: FdsRiskAssessment is not included in the protobuf structure.
	// It will be available (or nil) when the protobuf is converted to JSON and unmarshaled
	// into UnifiedPaymentSessionResponse which has the FdsRiskAssessment field.
}

func TestToPbUnifiedPaymentV2CallbackRequest_CustomerEdgeCases(t *testing.T) {
	now := time.Now()
	chargeUUID := uuid.New()

	chargeAdditionalInfo := `{
		"chargeStatus": "SUCCESS",
		"statementDescriptor": "Test Descriptor",
		"methodDetail": {"key": "value"}
	}`

	payment := &Payment{
		UUID:        "payment-uuid",
		ReferenceID: util.ValueToPtr("reference-id"),
		MerchantID:  "merchant-id",
		Amount:      decimal.NewFromFloat(1000),
		Currency:    "IDR",
		Status:      constant.StatusSuccess,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata: &map[string]any{
			"autoConfirm": true,
			"mode":        "REDIRECT",
			"paymentMethod": map[string]any{
				"type": "CARD",
			},
		},
	}

	charge := &orchestrator_model.AccountTransactionWithUseCase{
		UUID:                 chargeUUID,
		Currency:             "IDR",
		Credit:               1000,
		CreatedAt:            now,
		UpdatedAt:            now,
		TransactionTimestamp: now,
		Status:               constant.ChargeStatusSuccess,
		AdditionalInfo:       types.NullJSONText{JSONText: []byte(chargeAdditionalInfo), Valid: true},
	}

	t.Run("Customer with minimal information", func(t *testing.T) {
		customer := &unifiedPaymentModel.CustomerInformationResponse{
			CustomerID: "minimal-customer",
			GivenName:  "Jane",
			SureName:   "Smith",
			Surname:    util.ValueToPtr("Smith"),
			Email:      "jane.smith@example.com",
		}

		result := payment.ToPbUnifiedPaymentV2CallbackRequest(charge, customer)

		assert.NotNil(t, result.Customer)
		assert.Equal(t, "minimal-customer", result.Customer.CustomerId)
		assert.Equal(t, "Jane", result.Customer.GivenName)
		assert.Equal(t, "Smith", result.Customer.SureName)
		if result.Customer.Surname != nil {
			assert.Equal(t, "Smith", *result.Customer.Surname)
		}
		assert.Equal(t, "jane.smith@example.com", result.Customer.Email)
		assert.Nil(t, result.Customer.PhoneNumber)
		assert.Nil(t, result.Customer.RefundPreference)
		assert.Empty(t, result.Customer.StoredPaymentMethods)
	})

	t.Run("Customer with phone number only", func(t *testing.T) {
		customer := &unifiedPaymentModel.CustomerInformationResponse{
			CustomerID: "phone-customer",
			GivenName:  "Bob",
			SureName:   "Johnson",
			Surname:    util.ValueToPtr("Johnson"),
			Email:      "bob.johnson@example.com",
			PhoneNumber: &unifiedPaymentModel.UnifiedPaymentPhoneNumber{
				CountryCode: "+1",
				Number:      "5551234567",
			},
		}

		result := payment.ToPbUnifiedPaymentV2CallbackRequest(charge, customer)

		assert.NotNil(t, result.Customer.PhoneNumber)
		assert.Equal(t, "+1", result.Customer.PhoneNumber.CountryCode)
		assert.Equal(t, "5551234567", result.Customer.PhoneNumber.Number)
		assert.Equal(t, "Johnson", result.Customer.SureName)
		if result.Customer.Surname != nil {
			assert.Equal(t, "Johnson", *result.Customer.Surname)
		}
		assert.Nil(t, result.Customer.RefundPreference)
		assert.Empty(t, result.Customer.StoredPaymentMethods)
	})

	t.Run("Customer with refund preference without transfer destination", func(t *testing.T) {
		customer := &unifiedPaymentModel.CustomerInformationResponse{
			CustomerID: "refund-customer",
			GivenName:  "Alice",
			SureName:   "Williams",
			Surname:    util.ValueToPtr("Williams"),
			Email:      "alice.williams@example.com",
			RefundPreference: &unifiedPaymentModel.UnifiedPaymentRefundPreference{
				Method: "AUTO",
			},
		}

		result := payment.ToPbUnifiedPaymentV2CallbackRequest(charge, customer)

		assert.NotNil(t, result.Customer.RefundPreference)
		assert.Equal(t, "AUTO", result.Customer.RefundPreference.Method)
		assert.Nil(t, result.Customer.RefundPreference.TransferDestination)
		assert.Equal(t, "Williams", result.Customer.SureName)
		if result.Customer.Surname != nil {
			assert.Equal(t, "Williams", *result.Customer.Surname)
		}
	})

	t.Run("Customer with stored payment method without card details", func(t *testing.T) {
		customer := &unifiedPaymentModel.CustomerInformationResponse{
			CustomerID: "stored-payment-customer",
			GivenName:  "Charlie",
			SureName:   "Brown",
			Email:      "charlie.brown@example.com",
			StoredPaymentMethods: []*unifiedPaymentModel.CustomerPaymentMethodResponse{
				{
					Token:          "token-456",
					PaymentMethod:  "EWALLET",
					PaymentChannel: "OVO",
					Status:         "INACTIVE",
					CreatedAt:      now,
				},
			},
		}

		result := payment.ToPbUnifiedPaymentV2CallbackRequest(charge, customer)

		// Due to implementation bug, we get double the items (first will be nil)
		assert.Len(t, result.Customer.StoredPaymentMethods, 1)
		storedMethod := result.Customer.StoredPaymentMethods[0]
		assert.Equal(t, "token-456", storedMethod.Token)
		assert.Equal(t, "EWALLET", storedMethod.PaymentMethod)
		assert.Equal(t, "OVO", storedMethod.PaymentChannel)
		assert.Equal(t, "INACTIVE", storedMethod.Status)
		assert.Nil(t, storedMethod.Card)
	})

	t.Run("Customer with both SureName and Surname fields - verify both are populated", func(t *testing.T) {
		customer := &unifiedPaymentModel.CustomerInformationResponse{
			CustomerID: "both-fields-customer",
			GivenName:  "David",
			SureName:   "Brown",
			Surname:    util.ValueToPtr("Brown"),
			Email:      "david.brown@example.com",
		}

		result := payment.ToPbUnifiedPaymentV2CallbackRequest(charge, customer)

		assert.NotNil(t, result.Customer)
		assert.Equal(t, "both-fields-customer", result.Customer.CustomerId)
		assert.Equal(t, "David", result.Customer.GivenName)
		assert.Equal(t, "Brown", result.Customer.SureName)
		if result.Customer.Surname != nil {
			assert.Equal(t, "Brown", *result.Customer.Surname)
		}
		assert.Equal(t, "david.brown@example.com", result.Customer.Email)
	})
}

func TestToInternalCreditcardGetPaymentByUUID_ProcessingConfig(t *testing.T) {
	testCases := []struct {
		desc              string
		metadata          map[string]any
		expectedBankMID   string
		expectedMIDTag    string
		expectedAutoSplit *creditcardModel.AutoSplitPayment
		wantError         bool
	}{
		{
			desc: "success with both bank merchant id and merchant id tag",
			metadata: map[string]any{
				"paymentMethodOptions": map[string]any{
					"card": map[string]any{
						"processingConfig": map[string]any{
							"bankMerchantId": "BANK123",
							"merchantIdTag":  "tag1",
						},
					},
				},
			},
			expectedBankMID: "BANK123",
			expectedMIDTag:  "tag1",
			wantError:       false,
		},
		{
			desc: "success with only bank merchant id",
			metadata: map[string]any{
				"paymentMethodOptions": map[string]any{
					"card": map[string]any{
						"processingConfig": map[string]any{
							"bankMerchantId": "BANK456",
						},
					},
				},
			},
			expectedBankMID: "BANK456",
			expectedMIDTag:  "",
			wantError:       false,
		},
		{
			desc: "success with only merchant id tag",
			metadata: map[string]any{
				"paymentMethodOptions": map[string]any{
					"card": map[string]any{
						"processingConfig": map[string]any{
							"merchantIdTag": "tag2",
						},
					},
				},
			},
			expectedBankMID: "",
			expectedMIDTag:  "tag2",
			wantError:       false,
		},
		{
			desc: "success with no processing config",
			metadata: map[string]any{
				"paymentMethodOptions": map[string]any{
					"card": map[string]any{},
				},
			},
			expectedBankMID: "",
			expectedMIDTag:  "",
			wantError:       false,
		},
		{
			desc: "success with no card options",
			metadata: map[string]any{
				"paymentMethodOptions": map[string]any{},
			},
			expectedBankMID: "",
			expectedMIDTag:  "",
			wantError:       false,
		},
		{
			desc:            "success with empty metadata",
			metadata:        map[string]any{},
			expectedBankMID: "",
			expectedMIDTag:  "",
			wantError:       false,
		},
		{
			desc: "success with auto split payment",
			metadata: map[string]any{
				"autoSplitPayment": map[string]any{
					"transactionType": constant.AutoSplitPaymentTypeAuthentication,
					"processor":       "MPGS",
					"processorLimit":  float64(2000000000),
					"citMerchantId":   "CIT_MID_001",
					"mitMerchantId":   "MIT_MID_001",
				},
			},
			expectedAutoSplit: &creditcardModel.AutoSplitPayment{
				TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				Processor:       "MPGS",
				ProcessorLimit:  2000000000,
				CITMerchantID:   "CIT_MID_001",
				MITMerchantID:   "MIT_MID_001",
			},
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			referenceID := "REF123"
			payment := &Payment{
				UUID:            uuid.NewString(),
				Type:            "CHARGE",
				Status:          "COMPLETED",
				PaymentMethodID: "payment-method-id",
				Amount:          decimal.NewFromFloat(100000),
				Currency:        "IDR",
				ReferenceID:     &referenceID,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				Metadata:        &tc.metadata,
				PaymentMethod: PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			}

			result, err := payment.ToInternalCreditcardGetPaymentByUUID()

			if tc.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, result.Metadata)
			assert.NotNil(t, result.Metadata.ProcessingConfig)
			assert.Equal(t, tc.expectedBankMID, result.Metadata.ProcessingConfig.BankMerchantId)
			assert.Equal(t, tc.expectedMIDTag, result.Metadata.ProcessingConfig.MerchantIdTag)

			assert.Equal(t, tc.expectedAutoSplit, result.Metadata.AutoSplitPayment)
			if tc.expectedAutoSplit != nil {
				assert.NotNil(t, payment.AutoSplitPayment)
				assert.Equal(t, tc.expectedAutoSplit.TransactionType, payment.AutoSplitPayment.TransactionType)
				assert.Equal(t, tc.expectedAutoSplit.Processor, payment.AutoSplitPayment.Processor)
				assert.Equal(t, tc.expectedAutoSplit.ProcessorLimit, payment.AutoSplitPayment.ProcessorLimit)
				assert.Equal(t, tc.expectedAutoSplit.CITMerchantID, payment.AutoSplitPayment.CITMerchantID)
				assert.Equal(t, tc.expectedAutoSplit.MITMerchantID, payment.AutoSplitPayment.MITMerchantID)
			}
		})
	}
}

func TestPaymentIsFeeExempt(t *testing.T) {
	tests := []struct {
		payment         Payment
		wantIsFeeExempt bool
	}{
		{
			payment:         Payment{},
			wantIsFeeExempt: false,
		},
		{
			payment:         Payment{Type: constant.UnifiedPaymentTypeSingle},
			wantIsFeeExempt: false,
		},
		{
			payment:         Payment{Type: constant.UnifiedPaymentOneDollarAuthorization},
			wantIsFeeExempt: true,
		},
		{
			payment: Payment{
				RecurringPayment: &unifiedPaymentModel.MetadataRecurringPayment{
					InitiateFirstAuthorization: false,
				},
				Amount: decimal.NewFromInt(10_000),
			},
			wantIsFeeExempt: false,
		},
		{
			payment: Payment{
				RecurringPayment: &unifiedPaymentModel.MetadataRecurringPayment{
					InitiateFirstAuthorization: true,
					FirstAuthorizationMethod:   constant.RecurringContractAuthMethodFirstPayment,
				},
			},
			wantIsFeeExempt: false,
		},
		{
			payment: Payment{
				RecurringPayment: &unifiedPaymentModel.MetadataRecurringPayment{
					InitiateFirstAuthorization: true,
					FirstAuthorizationMethod:   constant.RecurringContractAuthMethodOneDollar,
				},
			},
			wantIsFeeExempt: true,
		},
		{
			payment: Payment{
				RecurringPayment: &unifiedPaymentModel.MetadataRecurringPayment{
					InitiateFirstAuthorization: false,
				},
				Amount: decimal.NewFromInt(0),
			},
			wantIsFeeExempt: true,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantIsFeeExempt, test.payment.IsFeeExempt())
	}
}

func TestExtractPaymentMetadata(t *testing.T) {
	tests := []struct {
		name     string
		payment  *Payment
		expected *unifiedPaymentModel.MetadataUnifiedPayment
	}{
		{
			name: "SUCCESS: Valid payment with VirtualTerminal",
			payment: &Payment{
				Metadata: &map[string]interface{}{
					"virtualTerminal": map[string]interface{}{
						"batchId":         "batch-123",
						"bookingId":       "booking-456",
						"travelAgentCode": "TAC-001",
						"travelAgentName": "Travel Agent ABC",
						"remarks":         "Test booking",
					},
					"isUnifiedPaymentV2": true,
					"isMigratingFromV1":  false,
					"autoConfirm":        true,
					"mode":               "redirect",
				},
			},
			expected: &unifiedPaymentModel.MetadataUnifiedPayment{
				IsUnifiedPaymentV2: true,
				IsMigratingFromV1:  false,
				AutoConfirm:        true,
				Mode:               "redirect",
				VirtualTerminal: &unifiedPaymentModel.VirtualTerminal{
					BatchID:         "batch-123",
					BookingID:       "booking-456",
					TravelAgentCode: "TAC-001",
					TravelAgentName: "Travel Agent ABC",
					Remarks:         "Test booking",
				},
			},
		},
		{
			name: "SUCCESS: Valid payment with all metadata fields",
			payment: &Payment{
				Metadata: &map[string]interface{}{
					"isUnifiedPaymentV2":  true,
					"isMigratingFromV1":   false,
					"autoConfirm":         false,
					"mode":                "api",
					"statementDescriptor": "Test Order",
					"clientRedirectUrl": map[string]interface{}{
						"successReturnUrl":    "https://example.com/success",
						"failureReturnUrl":    "https://example.com/failure",
						"expirationReturnUrl": "https://example.com/expiry",
					},
					"paymentMethod": map[string]interface{}{
						"type": "credit_card",
					},
					"clientMetadata": map[string]interface{}{
						"customField": "customValue",
					},
					"virtualTerminal": map[string]interface{}{
						"batchId":   "batch-789",
						"bookingId": "booking-101",
					},
				},
			},
			expected: &unifiedPaymentModel.MetadataUnifiedPayment{
				IsUnifiedPaymentV2:  true,
				IsMigratingFromV1:   false,
				AutoConfirm:         false,
				Mode:                "api",
				StatementDescriptor: "Test Order",
				ClientRedirectUrl: &unifiedPaymentModel.RedirectUrl{
					SuccessReturnUrl:    "https://example.com/success",
					FailureReturnUrl:    "https://example.com/failure",
					ExpirationReturnUrl: "https://example.com/expiry",
				},
				PaymentMethod: &unifiedPaymentModel.PaymentMethod{
					Type: "credit_card",
				},
				ClientMetadata: map[string]interface{}{
					"customField": "customValue",
				},
				VirtualTerminal: &unifiedPaymentModel.VirtualTerminal{
					BatchID:   "batch-789",
					BookingID: "booking-101",
				},
			},
		},
		{
			name: "SUCCESS: Payment with nil metadata returns empty struct",
			payment: &Payment{
				Metadata: nil,
			},
			expected: &unifiedPaymentModel.MetadataUnifiedPayment{},
		},
		{
			name: "SUCCESS: Payment with empty metadata",
			payment: &Payment{
				Metadata: &map[string]interface{}{},
			},
			expected: &unifiedPaymentModel.MetadataUnifiedPayment{},
		},
		{
			name: "SUCCESS: Payment with only VirtualTerminal field",
			payment: &Payment{
				Metadata: &map[string]interface{}{
					"virtualTerminal": map[string]interface{}{
						"batchId": "batch-only",
					},
				},
			},
			expected: &unifiedPaymentModel.MetadataUnifiedPayment{
				VirtualTerminal: &unifiedPaymentModel.VirtualTerminal{
					BatchID: "batch-only",
				},
			},
		},
		{
			name: "SUCCESS: Payment with VirtualTerminal containing all fields",
			payment: &Payment{
				Metadata: &map[string]interface{}{
					"virtualTerminal": map[string]interface{}{
						"batchId":         "batch-12345",
						"bookingId":       "booking-67890",
						"travelAgentCode": "TA-999",
						"travelAgentName": "Premium Travel",
						"remarks":         "Urgent booking - VIP client",
					},
				},
			},
			expected: func() *unifiedPaymentModel.MetadataUnifiedPayment {
				return &unifiedPaymentModel.MetadataUnifiedPayment{
					VirtualTerminal: &unifiedPaymentModel.VirtualTerminal{
						BatchID:         "batch-12345",
						BookingID:       "booking-67890",
						TravelAgentCode: "TA-999",
						TravelAgentName: "Premium Travel",
						Remarks:         "Urgent booking - VIP client",
					},
				}
			}(),
		},
		{
			name: "SUCCESS: Payment with complex nested metadata",
			payment: &Payment{
				Metadata: &map[string]interface{}{
					"paymentMethodOptions": map[string]interface{}{
						"card": map[string]interface{}{
							"captureMethod": "auto",
							"threeDsMethod": "3ds",
						},
					},
					"virtualTerminal": map[string]interface{}{
						"batchId":   "complex-batch",
						"bookingId": "complex-booking",
						"remarks":   "Complex test case",
					},
					"isUnifiedPaymentV2": true,
					"saveForFutureUse":   true,
					"showSavedPayment":   false,
				},
			},
			expected: func() *unifiedPaymentModel.MetadataUnifiedPayment {
				saveForFutureUse := true
				showSavedPayment := false
				return &unifiedPaymentModel.MetadataUnifiedPayment{
					IsUnifiedPaymentV2: true,
					SaveForFutureUse:   &saveForFutureUse,
					ShowSavedPayment:   &showSavedPayment,
					VirtualTerminal: &unifiedPaymentModel.VirtualTerminal{
						BatchID:   "complex-batch",
						BookingID: "complex-booking",
						Remarks:   "Complex test case",
					},
				}
			}(),
		},
		{
			name: "SUCCESS: Payment with boolean fields",
			payment: &Payment{
				Metadata: &map[string]interface{}{
					"isSnap":             true,
					"isUnifiedPaymentV2": true,
					"autoConfirm":        true,
					"saveForFutureUse":   false,
					"showSavedPayment":   true,
				},
			},
			expected: func() *unifiedPaymentModel.MetadataUnifiedPayment {
				isSnap := true
				saveForFutureUse := false
				showSavedPayment := true
				return &unifiedPaymentModel.MetadataUnifiedPayment{
					IsSnap:             &isSnap,
					IsUnifiedPaymentV2: true,
					AutoConfirm:        true,
					SaveForFutureUse:   &saveForFutureUse,
					ShowSavedPayment:   &showSavedPayment,
				}
			}(),
		},
		{
			name: "SUCCESS: Payment with time fields",
			payment: &Payment{
				Metadata: &map[string]interface{}{
					"canceledAt":         "2024-01-15T10:30:00Z",
					"isUnifiedPaymentV2": true,
				},
			},
			expected: func() *unifiedPaymentModel.MetadataUnifiedPayment {
				canceledAt, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
				return &unifiedPaymentModel.MetadataUnifiedPayment{
					IsUnifiedPaymentV2: true,
					CanceledAt:         &canceledAt,
				}
			}(),
		},
		{
			name: "SUCCESS: Payment with cancellation reason",
			payment: &Payment{
				Metadata: &map[string]interface{}{
					"cancellationReason": "User cancelled the payment",
					"isUnifiedPaymentV2": true,
				},
			},
			expected: &unifiedPaymentModel.MetadataUnifiedPayment{
				IsUnifiedPaymentV2: true,
				CancellationReason: "User cancelled the payment",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function (needs to be accessed via the package since it's internal)
			result := tt.payment.ToUnifiedPaymentMetadata()

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				// Compare the main struct fields
				assert.Equal(t, tt.expected.IsUnifiedPaymentV2, result.IsUnifiedPaymentV2)
				assert.Equal(t, tt.expected.IsMigratingFromV1, result.IsMigratingFromV1)
				assert.Equal(t, tt.expected.AutoConfirm, result.AutoConfirm)
				assert.Equal(t, tt.expected.Mode, result.Mode)
				assert.Equal(t, tt.expected.StatementDescriptor, result.StatementDescriptor)
				assert.Equal(t, tt.expected.CancellationReason, result.CancellationReason)

				// Compare VirtualTerminal
				if tt.expected.VirtualTerminal != nil {
					assert.NotNil(t, result.VirtualTerminal)
					assert.Equal(t, tt.expected.VirtualTerminal.BatchID, result.VirtualTerminal.BatchID)
					assert.Equal(t, tt.expected.VirtualTerminal.BookingID, result.VirtualTerminal.BookingID)
					assert.Equal(t, tt.expected.VirtualTerminal.TravelAgentCode, result.VirtualTerminal.TravelAgentCode)
					assert.Equal(t, tt.expected.VirtualTerminal.TravelAgentName, result.VirtualTerminal.TravelAgentName)
					assert.Equal(t, tt.expected.VirtualTerminal.Remarks, result.VirtualTerminal.Remarks)
				}

				// Compare boolean pointers
				if tt.expected.IsSnap != nil {
					assert.NotNil(t, result.IsSnap)
					assert.Equal(t, *tt.expected.IsSnap, *result.IsSnap)
				}
				if tt.expected.SaveForFutureUse != nil {
					assert.NotNil(t, result.SaveForFutureUse)
					assert.Equal(t, *tt.expected.SaveForFutureUse, *result.SaveForFutureUse)
				}
				if tt.expected.ShowSavedPayment != nil {
					assert.NotNil(t, result.ShowSavedPayment)
					assert.Equal(t, *tt.expected.ShowSavedPayment, *result.ShowSavedPayment)
				}

				// Compare time pointers
				if tt.expected.CanceledAt != nil {
					assert.NotNil(t, result.CanceledAt)
					assert.Equal(t, tt.expected.CanceledAt.Unix(), result.CanceledAt.Unix())
				}
			}
		})
	}
}

func TestToUnifiedPaymentResponse_PaymentUrlConditions(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name        string
		mode        string
		paymentType string
		acquirer    string
		initialURL  string
		expectedURL string
	}{
		{
			name:        "API mode, not Ewallet type, not DANA acquirer -> URL cleared",
			mode:        constant.UnifiedPaymentModeAPI,
			paymentType: constant.UnifiedPaymentMethodEWallet,
			acquirer:    constant.UnifiedPaymentEWalletShopeePayAcquirer,
			initialURL:  "https://payment.url",
			expectedURL: "",
		},
		{
			name:        "API mode, is Ewallet type, not DANA acquirer -> URL cleared",
			mode:        constant.UnifiedPaymentModeAPI,
			paymentType: constant.UnifiedPaymentMethodEWallet,
			acquirer:    constant.UnifiedPaymentEWalletShopeePayAcquirer,
			initialURL:  "https://payment.url",
			expectedURL: "",
		},
		{
			name:        "API mode, not Ewallet type, is DANA acquirer -> URL cleared",
			mode:        constant.UnifiedPaymentModeAPI,
			paymentType: constant.UnifiedPaymentMethodQris,
			acquirer:    constant.UnifiedPaymentEWalletDanaAcquirer,
			initialURL:  "https://payment.url",
			expectedURL: "",
		},
		{
			name:        "Redirect mode, not PAYMENT type, not DANA acquirer -> URL kept",
			mode:        constant.UnifiedPaymentModeRedirect,
			paymentType: constant.UnifiedPaymentMethodQris,
			acquirer:    constant.UnifiedPaymentEWalletDanaAcquirer,
			initialURL:  "https://payment.url",
			expectedURL: "https://payment.url",
		},
		{
			name:        "API mode, is Ewallet type, DANA acquirer -> URL kept",
			mode:        constant.UnifiedPaymentModeAPI,
			paymentType: constant.UnifiedPaymentMethodEWallet,
			acquirer:    constant.UnifiedPaymentEWalletDanaAcquirer,
			initialURL:  "https://payment.url",
			expectedURL: "https://payment.url",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metadata := map[string]any{
				"mode": tc.mode,
			}

			payment := &Payment{
				UUID:       "test-uuid",
				Currency:   "IDR",
				Amount:     decimal.NewFromFloat(1000),
				Status:     "pending",
				Metadata:   &metadata,
				PaymentURL: tc.initialURL,
				CreatedAt:  now,
				UpdatedAt:  now,
				PaymentMethod: PaymentMethod{
					Type:     tc.paymentType,
					Acquirer: tc.acquirer,
				},
			}

			resp := payment.ToUnifiedPaymentResponse()
			assert.Equal(t, tc.expectedURL, resp.PaymentUrl)
		})
	}
}

func TestPaymentGetGroupPaymentType(t *testing.T) {
	tests := []struct {
		name       string
		input      Payment
		wantResult string
	}{
		{
			name:       "Normal payment",
			input:      Payment{},
			wantResult: constant.GroupPaymentTypePayment,
		},
		{
			name: "Single payment",
			input: Payment{
				Type: constant.UnifiedPaymentTypeSingle,
			},
			wantResult: constant.GroupPaymentTypePayment,
		},
		{
			name: "Multiple payment",
			input: Payment{
				Type: constant.UnifiedPaymentTypeMultiple,
			},
			wantResult: constant.GroupPaymentTypePayment,
		},
		{
			name: "Recurring payment",
			input: Payment{
				RecurringContractID: util.ValueToPtr("1234"),
			},
			wantResult: constant.GroupPaymentTypeRecurringPayment,
		},
		{
			name: "Virtual terminal",
			input: Payment{
				Type: constant.PaymentTypeVirtualTerminal,
			},
			wantResult: constant.GroupPaymentTypeVirtualTerminal,
		},
		{
			name: "One dollar authorization",
			input: Payment{
				Type: constant.PaymentTypeOneDollarAuth,
			},
			wantResult: constant.GroupPaymentTypeOneDollarAuth,
		},
		{
			name: "Card-funded payout",
			input: Payment{
				Type: constant.PaymentTypeCardFundedPayout,
			},
			wantResult: constant.GroupPaymentTypeCardFundedPayout,
		},
		{
			name: "Split payment",
			input: Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				},
			},
			wantResult: constant.GroupPaymentTypeSplitPayment,
		},
		{
			name: "Split payment takes precedence over card-funded payout type",
			input: Payment{
				Type: constant.PaymentTypeCardFundedPayout,
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				},
			},
			wantResult: constant.GroupPaymentTypeSplitPayment,
		},
		{
			name: "Recurring payment takes precedence over split payment",
			input: Payment{
				RecurringContractID: util.ValueToPtr("1234"),
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				},
			},
			wantResult: constant.GroupPaymentTypeRecurringPayment,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, tt.input.GetGroupPaymentType())
		})
	}
}

func TestPaymentIsAutoSplitPaymentAuth(t *testing.T) {
	tests := []struct {
		name       string
		input      Payment
		wantResult bool
	}{
		{
			name:       "should return false when AutoSplitPayment is nil",
			input:      Payment{},
			wantResult: false,
		},
		{
			name: "should return false when TransactionType is not AUTHENTICATION",
			input: Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: "CAPTURE",
				},
			},
			wantResult: false,
		},
		{
			name: "should return true when TransactionType is AUTHENTICATION",
			input: Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				},
			},
			wantResult: true,
		},
		{
			name: "should return true with full AutoSplitPayment config",
			input: Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
					Processor:       "MPGS",
					ProcessorLimit:  2000000000,
					CITMerchantID:   "CIT_MID",
					MITMerchantID:   "MIT_MID",
				},
			},
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, tt.input.IsAutoSplitPaymentAuth())
		})
	}
}

func TestPaymentGetAutoSplitTotalSuccessAmount(t *testing.T) {
	expectedAmount := commonModel.Amount{Currency: "IDR", Value: "100000.00"}

	tests := []struct {
		name             string
		input            *Payment
		wantResult       *commonModel.Amount
		wantAutoSplitSet bool
	}{
		{
			name: "should return nil when AutoSplitPayment is nil and Metadata has no autoSplitPayment key",
			input: &Payment{
				Metadata: &map[string]any{},
			},
			wantResult: nil,
		},
		{
			name: "should return nil when AutoSplitPayment is nil and Metadata autoSplitPayment is not AutoSplitPayment type",
			input: &Payment{
				Metadata: &map[string]any{
					"autoSplitPayment": "invalid",
				},
			},
			wantResult: nil,
		},
		{
			name: "should return nil when AutoSplitPayment is nil and Metadata autoSplitPayment is nil",
			input: &Payment{
				Metadata: &map[string]any{
					"autoSplitPayment": (*unifiedPaymentModel.AutoSplitPayment)(nil),
				},
			},
			wantResult: nil,
		},
		{
			name: "should return nil when AutoSplitPayment is nil and Metadata autoSplitPayment is non-auth type",
			input: &Payment{
				Metadata: &map[string]any{
					"autoSplitPayment": &unifiedPaymentModel.AutoSplitPayment{
						TransactionType: constant.AutoSplitPaymentTypeFirstPayment,
						Summary: &unifiedPaymentModel.AutoSplitPaymentSummary{
							TotalSuccessfulChargeAmount: expectedAmount,
						},
					},
				},
			},
			wantResult:       nil,
			wantAutoSplitSet: true,
		},
		{
			name: "should populate AutoSplitPayment from Metadata and return amount when auth type with Summary",
			input: &Payment{
				Metadata: &map[string]any{
					"autoSplitPayment": &unifiedPaymentModel.AutoSplitPayment{
						TransactionType: constant.AutoSplitPaymentTypeAuthentication,
						Summary: &unifiedPaymentModel.AutoSplitPaymentSummary{
							TotalSuccessfulChargeAmount: expectedAmount,
						},
					},
				},
			},
			wantResult:       &expectedAmount,
			wantAutoSplitSet: true,
		},
		{
			name: "should return nil when AutoSplitPayment is auth but Summary is nil",
			input: &Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				},
			},
			wantResult: nil,
		},
		{
			name: "should return nil when AutoSplitPayment is not auth type",
			input: &Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeFirstPayment,
					Summary: &unifiedPaymentModel.AutoSplitPaymentSummary{
						TotalSuccessfulChargeAmount: expectedAmount,
					},
				},
			},
			wantResult: nil,
		},
		{
			name: "should return amount pointer when AutoSplitPayment is auth with Summary",
			input: &Payment{
				AutoSplitPayment: &unifiedPaymentModel.AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
					Summary: &unifiedPaymentModel.AutoSplitPaymentSummary{
						TotalSuccessfulChargeAmount: expectedAmount,
					},
				},
			},
			wantResult: &expectedAmount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.GetAutoSplitTotalSuccessAmount()
			assert.Equal(t, tt.wantResult, result)
			if tt.wantAutoSplitSet {
				assert.NotNil(t, tt.input.AutoSplitPayment)
			}
		})
	}
}

func TestPaymentSetStatusByAutoSplitStatus(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  string
		autoSplitStat  string
		expectedStatus string
	}{
		{
			name:           "success auto-split status should set to PAID",
			initialStatus:  constant.UnifiedPaymentSessionStatusProcessing,
			autoSplitStat:  constant.AutoSplitPaymentStatusSuccess,
			expectedStatus: constant.UnifiedPaymentSessionStatusPaid,
		},
		{
			name:           "partial success auto-split status should set to PAID",
			initialStatus:  constant.UnifiedPaymentSessionStatusProcessing,
			autoSplitStat:  constant.AutoSplitPaymentStatusPartialSuccess,
			expectedStatus: constant.UnifiedPaymentSessionStatusPaid,
		},
		{
			name:           "cancelled auto-split status should set to CANCELLED",
			initialStatus:  constant.UnifiedPaymentSessionStatusProcessing,
			autoSplitStat:  constant.AutoSplitPaymentStatusCancelled,
			expectedStatus: constant.UnifiedPaymentSessionStatusCancelled,
		},
		{
			name:           "failed auto-split status should set to CANCELLED",
			initialStatus:  constant.UnifiedPaymentSessionStatusProcessing,
			autoSplitStat:  constant.AutoSplitPaymentStatusFailed,
			expectedStatus: constant.UnifiedPaymentSessionStatusCancelled,
		},
		{
			name:           "processing auto-split status should not change status",
			initialStatus:  constant.UnifiedPaymentSessionStatusProcessing,
			autoSplitStat:  constant.AutoSplitPaymentStatusProcessing,
			expectedStatus: constant.UnifiedPaymentSessionStatusProcessing,
		},
		{
			name:           "unknown auto-split status should not change status",
			initialStatus:  "PENDING",
			autoSplitStat:  "UNKNOWN",
			expectedStatus: "PENDING",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Payment{Status: tt.initialStatus}
			p.SetStatusByAutoSplitStatus(tt.autoSplitStat)
			assert.Equal(t, tt.expectedStatus, p.Status)
		})
	}
}

func TestPaymentGetLedgerStatus(t *testing.T) {
	tests := []struct {
		name             string
		paymentStatus    string
		expectedLedgerSt string
	}{
		{
			name:             "PAID status should return SUCCESS",
			paymentStatus:    constant.UnifiedPaymentSessionStatusPaid,
			expectedLedgerSt: constant.StatusSuccess,
		},
		{
			name:             "SUCCESS status should return SUCCESS",
			paymentStatus:    constant.StatusSuccess,
			expectedLedgerSt: constant.StatusSuccess,
		},
		{
			name:             "CANCELLED status should return FAILED",
			paymentStatus:    constant.UnifiedPaymentSessionStatusCancelled,
			expectedLedgerSt: constant.StatusFailed,
		},
		{
			name:             "FAILED status should return FAILED",
			paymentStatus:    constant.StatusFailed,
			expectedLedgerSt: constant.StatusFailed,
		},
		{
			name:             "EXPIRED status should return FAILED",
			paymentStatus:    constant.UnifiedPaymentSessionStatusExpired,
			expectedLedgerSt: constant.StatusFailed,
		},
		{
			name:             "PROCESSING status should return PENDING",
			paymentStatus:    constant.UnifiedPaymentSessionStatusProcessing,
			expectedLedgerSt: constant.StatusPending,
		},
		{
			name:             "REQUIRE_ACTION status should return PENDING",
			paymentStatus:    constant.UnifiedPaymentSessionStatusRequireAction,
			expectedLedgerSt: constant.StatusPending,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Payment{Status: tt.paymentStatus}
			assert.Equal(t, tt.expectedLedgerSt, p.GetLedgerStatus())
		})
	}
}

func TestPaymentIsFinalStatus(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		wantResult bool
	}{
		{
			name:       "PAID is final status",
			status:     constant.UnifiedPaymentSessionStatusPaid,
			wantResult: true,
		},
		{
			name:       "CANCELLED is final status",
			status:     constant.UnifiedPaymentSessionStatusCancelled,
			wantResult: true,
		},
		{
			name:       "EXPIRED unified payment is final status",
			status:     constant.UnifiedPaymentSessionStatusExpired,
			wantResult: true,
		},
		{
			name:       "SUCCESS payment is final status",
			status:     paymentConstant.PAYMENT_STATUS_SUCCESS,
			wantResult: true,
		},
		{
			name:       "FAILED payment is final status",
			status:     paymentConstant.PaymentStatusFailed,
			wantResult: true,
		},
		{
			name:       "EXPIRED payment is final status",
			status:     paymentConstant.PaymentStatusExpired,
			wantResult: true,
		},
		{
			name:       "PROCESSING is not final status",
			status:     constant.UnifiedPaymentSessionStatusProcessing,
			wantResult: false,
		},
		{
			name:       "REQUIRE_PAYMENT_METHOD is not final status",
			status:     constant.UnifiedPaymentSessionStatusRequirePaymentMethod,
			wantResult: false,
		},
		{
			name:       "REQUIRE_ACTION is not final status",
			status:     constant.UnifiedPaymentSessionStatusRequireAction,
			wantResult: false,
		},
		{
			name:       "empty status is not final status",
			status:     "",
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Payment{Status: tt.status}
			assert.Equal(t, tt.wantResult, p.IsFinalStatus())
		})
	}
}
