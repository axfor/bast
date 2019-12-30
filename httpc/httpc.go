//Copyright 2018 The axx Authors. All rights reserved.

package httpc

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
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
	defaultRetries   = 3
	defaultTimeout   = Timeout(60*time.Second, 60*time.Second)
	defaultCookieJar http.CookieJar
	before           func(*HTTPClient) error
	after            func(*HTTPClient)
)

//HTTPClient http client
type HTTPClient struct {
	client             http.Client
	Request            *http.Request
	response           *http.Response
	insecureSkipVerify bool
	files              map[string]string
	params             url.Values
	Retry              int
	body               []byte
	EnableCookie       bool
	err                error
	Tag                string
}

// MarkTag sets an tag field
func (c *HTTPClient) MarkTag(tag string) *HTTPClient {
	c.Tag = tag
	return c
}

// Transport specifies the mechanism by which individual
// HTTP requests are made.
// If nil, DefaultTransport is used.
func (c *HTTPClient) Transport(transport http.RoundTripper) *HTTPClient {
	c.client.Transport = transport
	if t, ok := c.client.Transport.(*http.Transport); ok {
		if t.TLSClientConfig != nil {
			c.insecureSkipVerify = t.TLSClientConfig.InsecureSkipVerify
		}
	}
	return c
}

// Proxy specifies a function to return a proxy for a given
// Request. If the function returns a non-nil error, the
// request is aborted with the provided error.
//
// The proxy type is determined by the URL scheme. "http",
// "https", and "socks5" are supported. If the scheme is empty,
// "http" is assumed.
//
// If Proxy is nil or returns a nil *URL, no proxy is used.
func (c *HTTPClient) Proxy(proxy func(*http.Request) (*url.URL, error)) *HTTPClient {
	if t, ok := c.client.Transport.(*http.Transport); ok {
		t.Proxy = proxy
	}
	return c
}

// MaxIdleConnsPerHost ,if non-zero, controls the maximum idle
// (keep-alive) connections to keep per-host. If zero,
// DefaultMaxIdleConnsPerHost is used.
func (c *HTTPClient) MaxIdleConnsPerHost(maxIdleConnsPerHost int) *HTTPClient {
	if t, ok := c.client.Transport.(*http.Transport); ok {
		t.MaxIdleConnsPerHost = maxIdleConnsPerHost
	}
	return c
}

// Retries  maximum retries count
func (c *HTTPClient) Retries(retries int) *HTTPClient {
	c.Retry = retries
	return c
}

// UserAgent sets User-Agent header field
func (c *HTTPClient) UserAgent(useragent string) *HTTPClient {
	if c.Request.UserAgent() == "" {
		c.Request.Header.Set("User-Agent", useragent)
	}
	return c
}

// Accept sets Accept header field
func (c *HTTPClient) Accept(accept string) *HTTPClient {
	c.Request.Header.Set("Accept", accept)
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
func (c *HTTPClient) SetBasicAuth(username, password string) *HTTPClient {
	c.Request.SetBasicAuth(username, password)
	return c
}

// Dial specifies the dial function for creating unencrypted TCP connections.
//
// Dial runs concurrently with calls to RoundTrip.
// A RoundTrip call that initiates a dial may end up using
// a connection dialed previously when the earlier connection
// becomes idle before the later Dial completes.
//
// Deprecated: Use DialContext instead, which allows the transport
// to cancel dials as soon as they are no longer needed.
// If both are set, DialContext takes priority.
func (c *HTTPClient) Dial(dial func(netw, addr string) (net.Conn, error)) *HTTPClient {
	if t, ok := c.client.Transport.(*http.Transport); ok {
		t.Dial = dial
	}
	return c
}

// TLSClientConfig sets tls connection configurations if visiting https url.
func (c *HTTPClient) TLSClientConfig(config *tls.Config) *HTTPClient {
	if t, ok := c.client.Transport.(*http.Transport); ok {
		t.TLSClientConfig = config
	}
	return c
}

// SkipVerify controls whether a client verifies the
// server's certificate chain and host name.
// If InsecureSkipVerify is true, TLS accepts any certificate
// presented by the server and any host name in that certificate.
// In this mode, TLS is susceptible to man-in-the-middle attacks.
// This should be used only for testing.
func (c *HTTPClient) SkipVerify(insecureSkipVerify bool) *HTTPClient {
	c.insecureSkipVerify = insecureSkipVerify
	c.skipVerify(insecureSkipVerify)
	return c
}

func (c *HTTPClient) skipVerify(insecureSkipVerify bool) *HTTPClient {
	if t, ok := c.client.Transport.(*http.Transport); ok {
		if t.TLSClientConfig == nil {
			t.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: c.insecureSkipVerify,
			}
		} else {
			t.TLSClientConfig.InsecureSkipVerify = c.insecureSkipVerify
		}
	}
	return c
}

