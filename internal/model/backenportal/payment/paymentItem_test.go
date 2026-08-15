package paymentModel

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPaymentItem_PaymentItemFromDTO(t *testing.T) {
	amount := decimal.NewFromFloat(1000)
	totalAmount := decimal.NewFromFloat(1000)
	now := time.Now()
	metadata := "{\"testing\":\"testing\"}"
	metadataMap := map[string]any{"testing": "testing"}

	paymentItemDTO := &PaymentItemDTO{
		UUID:        "uuid-uuid-uuid",
		PaymentID:   "payment-id",
		Name:        "name",
		Description: "",
		Qty:         1,
		Currency:    "IDR",
		Amount:      amount,
		TotalAmount: totalAmount,
		Metadata:    nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	paymentItem := &PaymentItem{
		UUID:        "uuid-uuid-uuid",
		PaymentID:   "payment-id",
		Name:        "name",
		Description: "",
		Qty:         1,
		Currency:    "IDR",
		Amount:      amount,
		TotalAmount: totalAmount,
		Metadata:    &metadataMap,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	testCases := []struct {
		Name     string
		Input    *PaymentItemDTO
		Metadata *map[string]interface{}
		Expected *PaymentItem
	}{
		{
			Name: "it should return payment item",
			Input: &PaymentItemDTO{
				UUID:        "uuid-uuid-uuid",
				PaymentID:   "payment-id",
				Name:        "name",
				Description: "",
				Qty:         1,
				Currency:    "IDR",
				Amount:      amount,
				TotalAmount: totalAmount,
				Metadata:    &metadata,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			Expected: &PaymentItem{
				UUID:        "uuid-uuid-uuid",
				PaymentID:   "payment-id",
				Name:        "name",
				Description: "",
				Qty:         1,
				Currency:    "IDR",
				Amount:      amount,
				TotalAmount: totalAmount,
				Metadata:    &metadataMap,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		{
			Name:     "it should return payment item with nil metadata",
			Input:    paymentItemDTO,
			Expected: paymentItem,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Expected.PaymentItemFromDTO(tc.Input)
			require.Equal(t, tc.Expected, paymentItem)
		})
	}
}

func TestPaymentItemDTO_ToPaymentItem(t *testing.T) {
	amount := decimal.NewFromFloat(1000)
	totalAmount := decimal.NewFromFloat(1000)
	now := time.Now()

	paymentItemDTO := &PaymentItemDTO{
		UUID:        "uuid-uuid-uuid",
		PaymentID:   "payment-id",
		Name:        "name",
		Description: "",
		Qty:         1,
		Currency:    "IDR",
		Amount:      amount,
		TotalAmount: totalAmount,
		Metadata:    nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	paymentItem := &PaymentItem{
		UUID:        "uuid-uuid-uuid",
		PaymentID:   "payment-id",
		Name:        "name",
		Description: "",
		Qty:         1,
		Currency:    "IDR",
		Amount:      amount,
		TotalAmount: totalAmount,
		Metadata:    nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	testCases := []struct {
		Name     string
		Input    *PaymentItemDTO
		Expected *PaymentItem
	}{
		{
			Name:     "it should return payment item",
			Input:    paymentItemDTO,
			Expected: paymentItem,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newPaymentItem := tc.Input.ToPaymentItem()
			require.Equal(t, tc.Expected, newPaymentItem)
		})
	}
}

func TestPaymentItem_ToPaymentResponseItem(t *testing.T) {
	amount := decimal.NewFromFloat(1000)
	now := time.Now()
	description := "description"

	paymentItem := &PaymentItem{
		UUID:        "uuid-uuid-uuid",
		PaymentID:   "payment-id",
		Name:        "name",
		Description: description,
		Qty:         1,
		Currency:    "IDR",
		Amount:      amount,
		TotalAmount: amount,
		Metadata:    nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	paymentResponseItem := &PaymentResponseItem{
		ItemID:      "uuid-uuid-uuid",
		Name:        "name",
		Description: "description",
		Amount: Amount{
			Value:    amount,
			Currency: "IDR",
		},
		Qty: 1,
	}

	testCases := []struct {
		Name     string
		Input    *PaymentItem
		Expected *PaymentResponseItem
	}{
		{
			Name:     "it should return payment response item",
			Input:    paymentItem,
			Expected: paymentResponseItem,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newPaymentResponseItem := tc.Input.ToPaymentResponseItem()
			require.Equal(t, tc.Expected, newPaymentResponseItem)
		})
	}
}
