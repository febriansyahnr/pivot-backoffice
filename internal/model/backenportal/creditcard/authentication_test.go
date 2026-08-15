package card

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	"github.com/stretchr/testify/assert"
)

func TestEncryptedCardAuthenticationRequest_ToProcessorRequestModel(t *testing.T) {
	request := &EncryptedCardAuthenticationRequest{
		PaymentID:           "payment-123",
		MerchantID:          "merchant-456",
		ClientTransactionID: "client-789",
		CardID:              "card-abc",
		CVC:                 "123",
		Amount:              1000.50,
		Fee:                 50.25,
		Currency:            "IDR",
		ExternalThreeDsInfo: &ExternalThreeDsInfo{
			TransactionID: "transaction-123",
		},
	}

	result := request.ToProcessorRequestModel()

	if result.PaymentID != request.PaymentID {
		t.Errorf("Expected PaymentID %s, got %s", request.PaymentID, result.PaymentID)
	}
	if result.MerchantID != request.MerchantID {
		t.Errorf("Expected MerchantID %s, got %s", request.MerchantID, result.MerchantID)
	}
	if result.ClientTransactionID != request.ClientTransactionID {
		t.Errorf("Expected ClientTransactionID %s, got %s", request.ClientTransactionID, result.ClientTransactionID)
	}
	if result.CardID != request.CardID {
		t.Errorf("Expected CardID %s, got %s", request.CardID, result.CardID)
	}
	if result.CVC != request.CVC {
		t.Errorf("Expected CVC %s, got %s", request.CVC, result.CVC)
	}
	if result.Amount != request.Amount {
		t.Errorf("Expected Amount %f, got %f", request.Amount, result.Amount)
	}
	if result.Fee != request.Fee {
		t.Errorf("Expected Fee %f, got %f", request.Fee, result.Fee)
	}
	if result.Currency != request.Currency {
		t.Errorf("Expected Currency %s, got %s", request.Currency, result.Currency)
	}
	if result.ExternalThreeDsInfo.TransactionID != request.ExternalThreeDsInfo.TransactionID {
		t.Errorf("Expected ExternalThreeDsInfo.Transaction %s, got %s", request.ExternalThreeDsInfo.TransactionID, result.ExternalThreeDsInfo.TransactionID)
	}
}

