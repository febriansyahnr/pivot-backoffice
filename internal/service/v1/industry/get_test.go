package industry

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIndustryService_GetAllIndustries(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     []*industryModel.Industry
		mockError      error
		expectedResult []*industryModel.Industry
		expectedError  error
	}{
		{
			name: "Success - returns all industries",
			mockReturn: []*industryModel.Industry{
				{ParentIndustry: "Airlines", ChildIndustry: "Airlines, Air Carriers", CommonMCC: "4511"},
				{ParentIndustry: "Digital goods", ChildIndustry: "Games", CommonMCC: "5816"},
			},
			mockError: nil,
			expectedResult: []*industryModel.Industry{
				{ParentIndustry: "Airlines", ChildIndustry: "Airlines, Air Carriers", CommonMCC: "4511"},
				{ParentIndustry: "Digital goods", ChildIndustry: "Games", CommonMCC: "5816"},
			},
			expectedError: nil,
		},
		{
			name:           "Error - repository error",
			mockReturn:     []*industryModel.Industry{},
			mockError:      errors.New("database error"),
			expectedResult: nil,
			expectedError:  pkgErrs.New(response.HttpErrDatabase, constant.ErrGetAllCountries),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			testLogger, _ := logger.NewZapLogger(logger.Config{})
			service := NewIndustryService(mockRepo, testLogger)

			mockRepo.On("GetAllIndustries", mock.Anything, mock.Anything).Return(tc.mockReturn, tc.mockError)

			result, err := service.GetAllIndustries(context.Background(), &industryModel.SearchIndustryRequest{})

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestIndustryService_GetUniqueParentIndustries(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     []string
		mockError      error
		expectedResult []string
		expectedError  error
	}{
		{
			name:           "Success - returns parent industries",
			mockReturn:     []string{"Airlines", "Digital goods", "Retail"},
			mockError:      nil,
			expectedResult: []string{"Airlines", "Digital goods", "Retail"},
			expectedError:  nil,
		},
		{
			name:           "Error - repository error",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedResult: nil,
			expectedError:  errors.New("database error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			testLogger, _ := logger.NewZapLogger(logger.Config{})
			service := NewIndustryService(mockRepo, testLogger)

			mockRepo.On("GetUniqueParentIndustries", mock.Anything).Return(tc.mockReturn, tc.mockError)

			result, err := service.GetUniqueParentIndustries(context.Background())

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestIndustryService_GetChildIndustries(t *testing.T) {
	tests := []struct {
		name           string
		parentIndustry string
		mockReturn     []string
		mockError      error
		expectedResult []string
		expectedError  error
	}{
		{
			name:           "Success - returns child industries for Airlines",
			parentIndustry: "Airlines",
			mockReturn:     []string{"Airlines, Air Carriers", "Charter Airlines"},
			mockError:      nil,
			expectedResult: []string{"Airlines, Air Carriers", "Charter Airlines"},
			expectedError:  nil,
		},
		{
			name:           "Success - empty result for invalid parent",
			parentIndustry: "NonExistent",
			mockReturn:     []string{},
			mockError:      nil,
			expectedResult: []string{},
			expectedError:  nil,
		},
		{
			name:           "Error - repository error",
			parentIndustry: "Airlines",
			mockReturn:     nil,
			mockError:      errors.New("database error"),
			expectedResult: nil,
			expectedError:  errors.New("database error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			testLogger, _ := logger.NewZapLogger(logger.Config{})
			service := NewIndustryService(mockRepo, testLogger)

			mockRepo.On("GetChildIndustries", mock.Anything, tc.parentIndustry).Return(tc.mockReturn, tc.mockError)

			result, err := service.GetChildIndustries(context.Background(), tc.parentIndustry)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestIndustryService_GetMCCForIndustry(t *testing.T) {
	tests := []struct {
		name           string
		parentIndustry string
		childIndustry  string
		mockReturn     string
		mockError      error
		expectedResult string
		expectedError  error
	}{
		{
			name:           "Success - returns MCC for valid combination",
			parentIndustry: "Airlines",
			childIndustry:  "Airlines, Air Carriers",
			mockReturn:     "4511",
			mockError:      nil,
			expectedResult: "4511",
			expectedError:  nil,
		},
		{
			name:           "Success - empty result for invalid combination",
			parentIndustry: "NonExistent",
			childIndustry:  "NonExistent",
			mockReturn:     "",
			mockError:      nil,
			expectedResult: "",
			expectedError:  nil,
		},
		{
			name:           "Error - repository error",
			parentIndustry: "Airlines",
			childIndustry:  "Airlines, Air Carriers",
			mockReturn:     "",
			mockError:      errors.New("database error"),
			expectedResult: "",
			expectedError:  errors.New("database error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			testLogger, _ := logger.NewZapLogger(logger.Config{})
			service := NewIndustryService(mockRepo, testLogger)

			mockRepo.On("GetMCCForIndustry", mock.Anything, tc.parentIndustry, tc.childIndustry).Return(tc.mockReturn, tc.mockError)

			result, err := service.GetMCCForIndustry(context.Background(), tc.parentIndustry, tc.childIndustry)

			assert.Equal(t, tc.expectedResult, result)
			assert.Equal(t, tc.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}
