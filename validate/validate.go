package validate

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

var vfuncs = map[string]VerifyFunc{}

//VerifyFunc is store engine interface
type VerifyFunc interface {
	Verify(v *Validator, key string, val interface{}, param ...string) (bool, bool)
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

//Struct verify that the a struct or each element is a struct int the slice or each element is a struct int the map
//data is validate data
func (c *Validator) Struct(data interface{}) error {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	ok := false
	if t, v, ok = getStruct(t, v); ok {
		return c.structVerify(t, v)
	}
	if t, v, ok = getSlice(t, v); ok {
		for i := 0; i < v.Len(); i++ {
			v2 := v.Index(i)
			err := c.Struct(v2.Interface())
			if err != nil {
				return err
			}
		}
		return nil
	}
	if t, v, ok = getMap(t, v); ok {
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
	if !ok {
		return fmt.Errorf("%v must be a struct or a struct pointer", t)
	}
	return nil
}

//Request verify that the url.Values
//data is validate data
//rules is validate rule such as:
// 	key=required|int|min:1
// 	key=required|string|min:1
//	key=sometimes|date|required
func (c *Validator) Request(data url.Values, rules ...string) error {
	if rules == nil || len(rules) <= 0 {
		return nil
	}
	for _, r := range rules {
		rs := strings.Split(r, "=")
		if len(rs) != 2 && rs[0] == "" {
			continue
		}
		k := rs[0]
		tag := rs[1]
		tags := strings.Split(tag, "|")
		vs, vok := data[k]
		lg := 0
		if vok {
			lg = len(vs)
		}
		for _, tg := range tags {
			tgs := strings.Split(tg, ":")
			vf, ok := vfuncs[tgs[0]]
			if !ok {
				continue
			}
			if tgs != nil || len(tgs) > 1 {
				tgs = tgs[1:]
			}
			if !vok || lg <= 0 {
				if pass, continues := vf.Verify(c, k, nil, tgs...); !pass || !continues {
					return c.Error()
				}
				continue
			}
			for _, v := range vs {
				if pass, continues := vf.Verify(c, k, v, tgs...); !pass || !continues {
					return c.Error()
				}
			}
		}
	}
	return nil
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
		if tag == "" {
			continue
		}
		tags := strings.Split(tag, "|")
		ks := strings.Split(js, ",")[0]
		for _, tg := range tags {
			tgs := strings.Split(tg, ":")
			vf, ok := vfuncs[tgs[0]]
			if !ok {
				continue
			}
			if tgs != nil || len(tgs) > 1 {
				tgs = tgs[1:]
			}
			if pass, continues := vf.Verify(c, ks, rv, tgs...); !pass || !continues {
				return c.Error()
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
