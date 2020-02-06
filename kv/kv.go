//Copyright 2018 The axx Authors. All rights reserved.

package kv

import (
	"fmt"
	"sort"
	"strings"
)

//Stringer String()
type Stringer interface {
	String() string
}

//Any key value
//support url signature such as k1=v1&k2=v2
type Any map[interface{}]interface{}

//New new Any
//key type is interface and value type is interface
func New() Any {
	return make(Any)
}

//URL returns k1=v1&k2=v2
func (c *Any) URL() string {
	return c.Join("&")
}

//Join returns e.g k1=v1&k2=v2
func (c *Any) Join(sep string) string {
	if sep == "" {
		sep = "&"
	}
	var s []string
	ks, vs := "", ""
	for k, v := range *c {
		if k != nil && v != nil {
			ks, vs = "", ""
			if tv, ok := k.(Stringer); ok {
				ks = tv.String()
			} else {
				ks = fmt.Sprintf("%v", k)
			}
			if tv, ok := v.(Stringer); ok {
				vs = tv.String()
			} else {
				vs = fmt.Sprintf("%v", v)
			}
			s = append(s, ks+"="+vs)
		}
	}
	if s != nil {
		sort.Sort(sort.StringSlice(s))
	}
	return strings.Join(s, sep)
}

//String KV for string
type String map[string]string

//NewString new KS
//key type is string and value type is string
func NewString() String {
	return make(String)
}

//URL returns k1=v1&k2=v2
func (c *String) URL() string {
	return c.Join("&")
}

//Join returns e.g k1=v1&k2=v2
func (c *String) Join(sep string) string {
	if sep == "" {
		sep = "&"
	}
	var s []string
	for k, v := range *c {
		if k != "" && v != "" {
			s = append(s, k+"="+v)
		}
	}
	if s != nil {
		sort.Sort(sort.StringSlice(s))
	}
	return strings.Join(s, sep)
}
