package cardFundedPayoutController

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetReceipt(t *testing.T) {
	vld := validatorExt.New()
	service := serviceMocks.NewICardFundedPayoutService(t)

	payoutID := "6e34cfca-de00-4aa1-bc6b-069d878012c8"
	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		Name:       "John Doe",
		MerchantId: uuid.NewString(),
	}

	cfg := &config.Config{ServiceName: "testing"}
	handler := New(cfg, vld, service)

	router := chi.NewRouter()
	router.Get("/{payoutId}/receipt", handler.GetReceipt)

	testCases := []struct {
		name         string
		payoutID     string
		userClaim    *userModel.UserTokenClaims
		setupMock    func()
		wantStatus   int
		wantReceipt  bool
		receiptURL   string
	}{
		{
			name:       "ERROR: Invalid payoutId format",
			payoutID:   "invalid-uuid",
			userClaim:  validUserClaims,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ERROR: User not in Context",
			payoutID:   payoutID,
			userClaim:  nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "ERROR: Service error",
			payoutID:  payoutID,
			userClaim: validUserClaims,
			setupMock: func() {
				service.On("GetReceipt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "SUCCESS: Get receipt URL",
			payoutID:  payoutID,
			userClaim: validUserClaims,
			setupMock: func() {
				service.On("GetReceipt", mock.Anything, &model.GetReceiptRequest{
					PayoutID:   payoutID,
					MerchantID: validUserClaims.MerchantId,
				}).Once().Return(&model.GetReceiptResponse{
					ReceiptURL: "https://storage.googleapis.com/bucket/card-funded-payouts/receipt/Card_Funded_Payout_Receipt_6e34cfca-de00-4aa1-bc6b-069d878012c8.pdf?signed=xyz",
				}, nil)
			},
			wantStatus:  http.StatusOK,
			wantReceipt: true,
			receiptURL:  "https://storage.googleapis.com/bucket/card-funded-payouts/receipt/Card_Funded_Payout_Receipt_6e34cfca-de00-4aa1-bc6b-069d878012c8.pdf?signed=xyz",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/"+tt.payoutID+"/receipt", nil)

			// Set chi route context for path params
			chiCtx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Result().StatusCode)

			if tt.wantReceipt {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

				var apiResp response.ApiResponse
				err := json.Unmarshal(rec.Body.Bytes(), &apiResp)
				assert.NoError(t, err)

				data, ok := apiResp.Data.(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.receiptURL, data["receiptUrl"])
			}

			service.AssertExpectations(t)
		})
	}
}
