package test

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	pdkMySql "github.com/paper-indonesia/pdk/v2/mySqlExt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

func SetupMysql(ctx context.Context) (testcontainers.Container, mySqlExt.IMySqlExt, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "password",
			"MYSQL_DATABASE":      "backend_portal",
		},
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL"),
	}

	mysqlContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	host, err := mysqlContainer.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	port, err := mysqlContainer.MappedPort(ctx, "3306")
	if err != nil {
		return nil, nil, err
	}

	dbClient, err := mySqlExt.New(
		pdkMySql.Config{
			Host:     host,
			Port:     port.Port(),
			Username: "root",
			Password: "password",
			DBName:   "backend_portal",
		},
		// pdkMySql.WithLogger(pdkLog),
		pdkMySql.WithTracerProvider(trace.NewTracerProvider()),
		pdkMySql.WithMetricProvider(metric.NewMeterProvider()),
	)

	return mysqlContainer, dbClient, err
}