func TestGetAuthenticationResponseFromProcessor(t *testing.T) {
	tests := []struct {
		name              string
		processorResponse *creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse
		expected          *AuthenticationResponse
	}{
		{
			name: "complete_authentication_response",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				AcquirerTransactionID: "acquirer-123",
				Amount:                "1000.50",
				Currency:              "IDR",
				Message:               "Success",
				SessionID:             "session-456",
				Status:                "AUTHENTICATED",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					ActionURL:    "https://action.url",
					CreatedAt:    "2023-01-01T00:00:00Z",
					ThreeDSToken: "token-123",
					HTML:         "<html>3DS form</html>",
					Method:       "POST",
					URL:          "https://3ds.url",
					Version:      "2.1.0",
				},
			},
			expected: &AuthenticationResponse{
				AcquirerTransactionID: "acquirer-123",
				Amount:                "1000.50",
				Currency:              "IDR",
				Message:               "Success",
				SessionID:             "session-456",
				Status:                "AUTHENTICATED",
				AuthenticationURL: AuthenticationURLDetail{
					ActionURL:    "https://action.url",
					CreatedAt:    "2023-01-01T00:00:00Z",
					ThreeDSToken: "token-123",
					HTML:         "<html>3DS form</html>",
					Method:       "POST",
					URL:          "https://3ds.url",
					Version:      "2.1.0",
				},
			},
		},
		{
			name: "minimal_fields",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				Status: "PENDING",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					Method: "GET",
				},
			},
			expected: &AuthenticationResponse{
				Status: "PENDING",
				AuthenticationURL: AuthenticationURLDetail{
					Method: "GET",
				},
			},
		},
		{
			name: "empty_fields",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				AcquirerTransactionID: "",
				Amount:                "",
				Currency:              "",
				Message:               "",
				SessionID:             "",
				Status:                "",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					ActionURL:    "",
					CreatedAt:    "",
					ThreeDSToken: "",
					HTML:         "",
					Method:       "",
					URL:          "",
					Version:      "",
				},
			},
			expected: &AuthenticationResponse{
				AcquirerTransactionID: "",
				Amount:                "",
				Currency:              "",
				Message:               "",
				SessionID:             "",
				Status:                "",
				AuthenticationURL: AuthenticationURLDetail{
					ActionURL:    "",
					CreatedAt:    "",
					ThreeDSToken: "",
					HTML:         "",
					Method:       "",
					URL:          "",
					Version:      "",
				},
			},
		},
		{
			name: "failed_authentication",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				AcquirerTransactionID: "acquirer-failed-789",
				Amount:                "500.00",
				Currency:              "USD",
				Message:               "Authentication failed - invalid card",
				SessionID:             "session-failed-123",
				Status:                "FAILED",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					ActionURL:    "https://failure.url",
					CreatedAt:    "2023-10-01T15:30:45Z",
					ThreeDSToken: "",
					HTML:         "<div>Error: Authentication failed</div>",
					Method:       "GET",
					URL:          "https://error.url",
					Version:      "1.0.0",
				},
			},
			expected: &AuthenticationResponse{
				AcquirerTransactionID: "acquirer-failed-789",
				Amount:                "500.00",
				Currency:              "USD",
				Message:               "Authentication failed - invalid card",
				SessionID:             "session-failed-123",
				Status:                "FAILED",
				AuthenticationURL: AuthenticationURLDetail{
					ActionURL:    "https://failure.url",
					CreatedAt:    "2023-10-01T15:30:45Z",
					ThreeDSToken: "",
					HTML:         "<div>Error: Authentication failed</div>",
					Method:       "GET",
					URL:          "https://error.url",
					Version:      "1.0.0",
				},
			},
		},
		{
			name: "special_characters_in_fields",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				AcquirerTransactionID: "acquirer-特殊字符-123",
				Amount:                "1,234.56",
				Currency:              "EUR",
				Message:               "Message with special chars: !@#$%^&*()",
				SessionID:             "session-with-unicode-αβγ",
				Status:                "AUTHENTICATED",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					ActionURL:    "https://example.com/action?param=value&other=123",
					CreatedAt:    "2023-12-31T23:59:59.999Z",
					ThreeDSToken: "token-with-symbols-!@#$%",
					HTML:         "<form><input type=\"hidden\" name=\"PaReq\" value=\"test&amp;value\"/></form>",
					Method:       "POST",
					URL:          "https://secure.example.com/3ds?merchant=123&ref=abc",
					Version:      "2.2.0",
				},
			},
			expected: &AuthenticationResponse{
				AcquirerTransactionID: "acquirer-特殊字符-123",
				Amount:                "1,234.56",
				Currency:              "EUR",
				Message:               "Message with special chars: !@#$%^&*()",
				SessionID:             "session-with-unicode-αβγ",
				Status:                "AUTHENTICATED",
				AuthenticationURL: AuthenticationURLDetail{
					ActionURL:    "https://example.com/action?param=value&other=123",
					CreatedAt:    "2023-12-31T23:59:59.999Z",
					ThreeDSToken: "token-with-symbols-!@#$%",
					HTML:         "<form><input type=\"hidden\" name=\"PaReq\" value=\"test&amp;value\"/></form>",
					Method:       "POST",
					URL:          "https://secure.example.com/3ds?merchant=123&ref=abc",
					Version:      "2.2.0",
				},
			},
		},
		{
			name: "large_html_content",
			processorResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				AcquirerTransactionID: "large-content-test",
				Status:                "CHALLENGED",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					HTML:   `<html><head><title>3DS Challenge</title></head><body><form id="challengeForm" method="POST" action="https://challenge.example.com"><input type="hidden" name="creq" value="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9"/><script>document.getElementById('challengeForm').submit();</script></form></body></html>`,
					Method: "POST",
				},
			},
			expected: &AuthenticationResponse{
				AcquirerTransactionID: "large-content-test",
				Status:                "CHALLENGED",
				AuthenticationURL: AuthenticationURLDetail{
					HTML:   `<html><head><title>3DS Challenge</title></head><body><form id="challengeForm" method="POST" action="https://challenge.example.com"><input type="hidden" name="creq" value="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9"/><script>document.getElementById('challengeForm').submit();</script></form></body></html>`,
					Method: "POST",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAuthenticationResponseFromProcessor(tt.processorResponse)

			assert.Equal(t, tt.expected.AcquirerTransactionID, result.AcquirerTransactionID)
			assert.Equal(t, tt.expected.Amount, result.Amount)
			assert.Equal(t, tt.expected.Currency, result.Currency)
			assert.Equal(t, tt.expected.Message, result.Message)
			assert.Equal(t, tt.expected.SessionID, result.SessionID)
			assert.Equal(t, tt.expected.Status, result.Status)

			// Assert AuthenticationURL fields
			assert.Equal(t, tt.expected.AuthenticationURL.ActionURL, result.AuthenticationURL.ActionURL)
			assert.Equal(t, tt.expected.AuthenticationURL.CreatedAt, result.AuthenticationURL.CreatedAt)
			assert.Equal(t, tt.expected.AuthenticationURL.ThreeDSToken, result.AuthenticationURL.ThreeDSToken)
			assert.Equal(t, tt.expected.AuthenticationURL.HTML, result.AuthenticationURL.HTML)
			assert.Equal(t, tt.expected.AuthenticationURL.Method, result.AuthenticationURL.Method)
			assert.Equal(t, tt.expected.AuthenticationURL.URL, result.AuthenticationURL.URL)
			assert.Equal(t, tt.expected.AuthenticationURL.Version, result.AuthenticationURL.Version)
		})
	}
}

