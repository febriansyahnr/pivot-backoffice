package cardFundedPayoutController

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutDetail(t *testing.T) {
	vld := validatorExt.New()
	service := serviceMocks.NewICardFundedPayoutService(t)

	payoutID := "6e34cfca-de00-4aa1-bc6b-069d878012c8"
	validUserClaims := &userModel.UserTokenClaims{
		UUID:       uuid.NewString(),
		Name:       "John Doe", // NOSONAR
		MerchantId: uuid.NewString(),
	}

	cfg := &config.Config{ServiceName: "testing"}
	handler := New(cfg, vld, service)

	router := chi.NewRouter()
	router.Get("/{payoutId}", handler.GetPayoutDetail)

	testCases := []struct {
		name             string
		payoutID         string
		userClaim        *userModel.UserTokenClaims
		setupMock        func()
		wantStatusCode   int
		wantResponseBody string
	}{
		{
			name:      "SUCCESS",
			payoutID:  payoutID,
			userClaim: validUserClaims,
			setupMock: func() {
				createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
				approvalDate := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
				statusTimestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

				service.On(
					"GetPayoutDetail", mock.Anything, mock.Anything,
				).Once().Return(&model.GetPayoutDetailResponse{
					UUID:              payoutID,
					CreatedAt:         createdAt,
					ReferenceID:       "REF-123",
					Amount:            "1500000",
					Fee:               "15000",
					TotalAmount:       "1515000",
					TransactionStatus: "SUCCESS",
					ApprovalStatus:    "APPROVED",
					VendorID:          "vendor-001",
					VendorName:        "PT Sample Vendor",
					Remarks:           "Monthly payment",
					BankName:          "Bank Central Asia",
					AccountNumber:     "1234567890",
					AccountName:       "PT Sample Vendor",
					Card: model.CardInfo{
						LastFour: "4242",
						Brand:    "Visa",
						Channel:  "Credit",
						Name:     "Business Card",
						Issuer:   "BCA",
						Expiry:   "12/25",
					},
					ChargeIDs:     []string{"CHG-001", "CHG-002"},
					ApprovalDate:  &approvalDate,
					ApprovedBy:    util.ValueToPtr("admin@example.com"),
					CurrentStatus: "SUCCESS",
					StatusHistory: []model.StatusHistoryItem{
						{
							Status:      "pending",
							Label:       "Pending",
							Description: "Payout request created",
							Timestamp:   &statusTimestamp,
						},
						{
							Status:      "success",
							Label:       "Success",
							Description: "Payout completed successfully",
							Timestamp:   &approvalDate,
						},
					},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"uuid":"6e34cfca-de00-4aa1-bc6b-069d878012c8","createdAt":"2024-01-15T10:30:00Z","referenceId":"REF-123","amount":"1500000","fee":"15000","totalAmount":"1515000","transactionStatus":"SUCCESS","approvalStatus":"APPROVED","vendorId":"vendor-001","vendorName":"PT Sample Vendor","remarks":"Monthly payment","bankName":"Bank Central Asia","accountNumber":"1234567890","accountName":"PT Sample Vendor","card":{"lastFour":"4242","brand":"Visa","type":"Credit","name":"Business Card","issuer":"BCA","expiry":"12/25"},"chargeIds":["CHG-001","CHG-002"],"approvalDate":"2024-01-15T11:00:00Z","approvedBy":"admin@example.com","currentStatus":"SUCCESS","statusHistory":[{"status":"pending","label":"Pending","description":"Payout request created","timestamp":"2024-01-15T10:30:00Z"},{"status":"success","label":"Success","description":"Payout completed successfully","timestamp":"2024-01-15T11:00:00Z"}]}}`,
		},
		{
			name:             "ERROR: User not in Context",
			payoutID:         payoutID,
			userClaim:        nil,
			wantStatusCode:   http.StatusUnauthorized,
			wantResponseBody: `{"code":"41","message":"user not found","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:             "ERROR: Invalid payoutId format",
			payoutID:         "invalid-uuid",
			userClaim:        validUserClaims,
			wantStatusCode:   http.StatusBadRequest,
			wantResponseBody: `{"code":"40","message":"payoutId is required and must be a valid UUID","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "ERROR: Service error",
			payoutID:  payoutID,
			userClaim: validUserClaims,
			setupMock: func() {
				service.On("GetPayoutDetail", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantStatusCode:   http.StatusInternalServerError,
			wantResponseBody: `{"code":"99","message":"assert.AnError general error for testing","error":{"type":"UNKNOWN","details":[],"traceId":""},"data":null}`,
		},
		{
			name:      "SUCCESS: Minimal response",
			payoutID:  payoutID,
			userClaim: validUserClaims,
			setupMock: func() {
				createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
				service.On(
					"GetPayoutDetail", mock.Anything, mock.Anything,
				).Once().Return(&model.GetPayoutDetailResponse{
					UUID:              payoutID,
					CreatedAt:         createdAt,
					ReferenceID:       "REF-123",
					Amount:            "1500000",
					Fee:               "15000",
					TotalAmount:       "1515000",
					TransactionStatus: "PROCESSING",
					ApprovalStatus:    "WAITING",
					VendorID:          "vendor-001",
					VendorName:        "PT Sample Vendor",
					Remarks:           "Monthly payment",
					BankName:          "Bank Central Asia",
					AccountNumber:     "1234567890",
					AccountName:       "PT Sample Vendor",
					Card: model.CardInfo{
						LastFour: "4242",
						Brand:    "Visa",
						Channel:  "Credit",
						Name:     "Business Card",
						Issuer:   "BCA",
						Expiry:   "12/25",
					},
					ChargeIDs:     []string{},
					CurrentStatus: "PROCESSING",
					StatusHistory: []model.StatusHistoryItem{},
				}, nil)
			},
			wantStatusCode:   http.StatusOK,
			wantResponseBody: `{"code":"00","message":"OK","data":{"uuid":"6e34cfca-de00-4aa1-bc6b-069d878012c8","createdAt":"2024-01-15T10:30:00Z","referenceId":"REF-123","amount":"1500000","fee":"15000","totalAmount":"1515000","transactionStatus":"PROCESSING","approvalStatus":"WAITING","vendorId":"vendor-001","vendorName":"PT Sample Vendor","remarks":"Monthly payment","bankName":"Bank Central Asia","accountNumber":"1234567890","accountName":"PT Sample Vendor","card":{"lastFour":"4242","brand":"Visa","type":"Credit","name":"Business Card","issuer":"BCA","expiry":"12/25"},"chargeIds":[],"currentStatus":"PROCESSING","statusHistory":[]}}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/"+tt.payoutID, nil)

			// Set chi route context for path params
			chiCtx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

			if tt.userClaim != nil {
				req = req.WithContext(context.WithValue(req.Context(), constant.CtxUserInfoKey, tt.userClaim))
			}

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatusCode, rec.Result().StatusCode)
			if tt.wantStatusCode == http.StatusNotFound {
				assert.Equal(t, tt.wantResponseBody, rec.Body.String())
			} else {
				if !assert.JSONEq(t, tt.wantResponseBody, rec.Body.String()) {
					t.Log("Actual:", rec.Body.String())
				}
			}

			service.AssertExpectations(t)
		})
	}
}
