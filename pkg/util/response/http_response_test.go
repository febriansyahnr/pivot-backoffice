package response_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/xbCoreProcessor"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestSendGeneralResponseOK(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantCode int
	}{
		{
			name:     "should return 200 with valid data",
			data:     map[string]string{"key": "value"},
			wantCode: http.StatusOK,
		},
		{
			name:     "should return 200 with nil data",
			data:     nil,
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			response.SendGeneralResponseOK(w, tt.data)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var resp response.Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, response.HttpStatusOK, resp.Code)
		})
	}
}

func TestSendOpenApiResponseError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  int
		wantError string
	}{
		{
			name:      "should return 500 for internal error",
			err:       errors.New("internal server error"),
			wantCode:  http.StatusInternalServerError,
			wantError: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			response.SendOpenApiResponseError(w, tt.err)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var resp response.OpenApiResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantError, resp.Message)
		})
	}
}

func TestHandleDetailedError(t *testing.T) {
	tests := []struct {
		name                        string
		code                        string
		message                     string
		errorType                   string
		wantCode                    string
		exposeUnmappingRequestError bool
	}{
		{
			name:      "should handle required field error",
			code:      response.HttpStatusErrorRequest,
			message:   "Field validation for 'Name' failed on the 'required' tag",
			errorType: response.ErrTypeAPI,
			wantCode:  constant.ErrCodeFieldRequired,
		},
		{
			name:      "should handle numeric field error",
			code:      response.HttpStatusErrorRequest,
			message:   "Field validation for 'Amount' failed on the 'numeric' tag",
			errorType: response.ErrTypeAPI,
			wantCode:  constant.ErrCodeFormatInvalid,
		},
		{
			name:      "should handle duplicate reference error",
			code:      response.HttpStatusErrorDuplicatedCheck,
			message:   "duplicate reference id",
			errorType: response.ErrTypeAPI,
			wantCode:  constant.ErrCodeResourceAlreadyExists,
		},
		{
			name:                        "should handle invalid card number error with expose flag",
			code:                        response.HttpStatusErrorRequest,
			message:                     constant.ErrInvalidCardNumber.Error(),
			errorType:                   response.ErrTypeAPI,
			wantCode:                    constant.ErrCodeV2InvalidCardNumber,
			exposeUnmappingRequestError: true,
		},
		{
			name:                        "should handle card decryption error with expose flag",
			code:                        response.HttpStatusErrorRequest,
			message:                     constant.ErrFailedToDecryptCardPayment.Error(),
			errorType:                   response.ErrTypeAPI,
			wantCode:                    constant.ErrCodeV2CardDecryption,
			exposeUnmappingRequestError: true,
		},
		{
			name:                        "should handle invalid card information error with expose flag",
			code:                        response.HttpStatusErrorRequest,
			message:                     constant.ErrInvalidCardInformation.Error(),
			errorType:                   response.ErrTypeAPI,
			wantCode:                    constant.ErrCodeV2InvalidCardInfo,
			exposeUnmappingRequestError: true,
		},
		{
			name:                        "should handle missing foreign card billing info",
			code:                        response.HttpStatusErrorRequest,
			message:                     constant.ErrForeignCardBillingInformationMissing.Error(),
			errorType:                   response.ErrTypeAPI,
			wantCode:                    constant.ErrCodeV2ForeignCardBillingInfoMissing,
			exposeUnmappingRequestError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.exposeUnmappingRequestError {
				ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)
			}
			result := response.HandleDetailedError(ctx, tt.code, tt.message, tt.errorType)
			assert.Equal(t, tt.wantCode, result.Code)
		})
	}
}

func generateXBError(errorType string) error {
	if errorType == "" {
		errorType = "general"
	}
	xbResponse := &xbCoreProcessorModel.CommonResponse{}
	xbResponse.Message = "test error message"
	switch errorType {
	case "invalid":
		xbResponse.ErrorDetails = []*xbCoreProcessorModel.ApiErrorDetail{
			{
				Field:   "test",
				Message: "invalid field error",
			},
		}
		xbResponse.Message = "invalid field error"
	case "general":
		xbResponse.ErrorDetails = []*xbCoreProcessorModel.ApiErrorDetail{
			{
				Field:   "test",
				Message: "test",
			},
		}
	}
	jsonXBError, _ := json.Marshal(xbResponse)
	return pkgErrs.New(response.HttpErrRequest, errors.New(string(jsonXBError)))
}

