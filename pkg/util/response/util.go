package response

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pdk/go/snap"
)

func HttpStatusErrorCode(errType string) (string, int) {
	switch errType {
	case HttpErrDatabase:
		return HttpStatusErrorDatabase, http.StatusInternalServerError
	case HttpErrThirdParty:
		return HttpStatusErrorThirdParty, http.StatusInternalServerError
	case HttpErrNotFound:
		return HttpStatusErrorNotFound, http.StatusNotFound
	case HttpErrUnauthorized:
		return HttpStatusErrorUnauthorized, http.StatusUnauthorized
	case HttpErrForbidden:
		return HttpStatusErrorForbidden, http.StatusForbidden
	case HttpErrDupCheck, HttpStatusErrorConflict:
		return HttpStatusErrorDuplicatedCheck, http.StatusConflict
	case HttpErrRequest:
		return HttpStatusErrorRequest, http.StatusBadRequest
	case HttpErrBadGateway:
		return HttpStatusErrorBadGateway, http.StatusBadGateway
	case HttpErrServiceUnavailable:
		return HttpStatusServiceUnavailable, http.StatusServiceUnavailable
	case HttpErrValidation:
		return HttpStatusErrorValidation, http.StatusBadRequest
	case HttpErrUnprocessableContent:
		return HttpStatusErrorUnprocessableContent, http.StatusUnprocessableEntity
	case HttpErrTooManyRequest:
		return HttpStatusErrorUnprocessableContent, http.StatusTooManyRequests
	case HttpErrDailyLimitReached:
		return HttpStatusErrorDailyLimitReached, http.StatusTooManyRequests
	case HttpErrRequestLimitExceeded:
		return HttpStatusErrorRequestLimitExceeded, http.StatusTooManyRequests
	case HttpErrResourceLocked:
		return HttpStatusErrorUnprocessableContent, http.StatusLocked
	case HttpErrRequestTimeout:
		return HttpStatusErrorrRequestTimeout, http.StatusGatewayTimeout
	case HttpErrConflict:
		return HttpStatusErrorConflict, http.StatusConflict
	default:
		return HttpStatusErrorInternal, http.StatusInternalServerError
	}
}

func HttpSnapErrorCode(errType string) (string, int) {
	switch errType {
	case HttpErrUnauthorized:
		return snap.SNAP_UNAUTHORIZED, http.StatusUnauthorized

	case SnapErrInvalidTokenB2B:
		return errType, http.StatusUnauthorized

	case SnapErrBankNotSupported, SnapErrInvalidVA, SnapErrTransactionNotFound, SnapErrInvalidAmount, SnapErrInvalidPartner, SnapErrInvalidMerchant:
		return errType, http.StatusNotFound

	case SnapErrFieldFormat, SnapErrRequiredField, SnapErrConflict:
		return errType, http.StatusBadRequest

	case SnapErrDuplicatePartnerRef:
		return errType, http.StatusConflict

	default:
		return snap.SNAP_INTERNAL_SERVER_ERROR, http.StatusInternalServerError
	}
}

func GetErrorType(errCode string) string {
	switch errCode {
	case HttpErrNotFound,
		HttpErrUnauthorized,
		HttpErrForbidden,
		HttpErrDupCheck,
		HttpErrInternal,
		HttpErrDatabase,
		HttpErrRequest,
		HttpErrDailyLimitReached,
		HttpStatusErrorConflict:
		return ErrTypeAPI
	case HttpErrValidation:
		return ErrTypeAPIValidation
	case HttpErrRequestTimeout,
		HttpErrBadGateway,
		HttpErrServiceUnavailable:
		return ErrTypeGateway
	case HttpErrThirdParty:
		return ErrTypePartner
	default:
		return ErrTypeUnknown
	}
}

func GetErrorSource(errType string) string {
	switch errType {
	case ErrTypeAPI, ErrTypeAPIValidation, ErrTypeIdempotency:
		return ErrorSourceUpstream
	case ErrTypePartner, ErrTypeGateway:
		return ErrorSourceDownstream
	default:
		return ErrorSourceSystem
	}
}

func GetErrorSourceByHttpErrType(errType string) string {
	switch errType {
	case HttpErrRequest,
		HttpErrUnauthorized,
		HttpErrForbidden,
		HttpErrValidation,
		HttpErrNotFound,
		HttpErrUnprocessableContent,
		HttpErrDupCheck,
		HttpErrConflict,
		HttpErrTooManyRequest,
		HttpErrDailyLimitReached,
		HttpErrResourceLocked:
		return ErrorSourceUpstream
	case HttpErrRequestLimitExceeded,
		HttpErrRequestTimeout,
		HttpErrBadGateway,
		HttpErrServiceUnavailable,
		HttpErrThirdParty:
		return ErrorSourceDownstream
	default:
		return ErrorSourceSystem
	}
}

func GetErrorDetailMessage(errMsg string) string {
	switch errMsg {
	case
		constant.ErrMsgXExternalIdAlreadyExists:
		return constant.ErrDetailMsgXExternalId
	default:
		return constant.ErrDetailMsgGeneralError
	}
}
