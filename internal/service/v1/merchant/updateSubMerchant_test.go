package merchant

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/location"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateSubMerchant(t *testing.T) {
	ctx := context.Background()
	merchant := merchantModel.Merchant{}
	request := &merchantModel.UpdateMerchantRequest{
		Name:          "test-1",
		Description:   "description-1",
		Logo:          "logo-1",
		MerchantEmail: "email-1",
		MerchantPhone: "phone-1",
		PICEmail:      "pic email-1",
		PICPhone:      "pic phone-1",
		PICName:       "pic name-1",
		PICJobTitle:   "job title-1",
		DistrictId:    190,
	}

	userSvc := mocks.NewIUserService(t)

	tests := []struct {
		name          string
		setupMocks    func(repo *mockRepo.IMerchantRepository, locRepo *mockRepo.IAddrLocationRepository)
		expectedError bool
		request       *merchantModel.UpdateMerchantRequest
	}{
		{
			name:          "SUCCESS: Update",
			expectedError: false,
			setupMocks: func(repo *mockRepo.IMerchantRepository, locRepo *mockRepo.IAddrLocationRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
			},
			request: request,
		},
		{
			name:          "SUCCESS: Update without district",
			expectedError: false,
			setupMocks: func(repo *mockRepo.IMerchantRepository, locRepo *mockRepo.IAddrLocationRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			request: &merchantModel.UpdateMerchantRequest{
				Name:        "test-1",
				Description: "description-1",
				Logo:        "logo-1",
			},
		},
		{
			name:          "error create sub merchant user",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository, locRepo *mockRepo.IAddrLocationRepository) {
				request.PICInvitation = true

				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(&merchantModel.Merchant{
					PICInvitation: constant.MerchantPICNotInvited,
				}, nil)
				userSvc.On("FindUserByEmail", mock.Anything, mock.Anything).Once().Return(nil, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(&location.District{}, nil)
				repo.On("Update", mock.Anything, mock.Anything).Once().Return(nil)
				userSvc.On("CreateMerchantUser", mock.Anything, mock.Anything).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			request: request,
		},
		{
			name:          "error email already registered",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockRepo.IAddrLocationRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(&merchantModel.Merchant{
					PICInvitation: constant.MerchantPICNotInvited,
				}, nil)
				userSvc.On("FindUserByEmail", mock.Anything, mock.Anything).Once().Return(&user.User{}, nil)
			},
			request: request,
		},
		{
			name:          "error find user by email",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository, _ *mockRepo.IAddrLocationRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Once().Return(&merchantModel.Merchant{
					PICInvitation: constant.MerchantPICNotInvited,
				}, nil)
				userSvc.On("FindUserByEmail", mock.Anything, mock.Anything).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			request: request,
		},
		{
			name:          "error merchant not found",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository, locRepo *mockRepo.IAddrLocationRepository) {
				request.PICInvitation = false

				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
			request: request,
		},
		{
			name:          "error find merchant",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository, locRepo *mockRepo.IAddrLocationRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
			request: request,
		},
		{
			name:          "error get district by id",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository, locRepo *mockRepo.IAddrLocationRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			request: request,
		},
		{
			name:          "error district not found",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository, locRepo *mockRepo.IAddrLocationRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant, nil)
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Once().Return(nil, nil)
			},
			request: request,
		},
		{
			name:          "error update merchant",
			expectedError: true,
			setupMocks: func(repo *mockRepo.IMerchantRepository, locRepo *mockRepo.IAddrLocationRepository) {
				repo.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant, nil)
				repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("error"))
				locRepo.On("GetDistrictById", mock.Anything, constant.Uint16MockType()).Return(&location.District{}, nil)
			},
			request: request,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mockRepo.NewIMerchantRepository(t)
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			locRepo := mockRepo.NewIAddrLocationRepository(t)

			tt.setupMocks(repo, locRepo)

			s := New(repo, logger, nil, nil, nil, nil, WithLocationRepository(locRepo), WithUserService(userSvc))

			_, err := s.UpdateSubMerchant(ctx, tt.request)

			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

		})
	}

}
