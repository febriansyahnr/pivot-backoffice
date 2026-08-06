package xbPayoutController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/xbPayout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetList(t *testing.T) {
	cfg := &config.Config{}
	disbursementSvc := serviceMock.NewIDisbursementService(t)

	ctrl := New(cfg, WithDisbursementService(disbursementSvc))

	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, &userModel.UserTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	validResponse := &commonModel.PaginationResponse{
		Data: []*disbursementModel.DisbursementWithTransactionResponse{},
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    10,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	tests := []struct {
		name             string
		mockSetup        func()
		reqSetting       func(r *http.Request)
		expectedStatus   int
		expectedRespBody string
	}{
		{
			name: "ERROR: Invalid user info",
			mockSetup: func() {
				// empty modifier
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedRespBody: `{"code":"41", "data":null, "error":{"details":[], "traceId":"", "type":"API_ERROR"}, "message":"user not found"}`,
		},
		{
			name: "ERROR: GetList service error",
			mockSetup: func() {
				disbursementSvc.On("GetList",
					c.ValueCtxMockType(),
					mock.AnythingOfType(c.MockTypeDisbursementFilterRequestReference),
					c.Int64MockType(),
					c.Int64MockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusInternalServerError,
			expectedRespBody: `{"code":"99", "data":null, "error":{"details":[], "traceId":"", "type":"UNKNOWN"}, "message":"some error"}`,
		},
		{
			name: "SUCCESS",
			mockSetup: func() {
				disbursementSvc.On("GetList",
					c.ValueCtxMockType(),
					mock.AnythingOfType(c.MockTypeDisbursementFilterRequestReference),
					c.Int64MockType(),
					c.Int64MockType(),
				).Return(validResponse, nil)
			},
			reqSetting:       validRequestID,
			expectedStatus:   http.StatusOK,
			expectedRespBody: `{"code":"00", "data":[], "message":"OK", "pagination":{"page":1, "perPage":10, "totalItems":1, "totalPages":1}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/xb/payout?page=1&perPage=10&startDate=2024-09-01&endDate=2024-09-30", nil)

			if tc.reqSetting != nil {
				tc.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/api/v1/xb/payout", ctrl.GetList)

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.JSONEq(t, tc.expectedRespBody, rec.Body.String())
		})
	}
}
