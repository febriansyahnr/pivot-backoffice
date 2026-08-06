package stream

import (
	"context"
	"time"

	"github.com/paper-indonesia/pdk/v2/newRelicExt"
)

type Config struct {
	Host             string
	Port             int
	VHost            string
	Username         string
	Password         string
	HeartbeatSeconds int
	NR               newRelicExt.INewRelicExt
}

type ReadMessageConfig struct {
	StreamQueueName string
	ConsumerName    string
	RetryCount      int           // default: 3, set to -1 to disable retry
	RetryDelay      time.Duration // default: 200ms
	CommitSize      int           // default: 50
	CommitInterval  time.Duration // default: 30s
	Handler         func(context.Context, []byte) error
	ReconnectDelay  time.Duration // default: 3s
}

func (r *ReadMessageConfig) setDefaults() {
	if r.RetryCount < 0 {
		r.RetryCount = 0

	} else if r.RetryCount == 0 {
		r.RetryCount = 3
	}

	if r.RetryDelay == 0 {
		r.RetryDelay = 200 * time.Millisecond
	}

	if r.CommitSize <= 0 {
		r.CommitSize = 50
	}

	if r.CommitInterval == 0 {
		r.CommitInterval = 30 * time.Second
	}

	if r.ReconnectDelay == 0 {
		r.ReconnectDelay = 3 * time.Second
	}
}
