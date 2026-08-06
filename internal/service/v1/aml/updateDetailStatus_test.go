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
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAmlService_UpdateDetailStatusByProfileId(t *testing.T) {
	testCases := []struct {
		name          string
		profileID     string
		merchantID    string
		request       *amlcommon.UpdateDetailStatusRequest
		mockSetup     func(mockMerchantRepo *mocks.IMerchantRepository)
		wantErr       bool
		expectedError error
	}{
		{
			name:       "SUCCESS: update status to DISMISS",
			profileID:  "e_tr_wci_1224148",
			merchantID: "merchant-123",
			request: &amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
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
						Summary:      amlcommon.NodeSummary{Total: 2, FullMarksNum: 2},
						MatchedCount: 2,
						Attributes: amlcommon.ScreeningAttributes{
							DOB:         "1990-01-01",
							Name:        "John Doe",
							Score:       95,
							ReferenceID: "ref-existing-123",
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"John Doe:1990-01-01": existingScreeningData,
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

				mockMerchantRepo.On("UpdateThirdPartyScreeningData", mock.Anything, "merchant-123", mock.AnythingOfType("types.NullJSONText")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:       "SUCCESS: update status to ON_MONITOR",
			profileID:  "e_tr_wci_1224148",
			merchantID: "merchant-456",
			request: &amlcommon.UpdateDetailStatusRequest{
				Name:   "Jane Smith",
				DOB:    "1985-05-15",
				Status: amlcommon.DetailStatusOnMonitor,
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				existingScreeningData := &amlcommon.ScreeningResponse{
					Status:        constant.AML_STATUS_APPROVED,
					TransactionID: "existing-txn-456",
					ReferenceID:   "ref-existing-456",
					Result: &amlcommon.ScreeningResult{
						ID:            "existing-txn-456",
						CompletedAt:   "2025-08-20 12:32:39",
						TransactionID: "existing-txn-456",
						Detail: []amlcommon.ScreeningDetailItem{
							{
								NodeDetail: amlcommon.NodeDetail{
									ProfileID: "e_tr_wci_1224148",
									Name:      "Jane Smith",
								},
							},
						},
						Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
						MatchedCount: 1,
						Attributes: amlcommon.ScreeningAttributes{
							DOB:         "1985-05-15",
							Name:        "Jane Smith",
							Score:       95,
							ReferenceID: "ref-existing-456",
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"Jane Smith:1985-05-15": existingScreeningData,
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
			wantErr: false,
		},
		{
			name:       "FAILED: missing name",
			profileID:  "e_tr_wci_1224148",
			merchantID: "merchant-123",
			request: &amlcommon.UpdateDetailStatusRequest{
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				// No setup needed since validation should fail
			},
			wantErr:       true,
			expectedError: constant.ErrInvalidRequestPayload,
		},
		{
			name:       "FAILED: missing dob",
			profileID:  "e_tr_wci_1224148",
			merchantID: "merchant-123",
			request: &amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				// No setup needed since validation should fail
			},
			wantErr:       true,
			expectedError: constant.ErrInvalidRequestPayload,
		},
		{
			name:       "FAILED: invalid status",
			profileID:  "e_tr_wci_1224148",
			merchantID: "merchant-123",
			request: &amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: "INVALID_STATUS",
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				// No setup needed since validation should fail
			},
			wantErr:       true,
			expectedError: constant.ErrInvalidRequestPayload,
		},
		{
			name:       "FAILED: merchant not found",
			profileID:  "e_tr_wci_1224148",
			merchantID: "nonexistent-merchant",
			request: &amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "nonexistent-merchant").Return(nil, errors.New("merchant not found"))
			},
			wantErr:       true,
			expectedError: constant.ErrMerchantNotFound,
		},
		{
			name:       "FAILED: no screening data for merchant",
			profileID:  "e_tr_wci_1224148",
			merchantID: "merchant-empty",
			request: &amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-empty").Return(&merchant.Merchant{
					UUID: "merchant-empty",
					ThirdPartyScreeningData: types.NullJSONText{
						Valid: false,
					},
				}, nil)
			},
			wantErr:       true,
			expectedError: constant.ErrDataNotFound,
		},
		{
			name:       "FAILED: no screening data for person",
			profileID:  "e_tr_wci_1224148",
			merchantID: "merchant-no-person",
			request: &amlcommon.UpdateDetailStatusRequest{
				Name:   "Unknown Person",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"John Doe:1990-01-01": {
							Status:        constant.AML_STATUS_APPROVED,
							TransactionID: "existing-txn",
							ReferenceID:   "ref-existing",
						},
					},
				}
				existingDataJSON, _ := json.Marshal(existingData)

				mockMerchantRepo.On("FindMerchantByID", mock.Anything, "merchant-no-person").Return(&merchant.Merchant{
					UUID: "merchant-no-person",
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
			name:       "FAILED: profile ID not found in detail",
			profileID:  "e_tr_wci_nonexistent",
			merchantID: "merchant-123",
			request: &amlcommon.UpdateDetailStatusRequest{
				Name:   "John Doe",
				DOB:    "1990-01-01",
				Status: amlcommon.DetailStatusDismiss,
			},
			mockSetup: func(mockMerchantRepo *mocks.IMerchantRepository) {
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
									ProfileID: "e_tr_wci_different",
									Name:      "John Doe",
								},
							},
						},
						Summary:      amlcommon.NodeSummary{Total: 1, FullMarksNum: 1},
						MatchedCount: 1,
						Attributes: amlcommon.ScreeningAttributes{
							DOB:         "1990-01-01",
							Name:        "John Doe",
							Score:       95,
							ReferenceID: "ref-existing-123",
						},
					},
				}

				existingData := commonModel.ThirdPartyScreeningData{
					AML: map[string]*amlcommon.ScreeningResponse{
						"John Doe:1990-01-01": existingScreeningData,
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
			wantErr:       true,
			expectedError: constant.ErrDataNotFound,
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
			err := service.UpdateDetailStatusByProfileId(context.Background(), tc.profileID, tc.merchantID, tc.request)

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
