package platform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSubMerchantUserList(t *testing.T) {
	userClaims := &userModel.UserTokenClaims{
		Name:       "user",
		MerchantId: uuid.Nil.String(),
	}
	testCases := []struct {
		name         string
		setup        func(platformSvc *mocks.IPlatformService)
		setupClaims  func(r *http.Request) *http.Request
		setupParams  func() string
		expectedCode int
		expectedBody string
	}{
		{
			name: "SUCCESS: Get Merchant User List",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("keyword", "")
				params.Add("roleId", "")
				params.Add("page", "")
				params.Add("perPage", "")
				params.Add("sortBy", "name")
				params.Add("sortOrder", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
				platformSvc.On(
					"GetSubMerchantUserList",
					mock.Anything,
					mock.Anything,
				).Return(
					&commonModel.PaginationResponse{
						Meta: commonModel.Meta{
							Page:       1,
							PerPage:    10,
							TotalItems: 100,
							TotalPages: 10,
						},
						Data: []platform.SubMerchantUserResponse{
							{
								UUID:        uuid.Nil.String(),
								Name:        "name",
								Email:       "email",
								Role:        constant.RoleAdmin,
								Status:      constant.UserStatusActive,
								LastLoginAt: time.Time{},
							},
						},
					},
					nil,
				)
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"code":"00","message":"OK","data":{"data":[{"uuid":"00000000-0000-0000-0000-000000000000","name":"name","email":"email","role":"ADMIN","asMerchantPIC":false,"status":"ACTIVE","lastLoginAt":"0001-01-01T00:00:00Z"}],"meta":{"page":1,"perPage":10,"totalItems":100,"totalPages":10}}}`,
		},
		{
			name: "ERROR: Missing claims",
			setupClaims: func(r *http.Request) *http.Request {
				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("keyword", "")
				params.Add("roleId", "")
				params.Add("page", "")
				params.Add("perPage", "")
				params.Add("sortBy", "name")
				params.Add("sortOrder", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"code":"41","message":"invalid access","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid Page",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("keyword", "")
				params.Add("roleId", "")
				params.Add("page", "x")
				params.Add("perPage", "")
				params.Add("sortBy", "name")
				params.Add("sortOrder", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid page value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Invalid Page Size",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("keyword", "")
				params.Add("roleId", "")
				params.Add("page", "")
				params.Add("perPage", "x")
				params.Add("sortBy", "name")
				params.Add("sortOrder", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":"40","message":"invalid perPage value","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name: "ERROR: Error Get Merchant User List",
			setupClaims: func(r *http.Request) *http.Request {
				ctx := r.Context()
				ctx = context.WithValue(ctx, constant.CtxUserInfoKey, userClaims)
				r = r.WithContext(ctx)

				return r
			},
			setupParams: func() string {
				params := url.Values{}
				params.Add("merchantId", uuid.NewString())
				params.Add("keyword", "")
				params.Add("roleId", "")
				params.Add("page", "")
				params.Add("perPage", "")
				params.Add("sortBy", "name")
				params.Add("sortOrder", "")

				return params.Encode()
			},
			setup: func(platformSvc *mocks.IPlatformService) {
				platformSvc.On(
					"GetSubMerchantUserList",
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					errors.New("errors"),
				)
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"code":"99","message":"errors","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := logger.NewZapLogger(logger.Config{})
			svc := mocks.NewIPlatformService(t)
			tc.setup(svc)

			ctrl := New(logger, nil, svc)

			urlParams := tc.setupParams()
			req := httptest.NewRequest(http.MethodGet, "/users?"+urlParams, nil)
			req = tc.setupClaims(req)

			rr := httptest.NewRecorder()
			// Create the handler and serve the request
			handler := http.HandlerFunc(ctrl.GetSubMerchantUserList)
			handler.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tc.expectedCode, rr.Code)
			assert.JSONEqf(t, tc.expectedBody, rr.Body.String(), "Expected body\n%s\nbut got:\n%s\n", tc.expectedBody, rr.Body.String())

		})
	}
}
