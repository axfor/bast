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

//Error get error
func (c *Validator) Error() error {
	return c.err
}

//SetError set error
func (c *Validator) SetError(err error) {
	c.err = err
}

//Struct v
//data is validate data
//rules is validate rule such as:
// 	key=required|int|min:1
// 	key=required|min:1
//	key=sometimes|required
func (c *Validator) Struct(data interface{}) error {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	ok := false
	t, v, ok = getStruct(t, v)
	if !ok { //not struct
		t, v, ok = getSlice(t, v)
		if ok { //is slice
			for i := 0; i < v.Len(); i++ {
				v2 := v.Index(i)
				err := c.Struct(v2.Interface())
				if err != nil {
					return err
				}
			}
			return nil
		}
		t, v, ok = getMap(t, v)
		if ok { //is map
			iter := v.MapRange()
			for iter.Next() {
				v2 := iter.Value()
				err := c.Struct(v2.Interface())
				if err != nil {
					return err
				}
			}
			return nil
		}
	}
	if !ok {
		return fmt.Errorf("%v must be a struct or a struct pointer", t)
	}
	err := c.structVerify(t, v) //verify struct
	return err
}

//structVerify verify struct
func (c *Validator) structVerify(t reflect.Type, v reflect.Value) error {
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
			ks := strings.Split(js, ",")[0]
			for _, tg := range tags {
				if vf, ok := vfuncs[tg]; ok {
					if hasErr, ok := vf.Verify(c, ks, rv); !ok || hasErr {
						return c.Error()
					}
				}
			}
		}
	}
	return nil
}

func getStruct(t reflect.Type, v reflect.Value) (reflect.Type, reflect.Value, bool) {
	ok := false
	if ok = isStruct(t); ok {
	} else if ok = isStructPtr(t); ok {
		t = t.Elem()
		v = v.Elem()
	}
	return t, v, ok
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func isStructPtr(t reflect.Type) bool {
	return (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct)
}

func getSlice(t reflect.Type, v reflect.Value) (reflect.Type, reflect.Value, bool) {
	ok := false
	if ok = isSlice(t); ok {
	} else if ok = isSlicePtr(t); ok {
		t = t.Elem()
		v = v.Elem()
	}
	return t, v, ok
}

func isSlice(t reflect.Type) bool {
	return t.Kind() == reflect.Slice
}

func isSlicePtr(t reflect.Type) bool {
	return (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Slice)
}

func getMap(t reflect.Type, v reflect.Value) (reflect.Type, reflect.Value, bool) {
	ok := false
	if ok = isMap(t); ok {
	} else if ok = isMapPtr(t); ok {
		t = t.Elem()
		v = v.Elem()
	}
	return t, v, ok
}

func isMap(t reflect.Type) bool {
	return t.Kind() == reflect.Map
}

func isMapPtr(t reflect.Type) bool {
	return (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Map)
}

//Register a validator provide by the vfuncs name
func Register(name string, vf VerifyFunc) {
	if _, ok := vfuncs[name]; !ok {
		vfuncs[name] = vf
	}
}
