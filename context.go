// Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aixiaoxiang/bast/guid"
	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/logs"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

//const code
const (
	SerError                = 0       // error code
	SerOK                   = 1       // ok code
	SerDBError              = -10000  // db error code
	SerNoDataError          = -20000  // no data error code
	SerSignOutError         = -30000  // user sign out error code
	SerUserNotExistError    = -40000  // user not exist code
	SerInvalidParamError    = -50000  // invalid param  code
	SerInvalidUserAuthorize = -60000  // invalid user authorize  code
	SerExist                = -70000  // exist code
	SerNotExist             = -80000  // not exist code
	SerTry                  = -99999  // please try code
	SerMustFailed           = -111111 // must failed code
	SerFailed               = -222222 // failed code
	SerAuthorizationFailed  = -888888 // authorization failed code
)

//Context is app Context
type Context struct {
	//In A Request represents an HTTP request received by a server
	// or to be sent by a client.
	In *http.Request
	//Out A ResponseWriter interface is used by an HTTP handler to
	// construct an HTTP response.
	Out http.ResponseWriter
	//Params httprouter Params,/:name/:age
	Params httprouter.Params
	//isParseForm Parse tag
	isParseForm bool
	//Authorization is need authorization
	Authorization bool
	//IsAuthorization is authorization finish?
	IsAuthorization bool
}

//Msgs is response message
type Msgs struct {
	Code int    `gorm:"-" json:"code"`
	Msg  string `gorm:"-" json:"msg"`
}

//Data is response data
type Data struct {
	Msgs `gorm:"-"`
	Data interface{} `gorm:"-"  json:"data"`
}

//DataPage is Pagination data
type DataPage struct {
	Msgs
	Data  interface{} `gorm:"-"  json:"data"`
	Page  int         `gorm:"-"  json:"page"`
	Total int         `gorm:"-"  json:"total"`
}

/******Output method **********/

//JSON  output JSON Data to client
//v data
func (c *Context) JSON(v interface{}) {
	c.JSONWithCodeMsg(v, SerOK, "")
}

//JSONWithCode output JSON Data to client
//v data
//code is message code
func (c *Context) JSONWithCode(v interface{}, code int) {
	c.JSONWithCodeMsg(v, code, "")
}

//JSONWithMsg output JSON Data to client
//v data
//msg is string message
func (c *Context) JSONWithMsg(v interface{}, msg string) {
	c.JSONWithCodeMsg(v, SerOK, msg)
}

//JSONWithCodeMsg output JSON Data to client
//v data
//code is message code
//msg is string message
func (c *Context) JSONWithCodeMsg(v interface{}, code int, msg string) {
	_, isData := v.(*Data)
	if !isData {
		_, isData = v.(*Msgs)
	}
	if !isData {
		_, isData = v.(*DataPage)
	}
	if !isData {
		d := &Data{}
		d.Code = code
		d.Msg = msg
		d.Data = v
		c.JSONResult(d)
		d.Data = nil
		d = nil
	} else {
		c.JSONResult(v)
	}
}

//JSONWithPage output Pagination JSON Data to client
func (c *Context) JSONWithPage(v interface{}, page, total int) {
	c.JSONWithPageAndCodeMsg(v, page, total, SerOK, "")
}

//JSONWithPageAndCode output Pagination JSON Data to client
func (c *Context) JSONWithPageAndCode(v interface{}, page, total, code int, msg string) {
	c.JSONWithPageAndCodeMsg(v, page, total, code, msg)
}

//JSONWithPageAndCodeMsg output Pagination JSON Data to client
func (c *Context) JSONWithPageAndCodeMsg(v interface{}, page, total, code int, msg string) {
	d := &DataPage{}
	d.Data = v
	d.Page = page
	d.Total = total
	d.Code = code
	d.Msg = msg
	c.JSONResult(d)
	d.Data = nil
	d = nil
}

//JSONResult output Data to client
func (c *Context) JSONResult(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		logs.Info("JSONResult-Err=" + err.Error())
		c.StatusCode(http.StatusInternalServerError)
		return
	}
	c.Out.Header().Set("Content-Type", "application/json")
	c.Out.Write(data)
	data = nil
}

