//Copyright 2018 The axx Authors. All rights reserved.

package conf

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/logs"
	sessionConf "github.com/aixiaoxiang/bast/session/conf"
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
	frist      *AppConf
	rawConfs   []AppConf
	confHandle bool
	Confs      map[string]*AppConf
}

//AppConf  app config item
type AppConf struct {
	Key          string            `json:"key"`
	Name         string            `json:"name"`
	Addr         string            `json:"addr"`
	FileDir      string            `json:"fileDir"`
	Debug        bool              `json:"debug"`
	BaseURL      string            `json:"baseUrl"`
	IDNode       uint8             `json:"idNode"`   //id node
	Lang         string            `json:"lang"`     //lang
	SameSiteText string            `json:"sameSite"` //strict|lax|none
	Wrap         *bool             `json:"wrap"`     //wrap response body
	Session      *sessionConf.Conf `json:"session"`  //session
	Log          *logs.Conf        `json:"log"`      //log conf
	CORS         *CORS             `json:"cors"`     //CORS
	Conf         interface{}       `json:"conf"`     //user conf
	Extend       string            `json:"extend"`   //user extend
	ConfHandle   bool              `json:"-"`
	SameSite     http.SameSite
}

//Item default db config
type Item struct {
	Name      string `json:"name"`
	Title     string `json:"dbTitle"`
	User      string `json:"dbUser"`
	Pwd       string `json:"dbPwd"`
	Server    string `json:"dbServer"`
	Charset   string `json:"dbCharset"`
	ParseTime *bool  `json:"dbParseTime"`
	Loc       string `json:"dbLoc"`
}

//CORS  config
type CORS struct {
	AllowOrigin      string `json:"allowOrigin"`
	AllowMethods     string `json:"allowMethods"`
	AllowHeaders     string `json:"allowHeaders"`
	AllowCredentials string `json:"allowCredentials"`
	MaxAge           string `json:"maxAge"`
}

//Manager is manager all config objects
func Manager() *AppConfMgr {
	if confObj == nil {
		data, err := ioutil.ReadFile(Path())
		if err != nil {
			return nil
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
			logs.Errors("conf mgr init error", err)
			fmt.Println("conf mgr init error:" + err.Error())
			return nil
		}
		Init(appConf)
	}
	return confObj
}

//Init  config
func Init(appConf []AppConf) {
	lg := len(appConf)
	if lg == 0 {
		return
	}
	if confObj == nil {
		confObj = &AppConfMgr{}
		confObj.rawConfs = appConf
		confObj.Confs = make(map[string]*AppConf)
		for i := 0; i < lg; i++ {
			c := &appConf[i]
			c.SameSite = sameSite(c.SameSiteText)
			if c.Session == nil {
				c.Session = sessionConf.NewDefault()
			} else {
				if c.Session.LifeTime <= 0 {
					c.Session.LifeTime = 60 * 20
				}
				if c.Session.Name == "" {
					c.Session.Name = "_sid"
				}
				if c.Session.Engine == "" {
					c.Session.Engine = "memory"
				}
			}
			c.Session.SameSite = c.SameSite
			if c.FileDir != "" {
				if c.FileDir[len(c.FileDir)-1] != '/' {
					c.FileDir += "/"
				}
			}
			if confHandle != nil {
				err := confHandle(c)
				if err != nil {
					continue
				}
				c.ConfHandle = true
			}
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
		if confHandle != nil {
			confObj.confHandle = true
		}
	}
	if confFinishHandle != nil && confObj != nil {
		confFinishHandle(confObj.frist)
	}
}

//OK check conf
func OK() bool {
	return Manager() != nil
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

//WithPath returns the current app config
func WithPath(filePath string) *AppConf {
	flagConf = filePath
	return Conf()
}

//FileDir if app config configuration fileDir return it，orherwise return app exec path
func FileDir() string {
	c := Conf()
	if c != nil && c.FileDir != "" {
		return c.FileDir
	}
	return filepath.Dir(os.Args[0])
}

//SessionConf return session conf
func SessionConf() *sessionConf.Conf {
	c := Conf()
	if c == nil {
		return sessionConf.DefaultConf
	}
	return sessionConf.DefaultConf
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

//Config returns the key app config
func Config(key string) *AppConf {
	appConf := Manager()
	if appConf != nil && appConf.Confs != nil {
		c, ok := appConf.Confs[key]
		if c != nil && ok {
			return c
		}
	}
	return nil
}

//Confs returns the all app config
func Confs() []AppConf {
	appConf := Manager()
	if appConf != nil && appConf.Confs != nil {
		return appConf.rawConfs
	}
	return nil
}

//ConfsWithPath returns the all app config
func ConfsWithPath(path string) []AppConf {
	flagConf = path
	return Confs()
}

//UserConf  returns the current user config
func UserConf() interface{} {
	appConf := Conf()
	if appConf != nil && appConf.Conf != nil {
		return appConf.Conf
	}
	return nil
}

//LogConf  returns the current log config
func LogConf() *logs.Conf {
	appConf := Conf()
	if appConf != nil && appConf.Log != nil {
		return appConf.Log
	}
	return nil
}

//CORSConf  returns the current log config
func CORSConf() *CORS {
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
	return &CORS{
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

//Parse parse config path
func Parse(f *flag.FlagSet) string {
	exPath := filepath.Dir(os.Args[0])
	fs := f.Lookup("conf")
	if fs != nil {
		flagConf = fs.Value.String()
	}
	fs = f.Lookup("appkey")
	if fs != nil {
		flagAppKey = fs.Value.String()
	}
	if flagConf == "" {
		cf := exPath + "/config.conf"
		flagConf = cf
	}
	return flagConf
}

//Register regist handler of conf
func Register(handle ConfingHandle, finish ...FinishHandle) {
	confHandle = handle
	if finish != nil {
		confFinishHandle = finish[0]
	}
	if confObj == nil || confObj.confHandle {
		return
	}
	cf := confObj.rawConfs
	if cf == nil {
		return
	}
	lg := len(cf)
	if lg > 0 {
		confObj = nil
		Init(cf)
	}
}
