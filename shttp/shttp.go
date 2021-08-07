// Package shttp contains some golang HTTP client methods
package shttp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
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

func Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return doWithClient(ctx, req, http.DefaultClient)
}

func DoClient(ctx context.Context, req *http.Request, client *http.Client) (*http.Response, error) {
	return doWithClient(ctx, req, client)
}
