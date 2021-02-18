//Copyright 2018 The axx Authors. All rights reserved.

package conf

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/axfor/bast/ids"
	"github.com/axfor/bast/logs"
	sessionConf "github.com/axfor/bast/session/conf"
)

var (
	flagConf, flagAppKey string
	confObj              *AppConfMgr
	confHandle           ConfingHandle
	confFinishHandle     FinishHandle
)

//ConfingHandle handler  conf
type ConfingHandle func(appConf *AppConf) error

//FinishHandle finished handler conf
type FinishHandle func(appConf *AppConf) error

//AppConfMgr app config
type AppConfMgr struct {
	frist    *AppConf
	rawConfs []AppConf
	Confs    map[string]*AppConf
}

//AppConf  app config item
type AppConf struct {
	Key          string            `json:"key"`
	Name         string            `json:"name"`
	Addr         string            `json:"addr"`
	CertFile     string            `json:"certFile"` //tls cert file
	KeyFile      string            `json:"keyFile"`  //tls cert key file
	FileDir      string            `json:"fileDir"`
	Debug        bool              `json:"debug"`
	BaseURL      string            `json:"baseUrl"`
	IDNode       uint8             `json:"idNode"`    //id node
	Lang         string            `json:"lang"`      //lang
	Trans        string            `json:"trans"`     //trans
	SameSiteText string            `json:"sameSite"`  //strict|lax|none
	Wrap         *bool             `json:"wrap"`      //wrap response body
	Session      *sessionConf.Conf `json:"session"`   //session
	Log          *logs.Conf        `json:"log"`       //log conf
	CORS         *CORSConf         `json:"cors"`      //CORS
	Conf         interface{}       `json:"conf"`      //user conf
	Extend       string            `json:"extend"`    //user extend
	Page         *PaginationConf   `json:"page"`      //pagination conf
	Registry     *RegistryConf     `json:"registry"`  //service registry center
	Discovery    *DiscoveryConf    `json:"discovery"` //service discovery center
	Shutdown     int64             `json:"shutdown"`  //service shutdown timeout(default 60 second)
	SameSite     http.SameSite     `json:"-"`
	initTag      bool
}

//CORSConf  config
type CORSConf struct {
	AllowOrigin      string `json:"allowOrigin"`
	AllowMethods     string `json:"allowMethods"`
	AllowHeaders     string `json:"allowHeaders"`
	AllowCredentials string `json:"allowCredentials"`
	MaxAge           string `json:"maxAge"`
}

//PaginationConf  config
type PaginationConf struct {
	Page    string `json:"page"`
	Total   string `json:"total"`
	PageRow string `json:"pageRow"`
}

//RegistryConf  config
type RegistryConf struct {
	Enable      bool   `json:"enable"`    //
	BaseURL     string `json:"baseUrl"`   //
	Prefix      string `json:"prefix"`    //
	Endpoints   string `json:"endpoints"` //localhost:2379,localhost:22379
	DialTimeout int64  `json:"timeout"`   //second default 5s
	TTL         int64  `json:"ttl"`
}

//DiscoveryConf  config
type DiscoveryConf struct {
	Enable      bool   `json:"enable"`    //
	Prefix      string `json:"prefix"`    //
	Endpoints   string `json:"endpoints"` //localhost:2379,localhost:22379
	DialTimeout int64  `json:"timeout"`   //second default 5s
}

