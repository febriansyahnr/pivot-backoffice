package ipWhitelistController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	ipAddress = "237.84.2.178"
	subnet    = "24"
)

func TestCreate(t *testing.T) {
	payload := &ipwhitelistModel.CreateIPWhitelistConfiguration{
		MerchantID: uuid.NewString(),
		IP:         ipAddress,
		Subnet:     subnet,
		Priority:   1,
		Action:     "ALLOW",
	}

	ipConfig := &ipwhitelistModel.IPWhitelistConfiguration{
		ID:          uuid.Max.String(),
		MerchantID:  uuid.Max.String(),
		IP:          ipAddress,
		Subnet:      subnet,
		Priority:    1,
		Action:      "ALLOW",
		Status:      "ACTIVE",
		Description: "IP Whitelist",
	}

	testCases := []struct {
		name         string
		expectedCode int
		expectedBody string
		requestBody  func() []byte
		setup        func(svc *mockSvc.IIPWhitelistService)
		setupClaims  bool
	}{
		{
			name: "SUCCESS: Create IP Configuration",
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal(payload)
				return payloadRequestByte
			},
			setup: func(svc *mockSvc.IIPWhitelistService) {
				svc.On(
					"Create",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*ipwhitelistModel.CreateIPWhitelistConfiguration"),
				).Return(ipConfig, nil)
			},
			setupClaims:  true,
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"OK","data":{"id":"ffffffff-ffff-ffff-ffff-ffffffffffff","merchantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","ip":"237.84.2.178","subnet":"24","priority":1,"action":"ALLOW","status":"ACTIVE","description":"IP Whitelist","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}}`,
		},
		{
			name: "ERROR: Unable get claims",
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal(payload)
				return payloadRequestByte
			},
			setup: func(svc *mockSvc.IIPWhitelistService) {

			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			setupClaims:  false,
		},
		{
			name: "ERROR: Failed decode request",
			requestBody: func() []byte {
				return []byte("invalid")
			},
			setup: func(svc *mockSvc.IIPWhitelistService) {

			},
			setupClaims:  true,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","errors":"invalid character 'i' looking for beginning of value"}`,
		},
		{
			name: "ERROR: Payload failed validation",
			requestBody: func() []byte {
				modifiedRequest := *payload
				modifiedRequest.Action = ""
				payloadRequestByte, _ := json.Marshal(modifiedRequest)
				return payloadRequestByte
			},
			setup: func(svc *mockSvc.IIPWhitelistService) {
			},
			setupClaims:  true,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","errors":{"Action":"Key: 'CreateIPWhitelistConfiguration.Action' Error:Field validation for 'Action' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: Create IP Configuration",
			requestBody: func() []byte {
				payloadRequestByte, _ := json.Marshal(payload)
				return payloadRequestByte
			},
			setup: func(svc *mockSvc.IIPWhitelistService) {
				svc.On(
					"Create",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*ipwhitelistModel.CreateIPWhitelistConfiguration"),
				).Return(nil, errors.New("errors"))
			},
			setupClaims:  true,
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"99","message":"errors","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			svc := mockSvc.NewIIPWhitelistService(t)
			mockValidator := validator.New()
			tc.setup(svc)

			ctrl := New(svc, mockValidator)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBuffer(tc.requestBody()))

			if tc.setupClaims {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: uuid.NewString(),
				})
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(ctrl.Create)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())

		})
	}
}
