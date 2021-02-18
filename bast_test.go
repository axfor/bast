//Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"testing"
	"time"

	"github.com/axfor/bast/httpc"
)

var appStarted bool

func Benchmark_QPS(t *testing.B) {
	if !appStarted {
		startApp()
	}

	t.ResetTimer()
	t.ReportAllocs()

	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, err := httpc.Get("http://127.0.0.1:9999/bast").Param("n", t.N).String()
			if err != nil || r == "" {
				t.Error(err)
			}
		}
	})

}

func TestBastApp(t *testing.T) {
	t.Cleanup(exitApp)
	startApp()

	ok := httpc.Get("http://127.0.0.1:9999/bast").OK()
	if !ok {
		t.FailNow()
	}

	ok = httpc.Post("http://127.0.0.1:9999/bast").OK()
	if !ok {
		t.FailNow()
	}

	ok = httpc.Put("http://127.0.0.1:9999/bast").OK()
	if !ok {
		t.FailNow()
	}
}

func startApp() {
	appStarted = true
	go Run(":9999")

	time.Sleep(time.Second)

}

func init() {
	Get("/bast", func(ctx *Context) {
		ctx.Says("hello bast of get")
	})

	Post("/bast", func(ctx *Context) {
		ctx.Says("hello bast of post")
	})

	Put("/bast", func(ctx *Context) {
		ctx.Says("hello bast of put")
	})

}

func exitApp() {
	appStarted = false
	Shutdown(nil)
}
