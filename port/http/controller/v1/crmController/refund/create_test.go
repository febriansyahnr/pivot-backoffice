package v1CRMRefundController_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/refund"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreate(t *testing.T) {
	refundSvc := serviceMocks.NewIRefundService(t)
	cfg := &config.Config{}
	controller := New(cfg, WithRefundService(refundSvc))

	validRequest := &refundModel.CreateRefundThroughCRMRequest{
		CreateRefundRequest: refundModel.CreateRefundRequest{
			ClientReferenceID: "REF001",
			PaymentSessionID:  uuid.NewString(),
			Reason:            constant.RefundReasonOthers,
			IsFullAmount:      true,
		},
		MerchantID: "123456",
		Method:     constant.RefundMethodAuto,
	}

	tests := []struct {
		name         string
		setupMock    func()
		requestBody  func() []byte
		wantStatus   int
		wantResponse string
	}{
		{
			name: "ERROR: Invalid Payload",
			requestBody: func() []byte {
				return []byte(`{"missing-payload"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"40", "errors":"invalid request payload"}`,
		},
		{
			name: "ERROR: Field Format Invalid",
			requestBody: func() []byte {
				return []byte(`{"merchantId": "123", "method": "AUTO"}`)
			},
			wantStatus:   http.StatusBadRequest,
			wantResponse: `{"code":"40", "errors":{"Amount":"Key: 'CreateRefundThroughCRMRequest.CreateRefundRequest.Amount' Error:Field validation for 'Amount' failed on the 'required_if' tag", "ClientReferenceID":"Key: 'CreateRefundThroughCRMRequest.CreateRefundRequest.ClientReferenceID' Error:Field validation for 'ClientReferenceID' failed on the 'required' tag", "PaymentSessionID":"Key: 'CreateRefundThroughCRMRequest.CreateRefundRequest.PaymentSessionID' Error:Field validation for 'PaymentSessionID' failed on the 'required' tag", "Reason":"Key: 'CreateRefundThroughCRMRequest.CreateRefundRequest.Reason' Error:Field validation for 'Reason' failed on the 'required' tag"}}`,
		},
		{
			name: "ERROR: Service returns error",
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				refundSvc.On("Create", constant.ValueCtxMockType(), constant.PtrCreateRefundRequest()).
					Return(nil, errors.New("service error")).Once()
			},
			wantStatus:   http.StatusInternalServerError,
			wantResponse: `{"code":"99", "errors":"service error"}`,
		},
		{
			name: "SUCCESS",
			requestBody: func() []byte {
				reqBody, _ := json.Marshal(validRequest)
				return reqBody
			},
			setupMock: func() {
				refundSvc.On("Create", constant.ValueCtxMockType(), constant.PtrCreateRefundRequest()).
					Return(nil, nil)
			},
			wantStatus:   http.StatusOK,
			wantResponse: `{"code":"00", "data":null}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			reqBody := []byte{}
			if test.requestBody != nil {
				reqBody = test.requestBody()
			}
			req := httptest.NewRequest(http.MethodPost, "/refunds", bytes.NewBuffer(reqBody))
			rec := httptest.NewRecorder()

			ctx := req.Context()
			req = req.WithContext(ctx)

			router := chi.NewRouter()
			router.Post("/refunds", controller.Create)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantResponse, rec.Body.String())
		})
	}
}
