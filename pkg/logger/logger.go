package logger

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"sort"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger/encoder"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	zapLog *zap.Logger
}

var encoders = &sync.Map{}

func WithMaskingSensitiveData(fields []string) Option {
	return func(c *zap.Config) {

		sort.Strings(fields)
		buf, _ := json.Marshal(fields)

		id := base64.RawStdEncoding.EncodeToString(buf)
		if _, ok := encoders.LoadOrStore(id, true); !ok {
			_ = zap.RegisterEncoder(id, func(ec zapcore.EncoderConfig) (zapcore.Encoder, error) {
				return encoder.NewJSONEncoder(ec, encoder.WithMaskSensitiveData(fields)), nil
			})
		}

		c.Encoding = id
	}
}

func New(config Config, opts ...Option) (ILogger, error) {
	zapConfig := zap.Config{}

	if config.Environment == constant.EnvironmentLocal || config.Environment == constant.EnvironmentDevelopment {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		zapConfig.Development = true
	} else {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		zapConfig.Development = false
		zapConfig.Sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
	}

	zapConfig.Encoding = "json"
	zapConfig.EncoderConfig = zap.NewProductionEncoderConfig()
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}
	zapConfig.InitialFields = map[string]interface{}{
		"env": config.Environment,
		"app": config.ServiceName,
	}

	for _, opt := range opts {
		opt(&zapConfig)
	}

	zapLog, err := zapConfig.Build()
	return &logger{zapLog}, err
}

func (l *logger) getTraceId(ctx context.Context) string {
	if ctx.Value(pdkConst.CtxTraceIdKey) != nil {
		return ctx.Value(pdkConst.CtxTraceIdKey).(string)
	}
	return ""
}

func (l *logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	trace := l.getTraceId(ctx)
	if trace != "" {
		fields = append(fields, zap.String("trace_id", trace))
	}
	l.zapLog.Debug(msg, fields...)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	trace := l.getTraceId(ctx)
	if trace != "" {
		fields = append(fields, zap.String("trace_id", trace))
	}

	if !util.ContainsPrefix(constant.IgnoreLoggingPath, msg) {
		l.zapLog.Info(msg, fields...)
	}
}

func (l *logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	trace := l.getTraceId(ctx)
	if trace != "" {
		fields = append(fields, zap.String("trace_id", trace))
	}
	l.zapLog.Error(msg, fields...)
}

func (l *logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	trace := l.getTraceId(ctx)
	if trace != "" {
		fields = append(fields, zap.String("trace_id", trace))
	}
	l.zapLog.Warn(msg, fields...)
}

func (l *logger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	trace := l.getTraceId(ctx)
	if trace != "" {
		fields = append(fields, zap.String("trace_id", trace))
	}
	l.zapLog.Panic(msg, fields...)
}

func (l *logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	trace := l.getTraceId(ctx)
	if trace != "" {
		fields = append(fields, zap.String("trace_id", trace))
	}
	l.zapLog.Fatal(msg, fields...)
}

func (l *logger) Sync() error {
	return l.zapLog.Sync()
}

func (l *logger) GetLogger() *zap.Logger {
	return l.zapLog
}
