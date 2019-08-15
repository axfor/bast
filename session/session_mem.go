//Copyright 2018 The axx Authors. All rights reserved.
//Reference github.com/astaxie/beego/session

package session

import (
	"container/list"
	"sync"
	"time"
)

var memEngine = &memoryEngine{list: list.New(), data: map[string]*list.Element{}}

//memory is memory engine interface
type memory struct {
	lock         sync.RWMutex
	timeAccessed int64
	id           string                 //session id
	data         map[string]interface{} //data
}

//Set session set value by key
func (m *memory) Set(key string, value interface{}) error {
	m.lock.Lock()
	defer m.lock.Lock()
	m.data[key] = value
	return nil
}

//set session value by key
func (m *memory) Get(key string) interface{} {
	m.lock.RLock()
	defer m.lock.RLock()
	if v, ok := m.data[key]; ok {
		return v
	}
	return nil
}

//delete session value by key
func (m *memory) Delete(key string) error {
	m.lock.Lock()
	defer m.lock.Lock()
	delete(m.data, key)
	return nil
}

//return current session ID
func (m *memory) ID() string {
	return m.id
}

//clear all data
func (m *memory) Clear() error {
	m.lock.Lock()
	defer m.lock.Lock()
	m.data = nil
	m.data = map[string]interface{}{}
	return nil
}

type memoryEngine struct {
	lock     sync.RWMutex
	lifeTime int64
	list     *list.List
	data     map[string]*list.Element
}

//set session value by key
func (m *memoryEngine) Init(lifeTime int) error {
	m.lifeTime = int64(lifeTime)
	return nil
}

func (m *memoryEngine) Get(id string) (Store, error) {
	m.lock.RLock()
	if element, ok := m.data[id]; ok {
		go m.update(id)
		m.lock.RUnlock()
		return element.Value.(*memory), nil
	}
	m.lock.RUnlock()
	m.lock.Lock()
	defer m.lock.Unlock()
	sess := &memory{id: id, timeAccessed: time.Now().Unix(), data: map[string]interface{}{}}
	m.data[id] = m.list.PushFront(sess)
	return sess, nil
}

func (m *memoryEngine) Exist(id string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if _, ok := m.data[id]; ok {
		return true
	}
	return false
}

func (m *memoryEngine) Delete(id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if element, ok := m.data[id]; ok {
		delete(m.data, id)
		m.list.Remove(element)
		return nil
	}
	return nil
}

func (m *memoryEngine) update(id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if element, ok := m.data[id]; ok {
		element.Value.(*memory).timeAccessed = time.Now().Unix()
		m.list.MoveToFront(element)
		return nil
	}
	return nil
}

func (m *memoryEngine) GC() {
	m.lock.RLock()
	for {
		element := m.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*memory).timeAccessed + m.lifeTime) < time.Now().Unix() {
			m.lock.RUnlock()
			m.lock.Lock()
			m.list.Remove(element)
			delete(m.data, element.Value.(*memory).id)
			m.lock.Unlock()
			m.lock.RLock()
		} else {
			break
		}
	}
	m.lock.RUnlock()
}

func init() {
	Register("memory", memEngine)
}
