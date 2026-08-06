package crmDisbursementController_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/disbursement"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReversal(t *testing.T) {
	disbursementSvc := serviceMocks.NewIDisbursementService(t)

	handler := New(disbursementSvc, nil)

	router := chi.NewRouter()
	router.Post("/disbursements/{id}/reversal", handler.Reversal)

	reversalResult := &disbursementModel.ReversalTransactionResp{
		Id:             uuid.NewString(),
		DisbursementId: uuid.NewString(),
		ReversalAmount: 250_000.00,
	}
	rawReversalResult, _ := json.Marshal(reversalResult)

	tests := []struct {
		name           string
		disbursementId string
		requestBody    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid disbursement id",
			disbursementId: "AAAA",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid disbursement id"),
		},
		{
			name:           "ERROR:Invalid request body",
			requestBody:    "A",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid character 'A' looking for beginning of value"),
		},
		{
			name:           "ERROR:Empty create by field",
			requestBody:    `{"merchantId":"c12bcd47-6b1d-401b-8f03-e44b5597f97f","reason":"bla bla"}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"CreatedBy":"Key: 'ReversalTransactionReq.CreatedBy' Error:Field validation for 'CreatedBy' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR:Some error", // NOSONAR
			requestBody: `{"merchantId":"c12bcd47-6b1d-401b-8f03-e44b5597f97f","createdBy":"John", "reason":"bla bla"}`,
			setupMock: func() {
				disbursementSvc.On(
					"Reversal", c.ValueCtxMockType(), mock.AnythingOfType("*disbursementModel.ReversalTransactionReq"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS",
			requestBody: `{"merchantId":"c12bcd47-6b1d-401b-8f03-e44b5597f97f","createdBy":"John", "reason":"bla bla"}`,
			setupMock: func() {
				disbursementSvc.On(
					"Reversal", c.ValueCtxMockType(), mock.AnythingOfType("*disbursementModel.ReversalTransactionReq"),
				).Return(reversalResult, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00", "data": %s}`, string(rawReversalResult)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.disbursementId == "" {
				test.disbursementId = uuid.NewString()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/disbursements/"+test.disbursementId+"/reversal", strings.NewReader(test.requestBody))

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
