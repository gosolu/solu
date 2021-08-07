package slog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
