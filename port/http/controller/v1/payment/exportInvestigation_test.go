package payment_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExportInvestigation(t *testing.T) {
	service := serviceMocks.NewIPaymentService(t)
	handler := New(nil, validator.New(), nil, WithPaymentService(service))

	router := chi.NewRouter()
	router.Post("/cases/export", handler.ExportInvestigation)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
	}

	now := time.Now().UTC()
	fromDate := now.AddDate(0, 0, -7).Format(time.RFC3339)
	toDate := now.Format(time.RFC3339)

	tests := []struct {
		name            string
		userTokenClaims *user.UserTokenClaims
		requestBody     string
		setupMock       func()
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name:           "ERROR: User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:            "ERROR: Invalid request body",
			userTokenClaims: userTokenClaims,
			requestBody:     `invalid`,
			wantStatusCode:  http.StatusBadRequest,
		},
		{
			name:            "ERROR: Service error",
			userTokenClaims: userTokenClaims,
			requestBody:     fmt.Sprintf(`{"fromDate":"%s","toDate":"%s"}`, fromDate, toDate),
			setupMock: func() {
				service.On("ExportInvestigatedPayments", c.ValueCtxMockType(), mock.Anything).
					Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:            "SUCCESS: Export investigation",
			userTokenClaims: userTokenClaims,
			requestBody:     fmt.Sprintf(`{"fromDate":"%s","toDate":"%s"}`, fromDate, toDate),
			setupMock: func() {
				service.On("ExportInvestigatedPayments", c.ValueCtxMockType(), mock.Anything).
					Return(&paymentModel.InvestigationDownloadHistoryResponse{
						URL: "https://storage.googleapis.com/signed-url",
					}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"url":"https://storage.googleapis.com/signed-url"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			req := httptest.NewRequest(http.MethodPost, "/cases/export", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.userTokenClaims != nil {
				ctx := context.WithValue(req.Context(), c.CtxUserInfoKey, tt.userTokenClaims)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			if tt.wantRespBody != "" {
				assert.JSONEq(t, tt.wantRespBody, w.Body.String())
			}

			service.AssertExpectations(t)
		})
	}
}
