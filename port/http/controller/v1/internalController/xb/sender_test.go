package internalXbController_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/xb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateSender(t *testing.T) {
	validPayload := xbModel.CreateSenderRequest{
		Name:                 "John Doe",
		AccountType:          "Individual",
		Address:              "America St.",
		CountryCode:          "US",
		State:                "New York",
		City:                 "New York",
		Postcode:             "54321",
		IdentificationType:   "Passport",
		IdentificationNumber: "1234567890",
	}
	validUUID := uuid.MustParse("3f7be294-d5cf-44ef-8707-1615c1ca7aef")
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	tests := []struct {
		name             string
		mockSetup        func(svc *serviceMock.IXbPayoutService)
		setupBody        func(*testing.T) []byte
		setupHeader      func(*http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: unauthorized no provided merchant info",
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			setupHeader: func(req *http.Request) {
				// empty modifier
			},
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found", "error":{"details":[{"field":"", "message":"Invalid Merchant request"}], "traceId":"", "type":"API_ERROR"}, "message":"Merchant not found"}`,
		},
		{
			name: "ERROR: Invalid request body",
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			setupHeader: validRequestID,
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: Missing required request",
			setupBody: func(t *testing.T) []byte {
				payload := xbModel.CreateSenderRequest{
					AccountType: "Individual",
					CountryCode: "US",
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupHeader: validRequestID,
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"field_required","error":{"details":[{"field":"name","message":"Make sure name value is fulfilled"}],"traceId":"","type":"API_ERROR"},"message":"Mandatory field is missing"}`,
		},
		{
			name: "ERROR: CreateSender service error",
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupHeader: validRequestID,
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.CreateSenderRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "SUCCESS",
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupHeader: validRequestID,
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("CreateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.CreateSenderRequest"),
				).Return(&xbModel.CreateSenderResponse{
					UUID: validUUID,
				}, nil)
			},
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"accountType":"", "address":"", "bankAccountNumber":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "createdAt":"0001-01-01T00:00:00Z", "dob":"", "identificationNumber":"", "identificationType":"", "name":"", "postcode":"", "sourceOfIncome":"", "state":"", "uuid":"3f7be294-d5cf-44ef-8707-1615c1ca7aef"}, "message":"Success"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
			tt.mockSetup(xbPayoutSvc)

			ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

			req := httptest.NewRequest(http.MethodPost, "/open-api/v1/xb/sender", bytes.NewBuffer(tt.setupBody(t)))
			rec := httptest.NewRecorder()

			if tt.setupHeader != nil {
				tt.setupHeader(req)
			}

			router := chi.NewRouter()
			router.Post("/open-api/v1/xb/sender", ctrl.CreateSender)

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tt.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetListSender(t *testing.T) {
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}
	testCases := []struct {
		name             string
		mockSetup        func(svc *serviceMock.IXbPayoutService)
		queryParams      string
		setupHeader      func(*http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: unauthorized no provided merchant info",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupHeader: func(req *http.Request) {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found", "error":{"details":[{"field":"", "message":"Invalid Merchant request"}], "traceId":"", "type":"API_ERROR"}, "message":"Merchant not found"}`,
		},
		{
			name:        "ERROR: invalid page",
			queryParams: "?page=invalid",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:        "ERROR: invalid perPage",
			queryParams: "?perPage=invalid",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:        "ERROR: invalid fetchAll",
			queryParams: "?fetchAll=invalid",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:        "ERROR: invalid showDeactivated",
			queryParams: "?showDeactivated=invalid",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:        "ERROR: error get list sender",
			queryParams: "",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("GetListSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetListSenderRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:        "SUCCESS",
			queryParams: "",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("GetListSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetListSenderRequest"),
				).Once().Return(&xbModel.PaginationResponse{}, nil)
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":null, "message":"Success", "pagination":{"fetchAll":false, "page":0, "perPage":0, "totalItems":0, "totalPages":0}}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
			tt.mockSetup(xbPayoutSvc)

			ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

			req := httptest.NewRequest(http.MethodGet, "/open-api/v1/xb/sender"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			if tt.setupHeader != nil {
				tt.setupHeader(req)
			}

			router := chi.NewRouter()
			router.Get("/open-api/v1/xb/sender", ctrl.GetListSender)

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tt.expectedRespBody, rec.Body.String())
		})
	}
}