func TestAuthenticationResponseStruct(t *testing.T) {
	response := AuthenticationResponse{
		AcquirerTransactionID: "acquirer-123",
		AuthenticationURL: AuthenticationURLDetail{
			ActionURL: "https://action.url",
		},
	}

	if response.AcquirerTransactionID != "acquirer-123" {
		t.Errorf("Expected AcquirerTransactionID 'acquirer-123', got %s", response.AcquirerTransactionID)
	}
	if response.AuthenticationURL.ActionURL != "https://action.url" {
		t.Errorf("Expected ActionURL 'https://action.url', got %s", response.AuthenticationURL.ActionURL)
	}
}

func TestEncryptedCardAuthenticationResponseStruct(t *testing.T) {
	response := EncryptedCardAuthenticationResponse{
		CardID: "card-123",
		CardInfo: EncryptedCardInformationResponse{
			First8Digits: "12345678",
			Last4Digits:  "9876",
		},
		Bin: Bin{
			UUID:      uuid.New(),
			BinNumber: "123456",
			CardType:  "CREDIT",
		},
		AuthenticationResponse: AuthenticationResponse{
			Status: "SUCCESS",
		},
	}

	if response.CardID != "card-123" {
		t.Errorf("Expected CardID 'card-123', got %s", response.CardID)
	}
	if response.CardInfo.First8Digits != "12345678" {
		t.Errorf("Expected First8Digits '12345678', got %s", response.CardInfo.First8Digits)
	}
}

