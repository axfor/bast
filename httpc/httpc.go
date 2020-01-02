//Copyright 2018 The axx Authors. All rights reserved.

package httpc

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	//DefaultRetries default retries 3
	DefaultRetries = 3
	//DefaultCookieJar default cookie jar
	DefaultCookieJar http.CookieJar
	//DefaultTransport default transport
	defaultTransport *http.Transport
	//DefaultClient HTTP  clientdefault
	defaultClient *http.Client
	before        func(*Client) error
	after         func(*Client)
)

//Client http client
type Client struct {
	Req       *http.Request
	files     map[string]string
	params    url.Values
	client    *http.Client
	settings  Settings
	Tag       string
	resp      *http.Response
	err       error
	body      []byte
	bodyClose bool
}

//Settings of Client
type Settings struct {
	Transport    http.RoundTripper
	EnableCookie bool
	Retry        int
}

// MarkTag sets an tag field
func (c *Client) MarkTag(tag string) *Client {
	c.Tag = tag
	return c
}

// Client set http.Client
func (c *Client) Client(client *http.Client) *Client {
	if client != nil {
		c.client = client
	} else {
		c.client = defaultClient
	}
	return c
}

// Request set http.Request
func (c *Client) Request(request *http.Request) *Client {
	if request != nil {
		c.Req = request
	}
	return c
}

// Transport specifies the mechanism by which individual
// HTTP requests are made.
// If nil, DefaultTransport is used.
func (c *Client) Transport(transport http.RoundTripper) *Client {
	c.settings.Transport = transport
	return c
}

// NewTransport new *http.Transport
func NewTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// Settings set settings
// If nil, DefaultSettings is used.
func (c *Client) Settings(settings Settings) *Client {
	c.settings = settings
	return c
}

// Retries  maximum retries count
func (c *Client) Retries(retries int) *Client {
	c.settings.Retry = retries
	return c
}

// EnableCookie set enable cookie jar
func (c *Client) EnableCookie(enableCookie bool) *Client {
	c.settings.EnableCookie = enableCookie
	return c
}

// UserAgent sets User-Agent header field
func (c *Client) UserAgent(useragent string) *Client {
	if c.Req.UserAgent() == "" {
		c.Req.Header.Set("User-Agent", useragent)
	}
	return c
}

// Accept sets Accept header field
func (c *Client) Accept(accept string) *Client {
	c.Req.Header.Set("Accept", accept)
	return c
}

// SetBasicAuth sets the request's Authorization header to use HTTP
// Basic Authentication with the provided username and password.
//
// With HTTP Basic Authentication the provided username and password
// are not encrypted.
//
// Some protocols may impose additional requirements on pre-escaping the
// username and password. For instance, when used with OAuth2, both arguments
// must be URL encoded first with url.QueryEscape.
func (c *Client) SetBasicAuth(username, password string) *Client {
	c.Req.SetBasicAuth(username, password)
	return c
}

//CheckRedirect proxy method http.Get
func (c *Client) CheckRedirect(checkRedirect func(req *http.Request, via []*http.Request) error) *Client {
	c.client.CheckRedirect = checkRedirect
	return c
}

//Jar proxy method http.Get
func (c *Client) Jar(jar http.CookieJar) *Client {
	c.client.Jar = jar
	if jar != nil {
		c.settings.EnableCookie = true
	}
	return c
}

// Header add header item string in request.
func (c *Client) Header(key, value string) *Client {
	c.Req.Header.Set(key, value)
	if key == "Cookie" {
		c.settings.EnableCookie = true
	}
	return c
}

// Cookie add cookie into request.
func (c *Client) Cookie(cookie *http.Cookie) *Client {
	c.Req.AddCookie(cookie)
	if cookie != nil {
		c.settings.EnableCookie = true
	}
	return c
}

// File add file into request.
func (c *Client) File(fileName, path string) *Client {
	c.files[fileName] = path
	return c
}

// Files add files into request.
func (c *Client) Files(paths map[string]string) *Client {
	for k, v := range paths {
		c.files[k] = v
	}
	return c
}

