// Copyright 2018 The axx Authors. All rights reserved.

package bast

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/aixiaoxiang/bast/guid"
	"github.com/aixiaoxiang/bast/ids"
	"github.com/aixiaoxiang/bast/logs"
	"github.com/aixiaoxiang/bast/session"
	"github.com/aixiaoxiang/bast/validate"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
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
	KindAcceptJSON          = 0       // json
	KindAcceptXML           = 1       // xml
	KindAcceptYAML          = 2       // yaml
)

//default validator
var valid = validate.Validator{}

//Context is app Context
type Context struct {
	//In A Request represents an HTTP request received by a server
	// or to be sent by a client.
	In *http.Request
	//Accept
	Accept string
	//Kind Accept
	KindAccept int
	//Out A ResponseWriter interface is used by an HTTP handler to
	// construct an HTTP response.
	Out http.ResponseWriter
	//Params httprouter Params,/:name/:age
	Params httprouter.Params
	//isParseForm Parse tag
	isParseForm bool
	//NeedAuthorization is need authorization
	NeedAuthorization bool
	//IsAuthorization is authorization finish?
	IsAuthorization bool
	//Session is session
	Session session.Store
}

//Message is response message
type Message struct {
	Code int    `gorm:"-" json:"code"`
	Msg  string `gorm:"-" json:"msg"`
}

//MessageDetail is response detail message
type MessageDetail struct {
	Code   int    `gorm:"-" json:"code"`
	Msg    string `gorm:"-" json:"msg"`
	Detail string `gorm:"-" json:"detail"`
}

//Data is response data
type Data struct {
	Message `gorm:"-"`
	Data    interface{} `gorm:"-"  json:"data"`
}

//Pagination is Pagination data
type Pagination struct {
	Message
	Data  interface{} `gorm:"-"  json:"data"`
	Page  int         `gorm:"-"  json:"page"`
	Total int         `gorm:"-"  json:"total"`
}

//InvalidPagination is invalid Pagination data
type InvalidPagination struct {
	Pagination
	Invalid bool `gorm:"-"  json:"invalid"`
	Fix     bool `gorm:"-"  json:"fix"`
}

/**********data  start**********/

//Data  output Data data to client
//v data
func (c *Context) Data(v interface{}) {
	c.DataWithCodeMsg(v, SerOK, "")
}

//DataWithCode output Data data to client
//v data
//code is message code
func (c *Context) DataWithCode(v interface{}, code int) {
	c.DataWithCodeMsg(v, code, "")
}

//DataWithMsg output  data to client
//v data
//msg is string message
func (c *Context) DataWithMsg(v interface{}, msg string) {
	c.DataWithCodeMsg(v, SerOK, msg)
}

//DataWithCodeMsg output data to client
//v data
//code is message code
//msg is string message
func (c *Context) DataWithCodeMsg(v interface{}, code int, msg string) {
	c.DataResult(c.ObjWithCodeMsg(c, code, msg))
}

//Pagination output pagination data data to client
//v data
//page is page
//total is total row count
func (c *Context) Pagination(v interface{}, page, total int) {
	c.PaginationCodeMsg(v, page, total, SerOK, "")
}

//PaginationCode output pagination  data to client
//v data
//page is page
//total is total row count
//code is message code
func (c *Context) PaginationCode(v interface{}, page, total, code int) {
	c.PaginationCodeMsg(v, page, total, code, "")
}

//PaginationCodeMsg output pagination  data to client
//v data
//page is page
//total is total row count
//code is message code
//msg is string message
func (c *Context) PaginationCodeMsg(v interface{}, page, total, code int, msg string) {
	c.DataResult(c.ObjWithPaginationCodeMsg(v, page, total, code, msg))
}

//DataResult output Data data to client
func (c *Context) DataResult(v interface{}) {
	switch c.KindAccept {
	case KindAcceptJSON:
		c.JSONResult(v)
		break
	case KindAcceptXML:
		c.XMLResult(v)
		break
	case KindAcceptYAML:
		c.YAMLResult(v)
		break
	}
}

