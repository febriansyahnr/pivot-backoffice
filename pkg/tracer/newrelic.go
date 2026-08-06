package tracer

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
)

type nrSpanWrapper struct {
	trx     *newrelic.Transaction
	segment *newrelic.Segment
}

func (n *nrSpanWrapper) End() { n.segment.End() }

func (n *nrSpanWrapper) NewGoroutine(ctx context.Context) context.Context {
	return newrelic.NewContext(ctx, n.trx.NewGoroutine())
}

func (n *nrSpanWrapper) SetAttributes(attrs ...KeyValue) {
	for _, attr := range attrs {
		n.segment.AddAttribute(string(attr.Key), attr.Value.AsString())
	}
}