//Success output success result to client
//	msg is success message
func (c *Context) Success(msg string) {
	d := &Msgs{}
	d.Code = SerOK
	d.Msg = msg
	c.JSON(d)
	d = nil
}

//Failed 输出错误的JSON格式对象
//param:
//	msg is fail 消息
//	err error 消息
func (c *Context) Failed(msg string, err ...error) {
	c.FailResult(msg, SerError, err...)
}

//SignOutError 输出用户登出信息
//param:
//	msg is fail 消息
func (c *Context) SignOutError(msg string) {
	c.FailResult(msg, SerSignOutError)
}

//DBError 输出数据库错误信息
//param:
//	err db.error
func (c *Context) DBError(err error) {
	msg := "操作数据库错误"
	if err != nil {
		msg = "操作数据库错误,详情：" + err.Error()
	} else {
		msg = "操作数据库错误"
	}
	c.FailResult(msg, SerDBError)
}

//FailResult 输出通用的错误的消息
//param:
//	msg 失败/错误消息
//	errCode 失败/错误代码
//  err  error
func (c *Context) FailResult(msg string, errCode int, err ...error) {
	d := &Msgs{}
	if errCode == 0 {
		errCode = SerError
	}
	d.Code = errCode
	d.Msg = msg
	if err != nil && err[0] != nil {
		d.Msg += ",详情 ：" + err[0].Error()
	}
	c.JSON(d)
	d = nil
}

//NoData 输出无数据消息
//param:
//	err 消息
func (c *Context) NoData(msg ...string) {
	msgs := "抱歉！暂无数据"
	if msg == nil {
		msgs = "抱歉！暂无数据"
	} else {
		msgs = msg[0]
	}
	c.FailResult(msgs, SerNoDataError)
}

//Say 输出字节流数据
//param:
//	data 数据
func (c *Context) Say(data []byte) {
	c.Out.Write(data)
}

//SayStr 输出字符串信息
//param:
//	str 消息
func (c *Context) SayStr(str string) {
	c.Out.Write([]byte(str))
}

//SendFile 发送文件
//param:
//	fileName 文件全路径
//  rawFileName 文件原始名称,用于下载文件时的别名
func (c *Context) SendFile(fileName string, rawFileName ...string) {
	dir := filepath.Dir(fileName)
	fileName = filepath.Base(fileName)
	url := c.BaseURL("f/" + fileName)
	fileName = "/f/" + fileName
	fs := http.StripPrefix("/f/", http.FileServer(http.Dir(dir)))
	r, _ := http.NewRequest("GET", url, nil)
	raw := fileName
	if rawFileName != nil {
		raw = rawFileName[0]
		c.Out.Header().Set("Content-Disposition", "attachment; filename="+raw)
	}
	fs.ServeHTTP(c.Out, r)
	r = nil
	fs = nil
}

//JSONToStr JSON对象转化为字符串
//param:
//	obj 对象
func (c *Context) JSONToStr(obj interface{}) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", nil
	}
	return string(data), err
}

//StrToJSON 将字符串转化为JSON对象
//param:
//	str json字符串
//  obj 外部对象
func (c *Context) StrToJSON(str string, obj interface{}) error {
	return c.JSONDecode(strings.NewReader(str), obj)
}

//StatusCode 设置状态码
//param:
//	statusCode 状态代码
func (c *Context) StatusCode(statusCode int) {
	c.Out.WriteHeader(statusCode)
	c.Out.Write([]byte(http.StatusText(statusCode)))
}

//******get resuest data method **********/

