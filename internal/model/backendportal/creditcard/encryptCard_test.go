package card_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/creditcardCoreProcessor"
	"github.com/stretchr/testify/assert"
)

func TestEncryptCardRequest_ToProcessorRequestModel(t *testing.T) {
	tests := []struct {
		name     string
		request  *EncryptCardRequest
		expected *creditcardCoreProcessorModel.EncryptCardRequest
	}{
		{
			name: "complete_encrypt_card_request",
			request: &EncryptCardRequest{
				MerchantID:        "merchant-12345",
				ClientReferenceID: "client-ref-abc123",
				CardRequest: EncryptCardDetailRequest{
					Number:      "4111111111111111",
					ExpiryMonth: "12",
					ExpiryYear:  "25",
					CVC:         "123",
					NameOnCard:  "John Doe",
				},
				DeviceInformation: DeviceInformation{
					Type:           "web",
					UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
					IpAddress:      "192.168.1.100",
					AcceptLanguage: "en-US,en;q=0.9",
					CookieToken:    "cookie-token-xyz789",
					DeviceID:       "device-id-456",
					BrowserWidth:   "1920",
					BrowserHeight:  "1080",
					Country:        "US",
				},
				Metadata: map[string]string{
					"custom_field_1": "value1",
					"custom_field_2": "value2",
				},
			},
			expected: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "merchant-12345",
				ClientReferenceID: "client-ref-abc123",
				CardRequest: creditcardCoreProcessorModel.EncryptCardDetailRequest{
					Number:      "4111111111111111",
					ExpiryMonth: "12",
					ExpiryYear:  "25",
					CVC:         "123",
					NameOnCard:  "John Doe",
				},
				DeviceInformation: creditcardCoreProcessorModel.DeviceInformation{
					Type:           "web",
					UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
					IpAddress:      "192.168.1.100",
					AcceptLanguage: "en-US,en;q=0.9",
					CookieToken:    "cookie-token-xyz789",
					DeviceID:       "device-id-456",
					BrowserWidth:   "1920",
					BrowserHeight:  "1080",
					Country:        "US",
				},
				Metadata: map[string]string{
					"custom_field_1": "value1",
					"custom_field_2": "value2",
				},
			},
		},
		{
			name: "minimal_encrypt_card_request",
			request: &EncryptCardRequest{
				MerchantID:        "merchant-minimal",
				ClientReferenceID: "client-ref-minimal",
				CardRequest: EncryptCardDetailRequest{
					Number:      "5555555555554444",
					ExpiryMonth: "01",
					ExpiryYear:  "30",
					NameOnCard:  "Jane Smith",
				},
				DeviceInformation: DeviceInformation{
					Type: "mobile",
				},
			},
			expected: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "merchant-minimal",
				ClientReferenceID: "client-ref-minimal",
				CardRequest: creditcardCoreProcessorModel.EncryptCardDetailRequest{
					Number:      "5555555555554444",
					ExpiryMonth: "01",
					ExpiryYear:  "30",
					CVC:         "",
					NameOnCard:  "Jane Smith",
				},
				DeviceInformation: creditcardCoreProcessorModel.DeviceInformation{
					Type:           "mobile",
					UserAgent:      "",
					IpAddress:      "",
					AcceptLanguage: "",
					CookieToken:    "",
					DeviceID:       "",
					BrowserWidth:   "",
					BrowserHeight:  "",
					Country:        "",
				},
				Metadata: nil,
			},
		},
		{
			name: "amex_card_with_special_characters",
			request: &EncryptCardRequest{
				MerchantID:        "merchant-special-chars-!@#$%",
				ClientReferenceID: "client-ref-special-测试",
				CardRequest: EncryptCardDetailRequest{
					Number:      "378282246310005",
					ExpiryMonth: "12",
					ExpiryYear:  "99",
					CVC:         "1234",
					NameOnCard:  "José María García-López",
				},
				DeviceInformation: DeviceInformation{
					Type:           "tablet",
					UserAgent:      "Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
					IpAddress:      "2001:db8::8a2e:370:7334",
					AcceptLanguage: "es-ES,es;q=0.8,en;q=0.6",
					CookieToken:    "cookie-with-special-chars-!@#$%^&*()",
					DeviceID:       "device-id-with-unicode-测试",
					BrowserWidth:   "768",
					BrowserHeight:  "1024",
					Country:        "ES",
				},
				Metadata: map[string]string{
					"special_field_!@#": "value_with_unicode_测试",
					"empty_field":       "",
					"numeric_field":     "12345",
				},
			},
			expected: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "merchant-special-chars-!@#$%",
				ClientReferenceID: "client-ref-special-测试",
				CardRequest: creditcardCoreProcessorModel.EncryptCardDetailRequest{
					Number:      "378282246310005",
					ExpiryMonth: "12",
					ExpiryYear:  "99",
					CVC:         "1234",
					NameOnCard:  "José María García-López",
				},
				DeviceInformation: creditcardCoreProcessorModel.DeviceInformation{
					Type:           "tablet",
					UserAgent:      "Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
					IpAddress:      "2001:db8::8a2e:370:7334",
					AcceptLanguage: "es-ES,es;q=0.8,en;q=0.6",
					CookieToken:    "cookie-with-special-chars-!@#$%^&*()",
					DeviceID:       "device-id-with-unicode-测试",
					BrowserWidth:   "768",
					BrowserHeight:  "1024",
					Country:        "ES",
				},
				Metadata: map[string]string{
					"special_field_!@#": "value_with_unicode_测试",
					"empty_field":       "",
					"numeric_field":     "12345",
				},
			},
		},
		{
			name: "empty_strings_and_edge_cases",
			request: &EncryptCardRequest{
				MerchantID:        "",
				ClientReferenceID: "",
				CardRequest: EncryptCardDetailRequest{
					Number:      "",
					ExpiryMonth: "",
					ExpiryYear:  "",
					CVC:         "",
					NameOnCard:  "",
				},
				DeviceInformation: DeviceInformation{
					Type:           "",
					UserAgent:      "",
					IpAddress:      "",
					AcceptLanguage: "",
					CookieToken:    "",
					DeviceID:       "",
					BrowserWidth:   "",
					BrowserHeight:  "",
					Country:        "",
				},
				Metadata: map[string]string{},
			},
			expected: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "",
				ClientReferenceID: "",
				CardRequest: creditcardCoreProcessorModel.EncryptCardDetailRequest{
					Number:      "",
					ExpiryMonth: "",
					ExpiryYear:  "",
					CVC:         "",
					NameOnCard:  "",
				},
				DeviceInformation: creditcardCoreProcessorModel.DeviceInformation{
					Type:           "",
					UserAgent:      "",
					IpAddress:      "",
					AcceptLanguage: "",
					CookieToken:    "",
					DeviceID:       "",
					BrowserWidth:   "",
					BrowserHeight:  "",
					Country:        "",
				},
				Metadata: map[string]string{},
			},
		},
		{
			name: "discover_card_with_long_values",
			request: &EncryptCardRequest{
				MerchantID:        "merchant-with-very-long-identifier-that-exceeds-normal-length-requirements",
				ClientReferenceID: "client-reference-id-with-extremely-long-value-that-might-cause-issues-in-some-systems",
				CardRequest: EncryptCardDetailRequest{
					Number:      "6011111111111117",
					ExpiryMonth: "06",
					ExpiryYear:  "28",
					CVC:         "999",
					NameOnCard:  "Alexander Bartholomew Christopher Maximilian Van Der Berg-Johnson III",
				},
				DeviceInformation: DeviceInformation{
					Type:           "desktop",
					UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edge/91.0.864.59 Very-Long-User-Agent-String-That-Contains-Multiple-Browser-Identifiers",
					IpAddress:      "203.0.113.195",
					AcceptLanguage: "en-US,en;q=0.9,es;q=0.8,fr;q=0.7,de;q=0.6,it;q=0.5,pt;q=0.4",
					CookieToken:    "very-long-cookie-token-with-base64-encoded-data-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
					DeviceID:       "device-id-with-guid-12345678-90ab-cdef-1234-567890abcdef-and-additional-identifiers",
					BrowserWidth:   "3840",
					BrowserHeight:  "2160",
					Country:        "NL",
				},
				Metadata: map[string]string{
					"transaction_type":     "purchase",
					"integration_version":  "v2.1.0",
					"sdk_version":          "1.0.5",
					"platform":             "web",
					"session_id":           "session-12345678-90ab-cdef-1234-567890abcdef",
					"merchant_custom_data": "very-long-custom-data-field-that-contains-business-specific-information-and-metadata",
				},
			},
			expected: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "merchant-with-very-long-identifier-that-exceeds-normal-length-requirements",
				ClientReferenceID: "client-reference-id-with-extremely-long-value-that-might-cause-issues-in-some-systems",
				CardRequest: creditcardCoreProcessorModel.EncryptCardDetailRequest{
					Number:      "6011111111111117",
					ExpiryMonth: "06",
					ExpiryYear:  "28",
					CVC:         "999",
					NameOnCard:  "Alexander Bartholomew Christopher Maximilian Van Der Berg-Johnson III",
				},
				DeviceInformation: creditcardCoreProcessorModel.DeviceInformation{
					Type:           "desktop",
					UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edge/91.0.864.59 Very-Long-User-Agent-String-That-Contains-Multiple-Browser-Identifiers",
					IpAddress:      "203.0.113.195",
					AcceptLanguage: "en-US,en;q=0.9,es;q=0.8,fr;q=0.7,de;q=0.6,it;q=0.5,pt;q=0.4",
					CookieToken:    "very-long-cookie-token-with-base64-encoded-data-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
					DeviceID:       "device-id-with-guid-12345678-90ab-cdef-1234-567890abcdef-and-additional-identifiers",
					BrowserWidth:   "3840",
					BrowserHeight:  "2160",
					Country:        "NL",
				},
				Metadata: map[string]string{
					"transaction_type":     "purchase",
					"integration_version":  "v2.1.0",
					"sdk_version":          "1.0.5",
					"platform":             "web",
					"session_id":           "session-12345678-90ab-cdef-1234-567890abcdef",
					"merchant_custom_data": "very-long-custom-data-field-that-contains-business-specific-information-and-metadata",
				},
			},
		},
		{
			name: "nil_metadata_map",
			request: &EncryptCardRequest{
				MerchantID:        "merchant-nil-metadata",
				ClientReferenceID: "client-ref-nil-metadata",
				CardRequest: EncryptCardDetailRequest{
					Number:      "4000000000000002",
					ExpiryMonth: "03",
					ExpiryYear:  "26",
					CVC:         "456",
					NameOnCard:  "Test User",
				},
				DeviceInformation: DeviceInformation{
					Type:      "mobile",
					UserAgent: "Mobile App v1.0",
					Country:   "ID",
				},
				Metadata: nil,
			},
			expected: &creditcardCoreProcessorModel.EncryptCardRequest{
				MerchantID:        "merchant-nil-metadata",
				ClientReferenceID: "client-ref-nil-metadata",
				CardRequest: creditcardCoreProcessorModel.EncryptCardDetailRequest{
					Number:      "4000000000000002",
					ExpiryMonth: "03",
					ExpiryYear:  "26",
					CVC:         "456",
					NameOnCard:  "Test User",
				},
				DeviceInformation: creditcardCoreProcessorModel.DeviceInformation{
					Type:           "mobile",
					UserAgent:      "Mobile App v1.0",
					IpAddress:      "",
					AcceptLanguage: "",
					CookieToken:    "",
					DeviceID:       "",
					BrowserWidth:   "",
					BrowserHeight:  "",
					Country:        "ID",
				},
				Metadata: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ToProcessorRequestModel()

			// Assert basic fields
			assert.Equal(t, tt.expected.MerchantID, result.MerchantID)
			assert.Equal(t, tt.expected.ClientReferenceID, result.ClientReferenceID)

			// Assert CardRequest fields
			assert.Equal(t, tt.expected.CardRequest.Number, result.CardRequest.Number)
			assert.Equal(t, tt.expected.CardRequest.ExpiryMonth, result.CardRequest.ExpiryMonth)
			assert.Equal(t, tt.expected.CardRequest.ExpiryYear, result.CardRequest.ExpiryYear)
			assert.Equal(t, tt.expected.CardRequest.CVC, result.CardRequest.CVC)
			assert.Equal(t, tt.expected.CardRequest.NameOnCard, result.CardRequest.NameOnCard)

			// Assert DeviceInformation fields
			assert.Equal(t, tt.expected.DeviceInformation.Type, result.DeviceInformation.Type)
			assert.Equal(t, tt.expected.DeviceInformation.UserAgent, result.DeviceInformation.UserAgent)
			assert.Equal(t, tt.expected.DeviceInformation.IpAddress, result.DeviceInformation.IpAddress)
			assert.Equal(t, tt.expected.DeviceInformation.AcceptLanguage, result.DeviceInformation.AcceptLanguage)
			assert.Equal(t, tt.expected.DeviceInformation.CookieToken, result.DeviceInformation.CookieToken)
			assert.Equal(t, tt.expected.DeviceInformation.DeviceID, result.DeviceInformation.DeviceID)
			assert.Equal(t, tt.expected.DeviceInformation.BrowserWidth, result.DeviceInformation.BrowserWidth)
			assert.Equal(t, tt.expected.DeviceInformation.BrowserHeight, result.DeviceInformation.BrowserHeight)
			assert.Equal(t, tt.expected.DeviceInformation.Country, result.DeviceInformation.Country)

			// Assert Metadata
			assert.Equal(t, tt.expected.Metadata, result.Metadata)
		})
	}
}