/**********data  end**********/

/**********json  start**********/

//JSON  output JSON data to client
//v data
func (c *Context) JSON(v interface{}) {
	c.JSONWithCodeMsg(v, SerOK, "")
}

//JSONWithCode output JSON data to client
//v data
//code is message code
func (c *Context) JSONWithCode(v interface{}, code int) {
	c.JSONWithCodeMsg(v, code, "")
}

//JSONWithMsg output JSON data to client
//v data
//msg is string message
func (c *Context) JSONWithMsg(v interface{}, msg string) {
	c.JSONWithCodeMsg(v, SerOK, msg)
}

//JSONWithCodeMsg output JSON data to client
//v data
//code is message code
//msg is string message
func (c *Context) JSONWithCodeMsg(v interface{}, code int, msg string) {
	c.JSONResult(c.ObjWithCodeMsg(c, code, msg))
}

//JSONWithPagination output pagination JSON data to client
//v data
//page is page
//total is total row count
func (c *Context) JSONWithPagination(v interface{}, page, total int) {
	c.JSONWithPaginationCodeMsg(v, page, total, SerOK, "")
}

//JSONWithPaginationCode output pagination JSON data to client
//v data
//page is page
//total is total row count
//code is message code
func (c *Context) JSONWithPaginationCode(v interface{}, page, total, code int) {
	c.JSONWithPaginationCodeMsg(v, page, total, code, "")
}

//JSONWithPaginationCodeMsg output pagination JSON data to client
//v data
//page is page
//total is total row count
//code is message code
//msg is string message
func (c *Context) JSONWithPaginationCodeMsg(v interface{}, page, total, code int, msg string) {
	c.JSONResult(c.ObjWithPaginationCodeMsg(v, page, total, code, msg))
}

//JSONResult output json data to client
func (c *Context) JSONResult(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		logs.Errors("JSONResult error", err)
		c.StatusCode(http.StatusInternalServerError)
		return
	}
	c.Out.Header().Set("Content-Type", "application/json")
	c.Out.Write(data)
	data = nil
}

/**********json  end**********/

/**********xml  start**********/

//XML  output XML data to client
//v data
func (c *Context) XML(v interface{}) {
	c.XMLWithCodeMsg(v, SerOK, "")
}

//XMLWithCode output XML data to client
//v data
//code is message code
func (c *Context) XMLWithCode(v interface{}, code int) {
	c.XMLWithCodeMsg(v, code, "")
}

//XMLWithMsg output XML data to client
//v data
//msg is string message
func (c *Context) XMLWithMsg(v interface{}, msg string) {
	c.XMLWithCodeMsg(v, SerOK, msg)
}

//XMLWithCodeMsg output XML data to client
//v data
//code is message code
//msg is string message
func (c *Context) XMLWithCodeMsg(v interface{}, code int, msg string) {
	c.XMLResult(c.ObjWithCodeMsg(c, code, msg))
}

//XMLWithPagination output pagination XML data to client
//v data
//page is page
//total is total row count
func (c *Context) XMLWithPagination(v interface{}, page, total int) {
	c.XMLWithPaginationCodeMsg(v, page, total, SerOK, "")
}

//XMLWithPaginationCode output pagination XML data to client
//v data
//page is page
//total is total row count
//code is message code
func (c *Context) XMLWithPaginationCode(v interface{}, page, total, code int) {
	c.XMLWithPaginationCodeMsg(v, page, total, code, "")
}

//XMLWithPaginationCodeMsg output pagination XML data to client
//v data
//page is page
//total is total row count
//code is message code
//msg is string message
func (c *Context) XMLWithPaginationCodeMsg(v interface{}, page, total, code int, msg string) {
	c.XMLResult(c.ObjWithPaginationCodeMsg(v, page, total, code, msg))
}