//GetRawStr 获取请求体并转化为字符串
func (c *Context) GetRawStr() string {
	body, err := ioutil.ReadAll(c.In.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

//GetString 获取请求信息里面指定参数值
//param:
//	key 键值
func (c *Context) GetString(key string) string {
	d := c.GetStrings(key)
	if len(d) > 0 {
		return d[0]
	}
	return ""
}

//GetTrimString 获取请求信息里面指定参数值并移除空格
//param:
//	key 键值
func (c *Context) GetTrimString(key string) string {
	return strings.TrimSpace(c.GetString(key))
}

//GetStringSlice 获取请求信息里面指定参数值并以指定的字符分割成字符串数组
//param:
//	key 键值
//  sep 分割字符串
func (c *Context) GetStringSlice(key, sep string) *[]string {
	d := c.GetStrings(key)
	if len(d) > 0 {
		s := strings.Split(d[0], sep)
		return &s
	}
	return nil
}

//GetIntSlice 获取请求信息里面指定参数值并以指定的字符分割成整数数组
//param:
//	key 键值
//  sep 分割字符串
func (c *Context) GetIntSlice(key, sep string) *[]int64 {
	d := c.GetStrings(key)
	if len(d) > 0 {
		s := strings.Split(d[0], sep)
		lg := len(s)
		si := make([]int64, lg, lg)
		for i := 0; i < lg; i++ {
			si[i], _ = strconv.ParseInt(s[i], 10, 64)
		}
		return &si
	}
	return nil
}

//GetIntSliceAndRemovePrefix 获取请求信息里面指定参数值并以指定的字符分割成整数数组，并尝试移除前缀字符
//param:
//	key 	键值
//  sep 	分割字符串
//  prefix 	待移除的前缀字符
func (c *Context) GetIntSliceAndRemovePrefix(key, sep, prefix string) (*[]int64, bool) {
	d := c.GetStrings(key)
	has := false
	if len(d) > 0 {
		ss := d[0]
		if prefix != "" {
			has = strings.HasPrefix(ss, prefix)
			ss = strings.TrimPrefix(d[0], prefix)
		}
		s := strings.Split(ss, sep)
		lg := len(s)
		si := make([]int64, 0, lg)
		for i := 0; i < lg; i++ {
			n, err := strconv.ParseInt(s[i], 10, 64)
			if err == nil {
				si = append(si, n)
			}
		}
		if len(si) > 0 {
			return &si, has
		}
	}
	return nil, false
}

//GetParam 获取请求里面的路由参数值
//说明：xx/:name/:name2 里面的:name与:name2就是路由参数占位符
//param:
//	key 键值
func (c *Context) GetParam(key string) string {
	return c.Params.ByName(key)
}

//GetLeftLikeString 获取请求信息里面指定参数值并生成左匹配sql条件
//param:
//	key 键值
func (c *Context) GetLeftLikeString(key string) string {
	d := c.GetStrings(key)
	if len(d) > 0 {
		r := d[0]
		if r != "" {
			return r + "%"
		}
	}
	return ""
}

//GetRightLikeString 获取请求信息里面指定参数值并生成右匹配sql条件
//param:
//	key 键值
func (c *Context) GetRightLikeString(key string) string {
	d := c.GetStrings(key)
	if len(d) > 0 {
		r := d[0]
		if r != "" {
			return "%" + r
		}
	}
	return ""
}

//GetLikeString 获取请求信息里面指定参数值并生成左右匹配sql条件
//param:
//	key 键值
func (c *Context) GetLikeString(key string) string {
	d := c.GetStrings(key)
	if len(d) > 0 {
		r := d[0]
		if r != "" {
			return "%" + r + "%"
		}
	}
	return ""
}

//GetBool 获取请求信息里面指定参数值并转化位bool
//param:
//	key 键值
func (c *Context) GetBool(key string) bool {
	d := c.GetStrings(key)
	if len(d) > 0 {
		ok, err := strconv.ParseBool(d[0])
		if err == nil {
			return ok
		}
	}
	return false
}

//GetBoolWithDefault 获取请求信息里面指定参数值并转化位bool
//param:
//	key 键值
func (c *Context) GetBoolWithDefault(key string, defaults bool) bool {
	d := c.GetStrings(key)
	if len(d) > 0 {
		ok, err := strconv.ParseBool(d[0])
		if err == nil {
			return ok
		}
	}
	return defaults
}

//GetStrings 获取请求信息里面指定参数值的字符
//param:
//	key 键值
func (c *Context) GetStrings(key string) []string {
	c.ParseForm()
	return c.In.Form[key]
}

//Form 获取请求参数(含表单与URL)
func (c *Context) Form() url.Values {
	c.ParseForm()
	return c.In.Form
}

//PostForm 获取表单请求参数
func (c *Context) PostForm() url.Values {
	c.ParseForm()
	return c.In.PostForm
}

//Query 获取请求URL参数
func (c *Context) Query() url.Values {
	c.ParseForm()
	return c.In.URL.Query()
}

//GetInt 获取请求信息里面指定参数值并转化位int
//param:
//	key 键值
//	def 默认值
func (c *Context) GetInt(key string, def ...int) (int, error) {
	d := c.GetString(key)
	v, err := strconv.Atoi(d)
	if err != nil {
		if def != nil && len(def) > 0 {
			v = def[0]
		} else {
			v = 0
		}
	}
	return v, err
}

//GetIntVal 获取请求信息里面指定参数值并转化位int（不含错误信息）
//param:
//	key 键值
//	def 默认值
func (c *Context) GetIntVal(key string, def ...int) int {
	d := c.GetString(key)
	v, err := strconv.Atoi(d)
	if err != nil {
		if def != nil && len(def) > 0 {
			v = def[0]
		} else {
			v = 0
		}
	}
	return v
}

//GetInt64 获取请求信息里面指定参数值并转化位int64
//param:
//	key 键值
//	def 默认值
func (c *Context) GetInt64(key string, def ...int64) (int64, error) {
	d := c.GetString(key)
	v, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		if def != nil && len(def) > 0 {
			v = def[0]
		} else {
			v = 0
		}
	}
	return v, err
}

//GetFloat 获取请求信息里面指定参数值并转化位float64
//param:
//	key 键值
//	def 默认值
func (c *Context) GetFloat(key string, def ...float64) (float64, error) {
	d := c.GetString(key)
	v, err := strconv.ParseFloat(d, 64)
	if err != nil {
		if def != nil && len(def) > 0 {
			v = def[0]
		} else {
			v = 0
		}
	}
	return v, err
}

//Pages 获取请求信息里面的分页相关参数
//param:
//	page 	当前页
//	total 	总行数
//  pageRow 每页行数
func (c *Context) Pages() (page int, total int, pageRow int) {
	page, _ = c.GetInt("page")
	total, _ = c.GetInt("total")
	pageRow, _ = c.GetInt("pageRow", 100)
	if pageRow > 100 {
		pageRow = 100
	} else if pageRow <= 0 {
		pageRow = 100
	}
	return page, total, pageRow
}

//JSONObj 将当前请求流JSON格式转化为对象
//param:
//	obj 外部对象
func (c *Context) JSONObj(obj interface{}) error {
	return c.JSONDecode(c.In.Body, obj)
}

//GetJSON 将当前请求流JSON格式转化为对象
//param:
//	obj 外部对象
func (c *Context) GetJSON(obj interface{}) error {
	return c.JSONObj(obj)
}

//JSONDecode 将请求流转化为对象
//param:
//	r reader 阅读流
//	obj 外部对象
func (c *Context) JSONDecode(r io.Reader, obj interface{}) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, obj)
	// logs.Debug("JSONDecode=" + string(body))
	if err != nil {
		if app.Debug {
			logs.Debug("JSONDecode-Err=" + err.Error() + ",detail=" + string(body))
		} else {
			logs.Debug("JSONDecode-Err=" + err.Error())
		}
		body = nil
		return err
	}
	return err
}

