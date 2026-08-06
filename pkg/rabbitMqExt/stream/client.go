package stream

import (
	"errors"
	"time"

	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
)

type Client struct {
	env *stream.Environment
	cfg Config
}

func New(cfg Config) (*Client, error) {
	if cfg.NR == nil {
		return nil, errors.New("new relic agent must be set")
	}

	if cfg.HeartbeatSeconds <= 0 {
		cfg.HeartbeatSeconds = 30
	}

	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost(cfg.Host).
			SetPort(cfg.Port).
			SetUser(cfg.Username).
			SetPassword(cfg.Password).
			SetVHost(cfg.VHost).
			SetRequestedHeartbeat(time.Duration(cfg.HeartbeatSeconds) * time.Second).
			SetAddressResolver(stream.AddressResolver{Host: cfg.Host, Port: cfg.Port}),
	)
	if err != nil {
		return nil, err
	}
	return &Client{env: env, cfg: cfg}, nil
}
