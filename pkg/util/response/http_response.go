package response

import (
	"context"
	"encoding/json"
	e "errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/xbCoreProcessor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/dictionary"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validation"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"github.com/paper-indonesia/pdk/go/snap"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
)

func SendGeneralResponseOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := Response{
		Code: HttpStatusOK,
		Data: data,
	}

	encodeJsonErr := json.NewEncoder(w).Encode(resp)
	if encodeJsonErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendGeneralResponseError(w http.ResponseWriter, errMessage error) {
	errType, err := errors.ExtractError(errMessage)
	code, statusCode := HttpStatusErrorCode(errType)

	errorMsg := "unknown error"
	if err != nil {
		errorMsg = err.Error()
	}

	resp := Response{
		Code:  code,
		Error: errorMsg,
	}

	var validatorError validator.ValidationErrors
	if e.As(errMessage, &validatorError) {
		errs := make(map[string]interface{})
		for _, e := range validatorError {
			errs[e.Field()] = e.Translate(validatorExt.GetTranslator())
		}
		resp.Error = errs

	} else {
		if errs, ok := errMessage.(*validation.Fields); ok {
			resp.Code = errs.Code()
			resp.Error, statusCode = errs, errs.StatusCode()
		}
	}

	w.Header().Set(constant.HeaderContentType, constant.MIMEApplicationJSON)
	w.WriteHeader(statusCode)

	encodeJsonErr := json.NewEncoder(w).Encode(resp)
	if encodeJsonErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendOpenApiResponseOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := OpenApiResponse{
		Code:    HttpStatusOK,
		Data:    data,
		Message: "Success",
	}

	encodeJsonErr := json.NewEncoder(w).Encode(resp)
	if encodeJsonErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendOpenApiResponseSuccess(w http.ResponseWriter, statusCode int, msg string, data interface{}) {
	w.Header().Set(
		constant.HeaderContentType, constant.MIMEApplicationJSON,
	)
	w.WriteHeader(statusCode)

	resp := OpenApiResponse{
		Code:    HttpStatusOK,
		Data:    data,
		Message: msg,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendOpenApiResponseError(w http.ResponseWriter, errMessage error) {
	errType, err := errors.ExtractError(errMessage)
	code, statusCode := HttpStatusErrorCode(errType)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorMsg := "unknown error"
	if err != nil {
		errorMsg = err.Error()
	}

	errorObject := &OpenApiError{
		Type:           GetErrorType(errType), // TODO: map later
		Message:        errorMsg,              // TODO: map later
		Recommendation: "",                    // TODO: map later
	}

	resp := OpenApiResponse{
		Code:    code,
		Error:   errorObject,
		Message: err.Error(),
	}

	encodeJsonErr := json.NewEncoder(w).Encode(resp)
	if encodeJsonErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendOpenApiResponsePaginationOK(w http.ResponseWriter, data interface{}, meta interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := OpenApiResponse{
		Code:       HttpStatusOK,
		Message:    "Success",
		Data:       data,
		Pagination: &meta,
	}

	encodeJsonErr := json.NewEncoder(w).Encode(resp)
	if encodeJsonErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendApiResponseOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := ApiResponse{
		Code:    HttpStatusOK,
		Data:    data,
		Message: "OK",
	}

	encodeJsonErr := json.NewEncoder(w).Encode(resp)
	if encodeJsonErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendApiResponseSuccess(w http.ResponseWriter, statusCode int, msg string, data interface{}) {
	w.Header().Set(
		constant.HeaderContentType, constant.MIMEApplicationJSON,
	)
	w.WriteHeader(statusCode)

	resp := ApiResponse{
		Code:    HttpStatusOK,
		Data:    data,
		Message: msg,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendApiResponseCreated(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err := json.NewEncoder(w).Encode(ApiResponse{
		Code:    HttpStatusCreated,
		Data:    data,
		Message: "Created",
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendApiResponseError(ctx context.Context, w http.ResponseWriter, errMessage error) {
	errType, err := errors.ExtractError(errMessage)
	code, statusCode := HttpStatusErrorCode(errType)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorObject := &ApiError{
		Type: GetErrorType(errType),
	}

	if traceID, ok := ctx.Value(pdkConst.CtxTraceIdKey).(string); ok {
		errorObject.TraceId = traceID
	}

	var errDetails = make([]ApiErrorDetail, 0)
	var validatorError validator.ValidationErrors
	if e.As(errMessage, &validatorError) {
		err = constant.ErrInvalidValidation
		for _, vErr := range validatorError {
			vErrMessage := vErr.Translate(validatorExt.GetTranslator())

			if dictionary.Dict != nil {
				switch vErr.Tag() {
				case "min":
					vErrMessage = dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationErrValidationMin, vErr.Field(), vErr.Param())
				case "required", "required_with":
					vErrMessage = dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationErrValidationRequired, vErr.Field())
				case "email":
					vErrMessage = dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationErrValidationEmail, vErr.Field())
				case "numeric":
					vErrMessage = dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationErrValidationNumeric, vErr.Field())
				case "len":
					vErrMessage = dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationErrValidationLength, vErr.Field())
				case "oneof":
					vErrMessage = dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationErrValidationOneOf, vErr.Field(), strings.Join(strings.Split(vErr.Param(), " "), ", "))
				case "iso_8601_datetime":
					vErrMessage = dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationErrValidationIso8601Datetime, vErr.Field(), vErr.Tag())
				case "luhn":
					vErrMessage = dictionary.Dict.SetDictionaryMessage(ctx, dictionary.TranslationErrValidationLuhn, vErr.Field())
				}
			}

			errDetail := ApiErrorDetail{
				Field:   vErr.Field(),
				Message: vErrMessage,
			}

			errDetails = append(errDetails, errDetail)
		}
	}
	errorObject.Details = errDetails

	translationCode := ""
	translationMessage := "unknown error"
	if err != nil {
		translationCode = dictionary.GetAPITranslationCodeByError(err.Error())
		translationMessage = err.Error()
	}
	if translationCode != "" && dictionary.Dict != nil {
		translationMessage = dictionary.Dict.SetDictionaryMessage(ctx, translationCode)
	}

	resp := ApiResponse{
		Code:    code,
		Error:   errorObject,
		Message: translationMessage,
	}

	xbCoreProcessorResp := extractXbCoreProcessorResponse(err.Error())
	if xbCoreProcessorResp != nil {
		resp.Message = xbCoreProcessorResp.Message
		if len(xbCoreProcessorResp.ErrorDetails) > 0 {
			errDetails = make([]ApiErrorDetail, 0)
			for _, vErr := range xbCoreProcessorResp.ErrorDetails {
				errDetail := ApiErrorDetail{
					Field:   util.SnakeToCamel(vErr.Field),
					Message: vErr.Message,
				}
				errDetails = append(errDetails, errDetail)
			}
			errorObject.Details = errDetails
			resp.Error = errorObject
		}
	}

	encodeJsonErr := json.NewEncoder(w).Encode(resp)
	if encodeJsonErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendApiResponsePaginationOK(w http.ResponseWriter, data interface{}, meta interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := ApiResponse{
		Code:       HttpStatusOK,
		Message:    "OK",
		Data:       data,
		Pagination: &meta,
	}

	encodeJsonErr := json.NewEncoder(w).Encode(resp)
	if encodeJsonErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func SendOpenApiSnapResponseError(ctx context.Context, w http.ResponseWriter, err error) {
	if traceId, ok := ctx.Value(pdkConst.CtxTraceIdKey).(string); ok {
		w.Header().Set(constant.HeaderXRequestId, traceId)
	}
	w.Header().Set(constant.HeaderContentType, constant.MIMEApplicationJSON)

	snapService, ok := ctx.Value(constant.CtxSnapApiName).(string)
	if !ok {
		snapService = snap.SNAP_SERVICE_B2B
	}

	etyp, unwrapErr := errors.ExtractError(err)
	snapCode, statusCode := HttpSnapErrorCode(etyp)

	code, msg := snap.GenerateResponseCode(snapCode, snapService)

	w.WriteHeader(statusCode)

	if statusCode < http.StatusInternalServerError && unwrapErr != nil {
		msg = unwrapErr.Error()
	}
	_ = json.NewEncoder(w).Encode(&OpenApiSnapResp{ResponseCode: code, ResponseMessage: msg})
}

func SendOpenApiSnapResponseOK(ctx context.Context, w http.ResponseWriter, response interface{}) {
	if traceId, ok := ctx.Value(pdkConst.CtxTraceIdKey).(string); ok {
		w.Header().Set(constant.HeaderXRequestId, traceId)
	}
	w.Header().Set(constant.HeaderContentType, constant.MIMEApplicationJSON)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func platformWhitelistOldResponseFormat(merchantId string) bool {
	ff := ffcontext.NewEvaluationContext(merchantId)
	ff.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)
	oldResponse, _ := ffclient.BoolVariation(constant.FeatureFlagPlatformWhitelistOldResponseFormat, ff, false)

	return oldResponse
}

type OpenApiErrorResponseType1 func(w http.ResponseWriter, err error)
type OpenApiErrorResponseType2 func(ctx context.Context, w http.ResponseWriter, err error)

func SendOpenApiNonSnapResponseError(ctx context.Context, w http.ResponseWriter, errMessage error) {

	if merchantId, _ := ctx.Value(constant.CtxMerchantIDKey).(string); merchantId != "" && platformWhitelistOldResponseFormat(merchantId) {

		if _, ok := ctx.Value(constant.CtxCustomErrorResponse).(OpenApiErrorResponseType2); ok {
			ctx = context.WithValue(ctx, constant.CtxErrorInfo, nil) // Remove new error mapping

		} else if fn, ok := ctx.Value(constant.CtxCustomErrorResponse).(OpenApiErrorResponseType1); ok {
			fn(w, errMessage)
			return
		}
	}
	if errorInfo, _ := ctx.Value(constant.CtxErrorInfo).(error); errorInfo != nil {
		errMessage = errorInfo
	}

	errType, err := errors.ExtractError(errMessage)
	code, statusCode := HttpStatusErrorCode(errType)

	w.Header().Set(constant.HeaderContentType, constant.ContentTypeApplicationJson)
	w.WriteHeader(statusCode)

	errorMsg := "unknown error"
	if err != nil {
		errorMsg = err.Error()
	}
	response := HandleDetailedError(ctx, code, errorMsg, errType)
	useV2ErrorCode, _ := ctx.Value(constant.CtxUseV2ErrorCode).(bool)
	if useV2ErrorCode {
		response.Code = constant.MapErrorCodeToV2(response.Code)
		response.Message = constant.MapV2ErrorCodeToMessage(response.Code)
	}
	response = parseErrorMessage(response, errMessage)
	useErrorSource, _ := ctx.Value(constant.CtxUseErrorSource).(bool)
	if (useV2ErrorCode || useErrorSource) && response.Error != nil {
		response.Error.Source = GetErrorSourceByHttpErrType(errType)
	}

	if traceID, ok := ctx.Value(pdkConst.CtxTraceIdKey).(string); ok {
		response.Error.TraceId = traceID
	}
	encodeJsonErr := json.NewEncoder(w).Encode(response)
	if encodeJsonErr != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func parseErrorMessage(response *OpenApiErrorNonSnap, err error) *OpenApiErrorNonSnap {
	switch typ := err.(type) {
	default:
		return response

	case *constant.ErrRequiredField:
		return createResponse(constant.ErrCodeFieldRequired, constant.ErrMsgMandatoryField, ErrTypeAPI, typ.GetFieldName(), typ.Message())

	case *constant.ErrInvalidFieldFmt:
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, typ.GetFieldName(), typ.Message())

	case *constant.ErrResourceNotFound:
		return createResponse(constant.ErrCodeResourceMissing, typ.Message(), ErrTypeGateway, "", typ.Message())

	case *constant.ErrInvalidPayload:
		return createResponse(constant.ErrCodeV2APIValidationError, typ.Message(), ErrTypeAPI, "", typ.Message())

	case *constant.ErrFieldValidation:
		errs, ok := typ.OriginalError().(validator.ValidationErrors)
		if !ok {
			return response
		}
		response = &OpenApiErrorNonSnap{
			Code:    constant.ErrCodeV2APIValidationError,
			Message: constant.ErrMessageV2APIValidationError,
			Error:   &ApiError{Type: ErrTypeAPI},
		}
		for _, msg := range errs {
			fieldName := strings.ToLower(string(msg.Field()[0])) + msg.Field()[1:]
			response.Error.Details = append(response.Error.Details, ApiErrorDetail{
				Field:   fieldName,
				Message: translateFieldValidationError(fieldName, msg.Param(), msg.Error()),
			})
		}
		return response

	case *constant.ErrInternalPartner:
		errType := ErrTypeAPI
		if !strings.Contains(typ.Error(), HttpErrRequest) {
			errType = ErrTypePartner
		}
		message, detail := typ.GetResponseMessage()
		return createResponse(typ.GetResponseCode(), message, errType, "", detail)

	case constant.GeneralError:
		message, detail := typ.GetResponseMessage()
		return createResponse(typ.GetResponseCode(), message, ErrTypeAPI, "", detail)

	case *constant.ErrStringRequest:
		return createResponse(typ.GetResponseCode(), typ.Message(), ErrTypeAPI, "", typ.Message())
	}
}

func translateFieldValidationError(field, param, originalError string) string {
	if strings.Contains(originalError, "'required' tag") ||
		strings.Contains(originalError, "'required_with' tag") ||
		strings.Contains(originalError, "'required_if' tag") ||
		strings.Contains(originalError, "'required_without' tag") {
		return fmt.Sprintf(constant.ErrDetailMsgMandatoryField, field)

	} else if strings.Contains(originalError, "'min' tag") {
		return fmt.Sprintf(constant.ErrDetailMsgMakeSureValueIsAboveMin, field)

	} else if strings.Contains(originalError, "'max' tag") {
		return fmt.Sprintf(constant.ErrDetailMsgMakeSureValueIsBelowMax, field)

	} else if strings.Contains(originalError, "'email' tag") ||
		strings.Contains(originalError, "'uuid' tag") ||
		strings.Contains(originalError, "'iso_8601_datetime' tag") ||
		strings.Contains(originalError, "'iso4217' tag") ||
		strings.Contains(originalError, "'datetime' tag") {
		return fmt.Sprintf(constant.ErrDetailMsgFormatField, field)

	} else if strings.Contains(originalError, "'maxChar' tag") {
		return fmt.Sprintf(constant.ErrDetailMsgCharExceedTheMax, field, param)

	} else if strings.Contains(originalError, "'oneof' tag") ||
		strings.Contains(originalError, "'gtecsfield' tag") ||
		strings.Contains(originalError, "'gtefield' tag") ||
		strings.Contains(originalError, "'excluded_if' tag") {
		return fmt.Sprintf(constant.ErrDetailMsgMakeSureValueIsCorrect, field)
	}
	return originalError
}

// handleUnifiedPaymentValidationError handles unified payment specific validation errors in HandleDetailedError
// Returns nil if no specific handling is needed, allowing fallback to general error handling
func handleUnifiedPaymentValidationError(code, message string) *OpenApiErrorNonSnap {
	if code != HttpStatusErrorRequest {
		return nil
	}

	// Handle threeDsMethod validation error
	if strings.Contains(message, "ThreeDsMethod") && strings.Contains(message, "'oneof' tag") {
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, "threeDsMethod", "invalid threeDsMethod value")
	}

	// Handle captureMethod validation error
	if strings.Contains(message, "CaptureMethod") && strings.Contains(message, "'oneof' tag") {
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, "captureMethod", "invalid captureMethod value")
	}

	// Handle authenticationTime datetime validation error
	if strings.Contains(message, "AuthenticationTime") && strings.Contains(message, "'datetime' tag") {
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, "authenticationTime", "invalid authenticationTime format, must be RFC3339 (e.g., 2023-12-31T23:59:59Z)")
	}

	if strings.Contains(message, "ThreeDsInfo") {
		detailMessage := "invalid threeDsInfo format"
		if strings.Contains(message, "'required' tag") {
			detailMessage = "missing required threeDsInfo parameters for external 3DS flow"
		}
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, "", detailMessage)
	}

	// Return nil to indicate no specific handling, use general error handling
	return nil
}

func HandleDetailedError(ctx context.Context, code, message, errorType string) *OpenApiErrorNonSnap {
	exposeUnmappingRequestError, ok := ctx.Value(constant.CtxExposeUnmappingRequestError).(bool)
	if !ok {
		exposeUnmappingRequestError = false
	}

	if strings.EqualFold(code, HttpStatusErrorRequest) && exposeUnmappingRequestError {
		// card error confirmation
		if code == HttpStatusErrorRequest && message == constant.ErrInvalidCardNumber.Error() {
			return createResponse(constant.ErrCodeV2InvalidCardNumber, message, ErrTypeAPI, "", message)
		}

		if code == HttpStatusErrorRequest && message == constant.ErrForeignCardBillingInformationMissing.Error() {
			return createResponse(constant.ErrCodeV2ForeignCardBillingInfoMissing, message, ErrTypeAPI, "", message)
		}
		if code == HttpStatusErrorRequest && message == constant.ErrFailedToDecryptCardPayment.Error() {
			return createResponse(constant.ErrCodeV2CardDecryption, message, ErrTypeAPI, "", message)
		}

		if code == HttpStatusErrorRequest && message == constant.ErrInvalidCardInformation.Error() {
			return createResponse(constant.ErrCodeV2InvalidCardInfo, message, ErrTypeAPI, "", message)
		}

		if code == HttpStatusErrorRequest && message == constant.ErrUnifiedPaymentMetadataSizeLimitExceeded.Error() {
			return createResponse(constant.ErrCodeFieldFormatInvalid, message, ErrTypeAPI, "metadata", fmt.Sprintf("metadata exceeds maximum allowed length of %d characters", constant.UnifiedPaymentMaxMetadataLength))
		}

		// Check for unified payment specific validation errors
		if unifiedPaymentResp := handleUnifiedPaymentValidationError(code, message); unifiedPaymentResp != nil {
			return unifiedPaymentResp
		}

		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, "", message)
	} else if strings.EqualFold(code, HttpStatusErrorUnprocessableContent) && exposeUnmappingRequestError {
		return createResponse(constant.ErrCodeUnprocessableEntity, constant.ErrMsgUnprocessableEntity, ErrTypeAPI, "", message)
	}

	xbCoreProcessorResp := extractXbCoreProcessorResponse(message)
	switch {
	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "required"):
		fieldName := util.SnakeToCamel((xbCoreProcessorResp.ErrorDetails)[0].Field)
		msg := fmt.Sprintf(constant.ErrDetailMsgMandatoryField, fieldName)
		return createResponse(constant.ErrCodeFieldRequired, constant.ErrMsgMandatoryField, ErrTypeAPI, fieldName, msg)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "routing code is required when country code is provided"):
		return createResponse(constant.ErrCodeFieldRequired, constant.ErrMsgMandatoryField, ErrTypeAPI, "routingCode", "Make sure routingCode value is fulfilled when countryCode is provided")

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "country code is required when routing code is provided"):
		return createResponse(constant.ErrCodeFieldRequired, constant.ErrMsgMandatoryField, ErrTypeAPI, "countryCode", "Make sure countryCode value is fulfilled when routingCode is provided")

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "invalid routing code"):
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, "routingCode", "Make sure routingCode format is correct")

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && (xbCoreProcessorResp.ErrorDetails)[0].Field != "" && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "invalid"):
		errObj := xbCoreProcessorResp.ErrorDetails[0]
		fieldName := util.SnakeToCamel(errObj.Field)
		msg := fmt.Sprintf(constant.ErrDetailMsgFormatField, fieldName)
		if errObj.Message != "" {
			msg = errObj.Message
		}
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, fieldName, msg)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "remitter not found"):
		msg := fmt.Sprintf(constant.ErrDetailMsgNotExists, "senderId")
		return createResponse(constant.ErrCodeResourceMissing, msg, ErrTypeAPI, "senderId", msg)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "not found") && (xbCoreProcessorResp.ErrorDetails)[0].Field != "":
		fieldName := util.SnakeToCamel((xbCoreProcessorResp.ErrorDetails)[0].Field)
		msg := fmt.Sprintf(constant.ErrDetailMsgNotExists, fieldName)
		return createResponse(constant.ErrCodeResourceMissing, msg, ErrTypeAPI, fieldName, msg)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "already used") && (xbCoreProcessorResp.ErrorDetails)[0].Field != "":
		fieldName := util.SnakeToCamel((xbCoreProcessorResp.ErrorDetails)[0].Field)
		return createResponse(constant.ErrCodeResourceAlreadyExists, constant.ErrMsgAccountNumberAlreadyExists, ErrTypeAPI, fieldName, (xbCoreProcessorResp.ErrorDetails)[0].Message)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "invalid amount. please refer"):
		return createResponse(constant.ErrCodeInvalidAmount, constant.ErrMsgInvalidAmount, ErrTypeAPI, "destinationAmount", constant.ErrDetailMsgInvalidAmount)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "beneficiary_account_number should match pattern"):
		return createResponse(constant.ErrCodeInvalidAccountNumberFormat, constant.ErrMsgInvalidAccountFormat, ErrTypeAPI, "accountNumber", constant.ErrDetailMsgInvalidAccountFormat)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "ach code should match pattern"):
		return createResponse(constant.ErrCodeInvalidACHCodeFormat, constant.ErrMsgInvalidACHCodeFormat, ErrTypeAPI, "routingValue", constant.ErrDetailMsgInvalidAchCodeFormat)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "swift should match pattern"),
		xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), " routing_code_value (swift) must be 8 or 11 characters in length"):
		return createResponse(constant.ErrCodeInvalidSwiftCodeFormat, constant.ErrMsgInvalidSwiftCodeFormat, ErrTypeAPI, "routingValue", constant.ErrDetailMsgInvalidSwiftCodeFormat)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "remitter_account_type salary payments should be b2p"),
		xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "beneficiary_account_type salary payments should be b2p"):
		return createResponse(constant.ErrCodeUnallowedPurpose, constant.ErrMsgUnallowedPurpose, ErrTypeAPI, "purposeCode", constant.ErrDetailMsgInvalidSalaryPayment)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "statement_narrative"):
		return createResponse(constant.ErrCodeInvalidRemarkPattern, constant.ErrMsgInvalidRemarkPattern, ErrTypeAPI, "remark", constant.ErrDetailMsgInvalidRemarkPattern)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "bank name is invalid"):
		return createResponse(constant.ErrCodeInvalidBankNameFormat, constant.ErrMsgInvalidBankNameFormat, ErrTypeAPI, "bankName", constant.ErrDetailMsgInvalidBankNameFormat)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "swift code not found"):
		return createResponse(constant.ErrCodeSwiftCodeNotFound, constant.ErrMsgSwiftCodeNotFound, ErrTypeAPI, "swiftCode", constant.ErrDetailMsgSwiftCodeNotFound)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "config spread not found"):
		// Use business-friendly message instead of exposing internal config details
		errorDetail := xbCoreProcessorResp.ErrorDetails[0]
		userFriendlyMessage := "Currency pair not supported for this transaction"
		return createResponse(constant.ErrCodeCurrencyNotEnabled, userFriendlyMessage, ErrTypeAPI, util.SnakeToCamel(errorDetail.Field), userFriendlyMessage)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "address is invalid"):
		return createResponse(constant.ErrCodeInvalidAddressFormat, constant.ErrMsgInvalidAddressFormat, ErrTypeAPI, "address", constant.ErrDetailMsgInvalidAddressFormat)
	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "destination_amount should be"):
		return createResponse(constant.ErrCodeInvalidAmount, constant.ErrMsgInvalidAmount, ErrTypeAPI, "destinationAmount", constant.ErrDetailMsgInvalidAmount)
	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "remit_purpose_code is invalid / not supported"):
		return createResponse(constant.ErrCodeUnallowedPurpose, constant.ErrMsgUnallowedPurpose, ErrTypeAPI, "purposeCode", constant.ErrDetailMsgUnallowedPurposeCode)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "file size exceeds partner maximum limit"):
		// Forward the original error message from XB Core Processor which contains the dynamic limit
		originalMessage := (xbCoreProcessorResp.ErrorDetails)[0].Message
		return createResponse(constant.ErrCodeFileSizeExceedsLimit, constant.ErrMsgFileSizeExceedsLimit, ErrTypeAPI, "document", originalMessage)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0 && strings.Contains(strings.ToLower((xbCoreProcessorResp.ErrorDetails)[0].Message), "unsupported file type"):
		// Handle unsupported file type errors from XB Core Processor
		originalMessage := (xbCoreProcessorResp.ErrorDetails)[0].Message
		field := util.SnakeToCamel((xbCoreProcessorResp.ErrorDetails)[0].Field)
		return createResponse(constant.ErrCodeFieldValueInvalid, "Document format is not supported by the payment processor", ErrTypeAPI, field, originalMessage)

	case xbCoreProcessorResp != nil && len(xbCoreProcessorResp.ErrorDetails) > 0:
		// General error response for XB
		errMessage := ""
		field := ""
		detailErr := ""
		errObj := xbCoreProcessorResp.ErrorDetails[0]
		if xbCoreProcessorResp.Message != "" {
			errMessage = xbCoreProcessorResp.Message
		}
		if errObj.Field != "" {
			field = util.SnakeToCamel(errObj.Field)
		}
		if errObj.Message != "" {
			detailErr = errObj.Message
		}
		return createResponse(constant.ErrCodeFieldValueInvalid, errMessage, ErrTypeAPI, field, detailErr)
	case strings.Contains(strings.ToLower(message), "payout is expired"):
		return createResponse(constant.ErrCodePayoutAlreadyExpired, constant.ErrMsgPayoutAlreadyExpired, ErrTypeAPI, "expiredAt", constant.ErrDetailMsgPayoutAlreadyExpired)

	case strings.Contains(strings.ToLower(message), "merchant not found"):
		return createResponse(constant.ErrCodeMerchantNotFound, constant.ErrMsgMerchantNotFound, ErrTypeAPI, "", constant.ErrDetailMsgMerchantNotFound)

	case strings.Contains(strings.ToLower(message), constant.ErrMalformedRequestBodyPayload.Error()):
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, "", constant.ErrDetailMsgFieldFormatInvalid)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "not rfi"):
		return createResponse(constant.ErrCodeResourceMissing, constant.ErrMsgRfiIdNotExist, ErrTypeAPI, "", constant.ErrDetailMsgRfiIdNotExist)

	case strings.EqualFold(code, HttpStatusErrorUnauthorized) && strings.EqualFold(message, "invalid token"):
		return createResponse(constant.ErrCodeInvalidCredential, constant.ErrMsgAccessTokenInvalid, ErrTypeAPI, "", constant.ErrDetailMsgAccessToken)

	case strings.EqualFold(code, HttpStatusErrorUnauthorized) && strings.EqualFold(message, constant.ErrMsgUnauthorized):
		return createResponse(constant.ErrCodeInvalidCredential, constant.ErrMsgUnauthorized, ErrTypeAPI, "", constant.ErrMsgUnauthorized)

	case strings.EqualFold(code, HttpStatusErrorUnauthorized) && strings.Contains(message, "Merchant status is"):
		return createResponse(constant.ErrCodeInvalidMerchantStatus, message, ErrTypeAPI, "", constant.ErrMsgInvalidMerchantStatus)

	case strings.EqualFold(code, HttpStatusErrorRequest) &&
		(slices.Contains([]string{constant.ErrInvalidSortOrder.Error(), constant.ErrFilterDateInput.Error(), constant.ErrDateRangeExceedLimit.Error()}, message) || strings.HasSuffix(message, " date format")):
		return createResponse(constant.ErrCodeFieldValueInvalid, message, ErrTypeAPI, "", message)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "cannot unmarshal"):
		fieldName := extractInvalidField(message)
		msg := fmt.Sprintf(constant.ErrDetailMsgFormatField, fieldName)
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, fieldName, msg)

	case (strings.EqualFold(code, HttpStatusErrorRequest) || strings.EqualFold(code, HttpStatusErrorInternal) || strings.EqualFold(code, HttpStatusErrorUnprocessableContent)) && strings.Contains(strings.ToLower(message), "channel code not found"):
		return createResponse(constant.ErrCodeUnsupportedChannelCode, constant.ErrMsgChannelCodeNotSupported, ErrTypeAPI, "channelCode", constant.ErrDetailMsgChannelCode)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "unsupported currency"):
		return createResponse(constant.ErrCodeFieldValueInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, "currency", message)

	case message == constant.ErrBeneficiaryNameLengthExceeded.Error():
		return createResponse(constant.ErrCodeFieldFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPIValidation, "accountName", fmt.Sprintf("Account name length must not exceed %d characters", constant.DisbursementMaxLengthBeneficiaryName))

	case strings.Contains(strings.ToLower(message), "bulk disbursement not found") || strings.Contains(strings.ToLower(message), constant.ErrPayoutIsNotFound.Error()):
		return createResponse(constant.ErrCodeResourceMissing, constant.ErrMsgPayoutIdNotExist, ErrTypeAPI, "", constant.ErrDetailMsgPayoutIdNotExist)

	case strings.Contains(strings.ToLower(message), "disbursement by merchant and reference not found") || strings.Contains(strings.ToLower(message), constant.ErrDisbursementNotFound.Error()):
		return createResponse(constant.ErrCodeResourceMissing, constant.ErrMsgReferenceIdNotExist, ErrTypeAPI, "", constant.ErrDetailMsgReferenceIdNotExist)

	case strings.EqualFold(code, HttpStatusErrorUnprocessableContent) && strings.Contains(strings.ToLower(message), "inquiry not found"):
		return createResponse(constant.ErrCodeResourceMissing, constant.ErrMsgInquiryIdNotExist, ErrTypeAPI, "inquiryId", constant.ErrDetailMsgInquiryIdNotExist)

	case strings.EqualFold(code, HttpStatusErrorUnprocessableContent) && strings.Contains(strings.ToLower(message), "client reference id already exist"):
		return createResponse(constant.ErrCodeResourceAlreadyExists, constant.ErrMsgClientReferenceIdAlreadyExist, ErrTypeAPI, "", "")

	case strings.EqualFold(code, HttpStatusErrorUnprocessableContent) && strings.Contains(strings.ToLower(message), "expiry is not allowed to be less than current time"):
		return createResponse(constant.ErrCodeFormatInvalid, constant.ErrMsgExpiryLessThanCurrentTime, ErrTypeAPI, "", "")

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "max bulk disbursement request is 1000"):
		return createResponse(constant.ErrCodeFormatInvalid, constant.ErrMsgPayoutFormatInvalid, ErrTypeAPI, "", constant.ErrDetailMsgPayoutFormatInvalid)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "required"):
		fieldName := extractMandatoryField(message)
		msg := fmt.Sprintf(constant.ErrDetailMsgMandatoryField, fieldName)
		return createResponse(constant.ErrCodeFieldRequired, constant.ErrMsgMandatoryField, ErrTypeAPI, fieldName, msg)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "'numeric' tag"):
		fieldName := extractMandatoryField(message)
		return createResponse(constant.ErrCodeFormatInvalid, constant.ErrMsgFieldFormatInvalid, ErrTypeAPI, fieldName, fmt.Sprintf(constant.ErrDetailMsgFormatField, fieldName))

	case strings.EqualFold(code, HttpStatusErrorValidation) && exposeUnmappingRequestError && strings.Contains(strings.ToLower(message), "payout"):
		fieldName := extractMandatoryField(message)
		return createResponse(
			constant.ErrCodeFormatInvalid,
			constant.ErrMsgPayoutFormatInvalid,
			ErrTypeAPI,
			"",
			fmt.Sprintf(constant.ErrDetailMsgRequestFormatField, cases.Title(language.English).String(fieldName)),
		)
	case strings.EqualFold(code, HttpStatusErrorValidation) && strings.Contains(strings.ToLower(message), "invalid mid or mid tag"):
		return createResponse(constant.ErrCodeV2APIValidationError, constant.ErrProcessingConfigInvalidMID.Error(), ErrTypeAPIValidation, "processingConfig", constant.ErrProcessingConfigInvalidMID.Error())

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "'min' tag"):
		fieldName := extractMandatoryField(message)
		msg := fmt.Sprintf(constant.ErrDetailMsgAmount, fieldName)
		return createResponse(constant.ErrCodeFieldRequired, constant.ErrMsgAmountBelowMinimum, ErrTypeAPI, fieldName, msg)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "'max' tag"):
		fieldName := extractMandatoryField(message)
		msg := fmt.Sprintf(constant.ErrDetailAmountAboveMaxFmt, fieldName)
		return createResponse(constant.ErrCodeFieldRequired, constant.ErrMsgAmountAboveMaximum, ErrTypeAPI, fieldName, msg)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "min amount"):
		return createResponse(constant.ErrCodeAmountBelowLimit, constant.ErrMsgAmountBelowMinimum, ErrTypeAPI, "value", fmt.Sprintf(constant.ErrDetailMsgAmount, "value"))

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "max amount"):
		return createResponse(constant.ErrCodeAmountAboveLimit, constant.ErrMsgAmountAboveMaximum, ErrTypeAPI, "value", fmt.Sprintf(constant.ErrDetailAmountAboveMaxFmt, "value"))

	case (strings.EqualFold(code, HttpStatusErrorRequest) || strings.EqualFold(code, HttpStatusErrorDuplicatedCheck)) && (strings.Contains(strings.ToLower(message), "duplicate reference") || strings.Contains(strings.ToLower(message), "reference id already exist")):
		msg := fmt.Sprintf(constant.ErrDetailMsgId, "referenceId")
		return createResponse(constant.ErrCodeResourceAlreadyExists, constant.ErrMsgIdAlreadyExists, ErrTypeAPI, "referenceId", msg)

	case (strings.EqualFold(code, HttpStatusErrorRequest) || strings.EqualFold(code, HttpStatusErrorDuplicatedCheck)) && (strings.Contains(strings.ToLower(message), "payouts are in process")):
		return createResponse(constant.ErrCodePayoutInProcess, constant.ErrMsgPayoutAreBeingInProcess, ErrTypeAPI, "", constant.ErrDetailMsgPayoutInProcess)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(message, "X-Request-Id"),
		strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(message, "Idempotency-Key already exist"):
		return createResponse(constant.ErrCodeResourceAlreadyExists, constant.ErrMsgXRequestAlreadyExists, ErrTypeAPI, "X-Request-Id", constant.ErrDetailMsgXRequestId)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(message, "Idempotency-Key"):
		message = strings.Replace(message, "Idempotency-Key", "X-Request-Id", 1)
		return createResponse(constant.ErrCodeFieldFormatInvalid, message, ErrTypeAPI, "X-Request-Id", constant.ErrDetailMsgXRequestIdFormat)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(message, constant.ErrDetailMsgPayoutDstNotEligible):
		return createResponse(constant.ErrCodePayoutNotEligible, constant.ErrDetailMsgPayoutDstNotEligible, ErrTypeAPI, "destination account", message)

	case strings.EqualFold(code, HttpStatusErrorrRequestTimeout):
		return createResponse(constant.ErrCodeTimeout, constant.ErrMsgTimeout, ErrTypeGateway, "", constant.ErrDetailMsgServiceUnavailable)

	case strings.EqualFold(code, HttpStatusErrorForbidden) && strings.Contains(strings.ToLower(message), "insufficient balance"):
		return createResponse(constant.ErrCodeBalanceInsufficient, constant.ErrMsgBalanceInsufficient, ErrTypeAPI, "", constant.ErrDetailMsgBalanceInsufficient)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), "invalid status inquiry"):
		return createResponse(constant.ErrCodeInvalidStatusInquiry, constant.ErrMsgInvalidStatusInquiry, ErrTypeAPI, "inquiryId", constant.ErrDetailMsgInvalidStatusInquiry)

	case strings.EqualFold(code, HttpStatusErrorRequest) && strings.Contains(strings.ToLower(message), constant.ErrDisbursementStatusAlreadyApproved.Error()):
		return createResponse(constant.ErrCodePayoutSessionAlreadyConfirmed, constant.ErrMsgPayoutSessionAlreadyConfirmed, ErrTypeAPI, "", constant.ErrDetailMsgPayoutSessionAlreadyConfirmed)
	case strings.EqualFold(message, constant.ErrInvalidDisbursementAmount.Error()):
		return createResponse(constant.ErrCodeFieldValueInvalid, constant.ErrCodeFieldValueInvalid, ErrTypeAPI, "value", constant.ErrInvalidDisbursementAmount.Error())
	case message == constant.ErrMissingSubMerchantId.Error():
		return createResponse(constant.ErrCodeFieldRequired, constant.ErrMissingSubMerchantId.Error(), ErrTypeAPI, "X-SubMerchant-Id", constant.ErrMissingSubMerchantId.Error())

	case message == constant.ErrUnifiedPaymentMetadataSizeLimitExceeded.Error():
		return createResponse(constant.ErrCodeUnprocessableEntity, constant.ErrMessageV2APIValidationError, ErrTypeAPI, "metadata", fmt.Sprintf("metadata exceeds maximum allowed length of %d characters", constant.UnifiedPaymentMaxMetadataLength))

	case code == HttpStatusErrorUnprocessableContent || strings.Contains(message, "please wait for a moment"):
		return createResponse(constant.ErrCodeUnprocessableEntity, message, ErrTypeAPI, "", message)

	case code == HttpStatusErrorDailyLimitReached:
		return createResponse(constant.ErrCodeDailyPayoutLimitReached, constant.ErrMsgPayoutDailyLimitExceeded, ErrTypeAPI, "", message)

	case code == HttpStatusErrorDuplicatedCheck:
		return createResponse(constant.ErrCodeDuplicateError, constant.ErrMessageV2DuplicateError, ErrTypeGateway, "", message)

	case code == HttpStatusErrorForbidden:
		return createResponse(constant.ErrCodeForbiddenAccess, constant.ErrMessageV2RequestForbidden, ErrTypeAPI, "", message)

	case code == HttpStatusErrorNotFound:
		return createResponse(constant.ErrCodeDataNotFound, constant.ErrMessageV2NotFound, ErrTypeGateway, "", message)

	case code == HttpStatusErrorBadGateway:
		return createResponse(constant.ErrCodeBadGateway, constant.ErrMessageV2BadGateway, ErrTypeGateway, "", message)

	case code == HttpStatusServiceUnavailable:
		return createResponse(constant.ErrCodeServiceUnavailable, constant.ErrMessageV2ServiceUnavailable, ErrTypeGateway, "", message)

	case code == HttpStatusErrorRequestLimitExceeded:
		return createResponse(constant.ErrCodeFrequencyAboveLimit, constant.ErrCodeV2FrequencyAboveLimit, ErrTypeAPI, "", message)

	default:
		return createResponse(constant.ErrCodeGeneral, constant.ErrMsgGeneral, ErrTypeAPI, "", constant.ErrDetailMsgGeneralError)
	}
}

