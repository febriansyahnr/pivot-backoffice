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

func TestIndustryService_DeleteIndustry(t *testing.T) {
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
		uuid          string
		mockSetup     func(*mocks.IIndustryRepository)
		expectedError error
	}{
		{
			name: "SUCCESS: Delete industry",
			uuid: "test-uuid",
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("IsIndustryUsedByMerchants", mock.Anything, "Technology", "Software").
					Return(false, nil)
				m.On("Delete", mock.Anything, "test-uuid").Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "ERROR: Industry not found",
			uuid: "non-existent-uuid",
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "non-existent-uuid").
					Return(nil, nil)
			},
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrIndustryNotFound),
		},
		{
			name: "ERROR: Industry in use by merchants",
			uuid: "test-uuid",
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("IsIndustryUsedByMerchants", mock.Anything, "Technology", "Software").
					Return(true, nil)
			},
			expectedError: pkgErrs.New(response.HttpErrRequest, constant.ErrIndustryInUse),
		},
		{
			name: "ERROR: Database error on get by ID",
			uuid: "test-uuid",
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(nil, errors.New("database error"))
			},
			expectedError: pkgErrs.New(response.HttpErrDatabase, constant.ErrDeleteIndustry),
		},
		{
			name: "ERROR: Database error on checking if used by merchants",
			uuid: "test-uuid",
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("IsIndustryUsedByMerchants", mock.Anything, "Technology", "Software").
					Return(false, errors.New("database error"))
			},
			expectedError: pkgErrs.New(response.HttpErrDatabase, constant.ErrDeleteIndustry),
		},
		{
			name: "ERROR: Database error on delete",
			uuid: "test-uuid",
			mockSetup: func(m *mocks.IIndustryRepository) {
				m.On("GetIndustryByID", mock.Anything, "test-uuid").
					Return(existingIndustry, nil)
				m.On("IsIndustryUsedByMerchants", mock.Anything, "Technology", "Software").
					Return(false, nil)
				m.On("Delete", mock.Anything, "test-uuid").
					Return(errors.New("database error"))
			},
			expectedError: pkgErrs.New(response.HttpErrDatabase, constant.ErrDeleteIndustry),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mocks.IIndustryRepository{}
			tc.mockSetup(mockRepo)
			service := NewIndustryService(mockRepo, testLogger)

			err := service.DeleteIndustry(context.Background(), tc.uuid)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
