package response

import (
	"github.com/paper-indonesia/pdk/go/snap"
)

const (
	HttpStatusOK      string = "00"
	HttpStatusCreated string = "01"

	HttpStatusErrorInternal             string = "99"
	HttpStatusErrorDatabase             string = "98"
	HttpStatusErrorThirdParty           string = "50"
	HttpStatusErrorRequest              string = "40"
	HttpStatusErrorUnauthorized         string = "41"
	HttpStatusErrorValidation           string = "42"
	HttpStatusErrorForbidden            string = "43"
	HttpStatusErrorNotFound             string = "44"
	HttpStatusErrorUnprocessableContent string = "45"
	HttpStatusErrorDailyLimitReached    string = "46"
	HttpStatusErrorRequestLimitExceeded string = "47"
	HttpStatusErrorConflict             string = "48"
	HttpStatusErrorDuplicatedCheck      string = "49"
	HttpStatusErrorBadGateway           string = "52"
	HttpStatusServiceUnavailable        string = "53"
	HttpStatusErrorrRequestTimeout      string = "54"
	httpSnapCodeService                 string = "SNP-CR-"
)

const (
	HttpErrInternal             string = "ERROR_INTERNAL"
	HttpErrDatabase             string = "ERROR_DATABASE"
	HttpErrThirdParty           string = "ERROR_THIRD_PARTY"
	HttpErrRequest              string = "ERROR_REQUEST"
	HttpErrBadGateway           string = "ERROR_BAD_GATEWAY"
	HttpErrServiceUnavailable   string = "ERROR_SERVICE_UNAVAILABLE"
	HttpErrUnauthorized         string = "ERROR_UNAUTHORIZED"
	HttpErrForbidden            string = "ERROR_FORBIDDEN"
	HttpErrNotFound             string = "ERROR_NOT_FOUND"
	HttpErrDupCheck             string = "ERROR_DUPLICATE_CHECK"
	HttpErrValidation           string = "ERROR_VALIDATION"
	HttpErrUnprocessableContent string = "ERROR_UNPROCESSABLE_CONTENT"
	HttpErrTooManyRequest       string = "ERROR_TOO_MANY_REQUEST"
	HttpErrResourceLocked       string = "ERROR_RESOURCE_LOCKED"
	HttpErrRequestTimeout       string = "ERROR_REQUEST_TIMEOUT"
	HttpErrConflict             string = "ERROR_CONFLICT"
	HttpErrDailyLimitReached    string = "ERROR_DAILY_LIMIT_REACHED"
	HttpErrRequestLimitExceeded string = "REQUEST_LIMIT_EXCEEDED"
)

const (
	SnapErrBankNotSupported    string = snap.SNAP_BANK_NOT_SUPPORTED
	SnapErrFieldFormat         string = snap.SNAP_INVALID_FIELD
	SnapErrRequiredField       string = snap.SNAP_INVALID_MANDATORY
	SnapErrInvalidTokenB2B     string = snap.SNAP_INVALID_TOKEN_B2B
	SnapErrInvalidVA           string = snap.SNAP_INVALID_VA
	SnapErrTransactionNotFound string = snap.SNAP_TRANSACTION_NOT_FOUND
	SnapErrInvalidAmount       string = snap.SNAP_INVALID_AMOUNT
	SnapErrConflict            string = snap.SNAP_CONFLICT
	SnapErrDuplicatePartnerRef string = snap.SNAP_DUPLICATE_PARTNER_REFERENCE_NO
	SnapErrInvalidPartner      string = snap.SNAP_INVALID_PARTNER
	SnapErrInvalidMerchant     string = snap.SNAP_INVALID_MERCHANT
)

const (
	ErrTypeAPI           = "API_ERROR"
	ErrTypeAPIValidation = "API_VALIDATION_ERROR"
	ErrTypeIdempotency   = "IDEMPOTENCY_ERROR"
	ErrTypeUnknown       = "UNKNOWN"
	ErrTypePartner       = "PARTNER_ERROR"
	ErrTypeGateway       = "GATEWAY_ERROR"
)

const (
	ErrorSourceUpstream   = "UPSTREAM"
	ErrorSourceDownstream = "DOWNSTREAM"
	ErrorSourceSystem     = "SYSTEM"
)

func GetHttpCodeService(code string) string {
	return httpSnapCodeService + code
}

func GetHttpCodeServiceError(code string, errCode string) string {
	if errCode != "" {
		return "SNP_" + errCode
	}
	return "SNP_" + code
}
