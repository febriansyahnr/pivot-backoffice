package unifiedPaymentModel

import (
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fdsProcessor/fdsCommon"
	paymentCaptureModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/paymentCapture"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/proto/messages/callback"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChargeResponseSetFailureDetail(t *testing.T) {
	tests := []struct {
		name          string
		charge        *ChargeResponse
		expectedCode  string
		expectedMsg   string
		expectedRecom string
	}{
		{
			name: "card 3ds authentication failed",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: "failed",
						},
					},
				},
			},
			expectedCode:  "AUTHENTICATION_FAILED",
			expectedMsg:   "Card authentication failed. \n\nThe 3DS attempt was rejected by the issuer.",
			expectedRecom: "The cardholder should contact their issuer for clarification. The shopper can try again after resolving the issue with their issuer, or use another payment method.",
		},
		{
			name: "card declined by channel code 01",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationSuccess,
						},
						AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
							IssuerAuthorizationCode: "01",
						},
					},
				},
			},
			expectedCode:  "DECLINED_BY_CHANNEL",
			expectedMsg:   "Card payment failed. \n\nThe transaction was declined by the channel.",
			expectedRecom: "The cardholder should contact their issuer for clarification. The shopper can try again after resolving the issue with their issuer, or use another payment method.",
		},
		{
			name: "cancelled by user",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationFailed,
						},
						ResponseCode: &ChargePaymentMethodDetailCardResponseCode{
							GatewayCode: constant.CreditCardGatewayCodeAborted,
						},
					},
				},
			},
			expectedCode:  "CANCELLED_BY_USER",
			expectedMsg:   "Card payment failed. \n\nThe 3DS attempt was cancelled by the cardholder.",
			expectedRecom: "The shopper can try again or use another payment method.",
		},
		{
			name: "card expired code 54",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationSuccess,
						},
						AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
							IssuerAuthorizationCode: "54",
						},
					},
				},
			},
			expectedCode:  "DECLINED_BY_CHANNEL",
			expectedMsg:   "Card payment failed. \n\nThe transaction was declined by the issuer due to the card has already expired.",
			expectedRecom: "The shopper should try again with another valid card or use another payment method.",
		},
		{
			name: "invalid CVV code N7",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationSuccess,
						},
						AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
							IssuerAuthorizationCode: "N7",
						},
					},
				},
			},
			expectedCode:  "DECLINED_BY_CHANNEL",
			expectedMsg:   "Card payment failed. \n\nThe transaction was declined by the issuer due to the submitted CVV is invalid.",
			expectedRecom: "The shopper should try again with another valid card or use another payment method.",
		},
		{
			name: "invalid account code 14",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationSuccess,
						},
						AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
							IssuerAuthorizationCode: "14",
						},
					},
				},
			},
			expectedCode:  "INVALID_ACCOUNT",
			expectedMsg:   "Card payment failed. \n\nThe transaction was declined by the issuer due to the card being marked as invalid.",
			expectedRecom: "The shopper should try again with another valid card or use another payment method.",
		},
		{
			name: "suspected fraud code 04",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationSuccess,
						},
						AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
							IssuerAuthorizationCode: "04",
						},
					},
				},
			},
			expectedCode:  "SUSPECTED_FRAUD",
			expectedMsg:   "Card payment failed. \n\nThe transaction was declined by the issuer due to the card being marked as stolen or potential fraud.",
			expectedRecom: "The card was reported as lost, the shopper should be validated for authenticity and be referred to their issuer.",
		},
		{
			name: "suspected fraud account blocked code 34",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationSuccess,
						},
						AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
							IssuerAuthorizationCode: "34",
						},
					},
				},
			},
			expectedCode:  "SUSPECTED_FRAUD",
			expectedMsg:   "Card payment failed. \n\nThe transaction was declined by channel due to the account being blocked or suspected as fraud.",
			expectedRecom: "The channel has declined the transaction due to suspicion of fraud, the shopper should be validated for authenticity and be referred to their issuer.",
		},
		{
			name: "insufficient funds code 51",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationSuccess,
						},
						AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
							IssuerAuthorizationCode: "51",
						},
					},
				},
			},
			expectedCode:  "INSUFFICIENT_FUND",
			expectedMsg:   "Card payment failed. \n\nThe transaction was declined by the issuer due to credit limit or balance is not sufficient.",
			expectedRecom: "Insufficient funds in the cardholder's account. The shopper can try again after adding funds to their bank account, or use another payment method.",
		},
		{
			name: "channel unavailable code 19",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationSuccess,
						},
						AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
							IssuerAuthorizationCode: "19",
						},
					},
				},
			},
			expectedCode:  "CHANNEL_UNAVAILABLE",
			expectedMsg:   "Card payment failed. \n\nThe transaction failed due to the issuer being unavailable or having a system malfunction.",
			expectedRecom: "The issuing bank cannot be contacted. The shopper should try again or use another payment method.",
		},
		{
			name: "blocked by external fds",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{},
				},
				FdsRiskAssessment: &fdscommon.FdsRiskAssessment{
					Status: constant.FDS_STATUS_REJECTED,
				},
			},
			expectedCode:  "BLOCKED_BY_FDS",
			expectedMsg:   "Card payment failed. \n\nThe transaction was declined by FDS due to the transactions being categorized as high risk.",
			expectedRecom: "Verify and validate the transaction. The shopper can try again after the transaction has been validated.",
		},
		{
			name: "require review by fds",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{},
				},
				FdsRiskAssessment: &fdscommon.FdsRiskAssessment{
					Status: constant.FDS_STATUS_REVIEW,
				},
			},
			expectedCode:  "REQUIRE_REVIEW",
			expectedMsg:   "Card payment failed. \n\nThe transaction was deferred by FDS due to the transactions being categorized as suspicious.",
			expectedRecom: "Verify and validate the transaction. Approve or reject the transaction after it has been validated.",
		},
		{
			name: "charge expired",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusExpired,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: nil,
				},
			},
			expectedCode:  "CHARGE_EXPIRED",
			expectedMsg:   " charge failed due to the transaction time has exceeded the channel expiration time.",
			expectedRecom: "The shopper can try again or use another payment method.",
		},
		{
			name: "no failure when card is nil",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusFailed,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: nil,
				},
			},
		},
		{
			name: "no failure when status is not failed or expired",
			charge: &ChargeResponse{
				Status: constant.ChargeStatusWaitingForUserAction,
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsResult: constant.CreditCardAuthenticationSuccess,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.charge.SetFailureDetail()
			if tt.charge.FailureCode != tt.expectedCode {
				t.Errorf("expected failure code %q, got %q", tt.expectedCode, tt.charge.FailureCode)
			}
			if tt.charge.FailureMessage != tt.expectedMsg {
				t.Errorf("expected failure message %q, got %q", tt.expectedMsg, tt.charge.FailureMessage)
			}
			if tt.charge.Recommendation != tt.expectedRecom {
				t.Errorf("expected recommendation %q, got %q", tt.expectedRecom, tt.charge.Recommendation)
			}
		})
	}
}

