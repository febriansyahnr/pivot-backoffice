package test

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"

	pdkRedis "github.com/paper-indonesia/pdk/v2/redisExt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

func SetupRedis(ctx context.Context) (testcontainers.Container, redisExt.IRedisExt, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, nil, err
	}

	cacheClient, err := redisExt.New(
		pdkRedis.Config{
			Addr:     host + ":" + port.Port(),
			Password: "",
			DB:       0,
		},
		pdkRedis.WithTracerProvider(trace.NewTracerProvider()),
		pdkRedis.WithMetricProvider(metric.NewMeterProvider()),
	)

	return redisContainer, cacheClient, err
}