//XMLObj 将当前请求流XML格式转化为对象
//param:
//	obj 外部对象
func (c *Context) XMLObj(obj interface{}) error {
	return c.XMLDecode(c.In.Body, obj)
}

//GetXML 将当前请求流XML格式转化为对象
//param:
//	obj 外部对象
func (c *Context) GetXML(obj interface{}) error {
	return c.XMLObj(obj)
}

//XMLDecode 将请求流XML格式转化为对象
//param:
//	r reader 阅读流
//	obj 外部对象
func (c *Context) XMLDecode(r io.Reader, obj interface{}) error {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(body, obj)
	if err != nil {
		if app.Debug {
			logs.Debug("XMLDecode-Err=" + err.Error() + ",detail=" + string(body))
		} else {
			logs.Debug("XMLDecode-Err=" + err.Error())
		}
		body = nil
		return err
	}
	return err
}

//MapObj 将请求流转化为字典对象
func (c *Context) MapObj() map[string]interface{} {
	body, _ := ioutil.ReadAll(c.In.Body)
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(body), &result)
	if err == nil {
		return result
	}
	return nil
}

//ParseForm 解析请求表单信息
func (c *Context) ParseForm() {
	//没解析则解析
	if !c.isParseForm {
		c.In.ParseForm()
		c.isParseForm = true
	}
}