func TestChargeResponseRemoveUnusedResponse(t *testing.T) {
	charge := ChargeResponse{
		ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
			Card: &ChargePaymentMethodDetailCard{
				MerchantCategoryCode: "0000",       // NOSONAR
				Description:          "APPROVED",   // NOSONAR
				SettlementDate:       "2026-07-06", // NOSONAR
				Device:               &ChargePaymentMethodDetailCardDevice{},
				Error:                &ChargePaymentMethodDetailCardError{},
				ResponseCode:         &ChargePaymentMethodDetailCardResponseCode{},

				AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
					CallbackTransactionID: "sample",

					// auto split properties
					AcsReference:         "sample-acs-ref",
					AcsTransactionID:     "sample-acs-trx",
					AuthenticationTime:   new(time.Now()),
					AuthenticationScheme: "auth-scheme",
					TransactionStatus:    "success",
					TransactionID:        "sample trx id",
				},
			},
		},
	}
	assert.NotNil(t, charge.Card.ResponseCode)

	charge.RemoveUnusedResponse()
	assert.Nil(t, charge.Card.ResponseCode)
	assert.Nil(t, charge.Card.Device)
	assert.Nil(t, charge.Card.Error)
	assert.Empty(t, charge.Card.MerchantCategoryCode)
	assert.Empty(t, charge.Card.Description)
	assert.Empty(t, charge.Card.SettlementDate)
	assert.Empty(t, charge.Card.AuthenticationResult.CallbackTransactionID)
	assert.Empty(t, charge.Card.AuthenticationResult.AcsReference)
	assert.Empty(t, charge.Card.AuthenticationResult.AcsTransactionID)
	assert.Empty(t, charge.Card.AuthenticationResult.AuthenticationTime)
	assert.Empty(t, charge.Card.AuthenticationResult.AuthenticationScheme)
	assert.Empty(t, charge.Card.AuthenticationResult.TransactionStatus)
	assert.Empty(t, charge.Card.AuthenticationResult.TransactionID)

	charge.Card = nil
	charge.RemoveUnusedResponse()

	charge2 := ChargeResponse{
		ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
			Ewallet: &ChargePaymentMethodDetailEwallet{
				AppRedirectURL:     "pivotpay://",
				WebRedirectURL:     "https://",
				ReferenceNo:        "REF",
				PartnerReferenceNo: "PARTNER-REFF",
				Channel:            "PIVOTPAY",
			},
		},
	}
	assert.NotEmpty(t, charge2.Ewallet.AppRedirectURL)
	assert.NotEmpty(t, charge2.Ewallet.WebRedirectURL)
	assert.NotEmpty(t, charge2.Ewallet.ReferenceNo)
	assert.NotEmpty(t, charge2.Ewallet.PartnerReferenceNo)
	assert.NotEmpty(t, charge2.Ewallet.Channel)
	charge2.RemoveUnusedResponse()

	assert.Empty(t, charge2.Ewallet.AppRedirectURL)
	assert.Empty(t, charge2.Ewallet.WebRedirectURL)
	assert.Empty(t, charge2.Ewallet.ReferenceNo)
	assert.Empty(t, charge2.Ewallet.PartnerReferenceNo)
	assert.NotEmpty(t, charge2.Ewallet.Channel)

	// Test ProcessorReferenceID field
	charge3 := ChargeResponse{
		ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
			ProcessorReferenceID: "PROC-REF-123",
		},
	}
	assert.NotEmpty(t, charge3.ChargePaymentMethodDetails.ProcessorReferenceID)
	charge3.RemoveUnusedResponse()
	assert.Empty(t, charge3.ChargePaymentMethodDetails.ProcessorReferenceID)

	// Test VirtualAccount BankReferenceNo field
	charge4 := ChargeResponse{
		ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
			VirtualAccount: &ChargePaymentMethodDetailVirtualAccount{
				BankReferenceNo: "BANK-REF-456",
				Channel:         "BNI",
			},
		},
	}
	assert.NotEmpty(t, charge4.VirtualAccount.BankReferenceNo)
	assert.NotEmpty(t, charge4.VirtualAccount.Channel)
	charge4.RemoveUnusedResponse()
	assert.Empty(t, charge4.VirtualAccount.BankReferenceNo)
	assert.NotEmpty(t, charge4.VirtualAccount.Channel)

	// Test Qr StoreID field
	charge5 := ChargeResponse{
		ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
			Qr: &ChargePaymentMethodDetailQr{
				StoreID:   "STORE-789",
				Acquirer:  "QRIS",
				QrContent: "qr-content",
			},
		},
	}
	assert.NotEmpty(t, charge5.Qr.StoreID)
	assert.NotEmpty(t, charge5.Qr.Acquirer)
	assert.NotEmpty(t, charge5.Qr.QrContent)
	charge5.RemoveUnusedResponse()
	assert.Empty(t, charge5.Qr.StoreID)
	assert.NotEmpty(t, charge5.Qr.Acquirer)
	assert.NotEmpty(t, charge5.Qr.QrContent)

	// Test when VirtualAccount is nil
	charge6 := ChargeResponse{
		ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
			VirtualAccount: nil,
		},
	}
	charge6.RemoveUnusedResponse()

	// Test when Qr is nil
	charge7 := ChargeResponse{
		ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
			Qr: nil,
		},
	}
	charge7.RemoveUnusedResponse()
}

func TestSetEncryptionKeyForCard(t *testing.T) {
	data := &UnifiedPaymentSessionResponse{
		AutoConfirm: false,
		Mode:        constant.UnifiedPaymentModeAPI,
		PaymentMethod: &PaymentMethod{
			Type: constant.UnifiedPaymentMethodCard,
		},
	}
	data.SetEncryptionKeyForCard()

	require.NotNil(t, data.EncryptionKey)
	require.NotEmpty(t, data.EncryptionKey.String())
	require.Empty(t, data.EncryptionKey.GetPrivateKey())
	require.Zero(t, data.EncryptionKey.GetSecretVersion())
}

