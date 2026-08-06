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

func TestBlockUseCase(t *testing.T) {

	testcases := []struct {
		Name      string
		Request   *merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest
		WantErr   bool
		MockSetup func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService)
	}{
		{
			Name: "SUCCESS: Block Use Case",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
				SetStatus:  true,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, nil)
				repo.On("RegisterForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, nil)
				rabbitMqExt.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				merchantSvc.On("UpdateStatusByID", mock.Anything, constant.MerchantStatusDeactivated, "deactivated by parent merchant via dashboard", "merchant-id").Return(nil)
			},
		},
		{
			Name: "SUCCESS: Block Use Case with merchant user",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  uuid.NewString(),
				SetStatus:  true,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, nil)
				repo.On("RegisterForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, nil)
				rabbitMqExt.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				merchantSvc.On("UpdateStatusByID", mock.Anything, constant.MerchantStatusDeactivated, "deactivated by parent merchant via dashboard", "merchant-id").Return(nil)
			},
		},
		{
			Name: "SUCCESS: Block Use Case for existing data",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
				SetStatus:  true,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, nil)
				repo.On("RegisterForbiddenUsecase", mock.Anything, mock.Anything).Return(&merchantforbiddenusecaseModel.MerchantForbiddenUseCase{}, nil)
				rabbitMqExt.On(
					"PublishActivity",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				merchantSvc.On("UpdateStatusByID", mock.Anything, constant.MerchantStatusDeactivated, "deactivated by parent merchant via dashboard", "merchant-id").Return(nil)
			},
		},
		{
			Name: "ERROR: Validate submerchant",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "NOTEXIST",
				Requester:  uuid.NewString(),
			},
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("ValidateSubMerchantParent", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error"))
			},
		},
		{
			Name: "ERROR: Merchant not exists",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "NOTEXIST",
				Requester:  constant.UserSystemType,
			},
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		{
			Name: "ERROR: Error find merchant",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "NOTEXIST",
				Requester:  constant.UserSystemType,
			},
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			Name: "ERROR: Usecase not exists",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "NOTEXIST",
				Requester:  constant.UserSystemType,
			},
			WantErr: true,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
			},
		},
		{
			Name: "ERROR: Get forbidden usecase",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			Name: "ERROR: Register new forbidden usecase",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, nil)
				repo.On("RegisterForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			Name: "ERROR: Fail publish activity",
			Request: &merchantforbiddenusecaseModel.MerchantForbiddenUseCaseRequest{
				MerchantID: "merchant-id",
				UseCase:    "DISBURSEMENT",
				Requester:  constant.UserSystemType,
				SetStatus:  true,
			},
			WantErr: false,
			MockSetup: func(repo *repoMocks.IMerchantForbiddenUsecaseRepository, rabbitMqExt *mockRabbitMq.RabbitMQExt, merchantSvc *svcMocks.IMerchantService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, mock.Anything).Return(&merchant.Merchant{}, nil)
				merchantSvc.On("UpdateStatusByID", mock.Anything, constant.MerchantStatusDeactivated, "deactivated by parent merchant via dashboard", "merchant-id").Return(nil)
				repo.On("GetForbiddenUsecase", mock.Anything, mock.Anything).Return(nil, nil)
				repo.On("RegisterForbiddenUsecase", mock.Anything, mock.Anything).Return(&merchantforbiddenusecaseModel.MerchantForbiddenUseCase{}, nil)
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
			merchantSvc := svcMocks.NewIMerchantService(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			rabbitMqExt := mockRabbitMq.NewRabbitMQExt(t)

			tc.MockSetup(repo, rabbitMqExt, merchantSvc)

			service := New(loggerMock, repo, rabbitMqExt, merchantSvc)
			err := service.BlockUseCase(context.Background(), tc.Request)
			if tc.WantErr {
				assert.NotNil(t, err)
			}
		})
	}
}
