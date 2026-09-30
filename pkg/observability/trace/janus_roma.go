// Package trace provides OpenTelemetry tracing helpers (noop-safe skeleton for CI).
package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName     = "github.com/iiooiioo888/qianjunpo-reborn/phase5"
	AttrServiceJanus = "janus"
	AttrServiceRoma  = "roma"
)

// StartJanusToRomaSpan creates a client span for Janus → Roma RPC paths.
func StartJanusToRomaSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	tr := otel.Tracer(tracerName)
	ctx, span := tr.Start(ctx, "janus_to_roma/"+operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("qjp.edge", AttrServiceJanus),
			attribute.String("qjp.zone", AttrServiceRoma),
		),
	)
	return ctx, span
}

// EndSpan records err on the span and ends it.
func EndSpan(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