func TestSendOpenApiNonSnapResponseError(t *testing.T) {
	ctxValue := context.WithValue(context.Background(), constant.CtxTraceIdKey, "trace-123")
	ctxV2 := context.WithValue(context.Background(), constant.CtxUseV2ErrorCode, true)

	tests := []struct {
		name         string
		ctx          context.Context
		err          error
		wantCode     int
		wantSource   string
		wantCode_    string
		wantMessage  string
		wantField    string
		wantFieldMsg string
	}{
		{
			name:     "should handle basic error",
			ctx:      context.Background(),
			err:      errors.New("test error"),
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "should handle error with trace ID",
			ctx:      ctxValue,
			err:      errors.New("test error"),
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "invalid merchant status",
			ctx:      ctxValue,
			err:      pkgErrs.New(response.HttpErrUnauthorized, errors.New("Merchant status is closed. Reason: test")),
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "invalid merchant status",
			ctx:      ctxValue,
			err:      pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidSortOrder),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "cannot unmarshal request body",
			ctx:      ctxValue,
			err:      pkgErrs.New(response.HttpErrRequest, errors.New("cannot unmarshal request body")),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "required field",
			ctx:      ctxValue,
			err:      constant.NewErrRequiredField("test"),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid field format",
			ctx:      ctxValue,
			err:      constant.NewErrInvalidFieldFmt("test"),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "resource not found",
			ctx:      ctxValue,
			err:      constant.NewErrResourceNotFound("testing", "test"),
			wantCode: http.StatusNotFound,
		},
		{
			name:     "invalid payload",
			ctx:      ctxValue,
			err:      constant.NewErrInvalidPayload(errors.New("unmarshal failed")),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "error field validation",
			ctx:      ctxValue,
			err:      constant.NewErrFieldValidation(errors.New("validation error")),
			wantCode: http.StatusBadRequest,
		},
		{
			name: "error field validation-validator results",
			ctx:  ctxValue,
			err: func() error {
				err := validator.New().Struct(struct {
					Name          string `validate:"required"`
					Min           int    `validate:"min=1"`
					Max           int    `validate:"max=2"`
					Email         string `validate:"email"`
					StringNumeric string `validate:"numeric"`
				}{Max: 3, StringNumeric: "ABC"})
				return constant.NewErrFieldValidation(err)
			}(),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "invalid rules",
			ctx:      ctxValue,
			err:      constant.NewErrInvalidRules("abc", errors.New("testing")),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "xb error general",
			ctx:      ctxValue,
			err:      generateXBError("general"),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "xb error invalid field",
			ctx:      ctxValue,
			err:      generateXBError("invalid"),
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "payout amount invalid",
			ctx:      ctxValue,
			err:      pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidDisbursementAmount),
			wantCode: http.StatusBadRequest,
		},
		{
			name:       "v2 upstream source",
			ctx:        ctxV2,
			err:        pkgErrs.New(response.HttpErrRequest, errors.New("invalid request")),
			wantCode:   http.StatusBadRequest,
			wantSource: response.ErrorSourceUpstream,
		},
		{
			name:       "v2 downstream source",
			ctx:        ctxV2,
			err:        pkgErrs.New(response.HttpErrRequestTimeout, errors.New("partner timeout")),
			wantCode:   http.StatusGatewayTimeout,
			wantSource: response.ErrorSourceDownstream,
		},
		{
			name:       "v2 system source",
			ctx:        ctxV2,
			err:        pkgErrs.New(response.HttpErrInternal, errors.New("internal error")),
			wantCode:   http.StatusInternalServerError,
			wantSource: response.ErrorSourceSystem,
		},
		{
			name: "field transformation for snake_case country_code",
			ctx:  context.WithValue(context.Background(), constant.CtxTraceIdKey, "test-trace-openapi"),
			err: func() error {
				xbResponse := &xbCoreProcessorModel.CommonResponse{
					Code: "40", Message: "Invalid request. Please check your input and try again.", ErrorType: "API_ERROR",
					ErrorDetails: []*xbCoreProcessorModel.ApiErrorDetail{{Field: "country_code", Message: "Country code must be 2-letter ISO code"}},
				}
				jsonXBError, _ := json.Marshal(xbResponse)
				return pkgErrs.New(response.HttpErrRequest, errors.New(string(jsonXBError)))
			}(),
			wantCode:     http.StatusBadRequest,
			wantField:    "countryCode",
			wantFieldMsg: "Country code must be 2-letter ISO code",
		},
		{
			name: "unsupported file type response mapping",
			ctx:  context.WithValue(context.Background(), constant.CtxTraceIdKey, "test-trace-123"),
			err: pkgErrs.New(response.HttpErrRequest, errors.New(`{
				"code": "40",
				"error": "API_VALIDATION_ERROR",
				"error_type": "API_ERROR",
				"error_details": [{"field": "document", "message": "validation error on field 'document': unsupported file type '.zip'. Only PDF files are supported"}],
				"message": "Invalid request. Please check your input and try again."
			}`)),
			wantCode:     http.StatusBadRequest,
			wantCode_:    "field_value_invalid",
			wantMessage:  "Document format is not supported by the payment processor",
			wantField:    "document",
			wantFieldMsg: "validation error on field 'document': unsupported file type '.zip'. Only PDF files are supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			response.SendOpenApiNonSnapResponseError(tt.ctx, w, tt.err)

			assert.Equal(t, tt.wantCode, w.Code)
			assert.Equal(t, constant.ContentTypeApplicationJson, w.Header().Get("Content-Type"))

			var resp response.OpenApiErrorNonSnap
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.NotNil(t, resp.Error)

			if tt.wantSource != "" {
				assert.Equal(t, tt.wantSource, resp.Error.Source)
			}
			if tt.wantCode_ != "" {
				assert.Equal(t, tt.wantCode_, resp.Code)
			}
			if tt.wantMessage != "" {
				assert.Equal(t, tt.wantMessage, resp.Message)
			}
			if tt.wantField != "" {
				assert.Len(t, resp.Error.Details, 1)
				assert.Equal(t, tt.wantField, resp.Error.Details[0].Field)
				assert.Equal(t, tt.wantFieldMsg, resp.Error.Details[0].Message)
			}
		})
	}
}

