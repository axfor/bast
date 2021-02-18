//Copyright 2018 The axx Authors. All rights reserved.

package logs

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Skip constructs a no-op field, which is often useful when handling invalid
// inputs in other Field constructors.
func Skip() zap.Field {
	return zap.Skip()
}

// Binary constructs a field that carries an opaque binary blob.
//
// Binary data is serialized in an encoding-appropriate format. For example,
// zap's JSON encoder base64-encodes binary blobs. To log UTF-8 encoded text,
// use ByteString.
func Binary(key string, val []byte) zap.Field {
	return zap.Binary(key, val)
}

// Bool constructs a field that carries a bool.
func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// Boolp constructs a field that carries a *bool. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Boolp(key string, val *bool) zap.Field {
	return zap.Boolp(key, val)
}

// ByteString constructs a field that carries UTF-8 encoded text as a []byte.
// To log opaque binary blobs (which aren't necessarily valid UTF-8), use
// Binary.
func ByteString(key string, val []byte) zap.Field {
	return zap.ByteString(key, val)
}

// Complex128 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex128 to
// interface{}).
func Complex128(key string, val complex128) zap.Field {
	return zap.Complex128(key, val)
}

// Complex128p constructs a field that carries a *complex128. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Complex128p(key string, val *complex128) zap.Field {
	return zap.Complex128p(key, val)
}

// Complex64 constructs a field that carries a complex number. Unlike most
// numeric fields, this costs an allocation (to convert the complex64 to
// interface{}).
func Complex64(key string, val complex64) zap.Field {
	return zap.Complex64(key, val)
}

// Complex64p constructs a field that carries a *complex64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Complex64p(key string, val *complex64) zap.Field {
	return zap.Complex64p(key, val)
}

// Float64 constructs a field that carries a float64. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

// Float64p constructs a field that carries a *float64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Float64p(key string, val *float64) zap.Field {
	return zap.Float64p(key, val)
}

// Float32 constructs a field that carries a float32. The way the
// floating-point value is represented is encoder-dependent, so marshaling is
// necessarily lazy.
func Float32(key string, val float32) zap.Field {
	return zap.Float32(key, val)
}

// Float32p constructs a field that carries a *float32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Float32p(key string, val *float32) zap.Field {
	return zap.Float32p(key, val)
}

// Int constructs a field with the given key and value.
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Intp constructs a field that carries a *int. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Intp(key string, val *int) zap.Field {
	return zap.Intp(key, val)
}

// Int64 constructs a field with the given key and value.
func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Int64p constructs a field that carries a *int64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int64p(key string, val *int64) zap.Field {
	return zap.Int64p(key, val)

}

// Int32 constructs a field with the given key and value.
func Int32(key string, val int32) zap.Field {
	return zap.Int32(key, val)

}

// Int32p constructs a field that carries a *int32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int32p(key string, val *int32) zap.Field {
	return zap.Int32p(key, val)

}

// Int16 constructs a field with the given key and value.
func Int16(key string, val int16) zap.Field {
	return zap.Int16(key, val)

}

// Int16p constructs a field that carries a *int16. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int16p(key string, val *int16) zap.Field {
	return zap.Int16p(key, val)
}

// Int8 constructs a field with the given key and value.
func Int8(key string, val int8) zap.Field {
	return zap.Int8(key, val)
}

// Int8p constructs a field that carries a *int8. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Int8p(key string, val *int8) zap.Field {
	return zap.Int8p(key, val)

}

// String constructs a field with the given key and value.
func String(key string, val string) zap.Field {
	return zap.String(key, val)

}

// Stringp constructs a field that carries a *string. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Stringp(key string, val *string) zap.Field {
	return zap.Stringp(key, val)

}

// Uint constructs a field with the given key and value.
func Uint(key string, val uint) zap.Field {
	return zap.Uint(key, val)

}

// Uintp constructs a field that carries a *uint. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uintp(key string, val *uint) zap.Field {
	return zap.Uintp(key, val)

}

// Uint64 constructs a field with the given key and value.
func Uint64(key string, val uint64) zap.Field {
	return zap.Uint64(key, val)

}

