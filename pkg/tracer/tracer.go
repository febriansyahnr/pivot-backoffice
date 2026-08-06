package tracer

import (
	"context"
	"sync"

	"github.com/newrelic/go-agent/v3/newrelic"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Option func(t *Tracer)

type KeyValue attribute.KeyValue

type Tracer struct {
	name string
	otel trace.Tracer
	once sync.Once

	// options
	isOtelEnable bool
	attributes   []KeyValue
}

type Spaner interface {
	End()
	NewGoroutine(context.Context) context.Context
	SetAttributes(...KeyValue)
}

func New(name string) *Tracer {
	return &Tracer{name: name}
}

func (t *Tracer) Start(
	ctx context.Context,
	spanName string,
	opts ...Option,
) (context.Context, Spaner) {
	for _, opt := range opts {
		opt(t)
	}

	if t.isOtelEnable {
		t.once.Do(func() {
			t.otel = otel.Tracer(t.name)
		})
		attributes := make([]attribute.KeyValue, len(t.attributes))
		for i, attr := range t.attributes {
			attributes[i] = attribute.KeyValue(attr)
		}
		//nolint:spancheck
		newCtx, span := t.otel.Start(
			ctx,
			spanName,
			trace.WithAttributes(attributes...),
		)
		//nolint:spancheck
		return newCtx, &otelSpanWrapper{span}
	}

	trx := newrelic.FromContext(ctx)
	newCtx, segment := newrelic.NewContext(ctx, trx), trx.StartSegment(spanName)

	for _, attr := range t.attributes {
		segment.AddAttribute(string(attr.Key), attr.Value.AsString())
	}
	return newCtx, &nrSpanWrapper{trx: trx, segment: segment}
}

func WithOtelEnabled() Option {
	return func(t *Tracer) {
		t.isOtelEnable = true
	}
}

func WithAttributes(attrs ...KeyValue) Option {
	return func(t *Tracer) {
		t.attributes = append(t.attributes, attrs...)
	}
}
