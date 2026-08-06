package ipWhitelistController

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestList(t *testing.T) {

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
		setup        func(svc *mockSvc.IIPWhitelistService)
		setupRequest func(req *http.Request) *http.Request
		setupClaims  bool
	}{
		{
			name: "SUCCESS: Get IP Configuration List",
			setup: func(svc *mockSvc.IIPWhitelistService) {
				svc.On(
					"List",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(&commonModel.PaginationResponse{
					Data: []*ipwhitelistModel.IPWhitelistConfiguration{
						ipConfig,
					},
					Meta: commonModel.Meta{
						Page:       1,
						PerPage:    10,
						TotalItems: 1,
						TotalPages: 1,
					}}, nil)
			},
			setupRequest: func(req *http.Request) *http.Request {
				req.URL.RawQuery = "?ip=10.0.0.0&status=ACTIVE"
				return req
			},
			setupClaims:  true,
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"OK","data":{"data":[{"id":"ffffffff-ffff-ffff-ffff-ffffffffffff","merchantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","ip":"237.84.2.178","subnet":"24","priority":1,"action":"ALLOW","status":"ACTIVE","description":"IP Whitelist","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}],"meta":{"page":1,"perPage":10,"totalItems":1,"totalPages":1}}}`,
		},
		{
			name: "ERROR: Unable get claims",
			setup: func(svc *mockSvc.IIPWhitelistService) {

			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			setupClaims:  false,
		},
		{
			name: "ERROR: Invalid param values",
			setup: func(svc *mockSvc.IIPWhitelistService) {
			},
			setupRequest: func(req *http.Request) *http.Request {
				req.URL.RawQuery = "?ip=10.0.0.0&status=ACTIVES"
				return req
			},
			setupClaims:  true,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","errors":{"Status":"Key: 'GetIPWhitelistConfiguration.Status' Error:Field validation for 'Status' failed on the 'oneof' tag"}}`,
		},
		{
			name: "ERROR: Invalid page param value",
			setup: func(svc *mockSvc.IIPWhitelistService) {
			},
			setupRequest: func(req *http.Request) *http.Request {
				req.URL.RawQuery = "?ip=10.0.0.0&status=ACTIVE&page=abc"
				return req
			},
			setupClaims:  true,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid page number","error":{"type":"API_ERROR","message":"invalid page number","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid pageSize param value",
			setup: func(svc *mockSvc.IIPWhitelistService) {
			},
			setupRequest: func(req *http.Request) *http.Request {
				req.URL.RawQuery = "?ip=10.0.0.0&status=ACTIVE&page=1&perPage=abc"
				return req
			},
			setupClaims:  true,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid per page number","error":{"type":"API_ERROR","message":"invalid per page number","recommendation":""},"data":null}`,
		},
		{
			name: "ERROR: Get IP Configuration",
			setup: func(svc *mockSvc.IIPWhitelistService) {
				svc.On(
					"List",
					constant.ValueCtxMockType(),
					mock.Anything,
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
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			if tc.setupClaims {
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: uuid.NewString(),
				})
				req = req.WithContext(ctx)
			}

			if tc.setupRequest != nil {
				req = tc.setupRequest(req)
			}

			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(ctrl.GetList)
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
