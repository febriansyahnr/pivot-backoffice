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

func TestList(t *testing.T) {
	merchantId := uuid.NewString()
	input := &ipwhitelistModel.GetIPWhitelistConfiguration{
		MerchantID: merchantId,
		Page:       1,
		PageSize:   10,
	}

	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IIPWhitelistRepository)
		input   *ipwhitelistModel.GetIPWhitelistConfiguration
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IIPWhitelistRepository) {
				repo.On("List", mock.Anything, mock.Anything).Return([]*ipwhitelistModel.IPWhitelistConfiguration{
					{
						IP:         IPAddress,
						Subnet:     Subnet,
						Action:     constant.ActionBlock,
						MerchantID: merchantId,
					},
				}, int64(10), nil)

			},
			input:   input,
			wantErr: false,
		},
		{
			name: "ERROR: Get List",
			setup: func(repo *mocksRepo.IIPWhitelistRepository) {
				repo.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("error"))
			},
			input:   input,
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
			_, err := svc.(*IPWhitelistService).List(context.Background(), tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
		})
	}
}
