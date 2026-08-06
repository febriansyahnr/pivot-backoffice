package crmXbController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReConfirm(t *testing.T) {
	svc := serviceMocks.NewIXbPayoutService(t)
	mockLogger := loggerMocks.NewILogger(t)

	validPayoutID := uuid.NewString()
	validMerchantID := uuid.NewString()

	validReConfirmEvent := &xbModel.ReConfirmEvent{
		NeedAutoConfirm: false,
		MerchantID:      validMerchantID,
	}

	validReConfirmEventWithAutoConfirm := &xbModel.ReConfirmEvent{
		NeedAutoConfirm: true,
		MerchantID:      validMerchantID,
	}

	validConfirmResponse := &xbModel.ConfirmPayoutResponse{
		Uuid:       validPayoutID,
		MerchantId: validMerchantID,
	}

	validResponseInJson, err := json.Marshal(validReConfirmEvent)
	require.NoError(t, err)

	validResponseWithAutoConfirmInJson, err := json.Marshal(validReConfirmEventWithAutoConfirm)
	require.NoError(t, err)

	tests := []struct {
		name           string
		payoutID       string
		modifierMock   func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:     "ERROR: Invalid PayoutID format",
			payoutID: "invalid-uuid",
			modifierMock: func() {
				// No service calls expected
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"field_required","message":"Mandatory field is missing","error":{"type":"API_ERROR","traceId":"","details":[{"field":"id","message":"Make sure id value is fulfilled"}]}}`,
		},
		{
			name:     "ERROR: ReConfirm service error",
			payoutID: validPayoutID,
			modifierMock: func() {
				svc.On(
					"ReConfirm",
					constant.ValueCtxMockType(),
					&xbModel.ConfirmPayoutRequest{
						PayoutId: validPayoutID,
					},
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","traceId":"","details":[{"field":"","message":"Please contact our representative team"}]}}`,
		},
		{
			name:     "SUCCESS: NeedAutoConfirm is false",
			payoutID: validPayoutID,
			modifierMock: func() {
				svc.On(
					"ReConfirm",
					constant.ValueCtxMockType(),
					&xbModel.ConfirmPayoutRequest{
						PayoutId: validPayoutID,
					},
				).Once().Return(validReConfirmEvent, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseInJson) + `}`,
		},
		{
			name:     "SUCCESS: NeedAutoConfirm is true and Confirm succeeds",
			payoutID: validPayoutID,
			modifierMock: func() {
				svc.On(
					"ReConfirm",
					constant.ValueCtxMockType(),
					&xbModel.ConfirmPayoutRequest{
						PayoutId: validPayoutID,
					},
				).Once().Return(validReConfirmEventWithAutoConfirm, nil)

				svc.On(
					"Confirm",
					constant.ValueCtxMockType(),
					&xbModel.ConfirmPayoutRequest{
						PayoutId:   validPayoutID,
						MerchantId: validMerchantID,
					},
				).Once().Return(validConfirmResponse, nil)

				mockLogger.On("Info", constant.ValueCtxMockType(), "confirm payout", mock.Anything).Once()
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":` + string(validResponseWithAutoConfirmInJson) + `}`,
		},
		{
			name:     "ERROR: NeedAutoConfirm is true but Confirm fails",
			payoutID: validPayoutID,
			modifierMock: func() {
				svc.On(
					"ReConfirm",
					constant.ValueCtxMockType(),
					&xbModel.ConfirmPayoutRequest{
						PayoutId: validPayoutID,
					},
				).Once().Return(validReConfirmEventWithAutoConfirm, nil)

				svc.On(
					"Confirm",
					constant.ValueCtxMockType(),
					&xbModel.ConfirmPayoutRequest{
						PayoutId:   validPayoutID,
						MerchantId: validMerchantID,
					},
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"general_error","message":"General error","error":{"type":"API_ERROR","traceId":"","details":[{"field":"","message":"Please contact our representative team"}]}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modifierMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/xb/payout/%s/reconfirm", test.payoutID), nil)

			router := chi.NewRouter()
			router.Post("/xb/payout/{id}/reconfirm", New(svc, WithLogger(mockLogger)).ReConfirm)

			router.ServeHTTP(rec, req)

			require.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			assert.JSONEq(t, test.wantRespBody, rec.Body.String())
		})
	}
}
