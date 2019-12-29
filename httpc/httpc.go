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
	case uint, uint8, uint16, uint32, uint64:
	case int, int8, int16, int32, int64:
	case float32, float64:
	case string:
		c.Request.Body = ioutil.NopCloser(bytes.NewBufferString(t))
		c.Request.ContentLength = int64(len(t))
	case []byte:
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(t))
		c.Request.ContentLength = int64(len(t))
	default:
		return c
	}
	v := fmt.Sprint(data)
	c.Request.Body = ioutil.NopCloser(bytes.NewBufferString(v))
	c.Request.ContentLength = int64(len(v))
	return c
}

// JSON adds request raw body encoding by JSON.
func (c *HTTPClient) JSON(obj interface{}) (*HTTPClient, error) {
	if c.Request.Body == nil && obj != nil {
		data, err := json.Marshal(obj)
		if err != nil {
			return c, err
		}
		c.BodyWithContentType(data, "application/json")
	}
	return c, nil
}

// XML adds request raw body encoding by XML.
func (c *HTTPClient) XML(obj interface{}) (*HTTPClient, error) {
	if c.Request.Body == nil && obj != nil {
		data, err := xml.Marshal(obj)
		if err != nil {
			return c, err
		}
		c.BodyWithContentType(data, "application/xml")
	}
	return c, nil
}

// YAML adds request raw body encoding by YAML.
func (c *HTTPClient) YAML(obj interface{}) (*HTTPClient, error) {
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
	if c.Request.Body == nil && data != nil {
		c.Request.Body = ioutil.NopCloser(bytes.NewReader(data))
		c.Request.ContentLength = int64(len(data))
		c.Request.Header.Set("Content-Type", contentType)
	}
	return c
}

// Do add files into request.
func (c *HTTPClient) Do() (resp *http.Response, err error) {
	c.response = nil
	err = c.build()
	if err != nil {
		return
	}
	// retries default value is 0, it will run once.
	// retries equal to -1, it will run forever until success
	// retries is setted, it will retries fixed times.
	for i := 0; c.Retry == -1 || i <= c.Retry; i++ {
		resp, err = c.client.Do(c.Request)
		if err == nil {
			break
		}
	}
	c.response = resp
	return
}

func (c *HTTPClient) String() (string, error) {
	data, err := c.Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
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
	pathExist(filename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, c.response.Body)
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
	if c.EnableCookie {
		c.client.Jar = defaultCookieJar
	}
	if c.Request.Method == http.MethodGet && len(c.params) > 0 {
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
		}
	}
	return nil
}

//Get proxy method http.Get
func Get(url string) *HTTPClient {
	return createRequest(url, http.MethodGet)
}

//Post proxy http.Get
func Post(url string) *HTTPClient {
	return createRequest(url, http.MethodPost)
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
	}
	return c
}

// Timeout returns functions of connection dialer with timeout settings for http.Transport Dial field.
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

func pathExist(pathName string) bool {
	pathName = path.Dir(pathName)
	_, err := os.Stat(pathName)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(pathName, os.ModePerm)
		if err == nil {
			return true
		}
	}
	return false
}

func init() {
	defaultCookieJar, _ = cookiejar.New(nil)
}