//Init data
func Init() {
	if confObj == nil {
		data, err := ioutil.ReadFile(Path())
		if err != nil {
			logs.Errors("read conf error", err)
			return
		}
		s := strings.TrimSpace(string(data))
		isAdd := (s[0] != '[')
		if isAdd && s[0] != '{' {
			s = "{" + s + "}"
		}
		if isAdd {
			s = "[" + s + "]"
			data = []byte(s)
		}
		appConf := []AppConf{}
		err = json.Unmarshal(data, &appConf)
		if err != nil {
			logs.Errors("conf manager init error", err)
			return
		}
		if confObj != nil {
			return
		}
		confObj = &AppConfMgr{}
		confObj.rawConfs = appConf
		confObj.Confs = make(map[string]*AppConf)
		lg := len(appConf)
		if lg > 0 {
			for i := 0; i < lg; i++ {
				c := &appConf[i]
				appConfWithInit(c)
				if c.Key == flagAppKey && confObj.frist == nil {
					confObj.frist = c
				}
				confObj.Confs[c.Key] = c
			}
			if confObj.frist == nil {
				confObj.frist = &appConf[0]
			}
			//set default current id node
			if confObj.frist != nil {
				ids.SetIDNode(confObj.frist.IDNode)
			}
			callbackHandle()
		}
	}
}

//Manager is manager all config objects
func Manager() *AppConfMgr {
	if confObj == nil {
		Init()
	}
	return confObj
}

func appConfWithInit(c *AppConf) {
	if c.initTag {
		return
	}
	c.initTag = true
	if c.Lang == "" {
		c.Lang = "en"
	}
	c.SameSite = sameSite(c.SameSiteText)
	if c.Session == nil {
		c.Session = sessionConf.NewDefault()
	} else {
		if c.Session.LifeTime <= 0 {
			c.Session.LifeTime = 20
		}
		if c.Session.Name == "" {
			c.Session.Name = "_sid"
		}
		if c.Session.Engine == "" {
			c.Session.Engine = "memory"
		}
		c.Session.LifeTime *= 60
	}
	c.Session.SameSite = c.SameSite
	if c.FileDir != "" {
		if c.FileDir[len(c.FileDir)-1] != '/' {
			c.FileDir += "/"
		}
	}

	if c.Shutdown > 0 {
		c.Shutdown *= int64(time.Microsecond)
	}
}

//callback all handle
func callbackHandle() {
	if confObj == nil {
		return
	}
	lg := len(confObj.rawConfs)
	if lg == 0 {
		return
	}
	if confHandle != nil {
		for i := 0; i < lg; i++ {
			c := &confObj.rawConfs[i]
			err := confHandle(c)
			if err != nil {
				logs.Errors("callback conf handle error", err)
			}
		}
	}
	if confFinishHandle != nil {
		confFinishHandle(confObj.frist)
	}
}

//OK check conf
func OK() bool {
	return Manager() != nil
}

//Lang returns lang
func Lang() string {
	c := Conf()
	if c != nil {
		return c.Lang
	}
	return ""
}

//Trans returns Trans conf
func Trans() string {
	c := Conf()
	if c != nil {
		return c.Trans
	}
	return ""
}

//Confs returns the all app configs
func Confs() []AppConf {
	appConf := Manager()
	if appConf != nil && appConf.Confs != nil {
		return appConf.rawConfs
	}
	return nil
}

//Conf returns the current app config
func Conf() *AppConf {
	appConf := Manager()
	if appConf != nil && appConf.Confs != nil {
		if flagAppKey == "" {
			return appConf.frist
		}
		c, ok := appConf.Confs[flagAppKey]
		if c != nil && ok {
			return c
		}
	}
	return nil
}

//WithKey returns the key app config
func WithKey(key string) *AppConf {
	appConf := Manager()
	if appConf != nil && appConf.Confs != nil {
		c, ok := appConf.Confs[key]
		if c != nil && ok {
			return c
		}
	}
	return nil
}

//WithPath returns the current app config
func WithPath(path string) *AppConf {
	flagConf = path
	return Conf()
}

//WithPaths returns the all app config
func WithPaths(path string) []AppConf {
	flagConf = path
	return Confs()
}

//FileDir if app config configuration fileDir return it，orherwise return app exec path
func FileDir() string {
	c := Conf()
	if c != nil && c.FileDir != "" {
		return c.FileDir
	}
	return filepath.Dir(os.Args[0])
}

