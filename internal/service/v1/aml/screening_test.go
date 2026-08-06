package amlservice

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/amlProcessor"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAmlService_Screening(t *testing.T) {
	testCases := []struct {
		name       string
		request    *amlcommon.CheckRequest
		provider   string
		merchantID string
		mockSetup  func(
			mockRepo map[string]*mocks.IAmlProcessorRepository,
			mockMerchantRepo *mocks.IMerchantRepository,
		)
		expectedStatus string
		wantErr        bool
		expectedError  error
	}{
		{
			name: "SUCCESS: screening with approved status",
			request: &amlcommon.CheckRequest{
				Name: "John Doe",
				DOB:  "1990-01-01",
			},
			provider:   "advance_ai",
			merchantID: "merchant-123",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check response
				nodeResult := &amlcommon.NodeResult{
					Detail: []amlcommon.NodeDetail{
						{
							Name: "John Doe",
						},
					},
				}

				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(&amlcommon.CheckResponse{
					TransactionID: "txn-123",
					Data: amlcommon.CheckResponseData{
						Status: constant.AML_STATUS_APPROVED,
					},
				}, nil)

				// Mock Inquiry response
				amlRepo.On("Inquiry", mock.Anything, "txn-123").Return(&amlcommon.InquiryResponse{
					Code:          "SUCCESS",
					Message:       "SUCCESS",
					TransactionID: "txn-123",
					Data: amlcommon.InquiryResponseData{
						ID: "inquiry-123",
						Nodes: []amlcommon.Node{
							{
								Type:        amlcommon.NodeTypeAMLNameScreening,
								CompletedAt: "2025-08-20 12:32:39",
								Attributes: map[string]any{
									"dob":               "1990-01-01",
									"name":              "John Doe",
									"score":             95.0,
									"gender":            "MALE",
									"entityType":        "PERSON",
									"referenceId":       "test-ref-123",
									"placeOfBirth":      "Indonesia",
									"countryLocation":   "Indonesia",
									"registeredCountry": "Indonesia",
									"hitCategory":       []any{"LE", "PEP"},
								},
								Result: nodeResult,
							},
						},
					},
				}, nil)

				// Mock merchant repository
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-123").Return(&merchant.Merchant{
					UUID: "merchant-123",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: []byte(`{}`),
						Valid:    true,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-123", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			expectedStatus: constant.AML_STATUS_APPROVED,
			wantErr:        false,
		},
		{
			name: "SUCCESS: screening with review status",
			request: &amlcommon.CheckRequest{
				Name: "Jane Smith",
				DOB:  "1985-05-15",
			},
			provider:   "advance_ai",
			merchantID: "merchant-456",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check response with REVIEW status
				nodeResult := &amlcommon.NodeResult{
					Detail: []amlcommon.NodeDetail{
						{
							Name: "Jane Smith",
						},
					},
				}

				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(&amlcommon.CheckResponse{
					TransactionID: "txn-456",
					Data: amlcommon.CheckResponseData{
						Status: constant.AML_STATUS_REVIEW,
					},
				}, nil)

				// Mock Inquiry response
				amlRepo.On("Inquiry", mock.Anything, "txn-456").Return(&amlcommon.InquiryResponse{
					Code:          "SUCCESS",
					Message:       "SUCCESS",
					TransactionID: "txn-456",
					Data: amlcommon.InquiryResponseData{
						ID: "inquiry-456",
						Nodes: []amlcommon.Node{
							{
								Type:        amlcommon.NodeTypeAMLNameScreening,
								CompletedAt: "2025-08-20 12:32:39",
								Attributes: map[string]any{
									"dob":               "1985-05-15",
									"name":              "Jane Smith",
									"score":             95.0,
									"gender":            "FEMALE",
									"entityType":        "PERSON",
									"referenceId":       "test-ref-456",
									"placeOfBirth":      "Indonesia",
									"countryLocation":   "Indonesia",
									"registeredCountry": "Indonesia",
									"hitCategory":       []any{"PEP", "SIC"},
								},
								Result: nodeResult,
							},
						},
					},
				}, nil)

				// Mock merchant repository with existing data
				existingData := map[string]*amlcommon.ScreeningResponse{
					"John Doe:1990-01-01": {
						Status:        constant.AML_STATUS_APPROVED,
						TransactionID: "old-txn",
						ReferenceID:   "old-ref",
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-456").Return(&merchant.Merchant{
					UUID: "merchant-456",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-456", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			expectedStatus: constant.AML_STATUS_NEED_REVIEW,
			wantErr:        false,
		},
		{
			name: "SUCCESS: ENTITY type screening without DOB",
			request: &amlcommon.CheckRequest{
				Name:        "PT PAPER INDONESIA",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			provider:   "advance_ai",
			merchantID: "merchant-entity-123",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check response
				nodeResult := &amlcommon.NodeResult{
					Detail: []amlcommon.NodeDetail{
						{
							Name: "PT PAPER INDONESIA",
						},
					},
				}

				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(&amlcommon.CheckResponse{
					TransactionID: "txn-entity-123",
					Data: amlcommon.CheckResponseData{
						Status: constant.AML_STATUS_APPROVED,
					},
				}, nil)

				// Mock Inquiry response
				amlRepo.On("Inquiry", mock.Anything, "txn-entity-123").Return(&amlcommon.InquiryResponse{
					Code:          "SUCCESS",
					Message:       "SUCCESS",
					TransactionID: "txn-entity-123",
					Data: amlcommon.InquiryResponseData{
						ID: "inquiry-entity-123",
						Nodes: []amlcommon.Node{
							{
								Type:        amlcommon.NodeTypeAMLNameScreening,
								CompletedAt: "2025-08-20 12:32:39",
								Attributes: map[string]any{
									"name":              "PT PAPER INDONESIA",
									"score":             85.0,
									"entityType":        "ENTITY",
									"referenceId":       "test-ref-entity-123",
									"countryLocation":   "Indonesia",
									"registeredCountry": "Indonesia",
									"hitCategory":       []any{"SAN"},
								},
								Result: nodeResult,
							},
						},
					},
				}, nil)

				// Mock merchant repository
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-entity-123").Return(&merchant.Merchant{
					UUID: "merchant-entity-123",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: []byte(`{}`),
						Valid:    true,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-entity-123", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			expectedStatus: constant.AML_STATUS_APPROVED,
			wantErr:        false,
		},
		{
			name: "SUCCESS: ENTITY type screening with existing merchant data",
			request: &amlcommon.CheckRequest{
				Name:        "PT ABC COMPANY",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			provider:   "advance_ai",
			merchantID: "merchant-entity-existing",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check response
				nodeResult := &amlcommon.NodeResult{
					Detail: []amlcommon.NodeDetail{
						{
							Name: "PT ABC COMPANY",
						},
					},
				}

				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(&amlcommon.CheckResponse{
					TransactionID: "txn-entity-existing",
					Data: amlcommon.CheckResponseData{
						Status: constant.AML_STATUS_APPROVED,
					},
				}, nil)

				// Mock Inquiry response
				amlRepo.On("Inquiry", mock.Anything, "txn-entity-existing").Return(&amlcommon.InquiryResponse{
					Code:          "SUCCESS",
					Message:       "SUCCESS",
					TransactionID: "txn-entity-existing",
					Data: amlcommon.InquiryResponseData{
						ID: "inquiry-entity-existing",
						Nodes: []amlcommon.Node{
							{
								Type:        amlcommon.NodeTypeAMLNameScreening,
								CompletedAt: "2025-08-20 12:32:39",
								Attributes: map[string]any{
									"name":              "PT ABC COMPANY",
									"score":             90.0,
									"entityType":        "ENTITY",
									"referenceId":       "test-ref-entity-existing",
									"countryLocation":   "Indonesia",
									"registeredCountry": "Indonesia",
									"hitCategory":       []any{"LE"},
								},
								Result: nodeResult,
							},
						},
					},
				}, nil)

				// Mock merchant repository with existing person data
				existingData := map[string]*amlcommon.ScreeningResponse{
					"John Doe:1990-01-01": {
						Status:        constant.AML_STATUS_APPROVED,
						TransactionID: "old-person-txn",
						ReferenceID:   "old-person-ref",
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-entity-existing").Return(&merchant.Merchant{
					UUID: "merchant-entity-existing",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-entity-existing", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			expectedStatus: constant.AML_STATUS_APPROVED,
			wantErr:        false,
		},
		{
			name: "FAILED: provider not found",
			request: &amlcommon.CheckRequest{
				Name: "Test User",
				DOB:  "1992-03-10",
			},
			provider:   "invalid_provider",
			merchantID: "merchant-123",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// No provider setup - should fail
			},
			expectedStatus: constant.AML_STATUS_APPROVED,
			wantErr:        true,
			expectedError:  constant.ErrProviderNotFound,
		},
		{
			name: "FAILED: check error",
			request: &amlcommon.CheckRequest{
				Name: "Error User",
				DOB:  "1988-12-25",
			},
			provider:   "advance_ai",
			merchantID: "merchant-123",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check error
				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(nil, errors.New("check error"))
			},
			expectedStatus: constant.AML_STATUS_APPROVED,
			wantErr:        true,
			expectedError:  errors.New("check error"),
		},
		{
			name: "FAILED: inquiry error",
			request: &amlcommon.CheckRequest{
				Name: "Inquiry Error User",
				DOB:  "1995-07-20",
			},
			provider:   "advance_ai",
			merchantID: "merchant-123",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check success
				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(&amlcommon.CheckResponse{
					TransactionID: "txn-error",
					Data: amlcommon.CheckResponseData{
						Status: constant.AML_STATUS_APPROVED,
					},
				}, nil)

				// Mock Inquiry error
				amlRepo.On("Inquiry", mock.Anything, "txn-error").Return(nil, errors.New("inquiry error"))
			},
			expectedStatus: constant.AML_STATUS_APPROVED,
			wantErr:        true,
			expectedError:  errors.New("inquiry error"),
		},
		{
			name: "SUCCESS: merchant save error but screening continues",
			request: &amlcommon.CheckRequest{
				Name: "Save Error User",
				DOB:  "1991-11-08",
			},
			provider:   "advance_ai",
			merchantID: "merchant-save-error",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check and Inquiry success
				nodeResult := &amlcommon.NodeResult{
					Detail: []amlcommon.NodeDetail{
						{
							Name: "Save Error User",
						},
					},
				}

				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(&amlcommon.CheckResponse{
					TransactionID: "txn-save-error",
					Data: amlcommon.CheckResponseData{
						Status: constant.AML_STATUS_APPROVED,
					},
				}, nil)

				amlRepo.On("Inquiry", mock.Anything, "txn-save-error").Return(&amlcommon.InquiryResponse{
					Code:          "SUCCESS",
					TransactionID: "txn-save-error",
					Data: amlcommon.InquiryResponseData{
						ID: "inquiry-save-error",
						Nodes: []amlcommon.Node{
							{
								Type:        amlcommon.NodeTypeAMLNameScreening,
								CompletedAt: "2025-08-20 12:32:39",
								Attributes: map[string]any{
									"dob":               "1991-11-08",
									"name":              "Save Error User",
									"score":             95.0,
									"gender":            "MALE",
									"entityType":        "PERSON",
									"referenceId":       "test-ref-save-error",
									"placeOfBirth":      "Indonesia",
									"countryLocation":   "Indonesia",
									"registeredCountry": "Indonesia",
									"hitCategory":       []any{"LE", "OB"},
								},
								Result: nodeResult,
							},
						},
					},
				}, nil)

				// Mock merchant not found (save will fail but screening continues)
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-save-error").Return(nil, errors.New("merchant not found"))
			},
			expectedStatus: constant.AML_STATUS_APPROVED,
			wantErr:        false, // Should not fail even if save fails
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMerchantRepo := mocks.NewIMerchantRepository(t)
			mockRepoMap := make(map[string]*mocks.IAmlProcessorRepository)

			tc.mockSetup(mockRepoMap, mockMerchantRepo)

			// Convert mock map to interface map
			repoMap := make(map[string]repository.IAmlProcessorRepository)
			for k, v := range mockRepoMap {
				repoMap[k] = v
			}

			// Create service
			cfg := &config.Config{}
			service := &AmlService{
				cfg:                 cfg,
				logger:              mockLogger,
				thirdPartyProcessor: repoMap,
				merchantRepository:  mockMerchantRepo,
			}

			// Call the method
			result, err := service.Screening(context.Background(), tc.request, tc.provider, tc.merchantID)

			// Assertions
			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Equal(t, tc.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.expectedStatus, result.Status)
				assert.NotEmpty(t, result.TransactionID)
				assert.NotEmpty(t, result.ReferenceID)
				// For successful cases, we should have a screening result
				if result.Status != constant.AML_STATUS_APPROVED || tc.name == "SUCCESS: screening with review status" {
					assert.NotNil(t, result.Result)
				}
			}
		})
	}
}

