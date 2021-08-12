package slog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResetRotateTime(t *testing.T) {
	tests := []struct {
		Mode   fileRotateMode
		Expect func() int64
	}{
		{FileRotateHourly, func() int64 {
			ts := time.Now()
			return time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour()+1, 0, 0, 0, time.Local).Unix()
		}},
		{FileRotateDaily, func() int64 {
			ts := time.Now().AddDate(0, 0, 1)
			return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.Local).Unix()
		}},
		{FileRotateWeekly, func() int64 {
			ts := time.Now()
			ts = ts.AddDate(0, 0, 7-int(ts.Weekday()))
			return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.Local).Unix()
		}},
		{FileRotateMonthly, func() int64 {
			ts := time.Now().AddDate(0, 1, 0)
			return time.Date(ts.Year(), ts.Month(), 0, 0, 0, 0, 0, time.Local).Unix()
		}},
	}

	for _, c := range tests {
		timestamp := resetRotateTime(c.Mode)
		if timestamp != c.Expect() {
			t.Errorf("Expect: %d, got: %d for mode %d", c.Expect(), timestamp, c.Mode)
			return
		}
	}
}

func TestFileWriterRotate(t *testing.T) {
	dir := os.TempDir()
	filename := fmt.Sprintf("file_%d.log", time.Now().Unix())
	max := 1000
	writer, err := newFileWriter(dir, filename, max, FileRotateNone)
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := writer.Write([]byte(filename)); err != nil {
		t.Error(err)
		return
	}
	if err := writer.rotate(); err != nil {
		t.Error(err)
		return
	}
	os.Remove(filepath.Join(dir, filename))
}

func TestFileWriterMaxRotate(t *testing.T) {
	dir := os.TempDir()
	filename := fmt.Sprintf("file_%d.log", time.Now().Unix())
	max := 100
	writer, err := newFileWriter(dir, filename, max, FileRotateDaily)
	if err != nil {
		t.Error(err)
		return
	}
	content := bytes.Repeat([]byte("a"), max)
	if _, err := writer.Write(content); err != nil {
		t.Error(err)
		return
	}
	content = bytes.Repeat([]byte("b"), 10)
	if _, err := writer.Write(content); err != nil {
		t.Error(err)
		return
	}
	if writer.size != 10 {
		t.Errorf("write more than max should trigger a rotate")
		return
	}
	writer.Sync()
	os.Remove(filepath.Join(dir, filename))
}

func TestFileWriterWritter(t *testing.T) {
	dir := os.TempDir()
	filename := fmt.Sprintf("file_%d.log", time.Now().Unix())
	max := 2 << 10
	writer, err := newFileWriter(dir, filename, max, FileRotateHourly)
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := writer.Write([]byte("Hello")); err != nil {
		t.Error(err)
		return
	}
	if _, err := writer.Write([]byte("World")); err != nil {
		t.Error(err)
		return
	}
	if writer.size != 10 {
		t.Errorf("writer's size is not correct!")
		return
	}
}
