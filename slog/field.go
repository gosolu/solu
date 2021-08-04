package slog

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field ...
type Field = zap.Field

// Alias
var (
	B  = Bool
	Bp = Boolp

	T  = Time
	Tp = Timep

	I    = Int
	Ip   = Intp
	I8   = Int8
	I8p  = Int8p
	I16  = Int16
	I16p = Int16p
	I32  = Int32
	I32p = Int32p
	I64  = Int64
	I64p = Int64p

	U    = Uint
	Up   = Uintp
	U8   = Uint8
	U8p  = Uint8p
	U16  = Uint16
	U16p = Uint16p
	U32  = Uint32
	U32p = Uint32p
	U64  = Uint64
	U64p = Uint64p

	F    = Float32
	F32  = Float32
	F32p = Float32p
	F64  = Float64
	F64p = Float64p

	Str = zap.String
)

// Skip field
func Skip() Field {
	return zap.Skip()
}

// Binary field
func Binary(key string, val []byte) Field {
	return zap.Binary(key, val)
}

// Bool field
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Boolp field
func Boolp(key string, val *bool) Field {
	return zap.Boolp(key, val)
}

// ByteString field
func ByteString(key string, val []byte) Field {
	return zap.ByteString(key, val)
}

// Complex128 field
func Complex128(key string, val complex128) Field {
	return zap.Complex128(key, val)
}

// Complex128p field
func Complex128p(key string, val *complex128) Field {
	return zap.Complex128p(key, val)
}

// Complex64 field
func Complex64(key string, val complex64) Field {
	return zap.Complex64(key, val)
}

// Complex64p field
func Complex64p(key string, val *complex64) Field {
	return zap.Complex64p(key, val)
}

// Float32 field
func Float32(key string, val float32) Field {
	return zap.Float32(key, val)
}

// Float32p field
func Float32p(key string, val *float32) Field {
	return zap.Float32p(key, val)
}

// Float64 field
func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

// Float64p field
func Float64p(key string, val *float64) Field {
	return zap.Float64p(key, val)
}

// Int field
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Intp field
func Intp(key string, val *int) Field {
	return zap.Intp(key, val)
}

// Int8 field
func Int8(key string, val int8) Field {
	return zap.Int8(key, val)
}

// Int8p field
func Int8p(key string, val *int8) Field {
	return zap.Int8p(key, val)
}

// Int16 field
func Int16(key string, val int16) Field {
	return zap.Int16(key, val)
}

// Int16p field
func Int16p(key string, val *int16) Field {
	return zap.Int16p(key, val)
}

// Int32 field
func Int32(key string, val int32) Field {
	return zap.Int32(key, val)
}

// Int32p field
func Int32p(key string, val *int32) Field {
	return zap.Int32p(key, val)
}

// Int64 field
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Int64p field
func Int64p(key string, val *int64) Field {
	return zap.Int64p(key, val)
}

// Stringp field
func Stringp(key string, val *string) Field {
	return zap.Stringp(key, val)
}

// Uint field
func Uint(key string, val uint) Field {
	return zap.Uint(key, val)
}

// Uintp field
func Uintp(key string, val *uint) Field {
	return zap.Uintp(key, val)
}

// Uint8 field
func Uint8(key string, val uint8) Field {
	return zap.Uint8(key, val)
}

// Uint8p field
func Uint8p(key string, val *uint8) Field {
	return zap.Uint8p(key, val)
}

// Uint16 field
func Uint16(key string, val uint16) Field {
	return zap.Uint16(key, val)
}

// Uint16p field
func Uint16p(key string, val *uint16) Field {
	return zap.Uint16p(key, val)
}

// Uint32 field
func Uint32(key string, val uint32) Field {
	return zap.Uint32(key, val)
}

// Uint32p field
func Uint32p(key string, val *uint32) Field {
	return zap.Uint32p(key, val)
}

// Uint64 field
func Uint64(key string, val uint64) Field {
	return zap.Uint64(key, val)
}

// Uint64p field
func Uint64p(key string, val *uint64) Field {
	return zap.Uint64p(key, val)
}

// Uintptr field
func Uintptr(key string, val uintptr) Field {
	return zap.Uintptr(key, val)
}

// Uintptrp field
func Uintptrp(key string, val *uintptr) Field {
	return zap.Uintptrp(key, val)
}

// Reflect field
func Reflect(key string, val interface{}) Field {
	return zap.Reflect(key, val)
}

// Namespace field
func Namespace(key string) Field {
	return zap.Namespace(key)
}

// Stringer field
func Stringer(key string, val fmt.Stringer) Field {
	return zap.Stringer(key, val)
}

// Time field
func Time(key string, val time.Time) Field {
	return zap.Time(key, val)
}

// Timep field
func Timep(key string, val *time.Time) Field {
	return zap.Timep(key, val)
}

// Stack fiedl
func Stack(key string) Field {
	return zap.Stack(key)
}

// Duration field
func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

// Durationp field
func Durationp(key string, val *time.Duration) Field {
	return zap.Durationp(key, val)
}

// Object field
func Object(key string, val zapcore.ObjectMarshaler) Field {
	return zap.Object(key, val)
}

// String field
func String(key, val string) Field {
	return zap.String(key, val)
}

// Any field
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

// Err is shorthand for the common idiom NamedError("error", err).
func Err(err error) Field {
	return NamedError("error", err)
}

// NamedError field
func NamedError(key string, err error) Field {
	if err == nil {
		return Skip()
	}
	return Field{Key: key, Type: zapcore.ErrorType, Interface: err}
}