func TestAmlService_saveScreeningDataToMerchant(t *testing.T) {
	testCases := []struct {
		name          string
		merchantID    string
		request       *amlcommon.CheckRequest
		screeningResp *amlcommon.ScreeningResponse
		mockSetup     func(mockMerchantRepo *mocks.IMerchantRepository)
		wantErr       bool
		expectedError error
	}{
		{
			name:       "SUCCESS: save to empty merchant data",
			merchantID: "merchant-123",
			request: &amlcommon.CheckRequest{
				Name: "John Doe",
				DOB:  "1990-01-01",
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-123",
				ReferenceID:   "ref-123",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-123").Return(&merchant.Merchant{
					UUID: "merchant-123",
					ThirdPartyScreeningData: types.NullJSONText{
						Valid: false,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-123", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "FAILED: merchant not found",
			merchantID: "nonexistent-merchant",
			request: &amlcommon.CheckRequest{
				Name: "Test User",
				DOB:  "1985-06-15",
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-fail",
				ReferenceID:   "ref-fail",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "nonexistent-merchant").Return(nil, errors.New("merchant not found"))
			},
			wantErr:       true,
			expectedError: constant.ErrMerchantNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMerchantRepo := mocks.NewIMerchantRepository(t)

			tc.mockSetup(mockMerchantRepo)

			// Create service
			cfg := &config.Config{}
			service := &AmlService{
				cfg:                cfg,
				logger:             mockLogger,
				merchantRepository: mockMerchantRepo,
			}

			// Call the method
			err := service.saveScreeningDataToMerchant(context.Background(), tc.merchantID, tc.request, tc.screeningResp)

			// Assertions
			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Equal(t, tc.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Setup
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMerchantRepo := mocks.NewIMerchantRepository(t)
	cfg := &config.Config{}

	// Create processor map
	thirdPartyProcessor := make(map[string]repository.IAmlProcessorRepository)
	amlRepo := mocks.NewIAmlProcessorRepository(t)
	thirdPartyProcessor["advance_ai"] = amlRepo

	// Test New function
	service := New(cfg, mockLogger, mockMerchantRepo, thirdPartyProcessor)

	// Assertions
	assert.NotNil(t, service)
}

func TestNew_WithDependencies(t *testing.T) {
	// Setup
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMerchantRepo := mocks.NewIMerchantRepository(t)
	mockOutboundRepo := mocks.NewIOutboundRepository(t)
	cfg := &config.Config{}

	// Create processor map
	thirdPartyProcessor := make(map[string]repository.IAmlProcessorRepository)
	amlRepo := mocks.NewIAmlProcessorRepository(t)
	thirdPartyProcessor["advance_ai"] = amlRepo

	// Test New function with dependencies
	service := New(
		cfg,
		mockLogger,
		mockMerchantRepo,
		thirdPartyProcessor,
		WithOutboundRepository(mockOutboundRepo),
	)

	// Assertions
	assert.NotNil(t, service)
}

func TestAmlService_extractScreeningResult(t *testing.T) {
	testCases := []struct {
		name            string
		inquiryResponse *amlcommon.InquiryResponse
		expected        *amlcommon.ScreeningResult
		expectNil       bool
	}{
		{
			name: "SUCCESS: extract valid AML Name Screening result",
			inquiryResponse: &amlcommon.InquiryResponse{
				Code:          "SUCCESS",
				TransactionID: "txn-123",
				Data: amlcommon.InquiryResponseData{
					ID: "inquiry-123",
					Nodes: []amlcommon.Node{
						{
							Type:        amlcommon.NodeTypeAMLNameScreening,
							CompletedAt: "2025-08-20 12:32:39",
							Attributes: map[string]any{
								"dob":               "1990-01-01",
								"name":              "John Doe",
								"score":             95.0,
								"gender":            "MALE",
								"entityType":        "PERSON",
								"referenceId":       "ref-123",
								"placeOfBirth":      "Indonesia",
								"countryLocation":   "Indonesia",
								"registeredCountry": "Indonesia",
								"hitCategory":       []any{"PEP", "LE"},
							},
							Result: &amlcommon.NodeResult{
								Detail: []amlcommon.NodeDetail{
									{
										Name:      "John DOE",
										ProfileID: "e_tr_wci_123",
										Type:      "POLITICAL INDIVIDUAL",
									},
								},
								Summary: amlcommon.NodeSummary{
									Total:        1,
									FullMarksNum: 1,
									IconHitSet:   []string{"PEP"},
								},
								MatchedCount: 1,
							},
						},
					},
				},
			},
			expected: &amlcommon.ScreeningResult{
				ID:            "inquiry-123",
				CompletedAt:   "2025-08-20 12:32:39",
				TransactionID: "txn-123",
				Detail: []amlcommon.ScreeningDetailItem{
					{
						NodeDetail: amlcommon.NodeDetail{
							Name:      "John DOE",
							ProfileID: "e_tr_wci_123",
							Type:      "POLITICAL INDIVIDUAL",
						},
					},
				},
				Summary: amlcommon.NodeSummary{
					Total:        1,
					FullMarksNum: 1,
					IconHitSet:   []string{"PEP"},
				},
				MatchedCount: 1,
				Attributes: amlcommon.ScreeningAttributes{
					DOB:               "1990-01-01",
					Name:              "John Doe",
					Score:             95,
					Gender:            "MALE",
					EntityType:        "PERSON",
					ReferenceID:       "ref-123",
					PlaceOfBirth:      "Indonesia",
					CountryLocation:   "Indonesia",
					RegisteredCountry: "Indonesia",
					HitCategory:       []string{"PEP", "LE"},
				},
			},
			expectNil: false,
		},
		{
			name: "FAIL: no AML Name Screening node found",
			inquiryResponse: &amlcommon.InquiryResponse{
				Code:          "SUCCESS",
				TransactionID: "txn-456",
				Data: amlcommon.InquiryResponseData{
					ID: "inquiry-456",
					Nodes: []amlcommon.Node{
						{
							Type: "Other Screening",
							Result: &amlcommon.NodeResult{
								Detail: []amlcommon.NodeDetail{},
							},
						},
					},
				},
			},
			expected:  nil,
			expectNil: true,
		},
		{
			name: "FAIL: AML Name Screening node has no result",
			inquiryResponse: &amlcommon.InquiryResponse{
				Code:          "SUCCESS",
				TransactionID: "txn-789",
				Data: amlcommon.InquiryResponseData{
					ID: "inquiry-789",
					Nodes: []amlcommon.Node{
						{
							Type:   amlcommon.NodeTypeAMLNameScreening,
							Result: nil,
						},
					},
				},
			},
			expected:  nil,
			expectNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			cfg := &config.Config{}
			service := &AmlService{
				cfg:    cfg,
				logger: mockLogger,
			}

			// Call the method
			result := service.extractScreeningResult(tc.inquiryResponse)

			// Assertions
			if tc.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.expected.ID, result.ID)
				assert.Equal(t, tc.expected.CompletedAt, result.CompletedAt)
				assert.Equal(t, tc.expected.TransactionID, result.TransactionID)
				assert.Equal(t, tc.expected.MatchedCount, result.MatchedCount)
				assert.Equal(t, len(tc.expected.Detail), len(result.Detail))
				assert.Equal(t, tc.expected.Attributes.Name, result.Attributes.Name)
				assert.Equal(t, tc.expected.Attributes.DOB, result.Attributes.DOB)
				assert.Equal(t, tc.expected.Attributes.Score, result.Attributes.Score)
				assert.Equal(t, tc.expected.Attributes.HitCategory, result.Attributes.HitCategory)
			}
		})
	}
}

func TestAmlService_saveScreeningDataToMerchant_AdditionalCases(t *testing.T) {
	testCases := []struct {
		name          string
		merchantID    string
		request       *amlcommon.CheckRequest
		screeningResp *amlcommon.ScreeningResponse
		mockSetup     func(mockMerchantRepo *mocks.IMerchantRepository)
		wantErr       bool
		expectedError error
	}{
		{
			name:       "SUCCESS: save with existing merchant data",
			merchantID: "merchant-existing",
			request: &amlcommon.CheckRequest{
				Name: "Jane Smith",
				DOB:  "1985-05-15",
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-existing",
				ReferenceID:   "ref-existing",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				existingData := map[string]*amlcommon.ScreeningResponse{
					"John Doe:1990-01-01": {
						Status:        constant.AML_STATUS_APPROVED,
						TransactionID: "old-txn",
						ReferenceID:   "old-ref",
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-existing").Return(&merchant.Merchant{
					UUID: "merchant-existing",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-existing", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "FAILED: invalid existing JSON data",
			merchantID: "merchant-invalid-json",
			request: &amlcommon.CheckRequest{
				Name: "Test User",
				DOB:  "1992-03-10",
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-invalid",
				ReferenceID:   "ref-invalid",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-invalid-json").Return(&merchant.Merchant{
					UUID: "merchant-invalid-json",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: []byte(`invalid json`),
						Valid:    true,
					},
				}, nil)

				// Even with invalid JSON, the service should continue and save new data
				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-invalid-json", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			wantErr: false, // Changed to false since it should succeed after creating new map
		},
		{
			name:       "FAILED: database error during save",
			merchantID: "merchant-db-error",
			request: &amlcommon.CheckRequest{
				Name: "Database Error User",
				DOB:  "1988-12-25",
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-db-error",
				ReferenceID:   "ref-db-error",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-db-error").Return(&merchant.Merchant{
					UUID: "merchant-db-error",
					ThirdPartyScreeningData: types.NullJSONText{
						Valid: false,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-db-error", mock.AnythingOfType("types.NullJSONText")).Return(errors.New("database error"))
			},
			wantErr:       true,
			expectedError: errors.New("database error"),
		},
		{
			name:       "FAILED: missing name",
			merchantID: "merchant-valid",
			request: &amlcommon.CheckRequest{
				Name: "",
				DOB:  "1990-01-01",
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-valid",
				ReferenceID:   "ref-valid",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				// No setup needed since validation should fail before database call
			},
			wantErr:       true,
			expectedError: constant.ErrInvalidRequestPayload,
		},
		{
			name:       "FAILED: missing DOB",
			merchantID: "merchant-valid",
			request: &amlcommon.CheckRequest{
				Name: "John Doe",
				DOB:  "",
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-valid",
				ReferenceID:   "ref-valid",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				// No setup needed since validation should fail before database call
			},
			wantErr:       true,
			expectedError: constant.ErrInvalidRequestPayload,
		},
		{
			name:       "SUCCESS: save ENTITY type without DOB",
			merchantID: "merchant-entity-123",
			request: &amlcommon.CheckRequest{
				Name:        "PT PAPER INDONESIA",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-entity-123",
				ReferenceID:   "ref-entity-123",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-entity-123").Return(&merchant.Merchant{
					UUID: "merchant-entity-123",
					ThirdPartyScreeningData: types.NullJSONText{
						Valid: false,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-entity-123", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "SUCCESS: save ENTITY type with existing person data",
			merchantID: "merchant-mixed-data",
			request: &amlcommon.CheckRequest{
				Name:        "PT ABC COMPANY",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-entity-mixed",
				ReferenceID:   "ref-entity-mixed",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				// Existing merchant data with person screening
				existingData := map[string]*amlcommon.ScreeningResponse{
					"John Doe:1990-01-01": {
						Status:        constant.AML_STATUS_APPROVED,
						TransactionID: "old-person-txn",
						ReferenceID:   "old-person-ref",
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-mixed-data").Return(&merchant.Merchant{
					UUID: "merchant-mixed-data",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-mixed-data", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "SUCCESS: update existing ENTITY type data",
			merchantID: "merchant-entity-update",
			request: &amlcommon.CheckRequest{
				Name:        "PT PAPER INDONESIA",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_NEED_REVIEW,
				TransactionID: "txn-entity-update",
				ReferenceID:   "ref-entity-update",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				// Existing merchant data with same entity
				existingData := map[string]*amlcommon.ScreeningResponse{
					"PT PAPER INDONESIA": {
						Status:        constant.AML_STATUS_APPROVED,
						TransactionID: "old-entity-txn",
						ReferenceID:   "old-entity-ref",
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-entity-update").Return(&merchant.Merchant{
					UUID: "merchant-entity-update",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-entity-update", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "FAILED: ENTITY type missing name",
			merchantID: "merchant-valid",
			request: &amlcommon.CheckRequest{
				Name:        "",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-valid",
				ReferenceID:   "ref-valid",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				// No setup needed since validation should fail before database call
			},
			wantErr:       true,
			expectedError: constant.ErrInvalidRequestPayload,
		},
		{
			name:       "SUCCESS: PERSON type with DOB still works (backward compatibility)",
			merchantID: "merchant-person-compat",
			request: &amlcommon.CheckRequest{
				Name:        "Jane Smith",
				DOB:         "1985-05-15",
				SubjectType: constant.AML_SUBJECT_TYPE_PERSON,
			},
			screeningResp: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-person-compat",
				ReferenceID:   "ref-person-compat",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-person-compat").Return(&merchant.Merchant{
					UUID: "merchant-person-compat",
					ThirdPartyScreeningData: types.NullJSONText{
						Valid: false,
					},
				}, nil)

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-person-compat", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMerchantRepo := mocks.NewIMerchantRepository(t)

			tc.mockSetup(mockMerchantRepo)

			// Create service
			cfg := &config.Config{}
			service := &AmlService{
				cfg:                cfg,
				logger:             mockLogger,
				merchantRepository: mockMerchantRepo,
			}

			// Call the method
			err := service.saveScreeningDataToMerchant(context.Background(), tc.merchantID, tc.request, tc.screeningResp)

			// Assertions
			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Equal(t, tc.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
