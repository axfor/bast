//Copyright 2018 The axx Authors. All rights reserved.

package session

import (
	"testing"
	"time"

	"github.com/aixiaoxiang/bast/session/conf"
	"github.com/aixiaoxiang/bast/session/engine"
)

func Test_Session_Memory(t *testing.T) {
	cf := conf.DefaultConf
	cf.LifeTime = 1
	Init(cf)
	engine, ok := engine.Engines["memory"]
	if !ok {
		t.Fail()
		return
	}
	store, err := engine.Get("1")
	if err != nil {
		t.Fail()
		return
	}
	v2 := "ddddddd"
	err = store.Set("a", v2)
	if err != nil {
		t.Fail()
		return
	}
	v, ok := store.Get("a").(string)
	if !ok || v != v2 {
		t.Fail()
		return
	}
	<-time.After(time.Duration(cf.LifeTime*3) * time.Second)
	v, ok = store.Get("a").(string)
	if ok && v != "" {
		t.Error(v)
		return
	}
}

func Test_Session_Memory_With_Session_Lock(t *testing.T) {
	cf := conf.DefaultConf
	cf.LifeTime = 1
	cf.SessionLock = true
	Init(cf)
	engine, ok := engine.Engines["memory"]
	if !ok {
		t.Fail()
		return
	}
	store, err := engine.Get("1")
	if err != nil {
		t.Fail()
		return
	}
	v2 := "ddddddd"
	err = store.Set("a", v2)
	if err != nil {
		t.Fail()
		return
	}
	v, ok := store.Get("a").(string)
	if !ok || v != v2 {
		t.Fail()
		return
	}
	<-time.After(time.Duration(cf.LifeTime*3) * time.Second)
	v, ok = store.Get("a").(string)
	if ok && v != "" {
		t.Error(v)
		return
	}
}

func Test_Session_Redis(t *testing.T) {
	cf := conf.DefaultConf
	cf.LifeTime = 5
	cf.Engine = "redis"
	cf.Redis = &conf.RedisConf{
		Addrs:    "127.0.0.1:6379",
		PoolSize: 100,
	}
	err := Init(cf)
	if err != nil {
		t.Error(err)
		return
	}
	engine, ok := engine.Engines["redis"]
	if !ok {
		t.Fail()
		return
	}
	store, err := engine.Get("1")
	if err != nil {
		t.Error(err)
		return
	}
	v2 := "ddddddd"
	err = store.Set("a", v2)
	if err != nil {
		t.Fail()
		return
	}
	v, ok := store.Get("a").(string)
	if !ok || v != v2 {
		t.Fail()
		return
	}

	err = store.Commit()
	if err != nil {
		t.Error(err)
		return
	}

	<-time.After(time.Duration(3) * time.Second)
	store, err = engine.Get("1")
	if err != nil {
		t.Error(err)
		return
	}
	v, ok = store.Get("a").(string)
	if !ok || v == "" {
		t.Error(v)
		return
	}

	<-time.After(time.Duration(cf.LifeTime*3) * time.Second)
	store, err = engine.Get("1")
	if err != nil {
		t.Error(err)
		return
	}
	v, ok = store.Get("a").(string)
	if ok && v != "" {
		t.Error(v)
		return
	}
}

func Test_Session_Redis_With_Session_Lock(t *testing.T) {
	cf := conf.DefaultConf
	cf.LifeTime = 5
	cf.Engine = "redis"
	cf.SessionLock = true
	cf.Redis = &conf.RedisConf{
		Addrs:    "127.0.0.1:6379",
		PoolSize: 100,
	}
	err := Init(cf)
	if err != nil {
		t.Error(err)
		return
	}
	engine, ok := engine.Engines["redis"]
	if !ok {
		t.Fail()
		return
	}
	store, err := engine.Get("1")
	if err != nil {
		t.Error(err)
		return
	}
	v2 := "ddddddd"
	err = store.Set("a", v2)
	if err != nil {
		t.Fail()
		return
	}
	v, ok := store.Get("a").(string)
	if !ok || v != v2 {
		t.Fail()
		return
	}

	err = store.Commit()
	if err != nil {
		t.Error(err)
		return
	}

	<-time.After(time.Duration(3) * time.Second)
	store, err = engine.Get("1")
	if err != nil {
		t.Error(err)
		return
	}
	v, ok = store.Get("a").(string)
	if !ok || v == "" {
		t.Error(v)
		return
	}

	<-time.After(time.Duration(cf.LifeTime*3) * time.Second)
	store, err = engine.Get("1")
	if err != nil {
		t.Error(err)
		return
	}
	v, ok = store.Get("a").(string)
	if ok && v != "" {
		t.Error(v)
		return
	}
}
