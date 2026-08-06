package disbursementController_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	disbursementController "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestGetDisbursementInsight(t *testing.T) {
	validMerchantID := uuid.NewString()
	validUserClaims := &userModel.UserTokenClaims{
		MerchantId: validMerchantID,
	}

	expectedResponse := &disbursementModel.DisbursementInsightResponse{
		WaitingForApproval: disbursementModel.SummaryTransaction{
			Count: 5,
			Sum:   commonModel.Amount{Currency: "IDR", Value: "50000.00"},
		},
		Delayed: disbursementModel.SummaryTransaction{
			Count: 2,
			Sum:   commonModel.Amount{Currency: "IDR", Value: "20000.00"},
		},
		AllStatus: disbursementModel.AllStatusSummary{
			Success: disbursementModel.SummaryTransaction{
				Count: 100,
				Sum:   commonModel.Amount{Currency: "IDR", Value: "1000000.00"},
			},
			Pending: disbursementModel.SummaryTransaction{
				Count: 10,
				Sum:   commonModel.Amount{Currency: "IDR", Value: "100000.00"},
			},
			Failed: disbursementModel.SummaryTransaction{
				Count: 3,
				Sum:   commonModel.Amount{Currency: "IDR", Value: "30000.00"},
			},
		},
		FailureReasons: []disbursementModel.SummaryTransactionByReason{
			{
				ReasonType: "INSUFFICIENT_BALANCE",
				Count:      2,
				Sum:        commonModel.Amount{Currency: "IDR", Value: "20000.00"},
			},
		},
	}

	testCases := []struct {
		name           string
		setupMocks     func(*serviceMocks.IDisbursementService)
		setupRequest   func() *http.Request
		userClaims     *userModel.UserTokenClaims
		expectedStatus int
		expectedBody   func(*testing.T, []byte)
	}{
		{
			name: "SUCCESS - Get insights with default dates",
			setupMocks: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On("GetDisbursementInsight", 
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(filter disbursementModel.GetDisbursementInsightFilter) bool {
						return filter.MerchantID == validMerchantID
					}),
				).Return(expectedResponse, nil)
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/disbursement-insights", nil)
				return req
			},
			userClaims:     validUserClaims,
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body []byte) {
				var response response.ApiResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "00", response.Code)
				assert.Equal(t, "OK", response.Message)
				
				// Parse the data field
				dataBytes, _ := json.Marshal(response.Data)
				var insightResponse disbursementModel.DisbursementInsightResponse
				err = json.Unmarshal(dataBytes, &insightResponse)
				assert.NoError(t, err)
				
				assert.Equal(t, expectedResponse.WaitingForApproval.Count, insightResponse.WaitingForApproval.Count)
				assert.Equal(t, expectedResponse.AllStatus.Success.Count, insightResponse.AllStatus.Success.Count)
				assert.Len(t, insightResponse.FailureReasons, 1)
			},
		},
		{
			name: "SUCCESS - Get insights with custom date range",
			setupMocks: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On("GetDisbursementInsight",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.MatchedBy(func(filter disbursementModel.GetDisbursementInsightFilter) bool {
						return filter.MerchantID == validMerchantID &&
							filter.InsightStartDate.Format("2006-01-02") == "2024-01-01" &&
							filter.InsightEndDate.Format("2006-01-02") == "2024-01-31"
					}),
				).Return(expectedResponse, nil)
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/disbursement-insights?insightStartDate=2024-01-01T00:00:00Z&insightEndDate=2024-01-31T23:59:59Z", nil)
				return req
			},
			userClaims:     validUserClaims,
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body []byte) {
				var response response.ApiResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "00", response.Code)
			},
		},
		{
			name: "ERROR - Unauthorized (no user context)",
			setupMocks: func(mockService *serviceMocks.IDisbursementService) {
				// No mock setup needed as it should fail before service call
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/disbursement-insights", nil)
				return req
			},
			userClaims:     nil, // No user claims in context
			expectedStatus: http.StatusUnauthorized,
			expectedBody: func(t *testing.T, body []byte) {
				var response response.ApiResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.NotEqual(t, "00", response.Code)
				assert.Contains(t, response.Message, "user not found")
			},
		},
		{
			name: "ERROR - Service returns error",
			setupMocks: func(mockService *serviceMocks.IDisbursementService) {
				mockService.On("GetDisbursementInsight",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("disbursementModel.GetDisbursementInsightFilter"),
				).Return(nil, assert.AnError)
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/disbursement-insights", nil)
				return req
			},
			userClaims:     validUserClaims,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body []byte) {
				var response response.ApiResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.NotEqual(t, "00", response.Code)
			},
		},
		{
			name: "SUCCESS - Empty response from service",
			setupMocks: func(mockService *serviceMocks.IDisbursementService) {
				emptyResponse := &disbursementModel.DisbursementInsightResponse{
					WaitingForApproval: disbursementModel.SummaryTransaction{
						Count: 0,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
					},
					Delayed: disbursementModel.SummaryTransaction{
						Count: 0,
						Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
					},
					AllStatus: disbursementModel.AllStatusSummary{
						Success: disbursementModel.SummaryTransaction{
							Count: 0,
							Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
						},
						Pending: disbursementModel.SummaryTransaction{
							Count: 0,
							Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
						},
						Failed: disbursementModel.SummaryTransaction{
							Count: 0,
							Sum:   commonModel.Amount{Currency: "IDR", Value: "0.00"},
						},
					},
					FailureReasons: []disbursementModel.SummaryTransactionByReason{},
				}
				mockService.On("GetDisbursementInsight",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("disbursementModel.GetDisbursementInsightFilter"),
				).Return(emptyResponse, nil)
			},
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/disbursement-insights", nil)
				return req
			},
			userClaims:     validUserClaims,
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body []byte) {
				var response response.ApiResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "00", response.Code)
				
				// Parse the data field
				dataBytes, _ := json.Marshal(response.Data)
				var insightResponse disbursementModel.DisbursementInsightResponse
				err = json.Unmarshal(dataBytes, &insightResponse)
				assert.NoError(t, err)
				
				assert.Equal(t, 0, insightResponse.WaitingForApproval.Count)
				assert.Equal(t, 0, insightResponse.AllStatus.Success.Count)
				assert.Len(t, insightResponse.FailureReasons, 0)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockDisbursementService := serviceMocks.NewIDisbursementService(t)
			tc.setupMocks(mockDisbursementService)

			// Setup controller
			config := &config.Config{}
			validator := validator.New()
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			services := disbursementController.Services{
				DisbursementSvc: mockDisbursementService,
			}

			controller := disbursementController.New(
				config,
				validator,
				nil, // monitor
				services,
				nil, // rabbitMqExt
				nil, // gcs
				disbursementController.WithLogger(logger),
			)

			// Setup request
			req := tc.setupRequest()
			
			// Add user context if provided
			if tc.userClaims != nil {
				ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, tc.userClaims)
				req = req.WithContext(ctx)
			}

			// Setup response recorder
			rr := httptest.NewRecorder()

			// Execute
			controller.GetDisbursementInsight(rr, req)

			// Assert
			assert.Equal(t, tc.expectedStatus, rr.Code)
			
			if tc.expectedBody != nil {
				tc.expectedBody(t, rr.Body.Bytes())
			}

			// Verify mocks
			mockDisbursementService.AssertExpectations(t)
		})
	}
}

