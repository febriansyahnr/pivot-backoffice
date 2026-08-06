package paymentModel_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBookingPayloadToResponse(t *testing.T) {
	booking := BookingPayload{
		BookingID:       "booking-001",
		ReferenceID:     "ref-001",
		TravelAgentCode: "TRAVELOKA",
		Amount: Amount{
			Value:    decimal.NewFromFloat(1500000),
			Currency: "IDR",
		},
		Card: Card{
			Number:       "4111111111111111",
			SecurityCode: "123",
			Expiry:       Expiry{Month: "12", Year: "25"},
		},
		MerchantID:     "merchant-uuid",
		BankMerchantID: "bank-merchant-uuid",
		BatchID:        "batch-001",
		PaymentID:      "payment-uuid",
		CreatedBy:      "user-uuid",
	}

	originalNumber := booking.Card.Number
	originalSecurityCode := booking.Card.SecurityCode

	wantResponse := booking
	wantResponse.Card.Number = "411111xxxxxx1111"
	wantResponse.Card.SecurityCode = "xxx"

	response := booking.ToResponse()

	assert.Equal(t, wantResponse, response)
	assert.Equal(t, originalNumber, booking.Card.Number)
	assert.Equal(t, originalSecurityCode, booking.Card.SecurityCode)
}

func TestBookingPayloadToCardChargePayload(t *testing.T) {
	booking := BookingPayload{
		ReferenceID: "ref-001",
		Amount: Amount{
			Value:    decimal.NewFromFloat(1500000),
			Currency: "IDR",
		},
		Card: Card{
			Number:       "4111111111111111",
			SecurityCode: "123",
			Expiry:       Expiry{Month: "12", Year: "25"},
		},
		MerchantID: "merchant-uuid",
		PaymentID:  "payment-uuid",
	}

	wantPayload := CardChargePayload{
		PaymentID:           "payment-uuid",
		MerchantID:          "merchant-uuid",
		ClientTransactionID: "ref-001",
		Amount:              1500000,
		Currency:            "IDR",
		Card: CardChargePayloadCHD{
			Number:       "4111111111111111",
			SecurityCode: "123",
			Expiry:       Expiry{Month: "12", Year: "25"},
		},
		VCCTerminal: true,
	}
	assert.Equal(t, wantPayload, booking.ToCardChargePayload())
}

func TestBookingPayloadToVCCTerminalChargeMessage(t *testing.T) {
	booking := BookingPayload{
		BookingID:       "booking-001", // NOSONAR
		ReferenceID:     "ref-001",     // NOSONAR
		TravelAgentCode: "TRAVELOKA",   // NOSONAR
		Amount: Amount{
			Value:    decimal.NewFromFloat(1500000),
			Currency: "IDR", // NOSONAR
		},
		MerchantID:     "merchant-uuid",      // NOSONAR
		BankMerchantID: "bank-merchant-uuid", // NOSONAR
		BatchID:        "batch-001",          // NOSONAR
		PaymentID:      "payment-uuid",       // NOSONAR
		CreatedBy:      "user-uuid",          // NOSONAR
	}
	encryptedPayload := "encrypted-payload-abc123"

	wantPayload := VCCTerminalChargeMessage{
		MerchantID:       "merchant-uuid", // NOSONAR
		BatchID:          "batch-001",     // NOSONAR
		PaymentID:        "payment-uuid",  // NOSONAR
		EncryptedPayload: encryptedPayload,
	}

	payload := booking.ToVCCTerminalChargeMessage(encryptedPayload)

	assert.Equal(t, wantPayload, payload)
}

