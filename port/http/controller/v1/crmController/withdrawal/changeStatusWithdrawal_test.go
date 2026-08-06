package withdrawalCrmController_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	withdrawalModel "github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/withdrawal"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChangeStatusWithdrawal(t *testing.T) {
	service := serviceMocks.NewIWithdrawalService(t)

	handler := New(validator.New(), service)

	router := chi.NewRouter()
	router.Post("/withdrawals/{id}/change-status", handler.ChangeStatusWithdrawal)

	merchantId := uuid.NewString()
	withdrawalId := uuid.NewString()
	requestBody := fmt.Sprintf(`{"merchantId": "%s", "status": "SUCCESS"}`, merchantId)

	tests := []struct {
		name           string
		id             string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid UUID",
			id:             "not-a-uuid",
			requestBody:    requestBody,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "id is required"),
		},
		{
			name:           "ERROR: Invalid request body",
			id:             withdrawalId,
			requestBody:    "invalid-json",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid character 'i' looking for beginning of value"),
		},
		{
			name:           "ERROR: Missing merchantId and status",
			id:             withdrawalId,
			requestBody:    `{}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantID":"Key: 'WithdrawalChangeStatusRequest.MerchantID' Error:Field validation for 'MerchantID' failed on the 'required' tag","Status":"Key: 'WithdrawalChangeStatusRequest.Status' Error:Field validation for 'Status' failed on the 'required' tag"}}`,
		},
		{
			name:           "ERROR: Missing reasonType when status FAILED",
			id:             withdrawalId,
			requestBody:    fmt.Sprintf(`{"merchantId": "%s", "status": "FAILED"}`, merchantId),
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"ReasonType":"Key: 'WithdrawalChangeStatusRequest.ReasonType' Error:Field validation for 'ReasonType' failed on the 'required_if' tag"}}`,
		},
		{
			name:        "ERROR: Service returns error",
			id:          withdrawalId,
			requestBody: requestBody,
			setupMock: func() {
				service.On(
					"ChangeStatusWithdrawal", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalChangeStatusRequest"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS: Change status withdrawal",
			id:          withdrawalId,
			requestBody: requestBody,
			setupMock: func() {
				service.On(
					"ChangeStatusWithdrawal", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.WithdrawalChangeStatusRequest"),
				).Once().Return(&withdrawalModel.WithdrawalChangeStatusResponse{
					ID:         withdrawalId,
					MerchantID: merchantId,
					Status:     "SUCCESS",
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"id":"%s","merchantId":"%s","status":"SUCCESS"}}`, withdrawalId, merchantId),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.requestBody == "" {
				test.requestBody = requestBody
			}

			url := "/withdrawals/" + test.id + "/change-status"
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(test.requestBody))

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
