package snapCoreRepository

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/routingProcessor/accountInquiry"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankAccount"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateVirtualAccount(t *testing.T) {
	standardArgs := []interface{}{
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("snapCoreModel.InquiryAccountRequest"),
		mock.AnythingOfType(constant.MockTypeMapStringStringReference),
	}

	tests := []struct {
		name      string
		request   snapCoreModel.InquiryAccountRequest
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
		result    *snapCoreModel.InquiryAccountResponseData
		err       error
		errType   string
	}{
		{
			name: "ERROR: Failed to prepare HTTP request",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return(nil, 0, errors.New("failed to prepare http request"))
			},
			err: errors.New("failed to prepare http request"),
		},
		{
			name: "ERROR: Unmarshal process failed",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte("{[}"), 0, nil)
			},
			err: errors.New(`invalid character '[' looking for beginning of object key string`),
		},
		{
			name: "ERROR: Unprocessable entity",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte(`{"error": "Unprocessable Entity"}`), http.StatusUnprocessableEntity, nil)
			},
			errType: httpResponse.HttpErrRequest,
			err:     errors.New(`ERROR_REQUEST | "Unprocessable Entity"`),
			result: &snapCoreModel.InquiryAccountResponseData{
				ResponseCode: "", ResponseMessage: "", PartnerReferenceNo: "", BeneficiaryAccountName: "", BeneficiaryAccountNo: "", BeneficiaryBankCode: "", BeneficiaryBankName: "",
			},
		},
		{
			name: "ERROR: Internal server error",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte(`{"error": "Internal Server Error"}`), http.StatusInternalServerError, nil)
			},
			errType: httpResponse.HttpErrThirdParty,
			err:     errors.New(`ERROR_THIRD_PARTY | "Internal Server Error"`),
			result: &snapCoreModel.InquiryAccountResponseData{
				ResponseCode: "", ResponseMessage: "", PartnerReferenceNo: "", BeneficiaryAccountName: "", BeneficiaryAccountNo: "", BeneficiaryBankCode: "", BeneficiaryBankName: "",
			},
		},
		{
			name: "ERROR: Internal server error with error field containing JSON message",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte(`{"code":"500","error":{"message":"Database connection failed"}}`), http.StatusInternalServerError, nil)
			},
			errType: httpResponse.HttpErrThirdParty,
			err:     errors.New(`ERROR_THIRD_PARTY | Database connection failed`),
			result: &snapCoreModel.InquiryAccountResponseData{
				ResponseCode: "", ResponseMessage: "", PartnerReferenceNo: "", BeneficiaryAccountName: "", BeneficiaryAccountNo: "", BeneficiaryBankCode: "", BeneficiaryBankName: "",
			},
		},
		{
			name: "ERROR: 408 request timeout",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte(`{"code":"408","message":"Request timeout"}`), http.StatusRequestTimeout, nil)
			},
			errType: httpResponse.HttpErrRequestTimeout,
			err:     errors.New("ERROR_REQUEST_TIMEOUT | Request timeout"),
			result:  &snapCoreModel.InquiryAccountResponseData{},
		},
		{
			name: "ERROR: 429 too many requests",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte(`{"code":"429","message":"Too many requests"}`), http.StatusTooManyRequests, nil)
			},
			errType: httpResponse.HttpErrRequestLimitExceeded,
			err:     errors.New("REQUEST_LIMIT_EXCEEDED | Too many requests"),
			result:  &snapCoreModel.InquiryAccountResponseData{},
		},
		{
			name: "ERROR: 502 bad gateway",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte(`{"code":"502","message":"Bad gateway"}`), http.StatusBadGateway, nil)
			},
			errType: httpResponse.HttpErrBadGateway,
			err:     errors.New("ERROR_BAD_GATEWAY | Bad gateway"),
			result:  &snapCoreModel.InquiryAccountResponseData{},
		},
		{
			name: "ERROR: 503 service unavailable",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte(`{"code":"503","message":"Service unavailable"}`), http.StatusServiceUnavailable, nil)
			},
			errType: httpResponse.HttpErrServiceUnavailable,
			err:     errors.New("ERROR_SERVICE_UNAVAILABLE | Service unavailable"),
			result:  &snapCoreModel.InquiryAccountResponseData{},
		},
		{
			name: "ERROR: 504 gateway timeout",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte(`{"code":"504","message":"Gateway timeout"}`), http.StatusGatewayTimeout, nil)
			},
			errType: httpResponse.HttpErrRequestTimeout,
			err:     errors.New("ERROR_REQUEST_TIMEOUT | Gateway timeout"),
			result:  &snapCoreModel.InquiryAccountResponseData{},
		},
		{
			name: "ERROR: 502 with body stripped by CDN (non-JSON response)",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return([]byte(`error code: 502`), http.StatusBadGateway, nil)
			},
			errType: httpResponse.HttpErrBadGateway,
			err:     errors.New("ERROR_BAD_GATEWAY | partner returned HTTP 502"),
			result:  nil,
		},
		{
			name: "SUCCESS: Data found",
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).
					Return(
						[]byte(`{"data": {"responseCode": "200999200", "responseMessage": "success", "partnerReferenceNo": "ABC", "isVirtualAccount": true}}`), http.StatusOK, nil,
					)
			},
			result: &snapCoreModel.InquiryAccountResponseData{
				ResponseCode: "200999200", ResponseMessage: "success", PartnerReferenceNo: "ABC", IsVirtualAccount: true,
			},
		},
	}

	conf := &config.Config{}
	secret := &config.Secret{SnapCoreSecret: struct {
		InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
	}{InternalServiceKey: ""}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			mockHttp := httpMocks.NewIHTTPRequest(t)
			test.setupMock(mockHttp)
			result, err := New(conf, secret, mockLogger, mockHttp).GetBankAccountInquiry(context.Background(), test.request)
			if test.err == nil {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.err.Error())
				if test.errType != "" {
					extractedErrType, _ := pkgErrors.ExtractError(err)
					assert.Equal(t, test.errType, extractedErrType)
				}
			}
			assert.Equal(t, test.result, result)
		})
	}
}

