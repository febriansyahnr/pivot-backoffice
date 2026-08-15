package disbursementService_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreBankConfigModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankConfig"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/disbursement"
	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIsBankcodeOverbookingChannelAllowed(t *testing.T) {
	pkdLogger, _ := loggerMock.NewZapLogger(loggerMock.Config{})
	rdb := redisMock.NewIRedisExt(t)
	snapCoreRepo := repositoryMock.NewISnapCoreRepository(t)

	svc := New(&config.Config{}, pkdLogger, nil, nil, snapCoreRepo, nil,
		WithRedisClient(rdb))

	tests := []struct {
		name      string
		setupMock func()
		result    bool
		bankCode  string
	}{
		{
			name: "Empty bank code",
		},
		{
			name: "Error when getting list of overbooking cache",
			setupMock: func() {
				result := &redis.StringCmd{}
				result.SetErr(c.ErrSomeErrorForUnitTest)
				rdb.On("Get", c.BackgroundCtxMockType(), c.StringMockType()).Once().Return(result)
			},
			bankCode: "002",
		},
		{
			name: "Error when unmarshal list of overbooking cache",
			setupMock: func() {
				result := &redis.StringCmd{}
				result.SetVal(`["002"}]`)
				rdb.On("Get", c.BackgroundCtxMockType(), c.StringMockType()).Once().Return(result)
			},
			bankCode: "002",
		},
		{
			name: "Bank code is overbooking based on cache value",
			setupMock: func() {
				result := &redis.StringCmd{}
				cacheValue, _ := json.Marshal([]string{"002"})
				result.SetVal(string(cacheValue))
				rdb.On("Get", c.BackgroundCtxMockType(), c.StringMockType()).Once().Return(result)
			},
			bankCode: "002",
			result:   true,
		},
		{
			name: "Bank code is overbooking based on processor value",
			setupMock: func() {
				result := &redis.StringCmd{}
				result.SetErr(redis.Nil)
				rdb.On("Get", c.BackgroundCtxMockType(), c.StringMockType()).Once().Return(result)

				snapCoreRepo.On("GetBankCodeList", c.ValueCtxMockType(), mock.AnythingOfType("*snapCoreModel.GetBankCodeListRequest")).
					Once().Return(&snapCoreBankConfigModel.BankCodeListResponseData{
					TransferType: "INTRABANK",
					BankCodes:    &[]string{"002"},
				}, nil)

				rdb.On("Set", c.ValueCtxMockType(), c.StringMockType(), mock.Anything, mock.Anything).Return(&redis.StatusCmd{})
			},
			bankCode: "002",
			result:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			result := svc.IsBankcodeOverbookingChannelAllowed(context.Background(), test.bankCode, uuid.NewString())
			assert.Equal(t, test.result, result)
		})
	}
}
