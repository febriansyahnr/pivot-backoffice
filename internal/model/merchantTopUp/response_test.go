package merchantTopUp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToResponse(t *testing.T) {
	data := &MerchantTopUp{
		ID:              "uuid-uuid-uuid",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		ReferenceNumber: "reference-number",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	response := &MerchantTopUpResponse{
		UUID:            "uuid-uuid-uuid",
		MerchantID:      "merchant-id",
		PaymentMethodID: "payment-method-id",
		ReferenceNumber: "reference-number",
		CreatedAt:       data.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       data.UpdatedAt.Format(time.RFC3339),
	}

	testCases := []struct {
		Name     string
		Input    *MerchantTopUp
		Expected *MerchantTopUpResponse
	}{
		{
			Name:     "it should return response",
			Input:    data,
			Expected: response,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newResponse := tc.Input.ToResponse()
			require.Equal(t, tc.Expected, newResponse)
		})
	}
}
