package dukcapilservice

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	dukcapilmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/dukcapil"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockDukcapilGateway "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func TestVerifyIdentity(t *testing.T) {
	ctx := context.Background()

	expectedVerifyResult := &dukcapilmodel.VerifyResult{
		FullName:     "Sesuai (95)",
		Gender:       "Sesuai (100)",
		BirthDate:    "Sesuai (98)",
		BirthPlace:   "Sesuai (92)",
		Occupation:   "Sesuai",
		Address:      "Sesuai (90)",
		RT:           "Sesuai (88)",
		RW:           "Sesuai (85)",
		SubDistrict2: "Tidak Sesuai",
		SubDistrict:  "Sesuai (93)",
		District:     "Sesuai (96)",
		Province:     "Sesuai (97)",
		ResponseCode: "00",
		ResponseDesc: "Success",
	}

	expectedMerchant := &merchantModel.Merchant{
		UUID:                    "merchant-123",
		Name:                    "Test Merchant",
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		ThirdPartyScreeningData: types.NullJSONText{Valid: false},
	}

	testCases := []struct {
		Name             string
		IsSuccess        bool
		Request          *dukcapilmodel.IdentityVerificationRequest
		ExpectedError    string
		MockSetup        func(mockGateway *mockDukcapilGateway.IDukcapilGatewayRepository, mockMerchantRepo *mockMerchant.IMerchantRepository)
		ValidateResponse func(t *testing.T, response *dukcapilmodel.IdentityVerificationResponse)
	}{
		{
			Name:      "SUCCESS: verify identity with merchant storage",
			IsSuccess: true,
			Request: &dukcapilmodel.IdentityVerificationRequest{
				MerchantID: "merchant-123",
				VerifyRequest: &dukcapilmodel.VerifyRequest{
					NIK:      "1234567890123456",
					Name:     "John Doe",
					Gender:   "L",
					DOB:      "01-01-1990",
					POB:      "Jakarta",
					Job:      "Engineer",
					Address:  "Jl. Test No 1",
					RT:       "001",
					RW:       "002",
					Village:  "Kelurahan Test",
					District: "Kecamatan Test",
					Regency:  "Jakarta Selatan",
					Province: "DKI Jakarta",
				},
			},
			MockSetup: func(mockGateway *mockDukcapilGateway.IDukcapilGatewayRepository, mockMerchantRepo *mockMerchant.IMerchantRepository) {
				mockGateway.On(
					"VerifyIdentity",
					mock.Anything,
					mock.AnythingOfType("*dukcapilmodel.VerifyRequest"),
				).Return(expectedVerifyResult, nil)

				mockMerchantRepo.On(
					"FindMerchantByID",
					mock.Anything,
					"merchant-123",
				).Return(expectedMerchant, nil)

				mockMerchantRepo.On(
					"UpdateThirdPartyScreeningData",
					mock.Anything,
					"merchant-123",
					mock.AnythingOfType("types.NullJSONText"),
				).Return(nil)
			},
			ValidateResponse: func(t *testing.T, response *dukcapilmodel.IdentityVerificationResponse) {
				assert.Equal(t, dukcapilmodel.StatusNotMatched, response.Status) // Because name score (95) < threshold (100)
				assert.Len(t, response.FieldResults, 12)                         // All fields except NIK
				assert.NotEmpty(t, response.ReferenceID)

				// Check specific field results
				nameField := findFieldResult(response.FieldResults, dukcapilmodel.FieldName)
				assert.Equal(t, 95, nameField.Score)
				assert.Equal(t, 100, nameField.Threshold)
				assert.Equal(t, dukcapilmodel.StatusNotMatched, nameField.Status)

				genderField := findFieldResult(response.FieldResults, dukcapilmodel.FieldGender)
				assert.Equal(t, 100, genderField.Score)
				assert.Equal(t, 100, genderField.Threshold)
				assert.Equal(t, dukcapilmodel.StatusMatched, genderField.Status)

				villageField := findFieldResult(response.FieldResults, dukcapilmodel.FieldVillage)
				assert.Equal(t, 0, villageField.Score) // "Tidak Sesuai"
				assert.Equal(t, 95, villageField.Threshold)
				assert.Equal(t, dukcapilmodel.StatusNotMatched, villageField.Status)
			},
		},
		{
			Name:      "SUCCESS: verify identity without merchant (no storage)",
			IsSuccess: true,
			Request: &dukcapilmodel.IdentityVerificationRequest{
				MerchantID: "",
				VerifyRequest: &dukcapilmodel.VerifyRequest{
					NIK:  "1234567890123456",
					Name: "Jane Doe",
					DOB:  "01-01-1995",
				},
			},
			MockSetup: func(mockGateway *mockDukcapilGateway.IDukcapilGatewayRepository, mockMerchantRepo *mockMerchant.IMerchantRepository) {
				mockGateway.On(
					"VerifyIdentity",
					mock.Anything,
					mock.AnythingOfType("*dukcapilmodel.VerifyRequest"),
				).Return(expectedVerifyResult, nil)
				// No merchant repo calls expected when MerchantID is empty
			},
			ValidateResponse: func(t *testing.T, response *dukcapilmodel.IdentityVerificationResponse) {
				assert.NotEmpty(t, response.ReferenceID)
				assert.Len(t, response.FieldResults, 12)
			},
		},
		{
			Name:      "ERROR: Dukcapil gateway error",
			IsSuccess: false,
			Request: &dukcapilmodel.IdentityVerificationRequest{
				MerchantID: "merchant-123",
				VerifyRequest: &dukcapilmodel.VerifyRequest{
					NIK:  "1234567890123456",
					Name: "John Doe",
					DOB:  "01-01-1990",
				},
			},
			ExpectedError: "dukcapil service error",
			MockSetup: func(mockGateway *mockDukcapilGateway.IDukcapilGatewayRepository, mockMerchantRepo *mockMerchant.IMerchantRepository) {
				mockGateway.On(
					"VerifyIdentity",
					mock.Anything,
					mock.AnythingOfType("*dukcapilmodel.VerifyRequest"),
				).Return(nil, errors.New("dukcapil service error"))
			},
		},
		{
			Name:      "ERROR: merchant not found",
			IsSuccess: false,
			Request: &dukcapilmodel.IdentityVerificationRequest{
				MerchantID: "invalid-merchant",
				VerifyRequest: &dukcapilmodel.VerifyRequest{
					NIK:  "1234567890123456",
					Name: "John Doe",
					DOB:  "01-01-1990",
				},
			},
			ExpectedError: constant.ErrMerchantNotFound.Error(),
			MockSetup: func(mockGateway *mockDukcapilGateway.IDukcapilGatewayRepository, mockMerchantRepo *mockMerchant.IMerchantRepository) {
				mockGateway.On(
					"VerifyIdentity",
					mock.Anything,
					mock.AnythingOfType("*dukcapilmodel.VerifyRequest"),
				).Return(expectedVerifyResult, nil)

				mockMerchantRepo.On(
					"FindMerchantByID",
					mock.Anything,
					"invalid-merchant",
				).Return(nil, errors.New("merchant not found"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockGateway := mockDukcapilGateway.NewIDukcapilGatewayRepository(t)
			mockMerchantRepo := mockMerchant.NewIMerchantRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.MockSetup(mockGateway, mockMerchantRepo)

			cfg := &config.Config{
				Dukcapil: config.DukcapilConfig{
					FieldThresholds: config.DukcapilFieldThresholds{
						Name:     100,
						Gender:   100,
						DOB:      100,
						POB:      100,
						Job:      100,
						Address:  95,
						RT:       95,
						RW:       95,
						Village:  95,
						District: 95,
						Regency:  95,
						Province: 95,
					},
				},
			}

			svc := &DukcapilService{
				cfg:                 cfg,
				logger:              loggerMock,
				merchantRepository:  mockMerchantRepo,
				dukcapilGatewayRepo: mockGateway,
			}

			response, err := svc.VerifyIdentity(ctx, tc.Request)

			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotNil(t, response)
				if tc.ValidateResponse != nil {
					tc.ValidateResponse(t, response)
				}
			} else {
				require.Error(t, err)
				require.Nil(t, response)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			}

			mockGateway.AssertExpectations(t)
			mockMerchantRepo.AssertExpectations(t)
		})
	}
}

func TestValidateFields(t *testing.T) {
	cfg := &config.Config{
		Dukcapil: config.DukcapilConfig{
			FieldThresholds: config.DukcapilFieldThresholds{
				Name:     100,
				Gender:   100,
				DOB:      100,
				POB:      100,
				Job:      100,
				Address:  95,
				RT:       95,
				RW:       95,
				Village:  95,
				District: 95,
				Regency:  95,
				Province: 95,
			},
		},
	}

	svc := &DukcapilService{cfg: cfg}

	verifyResult := &dukcapilmodel.VerifyResult{
		FullName:     "Sesuai (95)",  // Below threshold (100)
		Gender:       "Sesuai",       // Default 100, meets threshold
		BirthDate:    "Sesuai (98)",  // Below threshold (100)
		BirthPlace:   "Sesuai (92)",  // Below threshold (100)
		Occupation:   "Sesuai",       // Default 100, meets threshold
		Address:      "Sesuai (96)",  // Above threshold (95)
		RT:           "Sesuai (88)",  // Below threshold (95)
		RW:           "Sesuai (85)",  // Below threshold (95)
		SubDistrict2: "Tidak Sesuai", // Below threshold (95)
		SubDistrict:  "Sesuai (93)",  // Below threshold (95)
		District:     "Sesuai (96)",  // Above threshold (95)
		Province:     "Sesuai (97)",  // Above threshold (95)
	}

	fieldResults := svc.validateFields(context.Background(), verifyResult)

	assert.Len(t, fieldResults, 12) // All 12 fields

	// Test fields with scores below thresholds (should be NOT_MATCHED)
	nameField := findFieldResult(fieldResults, dukcapilmodel.FieldName)
	assert.NotNil(t, nameField)
	assert.Equal(t, 95, nameField.Score)
	assert.Equal(t, 100, nameField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusNotMatched, nameField.Status)

	dobField := findFieldResult(fieldResults, dukcapilmodel.FieldDOB)
	assert.NotNil(t, dobField)
	assert.Equal(t, 98, dobField.Score)
	assert.Equal(t, 100, dobField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusNotMatched, dobField.Status)

	pobField := findFieldResult(fieldResults, dukcapilmodel.FieldPOB)
	assert.NotNil(t, pobField)
	assert.Equal(t, 92, pobField.Score)
	assert.Equal(t, 100, pobField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusNotMatched, pobField.Status)

	rtField := findFieldResult(fieldResults, dukcapilmodel.FieldRT)
	assert.NotNil(t, rtField)
	assert.Equal(t, 88, rtField.Score)
	assert.Equal(t, 95, rtField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusNotMatched, rtField.Status)

	rwField := findFieldResult(fieldResults, dukcapilmodel.FieldRW)
	assert.NotNil(t, rwField)
	assert.Equal(t, 85, rwField.Score)
	assert.Equal(t, 95, rwField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusNotMatched, rwField.Status)

	districtField := findFieldResult(fieldResults, dukcapilmodel.FieldDistrict)
	assert.NotNil(t, districtField)
	assert.Equal(t, 93, districtField.Score)
	assert.Equal(t, 95, districtField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusNotMatched, districtField.Status)

	// Test fields that meet thresholds (should be MATCHED)
	genderField := findFieldResult(fieldResults, dukcapilmodel.FieldGender)
	assert.NotNil(t, genderField)
	assert.Equal(t, 100, genderField.Score)
	assert.Equal(t, 100, genderField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusMatched, genderField.Status)

	jobField := findFieldResult(fieldResults, dukcapilmodel.FieldJob)
	assert.NotNil(t, jobField)
	assert.Equal(t, 100, jobField.Score)
	assert.Equal(t, 100, jobField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusMatched, jobField.Status)

	addressField := findFieldResult(fieldResults, dukcapilmodel.FieldAddress)
	assert.NotNil(t, addressField)
	assert.Equal(t, 96, addressField.Score)
	assert.Equal(t, 95, addressField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusMatched, addressField.Status)

	regencyField := findFieldResult(fieldResults, dukcapilmodel.FieldRegency)
	assert.NotNil(t, regencyField)
	assert.Equal(t, 96, regencyField.Score)
	assert.Equal(t, 95, regencyField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusMatched, regencyField.Status)

	provinceField := findFieldResult(fieldResults, dukcapilmodel.FieldProvince)
	assert.NotNil(t, provinceField)
	assert.Equal(t, 97, provinceField.Score)
	assert.Equal(t, 95, provinceField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusMatched, provinceField.Status)

	// Test field that's explicitly "Tidak Sesuai" (should be 0 score)
	villageField := findFieldResult(fieldResults, dukcapilmodel.FieldVillage)
	assert.NotNil(t, villageField)
	assert.Equal(t, 0, villageField.Score)
	assert.Equal(t, 95, villageField.Threshold)
	assert.Equal(t, dukcapilmodel.StatusNotMatched, villageField.Status)
}

func TestParseDukcapilFieldScore(t *testing.T) {
	testCases := []struct {
		Name          string
		ResponseValue string
		ExpectedScore int
	}{
		{
			Name:          "Sesuai with score in parentheses",
			ResponseValue: "Sesuai (95)",
			ExpectedScore: 95,
		},
		{
			Name:          "Sesuai with score 100 in parentheses",
			ResponseValue: "Sesuai (100)",
			ExpectedScore: 100,
		},
		{
			Name:          "Sesuai without score (default 100)",
			ResponseValue: "Sesuai",
			ExpectedScore: 100,
		},
		{
			Name:          "Tidak Sesuai (always 0)",
			ResponseValue: "Tidak Sesuai",
			ExpectedScore: 0,
		},
		{
			Name:          "Empty string",
			ResponseValue: "",
			ExpectedScore: 0,
		},
		{
			Name:          "Random string",
			ResponseValue: "Random Text",
			ExpectedScore: 0,
		},
		{
			Name:          "Case insensitive Tidak Sesuai",
			ResponseValue: "tidak sesuai",
			ExpectedScore: 0,
		},
		{
			Name:          "Case insensitive Sesuai",
			ResponseValue: "sesuai",
			ExpectedScore: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			score := dukcapilmodel.ParseDukcapilFieldScore(tc.ResponseValue)
			assert.Equal(t, tc.ExpectedScore, score)
		})
	}
}

