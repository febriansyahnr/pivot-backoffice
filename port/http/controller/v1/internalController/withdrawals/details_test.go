package internalWithdrawalsController_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/internalController/withdrawals"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWithdrawalByID(t *testing.T) {
	service := serviceMocks.NewIWithdrawalService(t)

	handler := New(nil, nil, WithWithdrawalService(service))

	route := chi.NewRouter()
	route.Get("/withdrawals/{id}", handler.GetWithdrawalByID)

	tests := []struct {
		name             string
		merchantAuth     *merchant.MerchantAuthTokenClaims
		withdrawalID     string
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:             "ERROR:Merchant auth not found",
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"merchant_not_found","message":"Merchant not found","error":{"type":"API_ERROR","details":[{"field":"","message":"Invalid Merchant request"}],"traceId":""}}`,
		},
		{
			name:             "ERROR:Invalid withdrawal ID format",
			withdrawalID:     "XXXX",
			merchantAuth:     &merchant.MerchantAuthTokenClaims{},
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"field_format_invalid","message":"Format Field is invalid","error":{"type":"API_ERROR","details":[{"field":"withdrawalId","message":"Make sure withdrawalId format is correct"}],"traceId":""}}`,
		},
		{
			name:         "ERROR:Data not found",
			merchantAuth: &merchant.MerchantAuthTokenClaims{},
			setupMock: func() {
				service.On("GetById", mock.Anything, mock.Anything).Once().Return(nil, constant.ErrDataNotFound)
			},
			wantStatusCode:   http.StatusNotFound,
			wantResponseBody: `{"code":"resource_missing","message":"The withdrawal detail with ID 01991460-38b6-7860-901c-317accd1cd85 cannot be found","error":{"type":"GATEWAY_ERROR","details":[{"field":"","message":"The withdrawal detail with ID 01991460-38b6-7860-901c-317accd1cd85 cannot be found"}],"traceId":""}}`,
		},
		{
			name:         "SUCCESS",
			merchantAuth: &merchant.MerchantAuthTokenClaims{},
			setupMock: func() {
				service.On("GetById", mock.Anything, mock.Anything).Once().Return(&withdrawal.WithdrawalDetailResponse{
					Id:                     "01991460-38b6-7860-901c-317accd1cd85",         // NOSONAR
					CreatedAt:              time.Date(2025, 9, 4, 10, 57, 54, 0, time.UTC), // NOSONAR
					UpdatedAt:              time.Date(2025, 9, 4, 10, 57, 54, 0, time.UTC), // NOSONAR
					Type:                   "MANUAL",                                       // NOSONAR
					Amount:                 10_000,                                         // NOSONAR
					Status:                 "SUCCESS",                                      // NOSONAR
					BeneficiaryAccountName: "Payout Balance",                               // NOSONAR
					Currency:               "IDR",                                          // NOSONAR
					MerchantID:             "aec6636d-7a02-4d93-a4c5-006b9c235068",         // NOSONAR
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"id":"01991460-38b6-7860-901c-317accd1cd85","merchantId":"aec6636d-7a02-4d93-a4c5-006b9c235068","withdrawal":{"referenceId":"","withdrawType":"BALANCE_TRANSFER","balanceType":"PAYOUT_BALANCE","isFullAmount":false,"amount":{"currency":"IDR","value":"10000"},"description":""},"status":"SUCCESS","createdAt":"2025-09-04T10:57:54Z","updatedAt":"2025-09-04T10:57:54Z"}}
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.withdrawalID == "" {
				test.withdrawalID = "01991460-38b6-7860-901c-317accd1cd85" // NOSONAR
			}
			if test.setupMock != nil {
				test.setupMock()
			}
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/withdrawals/"+test.withdrawalID, nil)

			if test.merchantAuth != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxMerchantInfo, test.merchantAuth))
			}

			route.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantResponseBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
