package disbursementService

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"

	redsync "github.com/go-redsync/redsync/v4"
)

func (s *DisbursementService) buildPayoutTransactionMutex(payoutID string) redisExt.IMutexer {
	return s.redisExt.NewMutex(
		"backend-portal:payouts:"+payoutID+":bank-transfer:lock",
		redsync.WithExpiry(time.Minute),
		redsync.WithRetryDelay((200 * time.Millisecond)),
		redsync.WithFailFast(true),
		redsync.WithTries(256),
	)
}
