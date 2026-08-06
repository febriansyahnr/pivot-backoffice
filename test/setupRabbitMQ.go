package test

import (
	"context"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pdk/v2/amqp"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	rmqUsername = "testuser"
	rmqPassword = "yourpassword"
)

func SetupRabbitMQ(
	ctx context.Context, log logger.ILogger,
) (testcontainers.Container, rabbitMqExt.IRabbitMQExt, error) {
	request := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.13-alpine",
		ExposedPorts: []string{"5672/tcp"},
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER": rmqUsername,
			"RABBITMQ_DEFAULT_PASS": rmqPassword,
		},
		WaitingFor: wait.ForLog("Server startup complete"),
	}

	rabbitMQContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started:          true,
		ContainerRequest: request,
	})
	if err != nil {
		return nil, nil, err
	}

	host, err := rabbitMQContainer.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	port, err := rabbitMQContainer.MappedPort(ctx, "5672")
	if err != nil {
		return nil, nil, err
	}

	rmq, err := rabbitMqExt.New(
		config.RabbitMQConfig[string]{Host: host, Port: port.Port()},
		config.RabbitMQSecret{Username: rmqUsername, Password: rmqPassword}, log,
		nil,
	)

	return rabbitMQContainer, rmq, err
}

func OpenConnectionRabbitMQ(ctx context.Context, container testcontainers.Container) (*amqp.Connection, error) {
	config := amqp.Config{
		Vhost:      "/",
		Heartbeat:  10 * time.Second,
		Properties: amqp.NewConnectionProperties(),
	}
	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}
	port, err := container.MappedPort(ctx, "5672")
	if err != nil {
		return nil, err
	}
	return amqp.DialConfig(
		fmt.Sprintf("amqp://%s:%s@%s:%s/", rmqUsername, rmqPassword, host, port.Port()), config,
	)
}