//XMLResult output xml data to client
func (c *Context) XMLResult(v interface{}) {
	data, err := xml.Marshal(v)
	if err != nil {
		logs.Errors("XMLResult error", err)
		c.StatusCode(http.StatusInternalServerError)
		return
	}
	c.Out.Header().Set("Content-Type", "application/xml")
	c.Out.Write(data)
	data = nil
}

/**********xml  end**********/

/**********yaml  start**********/

//YAML  output YAML data to client
//v data
func (c *Context) YAML(v interface{}) {
	c.YAMLWithCodeMsg(v, SerOK, "")
}

//YAMLWithCode output YAML data to client
//v data
//code is message code
func (c *Context) YAMLWithCode(v interface{}, code int) {
	c.YAMLWithCodeMsg(v, code, "")
}

//YAMLWithMsg output YAML data to client
//v data
//msg is string message
func (c *Context) YAMLWithMsg(v interface{}, msg string) {
	c.YAMLWithCodeMsg(v, SerOK, msg)
}

//YAMLWithCodeMsg output YAML data to client
//v data
//code is message code
//msg is string message
func (c *Context) YAMLWithCodeMsg(v interface{}, code int, msg string) {
	c.YAMLResult(c.ObjWithCodeMsg(c, code, msg))
}

//YAMLWithPagination output pagination YAML data to client
//v data
//page is page
//total is total row count
func (c *Context) YAMLWithPagination(v interface{}, page, total int) {
	c.YAMLWithPaginationCodeMsg(v, page, total, SerOK, "")
}

//YAMLWithPaginationCode output pagination YAML data to client
//v data
//page is page
//total is total row count
//code is message code
func (c *Context) YAMLWithPaginationCode(v interface{}, page, total, code int) {
	c.YAMLWithPaginationCodeMsg(v, page, total, code, "")
}

//YAMLWithPaginationCodeMsg output pagination YAML data to client
//v data
//page is page
//total is total row count
//code is message code
//msg is string message
func (c *Context) YAMLWithPaginationCodeMsg(v interface{}, page, total, code int, msg string) {
	c.YAMLResult(c.ObjWithPaginationCodeMsg(v, page, total, code, msg))
}

//YAMLResult output yaml data to client
func (c *Context) YAMLResult(v interface{}) {
	data, err := yaml.Marshal(v)
	if err != nil {
		logs.Errors("YAMLResult error", err)
		c.StatusCode(http.StatusInternalServerError)
		return
	}
	c.Out.Header().Set("Content-Type", "application/x+yaml")
	c.Out.Write(data)
	data = nil
}

/**********yaml  end**********/

//ObjWithCodeMsg return obj data
//v data
//code is message code
//msg is string message
func (c *Context) ObjWithCodeMsg(v interface{}, code int, msg string) interface{} {
	_, ok := v.(*Data)
	if !ok {
		_, ok = v.(*Message)
	}
	if !ok {
		_, ok = v.(*Pagination)
	}
	if !ok {
		_, ok = v.(*MessageDetail)
	}
	if !ok {
		_, ok = v.(*InvalidPagination)
	}
	if !ok {
		d := &Data{}
		d.Code = code
		d.Msg = msg
		d.Data = v
		return d
	}
	return v
}

//ObjWithPaginationCodeMsg return pagination obj data
//v data
//page is page
//total is total row count
//code is message code
//msg is string message
func (c *Context) ObjWithPaginationCodeMsg(v interface{}, page, total, code int, msg string) interface{} {
	d := &InvalidPagination{}
	_, _total, pageRow := c.Page()
	if _total == 0 {
		last := int(math.Ceil(float64(total) / float64(pageRow)))
		if page >= last {
			page = last - 1
			d.Fix = true
		}
	}
	page++
	if v != nil {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Array:
		case reflect.Map:
		case reflect.Slice:
			s := reflect.ValueOf(v)
			if s.Len() == 0 {
				d.Invalid = true
			}
			break
		}
	} else {
		d.Invalid = true
	}
	d.Data = v
	d.Page = page
	d.Total = total
	d.Code = code
	d.Msg = msg
	if d.Invalid || d.Fix {
		return d
	}
	return d.Pagination
}

