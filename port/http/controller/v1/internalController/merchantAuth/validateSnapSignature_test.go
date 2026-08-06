package internalMerchantAuthController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	merchantServiceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidateSNAPSignature(t *testing.T) {

	testCases := []struct {
		desc             string
		requestBody      string
		mockSetup        func(mockService *merchantServiceMocks.IMerchantService)
		expectedStatus   int
		expectedResponse string
	}{
		{
			desc: "SUCCESS",
			requestBody: `{
				"signature":"signature",
				"timestamp":"2025-04-30T11:05:31+07:00",
				"clientId":"clientId",
				"url":"url",
				"body":{},
				"method":"method",
				"accessToken":"accessToken"
			}`,
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On("ValidateSnapRequestSignature", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"Success","data":"ok"}`,
		},
		{
			desc: "ERROR: Invalid payload",
			requestBody: `{
				"timestamp":"timestampt",
				"clientId":"clientId",
				"url":"url",
				"body":"body",
				"method":"method",
				"accessToken":"accessToken"
			}`,
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":"40","message":"Key: 'ValidateSnapSignatureRequest.Signature' Error:Field validation for 'Signature' failed on the 'required' tag","error":{"type":"API_ERROR","message":"Key: 'ValidateSnapSignatureRequest.Signature' Error:Field validation for 'Signature' failed on the 'required' tag","recommendation":""},"data":null}`,
		},
		{
			desc: "ERROR",
			requestBody: `{
				"signature":"signature",
				"timestamp":"2025-04-30T11:05:31+07:00",
				"clientId":"clientId",
				"url":"url",
				"body":{},
				"method":"method",
				"accessToken":"accessToken"
			}`,
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On("ValidateSnapRequestSignature", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","message":"error","recommendation":""},"data":null}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockService := merchantServiceMocks.NewIMerchantService(t)
			ctx := context.Background()

			merchantAuthController := New(validator.New(), mockService)

			tc.mockSetup(mockService)

			baseUrl := "/api/internal/v1/validate"
			req := httptest.NewRequest(http.MethodPost, baseUrl, strings.NewReader(tc.requestBody))
			chiRouterCtx := chi.NewRouteContext()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(ctx)

			req.Header.Add("X-CLIENT-KEY", "CLIENT-KEY")

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(merchantAuthController.ValidateSNAPSignature)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			// assert.Equalf(t, tc.expectedResponse, httpRecorder.Body.String(), "expected %s, got %s", tc.expectedResponse, httpRecorder.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestGenerateSNAPSignature(t *testing.T) {

	testCases := []struct {
		desc             string
		requestBody      string
		mockSetup        func(mockService *merchantServiceMocks.IMerchantService)
		expectedStatus   int
		expectedResponse string
	}{
		{
			desc: "SUCCESS",
			requestBody: `{
				"timestamp":"2025-04-30T11:05:31+07:00",
				"clientId":"clientId",
				"url":"url",
				"body":{
				},
				"method":"method",
				"accessToken":"accessToken"
			}`,
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On("GenerateSnapRequestSignature", mock.Anything, mock.Anything).Return("signature", nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"Success","data":"ok"}`,
		},
		{
			desc: "ERROR: Invalid payload",
			requestBody: `{
				"timestamp":timestampt,
				"clientId":"clientId",
				"url":"url",
				"body":"body",
				"method":"method",
				"accessToken":"accessToken"
			}`,
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{\"code\":\"40\",\"message\":\"Key: 'ValidateSnapSignatureRequest.Signature' Error:Field validation for 'Signature' failed on the 'required' tag\",\"error\":{\"type\":\"API_ERROR\",\"message\":\"Key: 'ValidateSnapSignatureRequest.Signature' Error:Field validation for 'Signature' failed on the 'required' tag\",\"recommendation\":\"\"},\"data\":null}`,
		},
		{
			desc: "ERROR",
			requestBody: `{
				"signature":"signature",
				"timestamp":"2025-04-30T11:05:31+07:00",
				"clientId":"clientId",
				"url":"url",
				"body":{
				},
				"method":"method",
				"accessToken":"accessToken"
			}`,
			mockSetup: func(mockService *merchantServiceMocks.IMerchantService) {
				mockService.On("GenerateSnapRequestSignature", mock.Anything, mock.Anything).Return("", errors.New("error"))
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"error","error":{"type":"UNKNOWN","message":"error","recommendation":""},"data":null}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockService := merchantServiceMocks.NewIMerchantService(t)
			ctx := context.Background()

			merchantAuthController := New(validator.New(), mockService)

			tc.mockSetup(mockService)

			baseUrl := "/api/internal/v1/generate"
			req := httptest.NewRequest(http.MethodPost, baseUrl, strings.NewReader(tc.requestBody))
			chiRouterCtx := chi.NewRouteContext()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiRouterCtx))
			req = req.WithContext(ctx)

			req.Header.Add("X-CLIENT-KEY", "CLIENT-KEY")

			httpRecorder := httptest.NewRecorder()
			handler := http.HandlerFunc(merchantAuthController.GenerateSNAPSignature)
			handler.ServeHTTP(httpRecorder, req)

			if httpRecorder.Body.String() != "" {
				t.Logf("response: %s", httpRecorder.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, httpRecorder.Code)
			// assert.Equalf(t, tc.expectedResponse, httpRecorder.Body.String(), "expected %s, got %s", tc.expectedResponse, httpRecorder.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
