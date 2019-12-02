//Copyright 2018 The axx Authors. All rights reserved.

package session

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aixiaoxiang/bast/conf"
	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/logs"
)

//Store is store interface
type Store interface {
	Set(key string, value interface{}) error //set session value by key
	Get(key string) interface{}              //get session value by key
	Delete(key string) error                 //delete session value by key
	ID() string                              //return current session id
	Clear() error                            //clear all data
}

//Engine is store engine interface
type Engine interface {
	Init(lifeTime int) error      //set session value by key
	Get(id string) (Store, error) //get session value by key
	Exist(id string) bool         //get session value by key
	Delete(d string) error        //delete session value by key
	GC()                          //clean expired sessions
}

var engines = map[string]Engine{}

//Register a session provide by the engine name
func Register(name string, engine Engine) {
	if _, ok := engines[name]; !ok {
		engines[name] = engine
	}
}

// Start generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func Start(w http.ResponseWriter, r *http.Request) (Store, error) {
	if !conf.SessionEnable() {
		return nil, nil
	}

	sessionEngine := conf.SessionEngine()
	engine, ok := engines[sessionEngine]
	if !ok {
		return nil, nil
	}
	id, errs := getSid(r)
	if errs != nil {
		return nil, errs
	}

	if id != "" && engine.Exist(id) {
		return engine.Get(id)
	}

	// Generate a new session
	id = strconv.FormatInt(ids.ID(), 10)
	if errs != nil {
		return nil, errs
	}
	store, err := engine.Get(id)
	if err != nil {
		return nil, err
	}
	sessionName := conf.SessionName()
	lifeTime := conf.SessionLifeTime()
	sessionSource := conf.SessionSource()

	cookie := &http.Cookie{
		Name:     sessionName,
		Value:    url.QueryEscape(id),
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure(r),
		// Domain:   "",
	}
	if lifeTime > 0 {
		cookie.MaxAge = lifeTime
		cookie.Expires = time.Now().Add(time.Duration(lifeTime) * time.Second)
	}
	if sessionSource == "cookie" {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)
	if sessionSource == "header" {
		r.Header.Set(sessionName, id)
		w.Header().Set(sessionName, id)
	}
	return store, nil
}

func getSid(r *http.Request) (string, error) {
	sessionName := conf.SessionName()
	sessionSource := conf.SessionSource()
	cookie, errs := r.Cookie(sessionName)
	if errs != nil || cookie.Value == "" {
		var id string
		if sessionSource == "url" {
			errs := r.ParseForm()
			if errs != nil {
				return "", errs
			}
			id = r.FormValue(sessionName)
		}

		// if not found in Cookie / param, then read it from request headers
		if sessionSource == "header" && id == "" {
			sids, isFound := r.Header[sessionName]
			if isFound && len(sids) != 0 {
				return sids[0], nil
			}
		}
		return id, nil
	}

	// HTTP Request contains cookie for sessionid info.
	return url.QueryUnescape(cookie.Value)
}

// GC Start session gc process.
// it can do gc in times after gc lifetime.
func GC() {
	logs.Debug("start session gc")
	lt := conf.SessionLifeTime()
	for _, v := range engines {
		v.GC()
	}
	time.AfterFunc(time.Duration(lt)*time.Second, func() {
		if !conf.SessionEnable() {
			return
		}
		GC()
	})
	logs.Debug("end session gc")
}

func isSecure(req *http.Request) bool {
	if req.URL.Scheme != "" {
		return req.URL.Scheme == "https"
	}
	if req.TLS == nil {
		return false
	}
	return true
}

func init() {
	GC()
}
