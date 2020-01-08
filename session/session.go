//Copyright 2018 The axx Authors. All rights reserved.

package session

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/session/memory"
	"github.com/aixiaoxiang/bast/session/redis"
	"github.com/aixiaoxiang/bast/snowflake"

	"github.com/aixiaoxiang/bast/session/conf"
	"github.com/aixiaoxiang/bast/session/engine"
)

var cf *conf.Conf = conf.DefaultConf
var idNode *snowflake.Node

// Start generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func Start(w http.ResponseWriter, r *http.Request) (engine.Store, error) {
	if !cf.Enable {
		return nil, nil
	}
	sessionEngine := cf.Engine
	engine, ok := engine.Engines[sessionEngine]
	if !ok {
		return nil, nil
	}
	sid, errs := getSid(r)
	if errs != nil {
		return nil, errs
	}

	if sid != "" && engine.Exist(sid) {
		return engine.Get(sid)
	}
	if idNode == nil {
		idNode = ids.New()
		if idNode == nil {
			return nil, errors.New("generate session id error")
		}
	}
	// Generate a new session id
	sid = cf.Prefix + strconv.FormatInt(idNode.GenerateWithInt64(), 10) + cf.Suffix
	if errs != nil {
		return nil, errs
	}
	store, err := engine.Get(sid)
	if err != nil {
		return nil, err
	}

	cookie := &http.Cookie{
		Name:     cf.Name,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure(r),
		SameSite: cf.SameSite,
		MaxAge:   cf.LifeTime * 60,
		// Domain:   "",
	}
	if cf.LifeTime > 0 {
		cookie.MaxAge = cf.LifeTime
		cookie.Expires = time.Now().Add(time.Duration(cf.LifeTime) * time.Second)
	}
	if cf.Source == "cookie" {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)
	if cf.Source == "header" {
		r.Header.Set(cf.Name, sid)
		w.Header().Set(cf.Name, sid)
	}
	return store, nil
}

func getSid(r *http.Request) (string, error) {
	cookie, errs := r.Cookie(cf.Name)
	if errs != nil || cookie.Value == "" {
		var id string
		if cf.Source == "url" {
			errs := r.ParseForm()
			if errs != nil {
				return "", errs
			}
			id = r.FormValue(cf.Name)
		}

		// if not found in Cookie / param, then read it from request headers
		if cf.Source == "header" && id == "" {
			sids, isFound := r.Header[cf.Name]
			if isFound && len(sids) != 0 {
				return sids[0], nil
			}
		}
		return id, nil
	}

	return url.QueryUnescape(cookie.Value)
}

// Recycle session data
func recycle() {
	if !cf.Enable {
		return
	}
	needRecycle := false
	for _, v := range engine.Engines {
		if v.NeedRecycle() {
			needRecycle = true
			v.Recycle()
		}
	}
	if needRecycle {
		time.AfterFunc(time.Duration(cf.LifeTime)*time.Second, recycle)
	}
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

//Init session
func Init(c *conf.Conf) error {
	idNode = ids.New()
	cf = c
	err := memory.Init(c)
	if err != nil {
		return err
	}
	err = redis.Init(c)
	if err != nil {
		return err
	}
	if cf.Enable {
		time.AfterFunc(time.Duration(cf.LifeTime)*time.Second, recycle)
	}
	return nil
}