//CheckRedirect proxy method http.Get
func (c *HTTPClient) CheckRedirect(checkRedirect func(req *http.Request, via []*http.Request) error) *HTTPClient {
	c.client.CheckRedirect = checkRedirect
	return c
}

//Jar proxy method http.Get
func (c *HTTPClient) Jar(jar http.CookieJar) *HTTPClient {
	c.client.Jar = jar
	if jar != nil {
		c.EnableCookie = true
	}
	return c
}

// Header add header item string in request.
func (c *HTTPClient) Header(key, value string) *HTTPClient {
	c.Request.Header.Set(key, value)
	if key == "Cookie" {
		c.EnableCookie = true
	}
	return c
}

// Cookie add cookie into request.
func (c *HTTPClient) Cookie(cookie *http.Cookie) *HTTPClient {
	c.Request.AddCookie(cookie)
	if cookie != nil {
		c.EnableCookie = true
	}
	return c
}

// File add file into request.
func (c *HTTPClient) File(fileName, path string) *HTTPClient {
	c.files[fileName] = path
	return c
}

// Files add files into request.
func (c *HTTPClient) Files(paths map[string]string) *HTTPClient {
	for k, v := range paths {
		c.files[k] = v
	}
	return c
}

// Param adds query param in to request.
// params build query string as ?key1=value1&key2=value2...
func (c *HTTPClient) Param(key, value string) *HTTPClient {
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
func (c *HTTPClient) Params(values url.Values) *HTTPClient {
	c.params = values
	return c
}

// Body adds request raw body.
func (c *HTTPClient) Body(data interface{}) *HTTPClient {
	switch t := data.(type) {
	case string:
		c.Request.Body = ioutil.NopCloser(bytes.NewBufferString(t))
		c.Request.ContentLength = int64(len(t))
	case []byte:
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(t))
		c.Request.ContentLength = int64(len(t))
	case uint, uint8, uint16, uint32, uint64:
	case int, int8, int16, int32, int64:
	case float32, float64:
	default:
		return c
	}
	v := fmt.Sprint(data)
	c.Request.Body = ioutil.NopCloser(bytes.NewBufferString(v))
	c.Request.ContentLength = int64(len(v))
	return c
}