func TestChargeResponse_ToSnapPayment(t *testing.T) {
	testCases := []struct {
		name                 string
		chargeResponse       *ChargeResponse
		expectedType         string
		expectedResultType   interface{}
		shouldReturnOriginal bool
	}{
		{
			name: "SUCCESS: VA with ChargePaymentMethodDetails filled",
			chargeResponse: &ChargeResponse{
				ID:                              "va-charge-123",
				PaymentSessionClientReferenceID: "va-client-ref-456",
				Status:                          constant.StatusSuccess,
				Amount: Amount{
					Currency: "IDR",
					Value:    10000.00,
				},
				PaidAt: &[]time.Time{time.Date(2025, 9, 13, 6, 15, 49, 0, time.UTC)}[0],
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					VirtualAccount: &ChargePaymentMethodDetailVirtualAccount{
						VirtualAccountNumber:  "7663010211990757",
						VirtualAccountName:    "HARSYA REMITINDO",
						Channel:               "PERMATA",
						VirtualAccountTrxType: "CLOSED_DYNAMIC",
						ExpiryAt:              time.Date(2025, 9, 13, 23, 15, 49, 0, time.UTC),
						BankReferenceNo:       "202509122315486196345812",
					},
				},
			},
			expectedType:       "VA",
			expectedResultType: &SnapVAResponse{},
		},
		{
			name: "SUCCESS: QRIS with ChargePaymentMethodDetails filled",
			chargeResponse: &ChargeResponse{
				ID:                              "charge-123",
				PaymentSessionID:                "session-789",
				PaymentSessionClientReferenceID: "client-ref-456",
				Status:                          constant.StatusPending,
				Amount: Amount{
					Currency: "IDR",
					Value:    25000,
				},
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Qr: &ChargePaymentMethodDetailQr{
						QrContent:                "qr-content-data",
						Acquirer:                 "QRIS_ACQUIRER",
						StoreID:                  "STORE123",
						QrType:                   "DYNAMIC",
						RetrievalReferenceNumber: "RRN123456",
						MerchantName:             "Test Merchant",
						ExpiryAt:                 time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
					},
				},
			},
			expectedType:       "QRIS",
			expectedResultType: &SnapQRISResponse{},
		},
		{
			name: "SUCCESS: Return original when no conversion criteria met",
			chargeResponse: &ChargeResponse{
				Status:                     "UNKNOWN",
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{},
			},
			expectedType:         "original",
			shouldReturnOriginal: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.chargeResponse.ToSnapPayment()

			if tc.shouldReturnOriginal {
				assert.Equal(t, tc.chargeResponse, result, "Should return original ChargeResponse when no conversion criteria met")
			} else {
				assert.IsType(t, tc.expectedResultType, result, "Should return correct type")

				switch tc.expectedType {
				case "VA":
					vaResponse := result.(*SnapVAResponse)

					// Check main fields
					assert.NotEmpty(t, vaResponse.TrxID, "VA result should have trxId")
					assert.NotEmpty(t, vaResponse.VirtualAccountNo, "VA result should have virtualAccountNo")
					assert.NotEmpty(t, vaResponse.VirtualAccountName, "VA result should have virtualAccountName")

					// Check amount structures
					assert.Equal(t, tc.chargeResponse.Amount.Currency, vaResponse.PaidAmount.Currency, "PaidAmount currency should match")
					assert.Equal(t, formatAmountValue(tc.chargeResponse.Amount.Value), vaResponse.PaidAmount.Value, "PaidAmount value should be formatted correctly")
					assert.Equal(t, tc.chargeResponse.Amount.Currency, vaResponse.TotalAmount.Currency, "TotalAmount currency should match")
					assert.Equal(t, formatAmountValue(tc.chargeResponse.Amount.Value), vaResponse.TotalAmount.Value, "TotalAmount value should be formatted correctly")

					// Check additionalInfo structure
					assert.Equal(t, tc.chargeResponse.PaymentSessionClientReferenceID, vaResponse.AdditionalInfo.ReferenceID, "ReferenceId should match")
					assert.Equal(t, tc.chargeResponse.Status, vaResponse.AdditionalInfo.PaymentStatus, "PaymentStatus should match")
					assert.NotEmpty(t, vaResponse.AdditionalInfo.Issuer, "Issuer should not be empty")
					assert.NotEmpty(t, vaResponse.AdditionalInfo.BankReferenceID, "BankReferenceId should not be empty")
					assert.NotEmpty(t, vaResponse.AdditionalInfo.VirtualAccountTrxType, "VirtualAccountTrxType should not be empty")
					assert.Equal(t, "CLOSED_DYNAMIC", vaResponse.AdditionalInfo.VirtualAccountTrxType, "VirtualAccountTrxType should match test data")

				case "QRIS":
					qrisResponse := result.(*SnapQRISResponse)

					// Check main fields
					assert.Equal(t, tc.chargeResponse.PaymentSessionID, qrisResponse.OriginalReferenceNo, "OriginalReferenceNo should match")
					assert.Equal(t, tc.chargeResponse.PaymentSessionClientReferenceID, qrisResponse.OriginalPartnerReferenceNo, "OriginalPartnerReferenceNo should match")
					assert.Equal(t, "00", qrisResponse.LatestTransactionStatus, "LatestTransactionStatus should match")
					assert.Equal(t, tc.chargeResponse.Status, qrisResponse.TransactionStatusDesc, "TransactionStatusDesc should match")

					// Check amount structure
					assert.Equal(t, tc.chargeResponse.Amount.Currency, qrisResponse.Amount.Currency, "Amount currency should match")
					assert.Equal(t, fmt.Sprintf("%.0f", tc.chargeResponse.Amount.Value), qrisResponse.Amount.Value, "Amount value should be formatted correctly")

					// Check additionalInfo structure
					assert.Equal(t, tc.chargeResponse.Status, qrisResponse.AdditionalInfo.PaymentStatus, "PaymentStatus should match")
					assert.NotEmpty(t, qrisResponse.AdditionalInfo.RRN, "RRN should not be empty")
					assert.NotEmpty(t, qrisResponse.AdditionalInfo.QrType, "QRType should not be empty")
					assert.NotEmpty(t, qrisResponse.AdditionalInfo.MerchantName, "MerchantName should not be empty")
				}
			}
		})
	}
}

