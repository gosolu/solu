// package solu contains some components
package solu

import (
	"context"
	"fmt"
	"strings"
)

type (
	traceKeyType struct{}
	spanKeyType  struct{}
)

var (
	TraceKey traceKeyType
	SpanKey  spanKeyType
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

// TraceWith create a new context with a given trace ID
func TraceWith(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceKey, traceID)
}

// SpanWith create a new context with a given span ID
func SpanWith(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, SpanKey, spanID)
}

// Trace specification
// see https://w3c.github.io/trace-context
const (
	TraceVersion = "00" //
	TraceFlag    = "00" // currently no flag support

	TraceInvalidTid = "00000000000000000000000000000000"
	TraceInvalidSid = "0000000000000000"
)

// TraceparentValue returns traceparent http header value
func TraceparentValue(ctx context.Context) string {
	tid := TraceID(ctx)
	sid := SpanID(ctx)
	return fmt.Sprintf("%s-%s-%s-%s", TraceVersion, tid, sid, TraceFlag)
}

// ParseTraceparent parse formatted traceparent value
// It returns TraceID and SpanID, ignore version and flags
// If a value is not correct formatted, it will return an error.
func ParseTraceparent(val string) (string, string, error) {
	parts := strings.Split(val, "-")
	if len(parts) != 4 {
		return "", "", fmt.Errorf("invalid trace value")
	}
	tid := parts[1]
	if len(tid) != len(TraceInvalidTid) || tid == TraceInvalidTid {
		tid = ""
	}
	sid := parts[2]
	if len(sid) != len(TraceInvalidSid) || sid == TraceInvalidSid {
		sid = ""
	}
	return tid, sid, nil
}
