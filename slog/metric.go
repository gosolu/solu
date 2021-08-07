package slog

import "github.com/prometheus/client_golang/prometheus"

var (
	fileRotateCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "solu",
		Subsystem: "log",
		Name:      "file_rotate_counter",
		Help:      "File log rotate counter",
	}, []string{"state"})

	fileWriteCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "solu",
		Subsystem: "log",
		Name:      "file_write_counter",
		Help:      "File log write counter",
	}, []string{"state"})

	consoleWriteCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "solu",
		Subsystem: "log",
		Name:      "console_write_counter",
		Help:      "Console log counter",
	}, []string{"state"})
)

func init() {
	prometheus.MustRegister(
		fileRotateCounter,
		fileWriteCounter,
		consoleWriteCounter,
	)
}
