package validate

import (
	"fmt"
	"reflect"
	"strings"
)

var vfuncs = map[string]VerifyFunc{}

//VerifyFunc is store engine interface
type VerifyFunc interface {
	Verify(v *Validator, key string, val interface{}) (bool, bool)
}

//Validator a validate
type Validator struct {
	err error
}

func (c *Validator) Error() error {
	return c.err
}

func (c *Validator) SetError(err error) {
	c.err = err
}

//Struct v
//data is validate data
//rules is validate rule such as:
// 	key=required|int|min:1
// 	key=required|string|min:1
//	key=sometimes|required|date
func (c *Validator) Struct(data interface{}) error {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	switch {
	case isStruct(t):
	case isStructPtr(t):
		t = t.Elem()
		v = v.Elem()
		break
	default:
		return fmt.Errorf("%v must be a struct or a struct pointer", t)
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		rv := fv.Interface()
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				rv = nil
			} else {
				rv = fv.Elem().Interface()
			}
		}
		tag := f.Tag.Get("v")
		js := f.Tag.Get("json")
		if tag != "" {
			tags := strings.Split(tag, "|")
			for _, tg := range tags {
				if vf, ok := vfuncs[tg]; ok {
					if hasErr, ok := vf.Verify(c, js, rv); !ok || hasErr {
						return c.Error()
					}
				}
			}
		}
	}
	return nil
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func isStructPtr(t reflect.Type) bool {
	return (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct)
}

//Register a validator provide by the vfuncs name
func Register(name string, vf VerifyFunc) {
	if _, ok := vfuncs[name]; !ok {
		vfuncs[name] = vf
	}
}