func TestOrderedFieldMappings(t *testing.T) {
	verifyResult := &dukcapilmodel.VerifyResult{
		FullName:     "John Doe",
		Gender:       "Male",
		BirthDate:    "01-01-1990",
		BirthPlace:   "Jakarta",
		Occupation:   "Engineer",
		Address:      "Jl. Test",
		RT:           "001",
		RW:           "002",
		SubDistrict2: "Village",
		SubDistrict:  "District",
		District:     "Regency",
		Province:     "Province",
	}

	fieldMappings := dukcapilmodel.NewDukcapilFieldMappings(verifyResult)

	// Test that fields are returned in the expected order
	expectedOrder := []string{
		dukcapilmodel.DukcapilFieldName,     // 0
		dukcapilmodel.DukcapilFieldGender,   // 1
		dukcapilmodel.DukcapilFieldDOB,      // 2
		dukcapilmodel.DukcapilFieldPOB,      // 3
		dukcapilmodel.DukcapilFieldJob,      // 4
		dukcapilmodel.DukcapilFieldAddress,  // 5
		dukcapilmodel.DukcapilFieldRT,       // 6
		dukcapilmodel.DukcapilFieldRW,       // 7
		dukcapilmodel.DukcapilFieldVillage,  // 8
		dukcapilmodel.DukcapilFieldDistrict, // 9
		dukcapilmodel.DukcapilFieldRegency,  // 10
		dukcapilmodel.DukcapilFieldProvince, // 11
	}

	require.Len(t, fieldMappings.Fields, 12)

	for i, expectedField := range expectedOrder {
		actualField := fieldMappings.Fields[i].DukcapilField
		assert.Equal(t, expectedField, actualField, "Field at index %d should be %s but got %s", i, expectedField, actualField)
	}

	// Test that values are correctly mapped
	assert.Equal(t, "John Doe", fieldMappings.Fields[0].Value)   // Name
	assert.Equal(t, "Male", fieldMappings.Fields[1].Value)       // Gender
	assert.Equal(t, "01-01-1990", fieldMappings.Fields[2].Value) // DOB
	assert.Equal(t, "Jakarta", fieldMappings.Fields[3].Value)    // POB
	assert.Equal(t, "Engineer", fieldMappings.Fields[4].Value)   // Job
}

