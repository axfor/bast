//Copyright 2018 The axx Authors. All rights reserved.

package logs

import (
	"go.uber.org/zap"
)

// Err is shorthand for the common idiom NamedError("error", err).
func Err(err error) zap.Field {
	return zap.Error(err)
}

// NamedError constructs a field that lazily stores err.Error() under the
// provided key. Errors which also implement fmt.Formatter (like those produced
// by github.com/pkg/errors) will also have their verbose representation stored
// under key+"Verbose". If passed a nil error, the field is a no-op.
//
// For the common case in which the key is simply "error", the Error function
// is shorter and less repetitive.
func NamedError(key string, err error) zap.Field {
	return zap.NamedError(key, err)
}
