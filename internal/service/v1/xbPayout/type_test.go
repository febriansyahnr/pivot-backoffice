package xbPayoutService

import (
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

var (
	loggerMock    logger.ILogger
	pdkLoggerMock pdkLogger.ILogger
)

func init() {
	pdkLoggerMock, _ = pdkLogger.NewZapLogger(pdkLogger.Config{})
}