// Param adds query param in to request.
// params build query string as ?key1=value1&key2=value2...
func (c *Client) Param(key, value string) *Client {
	if c.params == nil {
		c.params = url.Values{}
	}
	if param, ok := c.params[key]; ok {
		c.params[key] = append(param, value)
	} else {
		c.params[key] = []string{value}
	}
	return c
}

// Params adds query param in to request.
// params build query string as ?key1=value1&key2=value2...
func (c *Client) Params(values url.Values) *Client {
	c.params = values
	return c
}

// Body adds request raw body.
func (c *Client) Body(data interface{}) *Client {
	switch t := data.(type) {
	case string:
		c.Req.Body = ioutil.NopCloser(bytes.NewBufferString(t))
		c.Req.ContentLength = int64(len(t))
		return c
	case []byte:
		c.Req.Body = ioutil.NopCloser(bytes.NewBuffer(t))
		c.Req.ContentLength = int64(len(t))
		return c
	case uint, uint8, uint16, uint32, uint64:
	case int, int8, int16, int32, int64:
	case float32, float64:
	default:
		return c
	}
	v := fmt.Sprint(data)
	c.Req.Body = ioutil.NopCloser(bytes.NewBufferString(v))
	c.Req.ContentLength = int64(len(v))
	return c
}

