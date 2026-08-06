package transferService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	mockRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {

	trnsfr := &transfer.Transfer{
		UUID:         uuid.New(),
		MerchantID:   uuid.New(),
		RecipientID:  uuid.New(),
		ReferenceID:  "reference-id",
		Amount:       100,
		TransferType: constant.MoneyFlowDirect,
		Currency:     constant.CurrencyIDR,
		Status:       constant.TransferStatusSuccess,
		Remarks:      "remarks",
		CreatedAt:    time.Now().UTC(),
	}
	testCases := []struct {
		name    string
		setup   func(merchantSvc *mocks.IMerchantService, repo *mockRepo.ITransferRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS: Get transfer list",
			setup: func(merchantSvc *mocks.IMerchantService, repo *mockRepo.ITransferRepository) {
				merchantSvc.On(
					"GetMerchantsByIDs",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("[]string"),
				).Return([]*merchant.Merchant{}, nil)

				repo.On(
					"GetList",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*transfer.GetTransferListRequest"),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return([]*transfer.Transfer{trnsfr}, int64(2), nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Get transfer list",
			setup: func(merchantSvc *mocks.IMerchantService, repo *mockRepo.ITransferRepository) {
				repo.On(
					"GetList",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("*transfer.GetTransferListRequest"),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(nil, int64(0), errors.New("errors"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantSvc := mocks.NewIMerchantService(t)
			logger, _ := logger.NewZapLogger(logger.Config{})
			repo := mockRepo.NewITransferRepository(t)
			tc.setup(merchantSvc, repo)

			svc := New(logger, nil, nil, nil, merchantSvc, repo)
			outputs, err := svc.GetList(context.Background(), &transfer.GetTransferListRequest{
				MerchantID:  uuid.New().String(),
				ReferenceID: uuid.New().String(),
			}, 1, 20)

			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, outputs)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestBuildListResponse(t *testing.T) {
	merchantId := uuid.New()
	recipientId := uuid.New()
	transferList := []*transfer.Transfer{
		{
			UUID:        uuid.New(),
			MerchantID:  merchantId,
			RecipientID: recipientId,
			Amount:      100,
		},
		{
			UUID:        uuid.New(),
			RecipientID: merchantId,
			MerchantID:  recipientId,
			Amount:      100,
		},
	}
	testCases := []struct {
		name    string
		setup   func(merchantSvc *mocks.IMerchantService)
		wantErr bool
	}{
		{
			name: "SUCCESS: Build Response",
			setup: func(merchantSvc *mocks.IMerchantService) {
				merchantSvc.On(
					"GetMerchantsByIDs",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("[]string"),
				).Return([]*merchant.Merchant{}, nil)

			},
		},
		{
			name: "ERROR: Build Response",
			setup: func(merchantSvc *mocks.IMerchantService) {
				merchantSvc.On(
					"GetMerchantsByIDs",
					constant.ValueCtxMockType(),
					mock.AnythingOfType("[]string"),
				).Return(nil, errors.New("error"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantSvc := mocks.NewIMerchantService(t)
			logger, _ := logger.NewZapLogger(logger.Config{})

			tc.setup(merchantSvc)
			svc := New(logger, nil, nil, nil, merchantSvc, nil)
			_, err := svc.(*TransferService).buildListResponse(context.Background(), uuid.NewString(), transferList)

			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
