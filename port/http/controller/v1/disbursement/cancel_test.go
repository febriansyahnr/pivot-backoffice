package disbursementController

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestControllerCancel(t *testing.T) {
	merchantId := "b60b4c2e-d9a3-4fed-8b26-04050023ec59"
	validUserClaims := &user.UserTokenClaims{
		UUID:       uuid.NewString(),
		MerchantId: merchantId,
		Role:       constant.RoleMaker,
	}

	// Valid payload with BatchBulkID
	validPayloadWithBulkID := disbursementModel.CancelPayoutRequest{
		BatchBulkID: []string{"bulk-id-1", "bulk-id-2"},
	}
	rawValidPayloadWithBulkID, err := json.Marshal(validPayloadWithBulkID)
	require.NoError(t, err)

	// Valid payload with BatchID
	batchIDs := []string{"batch-id-1", "batch-id-2"}
	validPayloadWithBatchID := disbursementModel.CancelPayoutRequest{
		BatchID: batchIDs,
	}
	rawValidPayloadWithBatchID, err := json.Marshal(validPayloadWithBatchID)
	require.NoError(t, err)

	// Valid payload with both BatchBulkID and BatchID
	bulkIDs := []string{"bulk-id-1", "bulk-id-2"}
	validPayloadWithBoth := disbursementModel.CancelPayoutRequest{
		BatchBulkID: bulkIDs,
		BatchID:     batchIDs,
	}
	rawValidPayloadWithBoth, err := json.Marshal(validPayloadWithBoth)
	require.NoError(t, err)

	testCases := []struct {
		name             string
		mockSetup        func(disbursementSvc *mocks.IDisbursementService)
		setupBody        func(*testing.T) []byte
		expectedStatus   int
		expectedResponse string
		userClaim        *user.UserTokenClaims
	}{
		{
			name: "SUCCESS - Cancel with BatchBulkID only",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"Cancel",
					constant.ValueCtxMockType(),
					&disbursementModel.CancelPayoutRequest{
						MerchantID:  merchantId,
						BatchBulkID: []string{"bulk-id-1", "bulk-id-2"},
					},
				).Return([]string{"cancelled-1", "cancelled-2"}, nil)
			},
			setupBody:        func(*testing.T) []byte { return rawValidPayloadWithBulkID },
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"total":2,"cancelledIds":["cancelled-1","cancelled-2"]}}`,
			userClaim:        validUserClaims,
		},
		{
			name: "SUCCESS - Cancel with BatchID only",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"Cancel",
					constant.ValueCtxMockType(),
					&disbursementModel.CancelPayoutRequest{
						MerchantID: merchantId,
						BatchID:    batchIDs,
					},
				).Return([]string{"cancelled-3", "cancelled-4"}, nil)
			},
			setupBody:      func(*testing.T) []byte { return rawValidPayloadWithBatchID },
			expectedStatus: http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"total":2,"cancelledIds":["cancelled-3","cancelled-4"]}}
`,
			userClaim: validUserClaims,
		},
		{
			name: "SUCCESS - Cancel with both BatchBulkID and BatchID",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"Cancel",
					constant.ValueCtxMockType(),
					&disbursementModel.CancelPayoutRequest{
						MerchantID:  merchantId,
						BatchBulkID: bulkIDs,
						BatchID:     batchIDs,
					},
				).Return([]string{"cancelled-5", "cancelled-6"}, nil)
			},
			setupBody:        func(*testing.T) []byte { return rawValidPayloadWithBoth },
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"total":2,"cancelledIds":["cancelled-5","cancelled-6"]}}`,
			userClaim:        validUserClaims,
		},
		{
			name: "ERROR - User not in Context",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				// No mock setup needed for this case
			},
			setupBody: func(t *testing.T) []byte {
				return rawValidPayloadWithBulkID
			},
			expectedStatus:   http.StatusUnauthorized,
			expectedResponse: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			userClaim:        nil,
		},
		{
			name: "ERROR - Invalid JSON payload",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				// No mock setup needed for this case
			},
			setupBody: func(t *testing.T) []byte {
				return []byte("{invalid json}")
			},
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":"40","message":"invalid character 'i' looking for beginning of object key string","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
			userClaim:        validUserClaims,
		},
		{
			name: "SUCCESS - Empty payload (no BatchBulkID or BatchID) - service handles validation",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"Cancel",
					constant.ValueCtxMockType(),
					&disbursementModel.CancelPayoutRequest{
						MerchantID: merchantId,
					},
				).Return([]string{}, nil)
			},
			setupBody: func(t *testing.T) []byte {
				emptyPayload := disbursementModel.CancelPayoutRequest{}
				payloadBytes, err := json.Marshal(emptyPayload)
				assert.NoError(t, err)
				return payloadBytes
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"total":0,"cancelledIds":[]}}`,
			userClaim:        validUserClaims,
		},
		{
			name: "SUCCESS - Empty arrays for both BatchBulkID and BatchID - service handles validation",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				emptyBulkIDs := []string{}
				emptyBatchIDs := []string{}
				disbursementSvc.On(
					"Cancel",
					constant.ValueCtxMockType(),
					&disbursementModel.CancelPayoutRequest{
						MerchantID:  merchantId,
						BatchBulkID: emptyBulkIDs,
						BatchID:     emptyBatchIDs,
					},
				).Return([]string{}, nil)
			},
			setupBody: func(t *testing.T) []byte {
				emptyArraysPayload := disbursementModel.CancelPayoutRequest{
					BatchBulkID: []string{},
					BatchID:     []string{},
				}
				payloadBytes, err := json.Marshal(emptyArraysPayload)
				assert.NoError(t, err)
				return payloadBytes
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"code":"00","message":"OK","data":{"total":0,"cancelledIds":[]}}`,
			userClaim:        validUserClaims,
		},
		{
			name: "ERROR - Cancel service returns error",
			mockSetup: func(disbursementSvc *mocks.IDisbursementService) {
				disbursementSvc.On(
					"Cancel",
					constant.ValueCtxMockType(),
					&disbursementModel.CancelPayoutRequest{
						MerchantID:  merchantId,
						BatchBulkID: []string{"bulk-id-1", "bulk-id-2"},
					},
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			setupBody:        func(*testing.T) []byte { return rawValidPayloadWithBulkID },
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":"99","message":"some error","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
			userClaim:        validUserClaims,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
				Environment: constant.EnvironmentStaging,
			}

			mockValidator := validator.New()
			disbursementSvc := mocks.NewIDisbursementService(t)

			tt.mockSetup(disbursementSvc)

			// Statsd Monitoring
			monitor, err := monitoring.New("backend-portal", "0.0.0.0", "1234")
			if err != nil {
				fmt.Printf("Unable to init monitoring, %v", err)
				panic(err)
			}

			mc := New(cfg, mockValidator, monitor, Services{DisbursementSvc: disbursementSvc}, nil, nil)

			// Create the HTTP request for the test case
			req := httptest.NewRequest(http.MethodPost, "/api/v1/disbursements/cancel", bytes.NewBuffer(tt.setupBody(t)))
			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}
			rr := httptest.NewRecorder()

			// Create the handler and serve the request
			handler := http.HandlerFunc(mc.Cancel)
			handler.ServeHTTP(rr, req)

			if rr.Body.String() != "" {
				t.Logf("Handler response body: %s", rr.Body.String())
			}

			// Assertions
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if !assert.JSONEq(t, tt.expectedResponse, rr.Body.String()) {
				t.Log("Output:", rr.Body.String())
			}

			disbursementSvc.AssertExpectations(t)
		})
	}
}
