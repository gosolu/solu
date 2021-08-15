package slog

import (
	"context"
	"testing"

	"github.com/gosolu/solu/internal/core"
)

func TestForkSpan(t *testing.T) {
	ctx := context.Background()

	// initial context's span ID is empty
	sid := core.SpanID(ctx)
	if sid != "" {
		t.Errorf("Span ID should be empty")
		return
	}

	// fork context should has a span ID
	ctx = Fork(ctx)
	sid = core.SpanID(ctx)
	if sid == "" {
		t.Errorf("Fork span returns empty ID")
		return
	}
	if len(sid) != 16 {
		t.Errorf("Incorrect span length: %s", sid)
		return
	}
}
