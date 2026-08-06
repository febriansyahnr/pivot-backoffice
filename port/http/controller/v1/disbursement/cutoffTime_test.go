package disbursementController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/disbursement"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCutOffTimeStatus(t *testing.T) {
	disbursementSvc := serviceMocks.NewIDisbursementService(t)

	config := &config.Config{
		DisbursementConfig: config.DisbursementConfig{
			CutOffTimeWindow: config.DisbursementCutOffTimeWindow{
				BannerShowBeforeMinute: 15,
				StartTime:              "00:00",
				EndTime:                "00:10",
			},
		},
	}
	handler := New(config, nil, nil, Services{DisbursementSvc: disbursementSvc}, nil, nil)

	router := chi.NewRouter()
	router.Get("/cut-off-time-status", handler.GetCutOffTimeStatus)

	userClaims := &user.UserTokenClaims{
		MerchantId: "ceba442d-54a6-41c3-8825-501d665bc8f6",
	}

	execTime := "2025-01-15T23:53:57+07:00"

	tests := []struct {
		name           string
		time           string
		userClaims     *user.UserTokenClaims
		isEarlyCheck   bool
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:           "ERROR:Invalid start time window format",
			userClaims:     userClaims,
			isEarlyCheck:   true,
			setupMock:      func() { config.DisbursementConfig.CutOffTimeWindow.StartTime = "" },
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid start time window format"),
		},
		{
			name:       "ERROR:Some error",
			userClaims: userClaims,
			setupMock: func() {
				config.DisbursementConfig.CutOffTimeWindow.StartTime = "00:00"
				disbursementSvc.On(
					"GetCutOffTimeStatus", c.ValueCtxMockType(), c.TimeMockType(), c.StringMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:         "when early check was enabled and got onGoing in warning state, then should return the result",
			isEarlyCheck: true,
			userClaims:   userClaims,
			setupMock: func() {
				disbursementSvc.On(
					"GetCutOffTimeStatus", c.ValueCtxMockType(), c.TimeMockType(), c.StringMockType(), mock.Anything,
				).Return(&disbursementModel.CutOffTimeStatusResponse{
					Status: c.DisbursementCutOffTimeStatusOngoing, Time: execTime,
				}, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"status":"ONGOING","time":"2025-01-15T23:53:57+07:00"}}`,
		},
		{
			name:       "when early check was disabled and got onGoing in warning state, then should return the result",
			userClaims: userClaims,
			setupMock: func() {
				disbursementSvc.On(
					"GetCutOffTimeStatus", c.ValueCtxMockType(), c.TimeMockType(), c.StringMockType(), mock.Anything,
				).Return(&disbursementModel.CutOffTimeStatusResponse{
					Status: c.DisbursementCutOffTimeStatusOngoing, Time: execTime,
				}, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"status":"ONGOING","time":"2025-01-15T23:53:57+07:00"}}`,
		},
		{
			name:       "when got whitelable in warning state, then should return the result",
			userClaims: userClaims,
			setupMock: func() {
				disbursementSvc.On(
					"GetCutOffTimeStatus", c.ValueCtxMockType(), c.TimeMockType(), c.StringMockType(), mock.Anything,
				).Return(&disbursementModel.CutOffTimeStatusResponse{
					Status: c.DisbursementCutOffTimeStatusWhitelisted, Time: execTime,
				}, nil).Once()
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"status":"WHITELISTED","time":"2025-01-15T23:53:57+07:00"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			endpoint := "/cut-off-time-status"
			rec := httptest.NewRecorder()

			if test.isEarlyCheck {
				endpoint += "?earlyCheck=true"
			}

			req := httptest.NewRequest(http.MethodGet, endpoint, nil)

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
