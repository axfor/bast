//Copyright 2018 The axx Authors. All rights reserved.

package conf

import "net/http"

//DefaultConf default conf
var DefaultConf = NewDefault()

//Conf  config
type Conf struct {
	Enable      bool          `json:"enable"`      //false
	LifeTime    int           `json:"lifeTime"`    //20 (min)
	Name        string        `json:"name"`        //_sid
	Prefix      string        `json:"prefix"`      //session id prefix
	Suffix      string        `json:"suffix"`      //session id suffix
	Engine      string        `json:"engine"`      //memory
	Source      string        `json:"source"`      //header|cookie
	SameSite    http.SameSite `json:"sameSite"`    //strict|lax|none
	SessionLock bool          `json:"sessionLock"` //every session a lock(default is false)
	Redis       *RedisConf    `json:"redis"`       //redis
}

//RedisConf  config
type RedisConf struct {
	Addrs    string `json:"addrs"`    //
	Password string `json:"password"` //
	PoolSize int    `json:"poolSize"` //
}

//NewDefault create default *session.Conf
func NewDefault() *Conf {
	c := &Conf{
		Enable:   true,
		LifeTime: 60 * 20,
		Name:     "_sid",
		Engine:   "memory",
		Source:   "cookie",
		SameSite: http.SameSiteDefaultMode,
	}
	return c
}