// JSONBody adds request raw body encoding by JSON.
func (c *HTTPClient) JSONBody(obj interface{}) *HTTPClient {
	_, err := c.JSONBodyWithError(obj)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

// JSONBodyWithError adds request raw body encoding by JSON.
func (c *HTTPClient) JSONBodyWithError(obj interface{}) (*HTTPClient, error) {
	if c.Request.Body == nil && obj != nil {
		data, err := json.Marshal(obj)
		if err != nil {
			return c, err
		}
		c.BodyWithContentType(data, "application/json")
	}
	return c, nil
}

// XMLBody adds request raw body encoding by XML.
func (c *HTTPClient) XMLBody(obj interface{}) *HTTPClient {
	_, err := c.XMLBodyWithError(obj)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

// XMLBodyWithError adds request raw body encoding by XML.
func (c *HTTPClient) XMLBodyWithError(obj interface{}) (*HTTPClient, error) {
	if c.Request.Body == nil && obj != nil {
		data, err := xml.Marshal(obj)
		if err != nil {
			return c, err
		}
		c.BodyWithContentType(data, "application/xml")
	}
	return c, nil
}

// YAMLBody adds request raw body encoding by YAML.
func (c *HTTPClient) YAMLBody(obj interface{}) *HTTPClient {
	_, err := c.YAMLBodyWithError(obj)
	if err != nil {
		c.err = err
		return c
	}
	return c
}

// YAMLBodyWithError adds request raw body encoding by YAML.
func (c *HTTPClient) YAMLBodyWithError(obj interface{}) (*HTTPClient, error) {
	if c.Request.Body == nil && obj != nil {
		data, err := yaml.Marshal(obj)
		if err != nil {
			return c, err
		}
		c.BodyWithContentType(data, "application/x+yaml")
	}
	return c, nil
}

// BodyWithContentType adds request raw body encoding by XML.
func (c *HTTPClient) BodyWithContentType(data []byte, contentType string) *HTTPClient {
	if c.Request.Body == nil && data != nil && len(data) > 0 {
		c.Request.Body = ioutil.NopCloser(bytes.NewReader(data))
		c.Request.ContentLength = int64(len(data))
		c.Request.Header.Set("Content-Type", contentType)
	}
	return c
}

func (c *HTTPClient) Error() error {
	if c.err != nil {
		return c.err
	}
	return nil
}

// HasError has error
func (c *HTTPClient) HasError() bool {
	return c.err != nil
}

// OK status code is 200
func (c *HTTPClient) OK() bool {
	return c.err == nil && c.response != nil && c.response.StatusCode == http.StatusOK
}

func (c *HTTPClient) String() (string, error) {
	data, err := c.Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Result returns the map that marshals from the body bytes as json or xml or yaml in response .
// default json
func (c *HTTPClient) Result(v interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	contentType := c.response.Header.Get("Content-Type")
	if contentType == "" || strings.Index(contentType, "application/json") >= 0 {
		return json.Unmarshal(data, v)
	} else if strings.Index(contentType, "application/xml") >= 0 {
		return xml.Unmarshal(data, v)
	} else if strings.Index(contentType, "application/x+yaml") >= 0 {
		return yaml.Unmarshal(data, v)
	}
	return json.Unmarshal(data, v)
}

// ToJSON returns the map that marshals from the body bytes as json in response .
// it calls Response inner.
func (c *HTTPClient) ToJSON(v interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToMap returns the map that marshals from the body bytes as json in response .
// it calls Response inner.
func (c *HTTPClient) ToMap(v *map[string]interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ToXML returns the map that marshals from the body bytes as xml in response .
// it calls Response inner.
func (c *HTTPClient) ToXML(v interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// ToYAML returns the map that marshals from the body bytes as yaml in response .
// it calls Response inner.
func (c *HTTPClient) ToYAML(v interface{}) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

// ToFile saves the body data in response to one file.
// it calls Response inner.
func (c *HTTPClient) ToFile(filename string) error {
	if c.response == nil {
		_, err := c.Do()
		if err != nil {
			return err
		}
	}
	if c.response.Body == nil {
		return nil
	}
	defer c.response.Body.Close()
	err := pathExistAndMkdir(filename)
	if err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, c.response.Body)
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
func (c *HTTPClient) Bytes() ([]byte, error) {
	if c.body != nil {
		return c.body, nil
	}
	var err error
	if c.response == nil {
		_, err := c.Do()
		if err != nil {
			return nil, err
		}
	}
	if c.response.Body == nil {
		return nil, errors.New("empty body")
	}
	defer c.response.Body.Close()
	if c.response.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(c.response.Body)
		if err != nil {
			return nil, err
		}
		c.body, err = ioutil.ReadAll(reader)
		return c.body, err
	}
	c.body, err = ioutil.ReadAll(c.response.Body)
	return c.body, err
}

func (c *HTTPClient) build() error {
	if c.EnableCookie && c.client.Jar == nil {
		c.client.Jar = defaultCookieJar
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
	if c.Request.Method == http.MethodGet {
		if has {
			rurl := c.Request.URL.String()
			if strings.Contains(rurl, "?") {
				rurl += "&" + urlParam
			} else {
				rurl += "?" + urlParam
			}
			urls, err := url.Parse(rurl)
			if err != nil {
				return err
			}
			c.Request.URL = urls
		}
		return nil
	}
	if (c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch ||
		c.Request.Method == http.MethodDelete) && c.Request.Body == nil {
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
			c.Request.Body = ioutil.NopCloser(bodyBuffer)
		} else if has {
			c.Header("Content-Type", "application/x-www-form-urlencoded")
			c.Body(urlParam)
		}
	}
	return nil
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
func Get(url string) *HTTPClient {
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
func Post(url string) *HTTPClient {
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
func Head(url string) *HTTPClient {
	return createRequest(url, http.MethodHead)
}

// Put returns *HTTPClient with PUT method
func Put(url string) *HTTPClient {
	return createRequest(url, http.MethodPut)
}

// Delete returns *HTTPClient with Delete method
func Delete(url string) *HTTPClient {
	return createRequest(url, http.MethodDelete)
}

// Patch returns *HTTPClient with Patch method
func Patch(url string) *HTTPClient {
	return createRequest(url, http.MethodPatch)
}

func createRequest(uri, method string) *HTTPClient {
	u, _ := url.Parse(uri)
	c := &HTTPClient{
		client: http.Client{
			Transport: &http.Transport{
				Dial:                defaultTimeout,
				MaxIdleConnsPerHost: 100,
			},
		},
		Request: &http.Request{
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
func (c *HTTPClient) Do() (resp *http.Response, err error) {
	c.response = nil
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
	for i := 0; c.Retry == -1 || i <= c.Retry; i++ {
		resp, err = c.client.Do(c.Request)
		if err == nil {
			break
		}
	}
	c.response = resp
	callAfter(c)
	return
}

func callBefore(c *HTTPClient) error {
	if before != nil {
		return before(c)
	}
	return nil
}

func callAfter(c *HTTPClient) {
	if after != nil {
		after(c)
	}
}

// Timeout returns functions of connection dialer with timeout settings for http.Transport Dial field
func Timeout(connTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, connTimeout)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, err
	}
}

//Before before handler for each network request
func Before(f func(*HTTPClient) error) {
	before = f
}

//After after handler for each network request
func After(f func(*HTTPClient)) {
	after = f
}

//init
func init() {
	defaultCookieJar, _ = cookiejar.New(nil)
}