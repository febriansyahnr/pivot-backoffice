package refundModel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildChannelDestination(t *testing.T) {
	tests := []struct {
		name            string
		response        *RefundResponse
		expectedMethod  string
		expectedChannel string
		expectedDetail  map[string]interface{}
	}{
		{
			name: "should not build when destinationType is not CHANNEL",
			response: &RefundResponse{
				DestinationType:         "ACCOUNT",
				RawChargeAdditionalInfo: map[string]interface{}{},
			},
			expectedMethod: "",
		},
		{
			name: "should build QRIS channel destination with RRN",
			response: &RefundResponse{
				DestinationType: "CHANNEL",
				PaymentChannel:  "BNC",
				RawChargeAdditionalInfo: map[string]interface{}{
					"methodDetail": map[string]interface{}{
						"qr": map[string]interface{}{
							"acquirer":                 "BNC",
							"merchantName":             "Test Merchant",
							"retrievalReferenceNumber": "2026011414355650144231551",
						},
					},
				},
			},
			expectedMethod:  "QRIS",
			expectedChannel: "BNC",
			expectedDetail: map[string]interface{}{
				"acquirer":     "BNC",
				"merchantName": "Test Merchant",
				"rrn":          "2026011414355650144231551",
			},
		},
		{
			name: "should build CREDIT_CARD channel destination with RRN",
			response: &RefundResponse{
				DestinationType: "CHANNEL",
				PaymentChannel:  "VISA",
				RawChargeAdditionalInfo: map[string]interface{}{
					"methodDetail": map[string]interface{}{
						"card": map[string]interface{}{
							"last4": "1234",
							"binInformations": map[string]interface{}{
								"issuingBank": "BCA",
								"brand":       "VISA",
							},
							"authorizationResult": map[string]interface{}{
								"retrievalReferenceNumber": "TRXCC8fb53668d1c717458198311",
							},
						},
					},
				},
			},
			expectedMethod:  "CREDIT_CARD",
			expectedChannel: "VISA",
			expectedDetail: map[string]interface{}{
				"last4Digit":  "1234",
				"cardIssuing": "BCA",
				"cardBrand":   "VISA",
				"rrn":         "TRXCC8fb53668d1c717458198311",
			},
		},
		{
			name: "should build EWALLET channel destination",
			response: &RefundResponse{
				DestinationType: "CHANNEL",
				PaymentChannel:  "SHOPEEPAY",
				RawChargeAdditionalInfo: map[string]interface{}{
					"methodDetail": map[string]interface{}{
						"ewallet": map[string]interface{}{
							"channel": "SHOPEEPAY",
						},
					},
				},
			},
			expectedMethod:  "EWALLET",
			expectedChannel: "SHOPEEPAY",
			expectedDetail: map[string]interface{}{
				"channel": "SHOPEEPAY",
			},
		},
		{
			name: "should handle rawChargeAdditionalInfo as JSON string",
			response: &RefundResponse{
				DestinationType:         "CHANNEL",
				PaymentChannel:          "BNC",
				RawChargeAdditionalInfo: `{"methodDetail":{"qr":{"acquirer":"BNC","retrievalReferenceNumber":"RRN123456"}}}`,
			},
			expectedMethod:  "QRIS",
			expectedChannel: "BNC",
			expectedDetail: map[string]interface{}{
				"acquirer": "BNC",
				"rrn":      "RRN123456",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.response.BuildChannelDestination()

			if tc.expectedMethod == "" {
				assert.Nil(t, tc.response.ChannelDestination)
				return
			}

			assert.NotNil(t, tc.response.ChannelDestination)
			assert.Equal(t, tc.expectedMethod, tc.response.ChannelDestination.PaymentMethod)
			assert.Equal(t, tc.expectedChannel, tc.response.ChannelDestination.PaymentChannel)
			for key, expected := range tc.expectedDetail {
				assert.Equal(t, expected, tc.response.ChannelDestination.PaymentDetail[key])
			}
		})
	}
}
