package shttp

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
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

type connMiddleware struct {
	sync.Mutex

	middlewares []func(context.Context) context.Context
}

func (cm *connMiddleware) Add(fn func(context.Context) context.Context) {
	cm.Lock()
	defer cm.Unlock()
	cm.middlewares = append(cm.middlewares, fn)
}

var gConnMiddles = &connMiddleware{
	middlewares: make([]func(context.Context) context.Context, 0),
}

func AddToConnContext(fns ...func(context.Context) context.Context) {
	for _, fn := range fns {
		gConnMiddles.Add(fn)
	}
}

type startupContextType struct{}

var startupContextKey startupContextType

func ListenAndServe(addr string, handler http.Handler) error {
	if handler != nil {
		gRouter.NotFound = handler
	}
	connContext := func(ctx context.Context, c net.Conn) context.Context {
		gConnMiddles.Lock()
		defer gConnMiddles.Unlock()

		ctx = context.WithValue(ctx, startupContextKey, time.Now().Unix())
		for _, fn := range gConnMiddles.middlewares {
			fnContext := fn(ctx)
			// check startup context to verify context
			if val := fnContext.Value(startupContextKey); val == nil {
				continue
			}
			ctx = fnContext
		}
		return ctx
	}
	server := &http.Server{
		Addr:        addr,
		Handler:     gRouter,
		ConnContext: connContext,
	}
	return server.ListenAndServe()
}
