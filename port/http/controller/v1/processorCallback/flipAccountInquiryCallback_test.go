package processorCallbackController

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFlipInquiryAccountCallback(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    func() io.Reader
		setupMock      func(req *http.Request, svc *mockService.IRoutingProcessorService)
		expectedStatus int
	}{
		{
			name:           "success",
			expectedStatus: http.StatusOK,
			setupMock: func(req *http.Request, svc *mockService.IRoutingProcessorService) {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				svc.On(
					"ProcessAccountInquiryCallback",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountResponseData"),
				).Return(nil)
			},
			requestBody: func() io.Reader {
				data := []byte(`{"id":218498,"user_id":15774748,"amount":100000,"status":"CANCELLED","reason":"","timestamp":"2024-10-23 21:29:26","bank_code":"bri","account_number":"999966660001","recipient_name":"Dummy Name","sender_bank":null,"remark":"ucup percobaan tra","receipt":"","time_served":"(not set)","bundle_id":0,"company_id":76778,"recipient_city":391,"created_from":"API","direction":"DOMESTIC_SPECIAL_TRANSFER","sender":{"sender_name":"PT Harsya Remitindo","sender_place_of_birth":0,"sender_date_of_birth":"","sender_address":"Biomedical Campus, Knowledge Tower Lt. 3, Kav. Digital Hub","sender_identity_type":"bank_acc","sender_identity_number":"999966660001","sender_country":100252,"sender_job":"company"},"fee":2400,"beneficiary_email":"","idempotency_key":"8764b1b8-0fb5-4fe1-baaf-d00f6ff7128f"}`)
				values := url.Values{}
				values.Add("data", string(data))
				return strings.NewReader(values.Encode())
			},
		},
		{
			name:           "inquiry account callback process error",
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(req *http.Request, svc *mockService.IRoutingProcessorService) {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				svc.On(
					"ProcessAccountInquiryCallback",
					mock.Anything,
					mock.AnythingOfType("*routingProcessorModel.InquiryAccountResponseData"),
				).Return(errors.New("some error"))
			},
			requestBody: func() io.Reader {
				data := []byte(`{"id":218498,"user_id":15774748,"amount":100000,"status":"CANCELLED","reason":"","timestamp":"2024-10-23 21:29:26","bank_code":"bri","account_number":"999966660001","recipient_name":"Dummy Name","sender_bank":null,"remark":"ucup percobaan tra","receipt":"","time_served":"(not set)","bundle_id":0,"company_id":76778,"recipient_city":391,"created_from":"API","direction":"DOMESTIC_SPECIAL_TRANSFER","sender":{"sender_name":"PT Harsya Remitindo","sender_place_of_birth":0,"sender_date_of_birth":"","sender_address":"Biomedical Campus, Knowledge Tower Lt. 3, Kav. Digital Hub","sender_identity_type":"bank_acc","sender_identity_number":"999966660001","sender_country":100252,"sender_job":"company"},"fee":2400,"beneficiary_email":"","idempotency_key":"8764b1b8-0fb5-4fe1-baaf-d00f6ff7128f"}`)
				values := url.Values{}
				values.Add("data", string(data))
				return strings.NewReader(values.Encode())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			disbursementSvc := mockService.NewIDisbursementService(t)
			routingProcessorSvc := mockService.NewIRoutingProcessorService(t)
			req, err := http.NewRequest("POST", "/callback/v1/flip/account-inquiry", tc.requestBody())
			if err != nil {
				t.Fatal(err)
			}

			tc.setupMock(req, routingProcessorSvc)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctrl := &processorCallbackController{
					disbursementSvc:     disbursementSvc,
					routingProcessorSvc: routingProcessorSvc,
				}

				ctrl.FlipInquiryAccountCallback(w, r)
			})

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}
