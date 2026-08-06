package merchant

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockUser "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCRMMerchantController_BulkCreateSubmerchant(t *testing.T) {
	tests := []struct {
		name                string
		merchantId          string
		kycType             string
		isInvitePIC         string
		csvContent          string
		csvFilename         string
		setupMocks          func(merchantSvc *mockMerchant.IMerchantService)
		expectedStatusCode  int
		expectedErrorCode   string
		expectedSuccessResp *merchant.BulkCreateSubMerchantResponse
	}{
		{
			name:        "SUCCESS: Valid bulk creation request",
			merchantId:  "123e4567-e89b-12d3-a456-426614174000",
			kycType:     "KYC",
			isInvitePIC: "true",
			csvContent:  "Name,ShortName,Logo,Email,Phone,Country,BusinessType,BusinessStructure,PICName,PICPhone,PICEmail,Address,PostCode,AccountNumber,ChannelCode\nTest Merchant,TM,logo.png,test@merchant.com,08123456789,ID,INDIVIDUAL,PT,John Doe,08987654321,john@test.com,Jl Test 123,12345,1234567890,BCA",
			csvFilename: "test.csv",
			setupMocks: func(merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.On("BulkCreateSubMerchant", mock.Anything, mock.MatchedBy(func(req *merchant.BulkCreateSubMerchantRequest) bool {
					return req.MerchantId == "123e4567-e89b-12d3-a456-426614174000" &&
						req.KYCType == constant.MerchantKYCTypeKYC &&
						req.FileName == "test.csv" &&
						len(req.SubmerchantDetails) == 1
				})).Return(&merchant.BulkCreateSubMerchantResponse{
					ID:          "session-123",
					TotalFailed: 0,
					Results:     []merchant.BulkCreateSubMerchantDetailResponse{},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccessResp: &merchant.BulkCreateSubMerchantResponse{
				ID:          "session-123",
				TotalFailed: 0,
				Results:     []merchant.BulkCreateSubMerchantDetailResponse{},
			},
		},
		{
			name:               "ERROR: Invalid merchant ID",
			merchantId:         "invalid-uuid",
			kycType:            "KYC",
			isInvitePIC:        "true",
			csvContent:         "test,data",
			csvFilename:        "test.csv",
			setupMocks:         func(merchantSvc *mockMerchant.IMerchantService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "40",
		},
		{
			name:               "ERROR: Missing merchant ID",
			merchantId:         "",
			kycType:            "KYC",
			isInvitePIC:        "true",
			csvContent:         "test,data",
			csvFilename:        "test.csv",
			setupMocks:         func(merchantSvc *mockMerchant.IMerchantService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "40",
		},
		{
			name:               "ERROR: Invalid KYC type",
			merchantId:         "123e4567-e89b-12d3-a456-426614174000",
			kycType:            "INVALID",
			isInvitePIC:        "true",
			csvContent:         "test,data",
			csvFilename:        "test.csv",
			setupMocks:         func(merchantSvc *mockMerchant.IMerchantService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "40",
		},
		{
			name:               "ERROR: Empty KYC type",
			merchantId:         "123e4567-e89b-12d3-a456-426614174000",
			kycType:            "",
			isInvitePIC:        "true",
			csvContent:         "test,data",
			csvFilename:        "test.csv",
			setupMocks:         func(merchantSvc *mockMerchant.IMerchantService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "40",
		},
		{
			name:               "ERROR: No file provided",
			merchantId:         "123e4567-e89b-12d3-a456-426614174000",
			kycType:            "KYC",
			isInvitePIC:        "true",
			csvContent:         "",
			csvFilename:        "",
			setupMocks:         func(merchantSvc *mockMerchant.IMerchantService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "40",
		},
		{
			name:               "ERROR: Non-CSV file",
			merchantId:         "123e4567-e89b-12d3-a456-426614174000",
			kycType:            "KYC",
			isInvitePIC:        "true",
			csvContent:         "test data",
			csvFilename:        "test.txt",
			setupMocks:         func(merchantSvc *mockMerchant.IMerchantService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "40",
		},
		{
			name:               "ERROR: Empty CSV file",
			merchantId:         "123e4567-e89b-12d3-a456-426614174000",
			kycType:            "KYC",
			isInvitePIC:        "true",
			csvContent:         "Name,ShortName,Logo\n",
			csvFilename:        "empty.csv",
			setupMocks:         func(merchantSvc *mockMerchant.IMerchantService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "40",
		},
		{
			name:        "ERROR: Service returns error",
			merchantId:  "123e4567-e89b-12d3-a456-426614174000",
			kycType:     "KYC",
			isInvitePIC: "true",
			csvContent:  "Name,ShortName,Logo,Email,Phone,Country,BusinessType,BusinessStructure,PICName,PICPhone,PICEmail,Address,PostCode,AccountNumber,ChannelCode\nTest Merchant,TM,logo.png,test@merchant.com,08123456789,ID,INDIVIDUAL,PT,John Doe,08987654321,john@test.com,Jl Test 123,12345,1234567890,BCA",
			csvFilename: "test.csv",
			setupMocks: func(merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.On("BulkCreateSubMerchant", mock.Anything, mock.Anything).Return(nil, pkgErrs.New(response.HttpErrInternal, errors.New("internal service error")))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorCode:  "99",
		},
		{
			name:        "SUCCESS: NonKYC type",
			merchantId:  "123e4567-e89b-12d3-a456-426614174000",
			kycType:     "NON_KYC",
			isInvitePIC: "true",
			csvContent:  "Name,ShortName,Logo,Email,Phone,Country,BusinessType,BusinessStructure,PICName,PICPhone,PICEmail,Address,PostCode,AccountNumber,ChannelCode\nTest Merchant,TM,logo.png,test@merchant.com,08123456789,ID,INDIVIDUAL,PT,John Doe,08987654321,john@test.com,Jl Test 123,12345,1234567890,BCA",
			csvFilename: "test.csv",
			setupMocks: func(merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.On("BulkCreateSubMerchant", mock.Anything, mock.MatchedBy(func(req *merchant.BulkCreateSubMerchantRequest) bool {
					return req.KYCType == constant.MerchantKYCTypeNonKYC
				})).Return(&merchant.BulkCreateSubMerchantResponse{
					ID:          "session-456",
					TotalFailed: 0,
					Results:     []merchant.BulkCreateSubMerchantDetailResponse{},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccessResp: &merchant.BulkCreateSubMerchantResponse{
				ID:          "session-456",
				TotalFailed: 0,
				Results:     []merchant.BulkCreateSubMerchantDetailResponse{},
			},
		},
		{
			name:        "SUCCESS: Multiple merchants in CSV",
			merchantId:  "123e4567-e89b-12d3-a456-426614174000",
			kycType:     "KYC",
			isInvitePIC: "true",
			csvContent:  "Name,ShortName,Logo,Email,Phone,Country,BusinessType,BusinessStructure,PICName,PICPhone,PICEmail,Address,PostCode,AccountNumber,ChannelCode\nMerchant 1,M1,logo1.png,merchant1@test.com,08111111111,ID,INDIVIDUAL,PT,PIC 1,08222222222,pic1@test.com,Address 1,11111,1111111111,BCA\nMerchant 2,M2,logo2.png,merchant2@test.com,08333333333,ID,INDIVIDUAL,PT,PIC 2,08444444444,pic2@test.com,Address 2,22222,2222222222,BCA",
			csvFilename: "multiple.csv",
			setupMocks: func(merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.On("BulkCreateSubMerchant", mock.Anything, mock.MatchedBy(func(req *merchant.BulkCreateSubMerchantRequest) bool {
					return len(req.SubmerchantDetails) == 2
				})).Return(&merchant.BulkCreateSubMerchantResponse{
					ID:          "session-789",
					TotalFailed: 0,
					Results:     []merchant.BulkCreateSubMerchantDetailResponse{},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccessResp: &merchant.BulkCreateSubMerchantResponse{
				ID:          "session-789",
				TotalFailed: 0,
				Results:     []merchant.BulkCreateSubMerchantDetailResponse{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockUserSvc := mockUser.NewIUserService(t)
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			validator := validator.New()
			logger, _ := logger.NewZapLogger(logger.Config{})

			tt.setupMocks(mockMerchantSvc)

			// Create controller
			controller := New(mockMerchantSvc, mockUserSvc, validator, mockRmq, WithLogger(logger))

			// Create multipart form request
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)

			// Add form fields
			_ = writer.WriteField("merchantId", tt.merchantId)
			_ = writer.WriteField("kycType", tt.kycType)
			_ = writer.WriteField("isInvitePIC", tt.isInvitePIC)

			// Add file if provided
			if tt.csvFilename != "" {
				part, err := writer.CreateFormFile("file", tt.csvFilename)
				assert.NoError(t, err)
				_, err = io.Copy(part, strings.NewReader(tt.csvContent))
				assert.NoError(t, err)
			}
			writer.Close()

			// Create HTTP request
			req := httptest.NewRequest(http.MethodPost, "/bulk-create-submerchant", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			// Execute
			controller.BulkCreateSubmerchant(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.expectedErrorCode != "" {
				// Verify error response structure
				assert.Contains(t, w.Body.String(), tt.expectedErrorCode)
			}

			if tt.expectedSuccessResp != nil {
				// Verify success response structure
				assert.Contains(t, w.Body.String(), tt.expectedSuccessResp.ID)
			}

			// Assert mock expectations
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}

func TestCRMMerchantController_GetBulkCreateSubmerchantSummary(t *testing.T) {
	tests := []struct {
		name                string
		merchantId          string
		sessionId           string
		setupMocks          func(merchantSvc *mockMerchant.IMerchantService)
		expectedStatusCode  int
		expectedErrorCode   string
		expectedSuccessResp *merchant.BulkCreateSubMerchantResponse
	}{
		{
			name:       "SUCCESS: Valid summary request",
			merchantId: "123e4567-e89b-12d3-a456-426614174000",
			sessionId:  "456e7890-e12c-34d5-b678-901234567890",
			setupMocks: func(merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.On("GetBulkCreateSubMerchantSummary", mock.Anything, mock.MatchedBy(func(req *merchant.GetBulkCreateSubMerchantSummaryRequest) bool {
					return req.MerchantId == "123e4567-e89b-12d3-a456-426614174000" &&
						req.ID == "456e7890-e12c-34d5-b678-901234567890"
				})).Return(&merchant.BulkCreateSubMerchantResponse{
					ID:           "456e7890-e12c-34d5-b678-901234567890",
					FileName:     "test.csv",
					TotalSuccess: 5,
					TotalFailed:  2,
					Results: []merchant.BulkCreateSubMerchantDetailResponse{
						{
							Row:          0,
							MerchantID:   "merchant-1",
							MerchantName: "Merchant 1",
						},
						{
							Row:   1,
							Error: "validation error",
						},
					},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccessResp: &merchant.BulkCreateSubMerchantResponse{
				ID:           "456e7890-e12c-34d5-b678-901234567890",
				FileName:     "test.csv",
				TotalSuccess: 5,
				TotalFailed:  2,
			},
		},
		{
			name:               "ERROR: Invalid merchant ID",
			merchantId:         "invalid-uuid",
			sessionId:          "456e7890-e12c-34d5-b678-901234567890",
			setupMocks:         func(merchantSvc *mockMerchant.IMerchantService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "40",
		},
		{
			name:               "ERROR: Invalid session ID",
			merchantId:         "123e4567-e89b-12d3-a456-426614174000",
			sessionId:          "invalid-uuid",
			setupMocks:         func(merchantSvc *mockMerchant.IMerchantService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrorCode:  "40",
		},
		{
			name:       "ERROR: Service returns error",
			merchantId: "123e4567-e89b-12d3-a456-426614174000",
			sessionId:  "456e7890-e12c-34d5-b678-901234567890",
			setupMocks: func(merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.On("GetBulkCreateSubMerchantSummary", mock.Anything, mock.Anything).Return(nil, pkgErrs.New(response.HttpErrInternal, errors.New("redis connection error")))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrorCode:  "99",
		},
		{
			name:       "SUCCESS: Empty results",
			merchantId: "123e4567-e89b-12d3-a456-426614174000",
			sessionId:  "456e7890-e12c-34d5-b678-901234567890",
			setupMocks: func(merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.On("GetBulkCreateSubMerchantSummary", mock.Anything, mock.Anything).Return(&merchant.BulkCreateSubMerchantResponse{
					ID:           "456e7890-e12c-34d5-b678-901234567890",
					FileName:     "empty.csv",
					TotalSuccess: 0,
					TotalFailed:  0,
					Results:      []merchant.BulkCreateSubMerchantDetailResponse{},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccessResp: &merchant.BulkCreateSubMerchantResponse{
				ID:           "456e7890-e12c-34d5-b678-901234567890",
				FileName:     "empty.csv",
				TotalSuccess: 0,
				TotalFailed:  0,
			},
		},
		{
			name:       "SUCCESS: All failed results",
			merchantId: "123e4567-e89b-12d3-a456-426614174000",
			sessionId:  "456e7890-e12c-34d5-b678-901234567890",
			setupMocks: func(merchantSvc *mockMerchant.IMerchantService) {
				merchantSvc.On("GetBulkCreateSubMerchantSummary", mock.Anything, mock.Anything).Return(&merchant.BulkCreateSubMerchantResponse{
					ID:           "456e7890-e12c-34d5-b678-901234567890",
					FileName:     "failed.csv",
					TotalSuccess: 0,
					TotalFailed:  3,
					Results: []merchant.BulkCreateSubMerchantDetailResponse{
						{
							Row:   0,
							Error: "validation error 1",
						},
						{
							Row:   1,
							Error: "validation error 2",
						},
						{
							Row:   2,
							Error: "validation error 3",
						},
					},
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedSuccessResp: &merchant.BulkCreateSubMerchantResponse{
				ID:           "456e7890-e12c-34d5-b678-901234567890",
				FileName:     "failed.csv",
				TotalSuccess: 0,
				TotalFailed:  3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockMerchantSvc := mockMerchant.NewIMerchantService(t)
			mockUserSvc := mockUser.NewIUserService(t)
			mockRmq := mockRabbitMq.NewRabbitMQExt(t)
			validator := validator.New()
			logger, _ := logger.NewZapLogger(logger.Config{})

			tt.setupMocks(mockMerchantSvc)

			// Create controller
			controller := New(mockMerchantSvc, mockUserSvc, validator, mockRmq, WithLogger(logger))

			// Create HTTP request with path parameters
			url := fmt.Sprintf("/merchant/%s/bulk-create-submerchant/%s", tt.merchantId, tt.sessionId)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.SetPathValue("merchantId", tt.merchantId)
			req.SetPathValue("sessionId", tt.sessionId)
			w := httptest.NewRecorder()

			// Execute
			controller.GetBulkCreateSubmerchantSummary(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.expectedErrorCode != "" {
				// Verify error response structure
				assert.Contains(t, w.Body.String(), tt.expectedErrorCode)
			}

			if tt.expectedSuccessResp != nil {
				// Verify success response structure
				assert.Contains(t, w.Body.String(), tt.expectedSuccessResp.ID)
				assert.Contains(t, w.Body.String(), tt.expectedSuccessResp.FileName)
			}

			// Assert mock expectations
			mockMerchantSvc.AssertExpectations(t)
		})
	}
}
