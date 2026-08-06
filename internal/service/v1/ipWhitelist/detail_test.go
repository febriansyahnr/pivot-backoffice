package ipwhitelistService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	mocksCache "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestDetail(t *testing.T) {
	merchantId := uuid.NewString()

	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IIPWhitelistRepository)
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IIPWhitelistRepository) {
				repo.On("Detail", mock.Anything, mock.Anything).Return(&ipwhitelistModel.IPWhitelistConfiguration{
					IP:         IPAddress,
					Subnet:     Subnet,
					Action:     constant.ActionBlock,
					MerchantID: merchantId,
				}, nil)

			},
			wantErr: false,
		},
		{
			name: "ERROR: Not found",
			setup: func(repo *mocksRepo.IIPWhitelistRepository) {
				repo.On("Detail", mock.Anything, mock.Anything).Return(nil, nil)

			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid merchant id not same",
			setup: func(repo *mocksRepo.IIPWhitelistRepository) {
				repo.On("Detail", mock.Anything, mock.Anything).Return(&ipwhitelistModel.IPWhitelistConfiguration{
					IP:         IPAddress,
					Subnet:     Subnet,
					Action:     constant.ActionBlock,
					MerchantID: "",
				}, nil)

			},
			wantErr: true,
		},
		{
			name: "ERROR: Get Detail",
			setup: func(repo *mocksRepo.IIPWhitelistRepository) {
				repo.On("Detail", mock.Anything, mock.Anything).Return(&ipwhitelistModel.IPWhitelistConfiguration{}, errors.New("error"))
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocksRepo.NewIIPWhitelistRepository(t)
			cache := mocksCache.NewIRedisExt(t)
			logger, _ := logger.NewZapLogger(logger.Config{})

			tc.setup(repo)
			svc := New(logger, repo, cache)
			_, err := svc.(*IPWhitelistService).Detail(context.Background(), merchantId, uuid.NewString())
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
		})
	}
}