// JSONBody adds request raw body encoding by JSON.
func (c *Client) JSONBody(obj interface{}) *Client {
	_, err := c.JSONBodyWithError(obj)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

// JSONBodyWithError adds request raw body encoding by JSON.
func (c *Client) JSONBodyWithError(obj interface{}) (*Client, error) {
	if c.Req.Body == nil && obj != nil {
		data, err := json.Marshal(obj)
		if err != nil {
			return c, err
		}
		c.BodyWithContentType(data, "application/json")
	}
	return c, nil
}

// XMLBody adds request raw body encoding by XML.
func (c *Client) XMLBody(obj interface{}) *Client {
	_, err := c.XMLBodyWithError(obj)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

// XMLBodyWithError adds request raw body encoding by XML.
func (c *Client) XMLBodyWithError(obj interface{}) (*Client, error) {
	if c.Req.Body == nil && obj != nil {
		data, err := xml.Marshal(obj)
		if err != nil {
			return c, err
		}
		c.BodyWithContentType(data, "application/xml")
	}
	return c, nil
}

// YAMLBody adds request raw body encoding by YAML.
func (c *Client) YAMLBody(obj interface{}) *Client {
	_, err := c.YAMLBodyWithError(obj)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

// YAMLBodyWithError adds request raw body encoding by YAML.
func (c *Client) YAMLBodyWithError(obj interface{}) (*Client, error) {
	if c.Req.Body == nil && obj != nil {
		data, err := yaml.Marshal(obj)
		if err != nil {
			return c, err
		}
		c.BodyWithContentType(data, "application/x+yaml")
	}
	return c, nil
}

// BodyWithContentType adds request raw body encoding by XML.
func (c *Client) BodyWithContentType(data []byte, contentType string) *Client {
	if c.Req.Body == nil && data != nil && len(data) > 0 {
		c.Req.Body = ioutil.NopCloser(bytes.NewReader(data))
		c.Req.ContentLength = int64(len(data))
		c.Req.Header.Set("Content-Type", contentType)
	}
	return c
}

func (c *Client) Error() error {
	if c.err != nil {
		return c.err
	}
	return nil
}

// HasError has error
func (c *Client) HasError() bool {
	return c.err != nil
}

// OK status code is 200
func (c *Client) OK() bool {
	return c.err == nil && c.resp != nil && c.resp.StatusCode == http.StatusOK
}

func (c *Client) String() (string, error) {
	data, err := c.Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Result returns the map that marshals from the body bytes as json or xml or yaml in response .
// default json
func (c *Client) Result(v interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	contentType := c.resp.Header.Get("Content-Type")
	if contentType == "" || strings.HasPrefix(contentType, "application/json") {
		return json.Unmarshal(data, v)
	} else if strings.HasPrefix(contentType, "application/xml") {
		return xml.Unmarshal(data, v)
	} else if strings.HasPrefix(contentType, "application/x+yaml") {
		return yaml.Unmarshal(data, v)
	}
	return json.Unmarshal(data, v)
}

// ToJSON returns the map that marshals from the body bytes as json in response .
// it calls Response inner.
func (c *Client) ToJSON(v interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToMap returns the map that marshals from the body bytes as json in response .
// it calls Response inner.
func (c *Client) ToMap(v *map[string]interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML returns the map that marshals from the body bytes as xml in response .
// it calls Response inner.
func (c *Client) ToXML(v interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// ToYAML returns the map that marshals from the body bytes as yaml in response .
// it calls Response inner.
func (c *Client) ToYAML(v interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (c *Client) ToFile(filename string) error {
	_, err := c.getResponse()
	if err != nil {
		return err
	}
	if c.resp.Body == nil {
		return nil
	}
	defer c.resp.Body.Close()
	c.bodyClose = true
	err = pathExistAndMkdir(filename)
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, c.resp.Body)
	return err
}

//Check that the file directory exists, there is no automatically created
func pathExistAndMkdir(filename string) (err error) {
	filename = path.Dir(filename)
	_, err = os.Stat(filename)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(filename, os.ModePerm)
		if err == nil {
			return nil
		}
	}
	return err
}

// Bytes returns the body []byte in response.
// it calls Response inner.
func (c *Client) Bytes() ([]byte, error) {
	if c.body != nil {
		return c.body, nil
	}
	_, err := c.getResponse()
	if err != nil {
		return nil, err
	}
	if c.resp.Body == nil {
		return nil, errors.New("empty body")
	}
	defer c.resp.Body.Close()
	c.bodyClose = true
	if c.resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(c.resp.Body)
		if err != nil {
			return nil, err
		}
		c.body, err = ioutil.ReadAll(reader)
		return c.body, err
	}
	c.body, err = ioutil.ReadAll(c.resp.Body)
	return c.body, err
}

func (c *Client) getResponse() (*http.Response, error) {
	if c.resp != nil {
		return c.resp, nil
	}
	if c.resp == nil {
		r, err := c.Do()
		if err == nil {
			return r, nil
		}
		return nil, err
	}
	return nil, errors.New("empty response")
}

// Get issues a GET to the specified URL. If the response is one of
// the following redirect codes, Get follows the redirect, up to a
// maximum of 10 redirects:
//
//    301 (Moved Permanently)
//    302 (Found)
//    303 (See Other)
//    307 (Temporary Redirect)
//    308 (Permanent Redirect)
//
// An error is returned if there were too many redirects or if there
// was an HTTP protocol error. A non-2xx response doesn't cause an
// error. Any returned error will be of type *url.Error. The url.Error
// value's Timeout method will report true if request timed out or was
// canceled.
//
// When err is nil, resp always contains a non-nil resp.Body.
// Caller should close resp.Body when done reading from it.
//
// Get is a wrapper around DefaultClient.Get.
//
// To make a request with custom headers, use NewRequest and
// DefaultClient.Do.
func Get(url string) *Client {
	return createRequest(url, http.MethodGet)
}

// Post issues a POST to the specified URL.
//
// Caller should close resp.Body when done reading from it.
//
// If the provided body is an io.Closer, it is closed after the
// request.
//
// Post is a wrapper around DefaultClient.Post.
//
// To set custom headers, use NewRequest and DefaultClient.Do.
//
// See the Client.Do method documentation for details on how redirects
// are handled.
func Post(url string) *Client {
	return createRequest(url, http.MethodPost)
}

// Head issues a HEAD to the specified URL. If the response is one of
// the following redirect codes, Head follows the redirect, up to a
// maximum of 10 redirects:
//
//    301 (Moved Permanently)
//    302 (Found)
//    303 (See Other)
//    307 (Temporary Redirect)
//    308 (Permanent Redirect)
//
// Head is a wrapper around DefaultClient.Head
func Head(url string) *Client {
	return createRequest(url, http.MethodHead)
}

// Put returns *HTTPClient with PUT method
func Put(url string) *Client {
	return createRequest(url, http.MethodPut)
}

// Delete returns *HTTPClient with Delete method
func Delete(url string) *Client {
	return createRequest(url, http.MethodDelete)
}

// Patch returns *HTTPClient with Patch method
func Patch(url string) *Client {
	return createRequest(url, http.MethodPatch)
}

func createRequest(uri, method string) *Client {
	u, _ := url.Parse(uri)
	c := &Client{
		client: defaultClient,
		Req: &http.Request{
			URL:        u,
			Method:     method,
			Header:     make(http.Header),
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
		},
		err: nil,
	}
	return c
}

// Do add files into request
func (c *Client) Do() (resp *http.Response, err error) {
	c.resp = nil
	if c.err != nil {
		return nil, c.err
	}
	err = c.build()
	if err != nil {
		return
	}
	err = callBefore(c)
	if err != nil {
		return
	}
	for i := 0; c.settings.Retry == -1 || i <= c.settings.Retry; i++ {
		resp, err = c.client.Do(c.Req)
		if err == nil {
			break
		}
	}
	c.resp = resp
	callAfter(c)
	return
}

func (c *Client) build() error {
	if c.settings.EnableCookie && c.client.Jar == nil {
		c.client.Jar = DefaultCookieJar
	}

	if c.settings.Transport != nil {
		c.client.Transport = c.settings.Transport
	} else {
		c.client.Transport = defaultTransport
	}
	var urlParam string
	var buf bytes.Buffer
	for k, v := range c.params {
		for i, vv := range v {
			if i > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(url.QueryEscape(k))
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(vv))
		}
	}
	urlParam = buf.String()
	has := len(c.params) > 0
	if c.Req.Method == http.MethodGet {
		if has {
			rurl := c.Req.URL.String()
			if strings.Contains(rurl, "?") {
				rurl += "&" + urlParam
			} else {
				rurl += "?" + urlParam
			}
			urls, err := url.Parse(rurl)
			if err != nil {
				return err
			}
			c.Req.URL = urls
		}
		return nil
	}
	if (c.Req.Method == http.MethodPost || c.Req.Method == http.MethodPut || c.Req.Method == http.MethodPatch ||
		c.Req.Method == http.MethodDelete) && c.Req.Body == nil {
		if len(c.files) > 0 {
			bodyBuffer := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuffer)
			for formname, filename := range c.files {
				fileWriter, err := bodyWriter.CreateFormFile(formname, filename)
				if err != nil {
					return err
				}
				fh, err := os.Open(filename)
				if err != nil {
					return err
				}
				//iocopy
				_, err = io.Copy(fileWriter, fh)
				fh.Close()
				if err != nil {
					return err
				}
			}
			for k, v := range c.params {
				for _, vv := range v {
					bodyWriter.WriteField(k, vv)
				}
			}
			bodyWriter.Close()
			c.Header("Content-Type", bodyWriter.FormDataContentType())
			c.Req.Body = ioutil.NopCloser(bodyBuffer)
		} else if has {
			c.Header("Content-Type", "application/x-www-form-urlencoded")
			c.Body(urlParam)
		}
	}
	return nil
}

//Clear all response data
func (c *Client) Clear() *Client {
	if c.resp != nil && !c.bodyClose {
		c.resp.Body.Close()
	}
	c.resp = nil
	c.body = nil
	c.bodyClose = false
	c.err = nil
	c.files = map[string]string{}
	c.params = url.Values{}
	c.settings = Settings{}
	c.Tag = ""
	return c
}

func callBefore(c *Client) error {
	if before != nil {
		return before(c)
	}
	return nil
}

func callAfter(c *Client) {
	if after != nil {
		after(c)
	}
}

//Before before handler for each network request
func Before(f func(*Client) error) {
	before = f
}

//After after handler for each network request
func After(f func(*Client)) {
	after = f
}

//init
func init() {
	DefaultCookieJar, _ = cookiejar.New(nil)
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		// dereference it to get a copy of the struct that the pointer points to
		defaultTransport = &(*t)
		defaultTransport.MaxIdleConns = 100
		defaultTransport.MaxIdleConnsPerHost = 100
	}
	// dereference it to get a copy of the struct that the pointer points to
	defaultClient = &(*http.DefaultClient)
}
