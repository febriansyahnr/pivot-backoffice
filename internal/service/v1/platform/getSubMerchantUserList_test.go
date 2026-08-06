package platformService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockservice "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSubMerchantUserList(t *testing.T) {

	testCases := []struct {
		name    string
		setup   func(merchantSvc *mockservice.IMerchantService, userSvc *mockservice.IUserService)
		wantErr bool
	}{
		{
			name: "SUCCESS: Get submerchant user list",
			setup: func(merchantSvc *mockservice.IMerchantService, userSvc *mockservice.IUserService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				userSvc.On(
					"ListUsersByMerchantID",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(
					&commonModel.PaginationResponse{
						Data: []*user.User{
							&user.User{
								UUID: uuid.NewString(),
							},
						},
					},
					nil,
				)
			},
		},
		{
			name: "ERROR: Validate submerchant parent",
			setup: func(merchantSvc *mockservice.IMerchantService, userSvc *mockservice.IUserService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Get submerchant user list",
			setup: func(merchantSvc *mockservice.IMerchantService, userSvc *mockservice.IUserService) {
				merchantSvc.On(
					"ValidateSubMerchantParent",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

				userSvc.On(
					"ListUsersByMerchantID",
					constant.ValueCtxMockType(),
					mock.Anything,
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(
					nil,
					errors.New("error"),
				)
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := logger.NewZapLogger(logger.Config{})
			merchantSvc := mockservice.NewIMerchantService(t)
			userSvc := mockservice.NewIUserService(t)
			tc.setup(merchantSvc, userSvc)

			svc := New(logger, nil, nil, merchantSvc, nil, nil, nil, WithUserService(userSvc))

			_, err := svc.GetSubMerchantUserList(context.Background(), &platform.GetSubMerchantUsersRequest{})
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
