package slog

import (
	"context"
	"testing"
)

func TestSpanID(t *testing.T) {
	ctx := context.Background()

	// initial context's span ID is empty
	sid := SpanID(ctx)
	if sid != "" {
		t.Errorf("Span ID should be empty")
		return
	}

	// fork context should has a span ID
	ctx = Fork(ctx)
	sid = SpanID(ctx)
	if sid == "" {
		t.Errorf("Fork span returns empty ID")
		return
	}
	if len(sid) != 16 {
		t.Errorf("Incorrect span length: %s", sid)
		return
	}
}
