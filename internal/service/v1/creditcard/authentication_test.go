package creditcard

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	card "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func TestCreateEncryptedCardAuthenticationLink(t *testing.T) {
	tests := []struct {
		name             string
		request          *card.EncryptedCardAuthenticationRequest
		mockAuthResponse *creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse
		mockAuthError    error
		expectedResult   *card.EncryptedCardAuthenticationResponse
		expectedError    error
	}{
		{
			name: "success",
			request: &card.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
				CardID:              "card-abc",
				CVC:                 "123",
				Amount:              1000.50,
				Fee:                 50.25,
				Currency:            "IDR",
			},
			mockAuthResponse: &creditcardCoreProcessorModel.EncryptedCardAuthenticationResponse{
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
				EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationDetail{
					First8Digits: "12345678",
					First6Digits: "123456",
					Last4Digits:  "9876",
					ExpiryMonth:  "12",
					ExpiryYear:   "25",
					Fingerprint:  "fingerprint-123",
					BinDetail: creditcardCoreProcessorModel.BinDetail{
						CardType:      "CREDIT",
						CardBrand:     "VISA",
						IssuerName:    "Bank ABC",
						IssuerCountry: "ID",
					},
				},
			},
			mockAuthError: nil,
			expectedResult: &card.EncryptedCardAuthenticationResponse{
				CardID: "card-abc",
				CardInfo: card.EncryptedCardInformationResponse{
					First8Digits:     "12345678",
					First6Digits:     "123456",
					Last4Digits:      "9876",
					ExpiryMonth:      "12",
					ExpiryYear:       "25",
					HasAssociatedCVC: true,
					Fingerprint:      "fingerprint-123",
				},
				Bin: card.Bin{
					CardType:      "CREDIT",
					CardBrand:     "VISA",
					IssuerName:    "Bank ABC",
					IssuerCountry: "ID",
				},
				AuthenticationResponse: card.AuthenticationResponse{
					AcquirerTransactionID: "acquirer-123",
					Amount:                "1000.50",
					Currency:              "IDR",
					Message:               "Success",
					SessionID:             "session-456",
					Status:                "AUTHENTICATED",
					AuthenticationURL: card.AuthenticationURLDetail{
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
			expectedError: nil,
		},
		{
			name: "error creating authentication link",
			request: &card.EncryptedCardAuthenticationRequest{
				PaymentID:           "payment-123",
				MerchantID:          "merchant-456",
				ClientTransactionID: "client-789",
				CardID:              "card-abc",
				CVC:                 "123",
				Amount:              1000.50,
				Fee:                 50.25,
				Currency:            "IDR",
			},
			mockAuthResponse: nil,
			mockAuthError:    errors.New("authentication service unavailable"),
			expectedResult:   nil,
			expectedError:    errors.New("authentication service unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCreditcardCoreProcessorRepo := &mockRepo.ICreditcardCoreProcessorRepository{}

			service := &CreditCardService{
				config:                      &config.Config{},
				logger:                      logger.NewSlogger(logger.Config{}),
				creditcardCoreProcessorRepo: mockCreditcardCoreProcessorRepo,
			}

			ctx := context.Background()

			mockCreditcardCoreProcessorRepo.On("CreateEncryptedCardAuthenticationLink", mock.Anything, mock.AnythingOfType("*creditcardCoreProcessorModel.EncryptedCardAuthenticationRequest")).
				Return(tt.mockAuthResponse, tt.mockAuthError)

			result, err := service.CreateEncryptedCardAuthenticationLink(ctx, tt.request)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.CardID, result.CardID)
				assert.Equal(t, tt.expectedResult.CardInfo.First8Digits, result.CardInfo.First8Digits)
				assert.Equal(t, tt.expectedResult.CardInfo.Last4Digits, result.CardInfo.Last4Digits)
				assert.Equal(t, tt.expectedResult.Bin.BinNumber, result.Bin.BinNumber)
				assert.Equal(t, tt.expectedResult.Bin.CardType, result.Bin.CardType)
				assert.Equal(t, tt.expectedResult.AuthenticationResponse.AcquirerTransactionID, result.AuthenticationResponse.AcquirerTransactionID)
				assert.Equal(t, tt.expectedResult.AuthenticationResponse.Status, result.AuthenticationResponse.Status)
				assert.Equal(t, tt.expectedResult.AuthenticationResponse.AuthenticationURL.ActionURL, result.AuthenticationResponse.AuthenticationURL.ActionURL)
			}

			mockCreditcardCoreProcessorRepo.AssertExpectations(t)
		})
	}
}

func TestAuthentication(t *testing.T) {
	cardRepo := mockRepo.NewICreditcardCoreProcessorRepository(t)

	service := New(nil, nil, nil, nil, nil, cardRepo)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *creditcardCoreProcessorModel.AuthenticationResponse
	}{
		{
			name: "ERROR: Some error", // NOSONAR
			setupMock: func() {
				cardRepo.On("Authentication", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				cardRepo.On("Authentication", mock.Anything, mock.Anything).Once().Return(&creditcardCoreProcessorModel.AuthenticationResponse{}, nil)
			},
			wantResult: &creditcardCoreProcessorModel.AuthenticationResponse{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.Authentication(t.Context(), creditcardCoreProcessorModel.AuthenticationRequest{})
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
