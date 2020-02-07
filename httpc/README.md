# Http Client 

` httpc is HTTP client library `

## install 

``` bash

 go get -u github.com/aixiaoxiang/bast/httpc

``` 

## Support for GET POST HEAD POST PUT PATCH DELETE etc.

`result support string,json,xml,yaml,file etc.`

### string result

``` golang

	result, err := httpc.Get("https://suggest.taobao.com/sug?code=utf-8&q=phone").String()
	if err != nil {
		//handling
	}

``` 

### json result

``` golang 

	type tb struct {
		Result [][]string `json:"result"`
	}

	rv := &tb{}

	err := httpc.Get("https://suggest.taobao.com/sug?code=utf-8").Param("q", "phone").ToJSON(rv)
	if err != nil {
		//handling
	}

```  
 
### xml result

``` golang 

	type tb struct {
		Result [][]string `json:"result"`
	}

	rv := &tb{}

	err := httpc.Get("https://suggest.taobao.com/sug?code=utf-8").Param("q", "phone").ToXML(rv)
	if err != nil {
		//handling
	}

```  


### yaml result

``` golang 

	type tb struct {
		Result [][]string `json:"result"`
	}

	rv := &tb{}

	err := httpc.Get("https://suggest.taobao.com/sug?code=utf-8").Param("q", "phone").ToYAML(rv)
	if err != nil {
		//handling
	}

```  
 
### save result to file

``` golang

	err := httpc.Post("https://suggest.taobao.com/sug?code=utf-8&q=phone").ToFile("./files/f.json")
	if err != nil {
		//handling
	}

``` 

### upload file to server

``` golang

	result, err := httpc.Post("https://suggest.taobao.com/sug?code=utf-8&q=phone").File("testFile", "./files/f.json").String()
	if err != nil {
		//handling
	}

```


### logging and seting title

``` golang 

	type tb struct {
		Result [][]string `json:"result"`
	}

	rv := &tb{}

	err := httpc.Get("https://suggest.taobao.com/sug?code=utf-8").Title("taobao").Logging().Param("q", "phone").ToJSON(rv)
	if err != nil {
		//handling
	}

```  
 
### mark tag and hook's
 

``` golang

	result, err := httpc.Post("https://suggest.taobao.com/sug?code=utf-8&q=phone").MarkTag("ai").String()
	if err != nil {
		//handling
	}

```   

> Global register hook Before and After

``` golang 

func init() {
	httpc.Before(func(c *httpc.Client) error {
		if c.Tag == "ai" {
			c.Header("xxxx-test-header", "httpc")
		} else {
			//others handling
		}
		return nil
	})

	httpc.After(func(c *httpc.Client) {
		if c.Tag == "ai" && c.OK() {
			//log..
		} else {
			//others handling
		}
	})
} 

```
