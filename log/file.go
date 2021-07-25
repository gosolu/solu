package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type rotateMode int

const (
	rotateModeNone   rotateMode = iota
	rotateModeHourly            = iota
	rotateModeDaily
	rotateModeWeekly
)

type fileLogger struct {
	mux sync.RWMutex

	// DO NOT copy this logger
	noCopy noCopy

	fs    *os.File
	size  int
	timer *time.Timer

	Dir      string
	Filename string
	MaxSize  int
	Rotate   rotateMode
}

func newFileLogger(dir, filename string, max int, rotate rotateMode) (*fileLogger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	logger := &fileLogger{
		Dir:      dir,
		Filename: filename,
		MaxSize:  max,
		Rotate:   rotate,
	}
	return logger, nil
}

const timestampFormat = "20060102150405"

func currentTimestamp() string {
	now := time.Now()
	return now.Format(timestampFormat)
}

func (fl *fileLogger) rotate() error {
	// close original log file
	if err := fl.fs.Close(); err != nil {
		return err
	}

	backupFilename := fmt.Sprintf("%s-%s", fl.Filename, currentTimestamp())

	oldpath := filepath.Join(fl.Dir, fl.Filename)
	newpath := filepath.Join(fl.Dir, backupFilename)
	if err := os.Rename(oldpath, newpath); err != nil {
		return err
	}
	if err := fl.initFile(); err != nil {
		return err
	}
	fl.size = 0
	return nil
}

func (fl *fileLogger) resetTimer() {
	switch fl.Rotate {
	case rotateModeHourly:
		fl.timer = time.NewTimer(time.Hour)
	case rotateModeDaily:
		fl.timer = time.NewTimer(24 * time.Hour)
	case rotateModeWeekly:
		fl.timer = time.NewTimer(7 * 24 * time.Hour)
	default:
		fl.timer = nil
	}
}

func (fl *fileLogger) initFile() error {
	if err := os.MkdirAll(fl.Dir, 0755); err != nil {
		return err
	}
	filename := filepath.Join(fl.Dir, fl.Filename)
	fs, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	fl.fs = fs
	return nil
}

func (fl *fileLogger) Write(data []byte) (int, error) {
	fl.mux.Lock()
	defer fl.mux.Unlock()

	if fl.fs == nil {
		if err := fl.initFile(); err != nil {
			return 0, err
		}
	}

	if fl.timer != nil {
		select {
		case <-fl.timer.C:
			if err := fl.rotate(); err != nil {
				return 0, err
			}
			fl.resetTimer()
		default:
		}
	}

	if fl.MaxSize > 0 && (fl.size+len(data)) > fl.MaxSize {
		if err := fl.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := fl.fs.Write(data)
	fl.size += n
	return n, err
}

func (fl *fileLogger) Sync() error {
	fl.mux.Lock()
	defer fl.mux.Unlock()
	return fl.fs.Sync()
}
