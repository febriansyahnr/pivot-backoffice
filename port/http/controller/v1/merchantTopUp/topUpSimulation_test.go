package merchantTopUp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/topUpSimulation"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTopUpSimulation(t *testing.T) {
	topUpSvc := serviceMocks.NewIMerchantTopUpService(t)

	handler := New(validatorExt.New(), topUpSvc)

	router := chi.NewRouter()
	router.Post("/top-up-simulation", handler.TopUpSimulation)

	userClaims := &user.UserTokenClaims{}
	payload := &model.TopupSimulationRequest{
		VANumber: "123456",
		TotalAmount: model.Amount{
			Currency: "IDR", Value: "10000.00",
		},
	}
	rawRequest, err := json.Marshal(payload)
	require.NoError(t, err)

	response := model.TopupSimulationResponseData{
		VANumber: "123456",
		TotalAmount: model.Amount{
			Currency: "IDR", Value: "10000.00",
		},
		RequestID:   "REQ111",
		ReferenceNo: "REF222",
	}
	rawResponse, err := json.Marshal(response)
	require.NoError(t, err)

	tests := []struct {
		name           string
		userClaims     *user.UserTokenClaims
		requestBody    []byte
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","errors":"user not found"}`,
		},
		{
			name:           "ERROR:Invalid request body",
			userClaims:     userClaims,
			requestBody:    []byte(`ABC`),
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character 'A' looking for beginning of value"}`,
		},
		{
			name:           "ERROR:Required value",
			userClaims:     userClaims,
			requestBody:    []byte(`{}`),
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"Currency":"Key: 'TopupSimulationRequest.TotalAmount.Currency' Error:Field validation for 'Currency' failed on the 'required' tag","VANumber":"Key: 'TopupSimulationRequest.VANumber' Error:Field validation for 'VANumber' failed on the 'required' tag","Value":"Key: 'TopupSimulationRequest.TotalAmount.Value' Error:Field validation for 'Value' failed on the 'required' tag"}}`,
		},
		{
			name:        "ERROR:Some error",
			userClaims:  userClaims,
			requestBody: rawRequest,
			setupMock: func() {
				topUpSvc.On("CreateTopupSimulation", mock.Anything, mock.Anything).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   `{"code":"99","errors":"some error"}`,
		},
		{
			name:        "SUCCESS",
			userClaims:  userClaims,
			requestBody: rawRequest,
			setupMock: func() {
				topUpSvc.On("CreateTopupSimulation", mock.Anything, mock.Anything).Return(&response, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","message":"OK","data":%s}`, string(rawResponse)),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/top-up-simulation", bytes.NewReader(test.requestBody))

			if test.userClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, test.userClaims))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
