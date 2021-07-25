package log

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

const (
	traceKeyName = "g_tid"

	timeKeyName       = "g_ts"
	levelKeyName      = "g_level"
	callerKeyName     = "g_caller"
	messageKeyName    = "g_msg"
	stacktraceKeyName = "g_stack"
)

func initLogger(level string, fw zapcore.WriteSyncer, stdout bool) (*zap.Logger, error) {
	var logLevel zapcore.Level
	switch strings.ToLower(level) {
	case LogLevelDebug:
		logLevel = zapcore.DebugLevel
	case LogLevelInfo:
		logLevel = zapcore.InfoLevel
	case LogLevelWarn:
		logLevel = zapcore.WarnLevel
	case LogLevelError:
		logLevel = zapcore.ErrorLevel
	case LogLevelFatal:
		logLevel = zapcore.FatalLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	if stdout {
		if isNil(fw) {
			fw = zapcore.AddSync(os.Stdout)
		} else {
			fw = zapcore.NewMultiWriteSyncer(fw, os.Stdout)
		}
	}
	if isNil(fw) {
		return nil, errors.New("No output writer")
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        timeKeyName,
		LevelKey:       levelKeyName,
		CallerKey:      callerKeyName,
		MessageKey:     messageKeyName,
		StacktraceKey:  stacktraceKeyName,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fw, logLevel)
	samplerCore := zapcore.NewSampler(core, time.Second, 100, 100)
	logger := zap.New(samplerCore, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.DPanicLevel))

	return logger, nil
}

func isNil(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// Logger ...
type Logger struct {
	core *zap.Logger

	// some original configurations
	dir, filename string
	level         string

	stdoutLog bool
	fileLog   bool

	fw *fileLogger // file writer
}

var gLogger *Logger

func init() {
	core, err := initLogger(LogLevelInfo, nil, true)
	if err != nil {
		panic(err)
	}
	gLogger = &Logger{
		core: core,
	}
}

func newTraceID() string {
	var bts [16]byte
	_, err := io.ReadFull(rand.Reader, bts[:])
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bts[:])
}

// Option for logger
type Option func(*Logger) error

// File option set logger's output file directory and filename
func File(dir, filename string) Option {
	return func(logger *Logger) error {
		if dir != "" {
			logger.dir = dir
		}
		if filename != "" {
			logger.filename = filename
		}
		return nil
	}
}

var allowedLevels = []string{
	LogLevelDebug,
	LogLevelInfo,
	LogLevelWarn,
	LogLevelError,
	LogLevelFatal,
}

var (
	ErrInvalidLevel = errors.New("Invalid log level")
)

// Level option set logger's log level
func Level(lvl string) Option {
	return func(logger *Logger) error {
		var valid bool
		for _, al := range allowedLevels {
			if lvl == al {
				valid = true
				break
			}
		}
		if !valid {
			return ErrInvalidLevel
		}
		logger.level = lvl
		return nil
	}
}

// Stdout option set logger output to stdout
func Stdout() Option {
	return func(logger *Logger) error {
		logger.stdoutLog = true
		return nil
	}
}

// New a logger
func New(opts ...Option) (*Logger, error) {
	logger := new(Logger)
	for _, opt := range opts {
		if err := opt(logger); err != nil {
			return nil, err
		}
	}

	var fw *fileLogger
	if logger.fileLog {
		var err error
		if logger.filename != "" {
			fw, err = newFileLogger(logger.dir, logger.filename, 0, rotateModeNone)
			if err != nil {
				return nil, err
			}
			logger.fw = fw
		}
	}

	core, err := initLogger(logger.level, fw, logger.stdoutLog)
	if err != nil {
		return nil, err
	}
	logger.core = core

	// Replace gLogger with current new logger
	gLogger = logger

	return logger, nil
}

// clone a new logger
func (log *Logger) clone() *Logger {
	cp := *log
	return &cp
}

type traceType struct{}

var traceKey traceType

// Trace create a new context mixed with logger
func Trace(ctx context.Context) context.Context {
	id := newTraceID()
	return context.WithValue(ctx, traceKey, id)
}

// TraceWith create a new context with a given trace ID
func TraceWith(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceKey, id)
}

// TraceID returns the current context trace ID
func TraceID(ctx context.Context) string {
	val := ctx.Value(traceKey)
	if val != nil {
		if value, ok := val.(string); ok {
			return value
		}
	}
	return ""
}

// global fork index
var (
	forkIndex uint64
)

