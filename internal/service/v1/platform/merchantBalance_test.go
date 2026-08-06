package platformService

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	mockservice "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSubMerchantBalances(t *testing.T) {

	var (
		parentMerchantId    = uuid.NewString()
		merchantId          = uuid.NewString()
		listSubMerchantResp = &commonModel.PaginationResponse{
			Data: []*merchant.Merchant{
				{
					UUID: merchantId,
					ParentID: sql.NullString{
						String: parentMerchantId,
						Valid:  true,
					},
				},
			},
			Meta: commonModel.Meta{
				Page:       1,
				PerPage:    10,
				TotalItems: 1,
				TotalPages: 1,
			},
		}
	)

	testCases := []struct {
		name        string
		request     *platform.GetBulkBalanceRequest
		setup       func(merchantSvc *mockservice.IMerchantService, orchestratorSvc *mockservice.IOrchestratorService)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "SUCCESS: Get submerchant balances",
			request: &platform.GetBulkBalanceRequest{
				MerchantID: parentMerchantId,
				Page:       1,
				PerPage:    10,
				Usecase:    "test-usecase",
			},
			setup: func(merchantSvc *mockservice.IMerchantService, orchestratorSvc *mockservice.IOrchestratorService) {
				merchantSvc.On(
					"ListSubMerchantByParentID",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(listSubMerchantResp, nil)

				orchestratorSvc.On(
					"GetMerchantBulkBalances",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(
					map[string]*account_model.AvailableBalanceResponse{
						merchantId: &account_model.AvailableBalanceResponse{
							Balance:  100000,
							Currency: "IDR",
						},
					},
					nil,
				)
			},
		},
		{
			name: "ERROR: List submerchants",
			request: &platform.GetBulkBalanceRequest{
				MerchantID: parentMerchantId,
				Page:       1,
				PerPage:    10,
				Usecase:    "test-usecase",
			},
			setup: func(merchantSvc *mockservice.IMerchantService, orchestratorSvc *mockservice.IOrchestratorService) {
				merchantSvc.On(
					"ListSubMerchantByParentID",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, fmt.Errorf("error"))
			},
			wantErr:     true,
			expectedErr: errPkg.New(response.HttpErrInternal, constant.ErrGetSubMerchantList),
		},
		{
			name: "ERROR: Invalid submerchant object",
			request: &platform.GetBulkBalanceRequest{
				MerchantID: parentMerchantId,
				Page:       1,
				PerPage:    10,
				Usecase:    "test-usecase",
			},
			setup: func(merchantSvc *mockservice.IMerchantService, orchestratorSvc *mockservice.IOrchestratorService) {
				merchantSvc.On(
					"ListSubMerchantByParentID",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(&commonModel.PaginationResponse{
					Data: map[string]any{},
				}, nil)
			},
			wantErr:     true,
			expectedErr: errPkg.New(response.HttpErrInternal, constant.ErrGetSubMerchantList),
		},
		{
			name: "ERROR: Get submerchant balances",
			request: &platform.GetBulkBalanceRequest{
				MerchantID: parentMerchantId,
				Page:       1,
				PerPage:    10,
				Usecase:    "test-usecase",
			},
			setup: func(merchantSvc *mockservice.IMerchantService, orchestratorSvc *mockservice.IOrchestratorService) {
				merchantSvc.On(
					"ListSubMerchantByParentID",
					constant.ValueCtxMockType(),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(listSubMerchantResp, nil)

				orchestratorSvc.On(
					"GetMerchantBulkBalances",
					constant.ValueCtxMockType(),
					mock.Anything,
				).Return(
					nil,
					fmt.Errorf("error"),
				)
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := logger.NewZapLogger(logger.Config{})
			merchantSvc := mockservice.NewIMerchantService(t)
			orchestratorSvc := mockservice.NewIOrchestratorService(t)
			tc.setup(merchantSvc, orchestratorSvc)

			svc := New(logger, nil, nil, merchantSvc, nil, nil, nil, WithOrchestratorService(orchestratorSvc))

			_, err := svc.GetSubMerchantBalances(context.Background(), tc.request)
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
