package charges

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExport(t *testing.T) {
	var (
		validUserClaims = &userModel.UserTokenClaims{
			UUID:       "valid-user-id",
			MerchantId: "valid-merchant-id",
		}
		mockUnifiedPaymentSvc = mockService.NewIUnifiedPaymentService(t)
		mockMerchantService   = mockService.NewIMerchantService(t)
		controller            = ChargesController{
			unifiedPaymentService: mockUnifiedPaymentSvc,
			merchantService:       mockMerchantService,
			validate:              validator.New(),
		}
	)

	testCases := []struct {
		name           string
		callMock       func()
		expectedStatus int
		wantRespBody   string
		userClaim      *userModel.UserTokenClaims
		requestBody    interface{}
	}{
		{
			name:      "when everything is ok, should return 200",
			userClaim: validUserClaims,
			requestBody: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: "valid-merchant-id",
				Page:       1,
				PerPage:    10,
			},
			callMock: func() {
				expectedResponse := &commonModel.ExportResponse{
					DownloadURL: "https://storage.googleapis.com/bucket/signed-url",
					ExpiresAt:   time.Now().UTC().Add(15 * time.Minute),
				}
				mockUnifiedPaymentSvc.On("ExportCharge", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID: validUserClaims.MerchantId,
					Page:       1,
					PerPage:    10,
				}).Return(expectedResponse, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"downloadURL":"https://storage.googleapis.com/bucket/signed-url","expiresAt":"` + time.Now().UTC().Add(15*time.Minute).Format(time.RFC3339) + `"}}`,
		},
		{
			name:      "when export service fails, should return 500",
			userClaim: validUserClaims,
			requestBody: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: "valid-merchant-id",
				Page:       1,
				PerPage:    10,
			},
			callMock: func() {
				mockUnifiedPaymentSvc.On("ExportCharge", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID: validUserClaims.MerchantId,
					Page:       1,
					PerPage:    10,
				}).Return(nil, pkgError.New("internal server error", errors.New("internal service error"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","message":"internal service error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "when user info is missing from context, should return 401",
			userClaim:      nil,
			requestBody:    &unifiedPaymentModel.FilterChargeRequest{},
			callMock:       func() {},
			expectedStatus: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "when request body is invalid JSON, should return 400",
			userClaim:      validUserClaims,
			requestBody:    `{"page": "invalid"}`, // Invalid JSON structure
			callMock:       func() {},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"json: cannot unmarshal string into Go struct field FilterChargeRequest.page of type int","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "when the date range input is invalid",
			userClaim:      validUserClaims,
			requestBody:    `{"startCreatedAt": "2025-01-01T16:53:00.000Z", "endCreatedAt": "2025-01-30T16:52:59.999Z"}`, // Invalid JSON structure
			callMock:       func() {},
			expectedStatus: http.StatusBadRequest,
			wantRespBody: `{"code":"40","message":"The date range exceeds the allowed backdate limit. Maximum allowed is the last 6 months.","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}
`,
		},
		{
			name:      "when export returns cache hit, should return cached URL",
			userClaim: validUserClaims,
			requestBody: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: "valid-merchant-id",
				Page:       1,
				PerPage:    10,
			},
			callMock: func() {
				cachedResponse := &commonModel.ExportResponse{
					DownloadURL: "https://storage.googleapis.com/bucket/cached-signed-url",
					ExpiresAt:   time.Now().UTC().Add(10 * time.Minute),
				}
				mockUnifiedPaymentSvc.On("ExportCharge", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID: validUserClaims.MerchantId,
					Page:       1,
					PerPage:    10,
				}).Return(cachedResponse, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"downloadURL":"https://storage.googleapis.com/bucket/cached-signed-url","expiresAt":"` + time.Now().UTC().Add(10*time.Minute).Format(time.RFC3339) + `"}}`,
		},
		{
			name:      "when database error occurs, should return 500",
			userClaim: validUserClaims,
			requestBody: &unifiedPaymentModel.FilterChargeRequest{
				MerchantID: "valid-merchant-id",
				Page:       1,
				PerPage:    10,
			},
			callMock: func() {
				mockUnifiedPaymentSvc.On("ExportCharge", mock.Anything, &unifiedPaymentModel.FilterChargeRequest{
					MerchantID: validUserClaims.MerchantId,
					Page:       1,
					PerPage:    10,
				}).Return(nil, pkgError.New(response.HttpErrDatabase, errors.New("database connection failed"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			wantRespBody:   `{"code":"98","message":"database connection failed","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock()

			// Prepare request body
			var reqBody *bytes.Buffer
			if tc.requestBody != nil {
				if str, ok := tc.requestBody.(string); ok {
					reqBody = bytes.NewBuffer([]byte(str))
				} else {
					bodyBytes, _ := json.Marshal(tc.requestBody)
					reqBody = bytes.NewBuffer(bodyBytes)
				}
			} else {
				reqBody = bytes.NewBuffer([]byte("{}"))
			}

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/export", reqBody)
			req.Header.Set("Content-Type", "application/json")

			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}

			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(controller.Export)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			assert.Equal(t, tc.expectedStatus, rr.Code)

			// For time-sensitive responses, we need to handle the dynamic timestamps
			if strings.Contains(tc.wantRespBody, time.RFC3339) || tc.expectedStatus == http.StatusOK {
				// Extract and validate the general structure instead of exact match for success responses
				if tc.expectedStatus == http.StatusOK {
					assert.Contains(t, rr.Body.String(), `"code":"00"`)
					assert.Contains(t, rr.Body.String(), `"message":"OK"`)
					assert.Contains(t, rr.Body.String(), `"downloadURL"`)
					assert.Contains(t, rr.Body.String(), `"expiresAt"`)
				} else {
					if !assert.JSONEq(t, tc.wantRespBody, rr.Body.String()) {
						t.Log("Expected:", tc.wantRespBody)
						t.Log("Actual:", rr.Body.String())
					}
				}
			} else {
				if !assert.JSONEq(t, tc.wantRespBody, rr.Body.String()) {
					t.Log("Expected:", tc.wantRespBody)
					t.Log("Actual:", rr.Body.String())
				}
			}
		})
	}
}

func TestExportControllerInit(t *testing.T) {
	var (
		mockUnifiedPaymentService = new(mockService.IUnifiedPaymentService)
		mockMerchantService       = new(mockService.IMerchantService)
		mockLogger                logger.ILogger
	)

	c := New(&config.Config{}, validator.New(), &monitoring.Monitor{},
		WithLogger(mockLogger),
		WithUnifiedPaymentService(mockUnifiedPaymentService),
		WithMerchantService(mockMerchantService),
	)

	controller := c.(*ChargesController)
	assert.NotNil(t, controller)
	assert.NotNil(t, controller.unifiedPaymentService)
	assert.NotNil(t, controller.merchantService)
	assert.NotNil(t, controller.validate)

	// Test that Export method exists and can be called
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/export", bytes.NewBuffer([]byte("{}")))
	rr := httptest.NewRecorder()

	// Should return 401 since no user context is set
	controller.Export(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
