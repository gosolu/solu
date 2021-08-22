package slog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gosolu/solu/internal/core"
)

var (
	features = ""

	enableMetric = false
)

func init() {
	fs := strings.Split(features, ",")
	for _, f := range fs {
		switch strings.ToLower(f) {
		case "metric":
			enableMetric = true
		}
	}
}

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

func initLogger(level string, sampling *zap.SamplingConfig, syncers ...zapcore.WriteSyncer) (*zap.Logger, error) {
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

	if len(syncers) < 1 {
		return nil, errors.New("No output writer")
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        core.LogTimeKey,
		LevelKey:       core.LogLevelKey,
		CallerKey:      core.LogCallerKey,
		MessageKey:     core.LogMessageKey,
		StacktraceKey:  core.LogTraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	syncer := zapcore.NewMultiWriteSyncer(syncers...)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), syncer, logLevel)
	if sampling != nil {
		core = zapcore.NewSampler(core, time.Second, sampling.Initial, sampling.Thereafter)
	}
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zap.DPanicLevel))

	return logger, nil
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// Logger ...
type Logger struct {
	core *zap.Logger

	// some original configurations
	level         string
	dir, filename string
	maxFileSize   int
	fileRotate    fileRotateMode
	sampleFirst   int
	sampleAfter   int

	sinkStdout bool
	sinkFile   bool
}

var gLogger *Logger

func init() {
	core, err := initLogger(LogLevelDebug, nil, zapcore.AddSync(os.Stdout))
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

func newSpanID() string {
	var bts [8]byte
	_, err := io.ReadFull(rand.Reader, bts[:])
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bts[:])
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

// Option for logger
type Option func(*Logger) error

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

// Sample configure log sampling first and thereafter within time.Second
func Sample(first, thereafter int) Option {
	return func(logger *Logger) error {
		if first <= 0 {
			return fmt.Errorf("invalid sample first: %d", first)
		}
		logger.sampleFirst = first
		logger.sampleAfter = thereafter
		return nil
	}
}

// File option set logger's output file directory and filename
func File(dir, filename string, max int, rotate fileRotateMode) Option {
	return func(logger *Logger) error {
		if dir == "" {
			dir, _ = os.Getwd()
		}
		logger.dir = dir
		if filename != "" {
			logger.filename = filename
			logger.sinkFile = true
		}
		if max > 0 {
			logger.maxFileSize = max
		}
		if rotate > 0 {
			logger.fileRotate = rotate
		}
		return nil
	}
}

// Stdout option set logger output to stdout
func Stdout() Option {
	return func(logger *Logger) error {
		logger.sinkStdout = true
		return nil
	}
}

// Init logger
func Init(opts ...Option) (*Logger, error) {
	logger := new(Logger)
	for _, opt := range opts {
		if err := opt(logger); err != nil {
			return nil, err
		}
	}

	syncers := make([]zapcore.WriteSyncer, 0, 2)
	if logger.sinkFile {
		fw, err := newFileWriter(logger.dir, logger.filename, logger.maxFileSize, logger.fileRotate)
		if err != nil {
			return nil, err
		}
		syncers = append(syncers, fw)
	}
	if logger.sinkStdout {
		writer := newConsoleWriter()
		syncers = append(syncers, zapcore.AddSync(writer))
	}
	// sampling configuration
	var sampling *zap.SamplingConfig
	if logger.sampleFirst > 0 {
		sampling = &zap.SamplingConfig{
			Initial:    logger.sampleFirst,
			Thereafter: logger.sampleAfter,
		}
	}

	core, err := initLogger(logger.level, sampling, syncers...)
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

// Trace create a new context mixed with logger
func Trace(ctx context.Context) context.Context {
	id := newTraceID()
	return core.TraceWith(ctx, id)
}

// TraceWith create a new context with a given trace ID
func TraceWith(ctx context.Context, id string) context.Context {
	return core.TraceWith(ctx, id)
}

// Fork context will create a new span that inherit parent context's trace ID
func Fork(ctx context.Context) context.Context {
	if core.TraceID(ctx) == "" {
		ctx = Trace(ctx)
	}
	sid := newSpanID()
	return core.SpanWith(ctx, sid)
}

// In try extract logger instance from context
func In(ctx context.Context) *Logger {
	tid := core.TraceID(ctx)
	if tid == "" {
		tid = newTraceID()
	}
	fields := []Field{
		Str(core.LogTraceKey, tid),
	}
	sid := core.SpanID(ctx)
	if sid != "" {
		fields = append(fields, Str(core.LogSpanKey, sid))
	}
	return gLogger.With(fields...)
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
