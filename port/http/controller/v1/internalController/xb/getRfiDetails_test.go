package internalXbController

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
)

func TestGetRfiDetails(t *testing.T) {
	validPayoutID := "35849378-a1c1-4cf6-bbdc-c651b6cdd054"
	validMerchantID := "4f0f1537-bf0f-49ec-88eb-04e67f779e5b"

	cfg := &config.Config{}
	xbPayoutSvc := serviceMock.NewIXbPayoutService(t)

	controller := New(cfg, WithXbPayoutService(xbPayoutSvc))

	validResponse := &xbModel.GetRfiDetailsResponse{
		Uuid:       validPayoutID,
		MerchantId: validMerchantID,
	}
	validRequestID := func(req *http.Request) {
		*req = *req.WithContext(context.WithValue(req.Context(), c.CtxMerchantInfo, &merchantModel.MerchantAuthTokenClaims{
			MerchantId: "12345",
		}))
		req.Header.Set("Content-Type", "application/json")
	}

	tests := []struct {
		name           string
		PayoutID       string
		setupBody      func(*testing.T) []byte
		reqSetting     func(r *http.Request)
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name: "ERROR: invalid merchant ID",
			modifierMock: func() {
				// empty modifier
			},
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"merchant_not_found","error":{"details":[{"field":"","message":"Invalid Merchant request"}],"traceId":"","type":"API_ERROR"},"message":"Merchant not found"}`,
		},
		{
			name:     "ERROR: Invalid PayoutID format",
			PayoutID: "invalid",
			setupBody: func(t *testing.T) []byte {
				return nil
			},
			modifierMock: func() {
				// empty modifier
			},
			reqSetting:     validRequestID,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_required", "error":{"details":[{"field":"id", "message":"Make sure id value is fulfilled"}], "traceId":"", "type":"API_ERROR"}, "message":"Mandatory field is missing"}`,
		},
		{
			name:     "ERROR: GetRfiDetails service error",
			PayoutID: validPayoutID,
			setupBody: func(t *testing.T) []byte {
				payload := xbModel.GetRfiDetailsRequest{
					PayoutId:   validPayoutID,
					MerchantId: validMerchantID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				xbPayoutSvc.On(
					"GetRfiDetails",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetRfiDetailsRequest"),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			reqSetting:     validRequestID,
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error", "error":{"details":[{"field":"", "message":"Please contact our representative team"}], "traceId":"", "type":"API_ERROR"}, "message":"General error"}`,
		},
		{
			name:     "SUCCESS",
			PayoutID: validPayoutID,
			setupBody: func(t *testing.T) []byte {
				payload := xbModel.GetRfiDetailsRequest{
					PayoutId:   validPayoutID,
					MerchantId: validMerchantID,
				}
				payloadRequestByte, err := json.Marshal(payload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			modifierMock: func() {
				xbPayoutSvc.On(
					"GetRfiDetails",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*xbModel.GetRfiDetailsRequest"),
				).Return(validResponse, nil)
			},
			reqSetting:     validRequestID,
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00", "data":{"merchantId":"4f0f1537-bf0f-49ec-88eb-04e67f779e5b", "referenceId":"", "uuid":"35849378-a1c1-4cf6-bbdc-c651b6cdd054"}, "message":"Success"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/xb/payout/%s/get-rfi", test.PayoutID), nil)

			if test.reqSetting != nil {
				test.reqSetting(req)
			}

			router := chi.NewRouter()
			router.Get("/xb/payout/{id}/get-rfi", controller.GetRfiDetails)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
