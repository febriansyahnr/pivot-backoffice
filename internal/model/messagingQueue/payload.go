package messagingQueueModel

import (
	"time"

	"github.com/paper-indonesia/pdk/v2/amqp"
)

type PublishSettlementProcessPayload struct {
	SettlementType string
	Day            int
	MessageTTL     time.Duration
	Payload        any
	ModifyMessage  func(*amqp.Publishing) // for rabbit mq
}
