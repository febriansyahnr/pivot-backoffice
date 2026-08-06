package disbursementController

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/test"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/testcontainers/testcontainers-go"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"golang.org/x/sync/errgroup"
)

var (
	consulContainer testcontainers.Container
	mysqlContainer  testcontainers.Container
	rmqContainer    testcontainers.Container
	redisContainer  testcontainers.Container
	loggerMock      logger.ILogger
	pdkLoggerMock   pdkLogger.ILogger
	db              mySqlExt.IMySqlExt
	publisher       rabbitMqExt.IRabbitMQExt
	cacheClient     redisExt.IRedisExt
	consulURL       string
)

func TestMain(m *testing.M) {
	statsdHost := "1.2.3.4" // NOSONAR
	mntr, _ := monitor.New("testing", statsdHost, "5555")
	monitor.SetGlobalMonitoring(mntr)

	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		log.Println(constant.SkipIntegrationTest)
		m.Run()
		return
	}

	var (
		errG errgroup.Group
		err  error
	)

	ctx := context.Background()

	defer func() {
		if loggerMock != nil {
			_ = loggerMock.Sync()
		}

		if pdkLoggerMock != nil {
			_ = pdkLoggerMock.Sync()
		}

		ffclient.Close()

		if db != nil {
			db.Close()
		}
		if publisher != nil {
			publisher.Close()
		}
		if consulContainer != nil {
			_ = consulContainer.Terminate(ctx)
		}
		if mysqlContainer != nil {
			mysqlContainer.Terminate(ctx)
		}
		if rmqContainer != nil {
			rmqContainer.Terminate(ctx)
		}
	}()

	// Setup logger
	loggerMock, pdkLoggerMock, err = test.SetupLogger()
	if err != nil {
		panic(err)
	}

	// Setup Goff
	errG.Go(func() error {
		var err error

		consulContainer, consulURL, err = test.SetupConsul(ctx)
		if err != nil {
			return err
		}

		err = test.SetupFeatureFlag(consulURL)
		if err != nil {
			return err
		}

		err = test.SetupGoff(ctx, consulURL, pdkLoggerMock)
		if err != nil {
			return err
		}

		return nil
	})

	// Setup Database
	errG.Go(func() error {
		var err error

		mysqlContainer, db, err = test.SetupMysql(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	// Setup RabbitMQ
	errG.Go(func() error {
		var err error

		rmqContainer, publisher, err = test.SetupRabbitMQ(ctx, pdkLoggerMock)
		if err != nil {
			return err
		}

		return nil
	})

	// Setup Redis
	errG.Go(func() error {
		var err error

		redisContainer, cacheClient, err = test.SetupRedis(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	// Wait for all goroutines to complete
	if err := errG.Wait(); err != nil {
		panic(err)
	}

	m.Run()
}
