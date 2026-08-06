package disbursementController

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestController_ApprovalActions(t *testing.T) {
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: uuid.NewString(),
		Role:       constant.RoleApprover,
	}

	validPayload := disbursementModel.ApprovalActionsRequest{
		Approve: []disbursementModel.ApproveActionObject{},
		Reject:  []disbursementModel.RejectActionObject{},
	}

	validRaw, err := json.Marshal(validPayload)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		mockSetup      func(disbursementSvc *mocks.IDisbursementService)
		setupBody      func(*testing.T) []byte
		expectedStatus int
		userClaim      *user.UserTokenClaims
	}{
		{
			name: "SUCCESS: Normal Payout",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"ValidateBatchPayoutItems", constant.ValueCtxMockType(), mock.Anything,
				).Return(&disbursementModel.ApprovalActionsRequest{}, nil)
				disbursementSvc.On(
					"ApprovalAction", constant.ValueCtxMockType(), mock.Anything,
				).Return(&disbursementModel.ApprovalActionsResponse{}, nil)
				disbursementSvc.On(
					"GetCutOffTimeStatus", constant.ValueCtxMockType(), constant.TimeMockType(), constant.StringMockType(), mock.Anything,
				).Return(&disbursementModel.CutOffTimeStatusResponse{}, nil)
			},
			setupBody:      func(t *testing.T) []byte { return validRaw },
			expectedStatus: http.StatusOK,
			userClaim:      validUserClaims,
		},
		{
			name: "SUCCESS: Payout Cut-off Time",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"ValidateBatchPayoutItems", constant.ValueCtxMockType(), mock.Anything,
				).Return(&disbursementModel.ApprovalActionsRequest{}, nil)
				disbursementSvc.On(
					"ApprovalAction", constant.ValueCtxMockType(), mock.Anything,
				).Return(&disbursementModel.ApprovalActionsResponse{}, nil)
				disbursementSvc.On(
					"GetCutOffTimeStatus", constant.ValueCtxMockType(), constant.TimeMockType(), constant.StringMockType(), mock.Anything,
				).Return(&disbursementModel.CutOffTimeStatusResponse{
					Status: constant.DisbursementCutOffTimeStatusOngoing,
				}, nil)
			},
			setupBody:      func(t *testing.T) []byte { return validRaw },
			expectedStatus: http.StatusAccepted,
			userClaim:      validUserClaims,
		},
		{
			name:           "ERROR: User not in Context",
			mockSetup:      func(*mocks.IDisbursementService) { /* Empty Body Function */ },
			setupBody:      func(t *testing.T) []byte { return []byte{} },
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "ERROR: Invalid payload",
			mockSetup:      func(*mocks.IDisbursementService) { /* Empty Body Function */ },
			setupBody:      func(t *testing.T) []byte { return []byte("{invalid json}") },
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name:           "ERROR: Invalid data",
			mockSetup:      func(*mocks.IDisbursementService) { /* Empty Body Function */ },
			setupBody:      func(t *testing.T) []byte { return []byte(`{"approve": [{"id": "abc"}]}`) },
			expectedStatus: http.StatusBadRequest,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR: Get payout cut-off time status",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"GetCutOffTimeStatus", constant.ValueCtxMockType(), constant.TimeMockType(), constant.StringMockType(), mock.Anything,
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			setupBody:      func(t *testing.T) []byte { return validRaw },
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR: Approval action service error",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"ValidateBatchPayoutItems", constant.ValueCtxMockType(), mock.Anything,
				).Return(&disbursementModel.ApprovalActionsRequest{}, nil)
				disbursementSvc.On(
					"ApprovalAction", constant.ValueCtxMockType(), mock.Anything,
				).Return(nil, constant.ErrSomeErrorForUnitTest)
				disbursementSvc.On(
					"GetCutOffTimeStatus", constant.ValueCtxMockType(), constant.TimeMockType(), constant.StringMockType(), mock.Anything,
				).Return(&disbursementModel.CutOffTimeStatusResponse{
					Status: constant.DisbursementCutOffTimeStatusOngoing,
				}, nil)
			},
			setupBody:      func(t *testing.T) []byte { return validRaw },
			expectedStatus: http.StatusInternalServerError,
			userClaim:      validUserClaims,
		},
		{
			name: "ERROR: cleanup request action service error",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"ValidateBatchPayoutItems", constant.ValueCtxMockType(), mock.Anything,
				).Return(nil, constant.ErrSomeErrorForUnitTest)
				disbursementSvc.On(
					"GetCutOffTimeStatus", constant.ValueCtxMockType(), constant.TimeMockType(), constant.StringMockType(), mock.Anything,
				).Return(&disbursementModel.CutOffTimeStatusResponse{
					Status: constant.DisbursementCutOffTimeStatusOngoing,
				}, nil)
			},
			setupBody:      func(t *testing.T) []byte { return validRaw },
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

			mc := New(cfg, mockValidator, nil, Services{
				DisbursementSvc: disbursementSvc,
			}, nil, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/api/v1/disbursements/approval-actions", bytes.NewBuffer(tt.setupBody(t)))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.ApprovalActions)
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
