package bast

import (
	"crypto/tls"
	"net/http"
)

var (

	//HTTPClient is default http client
	HTTPClient = &http.Client{}
	tr         = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	//HTTPSClient is default https client
	HTTPSClient = &http.Client{Transport: tr}
	//HTTP is default http client
	HTTP = &HTTPClientProxy{}
)

//HTTPClientProxy http client
type HTTPClientProxy struct {
	// http.Client
}

//Dos when http call HTTPClient.Do or https call HTTPSClient.Do
func (c *HTTPClientProxy) Dos(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "https" {
		return HTTPSClient.Do(req)
	}
	return HTTPClient.Do(req)
}
