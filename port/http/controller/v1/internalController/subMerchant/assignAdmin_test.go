package submerchant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockMerchantSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssignAdmin(t *testing.T) {
	testCases := []struct {
		name             string
		requestBody      func() []byte
		setupBody        func(req *http.Request)
		mockSetup        func(merchantSvcMocks *mockMerchantSvc.IMerchantService)
		expectedStatus   int
		expectedResponse string
	}{
		{
			name: "SUCCESS",
			setupBody: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			requestBody: func() []byte {
				payloadRequest := &merchantModel.SubMerchantAdminRequest{
					Email: "test@test.com",
					Name:  "test",
				}
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService) {
				merchantSvcMocks.On(
					"AssignSubMerchantAdmin",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*merchant.SubMerchantAdminRequest"),
				).Return(
					nil,
				)
			},
			expectedStatus:   200,
			expectedResponse: `{"code":"00"}`,
		},
		{
			name: "ERROR: Email already exists",
			setupBody: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			requestBody: func() []byte {
				payloadRequest := &merchantModel.SubMerchantAdminRequest{
					Email: "test@test.com", Name: "test", // NOSONAR
				}
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService) {
				merchantSvcMocks.On(
					"AssignSubMerchantAdmin", mock.Anything, mock.Anything,
				).Return(pkgErrs.New(response.HttpErrRequest, constant.ErrUserAlreadyExists))
			},
			expectedStatus:   400,
			expectedResponse: `{"code":"40","errors":"user already exists"}`,
		},
		{
			name: "ERROR: Empty SubmerchantId",
			setupBody: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, "")
			},
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal("[]]")
				return payloadRequestByte
			},
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService) {
			},
			expectedStatus:   400,
			expectedResponse: `{"code":"40","errors":"missing submerchant id"}`,
		},
		{
			name: "ERROR: Failed decode request",
			setupBody: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal("[]]")
				return payloadRequestByte
			},
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService) {
			},
			expectedStatus:   400,
			expectedResponse: `{"code":"40","errors":"json: cannot unmarshal string into Go value of type merchant.SubMerchantAdminRequest"}`,
		},
		{
			name: "ERROR: Failed validate request",
			setupBody: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			requestBody: func() []byte {
				payloadRequest := &merchantModel.SubMerchantAdminRequest{
					Email: "testtest.com",
					Name:  "test",
				}
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService) {
			},
			expectedStatus:   400,
			expectedResponse: `{"code":"40","errors":{"Email":"Key: 'SubMerchantAdminRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"}}`,
		},
		{
			name: "ERROR: Assign submerchant admin",
			setupBody: func(req *http.Request) {
				req.Header.Set(constant.HeaderXSubMerchantID, uuid.New().String())
			},
			requestBody: func() []byte {
				payloadRequest := &merchantModel.SubMerchantAdminRequest{
					Email: "test@test.com",
					Name:  "test",
				}
				payloadRequestByte, _ := json.Marshal(payloadRequest)
				return payloadRequestByte
			},
			mockSetup: func(merchantSvcMocks *mockMerchantSvc.IMerchantService) {
				merchantSvcMocks.On(
					"AssignSubMerchantAdmin",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*merchant.SubMerchantAdminRequest"),
				).Return(
					errors.New("errors"),
				)
			},
			expectedStatus:   500,
			expectedResponse: `{"code":"99","errors":"errors"}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			merchantSvc := mockMerchantSvc.NewIMerchantService(t)
			accountSvc := mockMerchantSvc.NewIAccountService(t)
			orchestratorSvc := mockMerchantSvc.NewIOrchestratorService(t)
			mockValidator := validator.New()
			tt.mockSetup(merchantSvc)

			mc := New(merchantSvc, accountSvc, orchestratorSvc, mockValidator)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/sub-merchants/admin", bytes.NewBuffer(tt.requestBody()))
			merchantAuth := &merchantModel.MerchantAuthTokenClaims{
				MerchantId: "aec6636d-7a02-4d93-a4c5-006b9c235068", // NOSONAR
			}
			ctx = context.WithValue(ctx, constant.CtxMerchantInfo, merchantAuth)
			ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, merchantAuth.MerchantId)

			req = req.WithContext(ctx)

			if tt.setupBody != nil {
				tt.setupBody(req)
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.AssignAdmin)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.JSONEq(t, tt.expectedResponse, rr.Body.String())
			merchantSvc.AssertExpectations(t)
		})
	}

}
