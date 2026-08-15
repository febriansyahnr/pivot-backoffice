package disbursementService

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/test"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"golang.org/x/sync/errgroup"
)

var (
	loggerMock    logger.ILogger
	pdkLoggerMock pdkLogger.ILogger
	db            mySqlExt.IMySqlExt
	redisClient   redisExt.IRedisExt
	publisher     rabbitMqExt.IRabbitMQExt

	mysqlContainer, rmqContainer, redisContainer testcontainers.Container
)

func TestMain(m *testing.M) {
	var err error

	if loggerMock, pdkLoggerMock, err = test.SetupLogger(); err != nil {
		panic(err)
	}
	defer pdkLoggerMock.Sync()
	defer loggerMock.Sync()

	_, _ = monitor.New("backend-portal", "0.0.0.0", "1234") // NOSONAR

	cwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(cwd, "backend-portal")

	targetPath := filepath.Join(projectRoot, "test", "consul", "backend-portal", "feature-flag.yaml")

	_ = ffclient.Init(ffclient.Config{
		Retriever:    &fileretriever.Retriever{Path: targetPath},
		DataExporter: ffclient.DataExporter{},
	})
	defer ffclient.Close()

	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		log.Println(constant.SkipIntegrationTest)

		m.Run()
		return
	}

	var errG errgroup.Group
	var ctx = context.Background()

	defer func() {
		if db != nil {
			db.Close()
		}
		if publisher != nil {
			publisher.Close()
		}
		if mysqlContainer != nil {
			mysqlContainer.Terminate(ctx)
		}
		if rmqContainer != nil {
			rmqContainer.Terminate(ctx)
		}
		if redisContainer != nil {
			redisClient.Close()
			redisContainer.Terminate(ctx)
		}
	}()

	// Setup Database
	errG.Go(func() (err error) {

		mysqlContainer, db, err = test.SetupMysql(ctx)
		return err
	})

	// Setup RabbitMQ
	errG.Go(func() (err error) {

		rmqContainer, publisher, err = test.SetupRabbitMQ(ctx, pdkLoggerMock)
		return err
	})

	errG.Go(func() (err error) {
		redisContainer, redisClient, err = test.SetupRedis(ctx)
		return err
	})

	// Wait for all goroutines to complete
	if err = errG.Wait(); err != nil {
		panic(err)
	}

	m.Run()
}

var (
	DisbursementMockType    = mock.AnythingOfType("*disbursementModel.Disbursement")
	NotificationMockType    = mock.AnythingOfType("*notification.PushNotification")
	CheckAccountReqMockType = mock.AnythingOfType("*beneficiaryAccountModel.CheckAccountRequest")
	BankTransferReqMockType = mock.AnythingOfType("*routingProcessorModel.BankTransferRequest")
)