// Uint64p constructs a field that carries a *uint64. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint64p(key string, val *uint64) zap.Field {
	return zap.Uint64p(key, val)

}

// Uint32 constructs a field with the given key and value.
func Uint32(key string, val uint32) zap.Field {
	return zap.Uint32(key, val)

}

// Uint32p constructs a field that carries a *uint32. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint32p(key string, val *uint32) zap.Field {
	return zap.Uint32p(key, val)

}

// Uint16 constructs a field with the given key and value.
func Uint16(key string, val uint16) zap.Field {
	return zap.Uint16(key, val)
}

// Uint16p constructs a field that carries a *uint16. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint16p(key string, val *uint16) zap.Field {
	return zap.Uint16p(key, val)

}

// Uint8 constructs a field with the given key and value.
func Uint8(key string, val uint8) zap.Field {
	return zap.Uint8(key, val)

}

// Uint8p constructs a field that carries a *uint8. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uint8p(key string, val *uint8) zap.Field {
	return zap.Uint8p(key, val)
}

// Uintptr constructs a field with the given key and value.
func Uintptr(key string, val uintptr) zap.Field {
	return zap.Uintptr(key, val)

}

// Uintptrp constructs a field that carries a *uintptr. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Uintptrp(key string, val *uintptr) zap.Field {
	return zap.Uintptrp(key, val)
}

// Reflect constructs a field with the given key and an arbitrary object. It uses
// an encoding-appropriate, reflection-based function to lazily serialize nearly
// any object into the logging context, but it's relatively slow and
// allocation-heavy. Outside tests, Any is always a better choice.
//
// If encoding fails (e.g., trying to serialize a map[int]string to JSON), Reflect
// includes the error message in the final log output.
func Reflect(key string, val interface{}) zap.Field {
	return zap.Reflect(key, val)

}

// Namespace creates a named, isolated scope within the logger's context. All
// subsequent fields will be added to the new namespace.
//
// This helps prevent key collisions when injecting loggers into sub-components
// or third-party libraries.
func Namespace(key string) zap.Field {
	return zap.Namespace(key)
}

// Stringer constructs a field with the given key and the output of the value's
// String method. The Stringer's String method is called lazily.
func Stringer(key string, val fmt.Stringer) zap.Field {
	return zap.Stringer(key, val)
}

// Time constructs a Field with the given key and value. The encoder
// controls how the time is serialized.
func Time(key string, val time.Time) zap.Field {
	return zap.Time(key, val)

}

// Timep constructs a field that carries a *time.Time. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Timep(key string, val *time.Time) zap.Field {
	return zap.Timep(key, val)

}

// Stack constructs a field that stores a stacktrace of the current goroutine
// under provided key. Keep in mind that taking a stacktrace is eager and
// expensive (relatively speaking); this function both makes an allocation and
// takes about two microseconds.
func Stack(key string) zap.Field {
	return zap.Stack(key)

}

// Duration constructs a field with the given key and value. The encoder
// controls how the duration is serialized.
func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)

}

// Durationp constructs a field that carries a *time.Duration. The returned Field will safely
// and explicitly represent `nil` when appropriate.
func Durationp(key string, val *time.Duration) zap.Field {
	return zap.Durationp(key, val)

}

// Object constructs a field with the given key and ObjectMarshaler. It
// provides a flexible, but still type-safe and efficient, way to add map- or
// struct-like user-defined types to the logging context. The struct's
// MarshalLogObject method is called lazily.
func Object(key string, val zapcore.ObjectMarshaler) zap.Field {
	return zap.Object(key, val)
}

// Any takes a key and an arbitrary value and chooses the best way to represent
// them as a field, falling back to a reflection-based approach only if
// necessary.
//
// Since byte/uint8 and rune/int32 are aliases, Any can't differentiate between
// them. To minimize surprises, []byte values are treated as binary blobs, byte
// values are treated as uint8, and runes are always treated as integers.
func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// NewEntryCaller makes an EntryCaller from the return signature of
// runtime.Caller.
func NewEntryCaller(pc uintptr, file string, line int, ok bool) zapcore.EntryCaller {
	return zapcore.NewEntryCaller(pc, file, line, ok)
}
