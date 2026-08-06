package ratelimiter

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/test"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

var (
	loggerMock    logger.ILogger
	pdkLoggerMock pdkLogger.ILogger
)

func TestMain(m *testing.M) {
	var err error

	// Setup logger
	loggerMock, pdkLoggerMock, err = test.SetupLogger()
	if err != nil {
		panic(err)
	}
	defer loggerMock.Sync()

	m.Run()
}
