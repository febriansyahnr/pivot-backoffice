package ipwhitelistService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/ipWhitelist"
	mocksCache "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mocksRepo "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

func TestValidateIP(t *testing.T) {
	merchantId := uuid.NewString()

	testCases := []struct {
		name           string
		env            string
		whitelistedIPs []string
		setup          func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt)
		input          string
		wantErr        bool
	}{
		{
			name: "SUCCESS: No configuration",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				result := &redis.StringCmd{}
				result.SetErr(redis.Nil)
				cache.On("Get", mock.Anything, mock.Anything).Once().Return(result)

				repo.On("List", mock.Anything, mock.Anything).Return([]*ipwhitelistModel.IPWhitelistConfiguration{
					{
						IP:     "0.0.0.0",
						Subnet: "",
						Action: constant.ActionAllow,
					},
					{
						IP:     IPAddress,
						Subnet: Subnet,
						Action: constant.ActionBlock,
					},
				}, int64(10), nil)
				cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&redis.StatusCmd{})

				result2 := &redis.StringCmd{}
				result2.SetVal("0.0.0.0,,ALLOW|1.1.1.1,24,BLOCK")
				cache.On("Get", mock.Anything, mock.Anything).Return(result2)
			},
			input:   "10.10.10.10",
			wantErr: false,
		},
		{
			name: "SUCCESS: Default All Allow",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				result := &redis.StringCmd{}
				result.SetVal("0.0.0.0,,ALLOW")
				cache.On("Get", mock.Anything, mock.Anything).Return(result)
			},
			input:   "10.10.10.10",
			wantErr: false,
		},
		{
			name: "SUCCESS: Default All Block",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				result := &redis.StringCmd{}
				result.SetVal("0.0.0.0,,BLOCK")
				cache.On("Get", mock.Anything, mock.Anything).Return(result)
			},
			input:   "10.10.10.10",
			wantErr: true,
		},
		{
			name: "SUCCESS: Default IP in Range Block",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				result := &redis.StringCmd{}
				result.SetVal("10.0.0.0,8,BLOCK")
				cache.On("Get", mock.Anything, mock.Anything).Return(result)
			},
			input:   "10.0.0.10",
			wantErr: true,
		},
		{
			name: "SUCCESS: Block, IP in Range",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				result := &redis.StringCmd{}
				result.SetVal("10.0.0.0,8,BLOCK|10.0.1.0,8,BLOCK")
				cache.On("Get", mock.Anything, mock.Anything).Return(result)
			},
			input:   "10.0.1.10",
			wantErr: true,
		},
		{
			name: "SUCCESS: Allowed due to Priority",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				result := &redis.StringCmd{}
				result.SetVal("10.0.0.0,24,BLOCK|10.0.1.0,,ALLOW|10.0.1.25,8,BLOCK")
				cache.On("Get", mock.Anything, mock.Anything).Return(result)
			},
			input:   "10.0.1.0",
			wantErr: false,
		},
		{
			name: "SUCCESS: Block, No matched configuration",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				result := &redis.StringCmd{}
				result.SetVal("1.0.0.0,,BLOCK")
				cache.On("Get", mock.Anything, mock.Anything).Return(result)
			},
			input:   "10.10.10.10",
			wantErr: true,
		},
		{
			name: "ERROR: Get from cache",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				cacheResult := &redis.StringCmd{}
				cacheResult.SetErr(errors.New("error"))
				cache.On("Get", mock.Anything, mock.Anything).Return(cacheResult)
			},
			input:   "192.172.0.1",
			wantErr: true,
		},
		{
			name: "ERROR: when got invalid IP address, then return error",
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
			},
			input:   "192.172",
			wantErr: true,
		},
		{
			name: "when got IP address that is not in the whitelist, then return error with ip information",
			env:  constant.EnvironmentStaging,
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
				result := &redis.StringCmd{}
				result.SetVal("0.0.0.0,,BLOCK")
				cache.On("Get", mock.Anything, mock.Anything).Return(result)
			},
			input:   "10.10.10.10",
			wantErr: true,
		},
		{
			name:           "when got IP address was whitelisted as internal ip, then should succeed",
			env:            constant.EnvironmentStaging,
			whitelistedIPs: []string{"10.11.12.13"},
			setup: func(repo *mocksRepo.IIPWhitelistRepository, cache *mocksCache.IRedisExt) {
			},
			input:   "10.11.12.13",
			wantErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocksRepo.NewIIPWhitelistRepository(t)
			cache := mocksCache.NewIRedisExt(t)
			logger, _ := logger.NewZapLogger(logger.Config{})
			cfg := &config.Config{
				Environment:    tc.env,
				WhitelistedIPs: tc.whitelistedIPs,
			}
			tc.setup(repo, cache)
			svc := New(logger, repo, cache, WithConfig(cfg))
			err := svc.ValidateIP(context.Background(), merchantId, tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
		})
	}

}
