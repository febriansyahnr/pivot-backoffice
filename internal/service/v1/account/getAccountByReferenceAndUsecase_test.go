package accountService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAccountByReferenceIDAndUsecase(t *testing.T) {
	referenceID := uuid.New()
	usecase := constant.ReferenceDisbursement
	userType := constant.UserTypeMerchant

	testcases := []struct {
		Name      string
		MockSetup func(mockRepo *repositoryMocks.IAccountRepository)
		WantErr   bool
	}{
		{
			Name: "SUCCESS: Get By Reference ID and Usecase",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository) {
				mockRepo.On(
					"GetByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&account_model.Account{}, nil)
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Get By Reference ID and Usecase",
			MockSetup: func(mockRepo *repositoryMocks.IAccountRepository) {
				mockRepo.On(
					"GetByReferenceIDAndUsecase",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("error"))

			},
			WantErr: true,
		},
	}
	for _, tc := range testcases {

		t.Run(tc.Name, func(t *testing.T) {
			loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewIAccountRepository(t)
			ctx := context.Background()

			tc.MockSetup(mockRepo)
			svc := New(loggerMock, nil, mockRepo, nil)
			account, err := svc.GetAccountByReferenceIDAndUsecase(ctx, referenceID, usecase, userType)

			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, account)
			}

		})
	}
}
