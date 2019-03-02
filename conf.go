package bast

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aixiaoxiang/bast/logs"
)

var (
	confObj    *AppConfMgr
	confHandle ConfHandle
)

//ConfHandle handler  conf
type ConfHandle func(appConf *AppConf) error

//AppConfMgr app config
type AppConfMgr struct {
	frist *AppConf
	Confs map[string]*AppConf
}

//AppConf  app config item
type AppConf struct {
	Key     string        `json:"key"`
	Name    string        `json:"name"`
	Addr    string        `json:"addr"`
	FileDir string        `json:"fileDir"`
	Debug   bool          `json:"debug"`
	BaseURL string        `json:"baseUrl"`
	Log     *logs.LogConf `json:"log"`
	Conf    interface{}   `json:"conf"`
	Extend  string        `json:"extend"`
}

//ConfItem default db config
type ConfItem struct {
	Name      string `json:"name"`
	Title     string `json:"dbTitle"`
	User      string `json:"dbUser"`
	Pwd       string `json:"dbPwd"`
	Server    string `json:"dbServer"`
	Charset   string `json:"dbCharset"`
	ParseTime *bool  `json:"dbParseTime"`
	Loc       string `json:"dbLoc"`
}

//ConfMgr  config
func ConfMgr() *AppConfMgr {
	if confObj == nil {
		data, err := ioutil.ReadFile(ConfPath())
		if err != nil {
			return nil
		}
		appConf := []AppConf{}
		err = json.Unmarshal(data, &appConf)
		if err != nil {
			logs.Err("conf mgr init error", err)
			return nil
		}
		ConfInit(appConf)
	}
	return confObj
}

//ConfInit  config
func ConfInit(appConf []AppConf) {
	lg := len(appConf)
	if lg == 0 {
		return
	}
	if confObj == nil {
		confObj = &AppConfMgr{}
		confObj.Confs = make(map[string]*AppConf)
		for i := 0; i < lg; i++ {
			c := &appConf[i]
			if confHandle != nil {
				err := confHandle(c)
				if err != nil {
					continue
				}
			}
			if c.Key == flagAppKey && confObj.frist == nil {
				confObj.frist = c
			}
			confObj.Confs[c.Key] = c
		}
		if confObj.frist == nil {
			confObj.frist = &appConf[0]
		}
	}
}

//ConfOK check conf
func ConfOK() bool {
	return ConfMgr() != nil
}

//Conf returns the current app config
func Conf() *AppConf {
	appConf := ConfMgr()
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

//ConfPath  returns the current config path
func ConfPath() string {
	return flagConf
}

//confParse parse config path
func confParse(f *flag.FlagSet) string {
	exPath := filepath.Dir(os.Args[0])
	fs := f.Lookup("conf")
	if fs != nil {
		flagConf = fs.Value.String()
	}
	if flagConf == "" {
		cf := exPath + "/config.conf"
		flagConf = cf
	}
	return flagConf
}

//RegistConfHandle handler  conf
func RegistConfHandle(f ConfHandle) {
	confHandle = f
}