func TestToCardResponseModel(t *testing.T) {
	tests := []struct {
		name              string
		processorResponse *creditcardCoreProcessorModel.EncryptedCardResponse
		expected          *EncryptedCardResponse
	}{
		{
			name: "complete_card_response",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardResponse{
				ClientReferenceID: "client-ref-complete",
				EncryptedCard:     "encrypted-card-data-base64-encoded-string",
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationResponse{
					First8Digits:     "41111111",
					First6Digits:     "411111",
					Last4Digits:      "1111",
					ExpiryMonth:      "12",
					ExpiryYear:       "25",
					HasAssociatedCVC: true,
					Fingerprint:      "fingerprint-hash-abc123",
				},
				DeviceInfomation: creditcardCoreProcessorModel.DeviceInformation{
					Type:           "web",
					UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
					IpAddress:      "192.168.1.100",
					AcceptLanguage: "en-US,en;q=0.9",
					CookieToken:    "cookie-token-xyz",
					DeviceID:       "device-id-123",
					BrowserWidth:   "1920",
					BrowserHeight:  "1080",
					Country:        "US",
				},
				BinDetail: creditcardCoreProcessorModel.Bin{
					UUID:          uuid.New(),
					BinNumber:     "411111",
					CardType:      "CREDIT",
					CardBrand:     "VISA",
					ConsumerType:  "CONSUMER",
					CardLevel:     "CLASSIC",
					IssuerName:    "Chase Bank",
					IssuerCountry: "US",
					Currency:      "USD",
					Status:        "ACTIVE",
					IsBlocked:     false,
					CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:     time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
				},
				CreatedAt: "2023-10-01T12:00:00Z",
				Metadata: map[string]string{
					"transaction_id": "txn-123",
					"source":         "web",
				},
			},
			expected: &EncryptedCardResponse{
				ClientReferenceID: "client-ref-complete",
				EncryptedCard:     "encrypted-card-data-base64-encoded-string",
				EncryptedCardInformation: EncryptedCardInformationResponse{
					First8Digits:     "41111111",
					First6Digits:     "411111",
					Last4Digits:      "1111",
					ExpiryMonth:      "12",
					ExpiryYear:       "25",
					HasAssociatedCVC: true,
					Fingerprint:      "fingerprint-hash-abc123",
				},
				DeviceInfomation: DeviceInformation{
					Type:           "web",
					UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
					IpAddress:      "192.168.1.100",
					AcceptLanguage: "en-US,en;q=0.9",
					CookieToken:    "cookie-token-xyz",
					DeviceID:       "device-id-123",
					BrowserWidth:   "1920",
					BrowserHeight:  "1080",
					Country:        "US",
				},
				BinDetail: Bin{
					UUID:          uuid.UUID{}, // Will be set from processor response
					BinNumber:     "411111",
					CardType:      "CREDIT",
					CardBrand:     "VISA",
					ConsumerType:  "CONSUMER",
					CardLevel:     "CLASSIC",
					IssuerName:    "Chase Bank",
					IssuerCountry: "US",
					Currency:      "USD",
					Status:        "ACTIVE",
					IsBlocked:     false,
					CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:     time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
				},
				CreatedAt: "2023-10-01T12:00:00Z",
				Metadata: map[string]string{
					"transaction_id": "txn-123",
					"source":         "web",
				},
			},
		},
		{
			name: "minimal_card_response",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardResponse{
				ClientReferenceID: "minimal-ref",
				EncryptedCard:     "minimal-encrypted-data",
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationResponse{
					First8Digits:     "55555555",
					First6Digits:     "555555",
					Last4Digits:      "4444",
					ExpiryMonth:      "01",
					ExpiryYear:       "30",
					HasAssociatedCVC: false,
					Fingerprint:      "minimal-fingerprint",
				},
				DeviceInfomation: creditcardCoreProcessorModel.DeviceInformation{
					Type: "mobile",
				},
				BinDetail: creditcardCoreProcessorModel.Bin{
					BinNumber: "555555",
					CardType:  "DEBIT",
					Status:    "ACTIVE",
				},
				CreatedAt: "2023-01-01T00:00:00Z",
			},
			expected: &EncryptedCardResponse{
				ClientReferenceID: "minimal-ref",
				EncryptedCard:     "minimal-encrypted-data",
				EncryptedCardInformation: EncryptedCardInformationResponse{
					First8Digits:     "55555555",
					First6Digits:     "555555",
					Last4Digits:      "4444",
					ExpiryMonth:      "01",
					ExpiryYear:       "30",
					HasAssociatedCVC: false,
					Fingerprint:      "minimal-fingerprint",
				},
				DeviceInfomation: DeviceInformation{
					Type: "mobile",
				},
				BinDetail: Bin{
					BinNumber: "555555",
					CardType:  "DEBIT",
					Status:    "ACTIVE",
				},
				CreatedAt: "2023-01-01T00:00:00Z",
				Metadata:  nil,
			},
		},
		{
			name: "card_response_with_special_characters",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardResponse{
				ClientReferenceID: "special-chars-测试-!@#$%",
				EncryptedCard:     "encrypted-data-with-unicode-测试-characters",
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationResponse{
					First8Digits:     "37828224",
					First6Digits:     "378282",
					Last4Digits:      "0005",
					ExpiryMonth:      "12",
					ExpiryYear:       "99",
					HasAssociatedCVC: true,
					Fingerprint:      "fingerprint-with-special-chars-αβγ",
				},
				DeviceInfomation: creditcardCoreProcessorModel.DeviceInformation{
					Type:           "tablet",
					UserAgent:      "Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X)",
					IpAddress:      "2001:db8::8a2e:370:7334",
					AcceptLanguage: "es-ES,es;q=0.8,en;q=0.6",
					CookieToken:    "cookie-with-unicode-测试",
					DeviceID:       "device-special-chars-!@#$%",
					BrowserWidth:   "768",
					BrowserHeight:  "1024",
					Country:        "ES",
				},
				BinDetail: creditcardCoreProcessorModel.Bin{
					UUID:          uuid.New(),
					BinNumber:     "378282",
					CardType:      "CREDIT",
					CardBrand:     "AMEX",
					ConsumerType:  "PREMIUM",
					CardLevel:     "GOLD",
					IssuerName:    "Banco de España & Co. (测试)",
					IssuerCountry: "ES",
					Currency:      "EUR",
					Status:        "ACTIVE",
					IsBlocked:     false,
					CreatedAt:     time.Date(2023, 5, 15, 14, 30, 45, 0, time.UTC),
					UpdatedAt:     time.Date(2023, 10, 1, 16, 45, 30, 0, time.UTC),
				},
				CreatedAt: "2023-10-01T16:45:30Z",
				Metadata: map[string]string{
					"unicode_field_测试": "value_with_special_chars_!@#$%",
					"empty_field":      "",
					"html_content":     "<div>HTML content &amp; entities</div>",
				},
			},
			expected: &EncryptedCardResponse{
				ClientReferenceID: "special-chars-测试-!@#$%",
				EncryptedCard:     "encrypted-data-with-unicode-测试-characters",
				EncryptedCardInformation: EncryptedCardInformationResponse{
					First8Digits:     "37828224",
					First6Digits:     "378282",
					Last4Digits:      "0005",
					ExpiryMonth:      "12",
					ExpiryYear:       "99",
					HasAssociatedCVC: true,
					Fingerprint:      "fingerprint-with-special-chars-αβγ",
				},
				DeviceInfomation: DeviceInformation{
					Type:           "tablet",
					UserAgent:      "Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X)",
					IpAddress:      "2001:db8::8a2e:370:7334",
					AcceptLanguage: "es-ES,es;q=0.8,en;q=0.6",
					CookieToken:    "cookie-with-unicode-测试",
					DeviceID:       "device-special-chars-!@#$%",
					BrowserWidth:   "768",
					BrowserHeight:  "1024",
					Country:        "ES",
				},
				BinDetail: Bin{
					UUID:          uuid.UUID{}, // Will be set from processor response
					BinNumber:     "378282",
					CardType:      "CREDIT",
					CardBrand:     "AMEX",
					ConsumerType:  "PREMIUM",
					CardLevel:     "GOLD",
					IssuerName:    "Banco de España & Co. (测试)",
					IssuerCountry: "ES",
					Currency:      "EUR",
					Status:        "ACTIVE",
					IsBlocked:     false,
					CreatedAt:     time.Date(2023, 5, 15, 14, 30, 45, 0, time.UTC),
					UpdatedAt:     time.Date(2023, 10, 1, 16, 45, 30, 0, time.UTC),
				},
				CreatedAt: "2023-10-01T16:45:30Z",
				Metadata: map[string]string{
					"unicode_field_测试": "value_with_special_chars_!@#$%",
					"empty_field":      "",
					"html_content":     "<div>HTML content &amp; entities</div>",
				},
			},
		},
		{
			name: "empty_fields_card_response",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardResponse{
				ClientReferenceID: "",
				EncryptedCard:     "",
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationResponse{
					First8Digits:     "",
					First6Digits:     "",
					Last4Digits:      "",
					ExpiryMonth:      "",
					ExpiryYear:       "",
					HasAssociatedCVC: false,
					Fingerprint:      "",
				},
				DeviceInfomation: creditcardCoreProcessorModel.DeviceInformation{
					Type:           "",
					UserAgent:      "",
					IpAddress:      "",
					AcceptLanguage: "",
					CookieToken:    "",
					DeviceID:       "",
					BrowserWidth:   "",
					BrowserHeight:  "",
					Country:        "",
				},
				BinDetail: creditcardCoreProcessorModel.Bin{},
				CreatedAt: "",
				Metadata:  map[string]string{},
			},
			expected: &EncryptedCardResponse{
				ClientReferenceID: "",
				EncryptedCard:     "",
				EncryptedCardInformation: EncryptedCardInformationResponse{
					First8Digits:     "",
					First6Digits:     "",
					Last4Digits:      "",
					ExpiryMonth:      "",
					ExpiryYear:       "",
					HasAssociatedCVC: false,
					Fingerprint:      "",
				},
				DeviceInfomation: DeviceInformation{
					Type:           "",
					UserAgent:      "",
					IpAddress:      "",
					AcceptLanguage: "",
					CookieToken:    "",
					DeviceID:       "",
					BrowserWidth:   "",
					BrowserHeight:  "",
					Country:        "",
				},
				BinDetail: Bin{},
				CreatedAt: "",
				Metadata:  map[string]string{},
			},
		},
		{
			name: "blocked_card_with_deleted_at",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardResponse{
				ClientReferenceID: "blocked-card-ref",
				EncryptedCard:     "blocked-card-encrypted-data",
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationResponse{
					First8Digits:     "99999999",
					First6Digits:     "999999",
					Last4Digits:      "9999",
					ExpiryMonth:      "12",
					ExpiryYear:       "99",
					HasAssociatedCVC: true,
					Fingerprint:      "blocked-card-fingerprint",
				},
				DeviceInfomation: creditcardCoreProcessorModel.DeviceInformation{
					Type:      "web",
					IpAddress: "10.0.0.1",
					Country:   "XX",
				},
				BinDetail: creditcardCoreProcessorModel.Bin{
					UUID:          uuid.New(),
					BinNumber:     "999999",
					CardType:      "CREDIT",
					CardBrand:     "UNKNOWN",
					Status:        "BLOCKED",
					IsBlocked:     true,
					IssuerName:    "Blocked Bank",
					IssuerCountry: "XX",
					CreatedAt:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:     time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
					DeletedAt:     sql.NullTime{Valid: true, Time: time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)},
				},
				CreatedAt: "2023-09-01T00:00:00Z",
				Metadata: map[string]string{
					"block_reason": "suspicious_activity",
					"blocked_at":   "2023-09-01T00:00:00Z",
				},
			},
			expected: &EncryptedCardResponse{
				ClientReferenceID: "blocked-card-ref",
				EncryptedCard:     "blocked-card-encrypted-data",
				EncryptedCardInformation: EncryptedCardInformationResponse{
					First8Digits:     "99999999",
					First6Digits:     "999999",
					Last4Digits:      "9999",
					ExpiryMonth:      "12",
					ExpiryYear:       "99",
					HasAssociatedCVC: true,
					Fingerprint:      "blocked-card-fingerprint",
				},
				DeviceInfomation: DeviceInformation{
					Type:      "web",
					IpAddress: "10.0.0.1",
					Country:   "XX",
				},
				BinDetail: Bin{
					UUID:          uuid.UUID{}, // Will be set from processor response
					BinNumber:     "999999",
					CardType:      "CREDIT",
					CardBrand:     "UNKNOWN",
					Status:        "BLOCKED",
					IsBlocked:     true,
					IssuerName:    "Blocked Bank",
					IssuerCountry: "XX",
					CreatedAt:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:     time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
					DeletedAt:     sql.NullTime{Valid: true, Time: time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)},
				},
				CreatedAt: "2023-09-01T00:00:00Z",
				Metadata: map[string]string{
					"block_reason": "suspicious_activity",
					"blocked_at":   "2023-09-01T00:00:00Z",
				},
			},
		},
		{
			name:              "nil_processor_response",
			processorResponse: nil,
			expected:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCardResponseModel(tt.processorResponse)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)

			// Assert basic fields
			assert.Equal(t, tt.expected.ClientReferenceID, result.ClientReferenceID)
			assert.Equal(t, tt.expected.EncryptedCard, result.EncryptedCard)
			assert.Equal(t, tt.expected.CreatedAt, result.CreatedAt)

			// Assert EncryptedCardInformation fields
			assert.Equal(t, tt.expected.EncryptedCardInformation.First8Digits, result.EncryptedCardInformation.First8Digits)
			assert.Equal(t, tt.expected.EncryptedCardInformation.First6Digits, result.EncryptedCardInformation.First6Digits)
			assert.Equal(t, tt.expected.EncryptedCardInformation.Last4Digits, result.EncryptedCardInformation.Last4Digits)
			assert.Equal(t, tt.expected.EncryptedCardInformation.ExpiryMonth, result.EncryptedCardInformation.ExpiryMonth)
			assert.Equal(t, tt.expected.EncryptedCardInformation.ExpiryYear, result.EncryptedCardInformation.ExpiryYear)
			assert.Equal(t, tt.expected.EncryptedCardInformation.HasAssociatedCVC, result.EncryptedCardInformation.HasAssociatedCVC)
			assert.Equal(t, tt.expected.EncryptedCardInformation.Fingerprint, result.EncryptedCardInformation.Fingerprint)

			// Assert DeviceInformation fields
			assert.Equal(t, tt.expected.DeviceInfomation.Type, result.DeviceInfomation.Type)
			assert.Equal(t, tt.expected.DeviceInfomation.UserAgent, result.DeviceInfomation.UserAgent)
			assert.Equal(t, tt.expected.DeviceInfomation.IpAddress, result.DeviceInfomation.IpAddress)
			assert.Equal(t, tt.expected.DeviceInfomation.AcceptLanguage, result.DeviceInfomation.AcceptLanguage)
			assert.Equal(t, tt.expected.DeviceInfomation.CookieToken, result.DeviceInfomation.CookieToken)
			assert.Equal(t, tt.expected.DeviceInfomation.DeviceID, result.DeviceInfomation.DeviceID)
			assert.Equal(t, tt.expected.DeviceInfomation.BrowserWidth, result.DeviceInfomation.BrowserWidth)
			assert.Equal(t, tt.expected.DeviceInfomation.BrowserHeight, result.DeviceInfomation.BrowserHeight)
			assert.Equal(t, tt.expected.DeviceInfomation.Country, result.DeviceInfomation.Country)

			// Assert BinDetail fields (excluding UUID and time fields for easier testing)
			assert.Equal(t, tt.expected.BinDetail.BinNumber, result.BinDetail.BinNumber)
			assert.Equal(t, tt.expected.BinDetail.CardType, result.BinDetail.CardType)
			assert.Equal(t, tt.expected.BinDetail.CardBrand, result.BinDetail.CardBrand)
			assert.Equal(t, tt.expected.BinDetail.ConsumerType, result.BinDetail.ConsumerType)
			assert.Equal(t, tt.expected.BinDetail.CardLevel, result.BinDetail.CardLevel)
			assert.Equal(t, tt.expected.BinDetail.IssuerName, result.BinDetail.IssuerName)
			assert.Equal(t, tt.expected.BinDetail.IssuerCountry, result.BinDetail.IssuerCountry)
			assert.Equal(t, tt.expected.BinDetail.Currency, result.BinDetail.Currency)
			assert.Equal(t, tt.expected.BinDetail.Status, result.BinDetail.Status)
			assert.Equal(t, tt.expected.BinDetail.IsBlocked, result.BinDetail.IsBlocked)

			// For time fields, check if they are properly set from input
			if !tt.processorResponse.BinDetail.CreatedAt.IsZero() {
				assert.Equal(t, tt.processorResponse.BinDetail.CreatedAt, result.BinDetail.CreatedAt)
			}
			if !tt.processorResponse.BinDetail.UpdatedAt.IsZero() {
				assert.Equal(t, tt.processorResponse.BinDetail.UpdatedAt, result.BinDetail.UpdatedAt)
			}
			assert.Equal(t, tt.processorResponse.BinDetail.DeletedAt, result.BinDetail.DeletedAt)

			// Assert Metadata
			assert.Equal(t, tt.expected.Metadata, result.Metadata)
		})
	}
}