//Success output success result to client
//	msg is success message
func (c *Context) Success(msg string) {
	v := &Message{}
	v.Code = SerOK
	v.Msg = msg
	c.Data(v)
}

//Successf output success result and format to client
func (c *Context) Successf(format string, a ...interface{}) {
	if a != nil && len(a) > 0 {
		format = fmt.Sprintf(format, a...)
	}
	v := &Message{}
	v.Code = SerOK
	v.Msg = format
	c.Data(v)
}

//Failed  output failed result to client
//param:
//	msg is fail message
//	err error
func (c *Context) Failed(msg string, err ...error) {
	c.FailResult(msg, SerError, err...)
}

//Faileds  output failed detail result to client
//param:
//	msg is fail message
//	detail is detail message
func (c *Context) Faileds(msg string, detail string) {
	v := &MessageDetail{}
	v.Code = SerError
	v.Msg = msg
	v.Detail = detail
	c.DataWithCode(v, SerError)
}

//Failedf output failed result and format to client
func (c *Context) Failedf(format string, a ...interface{}) {
	var err error
	if a != nil {
		lg := len(a)
		if lg > 0 {
			if a[lg-1] != nil {
				err, _ = a[lg-1].(error)
			}
			if err != nil {
				a = a[0 : lg-1]
			}
			if len(a) > 0 {
				format = fmt.Sprintf(format, a...)
			}
		}
	}
	c.FailResult(format, SerError, err)
}

//Result  output result to client
//param:
//	msg is fail message
//	detail is detail message
func (c *Context) Result(msg string, detail ...string) {
	v := &MessageDetail{}
	v.Code = SerOK
	v.Msg = msg
	if detail != nil {
		for _, s := range detail {
			if v.Detail != "" {
				v.Detail += ","
			}
			v.Detail += s
		}
	}
	c.DataWithCode(v, SerError)
	v = nil
}

//FailResult output fail result to client
//param:
//	msg failed message
//	errCode ailed message code
//  err  error
func (c *Context) FailResult(msg string, errCode int, err ...error) {
	v := &Message{}
	if errCode == 0 {
		errCode = SerError
	}
	v.Code = errCode
	v.Msg = msg
	if err != nil && err[0] != nil {
		v.Msg += ", [" + err[0].Error() + "]"
	}
	c.DataWithCode(v, errCode)
}

//SignOut output user signout to client
//param:
//	msg message
func (c *Context) SignOut(msg string) {
	c.FailResult(msg, SerSignOutError)
}

//NoData output no data result to client
//param:
//	err message
func (c *Context) NoData(msg string) {
	c.FailResult(msg, SerNoDataError)
}

//Say output raw bytes to client
//param:
//	data raw bytes
func (c *Context) Say(data []byte) {
	c.Out.Write(data)
}

//Says output string to client
//param:
//	str string
func (c *Context) Says(str string) {
	c.Out.Write([]byte(str))
}

//SendFile send file to client
//param:
//	fileName is file name
//  rawFileName is raw file name
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

//JSON2String JSON to string
//param:
//	obj is object
func (c *Context) JSON2String(obj interface{}) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", nil
	}
	return string(data), err
}

//String2JSON string to JSON
//param:
//	str json string
//  obj is object
func (c *Context) String2JSON(str string, obj interface{}) error {
	return c.JSONDecode(strings.NewReader(str), obj)
}

//Verify verify current request
//param:
//rules is validate rule such as:
// 	key1@required|int|min:1
// 	key2/key2_translator@required|string|min:1
//	key3@sometimes|required|data
func (c *Context) Verify(rules ...string) error {
	c.In.ParseForm()
	return valid.Request(c.In.Form, rules...)
}

