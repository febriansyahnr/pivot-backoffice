package industry

import (
	"context"
	"errors"
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

func TestIndustryService_UpdateIndustry(t *testing.T) {
	testLogger, _ := logger.NewZapLogger(logger.Config{})

	now := time.Now()
	existingIndustry := &industryModel.Industry{
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
		request       industryModel.UpdateIndustryRequest
		mockSetup     func(*mocks.IIndustryRepository)
		expected      *industryModel.Industry
		expectedError error
	}{
		{
			name: "SUCCESS: Update risk level only",
			request: industryModel.UpdateIndustryRequest{
				UUID:      "test-uuid",
				RiskLevel: ptr("Medium"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expected: &industryModel.Industry{
				UUID:           "test-uuid",
				ParentIndustry: "Technology",
				ChildIndustry:  "Software",
				RiskLevel:      "Medium",
				MCC:            "5734",
				CommonMCC:      "5734",
			},
			expectedError: nil,
		},
		{
			name: "SUCCESS: Update all fields",
			request: industryModel.UpdateIndustryRequest{
				UUID:           "test-uuid",
				ParentIndustry: ptr("Finance"),
				ChildIndustry:  ptr("Banking"),
				RiskLevel:      ptr("High"),
				MCC:            ptr("9999"),
				CommonMCC:      ptr("9999"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("GetByParentChildIndustry", mock.Anything, "Finance", "Banking").
					Return(nil, nil)
				m.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expected: &industryModel.Industry{
				UUID:           "test-uuid",
				ParentIndustry: "Finance",
				ChildIndustry:  "Banking",
				RiskLevel:      "High",
				MCC:            "9999",
				CommonMCC:      "9999",
			},
			expectedError: nil,
		},
		{
			name: "SUCCESS: Update parent industry only",
			request: industryModel.UpdateIndustryRequest{
				UUID:           "test-uuid",
				ParentIndustry: ptr("Retail"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("GetByParentChildIndustry", mock.Anything, "Retail", "Software").
					Return(nil, nil)
				m.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expected: &industryModel.Industry{
				UUID:           "test-uuid",
				ParentIndustry: "Retail",
				ChildIndustry:  "Software",
				RiskLevel:      "Low",
				MCC:            "5734",
				CommonMCC:      "5734",
			},
			expectedError: nil,
		},
		{
			name: "ERROR: Industry not found",
			request: industryModel.UpdateIndustryRequest{
				UUID:      "non-existent-uuid",
				RiskLevel: ptr("Medium"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "non-existent-uuid").
					Return(nil, nil)
			},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrIndustryNotFound),
		},
		{
			name: "ERROR: Duplicate parent-child combination",
			request: industryModel.UpdateIndustryRequest{
				UUID:           "test-uuid",
				ParentIndustry: ptr("Finance"),
				ChildIndustry:  ptr("Banking"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("GetByParentChildIndustry", mock.Anything, "Finance", "Banking").
					Return(&industryModel.Industry{UUID: "other-uuid"}, nil)
			},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrDuplicateIndustry),
		},
		{
			name: "ERROR: Same industry returned (no duplicate)",
			request: industryModel.UpdateIndustryRequest{
				UUID:           "test-uuid",
				ParentIndustry: ptr("Technology"),
				ChildIndustry:  ptr("Software"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("GetByParentChildIndustry", mock.Anything, "Technology", "Software").
					Return(existingIndustry, nil)
				m.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expected:      existingIndustry,
			expectedError: nil,
		},
		{
			name: "ERROR: Invalid risk level",
			request: industryModel.UpdateIndustryRequest{
				UUID:      "test-uuid",
				RiskLevel: ptr("Invalid"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
			},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidIndustryRisk),
		},
		{
			name: "ERROR: Database error on get by ID",
			request: industryModel.UpdateIndustryRequest{
				UUID:      "test-uuid",
				RiskLevel: ptr("Medium"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(nil, errors.New("database error"))
			},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrDatabase, constant.ErrUpdateIndustry),
		},
		{
			name: "ERROR: Database error on check duplicate",
			request: industryModel.UpdateIndustryRequest{
				UUID:           "test-uuid",
				ParentIndustry: ptr("Finance"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("GetByParentChildIndustry", mock.Anything, "Finance", "Software").
					Return(nil, errors.New("database error"))
			},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrDatabase, constant.ErrUpdateIndustry),
		},
		{
			name: "ERROR: Database error on update",
			request: industryModel.UpdateIndustryRequest{
				UUID:      "test-uuid",
				RiskLevel: ptr("Medium"),
			},
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("Update", mock.Anything, mock.Anything).
					Return(errors.New("database error"))
			},
			expected:      nil,
			expectedError: pkgErrs.New(response.HttpErrDatabase, constant.ErrUpdateIndustry),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			tc.mockSetup(mockRepo)
			service := NewIndustryService(mockRepo, testLogger)

			result, err := service.UpdateIndustry(context.Background(), tc.request)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.expected != nil {
					assert.Equal(t, tc.expected.ParentIndustry, result.ParentIndustry)
					assert.Equal(t, tc.expected.ChildIndustry, result.ChildIndustry)
					assert.Equal(t, tc.expected.RiskLevel, result.RiskLevel)
					assert.Equal(t, tc.expected.MCC, result.MCC)
					assert.Equal(t, tc.expected.CommonMCC, result.CommonMCC)
				}
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create pointer to string
func ptr(s string) *string {
	return &s
}
