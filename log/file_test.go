package log

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileLoggerRotate(t *testing.T) {
	dir := os.TempDir()
	filename := fmt.Sprintf("filelog_%d.log", time.Now().Unix())
	max := 1000
	mode := rotateModeHourly
	logger, err := newFileLogger(dir, filename, max, mode)
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := logger.Write([]byte(filename)); err != nil {
		t.Error(err)
		return
	}
	if err := logger.rotate(); err != nil {
		t.Error(err)
		return
	}
	os.Remove(filepath.Join(dir, filename))
}

func TestFileLoggerMaxRotate(t *testing.T) {
	dir := os.TempDir()
	filename := fmt.Sprintf("filelog_%d.log", time.Now().Unix())
	max := 100
	mode := rotateModeHourly
	logger, err := newFileLogger(dir, filename, max, mode)
	if err != nil {
		t.Error(err)
		return
	}
	content := bytes.Repeat([]byte("a"), max)
	if _, err := logger.Write(content); err != nil {
		t.Error(err)
		return
	}
	content = bytes.Repeat([]byte("b"), 10)
	if _, err := logger.Write(content); err != nil {
		t.Error(err)
		return
	}
	if logger.size != 10 {
		t.Errorf("write more than max should trigger a rotate")
		return
	}
	logger.Sync()
	os.Remove(filepath.Join(dir, filename))
}

func TestFileLoggerWritter(t *testing.T) {
	dir := os.TempDir()
	filename := fmt.Sprintf("filelog_%d.log", time.Now().Unix())
	max := 2 << 10
	mode := rotateModeHourly
	logger, err := newFileLogger(dir, filename, max, mode)
	if err != nil {
		t.Error(err)
		return
	}
	if _, err := logger.Write([]byte("Hello")); err != nil {
		t.Error(err)
		return
	}
	if _, err := logger.Write([]byte("World")); err != nil {
		t.Error(err)
		return
	}
	if logger.size != 10 {
		t.Errorf("logger size is not correct!")
		return
	}
}
