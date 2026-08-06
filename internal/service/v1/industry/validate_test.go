package industry

import (
	"context"
	"errors"
	"testing"

	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIndustryService_IsValidMCC(t *testing.T) {
	tests := []struct {
		name           string
		mcc            string
		mockReturn     bool
		mockError      error
		expectedResult bool
		expectedError  error
	}{
		{
			name:           "Success - valid MCC",
			mcc:            "4511",
			mockReturn:     true,
			mockError:      nil,
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "Success - invalid MCC",
			mcc:            "9999",
			mockReturn:     false,
			mockError:      nil,
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "Error - repository error",
			mcc:            "4511",
			mockReturn:     false,
			mockError:      errors.New("database error"),
			expectedResult: false,
			expectedError:  errors.New("database error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			testLogger, _ := logger.NewZapLogger(logger.Config{})
			service := NewIndustryService(mockRepo, testLogger)

			mockRepo.On("IsValidMCC", mock.Anything, tc.mcc).Return(tc.mockReturn, tc.mockError)

			result, err := service.IsValidMCC(context.Background(), tc.mcc)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestIndustryService_ValidateIndustry(t *testing.T) {
	tests := []struct {
		name           string
		parentIndustry string
		childIndustry  string
		mockReturn     []string
		mockError      error
		expectedResult bool
		expectedError  error
	}{
		{
			name:           "Success - valid combination",
			parentIndustry: "Airlines",
			childIndustry:  "Airlines, Air Carriers",
			mockReturn:     []string{"Airlines, Air Carriers", "Charter Airlines"},
			mockError:      nil,
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name:           "Success - invalid combination",
			parentIndustry: "Airlines",
			childIndustry:  "NonExistent",
			mockReturn:     []string{"Airlines, Air Carriers", "Charter Airlines"},
			mockError:      nil,
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "Error - repository error",
			parentIndustry: "Airlines",
			childIndustry:  "Airlines, Air Carriers",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedResult: false,
			expectedError:  errors.New("database error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			testLogger, _ := logger.NewZapLogger(logger.Config{})
			service := NewIndustryService(mockRepo, testLogger)

			mockRepo.On("GetChildIndustries", mock.Anything, tc.parentIndustry).Return(tc.mockReturn, tc.mockError)

			result, err := service.ValidateIndustry(context.Background(), tc.parentIndustry, tc.childIndustry)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestIndustryService_ValidateIndustryMCCCombination(t *testing.T) {
	tests := []struct {
		name           string
		parentIndustry string
		childIndustry  string
		mcc            string
		mockReturn     string
		mockError      error
		expectedError  string
	}{
		{
			name:           "Success - valid combination",
			parentIndustry: "Airlines",
			childIndustry:  "Airlines, Air Carriers",
			mcc:            "4511",
			mockReturn:     "4511",
			mockError:      nil,
			expectedError:  "",
		},
		{
			name:           "Error - MCC mismatch",
			parentIndustry: "Airlines",
			childIndustry:  "Airlines, Air Carriers",
			mcc:            "5816",
			mockReturn:     "4511",
			mockError:      nil,
			expectedError:  "MCC 5816 does not match expected MCC 4511 for Airlines - Airlines, Air Carriers combination",
		},
		{
			name:           "Error - invalid industry combination",
			parentIndustry: "NonExistent",
			childIndustry:  "NonExistent",
			mcc:            "4511",
			mockReturn:     "",
			mockError:      nil,
			expectedError:  "invalid parent and child industry combination",
		},
		{
			name:           "Error - repository error",
			parentIndustry: "Airlines",
			childIndustry:  "Airlines, Air Carriers",
			mcc:            "4511",
			mockReturn:     "",
			mockError:      errors.New("database error"),
			expectedError:  "error retrieving MCC for industry combination: database error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			testLogger, _ := logger.NewZapLogger(logger.Config{})
			service := NewIndustryService(mockRepo, testLogger)

			mockRepo.On("GetMCCForIndustry", mock.Anything, tc.parentIndustry, tc.childIndustry).Return(tc.mockReturn, tc.mockError)

			err := service.ValidateIndustryMCCCombination(context.Background(), tc.parentIndustry, tc.childIndustry, tc.mcc)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