func TestFieldNameTransformationFromSnakeCaseToCamelCase(t *testing.T) {
	ctx := context.WithValue(context.Background(), constant.CtxTraceIdKey, "test-trace-123")

	tests := []struct {
		name            string
		snakeCaseFields []string
		expectedFields  []string
		messages        []string
	}{
		{
			name:            "country_code should transform to countryCode",
			snakeCaseFields: []string{"country_code"},
			expectedFields:  []string{"countryCode"},
			messages:        []string{"Country code must be 2-letter ISO code"},
		},
		{
			name:            "account_number should transform to accountNumber",
			snakeCaseFields: []string{"account_number"},
			expectedFields:  []string{"accountNumber"},
			messages:        []string{"Account number is invalid"},
		},
		{
			name:            "bank_code should transform to bankCode",
			snakeCaseFields: []string{"bank_code"},
			expectedFields:  []string{"bankCode"},
			messages:        []string{"Bank code is required"},
		},
		{
			name:            "beneficiary_name should transform to beneficiaryName",
			snakeCaseFields: []string{"beneficiary_name"},
			expectedFields:  []string{"beneficiaryName"},
			messages:        []string{"Beneficiary name is required"},
		},
		{
			name:            "contact_country_code should transform to contactCountryCode",
			snakeCaseFields: []string{"contact_country_code"},
			expectedFields:  []string{"contactCountryCode"},
			messages:        []string{"Contact country code is invalid"},
		},
		{
			name:            "multiple fields should all transform",
			snakeCaseFields: []string{"country_code", "account_number", "bank_code"},
			expectedFields:  []string{"countryCode", "accountNumber", "bankCode"},
			messages:        []string{"Country code must be 2-letter ISO code", "Account number is required", "Bank code is invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorDetails := make([]*xbCoreProcessorModel.ApiErrorDetail, len(tt.snakeCaseFields))
			for i, field := range tt.snakeCaseFields {
				errorDetails[i] = &xbCoreProcessorModel.ApiErrorDetail{Field: field, Message: tt.messages[i]}
			}

			xbResponse := &xbCoreProcessorModel.CommonResponse{
				Code: "40", Message: "Invalid request. Please check your input and try again.", ErrorType: "API_ERROR",
				ErrorDetails: errorDetails,
			}
			jsonXBError, _ := json.Marshal(xbResponse)
			err := pkgErrs.New(response.HttpErrRequest, errors.New(string(jsonXBError)))

			w := httptest.NewRecorder()
			response.SendApiResponseError(ctx, w, err)

			var apiResp response.ApiResponse
			unmarshalErr := json.Unmarshal(w.Body.Bytes(), &apiResp)
			assert.NoError(t, unmarshalErr)
			assert.NotNil(t, apiResp.Error)
			assert.Len(t, apiResp.Error.Details, len(tt.expectedFields))

			for i, detail := range apiResp.Error.Details {
				assert.Equal(t, tt.expectedFields[i], detail.Field)
				assert.Equal(t, tt.messages[i], detail.Message)
			}

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, constant.ContentTypeApplicationJson, w.Header().Get("Content-Type"))
			assert.Equal(t, "API_ERROR", apiResp.Error.Type)
		})
	}
}