//StatusCode set current request statusCode
//param:
//	statusCode HTTP status code. such as: 200x,300x and so on
func (c *Context) StatusCode(statusCode int) {
	c.Out.WriteHeader(statusCode)
	c.Out.Write([]byte(http.StatusText(statusCode)))
}

//RawString getter raw string value from current request(request body)
func (c *Context) RawString() string {
	body, err := ioutil.ReadAll(c.In.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

//GetString  gets a string value from  the current request  based on the key
//param:
//	key is key name
func (c *Context) GetString(key string) string {
	d := c.GetStrings(key)
	if len(d) > 0 {
		return d[0]
	}
	return ""
}

//GetTrimString  Use the key to get a non-space string value from the current request
//param:
//	key is key name
func (c *Context) GetTrimString(key string) string {
	return strings.TrimSpace(c.GetString(key))
}

//GetStringSlice Use the key to get all string value from the current request
//param:
//	key is key name
//  sep spilt char
func (c *Context) GetStringSlice(key, sep string) []string {
	s := c.GetTrimString(key)
	if len(s) > 0 {
		ss := strings.Split(s, sep)
		if len(ss) > 0 {
			return ss
		}
	}
	return nil
}

//GetIntSlice Use the key to get all int value from the current request
//param:
//	key is key name
//  sep spilt char
func (c *Context) GetIntSlice(key, sep string) []int64 {
	s := c.GetTrimString(key)
	if len(s) > 0 {
		ss := strings.Split(s, sep)
		lg := len(ss)
		si := make([]int64, lg, lg)
		for i := 0; i < lg; i++ {
			si[i], _ = strconv.ParseInt(ss[i], 10, 64)
		}
		if len(si) > 0 {
			return si
		}
	}
	return nil
}

//GetIntSliceAndRemovePrefix Use the key to get all int value from the current request and remove prefix of each
//param:
//	key is key name
//  sep spilt char
//  prefix	remove prefix string
func (c *Context) GetIntSliceAndRemovePrefix(key, sep, prefix string) ([]int64, bool) {
	s := c.GetTrimString(key)
	has := false
	if len(s) > 0 {
		if prefix != "" {
			has = strings.HasPrefix(s, prefix)
			s = strings.TrimPrefix(s, prefix)
		}
		ss := strings.Split(s, sep)
		lg := len(ss)
		si := make([]int64, 0, lg)
		for i := 0; i < lg; i++ {
			n, err := strconv.ParseInt(ss[i], 10, 64)
			if err == nil {
				si = append(si, n)
			}
		}
		if len(si) > 0 {
			return si, has
		}
	}
	return nil, false
}

//GetParam  Use the key to get all int value from the current request url
//note：xx/:name/:name2
//param:
//	key key name
func (c *Context) GetParam(key string) string {
	return c.Params.ByName(key)
}

//GetLeftLikeString get a sql(left like 'xx%') string value from the current request  based on the key
//param:
//	key is key name
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

//GetRightLikeString get a sql(right like '%xx') string value from the current request  based on the key
//param:
//	key is key name
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

//GetLikeString  get a sql(like '%xx%') string value from the current request  based on the key
//param:
//	key is key name
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

//GetBool get a bool value  from the current request  based on the key
//param:
//	key is key name
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

//GetBoolValue get a bool value from the current request  based on the key
//param:
//	key is key name
//  def is default value
func (c *Context) GetBoolValue(key string, def bool) bool {
	d := c.GetStrings(key)
	if len(d) > 0 {
		ok, err := strconv.ParseBool(d[0])
		if err == nil {
			return ok
		}
	}
	return def
}

//GetStrings gets strings from the current request based on the key
//param:
//	key is key name
func (c *Context) GetStrings(key string) []string {
	c.ParseForm()
	return c.In.Form[key]
}

//GetInt gets a int value from the current request  based on the key
//param:
//	key is key name
//	def default value
func (c *Context) GetInt(key string) (int, error) {
	return strconv.Atoi(c.GetString(key))
}

//GetIntValue gets a int value  from the current request  based on the key（errors not included）
//param:
//	key is key name
//	def default value
func (c *Context) GetIntValue(key string, def int) int {
	d := c.GetString(key)
	v, err := strconv.Atoi(d)
	if err != nil {
		v = def
	}
	return v
}

//GetInt64 gets a int64 value  from the current request url  based on the key
//param:
//	key is key name
//	def default value
func (c *Context) GetInt64(key string) (int64, error) {
	return strconv.ParseInt(c.GetString(key), 10, 64)
}

//GetInt64Value gets a int64 value  from the current request  based on the key（errors not included）
//param:
//	key is key name
//	def default value
func (c *Context) GetInt64Value(key string, def int64) int64 {
	d := c.GetString(key)
	v, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		v = def
	}
	return v
}

//GetFloat gets a float value  from the current request uri  based on the key
//param:
//	key is key name
//	def default value
func (c *Context) GetFloat(key string) (float64, error) {
	return strconv.ParseFloat(c.GetString(key), 64)
}

//GetFloatValue gets a float value  from the current request  based on the key（errors not included）
//param:
//	key is key name
//	def default value
func (c *Context) GetFloatValue(key string, def float64) float64 {
	d := c.GetString(key)
	v, err := strconv.ParseFloat(d, 64)
	if err != nil {
		v = def
	}
	return v
}

//HasParam has a param from the current request based on the key(May not have a value)
//param:
//	key is key name
func (c *Context) HasParam(key string) bool {
	c.ParseForm()
	_, ok := c.In.Form[key]
	return ok
}

//Form gets all form params from the current(uri not included)
func (c *Context) Form() url.Values {
	c.ParseForm()
	return c.In.Form
}

//PostForm gets all form params from the current(uri and form)
func (c *Context) PostForm() url.Values {
	c.ParseForm()
	return c.In.PostForm
}

//Query gets all query params from the current request url
func (c *Context) Query() url.Values {
	c.ParseForm()
	return c.In.URL.Query()
}

//Page get pages param from the current request
//param:
//	page 	current page index(start 1)
//	total 	all data total count(cache total count for first service return)
//  pageRow page maximum size(default is 100 row)
func (c *Context) Page() (page int, total int, pageRow int) {
	page = c.GetIntValue("page", 0)
	total = c.GetIntValue("total", 0)
	pageRow = c.GetIntValue("pageRow", 100)
	if page > 0 {
		page--
	}
	if pageRow > 100 {
		pageRow = 100
	} else if pageRow <= 0 {
		pageRow = 100
	}
	return page, total, pageRow
}

//Pages get pages param from the current request and check last page
//param:
//	page 	 current page index(start 1)
//	total 	 all data total count(cache total count for first service return)
//  pageRow  page maximum size(default is 100 row)
func (c *Context) Pages() (page, total, pageRow int) {
	page, total, pageRow = c.Page()
	if total > 0 {
		last := int(math.Ceil(float64(total) / float64(pageRow)))
		if page >= last {
			total = 0
		}
	}
	return page, total, pageRow
}

//Offset return page offset
//param:
//	total 	all data total count(cache total count for first service return)
func (c *Context) Offset(total int) int {
	page, _total, pageRow := c.Page()
	if _total == 0 {
		last := int(math.Ceil(float64(total) / float64(pageRow)))
		if page >= last {
			page = last - 1
		}
	}
	offset := page * pageRow
	return offset
}

//Obj gets data from the current request body(json or xml or yaml fromat) and convert it to a objecet
//param:
//	obj 	target object
//  verify	verify obj
func (c *Context) Obj(obj interface{}, verify ...bool) error {
	switch c.KindAccept {
	case KindAcceptJSON:
		return c.JSONObj(obj, verify...)
	case KindAcceptXML:
		return c.XMLObj(obj, verify...)
	case KindAcceptYAML:
		return c.YAMLObj(obj, verify...)
	}
	return c.JSONObj(obj, verify...)

}

//JSONObj gets data from the current request body(json fromat) and convert it to a objecet
//param:
//	obj 	target object
//  verify	verify obj
func (c *Context) JSONObj(obj interface{}, verify ...bool) error {
	err := c.JSONDecode(c.In.Body, obj)
	if err == nil && verify != nil && verify[0] {
		err = valid.Struct(obj)
	}
	return err
}

//JSONDecode gets data from the r reader(json fromat) and convert it to a objecet
//param:
//	r is a reader
//	obj target object
func (c *Context) JSONDecode(r io.Reader, obj interface{}) (err error) {
	if app.Debug {
		body, err := ioutil.ReadAll(r)
		if err != nil {
			logs.Debug("JSONDecode error", zap.Error(err), zap.ByteString("detail", body))
			return err
		}
		err = json.Unmarshal(body, obj)
		body = nil
	} else {
		err = json.NewDecoder(r).Decode(obj)
	}
	if err != nil {
		logs.Debug("JSONDecode error", zap.Error(err))
	}
	return err
}

//XMLObj gets data from the current request(xml format) and convert it to a object
//param:
//	obj 	target object
//  verify	verify obj
func (c *Context) XMLObj(obj interface{}, verify ...bool) error {
	err := c.XMLDecode(c.In.Body, obj)
	if err == nil && verify != nil && verify[0] {
		err = valid.Struct(obj)
	}
	return err
}

//XMLDecode  gets data from the r reader(xml format) and convert it to a object
//param:
//	r is a reader
//	obj target object
func (c *Context) XMLDecode(r io.Reader, obj interface{}) (err error) {
	if app.Debug {
		body, err := ioutil.ReadAll(r)
		if err != nil {
			logs.Debug("XMLDecode error", zap.Error(err), zap.ByteString("detail", body))
			return err
		}
		err = xml.Unmarshal(body, obj)
		body = nil
	} else {
		err = xml.NewDecoder(r).Decode(obj)
	}
	if err != nil {
		logs.Debug("XMLDecode error", zap.Error(err))
	}
	return err
}

//YAMLObj gets data from the current request(yaml format) and convert it to a object
//param:
//	obj 	target object
//  verify	verify obj
func (c *Context) YAMLObj(obj interface{}, verify ...bool) error {
	err := c.YAMLDecode(c.In.Body, obj)
	if err == nil && verify != nil && verify[0] {
		err = valid.Struct(obj)
	}
	return err
}

//YAMLDecode  gets data from the r reader(yaml format) and convert it to a object
//param:
//	r is a reader
//	obj target object
func (c *Context) YAMLDecode(r io.Reader, obj interface{}) (err error) {
	if app.Debug {
		body, err := ioutil.ReadAll(r)
		if err != nil {
			logs.Debug("YAMLDecode error", zap.Error(err), zap.ByteString("detail", body))
			return err
		}
		err = yaml.Unmarshal(body, obj)
		body = nil
	} else {
		err = yaml.NewDecoder(r).Decode(obj)
	}
	if err != nil {
		logs.Debug("YAMLDecode error", zap.Error(err))
	}
	return err
}

//MapObj gets current request body and convert it to a map
func (c *Context) MapObj() map[string]interface{} {
	result := make(map[string]interface{})
	err := c.Obj(result)
	if err != nil {
		logs.Debug("MapObj error", zap.Error(err))
		return nil
	}
	return result
}

// ParseForm populates r.Form and r.PostForm.
//
// For all requests, ParseForm parses the raw query from the URL and updates
// r.Form.
//
// For POST, PUT, and PATCH requests, it also parses the request body as a form
// and puts the results into both r.PostForm and r.Form. Request body parameters
// take precedence over URL query string values in r.Form.
//
// For other HTTP methods, or when the Content-Type is not
// application/x-www-form-urlencoded, the request Body is not read, and
// r.PostForm is initialized to a non-nil, empty value.
//
// If the request Body's size has not already been limited by MaxBytesReader,
// the size is capped at 10MB.
//
// ParseMultipartForm calls ParseForm automatically.
// ParseForm is idempotent.
func (c *Context) ParseForm() {
	if !c.isParseForm {
		c.In.ParseForm()
		c.isParseForm = true
	}
}

// ParseMultipartForm parses a request body as multipart/form-data.
// The whole request body is parsed and up to a total of maxMemory bytes of
// its file parts are stored in memory, with the remainder stored on
// disk in temporary files.
// ParseMultipartForm calls ParseForm if necessary.
// After one call to ParseMultipartForm, subsequent calls have no effect.
func (c *Context) ParseMultipartForm(maxMemory int64) error {
	return c.In.ParseMultipartForm(maxMemory)
}

//SessionRead get session value by key
func (c *Context) SessionRead(key string) interface{} {
	if c.Session != nil {
		return c.Session.Get(key)
	}
	return nil
}

//SessionWrite set session value by key
func (c *Context) SessionWrite(key string, value interface{}) error {
	if c.Session != nil {
		return c.Session.Set(key, value)
	}
	return nil
}

//SessionDelete delete session value by key
func (c *Context) SessionDelete(key string) error {
	if c.Session != nil {
		return c.Session.Delete(key)
	}
	return nil
}

//SessionClear delete all session
func (c *Context) SessionClear() error {
	if c.Session != nil {
		return c.Session.Clear()
	}
	return nil
}

//SessionID get sessionID
func (c *Context) SessionID() string {
	if c.Session != nil {
		return c.Session.ID()
	}
	return ""
}

//URL get eequest url
func (c *Context) URL() string {
	return strings.Join([]string{c.BaseURL(), c.In.RequestURI}, "")
}

//DefaultFileURL returns full file url
//param:
//	url is relative path
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

//BaseURL gets root url(scheme+host) from current request
//param:
//	url relative path
func (c *Context) BaseURL(url ...string) string {
	baseURL := c.In.Header.Get("BaseUrl")
	if baseURL != "" {
		return baseURL + strings.Join(url, "")
	}
	return c.baseURL() + "/" + strings.Join(url, "")
}

//baseURL gets root url(scheme+host) from current request
func (c *Context) baseURL() string {
	scheme := "http://"
	if c.In.TLS != nil {
		scheme = "https://"
	}
	return strings.Join([]string{scheme, c.In.Host}, "")
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

//Redirect redirect
func (c *Context) Redirect(url string) {
	http.Redirect(c.Out, c.In, url, http.StatusFound)
}

//TemporaryRedirect redirect(note: 307 redirect，Can avoid data loss after POST redirection)
func (c *Context) TemporaryRedirect(url string) {
	http.Redirect(c.Out, c.In, url, http.StatusTemporaryRedirect)
}

//Reset current context to pool
func (c *Context) Reset() {
	c.isParseForm = false
	c.In = nil
	c.Out = nil
	c.Params = nil
	c.isParseForm = false
	c.NeedAuthorization = false
	c.IsAuthorization = false
	c.Session = nil
	c.Accept = ""
	c.KindAccept = 0
}

//ID return a ID
func (c *Context) ID() int64 {
	return ids.ID()
}

//GUID return a GUID
func (c *Context) GUID() string {
	return guid.GUID()
}

// Exist returns a boolean indicating whether the error is known to
// report that a file or directory does not exist. It is satisfied by
// ErrNotExist as well as some syscall errors.
func (c *Context) Exist(path string) bool {
	return PathExist(path)
}
