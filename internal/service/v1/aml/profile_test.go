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
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAmlService_Profile(t *testing.T) {
	testCases := []struct {
		name       string
		request    *amlcommon.CheckRequest
		provider   string
		merchantID string
		mockSetup  func(
			mockRepo map[string]*mocks.IAmlProcessorRepository,
			mockMerchantRepo *mocks.IMerchantRepository,
		)
		wantErr       bool
		expectedError error
	}{
		{
			name: "SUCCESS: profile with existing merchant data",
			request: &amlcommon.CheckRequest{
				Name: "Ir Joko WIDODO",
				DOB:  "1961-06-21",
			},
			provider:   "advance_ai",
			merchantID: "merchant-123",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				existingScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "existing-txn-123",
					ReferenceID:   "ref-existing-123",
					Result: &amlcommon.ScreeningResult{
						ID:            "existing-txn-123",
						CompletedAt:   "2025-08-20 12:32:39",
						TransactionID: "existing-txn-123",
						Detail: []amlcommon.ScreeningDetailItem{
							{
								NodeDetail: amlcommon.NodeDetail{
									ProfileID: "e_tr_wci_1224148",
									Name:      "Ir Joko WIDODO",
								},
							},
						},
						Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
						MatchedCount: 1,
						Attributes: amlcommon.ScreeningAttributes{
							DOB:               "1961-06-21",
							Name:              "Ir Joko WIDODO",
							Score:             95,
							Gender:            "MALE",
							EntityType:        "PERSON",
							ReferenceID:       "ref-existing-123",
							PlaceOfBirth:      "Indonesia",
							CountryLocation:   "Indonesia",
							RegisteredCountry: "Indonesia",
							HitCategory:       []string{"PEP"},
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"Ir Joko WIDODO:1961-06-21": existingScreeningData,
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-123").Return(&merchant.Merchant{
					UUID: "merchant-123",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)

				// Mock ProfileDetail response
				amlRepo.On("ProfileDetail", mock.Anything, "existing-txn-123", "e_tr_wci_1224148").Return(&amlcommon.ProfileDetailResponse{
					Code:    "SUCCESS",
					Message: "SUCCESS",
					Data: amlcommon.ProfileDetailData{
						ProfileID: "e_tr_wci_1224148",
						Name:      "Ir Joko WIDODO",
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: profile with new screening (no merchant)",
			request: &amlcommon.CheckRequest{
				Name: "Ir Joko WIDODO",
				DOB:  "1961-06-21",
			},
			provider:   "advance_ai",
			merchantID: "",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// Since we're calling s.Screening(), we need to mock the underlying repo calls that Screening makes
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check and Inquiry responses that the Screening method will make
				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(&amlcommon.CheckResponse{
					TransactionID: "new-txn-456",
				}, nil)

				amlRepo.On("Inquiry", mock.Anything, "new-txn-456").Return(&amlcommon.InquiryResponse{
					TransactionID: "new-txn-456",
					Data: amlcommon.InquiryResponseData{
						ID: "new-txn-456",
						Nodes: []amlcommon.Node{
							{
								Type:        amlcommon.NodeTypeAMLNameScreening,
								Name:        "Aml Name Screening",
								CompletedAt: "2025-08-20 12:32:39",
								Attributes: map[string]any{
									"dob":               "1961-06-21",
									"name":              "Ir Joko WIDODO",
									"score":             95.0,
									"gender":            "MALE",
									"entityType":        "PERSON",
									"referenceId":       "ref-new-456",
									"placeOfBirth":      "Indonesia",
									"countryLocation":   "Indonesia",
									"registeredCountry": "Indonesia",
									"hitCategory":       []any{"PEP"},
								},
								Result: &amlcommon.NodeResult{
									Detail: []amlcommon.NodeDetail{
										{
											ProfileID: "e_tr_wci_1224148",
											Name:      "Ir Joko WIDODO",
										},
									},
									Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
									MatchedCount: 1,
								},
							},
						},
					},
				}, nil)

				// Mock ProfileDetail response
				amlRepo.On("ProfileDetail", mock.Anything, "new-txn-456", "e_tr_wci_1224148").Return(&amlcommon.ProfileDetailResponse{
					Code:    "SUCCESS",
					Message: "SUCCESS",
					Data: amlcommon.ProfileDetailData{
						ProfileID: "e_tr_wci_1224148",
						Name:      "Ir Joko WIDODO",
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "FAILED: provider not found",
			request: &amlcommon.CheckRequest{
				Name: "Test User",
				DOB:  "1990-01-01",
			},
			provider:   "invalid_provider",
			merchantID: "merchant-123",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// No provider setup - should fail
			},
			wantErr:       true,
			expectedError: constant.ErrProviderNotFound,
		},
		{
			name: "FAILED: merchant not found",
			request: &amlcommon.CheckRequest{
				Name: "Test User",
				DOB:  "1990-01-01",
			},
			provider:   "advance_ai",
			merchantID: "nonexistent-merchant",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "nonexistent-merchant").Return(nil, errors.New("merchant not found"))
			},
			wantErr:       true,
			expectedError: constant.ErrMerchantNotFound,
		},
		{
			name: "FAILED: profileID not found in screening data",
			request: &amlcommon.CheckRequest{
				Name: "No Profile User",
				DOB:  "1990-01-01",
			},
			provider:   "advance_ai",
			merchantID: "merchant-no-profile",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Create mock existing data without profile screening or with empty results
				existingScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "existing-txn-no-profile",
					ReferenceID:   "ref-no-profile",
					Result: &amlcommon.ScreeningResult{
						ID:            "existing-txn-no-profile",
						CompletedAt:   "2025-08-20 12:32:39",
						TransactionID: "existing-txn-no-profile",
						Detail:        []amlcommon.ScreeningDetailItem{}, // Empty detail for no profile case
						Summary:       amlcommon.NodeSummary{Total: 0, FullMarksNum: 0},
						MatchedCount:  0,
						Attributes: amlcommon.ScreeningAttributes{
							DOB:         "1990-01-01",
							Name:        "No Profile User",
							ReferenceID: "ref-no-profile",
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"No Profile User:1990-01-01": existingScreeningData,
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-no-profile").Return(&merchant.Merchant{
					UUID: "merchant-no-profile",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)
			},
			wantErr:       true,
			expectedError: constant.ErrDataNotFound,
		},
		{
			name: "FAILED: ProfileDetail API error",
			request: &amlcommon.CheckRequest{
				Name: "API Error User",
				DOB:  "1990-01-01",
			},
			provider:   "advance_ai",
			merchantID: "merchant-api-error",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Create mock existing data with profile screening
				existingScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "existing-txn-error",
					ReferenceID:   "ref-error-123",
					Result: &amlcommon.ScreeningResult{
						ID:            "existing-txn-error",
						CompletedAt:   "2025-08-20 12:32:39",
						TransactionID: "existing-txn-error",
						Detail: []amlcommon.ScreeningDetailItem{
							{
								NodeDetail: amlcommon.NodeDetail{
									ProfileID: "e_tr_wci_1224148",
									Name:      "API Error User",
								},
							},
						},
						Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
						MatchedCount: 1,
						Attributes: amlcommon.ScreeningAttributes{
							DOB:               "1990-01-01",
							Name:              "API Error User",
							Score:             95,
							Gender:            "MALE",
							EntityType:        "PERSON",
							ReferenceID:       "ref-error-123",
							PlaceOfBirth:      "Indonesia",
							CountryLocation:   "Indonesia",
							RegisteredCountry: "Indonesia",
							HitCategory:       []string{"PEP"},
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"API Error User:1990-01-01": existingScreeningData,
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-api-error").Return(&merchant.Merchant{
					UUID: "merchant-api-error",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)

				// Mock ProfileDetail error
				amlRepo.On("ProfileDetail", mock.Anything, "existing-txn-error", "e_tr_wci_1224148").Return(nil, errors.New("API error"))
			},
			wantErr:       true,
			expectedError: errors.New("API error"),
		},
		{
			name: "SUCCESS: ENTITY type profile with existing merchant data",
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

				existingScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "existing-txn-entity-123",
					ReferenceID:   "ref-existing-entity-123",
					Result: &amlcommon.ScreeningResult{
						ID:            "existing-txn-entity-123",
						CompletedAt:   "2025-08-20 12:32:39",
						TransactionID: "existing-txn-entity-123",
						Detail: []amlcommon.ScreeningDetailItem{
							{
								NodeDetail: amlcommon.NodeDetail{
									ProfileID: "e_tr_wci_entity_123",
									Name:      "PT PAPER INDONESIA",
								},
							},
						},
						Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
						MatchedCount: 1,
						Attributes: amlcommon.ScreeningAttributes{
							Name:              "PT PAPER INDONESIA",
							Score:             85,
							EntityType:        "ENTITY",
							ReferenceID:       "ref-existing-entity-123",
							CountryLocation:   "Indonesia",
							RegisteredCountry: "Indonesia",
							HitCategory:       []string{"SAN"},
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"PT PAPER INDONESIA": existingScreeningData,
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-entity-123").Return(&merchant.Merchant{
					UUID: "merchant-entity-123",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)

				// Mock ProfileDetail response
				amlRepo.On("ProfileDetail", mock.Anything, "existing-txn-entity-123", "e_tr_wci_entity_123").Return(&amlcommon.ProfileDetailResponse{
					Code:    "SUCCESS",
					Message: "SUCCESS",
					Data: amlcommon.ProfileDetailData{
						ProfileID: "e_tr_wci_entity_123",
						Name:      "PT PAPER INDONESIA",
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: ENTITY type profile with mixed merchant data",
			request: &amlcommon.CheckRequest{
				Name:        "PT ABC COMPANY",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			provider:   "advance_ai",
			merchantID: "merchant-mixed-entity",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				entityScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "existing-txn-entity-mixed",
					ReferenceID:   "ref-existing-entity-mixed",
					Result: &amlcommon.ScreeningResult{
						ID:            "existing-txn-entity-mixed",
						CompletedAt:   "2025-08-20 12:32:39",
						TransactionID: "existing-txn-entity-mixed",
						Detail: []amlcommon.ScreeningDetailItem{
							{
								NodeDetail: amlcommon.NodeDetail{
									ProfileID: "e_tr_wci_entity_mixed",
									Name:      "PT ABC COMPANY",
								},
							},
						},
						Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
						MatchedCount: 1,
						Attributes: amlcommon.ScreeningAttributes{
							Name:              "PT ABC COMPANY",
							Score:             90,
							EntityType:        "ENTITY",
							ReferenceID:       "ref-existing-entity-mixed",
							CountryLocation:   "Indonesia",
							RegisteredCountry: "Indonesia",
							HitCategory:       []string{"LE"},
						},
					},
				}

				// Mix entity and person data
				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"PT ABC COMPANY": entityScreeningData,
						"John Doe:1990-01-01": {
							Status:        constant.AML_STATUS_APPROVED,
							TransactionID: "old-person-txn",
							ReferenceID:   "old-person-ref",
						},
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-mixed-entity").Return(&merchant.Merchant{
					UUID: "merchant-mixed-entity",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)

				// Mock ProfileDetail response
				amlRepo.On("ProfileDetail", mock.Anything, "existing-txn-entity-mixed", "e_tr_wci_entity_mixed").Return(&amlcommon.ProfileDetailResponse{
					Code:    "SUCCESS",
					Message: "SUCCESS",
					Data: amlcommon.ProfileDetailData{
						ProfileID: "e_tr_wci_entity_mixed",
						Name:      "PT ABC COMPANY",
					},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: ENTITY type profile with new screening (no merchant)",
			request: &amlcommon.CheckRequest{
				Name:        "PT NEW ENTITY",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			provider:   "advance_ai",
			merchantID: "",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// Since we're calling s.Screening(), we need to mock the underlying repo calls that Screening makes
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check and Inquiry responses that the Screening method will make
				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(&amlcommon.CheckResponse{
					TransactionID: "new-entity-txn-456",
				}, nil)

				amlRepo.On("Inquiry", mock.Anything, "new-entity-txn-456").Return(&amlcommon.InquiryResponse{
					TransactionID: "new-entity-txn-456",
					Data: amlcommon.InquiryResponseData{
						ID: "new-entity-txn-456",
						Nodes: []amlcommon.Node{
							{
								Type:        amlcommon.NodeTypeAMLNameScreening,
								Name:        "Aml Name Screening",
								CompletedAt: "2025-08-20 12:32:39",
								Attributes: map[string]any{
									"name":              "PT NEW ENTITY",
									"score":             88.0,
									"entityType":        "ENTITY",
									"referenceId":       "ref-new-entity-456",
									"countryLocation":   "Indonesia",
									"registeredCountry": "Indonesia",
									"hitCategory":       []any{"LE"},
								},
								Result: &amlcommon.NodeResult{
									Detail: []amlcommon.NodeDetail{
										{
											ProfileID: "e_tr_wci_new_entity_456",
											Name:      "PT NEW ENTITY",
										},
									},
									Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
									MatchedCount: 1,
								},
							},
						},
					},
				}, nil)

				// Mock ProfileDetail response
				amlRepo.On("ProfileDetail", mock.Anything, "new-entity-txn-456", "e_tr_wci_new_entity_456").Return(&amlcommon.ProfileDetailResponse{
					Code:    "SUCCESS",
					Message: "SUCCESS",
					Data: amlcommon.ProfileDetailData{
						ProfileID: "e_tr_wci_new_entity_456",
						Name:      "PT NEW ENTITY",
					},
				}, nil)
			},
			wantErr: false,
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
			result, err := service.Profile(context.Background(), tc.request, tc.provider, tc.merchantID, "")

			// Assertions
			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Equal(t, tc.expectedError.Error(), err.Error())
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "SUCCESS", result.Code)
				assert.NotEmpty(t, result.ReferenceID)
			}
		})
	}
}

func TestAmlService_findScreeningData(t *testing.T) {
	testCases := []struct {
		name       string
		request    *amlcommon.CheckRequest
		provider   string
		merchantID string
		mockSetup  func(
			mockRepo map[string]*mocks.IAmlProcessorRepository,
			mockMerchantRepo *mocks.IMerchantRepository,
		)
		expectedInquiryID string
		expectedProfileID string
		wantErr           bool
		expectedError     error
	}{
		{
			name: "SUCCESS: existing data found with profileID",
			request: &amlcommon.CheckRequest{
				Name: "Ir Joko WIDODO",
				DOB:  "1961-06-21",
			},
			provider:   "advance_ai",
			merchantID: "merchant-123",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// Create mock existing data with profile screening
				existingScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "existing-txn-123",
					ReferenceID:   "ref-existing-123",
					Result: &amlcommon.ScreeningResult{
						ID:            "existing-txn-123",
						CompletedAt:   "2025-08-20 12:32:39",
						TransactionID: "existing-txn-123",
						Detail: []amlcommon.ScreeningDetailItem{
							{
								NodeDetail: amlcommon.NodeDetail{
									ProfileID: "e_tr_wci_1224148",
									Name:      "Ir Joko WIDODO",
								},
							},
						},
						Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
						MatchedCount: 1,
						Attributes: amlcommon.ScreeningAttributes{
							DOB:               "1961-06-21",
							Name:              "Ir Joko WIDODO",
							Score:             95,
							Gender:            "MALE",
							EntityType:        "PERSON",
							ReferenceID:       "ref-existing-123",
							PlaceOfBirth:      "Indonesia",
							CountryLocation:   "Indonesia",
							RegisteredCountry: "Indonesia",
							HitCategory:       []string{"PEP"},
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"Ir Joko WIDODO:1961-06-21": existingScreeningData,
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-123").Return(&merchant.Merchant{
					UUID: "merchant-123",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)
			},
			expectedInquiryID: "existing-txn-123",
			expectedProfileID: "e_tr_wci_1224148",
			wantErr:           false,
		},
		{
			name: "SUCCESS: new screening without merchant",
			request: &amlcommon.CheckRequest{
				Name: "New User",
				DOB:  "1990-01-01",
			},
			provider:   "advance_ai",
			merchantID: "",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				amlRepo := mocks.NewIAmlProcessorRepository(t)
				mockRepo["advance_ai"] = amlRepo

				// Mock Check and Inquiry responses for new screening
				amlRepo.On("Check", mock.Anything, mock.AnythingOfType("*amlcommon.CheckRequest")).Return(&amlcommon.CheckResponse{
					TransactionID: "new-txn-456",
				}, nil)

				amlRepo.On("Inquiry", mock.Anything, "new-txn-456").Return(&amlcommon.InquiryResponse{
					TransactionID: "new-txn-456",
					Data: amlcommon.InquiryResponseData{
						ID: "new-txn-456",
						Nodes: []amlcommon.Node{
							{
								Type:        amlcommon.NodeTypeAMLNameScreening,
								Name:        "Aml Name Screening",
								CompletedAt: "2025-08-20 12:32:39",
								Attributes: map[string]any{
									"dob":               "1990-01-01",
									"name":              "New User",
									"score":             95.0,
									"gender":            "MALE",
									"entityType":        "PERSON",
									"referenceId":       "ref-new-999",
									"placeOfBirth":      "Indonesia",
									"countryLocation":   "Indonesia",
									"registeredCountry": "Indonesia",
									"hitCategory":       []any{"LE", "PEP"},
								},
								Result: &amlcommon.NodeResult{
									Detail: []amlcommon.NodeDetail{
										{
											ProfileID: "e_tr_wci_9999999",
											Name:      "New User",
										},
									},
									Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
									MatchedCount: 1,
								},
							},
						},
					},
				}, nil)
			},
			expectedInquiryID: "new-txn-456",
			expectedProfileID: "e_tr_wci_9999999",
			wantErr:           false,
		},
		{
			name: "FAILED: merchant not found",
			request: &amlcommon.CheckRequest{
				Name: "Test User",
				DOB:  "1990-01-01",
			},
			provider:   "advance_ai",
			merchantID: "nonexistent-merchant",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "nonexistent-merchant").Return(nil, errors.New("merchant not found"))
			},
			expectedInquiryID: "",
			expectedProfileID: "",
			wantErr:           true,
			expectedError:     constant.ErrMerchantNotFound,
		},
		{
			name: "FAILED: no data found for merchant with screening key",
			request: &amlcommon.CheckRequest{
				Name: "Unknown User",
				DOB:  "1995-01-01",
			},
			provider:   "advance_ai",
			merchantID: "merchant-empty",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-empty").Return(&merchant.Merchant{
					UUID: "merchant-empty",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: []byte(`{}`),
						Valid:    true,
					},
				}, nil)
			},
			expectedInquiryID: "",
			expectedProfileID: "",
			wantErr:           true,
			expectedError:     constant.ErrDataNotFound,
		},
		{
			name: "SUCCESS: existing data with empty profileID",
			request: &amlcommon.CheckRequest{
				Name: "No Profile User",
				DOB:  "1990-01-01",
			},
			provider:   "advance_ai",
			merchantID: "merchant-no-profile",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// Create mock existing data without AML Name Screening
				existingScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "existing-txn-no-profile",
					ReferenceID:   "ref-no-profile-2",
					Result: &amlcommon.ScreeningResult{
						ID:            "existing-txn-no-profile",
						CompletedAt:   "2025-08-20 12:32:39",
						TransactionID: "existing-txn-no-profile",
						Detail:        []amlcommon.ScreeningDetailItem{}, // Empty detail for no profile case
						Summary:       amlcommon.NodeSummary{Total: 0, FullMarksNum: 0},
						MatchedCount:  0,
						Attributes: amlcommon.ScreeningAttributes{
							DOB:         "1990-01-01",
							Name:        "No Profile User",
							ReferenceID: "ref-no-profile-2",
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"No Profile User:1990-01-01": existingScreeningData,
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-no-profile").Return(&merchant.Merchant{
					UUID: "merchant-no-profile",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)
			},
			expectedInquiryID: "existing-txn-no-profile",
			expectedProfileID: "",
			wantErr:           false,
		},
		{
			name: "SUCCESS: ENTITY type data found",
			request: &amlcommon.CheckRequest{
				Name:        "PT PAPER INDONESIA",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			provider:   "advance_ai",
			merchantID: "merchant-entity",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// Create mock existing ENTITY data
				existingScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "existing-entity-txn-123",
					ReferenceID:   "ref-existing-entity-123",
					Result: &amlcommon.ScreeningResult{
						ID:            "existing-entity-txn-123",
						CompletedAt:   "2025-08-20 12:32:39",
						TransactionID: "existing-entity-txn-123",
						Detail: []amlcommon.ScreeningDetailItem{
							{
								NodeDetail: amlcommon.NodeDetail{
									ProfileID: "e_tr_wci_entity_123",
									Name:      "PT PAPER INDONESIA",
								},
							},
						},
						Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
						MatchedCount: 1,
						Attributes: amlcommon.ScreeningAttributes{
							Name:              "PT PAPER INDONESIA",
							Score:             85,
							EntityType:        "ENTITY",
							ReferenceID:       "ref-existing-entity-123",
							CountryLocation:   "Indonesia",
							RegisteredCountry: "Indonesia",
							HitCategory:       []string{"SAN"},
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"PT PAPER INDONESIA": existingScreeningData, // Note: Entity key is name only
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-entity").Return(&merchant.Merchant{
					UUID: "merchant-entity",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)
			},
			expectedInquiryID: "existing-entity-txn-123",
			expectedProfileID: "e_tr_wci_entity_123",
			wantErr:           false,
		},
		{
			name: "SUCCESS: mixed data - find ENTITY in data with person entries",
			request: &amlcommon.CheckRequest{
				Name:        "PT ABC COMPANY",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			provider:   "advance_ai",
			merchantID: "merchant-mixed",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// Create mock mixed data with both entity and person screenings
				entityScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "entity-mixed-txn-456",
					ReferenceID:   "ref-entity-mixed-456",
					Result: &amlcommon.ScreeningResult{
						Detail: []amlcommon.ScreeningDetailItem{
							{
								NodeDetail: amlcommon.NodeDetail{
									ProfileID: "e_tr_wci_entity_456",
									Name:      "PT ABC COMPANY",
								},
							},
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"PT ABC COMPANY":    entityScreeningData, // Entity key
						"John Doe:1990-01-01": { // Person key
							Status:        constant.AML_STATUS_APPROVED,
							TransactionID: "person-txn",
							ReferenceID:   "person-ref",
						},
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-mixed").Return(&merchant.Merchant{
					UUID: "merchant-mixed",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)
			},
			expectedInquiryID: "entity-mixed-txn-456",
			expectedProfileID: "e_tr_wci_entity_456",
			wantErr:           false,
		},
		{
			name: "FAILED: ENTITY type data not found",
			request: &amlcommon.CheckRequest{
				Name:        "PT NONEXISTENT COMPANY",
				SubjectType: constant.AML_SUBJECT_TYPE_ENTITY,
			},
			provider:   "advance_ai",
			merchantID: "merchant-entity-empty",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// Only has person data, no entity data
				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"John Doe:1990-01-01": {
							Status:        constant.AML_STATUS_APPROVED,
							TransactionID: "person-only-txn",
							ReferenceID:   "person-only-ref",
						},
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-entity-empty").Return(&merchant.Merchant{
					UUID: "merchant-entity-empty",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)
			},
			expectedInquiryID: "",
			expectedProfileID: "",
			wantErr:           true,
			expectedError:     constant.ErrDataNotFound,
		},
		{
			name: "SUCCESS: backward compatibility - PERSON type still works",
			request: &amlcommon.CheckRequest{
				Name:        "Jane Smith",
				DOB:         "1985-05-15",
				SubjectType: constant.AML_SUBJECT_TYPE_PERSON,
			},
			provider:   "advance_ai",
			merchantID: "merchant-person-compat",
			mockSetup: func(
				mockRepo map[string]*mocks.IAmlProcessorRepository,
				mockMerchantRepo *mocks.IMerchantRepository,
			) {
				// Create mock existing PERSON data
				existingScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "person-compat-txn-789",
					ReferenceID:   "ref-person-compat-789",
					Result: &amlcommon.ScreeningResult{
						Detail: []amlcommon.ScreeningDetailItem{
							{
								NodeDetail: amlcommon.NodeDetail{
									ProfileID: "e_tr_wci_person_789",
									Name:      "Jane Smith",
								},
							},
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"Jane Smith:1985-05-15": existingScreeningData, // Person key with DOB
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-person-compat").Return(&merchant.Merchant{
					UUID: "merchant-person-compat",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingDataJSON,
						Valid:    true,
					},
				}, nil)
			},
			expectedInquiryID: "person-compat-txn-789",
			expectedProfileID: "e_tr_wci_person_789",
			wantErr:           false,
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
			inquiryID, profileID, err := service.findScreeningData(context.Background(), tc.request, tc.provider, tc.merchantID)

			// Assertions
			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.Equal(t, tc.expectedError.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedInquiryID, inquiryID)
			assert.Equal(t, tc.expectedProfileID, profileID)
		})
	}
}