func TestGetDisbursementInsight_QueryParameterParsing(t *testing.T) {
	validMerchantID := uuid.NewString()
	validUserClaims := &userModel.UserTokenClaims{
		MerchantId: validMerchantID,
	}

	testCases := []struct {
		name        string
		queryParams string
		expectError bool
		validateFilter func(t *testing.T, filter disbursementModel.GetDisbursementInsightFilter)
	}{
		{
			name:        "Valid RFC3339 date formats",
			queryParams: "?insightStartDate=2024-01-01T00:00:00Z&insightEndDate=2024-01-31T23:59:59Z",
			expectError: false,
			validateFilter: func(t *testing.T, filter disbursementModel.GetDisbursementInsightFilter) {
				assert.Equal(t, "2024-01-01", filter.InsightStartDate.Format("2006-01-02"))
				assert.Equal(t, "2024-01-31", filter.InsightEndDate.Format("2006-01-02"))
			},
		},
		{
			name:        "Only start date provided",
			queryParams: "?insightStartDate=2024-06-01T10:30:00Z",
			expectError: false,
			validateFilter: func(t *testing.T, filter disbursementModel.GetDisbursementInsightFilter) {
				assert.Equal(t, "2024-06-01", filter.InsightStartDate.Format("2006-01-02"))
				// End date should be default (today + 1 day)
				assert.True(t, filter.InsightEndDate.After(time.Now()))
			},
		},
		{
			name:        "Only end date provided", 
			queryParams: "?insightEndDate=2024-12-31T23:59:59Z",
			expectError: false,
			validateFilter: func(t *testing.T, filter disbursementModel.GetDisbursementInsightFilter) {
				assert.Equal(t, "2024-12-31", filter.InsightEndDate.Format("2006-01-02"))
				// Start date should be default (today) - allow some flexibility for timezone differences
				expectedDate := time.Now().UTC().Format("2006-01-02")
				actualDate := filter.InsightStartDate.Format("2006-01-02")
				// Allow current date or yesterday due to timezone differences
				assert.True(t, actualDate == expectedDate || actualDate == time.Now().AddDate(0, 0, -1).UTC().Format("2006-01-02"))
			},
		},
		{
			name:        "Invalid date format - should use defaults",
			queryParams: "?insightStartDate=invalid-date&insightEndDate=also-invalid",
			expectError: false,
			validateFilter: func(t *testing.T, filter disbursementModel.GetDisbursementInsightFilter) {
				// Should fallback to default dates when parsing fails
				expectedDate := time.Now().UTC().Format("2006-01-02")
				actualDate := filter.InsightStartDate.Format("2006-01-02")
				// Allow current date or yesterday due to timezone differences  
				assert.True(t, actualDate == expectedDate || actualDate == time.Now().AddDate(0, 0, -1).UTC().Format("2006-01-02"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			mockService := serviceMocks.NewIDisbursementService(t)
			
			var capturedFilter disbursementModel.GetDisbursementInsightFilter
			mockService.On("GetDisbursementInsight",
				mock.AnythingOfType(constant.MockTypeValueContextReference),
				mock.AnythingOfType("disbursementModel.GetDisbursementInsightFilter"),
			).Run(func(args mock.Arguments) {
				capturedFilter = args.Get(1).(disbursementModel.GetDisbursementInsightFilter)
			}).Return(&disbursementModel.DisbursementInsightResponse{}, nil)

			// Setup controller
			config := &config.Config{}
			validator := validator.New()
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			services := disbursementController.Services{
				DisbursementSvc: mockService,
			}

			controller := disbursementController.New(
				config,
				validator,
				nil,
				services,
				nil,
				nil,
				disbursementController.WithLogger(logger),
			)

			// Setup request
			url := "/api/v1/disbursement-insights" + tc.queryParams
			req := httptest.NewRequest(http.MethodGet, url, nil)
			ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, validUserClaims)
			req = req.WithContext(ctx)

			// Execute
			rr := httptest.NewRecorder()
			controller.GetDisbursementInsight(rr, req)

			// Validate the filter that was passed to the service
			if tc.validateFilter != nil {
				tc.validateFilter(t, capturedFilter)
			}

			// Verify basic success
			if !tc.expectError {
				assert.Equal(t, http.StatusOK, rr.Code)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetDisbursementInsight_Integration(t *testing.T) {
	// This test verifies the complete flow from HTTP request to service call
	merchantID := uuid.NewString()
	userClaims := &userModel.UserTokenClaims{
		MerchantId: merchantID,
	}

	expectedResponse := &disbursementModel.DisbursementInsightResponse{
		WaitingForApproval: disbursementModel.SummaryTransaction{
			Count: 15,
			Sum:   commonModel.Amount{Currency: "IDR", Value: "150000.00"},
		},
		Delayed: disbursementModel.SummaryTransaction{
			Count: 3,
			Sum:   commonModel.Amount{Currency: "IDR", Value: "30000.00"},
		},
		AllStatus: disbursementModel.AllStatusSummary{
			Success: disbursementModel.SummaryTransaction{
				Count: 500,
				Sum:   commonModel.Amount{Currency: "IDR", Value: "5000000.00"},
			},
			Pending: disbursementModel.SummaryTransaction{
				Count: 25,
				Sum:   commonModel.Amount{Currency: "IDR", Value: "250000.00"},
			},
			Failed: disbursementModel.SummaryTransaction{
				Count: 8,
				Sum:   commonModel.Amount{Currency: "IDR", Value: "80000.00"},
			},
		},
		FailureReasons: []disbursementModel.SummaryTransactionByReason{
			{
				ReasonType: "INSUFFICIENT_BALANCE",
				Count:      5,
				Sum:        commonModel.Amount{Currency: "IDR", Value: "50000.00"},
			},
			{
				ReasonType: "INVALID_ACCOUNT",
				Count:      3,
				Sum:        commonModel.Amount{Currency: "IDR", Value: "30000.00"},
			},
		},
	}

	// Setup mock
	mockService := serviceMocks.NewIDisbursementService(t)
	mockService.On("GetDisbursementInsight",
		mock.AnythingOfType(constant.MockTypeValueContextReference),
		mock.MatchedBy(func(filter disbursementModel.GetDisbursementInsightFilter) bool {
			return filter.MerchantID == merchantID
		}),
	).Return(expectedResponse, nil)

	// Setup controller
	config := &config.Config{}
	validator := validator.New()
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	services := disbursementController.Services{
		DisbursementSvc: mockService,
	}

	controller := disbursementController.New(
		config,
		validator,
		nil,
		services,
		nil,
		nil,
		disbursementController.WithLogger(logger),
	)

	// Setup request with custom date range
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/disbursement-insights?insightStartDate=2024-03-01T00:00:00Z&insightEndDate=2024-03-31T23:59:59Z",
		nil,
	)
	ctx := context.WithValue(req.Context(), constant.CtxUserInfoKey, userClaims)
	req = req.WithContext(ctx)

	// Execute
	rr := httptest.NewRecorder()
	controller.GetDisbursementInsight(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	// Parse and validate response body
	var apiResponse response.ApiResponse
	err := json.Unmarshal(rr.Body.Bytes(), &apiResponse)
	assert.NoError(t, err)
	assert.Equal(t, "00", apiResponse.Code)
	assert.Equal(t, "OK", apiResponse.Message)

	// Parse the data field
	dataBytes, _ := json.Marshal(apiResponse.Data)
	var insightResponse disbursementModel.DisbursementInsightResponse
	err = json.Unmarshal(dataBytes, &insightResponse)
	assert.NoError(t, err)

	// Validate all fields
	assert.Equal(t, expectedResponse.WaitingForApproval.Count, insightResponse.WaitingForApproval.Count)
	assert.Equal(t, expectedResponse.WaitingForApproval.Sum.Value, insightResponse.WaitingForApproval.Sum.Value)
	
	assert.Equal(t, expectedResponse.Delayed.Count, insightResponse.Delayed.Count)
	assert.Equal(t, expectedResponse.Delayed.Sum.Value, insightResponse.Delayed.Sum.Value)

	assert.Equal(t, expectedResponse.AllStatus.Success.Count, insightResponse.AllStatus.Success.Count)
	assert.Equal(t, expectedResponse.AllStatus.Pending.Count, insightResponse.AllStatus.Pending.Count)
	assert.Equal(t, expectedResponse.AllStatus.Failed.Count, insightResponse.AllStatus.Failed.Count)

	assert.Len(t, insightResponse.FailureReasons, 2)
	assert.Equal(t, "INSUFFICIENT_BALANCE", insightResponse.FailureReasons[0].ReasonType)
	assert.Equal(t, "INVALID_ACCOUNT", insightResponse.FailureReasons[1].ReasonType)

	mockService.AssertExpectations(t)
}