func TestChargeResponseToPbChargeResponse(t *testing.T) {
	testCases := []struct {
		name   string
		charge *ChargeResponse
		verify func(*testing.T, *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse)
	}{
		{
			name: "SUCCESS: Complete charge response with all fields",
			charge: &ChargeResponse{
				ID:                              "charge-123",
				PaymentSessionID:                "session-456",
				PaymentSessionClientReferenceID: "client-ref-789",
				Amount: Amount{
					Value:    10000,
					Currency: "IDR",
				},
				StatementDescriptor: "Test Transaction",
				Status:              "SUCCESS",
				IsCaptured:          true,
				FailureCode:         "",
				FailureMessage:      "",
				Recommendation:      "",
				CreatedAt:           time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
				AuthorizedAmount: &Amount{
					Value:    10000,
					Currency: "IDR",
				},
				CapturedAmount: &Amount{
					Value:    10000,
					Currency: "IDR",
				},
				PaidAt: func() *time.Time { t := time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC); return &t }(),
				FdsRiskAssessment: &fdscommon.FdsRiskAssessment{
					Score:          decimal.NewFromFloat(1.5),
					Level:          "LOW",
					Recommendation: "APPROVE",
					Status:         "PASSED",
					EvaluatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{},
			},
			verify: func(t *testing.T, pb *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse) {
				assert.Equal(t, "charge-123", pb.Id)
				assert.Equal(t, "session-456", pb.PaymentSessionId)
				assert.Equal(t, "client-ref-789", pb.PaymentSessionClientReferenceId)
				assert.Equal(t, float64(10000), pb.Amount.Value)
				assert.Equal(t, "IDR", pb.Amount.Currency)
				assert.Equal(t, "Test Transaction", pb.StatementDescriptor)
				assert.Equal(t, "SUCCESS", pb.Status)
				assert.True(t, pb.IsCaptured)
				assert.NotNil(t, pb.CreatedAt)
				assert.NotNil(t, pb.UpdatedAt)
				assert.NotNil(t, pb.AuthorizedAmount)
				assert.NotNil(t, pb.CapturedAmount)
				assert.NotNil(t, pb.FdsRiskAssessment)
			},
		},
		{
			name: "SUCCESS: Charge with Ewallet payment method",
			charge: &ChargeResponse{
				ID:                              "charge-ewallet-123",
				PaymentSessionID:                "session-ewallet-456",
				PaymentSessionClientReferenceID: "client-ref-ewallet-789",
				Amount: Amount{
					Value:    50000,
					Currency: "IDR",
				},
				StatementDescriptor: "Ewallet Payment",
				Status:              "PENDING",
				IsCaptured:          false,
				CreatedAt:           time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 2, 1, 1, 0, 0, 0, time.UTC),
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Ewallet: &ChargePaymentMethodDetailEwallet{
						Channel: "GOPAY",
					},
				},
			},
			verify: func(t *testing.T, pb *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse) {
				assert.Equal(t, "charge-ewallet-123", pb.Id)
				assert.Equal(t, "PENDING", pb.Status)
				assert.False(t, pb.IsCaptured)
				assert.NotNil(t, pb.Ewallet)
				assert.Equal(t, "GOPAY", pb.Ewallet.Channel)
				assert.Nil(t, pb.Qr)
				assert.Nil(t, pb.VirtualAccount)
				assert.Nil(t, pb.Card)
			},
		},
		{
			name: "SUCCESS: Charge with QR payment method",
			charge: &ChargeResponse{
				ID:                              "charge-qr-123",
				PaymentSessionID:                "session-qr-456",
				PaymentSessionClientReferenceID: "client-ref-qr-789",
				Amount: Amount{
					Value:    25000,
					Currency: "IDR",
				},
				StatementDescriptor: "QR Payment",
				Status:              "SUCCESS",
				IsCaptured:          true,
				CreatedAt:           time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 3, 1, 1, 0, 0, 0, time.UTC),
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Qr: &ChargePaymentMethodDetailQr{
						Acquirer:                 "QRIS",
						QrContent:                "qr-string-data",
						QrUrl:                    "https://qr.example.com/123",
						RetrievalReferenceNumber: "RRN123456",
						IssuerName:               "Test Bank",
						ExpiryAt:                 time.Date(2024, 3, 1, 2, 0, 0, 0, time.UTC),
						MerchantName:             "Test Merchant",
					},
				},
			},
			verify: func(t *testing.T, pb *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse) {
				assert.Equal(t, "charge-qr-123", pb.Id)
				assert.Equal(t, "SUCCESS", pb.Status)
				assert.True(t, pb.IsCaptured)
				assert.NotNil(t, pb.Qr)
				assert.Equal(t, "QRIS", pb.Qr.Acquirer)
				assert.Equal(t, "qr-string-data", pb.Qr.QrContent)
				assert.Equal(t, "https://qr.example.com/123", pb.Qr.QrUrl)
				assert.Equal(t, "RRN123456", pb.Qr.RetrievalReferenceNumber)
				assert.Equal(t, "Test Bank", pb.Qr.IssuerName)
				assert.Equal(t, "Test Merchant", pb.Qr.MerchantName)
				assert.NotNil(t, pb.Qr.ExpiryAt)
				assert.Nil(t, pb.Ewallet)
				assert.Nil(t, pb.VirtualAccount)
				assert.Nil(t, pb.Card)
			},
		},
		{
			name: "SUCCESS: Charge with Virtual Account payment method",
			charge: &ChargeResponse{
				ID:                              "charge-va-123",
				PaymentSessionID:                "session-va-456",
				PaymentSessionClientReferenceID: "client-ref-va-789",
				Amount: Amount{
					Value:    75000,
					Currency: "IDR",
				},
				StatementDescriptor: "VA Payment",
				Status:              "PENDING",
				IsCaptured:          false,
				CreatedAt:           time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 4, 1, 1, 0, 0, 0, time.UTC),
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					VirtualAccount: &ChargePaymentMethodDetailVirtualAccount{
						Channel:              "BNI",
						VirtualAccountNumber: "1234567890123456",
						VirtualAccountName:   "Test VA Name",
						ExpiryAt:             time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			verify: func(t *testing.T, pb *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse) {
				assert.Equal(t, "charge-va-123", pb.Id)
				assert.Equal(t, "PENDING", pb.Status)
				assert.False(t, pb.IsCaptured)
				assert.NotNil(t, pb.VirtualAccount)
				assert.Equal(t, "BNI", pb.VirtualAccount.Channel)
				assert.Equal(t, "1234567890123456", pb.VirtualAccount.VirtualAccountNumber)
				assert.Equal(t, "Test VA Name", pb.VirtualAccount.VirtualAccountName)
				assert.NotNil(t, pb.VirtualAccount.ExpiryAt)
				assert.Equal(t, "", pb.VirtualAccount.BankReferenceNo)
				assert.Nil(t, pb.Ewallet)
				assert.Nil(t, pb.Qr)
				assert.Nil(t, pb.Card)
			},
		},
		{
			name: "SUCCESS: Charge with Virtual Account payment method and BankReferenceNo",
			charge: &ChargeResponse{
				ID:                              "charge-va-with-bank-ref",
				PaymentSessionID:                "session-va-bank-ref",
				PaymentSessionClientReferenceID: "client-ref-va-bank-ref",
				Amount: Amount{
					Value:    80000,
					Currency: "IDR",
				},
				StatementDescriptor: "VA Payment with BankRef",
				Status:              "SUCCESS",
				IsCaptured:          true,
				CreatedAt:           time.Date(2024, 4, 10, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 4, 10, 1, 0, 0, 0, time.UTC),
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					VirtualAccount: &ChargePaymentMethodDetailVirtualAccount{
						Channel:               "BCA",
						VirtualAccountNumber:  "9876543210987654",
						VirtualAccountName:    "Test VA with BankRef",
						VirtualAccountTrxType: "CLOSED",
						ExpiryAt:              time.Date(2024, 4, 11, 0, 0, 0, 0, time.UTC),
						BankReferenceNo:       "BANKREF123456789",
					},
				},
			},
			verify: func(t *testing.T, pb *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse) {
				assert.Equal(t, "charge-va-with-bank-ref", pb.Id)
				assert.Equal(t, "SUCCESS", pb.Status)
				assert.True(t, pb.IsCaptured)
				assert.NotNil(t, pb.VirtualAccount)
				assert.Equal(t, "BCA", pb.VirtualAccount.Channel)
				assert.Equal(t, "9876543210987654", pb.VirtualAccount.VirtualAccountNumber)
				assert.Equal(t, "Test VA with BankRef", pb.VirtualAccount.VirtualAccountName)
				assert.Equal(t, "CLOSED", pb.VirtualAccount.VirtualAccountTrxType)
				assert.Equal(t, "BANKREF123456789", pb.VirtualAccount.BankReferenceNo)
				assert.NotNil(t, pb.VirtualAccount.ExpiryAt)
				assert.Nil(t, pb.Ewallet)
				assert.Nil(t, pb.Qr)
				assert.Nil(t, pb.Card)
			},
		},
		{
			name: "SUCCESS: Charge with Virtual Account payment method and empty BankReferenceNo",
			charge: &ChargeResponse{
				ID:                              "charge-va-empty-bank-ref",
				PaymentSessionID:                "session-va-empty-ref",
				PaymentSessionClientReferenceID: "client-ref-va-empty-ref",
				Amount: Amount{
					Value:    90000,
					Currency: "IDR",
				},
				StatementDescriptor: "VA Payment empty BankRef",
				Status:              "PENDING",
				IsCaptured:          false,
				CreatedAt:           time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 4, 15, 1, 0, 0, 0, time.UTC),
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					VirtualAccount: &ChargePaymentMethodDetailVirtualAccount{
						Channel:              "MANDIRI",
						VirtualAccountNumber: "5555666677778888",
						VirtualAccountName:   "Test VA empty BankRef",
						ExpiryAt:             time.Date(2024, 4, 16, 0, 0, 0, 0, time.UTC),
						BankReferenceNo:      "",
					},
				},
			},
			verify: func(t *testing.T, pb *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse) {
				assert.Equal(t, "charge-va-empty-bank-ref", pb.Id)
				assert.Equal(t, "PENDING", pb.Status)
				assert.False(t, pb.IsCaptured)
				assert.NotNil(t, pb.VirtualAccount)
				assert.Equal(t, "MANDIRI", pb.VirtualAccount.Channel)
				assert.Equal(t, "5555666677778888", pb.VirtualAccount.VirtualAccountNumber)
				assert.Equal(t, "Test VA empty BankRef", pb.VirtualAccount.VirtualAccountName)
				assert.Equal(t, "", pb.VirtualAccount.BankReferenceNo)
				assert.NotNil(t, pb.VirtualAccount.ExpiryAt)
				assert.Nil(t, pb.Ewallet)
				assert.Nil(t, pb.Qr)
				assert.Nil(t, pb.Card)
			},
		},
		{
			name: "SUCCESS: Charge with Card payment method",
			charge: &ChargeResponse{
				ID:                              "charge-card-123",
				PaymentSessionID:                "session-card-456",
				PaymentSessionClientReferenceID: "client-ref-card-789",
				Amount: Amount{
					Value:    100000,
					Currency: "IDR",
				},
				StatementDescriptor: "Card Payment",
				Status:              "SUCCESS",
				IsCaptured:          true,
				CreatedAt:           time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 5, 1, 1, 0, 0, 0, time.UTC),
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						First6:           "411111",
						First8:           "41111111",
						Last4:            "1111",
						Fingerprint:      "fingerprint123",
						CardHolderName:   "JOHN DOE",
						ACSURL:           "https://acs.example.com",
						BankMerchantID:   "BANK_MERCHANT_123",
						SaveForFutureUse: util.ValueToPtr(true),
						ExpMonth:         "12",
						ExpYear:          "25",
						BinInformations: ChargePaymentMethodDetailBinInformation{
							Type:        "CREDIT",
							IssuingBank: "Test Bank",
							Brand:       "VISA",
							Country:     "ID",
						},
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							ThreeDsVersion: "2.0",
							ThreeDsResult:  "SUCCESS",
							ThreeDsMethod:  "CHALLENGE",
							EciCode:        "05",
						},
						AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
							AcquirerReferenceNumber:  "ARN123456",
							RetrievalReferenceNumber: "RRN654321",
							Stan:                     "123456",
							AvsResult:                "M",
							CvvResult:                "M",
							AuthorizedAmount: Amount{
								Value:    100000,
								Currency: "IDR",
							},
							IssuerAuthorizationCode: "00",
						},
						ResponseCode: &ChargePaymentMethodDetailCardResponseCode{
							GatewayCode:           "SUCCESS",
							GatewayRecommendation: "APPROVE",
						},
					},
				},
			},
			verify: func(t *testing.T, pb *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse) {
				assert.Equal(t, "charge-card-123", pb.Id)
				assert.Equal(t, "SUCCESS", pb.Status)
				assert.True(t, pb.IsCaptured)
				assert.NotNil(t, pb.Card)
				assert.Equal(t, "411111", pb.Card.First6)
				assert.Equal(t, "41111111", pb.Card.First8)
				assert.Equal(t, "1111", pb.Card.Last4)
				assert.Equal(t, "JOHN DOE", pb.Card.CardHolderName)
				assert.Equal(t, "https://acs.example.com", pb.Card.AcsUrl)
				assert.Equal(t, "BANK_MERCHANT_123", pb.Card.BankMerchantId)
				assert.Equal(t, "12", pb.Card.ExpMonth)
				assert.Equal(t, "25", pb.Card.ExpYear)

				assert.NotNil(t, pb.Card.BinInformations)
				assert.Equal(t, "CREDIT", pb.Card.BinInformations.Type)
				assert.Equal(t, "Test Bank", pb.Card.BinInformations.IssuingBank)
				assert.Equal(t, "VISA", pb.Card.BinInformations.Brand)
				assert.Equal(t, "ID", pb.Card.BinInformations.Country)

				assert.NotNil(t, pb.Card.AuthenticationResult)
				assert.Equal(t, "2.0", pb.Card.AuthenticationResult.ThreeDsVersion)
				assert.Equal(t, "SUCCESS", pb.Card.AuthenticationResult.ThreeDsResult)
				assert.Equal(t, "CHALLENGE", pb.Card.AuthenticationResult.ThreeDsMethod)
				assert.Equal(t, "05", pb.Card.AuthenticationResult.EciCode)

				assert.NotNil(t, pb.Card.AuthorizationResult)
				assert.Equal(t, "ARN123456", pb.Card.AuthorizationResult.AcquirerReferenceNumber)
				assert.Equal(t, "RRN654321", pb.Card.AuthorizationResult.RetrievalReferenceNumber)
				assert.Equal(t, "123456", pb.Card.AuthorizationResult.Stan)
				assert.Equal(t, "M", pb.Card.AuthorizationResult.AvsResult)
				assert.Equal(t, "M", pb.Card.AuthorizationResult.CvvResult)
				assert.Equal(t, "00", pb.Card.AuthorizationResult.IssuerAuthorizationCode)
				assert.NotNil(t, pb.Card.AuthorizationResult.AuthorizedAmount)

				assert.NotNil(t, pb.Card.ResponseCode)
				assert.Equal(t, "SUCCESS", pb.Card.ResponseCode.GatewayCode)
				assert.Equal(t, "APPROVE", pb.Card.ResponseCode.GatewayRecommendation)

				assert.Nil(t, pb.Ewallet)
				assert.Nil(t, pb.Qr)
				assert.Nil(t, pb.VirtualAccount)
			},
		},
		{
			name: "SUCCESS: Card with nil ExpMonth and ExpYear",
			charge: &ChargeResponse{
				ID:                              "charge-card-nil-exp",
				PaymentSessionID:                "session-card-nil",
				PaymentSessionClientReferenceID: "client-ref-card-nil",
				Amount: Amount{
					Value:    50000,
					Currency: "IDR",
				},
				StatementDescriptor: "Card Payment Nil Exp",
				Status:              "SUCCESS",
				IsCaptured:          true,
				CreatedAt:           time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:           time.Date(2024, 7, 1, 1, 0, 0, 0, time.UTC),
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						First6:      "411111",
						Last4:       "1111",
						Fingerprint: "fingerprint456",
						ExpMonth:    "",
						ExpYear:     "",
						BinInformations: ChargePaymentMethodDetailBinInformation{
							Type:        "DEBIT",
							IssuingBank: "Test Bank 2",
							Brand:       "MASTERCARD",
							Country:     "US",
						},
					},
				},
			},
			verify: func(t *testing.T, pb *pb.UnifiedPaymentV2CallbackRequest_ChargeResponse) {
				assert.Equal(t, "charge-card-nil-exp", pb.Id)
				assert.NotNil(t, pb.Card)
				assert.Equal(t, "411111", pb.Card.First6)
				assert.Equal(t, "1111", pb.Card.Last4)
				assert.Equal(t, "", pb.Card.ExpMonth)
				assert.Equal(t, "", pb.Card.ExpYear)
				assert.Nil(t, pb.Card.AuthenticationResult)
				assert.Nil(t, pb.Card.AuthorizationResult)
				assert.Nil(t, pb.Card.ResponseCode)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.charge.ToPbChargeResponse()
			tc.verify(t, result)
		})
	}
}

func TestFormatAmountValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    float64
		expected string
	}{
		{
			name:     "whole number - no decimals",
			value:    10000.00,
			expected: "10000",
		},
		{
			name:     "zero - no decimals",
			value:    0.00,
			expected: "0",
		},
		{
			name:     "large whole number - no decimals",
			value:    1000000.00,
			expected: "1000000",
		},
		{
			name:     "amount with cents",
			value:    10000.50,
			expected: "10000.50",
		},
		{
			name:     "amount with single decimal",
			value:    10000.10,
			expected: "10000.10",
		},
		{
			name:     "small amount with decimals",
			value:    1.99,
			expected: "1.99",
		},
		{
			name:     "amount with two decimals",
			value:    12345.67,
			expected: "12345.67",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatAmountValue(tc.value)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestChargeResponseSetCaptureHistories(t *testing.T) {
	testCases := []struct {
		name            string
		charge          *ChargeResponse
		paymentCaptures []*paymentCaptureModel.PaymentCapture
		verify          func(*testing.T, *ChargeResponse)
	}{
		{
			name:   "SUCCESS: Set single capture history",
			charge: &ChargeResponse{},
			paymentCaptures: []*paymentCaptureModel.PaymentCapture{
				{
					ID:                     "capture-1",
					PaymentID:              "payment-1",
					ProcessorReferenceID:   util.ValueToPtr("proc-ref-1"),
					Status:                 "SUCCESS",
					ReleaseRemainingAmount: true,
					Currency:               "IDR",
					Amount:                 100000,
					CreatedAt:              time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:              time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
				},
			},
			verify: func(t *testing.T, c *ChargeResponse) {
				assert.Len(t, c.CaptureHistories, 1)
				assert.Equal(t, "capture-1", c.CaptureHistories[0].ID)
				assert.Equal(t, "SUCCESS", c.CaptureHistories[0].Status)
				assert.Equal(t, "IDR", c.CaptureHistories[0].Currency)
				assert.Equal(t, float64(100000), c.CaptureHistories[0].CapturedAmount)
				assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), c.CaptureHistories[0].CreatedAt)
			},
		},
		{
			name:   "SUCCESS: Set multiple capture histories",
			charge: &ChargeResponse{},
			paymentCaptures: []*paymentCaptureModel.PaymentCapture{
				{
					ID:                     "capture-1",
					PaymentID:              "payment-1",
					ProcessorReferenceID:   util.ValueToPtr("proc-ref-1"),
					Status:                 "SUCCESS",
					ReleaseRemainingAmount: true,
					Currency:               "IDR",
					Amount:                 50000,
					CreatedAt:              time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:              time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
				},
				{
					ID:                     "capture-2",
					PaymentID:              "payment-1",
					ProcessorReferenceID:   util.ValueToPtr("proc-ref-2"),
					Status:                 "SUCCESS",
					ReleaseRemainingAmount: false,
					Currency:               "IDR",
					Amount:                 30000,
					CreatedAt:              time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					UpdatedAt:              time.Date(2024, 1, 2, 1, 0, 0, 0, time.UTC),
				},
				{
					ID:                     "capture-3",
					PaymentID:              "payment-1",
					ProcessorReferenceID:   util.ValueToPtr("proc-ref-3"),
					Status:                 "PENDING",
					ReleaseRemainingAmount: false,
					Currency:               "IDR",
					Amount:                 20000,
					CreatedAt:              time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
					UpdatedAt:              time.Date(2024, 1, 3, 1, 0, 0, 0, time.UTC),
				},
			},
			verify: func(t *testing.T, c *ChargeResponse) {
				assert.Len(t, c.CaptureHistories, 3)

				// Verify first capture
				assert.Equal(t, "capture-1", c.CaptureHistories[0].ID)
				assert.Equal(t, "SUCCESS", c.CaptureHistories[0].Status)
				assert.Equal(t, "IDR", c.CaptureHistories[0].Currency)
				assert.Equal(t, float64(50000), c.CaptureHistories[0].CapturedAmount)
				assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), c.CaptureHistories[0].CreatedAt)

				// Verify second capture
				assert.Equal(t, "capture-2", c.CaptureHistories[1].ID)
				assert.Equal(t, "SUCCESS", c.CaptureHistories[1].Status)
				assert.Equal(t, "IDR", c.CaptureHistories[1].Currency)
				assert.Equal(t, float64(30000), c.CaptureHistories[1].CapturedAmount)
				assert.Equal(t, time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), c.CaptureHistories[1].CreatedAt)

				// Verify third capture
				assert.Equal(t, "capture-3", c.CaptureHistories[2].ID)
				assert.Equal(t, "PENDING", c.CaptureHistories[2].Status)
				assert.Equal(t, "IDR", c.CaptureHistories[2].Currency)
				assert.Equal(t, float64(20000), c.CaptureHistories[2].CapturedAmount)
				assert.Equal(t, time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), c.CaptureHistories[2].CreatedAt)
			},
		},
		{
			name:            "SUCCESS: Empty payment captures slice - no capture histories set",
			charge:          &ChargeResponse{},
			paymentCaptures: []*paymentCaptureModel.PaymentCapture{},
			verify: func(t *testing.T, c *ChargeResponse) {
				assert.Nil(t, c.CaptureHistories)
			},
		},
		{
			name:            "SUCCESS: Nil payment captures slice - no capture histories set",
			charge:          &ChargeResponse{},
			paymentCaptures: nil,
			verify: func(t *testing.T, c *ChargeResponse) {
				assert.Nil(t, c.CaptureHistories)
			},
		},
		{
			name: "SUCCESS: Overwrite existing capture histories",
			charge: &ChargeResponse{
				CaptureHistories: []*paymentCaptureModel.CaptureHistoryResponse{
					{
						ID:             "old-capture",
						Status:         "OLD",
						Currency:       "USD",
						CapturedAmount: 999,
						CreatedAt:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			paymentCaptures: []*paymentCaptureModel.PaymentCapture{
				{
					ID:                     "new-capture",
					PaymentID:              "payment-2",
					ProcessorReferenceID:   util.ValueToPtr("new-ref"),
					Status:                 "SUCCESS",
					ReleaseRemainingAmount: true,
					Currency:               "EUR",
					Amount:                 150000,
					CreatedAt:              time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:              time.Date(2024, 5, 1, 1, 0, 0, 0, time.UTC),
				},
			},
			verify: func(t *testing.T, c *ChargeResponse) {
				assert.Len(t, c.CaptureHistories, 1)
				assert.Equal(t, "new-capture", c.CaptureHistories[0].ID)
				assert.Equal(t, "SUCCESS", c.CaptureHistories[0].Status)
				assert.Equal(t, "EUR", c.CaptureHistories[0].Currency)
				assert.Equal(t, float64(150000), c.CaptureHistories[0].CapturedAmount)
				assert.Equal(t, time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), c.CaptureHistories[0].CreatedAt)
			},
		},
		{
			name:   "SUCCESS: Capture with different statuses",
			charge: &ChargeResponse{},
			paymentCaptures: []*paymentCaptureModel.PaymentCapture{
				{
					ID:        "capture-success",
					Status:    "SUCCESS",
					Currency:  "IDR",
					Amount:    100000,
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "capture-pending",
					Status:    "PENDING",
					Currency:  "IDR",
					Amount:    50000,
					CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "capture-failed",
					Status:    "FAILED",
					Currency:  "IDR",
					Amount:    25000,
					CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
				},
			},
			verify: func(t *testing.T, c *ChargeResponse) {
				assert.Len(t, c.CaptureHistories, 3)
				assert.Equal(t, "SUCCESS", c.CaptureHistories[0].Status)
				assert.Equal(t, "PENDING", c.CaptureHistories[1].Status)
				assert.Equal(t, "FAILED", c.CaptureHistories[2].Status)
			},
		},
		{
			name:   "SUCCESS: Capture with different currencies",
			charge: &ChargeResponse{},
			paymentCaptures: []*paymentCaptureModel.PaymentCapture{
				{
					ID:        "capture-idr",
					Status:    "SUCCESS",
					Currency:  "IDR",
					Amount:    100000,
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "capture-usd",
					Status:    "SUCCESS",
					Currency:  "USD",
					Amount:    10,
					CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			verify: func(t *testing.T, c *ChargeResponse) {
				assert.Len(t, c.CaptureHistories, 2)
				assert.Equal(t, "IDR", c.CaptureHistories[0].Currency)
				assert.Equal(t, float64(100000), c.CaptureHistories[0].CapturedAmount)
				assert.Equal(t, "USD", c.CaptureHistories[1].Currency)
				assert.Equal(t, float64(10), c.CaptureHistories[1].CapturedAmount)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.charge.SetCaptureHistories(tc.paymentCaptures)
			tc.verify(t, tc.charge)
		})
	}
}

func TestSetPaymentURLForAPIMode(t *testing.T) {
	tests := []struct {
		name           string
		payment        *UnifiedPaymentSessionResponse
		expectedPayURL string
	}{
		{
			name: "skip when mode is not API",
			payment: &UnifiedPaymentSessionResponse{
				Mode: "REDIRECT",
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Card: &ChargePaymentMethodDetailCard{
								ACSURL: "https://acs.example.com",
							},
						},
					},
				},
			},
			expectedPayURL: "",
		},
		{
			name: "skip when payment method is nil",
			payment: &UnifiedPaymentSessionResponse{
				Mode:          constant.UnifiedPaymentModeAPI,
				PaymentMethod: nil,
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Card: &ChargePaymentMethodDetailCard{
								ACSURL: "https://acs.example.com",
							},
						},
					},
				},
			},
			expectedPayURL: "",
		},
		{
			name: "skip when charge details is empty",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ChargeDetails: []*ChargeResponse{},
			},
			expectedPayURL: "",
		},
		{
			name: "skip when QRIS and payment type is MULTIPLE",
			payment: &UnifiedPaymentSessionResponse{
				Mode:        constant.UnifiedPaymentModeAPI,
				PaymentType: constant.UnifiedPaymentTypeMultiple,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodQris,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Qr: &ChargePaymentMethodDetailQr{
								QrContent: "qr-content",
							},
						},
					},
				},
			},
			expectedPayURL: "",
		},
		{
			name: "skip when subsequent recurring payment",
			payment: &UnifiedPaymentSessionResponse{
				Mode:                       constant.UnifiedPaymentModeAPI,
				RecurringID:                "recurring-123",
				InitiateFirstAuthorization: false,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Card: &ChargePaymentMethodDetailCard{
								ACSURL: "https://acs.example.com",
							},
						},
					},
				},
			},
			expectedPayURL: "",
		},
		{
			name: "skip when payment method is not CARD or EWALLET (VA)",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: "VIRTUAL_ACCOUNT",
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							VirtualAccount: &ChargePaymentMethodDetailVirtualAccount{
								Channel:              "BNI",
								VirtualAccountNumber: "1234567890",
							},
						},
					},
				},
			},
			expectedPayURL: "",
		},
		{
			name: "skip when payment method is not CARD or EWALLET (QR)",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodQris,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Qr: &ChargePaymentMethodDetailQr{
								QrContent: "qr-content",
							},
						},
					},
				},
			},
			expectedPayURL: "",
		},
		{
			name: "set payment URL from card ACSURL",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Card: &ChargePaymentMethodDetailCard{
								ACSURL: "https://acs.example.com/3ds",
							},
						},
					},
				},
			},
			expectedPayURL: "https://acs.example.com/3ds",
		},
		{
			name: "skip when charge has no card or ewallet",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							VirtualAccount: &ChargePaymentMethodDetailVirtualAccount{
								Channel: "BNI",
							},
						},
					},
				},
			},
			expectedPayURL: "",
		},
		{
			name: "skip ewallet DANA - do not replace URL",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Ewallet: &ChargePaymentMethodDetailEwallet{
								Channel:        constant.UnifiedPaymentEWalletDanaAcquirer,
								WebRedirectURL: "https://dana.example.com",
							},
						},
					},
				},
			},
			expectedPayURL: "",
		},
		{
			name: "set payment URL from ewallet web redirect",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Ewallet: &ChargePaymentMethodDetailEwallet{
								Channel:        "GOPAY",
								WebRedirectURL: "https://goplay.example.com/pay",
							},
						},
					},
				},
			},
			expectedPayURL: "https://goplay.example.com/pay",
		},
		{
			name: "set payment URL from ewallet app redirect for ShopeePay",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodEWallet,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Ewallet: &ChargePaymentMethodDetailEwallet{
								Channel:        constant.UnifiedPaymentEWalletShopeePayAcquirer,
								WebRedirectURL: "https://shopeepay.example.com/web",
								AppRedirectURL: "shopeepay://app/redirect",
							},
						},
					},
				},
			},
			expectedPayURL: "shopeepay://app/redirect",
		},
		{
			name: "first recurring payment with card should set payment URL",
			payment: &UnifiedPaymentSessionResponse{
				Mode:                       constant.UnifiedPaymentModeAPI,
				RecurringID:                "recurring-123",
				InitiateFirstAuthorization: true,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Card: &ChargePaymentMethodDetailCard{
								ACSURL: "https://acs.example.com/recurring",
							},
						},
					},
				},
			},
			expectedPayURL: "https://acs.example.com/recurring",
		},
		{
			name: "card takes priority when charge has both card and ewallet",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Card: &ChargePaymentMethodDetailCard{
								ACSURL: "https://acs.example.com/card",
							},
							Ewallet: &ChargePaymentMethodDetailEwallet{
								Channel:        "GOPAY",
								WebRedirectURL: "https://goplay.example.com/pay",
							},
						},
					},
				},
			},
			expectedPayURL: "https://acs.example.com/card",
		},
		{
			name: "uses first charge detail with card",
			payment: &UnifiedPaymentSessionResponse{
				Mode: constant.UnifiedPaymentModeAPI,
				PaymentMethod: &PaymentMethod{
					Type: constant.UnifiedPaymentMethodCard,
				},
				ChargeDetails: []*ChargeResponse{
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{},
					},
					{
						ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
							Card: &ChargePaymentMethodDetailCard{
								ACSURL: "https://acs.example.com/second",
							},
						},
					},
				},
			},
			expectedPayURL: "https://acs.example.com/second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.payment.SetPaymentURLForAPIMode()
			assert.Equal(t, tt.expectedPayURL, tt.payment.PaymentUrl)
		})
	}
}

