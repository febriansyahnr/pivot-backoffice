package fds

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckTransaction(t *testing.T) {
	validUUID := uuid.New().String()
	invalidUUID := "invalid-uuid"

	testCases := []struct {
		desc           string
		setupMocks     func(mockSvc *mocks.IFdsService)
		id             string
		expectedStatus int
		timeout        *int64
	}{
		{
			desc: "success",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On(
					"CheckTransaction",
					mock.Anything,
					validUUID,
					mock.AnythingOfType("*fdscommon.CheckTransactionRequest"),
				).Return(&fdscommon.CheckTransactionResponse{
					Status: constant.FDS_STATUS_PASSED,
				}, nil)
			},
			id:             validUUID,
			expectedStatus: http.StatusOK,
		},
		{
			desc:           "invalid UUID",
			setupMocks:     func(mockSvc *mocks.IFdsService) {},
			id:             invalidUUID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			desc: "service error",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On(
					"CheckTransaction",
					mock.Anything,
					validUUID,
					mock.AnythingOfType("*fdscommon.CheckTransactionRequest"),
				).Return(nil, errors.New("service error"))
			},
			id:             validUUID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			desc: "timeout case",
			setupMocks: func(mockSvc *mocks.IFdsService) {
				mockSvc.On(
					"CheckTransaction",
					mock.MatchedBy(func(ctx context.Context) bool {
						// Simulate timeout by causing delay
						time.Sleep(10 * time.Millisecond)
						return true
					}),
					validUUID,
					mock.AnythingOfType("*fdscommon.CheckTransactionRequest"),
				).Return(&fdscommon.CheckTransactionResponse{
					Status: constant.FDS_STATUS_NOT_EVALUATED,
				}, nil)
			},
			id:             validUUID,
			expectedStatus: http.StatusOK,
			timeout:        func(t int64) *int64 { return &t }(5), // 5ms timeout
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Create mock service
			timeout := int64(5000)
			if tc.timeout != nil {
				timeout = *tc.timeout
			}
			config := config.Config{
				FdsConfig: config.FdsConfig{
					Timeout: timeout,
				},
				Environment: constant.EnvironmentStaging,
			}
			mockSvc := mocks.NewIFdsService(t)
			mockValidator := validator.New()
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			// Setup expectations
			tc.setupMocks(mockSvc)

			// Create controller with mock service
			controller := New(&config, mockLogger, mockValidator, mockSvc)

			// Create request with chi context for URL params
			req, err := http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				"/internal/v1/fds/check/"+tc.id,
				nil,
			)
			req.Header.Set(constant.HeaderXSimulationCardNumber, "123456") // NOSONAR
			assert.NoError(t, err)

			// Create test response recorder
			rr := httptest.NewRecorder()

			// Setup chi router for this test
			testRouter := chi.NewRouter()
			testRouter.Post("/internal/v1/fds/check/{id}", controller.CheckTransaction)

			// Serve the request
			testRouter.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Verify that all expectations were met
			mockSvc.AssertExpectations(t)
		})
	}
}
