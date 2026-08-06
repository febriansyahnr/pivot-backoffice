package processorCallbackController

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDanaDisbursementCallback(t *testing.T) {
	tests := []struct {
		name           string
		request        func() io.Reader
		setupMock      func(req *http.Request, svc *mockService.IDisbursementService)
		expectedStatus int
	}{
		{
			name:           "success",
			expectedStatus: http.StatusOK,
			setupMock: func(req *http.Request, svc *mockService.IDisbursementService) {
				req.Header.Set("Content-Type", "application/json")

				svc.On(
					"ProcessUpdateTransferStatus",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferResponseData"),
				).Return(nil)
			},
			request: func() io.Reader {
				return strings.NewReader(`{
						"originalPartnerReferenceNo": "2020102900000000000001",
						"originalReferenceNo": "2020102977770000000009",
						"latestTransactionStatus": "00",
						"transactionStatusDesc": "success",
						"createdTime": "2020-12-21T17:48:41+07:00",
						"finishedTime": "2020-12-21T17:50:41+07:00",
						"additionalInfo": {}
					}`)
			},
		},
		{
			name:           "disbursement callback process error",
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(req *http.Request, svc *mockService.IDisbursementService) {
				req.Header.Set("Content-Type", "application/json")

				svc.On(
					"ProcessUpdateTransferStatus",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.BankTransferResponseData"),
				).Return(errors.New("some error"))
			},
			request: func() io.Reader {
				return strings.NewReader(`{
						"originalPartnerReferenceNo": "2020102900000000000001",
						"originalReferenceNo": "2020102977770000000009",
						"latestTransactionStatus": "00",
						"transactionStatusDesc": "success",
						"createdTime": "2020-12-21T17:48:41+07:00",
						"finishedTime": "2020-12-21T17:50:41+07:00",
						"additionalInfo": {}
					}`)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			disbursementSvc := mockService.NewIDisbursementService(t)
			req, err := http.NewRequest("POST", "/v1.0/debit/emoney/transfer-bank/notify.htm", tc.request())
			if err != nil {
				t.Fatal(err)
			}

			tc.setupMock(req, disbursementSvc)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctrl := &processorCallbackController{
					disbursementSvc: disbursementSvc,
				}

				ctrl.DanaDisbursementCallback(w, r)
			})

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}