func TestBookingPayloadValidation(t *testing.T) {
	vld := validatorExt.New()
	tests := []struct {
		request         BookingPayload
		wantErrContains string
	}{
		{
			request: BookingPayload{
				BookingID:   "BOOK001", // NOSONAR
				ReferenceID: "REF0001", // NOSONAR
				Amount: Amount{
					Value:    decimal.NewFromInt(10_000), // NOSONAR
					Currency: constant.CurrencyIDR,
				},
				TravelAgentCode: "TRAVELOKA", // NOSONAR
				Card: Card{
					Number: "4440000042200014", // NOSONAR
					Expiry: Expiry{
						Month: "02", // NOSONAR
						Year:  "30", // NOSONAR
					},
					SecurityCode: "123", // NOSONAR
				},
				Remarks: "12345678901234567890", // NOSONAR
			},
			wantErrContains: "",
		},
		{
			request: BookingPayload{
				BookingID:   "BOOK002", // NOSONAR
				ReferenceID: "REF0002", // NOSONAR
				Amount: Amount{
					Value:    decimal.NewFromInt(10_000), // NOSONAR
					Currency: constant.CurrencyIDR,
				},
				TravelAgentCode: "TRAVELOKA", // NOSONAR
				Card: Card{
					Number: "4440000042200014", // NOSONAR
					Expiry: Expiry{
						Month: "02", // NOSONAR
						Year:  "30", // NOSONAR
					},
					SecurityCode: "1", // NOSONAR
				},
			},
			wantErrContains: "Error:Field validation for 'SecurityCode' failed on the 'min' tag", // NOSONAR
		},
		{
			request: BookingPayload{
				BookingID:   "BOOK003", // NOSONAR
				ReferenceID: "REF0003", // NOSONAR
				Amount: Amount{
					Value:    decimal.NewFromInt(10_000), // NOSONAR
					Currency: constant.CurrencyIDR,
				},
				TravelAgentCode: "TRAVELOKA", // NOSONAR
				Card: Card{
					Number: "4440000042200014", // NOSONAR
					Expiry: Expiry{
						Month: "02", // NOSONAR
						Year:  "30", // NOSONAR
					},
					SecurityCode: "223", // NOSONAR
				},
				Remarks: "123456789012345678900", // NOSONAR
			},
			wantErrContains: "Error:Field validation for 'Remarks' failed on the 'maxChar' tag", // NOSONAR
		},
	}
	for _, test := range tests {

		err := vld.Struct(test.request)
		if test.wantErrContains == "" {
			assert.NoError(t, err)
		} else {
			assert.ErrorContains(t, err, test.wantErrContains)
		}
	}
}

func TestBookingPayloadToCreateUnifiedPaymentSessionRequest(t *testing.T) {
	booking := BookingPayload{
		BookingID:   "BOOK001", // NOSONAR
		ReferenceID: "REF0001", // NOSONAR
		Amount: Amount{
			Value:    decimal.NewFromInt(10_000), // NOSONAR
			Currency: constant.CurrencyIDR,
		},
		TravelAgentCode: "TRAVELOKA", // NOSONAR
		Card: Card{
			Number: "4440000042200014", // NOSONAR
			Expiry: Expiry{
				Month: "02", // NOSONAR
				Year:  "30", // NOSONAR
			},
			SecurityCode: "123", // NOSONAR
		},
		Remarks:        "test",    // NOSONAR
		BankMerchantID: "TEST001", // NOSONAR
		MerchantID:     "87c84777-217e-49bb-92f1-986e9c2235f3",
		BatchID:        "47daf0d9-5f2b-4db0-8bc7-893833145dab",
		PaymentID:      "36aa12f4-b9bd-46c9-aae4-9a5c315e1f16",
	}

	wantResult := &unifiedPaymentModel.CreateUnifiedPaymentSessionRequest{
		ClientReferenceID: "REF0001", // NOSONAR
		PaymentMethod: &unifiedPaymentModel.PaymentMethod{
			Type: constant.UnifiedPaymentMethodCard,
		},
		PaymentMethodOptions: unifiedPaymentModel.PaymentMethodOptions{
			Card: &unifiedPaymentModel.PaymentMethodOptionCard{
				ThreeDsMethod: constant.CardThreeDsMethodNever,
				ProcessingConfig: &unifiedPaymentModel.PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "TEST001", // NOSONAR
				},
			},
		},
		Mode:        constant.UnifiedPaymentModeRedirect,
		AutoConfirm: true,
		Amount: unifiedPaymentModel.Amount{
			Value:    10_000, // NOSONAR
			Currency: constant.CurrencyIDR,
		},
		MerchantID:  "87c84777-217e-49bb-92f1-986e9c2235f3",
		PaymentID:   "36aa12f4-b9bd-46c9-aae4-9a5c315e1f16",
		PaymentType: constant.TypeVirtualTerminal,
		CreatedFrom: constant.SourceDashboard,
		VirtualTerminal: &unifiedPaymentModel.VirtualTerminal{
			BatchID:            "47daf0d9-5f2b-4db0-8bc7-893833145dab",
			BookingID:          "BOOK001",   // NOSONAR
			TravelAgentCode:    "TRAVELOKA", // NOSONAR
			TravelAgentName:    "Traveloka", // NOSONAR
			Remarks:            "test",      // NOSONAR
			AcquirerMerchantID: "TEST001",   // NOSONAR
		},
	}

	conf := config.VccTerminalConfig{
		TravelAgents: config.MStrStr{
			"TRAVELOKA": "Traveloka", // NOSONAR
		},
	}
	result := booking.ToCreateUnifiedPaymentSessionRequest(conf)
	wantResult.ExpiryAt = result.ExpiryAt

	assert.Equal(t, wantResult, result)
}
