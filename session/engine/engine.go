//Copyright 2018 The axx Authors. All rights reserved.

package engine

import (
	"github.com/aixiaoxiang/bast/session/conf"
)

//Store is store interface
type Store interface {
	Set(key string, value interface{}) error //set session value by key
	Get(key string) interface{}              //get session value by key
	Delete(key string) error                 //delete session value by key
	ID() string                              //return current session id
	Clear() error                            //clear all data
	Commit() error                           //commit data to session store
}

//Engine is store engine interface
type Engine interface {
	Init(cf *conf.Conf) error     //set session value by key
	Get(id string) (Store, error) //get session value by key
	Exist(id string) bool         //get session value by key
	Delete(d string) error        //delete session value by key
	Recycle()                     //clean all expired sessions
	NeedRecycle() bool            //need recycle session store
}

//Engines all registered engine
var Engines = map[string]Engine{}

//Register a session provide by the engine name
func Register(name string, engine Engine) {
	if _, ok := Engines[name]; !ok {
		Engines[name] = engine
	}
}
