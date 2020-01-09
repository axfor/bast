//Copyright 2018 The axx Authors. All rights reserved.

package standalone

import (
	"errors"
	"sync"
	"time"

	"github.com/aixiaoxiang/bast/logs"
	"github.com/aixiaoxiang/bast/session/conf"
	"github.com/aixiaoxiang/bast/session/engine"
	"github.com/aixiaoxiang/bast/session/serde"
	"github.com/go-redis/redis"
)

//ErrorConn connection  error
var ErrorConn = errors.New("connection to redis error")

//ErrorNotFondRedisConf not fond redis conf
var ErrorNotFondRedisConf = errors.New("not fond redis conf")

var sengine = &sessionEngine{}

//store is redis engine interface
type sessionStore struct {
	lock *sync.RWMutex
	en   *sessionEngine
	id   string                 //session id
	data map[string]interface{} //data
}

//Set session set value by key
func (s *sessionStore) Set(key string, value interface{}) error {
	if s.lock != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
	}
	s.data[key] = value
	return nil
}

//set session value by key
func (s *sessionStore) Get(key string) interface{} {
	if s.lock != nil {
		s.lock.RLock()
		defer s.lock.RUnlock()
	}
	if v, ok := s.data[key]; ok {
		return v
	}
	return nil
}

//delete session value by key
func (s *sessionStore) Delete(key string) error {
	if s.lock != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
	}
	delete(s.data, key)
	return nil
}

//return current session ID
func (s *sessionStore) ID() string {
	return s.id
}

//clear all data
func (s *sessionStore) Clear() error {
	if s.lock != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
	}
	s.data = nil
	s.data = map[string]interface{}{}
	return nil
}

//commit data to session store
func (s *sessionStore) Commit() error {
	if s.en.c == nil {
		logs.Errors("connection error", ErrorConn)
		return ErrorConn
	}
	be, err := serde.Encode(s.data)
	if err != nil {
		return err
	}
	err = s.en.c.Set(s.id, string(be), time.Duration(s.en.cf.LifeTime)*time.Second).Err()
	return err
}

//NeedRecycle need recycle session data
func (s *sessionStore) NeedRecycle() bool {
	return false
}

type sessionEngine struct {
	c  *redis.Client
	cf *conf.Conf
}

//set session value by key
func (en *sessionEngine) Init(cf *conf.Conf) error {
	if cf.Redis == nil {
		logs.Errors("init conf", ErrorNotFondRedisConf)
		return ErrorNotFondRedisConf
	}
	en.cf = cf
	en.c = redis.NewClient(&redis.Options{
		Addr:     cf.Redis.Addrs,
		Password: cf.Redis.Password,
		PoolSize: cf.Redis.PoolSize,
	})
	err := en.c.Ping().Err()
	if err != nil {
		logs.Errors("connection error", err)
	}
	return err
}

func (en *sessionEngine) Get(id string) (engine.Store, error) {
	if en.c == nil {
		logs.Errors("connection error", ErrorConn)
		return nil, ErrorConn
	}
	var data map[string]interface{}
	values, err := en.c.Get(id).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if len(values) == 0 {
		data = make(map[string]interface{})
	} else {
		if data, err = serde.Decode([]byte(values)); err != nil {
			return nil, err
		}
	}
	s := &sessionStore{en: en, id: id, data: data}
	if en.cf != nil && en.cf.SessionLock {
		s.lock = &sync.RWMutex{}
	}
	return s, nil
}

func (en *sessionEngine) Exist(id string) bool {
	if en.c == nil {
		return false
	}
	if existed, err := en.c.Exists(id).Result(); err != nil || existed == 0 {
		return false
	}
	return true
}

func (en *sessionEngine) Delete(id string) error {
	if en.c == nil {
		logs.Errors("connection error", ErrorConn)
		return ErrorConn
	}
	_, err := en.c.Del(id).Result()
	return err
}

func (en *sessionEngine) Recycle() {
	//
}

//NeedRecycle need recycle session data
func (en *sessionEngine) NeedRecycle() bool {
	return false
}

//Init init
func Init(c *conf.Conf) error {
	if c.Engine == "redis" {
		err := sengine.Init(c)
		engine.Register("redis", sengine)
		return err
	}
	return nil
}
