//Copyright 2018 The axx Authors. All rights reserved.

package validate

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/axfor/bast/lang"
)

var vfuncs = map[string]VerifyFunc{}

//VerifyFunc is verify interface
type VerifyFunc interface {
	Verify(v *Validator, val Val) (bool, bool, error)
}

//Verify is customer verify interface
type Verify interface {
	Verify(v *Validator) error
}

//Validator a validate
type Validator struct {
	Lang string //default en
}

//Val a validate value
type Val struct {
	Key, TranKey, Param string
	Real, Expect        reflect.Kind
	Value               interface{}
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
		ks := ""
		pos := 0
		if js != "" {
			pos = strings.Index(js, ",")
			if pos != -1 {
				ks = js[0:pos]
			}
		}
		pos = strings.Index(tag, "@")
		tk := ""
		split := "|"
		if pos != -1 {
			tk = tag[0:pos]
			tag = tag[pos+1:]
			pos = strings.Index(tk, "/")
			if pos != -1 {
				split = tk[pos+1:]
				tk = tk[0:pos]
			}
		} else {
			tk = ks
		}
		tags := strings.Split(tag, split)
		val := Val{ks, lang.Transk(c.Lang, tk), "", real, real, rv}
		for _, tg := range tags {
			pos := strings.Index(tg, ":")
			fk := tg
			ps := ""
			if pos != -1 {
				fk = tg[0:pos]
				pos++
				ps = tg[pos:]
			}
			val.Param = ps
			vf, ok := vfuncs[fk]
			if !ok {
				continue
			}
			if pass, next, err := vf.Verify(c, val); !pass || !next {
				if err != nil {
					return err
				} else if !next {
					break
				}
			}
		}
	}
	if f, ok := v.Interface().(Verify); ok {
		return f.Verify(c)
	}
	return nil
}

//Request verify that the url.Values
//data is validate data
//rules is validate rule such as:
// 	key1@required|int|min:1
// 	key2/key2_translator@required|string|min:1|max:12
//	key3@sometimes|required|date
func (c *Validator) Request(data url.Values, rules ...string) error {
	if rules == nil || len(rules) <= 0 {
		return nil
	}
	for _, r := range rules {
		k, tag := "", ""
		split := "|"
		pos := strings.Index(r, "@")
		if pos != -1 {
			k = r[0:pos]
			tag = r[pos+1:]
			pos = strings.Index(k, ",")
			if pos != -1 {
				split = k[pos+1:]
				k = k[0:pos]
			}
		} else {
			continue
		}
		tk := ""
		pos = strings.Index(k, "/")
		if pos != -1 {
			tk = k[pos+1:]
			k = k[0:pos]
		} else {
			tk = k
		}
		tags := strings.Split(tag, split)
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
		val := Val{k, lang.Transk(c.Lang, tk), "", reflect.String, expect, nil}
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
				val.Param = ""
				if pass, next, err := vf.Verify(c, val); !pass || !next {
					if err != nil {
						return err
					} else if !next {
						break
					}
				}
				continue
			}
			for _, v := range vs {
				val.Value = v
				val.Param = ps
				if pass, next, err := vf.Verify(c, val); !pass || !next {
					if err != nil {
						return err
					} else if !next {
						break
					}
				}
			}
		}
	}
	return nil
}

//Trans translator
func (c *Validator) Trans(key string, param ...string) string {
	return lang.Trans(c.Lang, "v."+key, param...)
}

//Register a validator provide by the vfuncs name
func Register(name string, vf VerifyFunc) {
	if _, ok := vfuncs[name]; !ok {
		vfuncs[name] = vf
	}
}
