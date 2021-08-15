// Package shttp contains some golang HTTP utilities
package shttp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gosolu/solu/internal/core"
)

const (
	TraceparentHeader = "traceparent"
	TracestateHeader  = "tracestate"
)

func metricLabels(res *http.Response, dur time.Duration) []string {
	labels := make([]string, 0, len(requestLabels))
	labels = append(labels, res.Request.Method)
	labels = append(labels, res.Request.Host)
	labels = append(labels, res.Request.URL.Path)
	labels = append(labels, strconv.Itoa(res.StatusCode))
	labels = append(labels, strconv.FormatInt(dur.Milliseconds(), 10))
	return labels
}

func doWithClient(ctx context.Context, req *http.Request, client *http.Client) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}
	if client == nil {
		return nil, fmt.Errorf("invalid client")
	}
	// add trace header
	if req.Header.Get(TraceparentHeader) == "" {
		trace := core.TraceparentValue(ctx)
		req.Header.Set(TraceparentHeader, trace)
	}
	start := time.Now()
	res, err := client.Do(req)
	labels := metricLabels(res, time.Now().Sub(start))
	if err != nil {
		httpErrorCounter.WithLabelValues(labels...).Inc()
	} else {
		httpRequestCounter.WithLabelValues(labels...).Inc()
	}
	return res, err
}

// Do http request, use http.DefaultClient.Do
// Wrap request with metrics and trace headers
func Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return doWithClient(ctx, req, http.DefaultClient)
}

// DoClient do http request, use client
// Wrap request with metrics and trace headers
func DoClient(ctx context.Context, req *http.Request, client *http.Client) (*http.Response, error) {
	return doWithClient(ctx, req, client)
}

func mergeTrace(ctx context.Context, res *http.Response) context.Context {
	if res == nil {
		return ctx
	}
	header := res.Header.Get(TraceparentHeader)
	if header == "" {
		return ctx
	}
	tid, sid, err := core.ParseTraceparent(header)
	if err != nil {
		return ctx
	}
	if core.TraceID(ctx) == "" {
		ctx = core.TraceWith(ctx, tid)
	}
	if core.SpanID(ctx) == "" {
		ctx = core.SpanWith(ctx, sid)
	}
	return ctx
}

// InheritTrace extract trace informations from http response and return a context
// with trace ID and span ID if them exist.
func InheritTrace(res *http.Response) context.Context {
	ctx := context.Background()
	return mergeTrace(ctx, res)
}

// FulfillTrace extract trace informations from http response and fulfill income
// context with trace or span IDs.
func FulfillTrace(ctx context.Context, res *http.Response) context.Context {
	return mergeTrace(ctx, res)
}
