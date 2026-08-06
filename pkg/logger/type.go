package logger

import (
	"context"

	"go.uber.org/zap"
)

type ILogger interface {
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Panic(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
	Sync() error

	GetLogger() *zap.Logger
}

type Config struct {
	Environment string
	ServiceName string
}

type Option func(*zap.Config)
