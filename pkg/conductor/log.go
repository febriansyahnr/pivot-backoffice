package conductor

import (
	"github.com/conductor-sdk/conductor-go/sdk/log"
	"go.uber.org/zap"
)

func NewZapLogger(l *zap.Logger) ILogger {
	return log.NewZap(l)
}
