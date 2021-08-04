// Package shttp contains some golang HTTP client methods
package shttp

import (
	"context"
	"io"
)

type Options struct {
}

func Get(ctx context.Context, addr string) {
}

func Post(ctx context.Context, addr string, body io.Reader) {
}

func Head(ctx context.Context, addr string) {
}

func Put(ctx context.Context, addr string) {
}

func Delete(ctx context.Context, addr string) {
}

func Option(ctx context.Context, addr string) {
}