func TestGetSenderById(t *testing.T) {
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}
	testCases := []struct {
		name             string
		mockSetup        func(svc *serviceMock.IXbPayoutService)
		id               string
		setupHeader      func(*http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: unauthorized no provided merchant info",
			id:   "99999999-9999-9999-9999-999999999999",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupHeader: func(req *http.Request) {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found", "error":{"details":[{"field":"", "message":"Invalid Merchant request"}], "traceId":"", "type":"API_ERROR"}, "message":"Merchant not found"}`,
		},
		{
			name: "ERROR: invalid id",
			id:   "invalid",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"field_required", "error":{"details":[{"field":"id", "message":"Make sure id value is fulfilled"}], "traceId":"", "type":"API_ERROR"}, "message":"Mandatory field is missing"}`,
		},
		{
			name: "ERROR: error get sender by id",
			id:   "99999999-9999-9999-9999-999999999999",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("GetSenderById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetSenderByIdRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			id:   "99999999-9999-9999-9999-999999999999",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("GetSenderById",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetSenderByIdRequest"),
				).Once().Return(&xbModel.CreateSenderResponse{}, nil)
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"accountType":"", "address":"", "bankAccountNumber":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "createdAt":"0001-01-01T00:00:00Z", "dob":"", "identificationNumber":"", "identificationType":"", "name":"", "postcode":"", "sourceOfIncome":"", "state":"", "uuid":"00000000-0000-0000-0000-000000000000"}, "message":"Success"}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
			tt.mockSetup(xbPayoutSvc)

			ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

			req := httptest.NewRequest(http.MethodGet, "/open-api/v1/xb/sender/"+tt.id, nil)
			rec := httptest.NewRecorder()

			if tt.setupHeader != nil {
				tt.setupHeader(req)
			}

			router := chi.NewRouter()
			router.Get("/open-api/v1/xb/sender/{id}", ctrl.GetSenderById)

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tt.expectedRespBody, rec.Body.String())
		})
	}
}

func TestUpdateSender(t *testing.T) {
	validPayload := xbModel.UpdateSenderRequest{
		Name:        "John Doe",
		AccountType: "Individual",
		Address:     "America St.",
		CountryCode: "US",
		State:       "New York",
		City:        "New York",
		Postcode:    "54321",
	}
	validId := uuid.MustParse("3f7be294-d5cf-44ef-8707-1615c1ca7aef")
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}
	testCases := []struct {
		name             string
		mockSetup        func(svc *serviceMock.IXbPayoutService)
		id               string
		setupBody        func(*testing.T) []byte
		setupHeader      func(*http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: unauthorized no provided merchant info",
			id:   "99999999-9999-9999-9999-999999999999",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			setupHeader: func(req *http.Request) {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found", "error":{"details":[{"field":"", "message":"Invalid Merchant request"}], "traceId":"", "type":"API_ERROR"}, "message":"Merchant not found"}`,
		},
		{
			name: "ERROR: invalid id",
			id:   "invalid",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"field_required", "error":{"details":[{"field":"id", "message":"Make sure id value is fulfilled"}], "traceId":"", "type":"API_ERROR"}, "message":"Mandatory field is missing"}`,
		},
		{
			name: "ERROR: Invalid request body",
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			id:          "99999999-9999-9999-9999-999999999999",
			setupHeader: validRequestID,
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"general_error","error":{"details":[{"field":"","message":"Please contact our representative team"}],"traceId":"","type":"API_ERROR"},"message":"General error"}`,
		},
		{
			name: "ERROR: error update sender",
			id:   validId.String(),
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("UpdateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.UpdateSenderRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			id:   validId.String(),
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("UpdateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.UpdateSenderRequest"),
				).Once().Return(&xbModel.CreateSenderResponse{}, nil)
			},
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"accountType":"", "address":"", "bankAccountNumber":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "createdAt":"0001-01-01T00:00:00Z", "dob":"", "identificationNumber":"", "identificationType":"", "name":"", "postcode":"", "sourceOfIncome":"", "state":"", "uuid":"00000000-0000-0000-0000-000000000000"}, "message":"Success"}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
			tt.mockSetup(xbPayoutSvc)

			ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

			req := httptest.NewRequest(http.MethodPut, "/open-api/v1/xb/sender/"+tt.id, bytes.NewBuffer(tt.setupBody(t)))
			rec := httptest.NewRecorder()

			if tt.setupHeader != nil {
				tt.setupHeader(req)
			}

			router := chi.NewRouter()
			router.Put("/open-api/v1/xb/sender/{id}", ctrl.UpdateSender)

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tt.expectedRespBody, rec.Body.String())
		})
	}
}

func TestDeactivateSender(t *testing.T) {
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}
	testCases := []struct {
		name             string
		mockSetup        func(svc *serviceMock.IXbPayoutService)
		id               string
		setupHeader      func(*http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: unauthorized no provided merchant info",
			id:   "99999999-9999-9999-9999-999999999999",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupHeader: func(req *http.Request) {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"merchant_not_found", "error":{"details":[{"field":"", "message":"Invalid Merchant request"}], "traceId":"", "type":"API_ERROR"}, "message":"Merchant not found"}`,
		},
		{
			name: "ERROR: invalid id",
			id:   "invalid",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				// empty modifier
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusBadRequest,
			expectedRespBody: `{"code":"field_required", "error":{"details":[{"field":"id", "message":"Make sure id value is fulfilled"}], "traceId":"", "type":"API_ERROR"}, "message":"Mandatory field is missing"}`,
		},
		{
			name: "ERROR: error get sender by id",
			id:   "99999999-9999-9999-9999-999999999999",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("DeactivateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetSenderByIdRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name: "SUCCESS",
			id:   "99999999-9999-9999-9999-999999999999",
			mockSetup: func(svc *serviceMock.IXbPayoutService) {
				svc.On("DeactivateSender",
					c.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetSenderByIdRequest"),
				).Once().Return(&xbModel.CreateSenderResponse{}, nil)
			},
			setupHeader:      validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":{"accountType":"", "address":"", "bankAccountNumber":"", "city":"", "contactCountryCode":"", "contactNumber":"", "countryCode":"", "createdAt":"0001-01-01T00:00:00Z", "dob":"", "identificationNumber":"", "identificationType":"", "name":"", "postcode":"", "sourceOfIncome":"", "state":"", "uuid":"00000000-0000-0000-0000-000000000000"}, "message":"Success"}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			xbPayoutSvc := serviceMock.NewIXbPayoutService(t)
			tt.mockSetup(xbPayoutSvc)

			ctrl := New(cfg, WithXbPayoutService(xbPayoutSvc))

			req := httptest.NewRequest(http.MethodPatch, "/open-api/v1/xb/sender/"+tt.id+"/deactivate", nil)
			rec := httptest.NewRecorder()

			if tt.setupHeader != nil {
				tt.setupHeader(req)
			}

			router := chi.NewRouter()
			router.Patch("/open-api/v1/xb/sender/{id}/deactivate", ctrl.DeactivateSender)

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tt.expectedRespBody, rec.Body.String())
		})
	}
}
