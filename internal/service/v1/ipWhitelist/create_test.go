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
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

const (
	IPAddress = "1.1.1.1"
	Subnet    = "24"
)

func TestCreate(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt)
		input   *ipwhitelistModel.CreateIPWhitelistConfiguration
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)

				repo.On("List", mock.Anything, mock.Anything).Return([]*ipwhitelistModel.IPWhitelistConfiguration{
					{
						IP:     IPAddress,
						Subnet: Subnet,
						Action: constant.ActionBlock,
					},
				}, int64(10), nil)
				cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&redis.StatusCmd{})

			},
			input: &ipwhitelistModel.CreateIPWhitelistConfiguration{
				MerchantID:  uuid.NewString(),
				IP:          IPAddress,
				Subnet:      Subnet,
				Priority:    0,
				Action:      constant.ActionBlock,
				Description: "description",
			},
			wantErr: false,
		},
		{
			name: "ERROR: Invalid request",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			input: &ipwhitelistModel.CreateIPWhitelistConfiguration{
				MerchantID:  uuid.NewString(),
				IP:          IPAddress,
				Subnet:      Subnet,
				Priority:    0,
				Action:      constant.ActionBlock,
				Description: "description",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Update database",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			input: &ipwhitelistModel.CreateIPWhitelistConfiguration{
				MerchantID:  uuid.NewString(),
				IP:          IPAddress,
				Subnet:      Subnet,
				Priority:    0,
				Action:      constant.ActionBlock,
				Description: "description",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Update cache",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)

				repo.On("List", mock.Anything, mock.Anything).Return([]*ipwhitelistModel.IPWhitelistConfiguration{
					{
						IP:     IPAddress,
						Subnet: Subnet,
						Action: constant.ActionBlock,
					},
				}, int64(10), nil)

				cacheResult := &redis.StatusCmd{}
				cacheResult.SetErr(errors.New("error"))
				cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cacheResult)

			},
			input: &ipwhitelistModel.CreateIPWhitelistConfiguration{
				MerchantID:  uuid.NewString(),
				IP:          IPAddress,
				Subnet:      Subnet,
				Priority:    0,
				Action:      constant.ActionBlock,
				Description: "description",
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocksRepo.NewIIPWhitelistRepository(t)
			cache := mocksCache.NewIRedisExt(t)
			logger, _ := logger.NewZapLogger(logger.Config{})

			tc.setup(repo, cache)
			svc := New(logger, repo, cache)
			_, err := svc.Create(context.Background(), tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
		})
	}
}

func TestUpdateCache(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt)
		wantErr bool
	}{
		{
			name: "SUCCESS",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				repo.On("List", mock.Anything, mock.Anything).Return([]*ipwhitelistModel.IPWhitelistConfiguration{
					{
						IP:     IPAddress,
						Subnet: Subnet,
						Action: constant.ActionBlock,
					},
				}, int64(10), nil)
				cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&redis.StatusCmd{})

			},
			wantErr: false,
		},
		{
			name: "ERROR: Get List",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				repo.On("List", mock.Anything, mock.Anything).Return([]*ipwhitelistModel.IPWhitelistConfiguration{}, int64(0), errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Update cache",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				repo.On("List", mock.Anything, mock.Anything).Return([]*ipwhitelistModel.IPWhitelistConfiguration{
					{
						IP:     IPAddress,
						Subnet: Subnet,
						Action: constant.ActionBlock,
					},
				}, int64(10), nil)

				cacheResult := &redis.StatusCmd{}
				cacheResult.SetErr(errors.New("error"))
				cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cacheResult)

			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocksRepo.NewIIPWhitelistRepository(t)
			cache := mocksCache.NewIRedisExt(t)
			logger, _ := logger.NewZapLogger(logger.Config{})

			tc.setup(repo, cache)
			svc := New(logger, repo, cache)
			err := svc.(*IPWhitelistService).updateCache(context.Background(), uuid.NewString())
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
		})
	}
}
