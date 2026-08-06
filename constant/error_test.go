package constant_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestMapErrorCodeToV2(t *testing.T) {
	tests := []struct {
		name     string
		v1Code   string
		expected string
	}{
		{"Invalid Credential", ErrCodeInvalidCredential, ErrCodeV2InvalidCredential},
		{"Field Format Invalid", ErrCodeFieldFormatInvalid, ErrCodeV2APIValidationError},
		{"Format Invalid", ErrCodeFormatInvalid, ErrCodeV2APIValidationError},
		{"Field Required", ErrCodeFieldRequired, ErrCodeV2APIValidationError},
		{"Unsupported Channel Code", ErrCodeUnsupportedChannelCode, ErrCodeV2APIValidationError},
		{"Unprocessable Entity", ErrCodeUnprocessableEntity, ErrCodeV2APIValidationError},
		{"Invalid Status Inquiry", ErrCodeInvalidStatusInquiry, ErrCodeV2APIValidationError},
		{"Amount Below Limit", ErrCodeAmountBelowLimit, ErrCodeV2FrequencyAboveLimit},
		{"Amount Above Limit", ErrCodeAmountAboveLimit, ErrCodeV2FrequencyAboveLimit},
		{"Daily Limit Reached", ErrCodeDailyLimitReached, ErrCodeV2FrequencyAboveLimit},
		{"Daily Payout Limit Reached", ErrCodeDailyPayoutLimitReached, ErrCodeV2FrequencyAboveLimit},
		{"Resource Already Exists", ErrCodeResourceAlreadyExists, ErrCodeV2DuplicateError},
		{"Payout In Process", ErrCodePayoutInProcess, ErrCodeV2DuplicateError},
		{"Duplicate Error", ErrCodeDuplicateError, ErrCodeV2DuplicateError},
		{"Service Unavailable", ErrCodeServiceUnavailable, ErrCodeV2ServiceUnavailable},
		{"Timeout", ErrCodeTimeout, ErrCodeV2GatewayTimeout},
		{"Resource Missing", ErrCodeResourceMissing, ErrCodeV2ResourceNotFound},
		{"Data Not Found", ErrCodeDataNotFound, ErrCodeV2NotFound},
		{"General Error", ErrCodeGeneral, ErrCodeV2InternalError},
		{"Balance Insufficient", ErrCodeBalanceInsufficient, ErrCodeV2RequestForbidden},
		{"Forbidden Access", ErrCodeForbiddenAccess, ErrCodeV2RequestForbidden},
		{"Bad Gateway", ErrCodeBadGateway, ErrCodeV2BadGateway},
		{"Frequency Above Limit", ErrCodeFrequencyAboveLimit, ErrCodeV2FrequencyAboveLimit},
		{"Invalid Card Number", ErrCodeInvalidCardNumber, ErrCodeV2InvalidCardNumber},
		{"Invalid Card Info", ErrCodeInvalidCardInfo, ErrCodeV2InvalidCardInfo},
		{"Card Decryption Failed", ErrCodeCardDecryption, ErrCodeV2CardDecryption},
		{"Unknown Error", "unknown_error_code", ErrCodeV2InternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := MapErrorCodeToV2(tt.v1Code)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestMapV2ErrorCodeToMessage(t *testing.T) {
	tests := []struct {
		name     string
		v2Code   string
		expected string
	}{
		// Line 830-831: Invalid Credential
		{"Invalid Credential", ErrCodeV2InvalidCredential, ErrMessageV2InvalidCredential},

		// Line 832-833: Resource Not Complete
		{"Resource Not Complete", ErrCodeV2ResourceNotComplete, ErrMessageV2ResourceNotComplete},

		// Line 834-835: API Validation Error
		{"API Validation Error", ErrCodeV2APIValidationError, ErrMessageV2APIValidationError},

		// Line 836-837: Request Forbidden
		{"Request Forbidden", ErrCodeV2RequestForbidden, ErrMessageV2RequestForbidden},

		// Line 838-839: Not Found
		{"Not Found", ErrCodeV2NotFound, ErrMessageV2NotFound},

		// Line 840-841: Resource Not Found
		{"Resource Not Found", ErrCodeV2ResourceNotFound, ErrMessageV2ResourceNotFound},

		// Line 842-843: Duplicate Error
		{"Duplicate Error", ErrCodeV2DuplicateError, ErrMessageV2DuplicateError},

		// Line 844-845: Idempotency Error
		{"Idempotency Error", ErrCodeV2IdempotencyError, ErrMessageV2IdempotencyError},

		// Line 846-847: Frequency Above Limit
		{"Frequency Above Limit", ErrCodeV2FrequencyAboveLimit, ErrMessageV2FrequencyAboveLimit},

		// Line 848-849: Database Error
		{"Database Error", ErrCodeV2DatabaseError, ErrMessageV2DatabaseError},

		// Line 850-851: Internal Error
		{"Internal Error", ErrCodeV2InternalError, ErrMessageV2InternalError},

		// Line 852-853: Bad Gateway
		{"Bad Gateway", ErrCodeV2BadGateway, ErrMessageV2BadGateway},

		// Line 854-855: Service Unavailable
		{"Service Unavailable", ErrCodeV2ServiceUnavailable, ErrMessageV2ServiceUnavailable},

		// Line 856-857: Gateway Timeout
		{"Gateway Timeout", ErrCodeV2GatewayTimeout, ErrMessageV2GatewayTimeout},

		// Line 858-859: Invalid Card Number
		{"Invalid Card Number", ErrCodeV2InvalidCardNumber, ErrMessageV2InvalidCardPaymentNumber},

		// Line 860-861: Card Decryption Error
		{"Card Decryption Error", ErrCodeV2CardDecryption, ErrMessageV2InvalidCardPaymentDecryption},

		// Line 862-863: Invalid Card Information
		{"Invalid Card Information", ErrCodeV2InvalidCardInfo, ErrMessageV2InvalidCardPaymentInformation},

		// Line 864-865: Default case
		{"Unknown Error", "unknown_error_code", "Unknown error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := MapV2ErrorCodeToMessage(tt.v2Code)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestErrStringRequest(t *testing.T) {
	err := NewErrStringRequest("ERROR_REQUEST", ErrCodeV2APIValidationError, "Invalid Number")
	assert.Equal(t, "ERROR_REQUEST | Invalid Number", err.Error())

	if et, ok := err.(*ErrStringRequest); assert.True(t, ok) {
		assert.Equal(t, "Invalid Number", et.Message())
		assert.Equal(t, ErrCodeV2APIValidationError, et.GetResponseCode())
	}
}