func TestDetermineOverallStatus(t *testing.T) {
	svc := &DukcapilService{}

	testCases := []struct {
		Name           string
		FieldResults   []dukcapilmodel.DukcapilFieldResult
		ExpectedStatus string
	}{
		{
			Name: "All fields matched",
			FieldResults: []dukcapilmodel.DukcapilFieldResult{
				{Field: "name", Status: dukcapilmodel.StatusMatched},
				{Field: "gender", Status: dukcapilmodel.StatusMatched},
			},
			ExpectedStatus: dukcapilmodel.StatusMatched,
		},
		{
			Name: "One field not matched",
			FieldResults: []dukcapilmodel.DukcapilFieldResult{
				{Field: "name", Status: dukcapilmodel.StatusMatched},
				{Field: "gender", Status: dukcapilmodel.StatusNotMatched},
			},
			ExpectedStatus: dukcapilmodel.StatusNotMatched,
		},
		{
			Name:           "No fields",
			FieldResults:   []dukcapilmodel.DukcapilFieldResult{},
			ExpectedStatus: dukcapilmodel.StatusMatched,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			status := svc.determineOverallStatus(tc.FieldResults)
			assert.Equal(t, tc.ExpectedStatus, status)
		})
	}
}

func TestStoreVerificationResult(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		Name          string
		IsSuccess     bool
		MerchantID    string
		Request       *dukcapilmodel.VerifyRequest
		Response      *dukcapilmodel.IdentityVerificationResponse
		ExpectedError string
		MockSetup     func(mockMerchantRepo *mockMerchant.IMerchantRepository)
	}{
		{
			Name:       "SUCCESS: store new screening data",
			IsSuccess:  true,
			MerchantID: "merchant-123",
			Request: &dukcapilmodel.VerifyRequest{
				Name: "John Doe",
				DOB:  "01-01-1990",
			},
			Response: &dukcapilmodel.IdentityVerificationResponse{
				Status: dukcapilmodel.StatusMatched,
			},
			MockSetup: func(mockMerchantRepo *mockMerchant.IMerchantRepository) {
				merchant := &merchantModel.Merchant{
					UUID:                    "merchant-123",
					ThirdPartyScreeningData: types.NullJSONText{Valid: false},
				}

				mockMerchantRepo.On(
					"FindMerchantByID",
					mock.Anything,
					"merchant-123",
				).Return(merchant, nil)

				mockMerchantRepo.On(
					"UpdateThirdPartyScreeningData",
					mock.Anything,
					"merchant-123",
					mock.AnythingOfType("types.NullJSONText"),
				).Return(nil)
			},
		},
		{
			Name:       "SUCCESS: update existing screening data",
			IsSuccess:  true,
			MerchantID: "merchant-456",
			Request: &dukcapilmodel.VerifyRequest{
				Name: "Jane Doe",
				DOB:  "01-01-1995",
			},
			Response: &dukcapilmodel.IdentityVerificationResponse{
				Status: dukcapilmodel.StatusNotMatched,
			},
			MockSetup: func(mockMerchantRepo *mockMerchant.IMerchantRepository) {
				existingData := commonModel.ThirdPartyScreeningData{
					Dukcapil: map[string]*dukcapilmodel.DukcapilScreeningData{
						"John Doe:01-01-1990": {
							Status: dukcapilmodel.StatusMatched,
						},
					},
				}
				existingJSON, _ := json.Marshal(existingData)

				merchant := &merchantModel.Merchant{
					UUID: "merchant-456",
					ThirdPartyScreeningData: types.NullJSONText{
						JSONText: existingJSON,
						Valid:    true,
					},
				}

				mockMerchantRepo.On(
					"FindMerchantByID",
					mock.Anything,
					"merchant-456",
				).Return(merchant, nil)

				mockMerchantRepo.On(
					"UpdateThirdPartyScreeningData",
					mock.Anything,
					"merchant-456",
					mock.AnythingOfType("types.NullJSONText"),
				).Return(nil)
			},
		},
		{
			Name:       "ERROR: missing full name",
			IsSuccess:  false,
			MerchantID: "merchant-123",
			Request: &dukcapilmodel.VerifyRequest{
				Name: "",
				DOB:  "01-01-1990",
			},
			Response:      &dukcapilmodel.IdentityVerificationResponse{},
			ExpectedError: constant.ErrInvalidRequestPayload.Error(),
			MockSetup: func(mockMerchantRepo *mockMerchant.IMerchantRepository) {
				// No mock setup needed as validation fails early
			},
		},
		{
			Name:       "ERROR: missing birth date",
			IsSuccess:  false,
			MerchantID: "merchant-123",
			Request: &dukcapilmodel.VerifyRequest{
				Name: "John Doe",
				DOB:  "",
			},
			Response:      &dukcapilmodel.IdentityVerificationResponse{},
			ExpectedError: constant.ErrInvalidRequestPayload.Error(),
			MockSetup: func(mockMerchantRepo *mockMerchant.IMerchantRepository) {
				// No mock setup needed as validation fails early
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockMerchantRepo := mockMerchant.NewIMerchantRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.MockSetup(mockMerchantRepo)

			svc := &DukcapilService{
				logger:             loggerMock,
				merchantRepository: mockMerchantRepo,
			}

			err := svc.storeVerificationResult(ctx, tc.MerchantID, tc.Response, tc.Request, &dukcapilmodel.VerifyResult{})

			if tc.IsSuccess {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			}

			mockMerchantRepo.AssertExpectations(t)
		})
	}
}

// Helper function to find a field result by field name
func findFieldResult(fieldResults []dukcapilmodel.DukcapilFieldResult, fieldName string) *dukcapilmodel.DukcapilFieldResult {
	for _, result := range fieldResults {
		if result.Field == fieldName {
			return &result
		}
	}
	return nil
}
