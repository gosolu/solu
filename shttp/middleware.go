package shttp

import (
	"context"
	"net/http"
)

// MiddlewareHandle handle http request before user-defined handler func.
//
// For example, if you need authorize client before real operations:
//
//	func authMiddleware(w http.ResponseWriter, r *http.Request) context.Context {
//		auth, err := r.Cookie("auth_key")
//		if err != nil || auth.Valid() != nil {
//			return r.Context()
//		}
//		if auth.Value == "" {
//			return r.Context()
//		}
//		return context.WithValue(r.Context(), "user", auth)
//	}
//
// A middleware can abort request or allow it continue.
type MiddlewareHandle func(w http.ResponseWriter, r *http.Request) context.Context

// ContextWranFn is a function wrap context with value and return a new context.
type ContextWrapFn func(context.Context) context.Context

func abortHandleFunc(w http.ResponseWriter, r *http.Request) {
	abort, ok := r.Context().Value(abortContextKey).(*abortValueType)
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(abort.status)
	w.Write([]byte(abort.reason))
}

func With(handle http.HandlerFunc, middlewares ...MiddlewareHandle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, mid := range middlewares {
			ctx := mid(w, r)
			if isAborted(ctx) {
				abortHandleFunc(w, r)
				return
			}
			r = r.WithContext(ctx)
		}
		handle(w, r)
	}
}
