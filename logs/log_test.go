//Copyright 2018 The axx Authors. All rights reserved.

package logs

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_Info(t *testing.T) {
	f := "./logs/logs.log"
	Info("info")
	Debug("debug")
	Error("error")
	defer os.RemoveAll("./logs")
	c, err := ioutil.ReadFile(f)
	if len(c) <= 0 || err != nil {
		t.Fatal(err)
	}
}

func init() {
	Init(nil)
}