//ParseMultipartForm 解析请求多表单信息
//param:
//	maxMemory 最大内存大小
func (c *Context) ParseMultipartForm(maxMemory int64) error {
	return c.In.ParseMultipartForm(maxMemory)
}

//URL 获取请求的完整URL
func (c *Context) URL() string {
	return strings.Join([]string{c.BaseURL(), c.In.RequestURI}, "")
}

//DefaultFileURL returns full file url
//param:
//	url is relative url
func (c *Context) DefaultFileURL(url string) string {
	if url != "" {
		if url[0] != 'f' {
			url = "f/" + url
		}
		baseURL := c.In.Header.Get("BaseUrl")
		if baseURL != "" {
			return baseURL + url
		}
		return c.baseURL() + "/" + url
	}
	return ""
}

//BaseURL 获取请求的基URL
//param:
//	url 相对地址
func (c *Context) BaseURL(url ...string) string {
	baseURL := c.In.Header.Get("BaseUrl")
	if baseURL != "" {
		return baseURL + strings.Join(url, "")
	}
	return c.baseURL() + "/" + strings.Join(url, "")
}

//baseURL 获取请求的基URL-内部使用
func (c *Context) baseURL() string {
	scheme := "http://"
	if c.In.TLS != nil {
		scheme = "https://"
	}
	return strings.Join([]string{scheme, c.In.Host}, "")
}

//Redirect 重定向
func (c *Context) Redirect(url string) {
	http.Redirect(c.Out, c.In, url, http.StatusFound)
}

//ClientIP return request client ip
func (c *Context) ClientIP() string {
	ps := c.Proxys()
	if len(ps) > 0 && ps[0] != "" {
		realIP, _, err := net.SplitHostPort(ps[0])
		if err != nil {
			realIP = ps[0]
		}
		return realIP
	}
	if ip, _, err := net.SplitHostPort(c.In.RemoteAddr); err == nil {
		return ip
	}
	return c.In.RemoteAddr
}

// Proxys return request proxys
// if request header has X-Real-IP, return it
// if request header has X-Forwarded-For, return it
func (c *Context) Proxys() []string {
	if v := c.In.Header.Get("X-Real-IP"); v != "" {
		return strings.Split(v, ",")
	}
	if v := c.In.Header.Get("X-Forwarded-For"); v != "" {
		return strings.Split(v, ",")
	}
	return []string{}
}

//TemporaryRedirect 重定向(307重定向，可以避免POST重定向后数据丢失)
func (c *Context) TemporaryRedirect(url string) {
	http.Redirect(c.Out, c.In, url, http.StatusTemporaryRedirect)
}

//Reset 重置请求与响应对象
func (c *Context) Reset() {
	c.isParseForm = false
	c.In = nil
	c.Out = nil
	c.Params = nil
	c.isParseForm = false
	c.Authorization = false
	c.IsAuthorization = false
}

/******log method **********/

//I info日志记录
func (c *Context) I(msg string, fields ...zap.Field) {
	logs.I(msg, fields...)
}

//D debug日志记录
func (c *Context) D(msg string, fields ...zap.Field) {
	logs.D(msg, fields...)
}

//E Error日志记录
func (c *Context) E(msg string, fields ...zap.Field) {
	logs.E(msg, fields...)
}

//Err Error日志记录
func (c *Context) Err(msg string, err error) {
	if msg == "" {
		msg = "发生错误"
	}
	if err != nil {
		msg += "，详情：" + err.Error()
	}
	logs.E(msg)
}

/******ID method **********/

//ID 快捷ID生成
func (c *Context) ID() int64 {
	return ids.ID()
}

/******GUID method **********/

//GUID 快捷创建一个GUID
func (c *Context) GUID() string {
	return guid.GUID()
}

//Exist 判断文件夹是否存在
func (c *Context) Exist(path string) bool {
	return PathExist(path)
}
