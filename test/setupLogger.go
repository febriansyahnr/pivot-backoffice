package test

import (
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func SetupLogger() (logger.ILogger, pdkLogger.ILogger, error) {
	logger, err := logger.New(logger.Config{
		Environment: "test",
		ServiceName: "backend-portal",
	})
	if err != nil {
		return nil, nil, err
	}

	pdkLogger, err := pdkLogger.NewZapLogger(pdkLogger.Config{
		IsDevelopment: true,
		Environment:   "test",
		ServiceName:   "backend-portal",
	})

	return logger, pdkLogger, err
}
