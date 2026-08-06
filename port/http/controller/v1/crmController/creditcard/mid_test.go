package crmCreditcardController

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestGetMIDList(t *testing.T) {
	svc := serviceMocks.NewICreditCardService(t)

	tests := []struct {
		name           string
		queryParams    string
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Invalid page parameter",
			queryParams: "page=invalid",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid page number"}`,
		},
		{
			name:        "ERROR: Invalid perPage parameter",
			queryParams: "perPage=invalid",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid per page number"}`,
		},
		{
			name:        "ERROR: Service error",
			queryParams: "page=1&perPage=10&mid=TEST_MID&acquirer=TEST_ACQUIRER&name=TEST_NAME&type=TEST_TYPE&transactionType=TEST_TRANSACTION_TYPE&installmentType=TEST_INSTALLMENT_TYPE&isActive=false&isDefault=false",
			modifierMock: func() {
				svc.On("GetMIDList", constant.ValueCtxMockType(), mock.Anything).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "errors":"some error"}`,
		},
		{
			name:        "SUCCESS with complete params",
			queryParams: "page=1&perPage=10&mid=TEST_MID&acquirer=TEST_ACQUIRER&name=TEST_NAME&type=TEST_TYPE&transactionType=TEST_TRANSACTION_TYPE&installmentType=TEST_INSTALLMENT_TYPE&isActive=true&isDefault=true",
			modifierMock: func() {
				mockResponse := &commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{
						Page:    1,
						PerPage: 10,
					},
				}
				svc.On("GetMIDList", constant.ValueCtxMockType(), mock.Anything).
					Return(mockResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":[], "message": "OK", "pagination":{"page":1, "perPage":10, "totalItems":0, "totalPages":0}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/crm/v1/creditcard/mid/list?%s", test.queryParams), nil)

			router := chi.NewRouter()
			router.Get("/crm/v1/creditcard/mid/list", New(&config.Config{}, &config.Secret{}, svc).GetMIDList)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestGetMIDMapList(t *testing.T) {
	svc := serviceMocks.NewICreditCardService(t)

	tests := []struct {
		name           string
		queryParams    string
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Invalid page parameter",
			queryParams: "page=invalid",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid page number"}`,
		},
		{
			name:        "ERROR: Invalid perPage parameter",
			queryParams: "perPage=invalid",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid per page number"}`,
		},
		{
			name:        "ERROR: Service error",
			queryParams: "page=1&perPage=10",
			modifierMock: func() {
				svc.On("GetMIDMapList", constant.ValueCtxMockType(), 10, 1, "").
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "errors":"some error"}`,
		},
		{
			name:        "SUCCESS",
			queryParams: "page=1&perPage=10",
			modifierMock: func() {
				mockResponse := &commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{
						Page:    1,
						PerPage: 10,
					},
				}
				svc.On("GetMIDMapList", constant.ValueCtxMockType(), 10, 1, "").
					Return(mockResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":[], "message": "OK", "pagination":{"page":1, "perPage":10, "totalItems":0, "totalPages":0}}`,
		},
		{
			name:        "SUCCESS with merchantId filter",
			queryParams: "page=1&perPage=10&merchantId=merchant-123",
			modifierMock: func() {
				mockResponse := &commonModel.PaginationResponse{
					Data: []interface{}{},
					Meta: commonModel.Meta{
						Page:    1,
						PerPage: 10,
					},
				}
				svc.On("GetMIDMapList", constant.ValueCtxMockType(), 10, 1, "merchant-123").
					Return(mockResponse, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":[], "message": "OK", "pagination":{"page":1, "perPage":10, "totalItems":0, "totalPages":0}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/crm/v1/creditcard/mid/map/list?%s", test.queryParams), nil)

			router := chi.NewRouter()
			router.Get("/crm/v1/creditcard/mid/map/list", New(&config.Config{}, &config.Secret{}, svc).GetMIDMapList)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestCreateMID(t *testing.T) {
	svc := serviceMocks.NewICreditCardService(t)

	tests := []struct {
		name           string
		requestBody    []byte
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			requestBody: []byte("{invalid JSON"),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name:        "ERROR: Bad Request - Failed Validation",
			requestBody: []byte(`{"mid": ""}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"42", "errors":{"BaseURL":"Key: 'CreateMIDRequest.BaseURL' Error:Field validation for 'BaseURL' failed on the 'required' tag", "Mid":"Key: 'CreateMIDRequest.Mid' Error:Field validation for 'Mid' failed on the 'required' tag", "Name":"Key: 'CreateMIDRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag", "Password":"Key: 'CreateMIDRequest.Password' Error:Field validation for 'Password' failed on the 'required' tag", "PrincipalAvailable":"Key: 'CreateMIDRequest.PrincipalAvailable' Error:Field validation for 'PrincipalAvailable' failed on the 'required' tag", "Processor":"Key: 'CreateMIDRequest.Processor' Error:Field validation for 'Processor' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR: Service error",
			requestBody: []byte(`{"mid": "TEST_MID", "name": "Test MID", "processor": "TEST_PROCESSOR", "principalAvailable": ["VISA"], "baseUrl": "https://test.com", "password": "test123"}`),
			modifierMock: func() {
				svc.On("CreateMID", constant.ValueCtxMockType(), mock.Anything).
					Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "errors":"some error"}`,
		},
		{
			name:        "SUCCESS",
			requestBody: []byte(`{"mid": "TEST_MID", "name": "Test MID", "processor": "TEST_PROCESSOR", "principalAvailable": ["VISA"], "baseUrl": "https://test.com", "password": "test123"}`),
			modifierMock: func() {
				svc.On("CreateMID", constant.ValueCtxMockType(), mock.Anything).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"created": true}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/crm/v1/creditcard/mid", bytes.NewBuffer(test.requestBody))

			router := chi.NewRouter()
			router.Post("/crm/v1/creditcard/mid", New(&config.Config{}, &config.Secret{}, svc).CreateMID)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}

func TestUpdateMID(t *testing.T) {
	svc := serviceMocks.NewICreditCardService(t)

	validID := uuid.NewString()

	tests := []struct {
		name           string
		id             string
		requestBody    []byte
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR: Invalid ID format",
			id:          "invalid",
			requestBody: []byte(`{}`),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"id is required"}`,
		},
		{
			name:        "ERROR: Bad Request - Invalid JSON",
			id:          validID,
			requestBody: []byte("{invalid JSON"),
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40", "errors":"invalid character 'i' looking for beginning of object key string"}`,
		},
		{
			name:        "ERROR: Service error",
			id:          validID,
			requestBody: []byte(`{"name": "Updated MID"}`),
			modifierMock: func() {
				svc.On("UpdateMID", constant.ValueCtxMockType(), mock.Anything).
					Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99", "errors":"some error"}`,
		},
		{
			name:        "SUCCESS",
			id:          validID,
			requestBody: []byte(`{"name": "Updated MID"}`),
			modifierMock: func() {
				svc.On("UpdateMID", constant.ValueCtxMockType(), mock.Anything).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"updated": true}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/crm/v1/creditcard/mid/%s", test.id), bytes.NewBuffer(test.requestBody))

			router := chi.NewRouter()
			router.Put("/crm/v1/creditcard/mid/{id}", New(&config.Config{}, &config.Secret{}, svc).UpdateMID)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
