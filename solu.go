// package solu contains some components
package solu

import (
	"context"
	"net/http"
)

type (
	abortContextType struct{}
	abortValueType   struct {
		status int
		reason string
	}
)

var (
	abortContextKey   abortContextType
	abortContextValue bool
)

func isAborted(ctx context.Context) bool {
	val, ok := ctx.Value(abortContextKey).(*abortValueType)
	if !ok {
		return false
	}
	if val == nil {
		return false
	}
	return val != nil
}

func Abort(ctx context.Context) context.Context {
	val := &abortValueType{
		status: http.StatusOK,
		reason: "server abort",
	}
	return context.WithValue(ctx, abortContextKey, val)
}

func AbortWithStatus(ctx context.Context, status int) context.Context {
	val := &abortValueType{
		status: status,
		reason: "server abort",
	}
	return context.WithValue(ctx, abortContextKey, val)
}