//Session return session conf
func Session() *sessionConf.Conf {
	c := Conf()
	if c != nil {
		return c.Session
	}
	return sessionConf.DefaultConf
}

//Page return Pagination conf
func Page() *PaginationConf {
	var p *PaginationConf
	c := Conf()
	if c != nil && c.Page != nil {
		p = c.Page
	}
	if p == nil {
		p = &PaginationConf{
			Page:    "page",
			Total:   "total",
			PageRow: "pageRow",
		}
	}
	if p.Page == "" {
		p.Page = "page"
	}
	if p.Total == "" {
		p.Total = "total"
	}
	if p.PageRow == "" {
		p.PageRow = "pageRow"
	}
	return p
}

//Registry return service registry conf
func Registry() *RegistryConf {
	c := Conf()
	if c != nil {
		return c.Registry
	}
	return nil
}

//Discovery return service discovery conf
func Discovery() *DiscoveryConf {
	c := Conf()
	if c != nil {
		return c.Discovery
	}
	return nil
}

//SameSite if app config configuration cookie sameSite return it，orherwise return 'None'
func SameSite() http.SameSite {
	c := Conf()
	return c.SameSite
}

func sameSite(same string) http.SameSite {
	if same != "" {
		switch same {
		case "lax":
			return http.SameSiteLaxMode //Lax
		case "strict":
			return http.SameSiteStrictMode //Strict
		default:
			return http.SameSiteNoneMode //None
		}
	}
	return http.SameSiteDefaultMode
}

//User  returns the current user config
func User() interface{} {
	appConf := Conf()
	if appConf != nil && appConf.Conf != nil {
		return appConf.Conf
	}
	return nil
}

//Log  returns the current log config
func Log() *logs.Conf {
	appConf := Conf()
	if appConf != nil && appConf.Log != nil {
		return appConf.Log
	}
	return nil
}

//CORS  returns the current log config
func CORS() *CORSConf {
	appConf := Conf()
	if appConf != nil && appConf.CORS != nil {
		if appConf.CORS.AllowMethods == "" {
			appConf.CORS.AllowMethods = "GET, POST, OPTIONS, PATCH, PUT, DELETE, HEAD,UPDATE"
		}
		if appConf.CORS.AllowHeaders == "" {
			appConf.CORS.AllowHeaders = "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma, BaseUrl, baseurl"
		}
		if appConf.CORS.AllowCredentials == "" {
			appConf.CORS.AllowCredentials = "true"
		}
		if appConf.CORS.MaxAge == "" {
			appConf.CORS.MaxAge = "1728000"
		}
		return appConf.CORS
	}
	return &CORSConf{
		AllowOrigin:      "",
		AllowMethods:     "GET, POST, OPTIONS, PATCH, PUT, DELETE, HEAD,UPDATE",
		AllowHeaders:     "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma, BaseUrl, baseurl",
		AllowCredentials: "true",
		MaxAge:           "1728000",
	}
}

//Wrap  wrap response body
func Wrap() bool {
	appConf := Conf()
	if appConf != nil && appConf.Wrap != nil {
		return *appConf.Wrap
	}
	return true
}

//Path  returns the current config path
func Path() string {
	return flagConf
}

//SetPath  set the current config path
func SetPath(confPath string) {
	flagConf = confPath
}

//Command command
func Command(conf, appkey *string) {
	exPath := filepath.Dir(os.Args[0])
	if *conf == "" {
		*conf = exPath + "/config.conf"
	}
	flagConf = *conf
	flagAppKey = *appkey
}

//Register regist handler of conf
func Register(handle ConfingHandle, finish ...FinishHandle) {
	confHandle = handle
	if finish != nil {
		confFinishHandle = finish[0]
	}
	callbackHandle()
}

//Shutdown returns service shutdown timeout
func Shutdown() time.Duration {
	c := Conf()
	if c != nil && c.Shutdown > 0 {
		return time.Duration(c.Shutdown)
	}
	return time.Duration(60 * int64(time.Second))
}
