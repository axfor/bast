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
	Verify(v *Validator, val Val, param string) (bool, bool, error)
}

//Validator a validate
type Validator struct {
}

//Val a validate value
type Val struct {
	Key    string
	Value  interface{}
	Real   reflect.Kind
	Expect reflect.Kind
}

//Struct verify that the a struct or each element is a struct int the slice or each element is a struct int the map
//data is validate data
func (c *Validator) Struct(data interface{}) error {
	v := reflect.ValueOf(data)
	ok := false
	kd := v.Kind()
	if kd == reflect.Ptr {
		v = v.Elem()
		kd = v.Kind()
	}
	if kd == reflect.Struct {
		return c.structVerify(v)
	} else if kd == reflect.Slice {
		lg := v.Len()
		for i := 0; i < lg; i++ {
			v2 := v.Index(i)
			err := c.Struct(v2.Interface())
			if err != nil {
				return err
			}
		}
		return nil
	} else if kd == reflect.Map {
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
		return fmt.Errorf("must be a struct or a struct pointer")
	}
	return nil
}

//structVerify verify struct
func (c *Validator) structVerify(v reflect.Value) error {
	numField := v.NumField()
	t := v.Type()
	for i := 0; i < numField; i++ {
		f := t.Field(i).Tag
		tag := f.Get("v")
		if tag == "" {
			continue
		}
		fv := v.Field(i)
		real := fv.Kind()
		var rv interface{} = nil
		if real == reflect.Ptr {
			fvv := fv.Elem()
			real = fvv.Kind()
			if fv.IsNil() {
				rv = nil
			} else {
				rv = fvv.Interface()
			}
		} else {
			rv = fv.Interface()
		}
		js := f.Get("json")
		tags := strings.Split(tag, "|")
		ks := strings.Split(js, ",")[0]
		val := Val{ks, rv, real, real}
		for _, tg := range tags {
			pos := strings.Index(tg, ":")
			fk := tg
			ps := ""
			if pos != -1 {
				fk = tg[0:pos]
				pos++
				ps = tg[pos:]
			}
			vf, ok := vfuncs[fk]
			if !ok {
				continue
			}
			if pass, next, err := vf.Verify(c, val, ps); !pass || !next {
				return err
			}
		}
	}
	return nil
}

//Request verify that the url.Values
//data is validate data
//rules is validate rule such as:
// 	key1=required|int|min:1
// 	key2=required|string|min:1|max:12
//	key3=sometimes|required|date
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
		expect := reflect.String
		if strings.Index(tag, "int") >= 0 {
			expect = reflect.Int
		} else if strings.Index(tag, "date") >= 0 {
			expect = Date
		} else if strings.Index(tag, "email") >= 0 {
			expect = Email
		}
		val := Val{k, nil, reflect.String, expect}
		for _, tg := range tags {
			pos := strings.Index(tg, ":")
			fk := tg
			ps := ""
			if pos != -1 {
				fk = tg[0:pos]
				pos++
				ps = tg[pos:]
			}
			vf, ok := vfuncs[fk]
			if !ok {
				continue
			}
			if !vok || lg <= 0 {
				val.Value = nil
				if pass, next, err := vf.Verify(c, val, ps); !pass || !next {
					return err
				}
				continue
			}
			for _, v := range vs {
				val.Value = v
				if pass, next, err := vf.Verify(c, val, ps); !pass || !next {
					return err
				}
			}
		}
	}
	return nil
}

//Register a validator provide by the vfuncs name
func Register(name string, vf VerifyFunc) {
	if _, ok := vfuncs[name]; !ok {
		vfuncs[name] = vf
	}
}