func TestToEncryptedCardAuthenticationResponse(t *testing.T) {
	tests := []struct {
		name         string
		request      *EncryptedCardAuthenticationRequest
		authResponse *creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse
		expected     *EncryptedCardAuthenticationResponse
	}{
		{
			name: "complete_valid_response",
			request: &EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
				CardID:              "card-abc-def",
				CVC:                 "123",
				Amount:              1000.50,
				Fee:                 25.00,
				Currency:            "IDR",
			},
			authResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				AcquirerTransactionID: "acquirer-txn-123",
				Amount:                "1000.50",
				Currency:              "IDR",
				Message:               "Authentication successful",
				SessionID:             "session-456",
				Status:                "AUTHENTICATED",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					ActionURL:    "https://action.example.com",
					CreatedAt:    "2023-10-01T12:00:00Z",
					ThreeDSToken: "threeds-token-123",
					HTML:         "<form>3DS Form</form>",
					Method:       "POST",
					URL:          "https://3ds.example.com",
					Version:      "2.1.0",
				},
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationDetail{
					First8Digits: "12345678",
					First6Digits: "123456",
					Last4Digits:  "9876",
					ExpiryMonth:  "12",
					ExpiryYear:   "25",
					Fingerprint:  "fp-abc123",
				},
			},
			expected: &EncryptedCardAuthenticationResponse{
				CardID: "card-abc-def",
				CardInfo: EncryptedCardInformationResponse{
					First8Digits:     "12345678",
					First6Digits:     "123456",
					Last4Digits:      "9876",
					ExpiryMonth:      "12",
					ExpiryYear:       "25",
					HasAssociatedCVC: true,
					Fingerprint:      "fp-abc123",
				},
				Bin: Bin{
					UUID:          uuid.UUID{}, // Will be set to match
					BinNumber:     "123456",
					CardType:      "CREDIT",
					CardBrand:     "VISA",
					ConsumerType:  "CONSUMER",
					CardLevel:     "CLASSIC",
					IssuerName:    "Bank Central Asia",
					IssuerCountry: "ID",
					Currency:      "IDR",
					Status:        "ACTIVE",
					IsBlocked:     false,
					CreatedAt:     time.Time{}, // Will be set to match
					UpdatedAt:     time.Time{}, // Will be set to match
				},
				AuthenticationResponse: AuthenticationResponse{
					AcquirerTransactionID: "acquirer-txn-123",
					Amount:                "1000.50",
					Currency:              "IDR",
					Message:               "Authentication successful",
					SessionID:             "session-456",
					Status:                "AUTHENTICATED",
					AuthenticationURL: AuthenticationURLDetail{
						ActionURL:    "https://action.example.com",
						CreatedAt:    "2023-10-01T12:00:00Z",
						ThreeDSToken: "threeds-token-123",
						HTML:         "<form>3DS Form</form>",
						Method:       "POST",
						URL:          "https://3ds.example.com",
						Version:      "2.1.0",
					},
				},
			},
		},
		{
			name: "minimal_required_fields",
			request: &EncryptedCardAuthenticationRequest{
				CardID: "card-minimal",
			},
			authResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				Status: "PENDING",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					Method: "GET",
				},
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationDetail{
					First8Digits: "11111111",
					First6Digits: "111111",
					Last4Digits:  "1111",
					ExpiryMonth:  "01",
					ExpiryYear:   "30",
					Fingerprint:  "fp-minimal",
				},
			},
			expected: &EncryptedCardAuthenticationResponse{
				CardID: "card-minimal",
				CardInfo: EncryptedCardInformationResponse{
					First8Digits:     "11111111",
					First6Digits:     "111111",
					Last4Digits:      "1111",
					ExpiryMonth:      "01",
					ExpiryYear:       "30",
					HasAssociatedCVC: true,
					Fingerprint:      "fp-minimal",
				},
				Bin: Bin{
					BinNumber: "111111",
					CardType:  "DEBIT",
					Status:    "ACTIVE",
				},
				AuthenticationResponse: AuthenticationResponse{
					Status: "PENDING",
					AuthenticationURL: AuthenticationURLDetail{
						Method: "GET",
					},
				},
			},
		},
		{
			name: "empty_optional_fields",
			request: &EncryptedCardAuthenticationRequest{
				CardID: "card-empty-fields",
			},
			authResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{},
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationDetail{
					First8Digits: "",
					First6Digits: "",
					Last4Digits:  "",
					ExpiryMonth:  "",
					ExpiryYear:   "",
					Fingerprint:  "",
				},
			},
			expected: &EncryptedCardAuthenticationResponse{
				CardID: "card-empty-fields",
				CardInfo: EncryptedCardInformationResponse{
					First8Digits:     "",
					First6Digits:     "",
					Last4Digits:      "",
					ExpiryMonth:      "",
					ExpiryYear:       "",
					Fingerprint:      "",
					HasAssociatedCVC: true,
				},
				Bin: Bin{},
				AuthenticationResponse: AuthenticationResponse{
					AuthenticationURL: AuthenticationURLDetail{},
				},
			},
		},
		{
			name: "special_characters_in_fields",
			request: &EncryptedCardAuthenticationRequest{
				CardID: "card-special-chars-!@#$%",
			},
			authResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				Message: "Error: Special chars in message!",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					HTML: "<script>alert('test')</script>",
					URL:  "https://example.com?param=value&other=123",
				},
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationDetail{
					First8Digits: "00000000",
					First6Digits: "000000",
					Last4Digits:  "0000",
					ExpiryMonth:  "99",
					ExpiryYear:   "99",
					Fingerprint:  "fp-with-special-chars-!@#",
				},
			},
			expected: &EncryptedCardAuthenticationResponse{
				CardID: "card-special-chars-!@#$%",
				CardInfo: EncryptedCardInformationResponse{
					First8Digits:     "00000000",
					First6Digits:     "000000",
					Last4Digits:      "0000",
					ExpiryMonth:      "99",
					ExpiryYear:       "99",
					Fingerprint:      "fp-with-special-chars-!@#",
					HasAssociatedCVC: true,
				},
				Bin: Bin{
					BinNumber:     "000000",
					CardBrand:     "UNKNOWN",
					IssuerName:    "Bank with Special Chars & Co.",
					IssuerCountry: "XX",
				},
				AuthenticationResponse: AuthenticationResponse{
					Message: "Error: Special chars in message!",
					AuthenticationURL: AuthenticationURLDetail{
						HTML: "<script>alert('test')</script>",
						URL:  "https://example.com?param=value&other=123",
					},
				},
			},
		},
		{
			name: "blocked_card_bin",
			request: &EncryptedCardAuthenticationRequest{
				CardID: "card-blocked",
			},
			authResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
				Status:  "FAILED",
				Message: "Card is blocked",
				AuthenticationURL: creditcardCoreProcessorModel.AuthenticationURLDetail{
					Version: "1.0.0",
				},
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationDetail{
					First8Digits: "99999999",
					First6Digits: "999999",
					Last4Digits:  "9999",
					ExpiryMonth:  "12",
					ExpiryYear:   "99",
					Fingerprint:  "fp-blocked",
				},
			},
			expected: &EncryptedCardAuthenticationResponse{
				CardID: "card-blocked",
				CardInfo: EncryptedCardInformationResponse{
					First8Digits:     "99999999",
					First6Digits:     "999999",
					Last4Digits:      "9999",
					ExpiryMonth:      "12",
					ExpiryYear:       "99",
					HasAssociatedCVC: true,
					Fingerprint:      "fp-blocked",
				},
				Bin: Bin{
					UUID:          uuid.UUID{}, // Will be set to match
					BinNumber:     "999999",
					CardType:      "CREDIT",
					CardBrand:     "MASTERCARD",
					Status:        "BLOCKED",
					IsBlocked:     true,
					IssuerCountry: "US",
					CreatedAt:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:     time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
					DeletedAt:     sql.NullTime{Valid: true, Time: time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC)},
				},
				AuthenticationResponse: AuthenticationResponse{
					Status:  "FAILED",
					Message: "Card is blocked",
					AuthenticationURL: AuthenticationURLDetail{
						Version: "1.0.0",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToEncryptedCardAuthenticationResponse(tt.request, tt.authResponse)

			// Assert basic fields
			assert.Equal(t, tt.expected.CardID, result.CardID)

			// Assert CardInfo fields
			assert.Equal(t, tt.expected.CardInfo.First8Digits, result.CardInfo.First8Digits)
			assert.Equal(t, tt.expected.CardInfo.First6Digits, result.CardInfo.First6Digits)
			assert.Equal(t, tt.expected.CardInfo.Last4Digits, result.CardInfo.Last4Digits)
			assert.Equal(t, tt.expected.CardInfo.ExpiryMonth, result.CardInfo.ExpiryMonth)
			assert.Equal(t, tt.expected.CardInfo.ExpiryYear, result.CardInfo.ExpiryYear)
			assert.Equal(t, tt.expected.CardInfo.HasAssociatedCVC, result.CardInfo.HasAssociatedCVC)
			assert.Equal(t, tt.expected.CardInfo.Fingerprint, result.CardInfo.Fingerprint)

			// Assert AuthenticationResponse fields
			assert.Equal(t, tt.expected.AuthenticationResponse.AcquirerTransactionID, result.AuthenticationResponse.AcquirerTransactionID)
			assert.Equal(t, tt.expected.AuthenticationResponse.Amount, result.AuthenticationResponse.Amount)
			assert.Equal(t, tt.expected.AuthenticationResponse.Currency, result.AuthenticationResponse.Currency)
			assert.Equal(t, tt.expected.AuthenticationResponse.Message, result.AuthenticationResponse.Message)
			assert.Equal(t, tt.expected.AuthenticationResponse.SessionID, result.AuthenticationResponse.SessionID)
			assert.Equal(t, tt.expected.AuthenticationResponse.Status, result.AuthenticationResponse.Status)

			// Assert AuthenticationURL fields
			assert.Equal(t, tt.expected.AuthenticationResponse.AuthenticationURL.ActionURL, result.AuthenticationResponse.AuthenticationURL.ActionURL)
			assert.Equal(t, tt.expected.AuthenticationResponse.AuthenticationURL.CreatedAt, result.AuthenticationResponse.AuthenticationURL.CreatedAt)
			assert.Equal(t, tt.expected.AuthenticationResponse.AuthenticationURL.ThreeDSToken, result.AuthenticationResponse.AuthenticationURL.ThreeDSToken)
			assert.Equal(t, tt.expected.AuthenticationResponse.AuthenticationURL.HTML, result.AuthenticationResponse.AuthenticationURL.HTML)
			assert.Equal(t, tt.expected.AuthenticationResponse.AuthenticationURL.Method, result.AuthenticationResponse.AuthenticationURL.Method)
			assert.Equal(t, tt.expected.AuthenticationResponse.AuthenticationURL.URL, result.AuthenticationResponse.AuthenticationURL.URL)
			assert.Equal(t, tt.expected.AuthenticationResponse.AuthenticationURL.Version, result.AuthenticationResponse.AuthenticationURL.Version)
		})
	}
}
