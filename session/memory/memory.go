//Copyright 2018 The axx Authors. All rights reserved.

package memory

import (
	"container/list"
	"sync"
	"time"

	"github.com/axfor/bast/logs"
	"github.com/axfor/bast/session/conf"
	"github.com/axfor/bast/session/engine"
)

var sengine = &sessionEngine{list: list.New(), data: map[string]*list.Element{}}

//store is memory engine interface
type sessionStore struct {
	lock *sync.RWMutex
	time int64
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
	return nil
}

type sessionEngine struct {
	lock     sync.RWMutex
	lifeTime int64
	list     *list.List
	data     map[string]*list.Element
	cf       *conf.Conf
}

//set session value by key
func (en *sessionEngine) Init(cf *conf.Conf) error {
	en.cf = cf
	en.lifeTime = int64(cf.LifeTime)
	return nil
}

func (en *sessionEngine) Get(id string) (engine.Store, error) {
	en.lock.RLock()
	if element, ok := en.data[id]; ok {
		go en.update(id)
		en.lock.RUnlock()
		return element.Value.(*sessionStore), nil
	}
	en.lock.RUnlock()
	en.lock.Lock()
	defer en.lock.Unlock()
	s := &sessionStore{id: id, time: time.Now().Unix(), data: map[string]interface{}{}}
	if en.cf != nil && en.cf.SessionLock {
		s.lock = &sync.RWMutex{}
	}
	en.data[id] = en.list.PushFront(s)
	return s, nil
}

func (en *sessionEngine) Exist(id string) bool {
	en.lock.RLock()
	defer en.lock.RUnlock()
	if _, ok := en.data[id]; ok {
		return true
	}
	return false
}

func (en *sessionEngine) Delete(id string) error {
	en.lock.Lock()
	defer en.lock.Unlock()
	if element, ok := en.data[id]; ok {
		delete(en.data, id)
		en.list.Remove(element)
		return nil
	}
	return nil
}

func (en *sessionEngine) update(id string) error {
	en.lock.Lock()
	defer en.lock.Unlock()
	if element, ok := en.data[id]; ok {
		element.Value.(*sessionStore).time = time.Now().Unix()
		en.list.MoveToFront(element)
		return nil
	}
	return nil
}

func (en *sessionEngine) Recycle() {
	en.lock.RLock()
	logs.Debug("start session recycle of memory")
	for {
		element := en.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*sessionStore).time + en.lifeTime) < time.Now().Unix() {
			en.lock.RUnlock()
			en.lock.Lock()
			en.list.Remove(element)
			s := element.Value.(*sessionStore)
			id := s.id
			delete(en.data, id)
			s.Clear()
			s = nil
			en.lock.Unlock()
			en.lock.RLock()
		} else {
			break
		}
	}
	logs.Debug("complete session recycle of memory")
	en.lock.RUnlock()
}

//NeedRecycle need recycle session data
func (en *sessionEngine) NeedRecycle() bool {
	return true
}

//Init init
func Init(c *conf.Conf) error {
	if c.Engine == "memory" {
		err := sengine.Init(c)
		engine.Register("memory", sengine)
		return err
	}
	return nil
}

func init() {
}
