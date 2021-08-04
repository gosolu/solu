package slog

import "testing"

func TestConsoleWriter(t *testing.T) {
	str := "Hello, world!"
	writer := newConsoleWriter()
	n, err := writer.Write([]byte(str))
	if err != nil {
		t.Error(err)
		return
	}
	if n != len(str) {
		t.Errorf("write failed, length not match")
		return
	}
}
