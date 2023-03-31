package slog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type fileRotateMode int

const (
	FileRotateNone fileRotateMode = iota
	FileRotateHourly
	FileRotateDaily
	FileRotateWeekly
	FileRotateMonthly
)

type fileWriter struct {
	mux sync.RWMutex

	// DO NOT copy this logger
	noCopy noCopy

	fs       *os.File
	size     int
	deadline int64

	Dir      string
	Filename string
	MaxSize  int
	Rotate   fileRotateMode
}

func resetRotateTime(mode fileRotateMode) int64 {
	now := time.Now()
	var next time.Time
	switch mode {
	case FileRotateHourly:
		next = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.Local)
	case FileRotateDaily:
		next = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	case FileRotateWeekly:
		nDays := 7 - now.Weekday()
		next = time.Date(now.Year(), now.Month(), now.Day()+int(nDays), 0, 0, 0, 0, time.Local)
	case FileRotateMonthly:
		next = time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.Local)
	case FileRotateNone:
		fallthrough
	default:
		return 0
	}
	return next.Unix()
}

func newFileWriter(dir, filename string, max int, rotate fileRotateMode) (*fileWriter, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	var deadline int64
	if rotate > 0 {
		deadline = resetRotateTime(rotate)
	}
	logger := &fileWriter{
		Dir:      dir,
		Filename: filename,
		MaxSize:  max,
		Rotate:   rotate,

		deadline: deadline,
	}
	if err := logger.initFile(); err != nil {
		return nil, err
	}
	return logger, nil
}

const timestampFormat = "20060102150405"

func timestampSuffix(t int64) string {
	return time.Unix(t, 0).Format(time.DateTime)
}

func (fw *fileWriter) rotate() (err error) {
	if enableMetric {
		defer func() {
			label := "ok"
			if err != nil {
				label = "error"
			}
			fileRotateCounter.WithLabelValues(label).Inc()
		}()
	}
	// close original log file
	if err = fw.fs.Close(); err != nil {
		return err
	}

	backupFilename := fmt.Sprintf("%s-%s", fw.Filename, timestampSuffix(fw.deadline))

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
	fw.deadline = resetRotateTime(fw.Rotate)
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

	if enableMetric {
		defer func() {
			label := "ok"
			if err != nil {
				label = "error"
			}
			fileWriteCounter.WithLabelValues(label).Inc()
		}()
	}

	if fw.fs == nil {
		if err = fw.initFile(); err != nil {
			return 0, err
		}
	}

	if fw.deadline > 0 {
		now := time.Now()
		if now.Unix() > fw.deadline {
			if err = fw.rotate(); err != nil {
				return 0, err
			}
			fw.resetTimer()
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
