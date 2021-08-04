package slog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type fileWriter struct {
	mux sync.RWMutex

	// DO NOT copy this logger
	noCopy noCopy

	fs    *os.File
	size  int
	timer *time.Timer

	Dir      string
	Filename string
	MaxSize  int
	Rotate   time.Duration
}

func newFileWriter(dir, filename string, max int, rotate time.Duration) (*fileWriter, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	var timer *time.Timer
	if rotate > 0 {
		timer = time.NewTimer(rotate)
	}
	logger := &fileWriter{
		Dir:      dir,
		Filename: filename,
		MaxSize:  max,
		Rotate:   rotate,

		timer: timer,
	}
	if err := logger.initFile(); err != nil {
		return nil, err
	}
	return logger, nil
}

const timestampFormat = "20060102150405"

func currentTimestamp() string {
	now := time.Now()
	return now.Format(timestampFormat)
}

func (fw *fileWriter) rotate() (err error) {
	defer func() {
		label := "ok"
		if err != nil {
			label = "error"
		}
		fileRotateCounter.WithLabelValues(label).Inc()
	}()
	// close original log file
	if err = fw.fs.Close(); err != nil {
		return err
	}

	backupFilename := fmt.Sprintf("%s-%s", fw.Filename, currentTimestamp())

	oldpath := filepath.Join(fw.Dir, fw.Filename)
	newpath := filepath.Join(fw.Dir, backupFilename)
	if err = os.Rename(oldpath, newpath); err != nil {
		return err
	}
	if err = fw.initFile(); err != nil {
		return err
	}
	fw.size = 0

	return nil
}

func (fw *fileWriter) resetTimer() {
	fw.timer = time.NewTimer(fw.Rotate)
}

func (fw *fileWriter) initFile() error {
	if err := os.MkdirAll(fw.Dir, 0755); err != nil {
		return err
	}
	filename := filepath.Join(fw.Dir, fw.Filename)
	fs, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fw.fs = fs
	return nil
}

func (fw *fileWriter) Write(data []byte) (n int, err error) {
	fw.mux.Lock()
	defer fw.mux.Unlock()

	defer func() {
		label := "ok"
		if err != nil {
			label = "error"
		}
		fileWriteCounter.WithLabelValues(label).Inc()
	}()

	if fw.fs == nil {
		if err = fw.initFile(); err != nil {
			return 0, err
		}
	}

	if fw.timer != nil {
		select {
		case <-fw.timer.C:
			if err = fw.rotate(); err != nil {
				return 0, err
			}
			fw.resetTimer()
		default:
		}
	}

	if fw.MaxSize > 0 && (fw.size+len(data)) > fw.MaxSize {
		if err = fw.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = fw.fs.Write(data)
	fw.size += n

	return n, err
}

func (fw *fileWriter) Sync() error {
	fw.mux.Lock()
	defer fw.mux.Unlock()
	return fw.fs.Sync()
}
