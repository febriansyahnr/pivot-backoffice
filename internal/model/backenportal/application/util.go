package application

import (
	"os"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
)

func getServiceName(serviceName string) string {
	// First check environment variable
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		serviceName = name
	}

	// Get mode from command line args first (since this is how we run it)
	mode := os.Getenv("MODE")
	if len(os.Args) > 1 {
		mode = os.Args[1] // The command (serveHttp/serveRmqConsumer) is always the first argument
	}

	// Append suffix based on mode
	switch mode {
	case "serveHttp":
		return serviceName + "-http"
	case "serveConsumer":
		return serviceName + "-consumer"
	default:
		return serviceName
	}
}

func (a *Application) GetLogger() pdkLogger.ILogger {
	return a.pdkLog
}
