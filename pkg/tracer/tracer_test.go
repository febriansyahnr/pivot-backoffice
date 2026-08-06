package tracer_test

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/tracer"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func TestNew(t *testing.T) {
	tr := tracer.New("test-tracer")
	assert.NotNil(t, tr)
}

func TestWithOtelEnabled(t *testing.T) {
	tr := tracer.New("test")
	_, span := tr.Start(context.Background(), "test-span", tracer.WithOtelEnabled())
	assert.NotNil(t, span)
	span.End()
}

func TestWithAttributes(t *testing.T) {
	tr := tracer.New("test")
	_, span := tr.Start(context.Background(), "test-span",
		tracer.WithOtelEnabled(),
		tracer.WithAttributes(tracer.KeyValue(attribute.String("key", "value"))),
	)
	assert.NotNil(t, span)
	span.End()
}

func newNRContext(t *testing.T) context.Context {
	t.Helper()
	app, err := newrelic.NewApplication(newrelic.ConfigEnabled(false))
	assert.NoError(t, err)
	trx := app.StartTransaction("test-trx")
	return newrelic.NewContext(context.Background(), trx)
}

func TestStartOtelSpan(t *testing.T) {
	tr := tracer.New("test")
	ctx, span := tr.Start(context.Background(), "otel-span", tracer.WithOtelEnabled())
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	span.End()
}

func TestStartNRSpan(t *testing.T) {
	tr := tracer.New("test")
	ctx := newNRContext(t)
	newCtx, span := tr.Start(ctx, "nr-span")
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	span.End()
}

func TestStartOtelSpanWithAttributes(t *testing.T) {
	tr := tracer.New("test")
	_, span := tr.Start(context.Background(), "otel-span-attrs",
		tracer.WithOtelEnabled(),
		tracer.WithAttributes(tracer.KeyValue(attribute.String("key", "value"))),
	)
	assert.NotNil(t, span)
	span.End()
}

func TestStartNRSpanWithAttributes(t *testing.T) {
	tr := tracer.New("test")
	ctx := newNRContext(t)
	_, span := tr.Start(ctx, "nr-span-attrs",
		tracer.WithAttributes(tracer.KeyValue(attribute.String("key", "value"))),
	)
	assert.NotNil(t, span)
	span.End()
}

func TestSpanerNewGoroutineOtel(t *testing.T) {
	tr := tracer.New("test")
	ctx := context.Background()
	_, span := tr.Start(ctx, "test-span", tracer.WithOtelEnabled())
	defer span.End()

	result := span.NewGoroutine(ctx)
	assert.NotNil(t, result)
}

func TestSpanerSetAttributesOtel(t *testing.T) {
	tr := tracer.New("test")
	_, span := tr.Start(context.Background(), "test-span", tracer.WithOtelEnabled())
	defer span.End()

	span.SetAttributes(tracer.KeyValue(attribute.String("key", "value")))
}

func TestSpanerNewGoroutineNR(t *testing.T) {
	ctx := newNRContext(t)
	tr := tracer.New("test")

	_, span := tr.Start(ctx, "test-segment")
	defer span.End()

	newCtx := span.NewGoroutine(ctx)
	assert.NotNil(t, newCtx)
}

func TestSpanerSetAttributesNR(t *testing.T) {
	ctx := newNRContext(t)
	tr := tracer.New("test")

	_, span := tr.Start(ctx, "test-segment")
	span.SetAttributes(tracer.KeyValue(attribute.String("key", "value")))
	span.End()
}