// Fork context that inherit parent context's trace ID
func Fork(ctx context.Context) context.Context {
	var tid string
	val := ctx.Value(traceKey)
	if val != nil {
		if v, ok := val.(string); ok {
			tid = v
		}
	}
	pexists := true
	if tid == "" {
		tid = newTraceID()
		pexists = false
	}
	if pexists {
		idx := atomic.AddUint64(&forkIndex, 1)
		tid = fmt.Sprintf("%s:%d", tid, idx)
	}
	return context.WithValue(ctx, traceKey, tid)
}

// In try extract logger instance from context
func In(ctx context.Context) *Logger {
	val := ctx.Value(traceKey)
	if val == nil {
		return gLogger.With(String(traceKeyName, newTraceID()))
	}
	v, ok := val.(string)
	if !ok {
		return gLogger.With(String(traceKeyName, newTraceID()))
	}
	return gLogger.With(String(traceKeyName, v))
}

// With fields
func (log *Logger) With(fields ...Field) *Logger {
	l := log.clone()
	l.core = log.core.With(fields...)
	return l
}

// Named create a named logger
func (log *Logger) Named(name string) *Logger {
	l := log.clone()
	l.core = log.core.Named(name)
	return l
}

// Debug log
func (log *Logger) Debug(msg string) {
	log.core.Debug(msg)
}

// Info log
func (log *Logger) Info(msg string) {
	log.core.Info(msg)
}

func (log *Logger) Infof(template string, args ...interface{}) {
	log.core.Sugar().Infof(template, args...)
}

// Warn log
func (log *Logger) Warn(msg string) {
	log.core.Warn(msg)
}

func (log *Logger) Warnf(template string, args ...interface{}) {
	log.core.Sugar().Warnf(template, args...)
}

// Error log
func (log *Logger) Error(msg string) {
	log.core.Error(msg)
}

// DPanic log
func (log *Logger) DPanic(msg string) {
	log.core.DPanic(msg)
}

// Panic log
func (log *Logger) Panic(msg string) {
	log.core.Panic(msg)
}

// Fatal log
func (log *Logger) Fatal(msg string) {
	log.core.Fatal(msg)
}

// Sync flush buffered logs
func (log *Logger) Sync() error {
	return log.core.Sync()
}

// With zap fields
func With(fileds ...Field) *Logger {
	return gLogger.With(fileds...)
}

// Print log
func Print(msg string) {
	gLogger.Info(msg)
}

// Printf log
func Printf(template string, args ...interface{}) {
	gLogger.Infof(template, args...)
}

// Println log
func Println(msg string) {
	gLogger.Info(msg)
}

// Fatal log
func Fatal(msg string) {
	gLogger.Fatal(msg)
}

// Fatalf log
func Fatalf(template string, args ...interface{}) {
	gLogger.core.Sugar().Fatalf(template, args...)
}

// Fatalw log
func Fatalw(msg string, keysAndValues ...interface{}) {
	gLogger.core.Sugar().Fatalw(msg, keysAndValues...)
}

// Fatalln log
func Fatalln(args ...interface{}) {
	gLogger.core.Sugar().Fatal(args...)
}

// Panic log
func Panic(msg string) {
	gLogger.Panic(msg)
}

// Panicf log
func Panicf(template string, args ...interface{}) {
	gLogger.core.Sugar().Panicf(template, args...)
}

// Panicw log
func Panicw(msg string, keysAndValues ...interface{}) {
	gLogger.core.Sugar().Panicw(msg, keysAndValues...)
}

// Debug log
func Debug(msg string) {
	gLogger.Debug(msg)
}

// Debugf log
func Debugf(template string, args ...interface{}) {
	gLogger.core.Sugar().Debugf(template, args...)
}

// Debugw log
func Debugw(msg string, keysAndValues ...interface{}) {
	gLogger.core.Sugar().Debugw(msg, keysAndValues...)
}

// Info log
func Info(msg string) {
	gLogger.Info(msg)
}

// Infof log
func Infof(template string, args ...interface{}) {
	gLogger.Infof(template, args...)
}

// Infow log
func Infow(msg string, keysAndValues ...interface{}) {
	gLogger.core.Sugar().Infow(msg, keysAndValues...)
}

// Warn log
func Warn(msg string) {
	gLogger.Warn(msg)
}

// Warnf log
func Warnf(template string, args ...interface{}) {
	gLogger.Warnf(template, args...)
}

// Warnw log
func Warnw(msg string, keysAndValues ...interface{}) {
	gLogger.core.Sugar().Warnw(msg, keysAndValues...)
}

// Error log
func Error(msg string) {
	gLogger.Error(msg)
}

// Errorf log
func Errorf(template string, args ...interface{}) {
	gLogger.core.Sugar().Errorf(template, args...)
}

// Errorw log
func Errorw(msg string, keysAndValues ...interface{}) {
	gLogger.core.Sugar().Errorw(msg, keysAndValues...)
}
