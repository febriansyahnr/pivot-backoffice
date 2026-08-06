package installmentplan

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInstallmentPlanService_List(t *testing.T) {

	req := &installmentPlanModel.FilterRequest{
		MerchantID: "merchant-id",
		Page:       1,
		PageSize:   10,
	}

	testCases := []struct {
		name      string
		wantErr   bool
		errDetail error
		mockSetup func(*repositoryMocks.IInstallmentPlanRepository)
	}{
		{
			name:    "SUCCESS: Get List",
			wantErr: false,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository) {
				repo.On("List", mock.Anything, mock.Anything).Return([]*installmentPlanModel.InstallmentPlan{
					{
						UUID: uuid.New().String(),
					},
				}, int64(1), nil)
			},
		},
		{
			name:      "ERROR: Get List",
			wantErr:   true,
			errDetail: pkgErrors.New(response.HttpErrDatabase, constant.ErrGetInstallmentPlan),
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository) {
				repo.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), fmt.Errorf("error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewIInstallmentPlanRepository(t)
			mockCreditCardSvc := serviceMocks.NewICreditCardService(t)

			tc.mockSetup(mockRepo)

			svc := NewInstallmentPlanService(mockLogger, mockRepo, mockCreditCardSvc, nil)
			result, _, err := svc.List(context.Background(), req)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
				require.Equal(t, tc.errDetail, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
			mockCreditCardSvc.AssertExpectations(t)
		})
	}
}
