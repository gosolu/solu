package shttp

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
	abortContextKey abortContextType

	abortDefaultReason = "Server encounter an error"
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
		status: http.StatusInternalServerError,
		reason: abortDefaultReason,
	}
	return context.WithValue(ctx, abortContextKey, val)
}

func AbortWithStatus(ctx context.Context, status int) context.Context {
	val := &abortValueType{
		status: status,
		reason: abortDefaultReason,
	}
	return context.WithValue(ctx, abortContextKey, val)
}

func AbortWithStatusReason(ctx context.Context, status int, reason string) context.Context {
	val := &abortValueType{
		status: status,
		reason: reason,
	}
	return context.WithValue(ctx, abortContextKey, val)
}

func AbortWithError(ctx context.Context, err error) context.Context {
	var reason string
	if err != nil {
		reason = err.Error()
	}
	val := &abortValueType{
		status: http.StatusInternalServerError,
		reason: reason,
	}
	return context.WithValue(ctx, abortContextKey, val)
}
