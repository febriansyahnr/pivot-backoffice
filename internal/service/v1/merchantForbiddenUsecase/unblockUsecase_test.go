package merchantForbiddenUsecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	merchantforbiddenusecaseModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchantForbiddenUsecase"
	mockRabbitMq "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	svcMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUnblockUsecase(t *testing.T) {
	response := []*merchantforbiddenusecaseModel.MerchantForbiddenUseCase{
		{
			MerchantID: "merchant-id",
			UseCase:    "DISBURSEMENT",
		},
	}
	testcases := []struct {
		Name      string
		Request   *merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest
		WantErr   bool
		MockSetup func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService)
	}{
		{
			Name: "SUCCESS: Unblock usecase",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
				SetStatus:  true,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(response, nil)
				repo.On("RemoveForbiddenUsecase", mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				merchantSvc.On("UpdateStatusByID", mock.Anything, constant.MerchantStatusActive, "reactivated by parent merchant via dashboard", "merchant-id").Return(nil)
				rabbitMqExt.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
		{
			Name: "SUCCESS: Unblock usecase with merchant requester",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  uuid.NewString(),
				SetStatus:  true,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				merchantSvc.On("UpdateStatusByID", mock.Anything, constant.MerchantStatusActive, "reactivated by parent merchant via dashboard", "merchant-id").Return(nil)
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(response, nil)
				repo.On("RemoveForbiddenUsecase", mock.Anything, mock.Anything).Return(nil)
				rabbitMqExt.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
			},
		},
		{
			Name: "SUCCESS: Usecase not exists",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				repo.On(
					"GetForbiddenUsecase",
					mock.Anything,
					mock.Anything,
				).Return([]*merchantforbiddenusecaseModel.MerchantForbiddenUseCase{}, nil)
			},
		},
		{
			Name: "ERROR: unable to find merchant",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
			},
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		{
			Name: "ERROR: error find merchant",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
			},
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			Name: "ERROR: error get forbidden usecase",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
			},
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			Name: "ERROR: Error remove usecase",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
			},
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(response, nil)
				repo.On("RemoveForbiddenUsecase", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
		},
		{
			Name: "ERROR: Publish activity",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
				SetStatus:  true,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				merchantSvc.On("UpdateStatusByID", mock.Anything, constant.MerchantStatusActive, "reactivated by parent merchant via dashboard", "merchant-id").Return(nil)
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(response, nil)
				repo.On("RemoveForbiddenUsecase", mock.Anything, mock.Anything).Return(nil)
				rabbitMqExt.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("error"))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			repo := repoMocks.NewIMerchantForbiddenUsecaseRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			rabbitMqExt := mockRabbitMq.NewRabbitMQExt(t)
			merchantSvc := svcMocks.NewIMerchantService(t)

			tc.MockSetup(repo, rabbitMqExt, merchantSvc)

			service := New(loggerMock, repo, rabbitMqExt, merchantSvc)
			err := service.UnblockUseCase(context.Background(), tc.Request)
			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
