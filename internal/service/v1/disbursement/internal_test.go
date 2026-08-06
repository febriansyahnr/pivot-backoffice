package disbursementService

import (
	"testing"

	redisMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBuildPayoutTransactionMutex(t *testing.T) {
	redisExt := redisMock.NewIRedisExt(t)
	mutex := redisMock.NewIMutexer(t)

	service := &DisbursementService{
		redisExt: redisExt,
	}

	redisExt.On(
		"NewMutex", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(mutex)
	assert.Equal(t, mutex, service.buildPayoutTransactionMutex(uuid.NewString()))
}
