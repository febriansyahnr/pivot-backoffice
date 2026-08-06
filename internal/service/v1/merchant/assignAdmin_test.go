package merchant

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockSvc "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssignSubMerchantAdmin(t *testing.T) {

	testCases := []struct {
		Name      string
		Setup     func(userSvc *mockSvc.IUserService, repo *mockRepo.IMerchantRepository)
		ExpectErr bool
	}{
		{
			Name: "SUCCESS: Assign submerchant admin",
			Setup: func(userSvc *mockSvc.IUserService, repo *mockRepo.IMerchantRepository) {
				repo.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&merchant.Merchant{Name: "name"}, nil)

				userSvc.On(
					"CreateMerchantUser",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*user.MerchantUserRequest"),
				).Return(
					&userModel.User{
						Email: "email@example.id"},
					nil,
				)
			},
			ExpectErr: false,
		},
		{
			Name: "ERROR: Assign submerchant admin",
			Setup: func(userSvc *mockSvc.IUserService, repo *mockRepo.IMerchantRepository) {
				repo.On(
					"FindMerchantByID",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(&merchant.Merchant{Name: "name"}, nil)

				userSvc.On(
					"CreateMerchantUser",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*user.MerchantUserRequest"),
				).Return(
					nil,
					errors.New("errors"),
				)
			},
			ExpectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			userSvc := mockSvc.NewIUserService(t)
			repo := mockRepo.NewIMerchantRepository(t)
			tc.Setup(userSvc, repo)
			svc := New(repo, nil, nil, nil, nil, nil, WithUserService(userSvc))
			err := svc.AssignSubMerchantAdmin(context.Background(), &merchant.SubMerchantAdminRequest{
				MerchantId: uuid.NewString(),
				Email:      "email@example.id",
				Name:       "name",
			})

			if tc.ExpectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
