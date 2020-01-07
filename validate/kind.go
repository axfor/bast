//Copyright 2018 The axx Authors. All rights reserved.

package validate

import "reflect"

// A Kind represents the specific kind of type that a Type represents.
// The zero Kind is not a valid kind.
const (
	Invalid = reflect.Invalid
	Bool    = reflect.Bool
	Int     = reflect.Int
	Int8    = reflect.Int8
	Int16   = reflect.Int16
	Int32   = reflect.Int32
	Int64   = reflect.Int64
	Uint    = reflect.Uint
	Uint8   = reflect.Uint8
	Uint16  = reflect.Uint16
	Uint32  = reflect.Uint32
	Uint64  = reflect.Uint64
	Float32 = reflect.Float32
	Float64 = reflect.Float64
	String  = reflect.String
	Date    = reflect.UnsafePointer + 255
	Email   = Date + 1
)
