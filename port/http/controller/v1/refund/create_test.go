package refund

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mockService "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	mockLogger := loggerMocks.NewILogger(t)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Maybe()

	var (
		validUserClaims = &userModel.UserTokenClaims{
			UUID:       "valid-user-id",
			MerchantId: "valid-merchant-id",
		}
		mockRefundSvc = mockService.NewIRefundService(t)
		controller    = RefundController{
			refundService: mockRefundSvc,
			validate:      validatorExt.New(),
			logger:        mockLogger,
		}
		fixedTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	)

	testCases := []struct {
		name           string
		userClaim      *userModel.UserTokenClaims
		body           string
		setupMock      func()
		expectedStatus int
		wantRespBody   string
		containsCheck  string
	}{
		{
			name:           "ERROR: user info missing from context returns 401",
			userClaim:      nil,
			body:           `{}`,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			wantRespBody:   `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: invalid JSON body returns 400",
			userClaim:      validUserClaims,
			body:           `{invalid-json`,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","message":"invalid request payload","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:           "ERROR: validation failure on missing required fields returns 400",
			userClaim:      validUserClaims,
			body:           `{}`,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			containsCheck:  `"code":"40"`,
		},
		{
			name:           "ERROR: validation failure on invalid method returns 400",
			userClaim:      validUserClaims,
			body:           `{"clientReferenceId":"ref-1","paymentSessionId":"sess-1","isFullAmount":true,"reason":"DUPLICATE","method":"INVALID"}`,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			containsCheck:  `"code":"40"`,
		},
		{
			name:      "ERROR: service error returns 500",
			userClaim: validUserClaims,
			body:      `{"clientReferenceId":"ref-123","paymentSessionId":"session-123","isFullAmount":true,"reason":"DUPLICATE","method":"AUTO"}`,
			setupMock: func() {
				mockRefundSvc.On("Create", mock.Anything, mock.Anything).
					Return(nil, pkgError.New(response.HttpErrDatabase, errors.New("database error"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			wantRespBody:   `{"code":"98","message":"database error","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "SUCCESS: creates refund with auto method",
			userClaim: validUserClaims,
			body:      `{"clientReferenceId":"ref-123","paymentSessionId":"session-123","isFullAmount":false,"amount":{"value":"50000","currency":"IDR"},"reason":"REQUESTED_BY_CUSTOMER","method":"AUTO"}`,
			setupMock: func() {
				mockRefundSvc.On("Create", mock.Anything, mock.Anything).
					Return(&refundModel.RefundResponse{
						ID:                "refund-id-123",
						ClientReferenceID: "ref-123",
						PaymentSessionID:  "session-123",
						ChargeID:          "charge-001",
						Amount: commonModel.Amount{
							Value:    "50000",
							Currency: "IDR",
						},
						CapturedAmount: commonModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						IsFullAmount: false,
						Status:       constant.RefundStatusPending,
						Reason:       constant.RefundReasonRequestedByCustomer,
						Method:       constant.RefundMethodAuto,
						CreatedAt:    fixedTime,
						UpdatedAt:    fixedTime,
					}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"refund-id-123","clientReferenceId":"ref-123","paymentSessionId":"session-123","chargeId":"charge-001","capturedAmount":{"value":"100000","currency":"IDR"},"isFullAmount":false,"amount":{"value":"50000","currency":"IDR"},"status":"PENDING","reason":"REQUESTED_BY_CUSTOMER","description":"","method":"AUTO","createdAt":"2024-01-15T10:30:00Z","updatedAt":"2024-01-15T10:30:00Z"}}`,
		},
		{
			name:      "SUCCESS: creates refund with transfer only method",
			userClaim: validUserClaims,
			body: `{
				"clientReferenceId":"ref-456",
				"paymentSessionId":"session-456",
				"isFullAmount":true,
				"reason":"CANCELLATION",
				"method":"TRANSFER_ONLY",
				"transferDestination":{
					"channelCode":"BCA",
					"channelInformation":{"accountNumber":"1234567890","accountName":"JOHN DOE"}
				}
			}`,
			setupMock: func() {
				mockRefundSvc.On("Create", mock.Anything, mock.Anything).
					Return(&refundModel.RefundResponse{
						ID:                "refund-id-456",
						ClientReferenceID: "ref-456",
						PaymentSessionID:  "session-456",
						ChargeID:          "charge-002",
						Amount: commonModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						CapturedAmount: commonModel.Amount{
							Value:    "100000",
							Currency: "IDR",
						},
						IsFullAmount:    true,
						Status:          constant.RefundStatusPending,
						Reason:          constant.RefundReasonCancellation,
						Method:          constant.RefundMethodTransferOnly,
						DestinationType: "ACCOUNT",
						TransferDestination: &refundModel.TransferDestination{
							ChannelCode: "BCA",
							ChannelInformation: refundModel.ChannelInformation{
								AccountNumber: "1234567890",
								AccountName:   "JOHN DOE",
							},
						},
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"id":"refund-id-456","clientReferenceId":"ref-456","paymentSessionId":"session-456","chargeId":"charge-002","capturedAmount":{"value":"100000","currency":"IDR"},"isFullAmount":true,"amount":{"value":"100000","currency":"IDR"},"status":"PENDING","reason":"CANCELLATION","description":"","destinationType":"ACCOUNT","method":"TRANSFER_ONLY","transferDestination":{"channelCode":"BCA","channelInformation":{"accountNumber":"1234567890","accountName":"JOHN DOE"},"description":""},"createdAt":"2024-01-15T10:30:00Z","updatedAt":"2024-01-15T10:30:00Z"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			req := httptest.NewRequest(http.MethodPost, "/api/v1/refunds", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			if tc.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaim))
			}

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(controller.Create)
			handler.ServeHTTP(rr, req)

			t.Logf("Response body: %s", rr.Body.String())

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.containsCheck != "" {
				assert.Contains(t, rr.Body.String(), tc.containsCheck)
			} else if tc.wantRespBody != "" {
				if !assert.JSONEq(t, tc.wantRespBody, rr.Body.String()) {
					t.Log("Expected:", tc.wantRespBody)
					t.Log("Actual:", rr.Body.String())
				}
			}
		})
	}
}