func TestAmlService_extractProfileIDFromScreeningResponse(t *testing.T) {
	testCases := []struct {
		name              string
		screeningResponse *amlcommon.ScreeningResponse
		expectedProfileID string
	}{
		{
			name: "SUCCESS: extract profileID from screening response with result",
			screeningResponse: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-123",
				ReferenceID:   "ref-123",
				Result: &amlcommon.ScreeningResult{
					ID: "inquiry-123",
					Detail: []amlcommon.ScreeningDetailItem{
						{
							NodeDetail: amlcommon.NodeDetail{
								ProfileID: "e_tr_wci_1224148",
								Name:      "John Doe",
							},
						},
						{
							NodeDetail: amlcommon.NodeDetail{
								ProfileID: "e_tr_wci_9999999",
								Name:      "Jane Smith",
							},
						},
					},
				},
			},
			expectedProfileID: "e_tr_wci_1224148", // Should return first profileID
		},
		{
			name: "FAIL: no result in screening response",
			screeningResponse: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-456",
				ReferenceID:   "ref-456",
				Result:        nil,
			},
			expectedProfileID: "",
		},
		{
			name: "FAIL: empty detail array",
			screeningResponse: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-789",
				ReferenceID:   "ref-789",
				Result: &amlcommon.ScreeningResult{
					ID:     "inquiry-789",
					Detail: []amlcommon.ScreeningDetailItem{},
				},
			},
			expectedProfileID: "",
		},
		{
			name: "FAIL: first detail has empty profileID",
			screeningResponse: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-empty",
				ReferenceID:   "ref-empty",
				Result: &amlcommon.ScreeningResult{
					ID: "inquiry-empty",
					Detail: []amlcommon.ScreeningDetailItem{
						{
							NodeDetail: amlcommon.NodeDetail{
								ProfileID: "",
								Name:      "No Profile User",
							},
						},
					},
				},
			},
			expectedProfileID: "",
		},
		{
			name: "SUCCESS: extract non-empty profileID from first detail",
			screeningResponse: &amlcommon.ScreeningResponse{
				Status:        constant.AML_STATUS_APPROVED,
				TransactionID: "txn-single",
				ReferenceID:   "ref-single",
				Result: &amlcommon.ScreeningResult{
					ID: "inquiry-single",
					Detail: []amlcommon.ScreeningDetailItem{
						{
							NodeDetail: amlcommon.NodeDetail{
								ProfileID: "e_tr_wci_single_123",
								Name:      "Single User",
							},
						},
					},
				},
			},
			expectedProfileID: "e_tr_wci_single_123",
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
			profileID := service.extractProfileIDFromScreeningResponse(tc.screeningResponse)

			// Assertions
			assert.Equal(t, tc.expectedProfileID, profileID)
		})
	}
}
