package industry

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	industryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/industry"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIndustryService_CreateIndustry(t *testing.T) {
	testLogger, _ := logger.NewZapLogger(logger.Config{})

	now := time.Now()
	validIndustry := &industryModel.Industry{
		UUID:           "test-uuid",
		ParentIndustry: "Technology",
		ChildIndustry:  "Software",
		RiskLevel:      "Low",
		MCC:            "5734",
		CommonMCC:      "5734",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	tests := []struct {
		name          string
		request       industryModel.CreateIndustryRequest
		mockSetup     func(*mocks.IIndustryRepository)
		expected      *industryModel.Industry
		expectedError error
	}{
		{
			name: "SUCCESS: Create industry with valid request",
			request: industryModel.CreateIndustryRequest{
				ParentIndustry: "Technology",
				ChildIndustry:  "Software",
				RiskLevel:      "Low",
				MCC:            "5734",
				CommonMCC:      "5734",
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetByParentChildIndustry", mock.Anything, "Technology", "Software").
					Return(nil, nil)
				m.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expected:      validIndustry,
			expectedError: nil,
		},
		{
			name: "ERROR: Validation fails - missing parent industry",
			request: industryModel.CreateIndustryRequest{
				ParentIndustry: "",
				ChildIndustry:  "Software",
				RiskLevel:      "Low",
				MCC:            "5734",
				CommonMCC:      "5734",
			},
			mockSetup:     func(m *mocks.IIndustryRepository) {},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrParentIndustryRequired),
		},
		{
			name: "ERROR: Validation fails - missing child industry",
			request: industryModel.CreateIndustryRequest{
				ParentIndustry: "Technology",
				ChildIndustry:  "",
				RiskLevel:      "Low",
				MCC:            "5734",
				CommonMCC:      "5734",
			},
			mockSetup:     func(m *mocks.IIndustryRepository) {},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrChildIndustryRequired),
		},
		{
			name: "ERROR: Validation fails - invalid risk level",
			request: industryModel.CreateIndustryRequest{
				ParentIndustry: "Technology",
				ChildIndustry:  "Software",
				RiskLevel:      "Invalid",
				MCC:            "5734",
				CommonMCC:      "5734",
			},
			mockSetup:     func(m *mocks.IIndustryRepository) {},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidIndustryRisk),
		},
		{
			name: "ERROR: Validation fails - missing MCC",
			request: industryModel.CreateIndustryRequest{
				ParentIndustry: "Technology",
				ChildIndustry:  "Software",
				RiskLevel:      "Low",
				MCC:            "",
				CommonMCC:      "5734",
			},
			mockSetup:     func(m *mocks.IIndustryRepository) {},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrMCCRequired),
		},
		{
			name: "ERROR: Validation fails - missing CommonMCC",
			request: industryModel.CreateIndustryRequest{
				ParentIndustry: "Technology",
				ChildIndustry:  "Software",
				RiskLevel:      "Low",
				MCC:            "5734",
				CommonMCC:      "",
			},
			mockSetup:     func(m *mocks.IIndustryRepository) {},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrCommonMCCRequired),
		},
		{
			name: "ERROR: Duplicate industry exists",
			request: industryModel.CreateIndustryRequest{
				ParentIndustry: "Technology",
				ChildIndustry:  "Software",
				RiskLevel:      "Low",
				MCC:            "5734",
				CommonMCC:      "5734",
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetByParentChildIndustry", mock.Anything, "Technology", "Software").
					Return(&industryModel.Industry{UUID: "existing-uuid"}, nil)
			},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrDuplicateIndustry),
		},
		{
			name: "ERROR: Database error on checking duplicate",
			request: industryModel.CreateIndustryRequest{
				ParentIndustry: "Technology",
				ChildIndustry:  "Software",
				RiskLevel:      "Low",
				MCC:            "5734",
				CommonMCC:      "5734",
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetByParentChildIndustry", mock.Anything, "Technology", "Software").
					Return(nil, assert.AnError)
			},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrDatabase, constant.ErrCreateIndustry),
		},
		{
			name: "ERROR: Database error on create",
			request: industryModel.CreateIndustryRequest{
				ParentIndustry: "Technology",
				ChildIndustry:  "Software",
				RiskLevel:      "Low",
				MCC:            "5734",
				CommonMCC:      "5734",
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetByParentChildIndustry", mock.Anything, "Technology", "Software").
					Return(nil, nil)
				m.On("Create", mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrDatabase, constant.ErrCreateIndustry),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			tc.mockSetup(mockRepo)
			service := NewIndustryService(mockRepo, testLogger)

			result, err := service.CreateIndustry(context.Background(), tc.request)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.request.ParentIndustry, result.ParentIndustry)
				assert.Equal(t, tc.request.ChildIndustry, result.ChildIndustry)
				assert.Equal(t, tc.request.RiskLevel, result.RiskLevel)
				assert.Equal(t, tc.request.MCC, result.MCC)
				assert.Equal(t, tc.request.CommonMCC, result.CommonMCC)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
