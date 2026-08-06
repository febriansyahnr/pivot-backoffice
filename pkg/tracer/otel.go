package tracer

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type otelSpanWrapper struct {
	span trace.Span
}

func (o *otelSpanWrapper) End() {
	o.span.End()
}

func (o *otelSpanWrapper) NewGoroutine(ctx context.Context) context.Context { return ctx }

func (o *otelSpanWrapper) SetAttributes(attrs ...KeyValue) {
	otelAttrs := make([]attribute.KeyValue, len(attrs))
	for i, attr := range attrs {
		otelAttrs[i] = attribute.KeyValue(attr)
	}
	o.span.SetAttributes(otelAttrs...)
}