func TestBuildAutoSplitDetails(t *testing.T) {
	tests := []struct {
		name     string
		charges  []ChargeResponse
		expected *AutoSplitDetails
	}{
		{
			name:    "empty charges returns processing status",
			charges: []ChargeResponse{},
			expected: &AutoSplitDetails{
				Status:                      constant.AutoSplitStatusProcessing,
				NumberOfCharges:             0,
				NumberOfSuccessfulCharges:   0,
				NumberOfInProcessCharges:    0,
				NumberOfFailedCharges:       0,
				TotalSuccessfulChargeAmount: nil,
				TotalFailedChargeAmount:     nil,
				TotalInProcessChargeAmount:  nil,
				ChargesDetails:              []ChargeResponse{},
			},
		},
		{
			name: "single processing charge",
			charges: []ChargeResponse{
				{
					Status: constant.ChargeStatusProcessing,
					Amount: Amount{Value: 100000, Currency: "IDR"},
				},
			},
			expected: &AutoSplitDetails{
				Status:                      constant.AutoSplitStatusProcessing,
				NumberOfCharges:             1,
				NumberOfSuccessfulCharges:   0,
				NumberOfInProcessCharges:    1,
				NumberOfFailedCharges:       0,
				TotalSuccessfulChargeAmount: nil,
				TotalFailedChargeAmount:     nil,
				TotalInProcessChargeAmount:  &Amount{Value: 100000, Currency: "IDR"},
			},
		},
		{
			name: "single successful charge",
			charges: []ChargeResponse{
				{
					Status: constant.ChargeStatusSuccess,
					Amount: Amount{Value: 250000, Currency: "IDR"},
				},
			},
			expected: &AutoSplitDetails{
				Status:                      constant.AutoSplitStatusSuccess,
				NumberOfCharges:             1,
				NumberOfSuccessfulCharges:   1,
				TotalSuccessfulChargeAmount: &Amount{Value: 250000, Currency: "IDR"},
			},
		},
		{
			name: "single failed charge",
			charges: []ChargeResponse{
				{
					Status: constant.ChargeStatusFailed,
					Amount: Amount{Value: 50000, Currency: "IDR"},
				},
			},
			expected: &AutoSplitDetails{
				Status:                  constant.AutoSplitStatusFailed,
				NumberOfCharges:         1,
				NumberOfFailedCharges:   1,
				TotalFailedChargeAmount: &Amount{Value: 50000, Currency: "IDR"},
			},
		},
		{
			name: "partial success - mix of successful and failed charges",
			charges: []ChargeResponse{
				{
					Status: constant.ChargeStatusSuccess,
					Amount: Amount{Value: 2000000000, Currency: "IDR"},
				},
				{
					Status: constant.ChargeStatusSuccess,
					Amount: Amount{Value: 1000000000, Currency: "IDR"},
				},
				{
					Status: constant.ChargeStatusFailed,
					Amount: Amount{Value: 500000000, Currency: "IDR"},
				},
			},
			expected: &AutoSplitDetails{
				Status:                      constant.AutoSplitStatusPartialSuccess,
				NumberOfCharges:             3,
				NumberOfSuccessfulCharges:   2,
				NumberOfInProcessCharges:    0,
				NumberOfFailedCharges:       1,
				TotalSuccessfulChargeAmount: &Amount{Value: 3000000000, Currency: "IDR"},
				TotalFailedChargeAmount:     &Amount{Value: 500000000, Currency: "IDR"},
			},
		},
		{
			name: "all charges in various pending states count as in-process",
			charges: []ChargeResponse{
				{
					Status: constant.ChargeStatusProcessing,
					Amount: Amount{Value: 100000, Currency: "IDR"},
				},
				{
					Status: constant.ChargeStatusWaitingForCapture,
					Amount: Amount{Value: 200000, Currency: "IDR"},
				},
				{
					Status: constant.ChargeStatusWaitingForUserAction,
					Amount: Amount{Value: 300000, Currency: "IDR"},
				},
			},
			expected: &AutoSplitDetails{
				Status:                     constant.AutoSplitStatusProcessing,
				NumberOfCharges:            3,
				NumberOfInProcessCharges:   3,
				TotalInProcessChargeAmount: &Amount{Value: 600000, Currency: "IDR"},
			},
		},
		{
			name: "expired charges count as failed",
			charges: []ChargeResponse{
				{
					Status: constant.ChargeStatusExpired,
					Amount: Amount{Value: 75000, Currency: "IDR"},
				},
			},
			expected: &AutoSplitDetails{
				Status:                  constant.AutoSplitStatusFailed,
				NumberOfCharges:         1,
				NumberOfFailedCharges:   1,
				TotalFailedChargeAmount: &Amount{Value: 75000, Currency: "IDR"},
			},
		},
		{
			name: "all successful across many charges",
			charges: []ChargeResponse{
				{Status: constant.ChargeStatusSuccess, Amount: Amount{Value: 100000, Currency: "IDR"}},
				{Status: constant.ChargeStatusSuccess, Amount: Amount{Value: 200000, Currency: "IDR"}},
				{Status: constant.ChargeStatusSuccess, Amount: Amount{Value: 300000, Currency: "IDR"}},
				{Status: constant.ChargeStatusSuccess, Amount: Amount{Value: 400000, Currency: "IDR"}},
			},
			expected: &AutoSplitDetails{
				Status:                      constant.AutoSplitStatusSuccess,
				NumberOfCharges:             4,
				NumberOfSuccessfulCharges:   4,
				TotalSuccessfulChargeAmount: &Amount{Value: 1000000, Currency: "IDR"},
			},
		},
		{
			name: "successful + in-process = processing status",
			charges: []ChargeResponse{
				{Status: constant.ChargeStatusSuccess, Amount: Amount{Value: 500000, Currency: "IDR"}},
				{Status: constant.ChargeStatusProcessing, Amount: Amount{Value: 300000, Currency: "IDR"}},
			},
			expected: &AutoSplitDetails{
				Status:                      constant.AutoSplitStatusProcessing,
				NumberOfCharges:             2,
				NumberOfSuccessfulCharges:   1,
				NumberOfInProcessCharges:    1,
				TotalSuccessfulChargeAmount: &Amount{Value: 500000, Currency: "IDR"},
				TotalInProcessChargeAmount:  &Amount{Value: 300000, Currency: "IDR"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildAutoSplitDetails(tt.charges)

			assert.Equal(t, tt.expected.Status, result.Status)
			assert.Equal(t, tt.expected.NumberOfCharges, result.NumberOfCharges)
			assert.Equal(t, tt.expected.NumberOfSuccessfulCharges, result.NumberOfSuccessfulCharges)
			assert.Equal(t, tt.expected.NumberOfInProcessCharges, result.NumberOfInProcessCharges)
			assert.Equal(t, tt.expected.NumberOfFailedCharges, result.NumberOfFailedCharges)
			assert.Equal(t, tt.expected.TotalSuccessfulChargeAmount, result.TotalSuccessfulChargeAmount)
			assert.Equal(t, tt.expected.TotalFailedChargeAmount, result.TotalFailedChargeAmount)
			assert.Equal(t, tt.expected.TotalInProcessChargeAmount, result.TotalInProcessChargeAmount)
			assert.Equal(t, len(tt.charges), len(result.ChargesDetails))
		})
	}
}
