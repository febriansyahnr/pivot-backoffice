package crmDisbursementController

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	mocks_service "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
)

func TestHandler_GetPayoutStatusAndRouting(t *testing.T) {
	type Mockers struct {
		disbursementSvc *mocks_service.IDisbursementService
	}

	mockResponse := &disbursementModel.CRMPayoutStatusResponse{
		Code: "00",
		Data: &disbursementModel.CRMPayoutStatusResponseData{
			DisbursementUUID:   "disbursement-uuid-123",
			ReferenceID:        "123e4567-e89b-12d3-a456-426614174000",
			Status:             "SUCCESS",
			ApprovalStatus:     "APPROVED",
			Amount:             "100000.00",
			BeneficiaryAccount: "1234567890",
			BeneficiaryName:    "John Doe",
			BeneficiaryBank:    "Bank ABC",
			TransactionDate:    "2024-01-01T10:00:00Z",
			CreatedAt:          "2024-01-01T10:00:00Z",
			UpdatedAt:          "2024-01-01T10:05:00Z",
			TransferLogs: []disbursementModel.RoutingHistoryItem{
				{
					Order:       1,
					BankName:    "Bank ABC",
					Status:      "FAILED",
					ResponseMsg: "Insufficient funds",
					Timestamp:   "2024-01-01T10:00:00Z",
				},
				{
					Order:       2,
					BankName:    "Bank DEF",
					Status:      "SUCCESS",
					ResponseMsg: "",
					Timestamp:   "2024-01-01T10:01:00Z",
				},
			},
		},
	}

	testCases := []struct {
		desc           string
		requestBody    interface{}
		expectedStatus int
		wantError      bool
		setupMock      func(mockers Mockers)
	}{
		{
			desc: "success - valid request",
			requestBody: disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: "123e4567-e89b-12d3-a456-426614174000",
			},
			expectedStatus: http.StatusOK,
			wantError:      false,
			setupMock: func(mockers Mockers) {
				mockers.disbursementSvc.On(
					"GetPayoutStatusAndRouting",
					mock.Anything,
					mock.AnythingOfType("*disbursementModel.CRMSinglePayoutStatusRequest"),
				).Return(mockResponse, nil)
			},
		},
		{
			desc: "error - validation failed, empty reference ID",
			requestBody: disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: "",
			},
			expectedStatus: http.StatusBadRequest,
			wantError:      true,
			setupMock: func(mockers Mockers) {
				// No service call expected due to validation failure
			},
		},
		{
			desc:           "error - invalid JSON request body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			wantError:      true,
			setupMock: func(mockers Mockers) {
				// No service call expected due to JSON decode failure
			},
		},
		{
			desc: "error - service returns transaction not found",
			requestBody: disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: "00000000-0000-0000-0000-000000000000",
			},
			expectedStatus: http.StatusNotFound,
			wantError:      true,
			setupMock: func(mockers Mockers) {
				mockers.disbursementSvc.On(
					"GetPayoutStatusAndRouting",
					mock.Anything,
					mock.AnythingOfType("*disbursementModel.CRMSinglePayoutStatusRequest"),
				).Return(nil, pkgErrs.New(httpResponse.HttpErrNotFound, errors.New("transaction not found")))
			},
		},
		{
			desc: "error - service returns database error",
			requestBody: disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: "123e4567-e89b-12d3-a456-426614174000",
			},
			expectedStatus: http.StatusInternalServerError,
			wantError:      true,
			setupMock: func(mockers Mockers) {
				mockers.disbursementSvc.On(
					"GetPayoutStatusAndRouting",
					mock.Anything,
					mock.AnythingOfType("*disbursementModel.CRMSinglePayoutStatusRequest"),
				).Return(nil, pkgErrs.New(httpResponse.HttpErrDatabase, errors.New("database error")))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockers := Mockers{
				disbursementSvc: new(mocks_service.IDisbursementService),
			}
			tc.setupMock(mockers)

			handler := &handler{
				validator:       validatorExt.New(),
				disbursementSvc: mockers.disbursementSvc,
			}

			var reqBody []byte
			var err error
			if tc.requestBody != "invalid json" {
				reqBody, err = json.Marshal(tc.requestBody)
				assert.NoError(t, err)
			} else {
				reqBody = []byte(tc.requestBody.(string))
			}

			req := httptest.NewRequest(http.MethodPost, "/crm/v1/disbursements/payout-status", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.Background())

			rr := httptest.NewRecorder()

			handler.GetPayoutStatusAndRouting(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tc.wantError {
				// For error cases, the controller should return error response
				assert.NotEqual(t, http.StatusOK, rr.Code)
			} else {
				// For success cases, expect ApiResponse structure
				assert.Equal(t, "00", response["code"])
				assert.Equal(t, "OK", response["message"])
				assert.NotNil(t, response["data"])
				
				// The service response is nested inside the API response data
				serviceResponse := response["data"].(map[string]interface{})
				assert.Equal(t, "00", serviceResponse["code"])
				assert.NotNil(t, serviceResponse["data"])
				
				// Verify actual response data structure
				responseData := serviceResponse["data"].(map[string]interface{})
				assert.Equal(t, mockResponse.Data.DisbursementUUID, responseData["disbursementUuid"])
				assert.Equal(t, mockResponse.Data.ReferenceID, responseData["referenceId"])
				assert.Equal(t, mockResponse.Data.Status, responseData["status"])
				assert.NotNil(t, responseData["transferLogs"])
			}

			mockers.disbursementSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetPayoutStatusAndRouting_ValidateRequestStructure(t *testing.T) {
	handler := &handler{
		validator: validatorExt.New(),
	}

	testCases := []struct {
		desc      string
		request   disbursementModel.CRMSinglePayoutStatusRequest
		wantError bool
	}{
		{
			desc: "valid request with single reference ID",
			request: disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: "43ce9f2e-4b46-4bff-9431-9be9b11cf7c2",
			},
			wantError: false,
		},
		{
			desc: "invalid request - empty reference ID",
			request: disbursementModel.CRMSinglePayoutStatusRequest{
				ReferenceID: "",
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := handler.validator.Struct(tc.request)
			
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_GetPayoutStatusAndRouting_ResponseFormat(t *testing.T) {
	mockResponse := &disbursementModel.CRMPayoutStatusResponse{
		Code: "00",
		Data: &disbursementModel.CRMPayoutStatusResponseData{
			DisbursementUUID:   "disbursement-uuid-123",
			ReferenceID:        "123e4567-e89b-12d3-a456-426614174000",
			Status:             "SUCCESS",
			ApprovalStatus:     "APPROVED",
			Amount:             "150000.00",
			BeneficiaryAccount: "9876543210",
			BeneficiaryName:    "Jane Smith",
			BeneficiaryBank:    "Bank XYZ",
			TransactionDate:    "2024-01-01T09:30:00Z",
			CreatedAt:          "2024-01-01T09:30:00Z",
			UpdatedAt:          "2024-01-01T09:35:00Z",
			TransferLogs: []disbursementModel.RoutingHistoryItem{
				{
					Order:       1,
					BankName:    "Bank XYZ",
					Status:      "SUCCESS",
					ResponseMsg: "Transaction completed successfully",
					Timestamp:   "2024-01-01T09:35:00Z",
				},
			},
		},
	}

	disbursementSvc := new(mocks_service.IDisbursementService)
	disbursementSvc.On(
		"GetPayoutStatusAndRouting",
		mock.Anything,
		mock.AnythingOfType("*disbursementModel.CRMSinglePayoutStatusRequest"),
	).Return(mockResponse, nil)

	handler := &handler{
		validator:       validatorExt.New(),
		disbursementSvc: disbursementSvc,
	}

	requestBody := disbursementModel.CRMSinglePayoutStatusRequest{
		ReferenceID: "123e4567-e89b-12d3-a456-426614174000",
	}

	reqBodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/crm/v1/disbursements/payout-status", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.Background())

	rr := httptest.NewRecorder()

	handler.GetPayoutStatusAndRouting(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Expect ApiResponse structure
	assert.Equal(t, "00", response["code"])
	assert.Equal(t, "OK", response["message"])
	assert.NotNil(t, response["data"])
	
	// The service response is nested inside the API response data
	serviceResponse := response["data"].(map[string]interface{})
	assert.Equal(t, "00", serviceResponse["code"])
	assert.NotNil(t, serviceResponse["data"])
	
	// Verify actual response data structure
	responseData := serviceResponse["data"].(map[string]interface{})
	assert.Equal(t, mockResponse.Data.DisbursementUUID, responseData["disbursementUuid"])
	assert.Equal(t, mockResponse.Data.ReferenceID, responseData["referenceId"])
	assert.Equal(t, mockResponse.Data.Status, responseData["status"])
	assert.Equal(t, mockResponse.Data.Amount, responseData["amount"])
	assert.Equal(t, mockResponse.Data.BeneficiaryAccount, responseData["beneficiaryAccount"])
	assert.Equal(t, mockResponse.Data.BeneficiaryName, responseData["beneficiaryName"])
	assert.Equal(t, mockResponse.Data.BeneficiaryBank, responseData["beneficiaryBank"])
	assert.NotEmpty(t, responseData["transferLogs"])
	
	transferLogs := responseData["transferLogs"].([]interface{})
	assert.Len(t, transferLogs, 1)
	
	firstLog := transferLogs[0].(map[string]interface{})
	assert.Equal(t, float64(1), firstLog["order"])
	assert.Equal(t, "Bank XYZ", firstLog["bankName"])
	assert.Equal(t, "SUCCESS", firstLog["status"])
	assert.Equal(t, "Transaction completed successfully", firstLog["responseMessage"])

	disbursementSvc.AssertExpectations(t)
}

func TestHandler_GetBatchPayoutStatusAndRouting(t *testing.T) {
	type Mockers struct {
		disbursementSvc *mocks_service.IDisbursementService
	}

	mockBatchResponse := &disbursementModel.CRMBatchPayoutStatusResponse{
		Code: "00",
		Data: []disbursementModel.CRMPayoutStatusResult{
			{
				ReferenceID: "123e4567-e89b-12d3-a456-426614174001",
				Success:     true,
				Data: &disbursementModel.CRMPayoutStatusResponseData{
					DisbursementUUID:   "disbursement-uuid-1",
					ReferenceID:        "123e4567-e89b-12d3-a456-426614174001",
					Status:             "SUCCESS",
					ApprovalStatus:     "APPROVED",
					Amount:             "100000.00",
					BeneficiaryAccount: "1234567890",
					BeneficiaryName:    "John Doe",
					BeneficiaryBank:    "Bank ABC",
					TransferLogs: []disbursementModel.RoutingHistoryItem{
						{
							Order:       1,
							BankName:    "Bank ABC",
							Status:      "SUCCESS",
							ResponseMsg: "",
							Timestamp:   "2024-01-01T10:00:00Z",
						},
					},
				},
			},
			{
				ReferenceID: "123e4567-e89b-12d3-a456-426614174002",
				Success:     false,
				Error: &disbursementModel.CRMPayoutStatusError{
					Code:    "ERROR",
					Message: "Transaction not found",
				},
			},
		},
	}

	testCases := []struct {
		desc           string
		requestBody    interface{}
		expectedStatus int
		wantError      bool
		setupMock      func(mockers Mockers)
	}{
		{
			desc: "success - valid batch request",
			requestBody: disbursementModel.CRMBatchPayoutStatusRequest{
				ReferenceIDs: []string{"123e4567-e89b-12d3-a456-426614174001", "123e4567-e89b-12d3-a456-426614174002"},
			},
			expectedStatus: http.StatusOK,
			wantError:      false,
			setupMock: func(mockers Mockers) {
				mockers.disbursementSvc.On(
					"GetBatchPayoutStatusAndRouting",
					mock.Anything,
					mock.AnythingOfType("*disbursementModel.CRMBatchPayoutStatusRequest"),
				).Return(mockBatchResponse, nil)
			},
		},
		{
			desc: "error - empty reference IDs array",
			requestBody: disbursementModel.CRMBatchPayoutStatusRequest{
				ReferenceIDs: []string{},
			},
			expectedStatus: http.StatusBadRequest,
			wantError:      true,
			setupMock: func(mockers Mockers) {
				// No service call expected due to validation failure
			},
		},
		{
			desc:           "error - invalid JSON request body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			wantError:      true,
			setupMock: func(mockers Mockers) {
				// No service call expected due to JSON decode failure
			},
		},
		{
			desc: "error - service returns error",
			requestBody: disbursementModel.CRMBatchPayoutStatusRequest{
				ReferenceIDs: []string{"123e4567-e89b-12d3-a456-426614174001"},
			},
			expectedStatus: http.StatusInternalServerError,
			wantError:      true,
			setupMock: func(mockers Mockers) {
				mockers.disbursementSvc.On(
					"GetBatchPayoutStatusAndRouting",
					mock.Anything,
					mock.AnythingOfType("*disbursementModel.CRMBatchPayoutStatusRequest"),
				).Return(nil, pkgErrs.New(httpResponse.HttpErrDatabase, errors.New("database error")))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			mockers := Mockers{
				disbursementSvc: new(mocks_service.IDisbursementService),
			}
			tc.setupMock(mockers)

			handler := &handler{
				validator:       validatorExt.New(),
				disbursementSvc: mockers.disbursementSvc,
			}

			var reqBody []byte
			var err error
			if tc.requestBody != "invalid json" {
				reqBody, err = json.Marshal(tc.requestBody)
				assert.NoError(t, err)
			} else {
				reqBody = []byte(tc.requestBody.(string))
			}

			req := httptest.NewRequest(http.MethodPost, "/crm/v1/disbursements/batch-payout-status", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(context.Background())

			rr := httptest.NewRecorder()

			handler.GetBatchPayoutStatusAndRouting(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			var response map[string]interface{}
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tc.wantError {
				assert.NotEqual(t, "00", response["code"])
				// Error responses might have "message" field, check if it exists
				if message, exists := response["message"]; exists && message != nil {
					assert.NotEmpty(t, message)
				}
			} else {
				// Expect ApiResponse structure
				assert.Equal(t, "00", response["code"])
				assert.Equal(t, "OK", response["message"])
				assert.NotNil(t, response["data"])
				
				// The service response is nested inside the API response data
				serviceResponse := response["data"].(map[string]interface{})
				assert.Equal(t, "00", serviceResponse["code"])
				assert.NotNil(t, serviceResponse["data"])
				
				// Verify batch response data structure - should be an array inside the service response
				responseData := serviceResponse["data"].([]interface{})
				assert.Len(t, responseData, len(mockBatchResponse.Data))
				
				// Verify first result structure
				firstResult := responseData[0].(map[string]interface{})
				assert.Equal(t, "123e4567-e89b-12d3-a456-426614174001", firstResult["referenceId"])
				assert.Equal(t, true, firstResult["success"])
				assert.NotNil(t, firstResult["data"])
				
				// Verify second result structure (error case)
				secondResult := responseData[1].(map[string]interface{})
				assert.Equal(t, "123e4567-e89b-12d3-a456-426614174002", secondResult["referenceId"])
				assert.Equal(t, false, secondResult["success"])
				assert.NotNil(t, secondResult["error"])
			}

			mockers.disbursementSvc.AssertExpectations(t)
		})
	}
}