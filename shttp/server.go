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

type contextMiddleware struct {
	sync.Mutex

	middlewares []ContextWrapFn
}

func (cm *contextMiddleware) Add(fn ContextWrapFn) {
	cm.Lock()
	defer cm.Unlock()
	cm.middlewares = append(cm.middlewares, fn)
}

var (
	gConnMiddles = &contextMiddleware{
		middlewares: make([]ContextWrapFn, 0),
	}

	gBaseMiddles = &contextMiddleware{
		middlewares: make([]ContextWrapFn, 0),
	}
)

// AddToBaseContext add middlewares to base context
func AddToBaseContext(fns ...ContextWrapFn) {
	for _, fn := range fns {
		gBaseMiddles.Add(fn)
	}
}

func AddToConnContext(fns ...ContextWrapFn) {
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
	baseContext := func(ln net.Listener) context.Context {
		gBaseMiddles.Lock()
		defer gBaseMiddles.Unlock()

		ctx := context.Background()
		ctx = context.WithValue(ctx, startupContextKey, time.Now().Unix())
		for _, fn := range gBaseMiddles.middlewares {
			fnCtx := fn(ctx)
			// check startup context to verify context
			if val := fnCtx.Value(startupContextKey); val == nil {
				continue
			}
			ctx = fnCtx
		}
		return ctx
	}

	connContext := func(ctx context.Context, c net.Conn) context.Context {
		gConnMiddles.Lock()
		defer gConnMiddles.Unlock()

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
		BaseContext: baseContext,
		ConnContext: connContext,
	}
	return server.ListenAndServe()
}
