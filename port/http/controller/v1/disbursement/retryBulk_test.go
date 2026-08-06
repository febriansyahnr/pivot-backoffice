package disbursementController

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryBulk(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
		Role:       constant.RoleApprover,
	}

	validPayload := disbursementModel.RetryBulkRequest{
		BulkDisbursementID: uuid.NewString(),
	}

	testCases := []struct {
		name           string
		mockSetup      func(disbursementSvc *mocks.IDisbursementService)
		setupBody      func(*testing.T) []byte
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name: "SUCCESS",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.
					On(
						"RetryBulk",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*disbursementModel.RetryBulkRequest")).
					Return(nil)
			},
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name:      "ERROR: User not in Context",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {},
			setupBody: func(t *testing.T) []byte {
				return []byte{}
			},
			expectedStatus: http.StatusUnauthorized,
			userClaim:      nil,
		},
		{
			name:      "ERROR: Invalid payload",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {},
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:      "ERROR: Missing required payload",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {},
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(disbursementModel.RetryBulkRequest{})
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR: RetryBulk",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.
					On(
						"RetryBulk",
						mock.AnythingOfType(constant.MockTypeValueContextReference),
						mock.AnythingOfType("*disbursementModel.RetryBulkRequest")).
					Return(constant.ErrSomeErrorForUnitTest)
			},
			setupBody: func(t *testing.T) []byte {
				payloadRequestByte, err := json.Marshal(validPayload)
				assert.NoError(t, err)
				return payloadRequestByte
			},
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			mockValidator := validator.New()
			disbursementSvc := mocks.NewIDisbursementService(t)

			tt.mockSetup(disbursementSvc)

			mc := New(cfg, mockValidator, nil, Services{DisbursementSvc: disbursementSvc}, nil, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/api/v1/disbursements/bulk/retry", bytes.NewBuffer(tt.setupBody(t)))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.RetryBulk)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)

			disbursementSvc.AssertExpectations(t)
		})
	}
}