func TestBankAccountInquiry(t *testing.T) {
	standardArgs := []interface{}{
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("snapCoreModel.InquiryAccountRequest"),
		mock.AnythingOfType(constant.MockTypeMapStringStringReference),
	}

	tests := []struct {
		name      string
		request   *routingProcessorModel.InquiryAccountRequest
		setupMock func(mockHttp *httpMocks.IHTTPRequest)
		result    *routingProcessorModel.InquiryAccountResponseData
		err       error
	}{
		{
			name: "SUCCESS: Bank account inquiry successful",
			request: &routingProcessorModel.InquiryAccountRequest{
				BeneficiaryBankCode:    "008",
				BeneficiaryAccountNo:   "1234567890",
				BeneficiaryAccountName: "John Doe",
				PartnerReferenceNo:     "REF123",
			},
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).
					Return(
						[]byte(`{"data": {"responseCode": "200999200", "responseMessage": "success", "partnerReferenceNo": "REF123", "beneficiaryAccountName": "John Doe", "beneficiaryAccountNo": "1234567890", "beneficiaryBankCode": "008", "beneficiaryBankName": "Bank Mandiri", "isVirtualAccount": true}}`), http.StatusOK, nil,
					)
			},
			result: &routingProcessorModel.InquiryAccountResponseData{
				ResponseCode:           "200999200",
				ResponseMessage:        "success",
				PartnerReferenceNo:     "REF123",
				BeneficiaryAccountName: "John Doe",
				BeneficiaryAccountNo:   "1234567890",
				BeneficiaryBankCode:    "008",
				BeneficiaryBankName:    "Bank Mandiri",
				IsVirtualAccount:       true,
			},
			err: nil,
		},
		{
			name: "ERROR: Response is nil",
			request: &routingProcessorModel.InquiryAccountRequest{
				BeneficiaryBankCode:    "008",
				BeneficiaryAccountNo:   "1234567890",
				BeneficiaryAccountName: "John Doe",
				PartnerReferenceNo:     "REF123",
			},
			setupMock: func(mockHttp *httpMocks.IHTTPRequest) {
				mockHttp.On("POST", standardArgs...).Return(nil, 0, errors.New("network error"))
			},
			result: nil,
			err:    errors.New("network error"),
		},
	}

	conf := &config.Config{}
	secret := &config.Secret{SnapCoreSecret: struct {
		InternalServiceKey string "mapstructure:\"INTERNAL_SERVICE_KEY\""
	}{InternalServiceKey: ""}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			mockHttp := httpMocks.NewIHTTPRequest(t)
			test.setupMock(mockHttp)
			result, err := New(conf, secret, mockLogger, mockHttp).BankAccountInquiry(context.Background(), test.request)
			if test.err == nil {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.err.Error())
			}
			assert.Equal(t, test.result, result)
		})
	}
}
