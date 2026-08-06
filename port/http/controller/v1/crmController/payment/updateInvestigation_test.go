package v1CrmPaymentController_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateInvestigation(t *testing.T) {
	svc := serviceMocks.NewIPaymentService(t)
	h := New(svc)

	validPaymentID := uuid.NewString()
	validNotes := "Bank confirmed payment received"
	completedAt := time.Now().UTC()

	tests := []struct {
		name           string
		paymentID      string
		requestBody    any
		mockService    func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:      "SUCCESS: Update to INVESTIGATION_SUCCESS with notes",
			paymentID: validPaymentID,
			requestBody: paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
				Notes:               &validNotes,
			},
			mockService: func() {
				svc.On("UpdateInvestigationStatus", mock.Anything, validPaymentID, mock.MatchedBy(func(req *paymentModel.UpdateInvestigationRequest) bool {
					return req.InvestigationStatus == paymentConst.InvestigationStatusSuccess && *req.Notes == validNotes
				})).Once().Return(&paymentModel.UpdateInvestigationResponse{
					PaymentReferenceID:  validPaymentID,
					InvestigationStatus: paymentConst.InvestigationStatusSuccess,
					CompletedAt:         &completedAt,
					LastUpdatedAt:       completedAt,
					Notes:               &validNotes,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"paymentReferenceId":"%s","investigationStatus":"INVESTIGATION_SUCCESS","completedAt":"%s","lastUpdatedAt":"%s","notes":"%s"},"message":"OK"}`, validPaymentID, completedAt.Format(time.RFC3339Nano), completedAt.Format(time.RFC3339Nano), validNotes),
		},
		{
			name:      "SUCCESS: Update to INVESTIGATION_FAILED",
			paymentID: validPaymentID,
			requestBody: paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusFailed,
			},
			mockService: func() {
				svc.On("UpdateInvestigationStatus", mock.Anything, validPaymentID, mock.MatchedBy(func(req *paymentModel.UpdateInvestigationRequest) bool {
					return req.InvestigationStatus == paymentConst.InvestigationStatusFailed
				})).Once().Return(&paymentModel.UpdateInvestigationResponse{
					PaymentReferenceID:  validPaymentID,
					InvestigationStatus: paymentConst.InvestigationStatusFailed,
					CompletedAt:         &completedAt,
					LastUpdatedAt:       completedAt,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"paymentReferenceId":"%s","investigationStatus":"INVESTIGATION_FAILED","completedAt":"%s","lastUpdatedAt":"%s","notes":null},"message":"OK"}`, validPaymentID, completedAt.Format(time.RFC3339Nano), completedAt.Format(time.RFC3339Nano)),
		},
		{
			name:           "ERROR: Invalid paymentID format",
			paymentID:      "invalid-uuid",
			requestBody:    paymentModel.UpdateInvestigationRequest{InvestigationStatus: paymentConst.InvestigationStatusSuccess},
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"payment id is not valid"}`,
		},
		{
			name:           "ERROR: Empty paymentID",
			paymentID:      "",
			requestBody:    paymentModel.UpdateInvestigationRequest{InvestigationStatus: paymentConst.InvestigationStatusSuccess},
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"payment id is not valid"}`,
		},
		{
			name:           "ERROR: Invalid JSON payload",
			paymentID:      validPaymentID,
			requestBody:    "invalid-json",
			mockService:    func() {},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid request payload"}`,
		},
		{
			name:      "ERROR: Service - Payment not found",
			paymentID: validPaymentID,
			requestBody: paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockService: func() {
				svc.On("UpdateInvestigationStatus", mock.Anything, validPaymentID, mock.Anything).
					Once().Return(nil, pkgErr.New(httpResponse.HttpErrNotFound, constant.ErrPaymentNotFound))
			},
			wantStatusCode: http.StatusNotFound,
			wantRespBody:   `{"code":"44","errors":"payment not found"}`,
		},
		{
			name:      "ERROR: Service - Investigation already finalized",
			paymentID: validPaymentID,
			requestBody: paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockService: func() {
				svc.On("UpdateInvestigationStatus", mock.Anything, validPaymentID, mock.Anything).
					Once().Return(nil, pkgErr.New(httpResponse.HttpErrRequest, constant.ErrInvestigationAlreadyFinalized))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"investigation is already finalized and cannot be modified"}`,
		},
		{
			name:      "ERROR: Service - Payment not under investigation",
			paymentID: validPaymentID,
			requestBody: paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockService: func() {
				svc.On("UpdateInvestigationStatus", mock.Anything, validPaymentID, mock.Anything).
					Once().Return(nil, pkgErr.New(httpResponse.HttpErrRequest, constant.ErrInvestigationNotFound))
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"payment is not under investigation"}`,
		},
		{
			name:      "ERROR: Service - Internal error",
			paymentID: validPaymentID,
			requestBody: paymentModel.UpdateInvestigationRequest{
				InvestigationStatus: paymentConst.InvestigationStatusSuccess,
			},
			mockService: func() {
				svc.On("UpdateInvestigationStatus", mock.Anything, validPaymentID, mock.Anything).
					Once().Return(nil, pkgErr.New(httpResponse.HttpErrDatabase, constant.ErrSomeErrorForUnitTest))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"98","errors":"some error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockService()

			var body []byte
			if str, ok := test.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(test.requestBody)
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/payments/%s/investigation", test.paymentID), bytes.NewReader(body))

			router := chi.NewRouter()
			router.Post("/payments/{id}/investigation", h.UpdateInvestigation)
			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
