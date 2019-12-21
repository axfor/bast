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

	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/logs"
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
	Key             string        `json:"key"`
	Name            string        `json:"name"`
	Addr            string        `json:"addr"`
	FileDir         string        `json:"fileDir"`
	Debug           bool          `json:"debug"`
	BaseURL         string        `json:"baseUrl"`
	IDNode          uint8         `json:"idNode"`
	Log             *logs.LogConf `json:"log"`
	Conf            interface{}   `json:"conf"`
	Extend          string        `json:"extend"`
	SessionEnable   bool          `json:"sessionEnable"`   //false
	SessionLifeTime int           `json:"sessionLifeTime"` //20 (min)
	SessionName     string        `json:"sessionName"`     //_sid
	SessionEngine   string        `json:"sessionEngine"`   //memory
	SessionSource   string        `json:"sessionSource"`   //url|header|cookie
	SameSite        string        `json:"sameSite"`        //strict|lax|none
	Lang            string        `json:"lang"`
	ConfHandle      bool          `json:"-"`
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

//Manager is manager all config objects
func Manager() *AppConfMgr {
	if confObj == nil {
		data, err := ioutil.ReadFile(Path())
		if err != nil {
			return nil
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
			ids.SetCurrentIDNode(confObj.frist.IDNode)
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

//FileDir if app config configuration fileDir return it，orherwise return app exec path
func FileDir() string {
	c := Conf()
	if c != nil && c.FileDir != "" {
		return c.FileDir
	}
	return filepath.Dir(os.Args[0])
}

//SessionLifeTime if app config configuration sessionLifeTime return it，orherwise return 20min
func SessionLifeTime() int {
	c := Conf()
	if c != nil && c.SessionLifeTime > 0 {
		return c.SessionLifeTime
	}
	return 60 * 20
}

//SessionName if app config configuration sessionName return it，orherwise return '_sid'
func SessionName() string {
	c := Conf()
	if c != nil && c.SessionName != "" {
		return c.SessionName
	}
	return "_sid"
}

//SessionEngine if app config configuration sessionEngine return it，orherwise return 'memory'
func SessionEngine() string {
	c := Conf()
	if c != nil && c.SessionEngine != "" {
		return c.SessionEngine
	}
	return "memory"
}

//SessionEnable if app config configuration sessionEnable return it，orherwise return false
func SessionEnable() bool {
	c := Conf()
	if c != nil {
		return c.SessionEnable
	}
	return false
}

//SessionSource if app config configuration sessionSource return it，orherwise return 'cookie'
func SessionSource() string {
	c := Conf()
	if c != nil && c.SessionSource != "" {
		return c.SessionSource
	}
	return "cookie"
}

//SameSite if app config configuration cookie sameSite return it，orherwise return 'None'
func SameSite() http.SameSite {
	c := Conf()
	if c != nil && c.SameSite != "" {
		switch c.SameSite {
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

//UserConf  returns the current user config
func UserConf() interface{} {
	appConf := Conf()
	if appConf != nil && appConf.Conf != nil {
		return appConf.Conf
	}
	return nil
}

//LogConf  returns the current log config
func LogConf() *logs.LogConf {
	appConf := Conf()
	if appConf != nil && appConf.Log != nil {
		return appConf.Log
	}
	return nil
}

//Path  returns the current config path
func Path() string {
	return flagConf
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

//RegistConf regist handler conf
func RegistConf(handle ConfingHandle, finish ...FinishHandle) {
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