func createResponse(code, message, errorType, field, detailMsg string) *OpenApiErrorNonSnap {
	r := &OpenApiErrorNonSnap{
		Code:    code,
		Message: message,
		Error: &ApiError{
			Type:    errorType,
			Details: []ApiErrorDetail{{Field: field, Message: detailMsg}},
		},
	}

	return r
}

func extractInvalidField(errorMessage string) string {
	pattern := `(?i)field\s+([^ ]+)\.([^ ]+)`
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(errorMessage)
	if len(match) > 2 {
		fieldPath := match[2]
		parts := strings.Split(fieldPath, ".")
		fieldName := parts[len(parts)-1]
		return fieldName
	}
	return ""
}

func extractMandatoryField(errorMessage string) string {
	patterns := []string{
		`[Ff]ield validation for '(\w+)'`,
		`(?i)([a-zA-Z0-9_]+)\s+is\s+required`,
		`(?i)([a-zA-Z0-9_]+)\s+is\s+invalid`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		match := re.FindStringSubmatch(errorMessage)
		if len(match) > 1 {
			fieldName := match[1]
			camelCaseField := strcase.ToLowerCamel(fieldName)
			return camelCaseField
		}
	}
	return ""
}

func extractXbCoreProcessorResponse(response string) *xbCoreProcessorModel.CommonResponse {
	var resp xbCoreProcessorModel.CommonResponse
	err := json.Unmarshal([]byte(response), &resp)
	if err != nil {
		return nil
	}
	return &resp
}
