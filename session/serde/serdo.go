//Copyright 2018 The axx Authors. All rights reserved.

package serde

import (
	"bytes"
	"encoding/gob"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(map[string]string{})
	gob.Register(map[string]int{})
	gob.Register(map[string]int64{})
	gob.Register([]interface{}{})
}

//Encode encode obj(map[string]interface{}) to []byte
func Encode(obj map[string]interface{}) ([]byte, error) {
	for _, v := range obj {
		gob.Register(v)
	}
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return []byte(""), err
	}
	return buf.Bytes(), nil
}

// Decode decode []byte to map obj(map[string]interface{})
func Decode(data []byte) (map[string]interface{}, error) {
	var ret map[string]interface{}
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err := dec.Decode(&ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
