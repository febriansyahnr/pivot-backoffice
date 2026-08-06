package merchant

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service/v1/users"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	validMerchantId := uuid.NewString()
	districtId := uint16(123)

	validPayload := &merchantModel.CRMUpdateMerchantRequest{
		MerchantID:        validMerchantId,
		Name:              "Test Merchant",
		Description:       "Test Description",
		Website:           "https://test-merchant.com",
		Address:           "Test Address",
		DistrictId:        districtId,
		PostCode:          "12345",
		Logo:              "test-logo.png",
		MerchantEmail:     "test@merchant.com",
		MerchantPhone:     "1234567890",
		PICEmail:          "pic@merchant.com",
		PICPhone:          "0987654321",
		BusinessCountry:   "IDN",
		BusinessStructure: "PT",
		BusinessType:      "type1",
		PICName:           "Test PIC",
		PICJobTitle:       "Test Job Title",
		CountryOfEntity:   "IDN",
		DigitalStatus:     "Digital",
	}

	testCases := []struct {
		name           string
		merchantId     string
		setupBody      func(*testing.T) []byte
		modifierMock   func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:       "SUCCESS",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
				svc.On("Update", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				mockRmq.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"uuid":"","externalId":"","name":"","shortName":"","description":"","website":"","address":"","postcode":"","logo":"","mid":"","merchantEmail":"","merchantPhone":"","picEmail":"","picPhone":"","picName":"","picJobTitle":"","businessType":"","businessStructure":"","businessCountry":"","parentIndustry":"","childIndustry":"","mcc":"","countryOfEntity":"","digitalStatus":"","kycStatus":"","status":"","riskLevel":"","reasonStatus":"","parentId":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:       "ERROR invalid merchant ID",
			merchantId: "invalid-uuid",
			setupBody: func(t *testing.T) []byte {
				return []byte{}
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"id is required"}`,
		},
		{
			name:       "ERROR invalid payload",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) []byte {
				return []byte{}
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"EOF"}`,
		},
		{
			name:       "ERROR invalid json",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) []byte {
				return []byte(`{"invalid-json"}`)
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character '}' after object key"}`,
		},
		{
			name:       "ERROR Invalid industry ID",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) []byte {
				adjustPayload := *validPayload
				adjustPayload.IndustryID = "invalid-industry-id"
				payloadRequestByte, err := json.Marshal(adjustPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"IndustryID":"Key: 'CRMUpdateMerchantRequest.IndustryID' Error:Field validation for 'IndustryID' failed on the 'uuid' tag"}}`,
		},
		{
			name:       "ERROR failed to update merchant",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
				svc.On("Update", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything).Return(nil, errors.New("failed to update merchant"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"failed to update merchant"}`,
		},
		{
			name:       "ERROR missing required params",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) []byte {
				return []byte(`{"name":"Test Merchant"}`)
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody: `{
			  "code": "40",
			  "errors": {
				"Address": "Key: 'CRMUpdateMerchantRequest.Address' Error:Field validation for 'Address' failed on the 'required' tag",
				"BusinessCountry": "Key: 'CRMUpdateMerchantRequest.BusinessCountry' Error:Field validation for 'BusinessCountry' failed on the 'required' tag",
				"BusinessStructure": "Key: 'CRMUpdateMerchantRequest.BusinessStructure' Error:Field validation for 'BusinessStructure' failed on the 'required' tag",
				"BusinessType": "Key: 'CRMUpdateMerchantRequest.BusinessType' Error:Field validation for 'BusinessType' failed on the 'required' tag",
				"DistrictId": "Key: 'CRMUpdateMerchantRequest.DistrictId' Error:Field validation for 'DistrictId' failed on the 'required' tag",
				"MerchantEmail": "Key: 'CRMUpdateMerchantRequest.MerchantEmail' Error:Field validation for 'MerchantEmail' failed on the 'required' tag",
				"MerchantPhone": "Key: 'CRMUpdateMerchantRequest.MerchantPhone' Error:Field validation for 'MerchantPhone' failed on the 'required' tag",
				"PICEmail": "Key: 'CRMUpdateMerchantRequest.PICEmail' Error:Field validation for 'PICEmail' failed on the 'required' tag",
				"PICName": "Key: 'CRMUpdateMerchantRequest.PICName' Error:Field validation for 'PICName' failed on the 'required' tag",
				"PICPhone": "Key: 'CRMUpdateMerchantRequest.PICPhone' Error:Field validation for 'PICPhone' failed on the 'required' tag",
				"PostCode": "Key: 'CRMUpdateMerchantRequest.PostCode' Error:Field validation for 'PostCode' failed on the 'required' tag",
				"Website": "Key: 'CRMUpdateMerchantRequest.Website' Error:Field validation for 'Website' failed on the 'required' tag"
			  }
			}`,
		},
		{
			name:       "ERROR missing both logo and logoFile",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) []byte {
				payload := *validPayload
				payload.Logo = ""
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"either 'logo' or 'logoFile' must be provided"}`,
		},
		{
			name:       "SUCCESS with shortName in JSON",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) []byte {
				payload := *validPayload
				payload.ShortName = "TESTSHORT"
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
				svc.On("Update", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				mockRmq.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"uuid":"","externalId":"","name":"","shortName":"","description":"","website":"","address":"","postcode":"","logo":"","mid":"","merchantEmail":"","merchantPhone":"","picEmail":"","picPhone":"","picName":"","picJobTitle":"","businessType":"","businessStructure":"","businessCountry":"","parentIndustry":"","childIndustry":"","mcc":"","countryOfEntity":"","digitalStatus":"","kycStatus":"","status":"","riskLevel":"","reasonStatus":"","parentId":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:       "SUCCESS - verify update request fields are mapped correctly",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
				svc.On("Update", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything).
					Run(func(args mock.Arguments) {
						req := args.Get(1).(*merchantModel.UpdateMerchantRequest)
						assert.Equal(t, "type1", req.BusinessType)
						assert.Equal(t, "PT", req.BusinessStructure)
						assert.Equal(t, "IDN", req.BusinessCountry)
						assert.Equal(t, "Test PIC", req.PICName)
						assert.Equal(t, "Test Job Title", req.PICJobTitle)
					}).
					Return(&merchantModel.Merchant{}, nil)
				mockRmq.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"uuid":"","externalId":"","name":"","shortName":"","description":"","website":"","address":"","postcode":"","logo":"","mid":"","merchantEmail":"","merchantPhone":"","picEmail":"","picPhone":"","picName":"","picJobTitle":"","businessType":"","businessStructure":"","businessCountry":"","parentIndustry":"","childIndustry":"","mcc":"","countryOfEntity":"","digitalStatus":"","kycStatus":"","status":"","riskLevel":"","reasonStatus":"","parentId":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockUserSvc := mockUser.NewIUserService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)

			tc.modifierMock(mockMerchantSvc, mockRmq)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/crm/merchants/%s", tc.merchantId), bytes.NewBuffer(tc.setupBody(t)))

			router := chi.NewRouter()
			router.Put("/v1/crm/merchants/{id}", New(mockMerchantSvc, mockUserSvc, mockValidator, mockRmq).Update)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, tc.wantRespBody, rec.Body.String(), "expected: %s, got: %s", tc.wantRespBody, rec.Body.String())
		})
	}
}

func TestUpdateMultipartFormData(t *testing.T) {
	validMerchantId := uuid.NewString()

	testCases := []struct {
		name           string
		merchantId     string
		setupBody      func(*testing.T) (*bytes.Buffer, string)
		modifierMock   func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:       "SUCCESS with multipart/form-data",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				_ = writer.WriteField("name", "Test Merchant")
				_ = writer.WriteField("shortName", "TESTSHORT")
				_ = writer.WriteField("description", "Test Description")
				_ = writer.WriteField("address", "Test Address")
				_ = writer.WriteField("districtId", "123")
				_ = writer.WriteField("postcode", "12345")
				_ = writer.WriteField("website", "https://test.com")
				_ = writer.WriteField("logo", "test-logo.png")
				_ = writer.WriteField("merchantEmail", "test@merchant.com")
				_ = writer.WriteField("merchantPhone", "1234567890")
				_ = writer.WriteField("picName", "Test PIC")
				_ = writer.WriteField("picEmail", "pic@merchant.com")
				_ = writer.WriteField("picPhone", "0987654321")
				_ = writer.WriteField("picJobTitle", "Manager")
				_ = writer.WriteField("businessType", "type1")
				_ = writer.WriteField("businessStructure", "PT")
				_ = writer.WriteField("businessCountry", "IDN")
				_ = writer.WriteField("countryOfEntity", "IDN")
				_ = writer.WriteField("digitalStatus", "Digital")

				writer.Close()
				return body, writer.FormDataContentType()
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
				svc.On("Update", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything).Return(&merchantModel.Merchant{}, nil)
				mockRmq.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"uuid":"","externalId":"","name":"","shortName":"","description":"","website":"","address":"","postcode":"","logo":"","mid":"","merchantEmail":"","merchantPhone":"","picEmail":"","picPhone":"","picName":"","picJobTitle":"","businessType":"","businessStructure":"","businessCountry":"","parentIndustry":"","childIndustry":"","mcc":"","countryOfEntity":"","digitalStatus":"","kycStatus":"","status":"","riskLevel":"","reasonStatus":"","parentId":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:       "SUCCESS with multipart/form-data and logo file",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				_ = writer.WriteField("name", "Test Merchant")
				_ = writer.WriteField("description", "Test Description")
				_ = writer.WriteField("address", "Test Address")
				_ = writer.WriteField("districtId", "123")
				_ = writer.WriteField("postcode", "12345")
				_ = writer.WriteField("website", "https://test.com")
				_ = writer.WriteField("logo", "test-logo.png")
				_ = writer.WriteField("merchantEmail", "test@merchant.com")
				_ = writer.WriteField("merchantPhone", "1234567890")
				_ = writer.WriteField("picName", "Test PIC")
				_ = writer.WriteField("picEmail", "pic@merchant.com")
				_ = writer.WriteField("picPhone", "0987654321")
				_ = writer.WriteField("businessType", "type1")
				_ = writer.WriteField("businessStructure", "PT")
				_ = writer.WriteField("businessCountry", "IDN")

				// Add a logo file
				fileWriter, _ := writer.CreateFormFile("logoFile", "test.png")
				fileWriter.Write([]byte("fake image data"))

				writer.Close()
				return body, writer.FormDataContentType()
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
				svc.On("Update", mock.AnythingOfType(constant.MockTypeValueContextReference), mock.Anything).Return(&merchantModel.Merchant{Logo: "https://cdn.example.com/logo.png"}, nil)
				mockRmq.On(
					"PublishActivity",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.Anything,
				).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"uuid":"","externalId":"","name":"","shortName":"","description":"","website":"","address":"","postcode":"","logo":"https://cdn.example.com/logo.png","mid":"","merchantEmail":"","merchantPhone":"","picEmail":"","picPhone":"","picName":"","picJobTitle":"","businessType":"","businessStructure":"","businessCountry":"","parentIndustry":"","childIndustry":"","mcc":"","countryOfEntity":"","digitalStatus":"","kycStatus":"","status":"","riskLevel":"","reasonStatus":"","parentId":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name:       "ERROR invalid districtId in multipart form",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				_ = writer.WriteField("name", "Test Merchant")
				_ = writer.WriteField("districtId", "invalid")

				writer.Close()
				return body, writer.FormDataContentType()
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"strconv.ParseUint: parsing \"invalid\": invalid syntax"}`,
		},
		{
			name:       "ERROR parsing multipart form - invalid content type boundary",
			merchantId: validMerchantId,
			setupBody: func(t *testing.T) (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				body.WriteString("invalid multipart data")
				return body, "multipart/form-data; boundary=----invalid"
			},
			modifierMock: func(svc *mockMerchant.IMerchantService, mockRmq *mockRabbitMq.RabbitMQExt) {
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"multipart: NextPart: EOF"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockUserSvc := mockUser.NewIUserService(t)
			mockValidator := validator.New()
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.modifierMock(mockMerchantSvc, mockRmq)

			rec := httptest.NewRecorder()
			body, contentType := tc.setupBody(t)
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/crm/merchants/%s", tc.merchantId), body)
			req.Header.Set("Content-Type", contentType)

			router := chi.NewRouter()
			router.Put("/v1/crm/merchants/{id}", New(mockMerchantSvc, mockUserSvc, mockValidator, mockRmq, WithLogger(mockLogger)).Update)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEqf(t, tc.wantRespBody, rec.Body.String(), "expected: %s, got: %s", tc.wantRespBody, rec.Body.String())
		})
	}
}
