package walletInsightService

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTotalBalance(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(svc *mocks.IOrchestratorService, cache *mockRedis.IRedisExt)
		wantErr bool
	}{
		{
			name: "SUCCESS: Cache",
			setup: func(svc *mocks.IOrchestratorService, cache *mockRedis.IRedisExt) {
				strResult := redis.StringCmd{}
				strResult.SetVal(`{"totalBalance":0}`)
				cache.On("Get", mock.Anything, mock.Anything).Return(&strResult)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS",
			setup: func(svc *mocks.IOrchestratorService, cache *mockRedis.IRedisExt) {
				strResult := redis.StringCmd{}
				strResult.SetVal(``)
				cache.On("Get", mock.Anything, mock.Anything).Return(&strResult)

				svc.On("GetWalletCustomersTotalBalance", mock.Anything, mock.Anything).Return(float64(0), nil)

				statResult := redis.StatusCmd{}
				statResult.SetVal(`OK`)
				cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&statResult)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Get from Cache",
			setup: func(svc *mocks.IOrchestratorService, cache *mockRedis.IRedisExt) {
				strResult := redis.StringCmd{}
				strResult.SetErr(errors.New("error"))
				cache.On("Get", mock.Anything, mock.Anything).Return(&strResult)

				svc.On("GetWalletCustomersTotalBalance", mock.Anything, mock.Anything).Return(float64(0), nil)

				statResult := redis.StatusCmd{}
				statResult.SetVal(`OK`)
				cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&statResult)

			},
			wantErr: false,
		},
		{
			name: "ERROR: Unmarshal from Cache",
			setup: func(svc *mocks.IOrchestratorService, cache *mockRedis.IRedisExt) {
				strResult := redis.StringCmd{}
				strResult.SetVal(`{}}`)
				cache.On("Get", mock.Anything, mock.Anything).Return(&strResult)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Get Balance",
			setup: func(svc *mocks.IOrchestratorService, cache *mockRedis.IRedisExt) {
				strResult := redis.StringCmd{}
				strResult.SetVal(``)
				cache.On("Get", mock.Anything, mock.Anything).Return(&strResult)

				svc.On("GetWalletCustomersTotalBalance", mock.Anything, mock.Anything).Return(float64(0), errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Store to Cache",
			setup: func(svc *mocks.IOrchestratorService, cache *mockRedis.IRedisExt) {
				strResult := redis.StringCmd{}
				strResult.SetVal(``)
				cache.On("Get", mock.Anything, mock.Anything).Return(&strResult)

				svc.On("GetWalletCustomersTotalBalance", mock.Anything, mock.Anything).Return(float64(0), nil)

				statResult := redis.StatusCmd{}
				statResult.SetErr(errors.New("error"))
				cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&statResult)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			orcSvc := mocks.NewIOrchestratorService(t)
			cache := mockRedis.NewIRedisExt(t)

			tc.setup(orcSvc, cache)
			svc := New(orcSvc, logger, cache)

			_, err := svc.TotalBalance(context.Background(), uuid.NewString(), true)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
