package log

import (
	"io"
	"os"
)

type consoleWriter struct {
	writer io.Writer
}

func newConsoleWriter() *consoleWriter {
	return &consoleWriter{
		writer: os.Stdout,
	}
}

func (cw *consoleWriter) Write(data []byte) (int, error) {
	n, err := cw.writer.Write(data)

	label := "ok"
	if err != nil {
		label = "error"
	}
	consoleWriteCounter.WithLabelValues(label).Inc()

	return n, err
}
