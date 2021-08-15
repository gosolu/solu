package core

import (
	"context"
	"fmt"
)

type (
	TraceKeyType int
	SpanKeyType  int
)

var (
	TraceKey TraceKeyType
	SpanKey  SpanKeyType
)

// TraceID returns the current context trace ID
func TraceID(ctx context.Context) string {
	val := ctx.Value(TraceKey)
	if val != nil {
		if value, ok := val.(string); ok {
			return value
		}
	}
	return ""
}

// SpanID returns context's span ID
func SpanID(ctx context.Context) string {
	val := ctx.Value(SpanKey)
	if val != nil {
		if value, ok := val.(string); ok {
			return value
		}
	}
	return ""
}

// Trace specification
// see https://w3c.github.io/trace-context
const (
	TraceVersion = "00" //
	TraceFlag    = "00" // currently no flag support
)

// TraceparentValue returns traceparent http header value
func TraceparentValue(ctx context.Context) string {
	tid := TraceID(ctx)
	sid := SpanID(ctx)
	return fmt.Sprintf("%s-%s-%s-%s", TraceVersion, tid, sid, TraceFlag)
}
