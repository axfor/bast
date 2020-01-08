# bast [![Build Status](https://travis-ci.org/aixiaoxiang/bast.svg?branch=master)](https://travis-ci.org/aixiaoxiang/bast)

# A lightweight RESTful  for Golang



> Install

``` bash

 go get -u github.com/aixiaoxiang/bast

 ```

# Router doc(Request example)

> Router
 

## Get

``` golang
//Person struct  
type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Addr string `json:"addr"`
}
 
bast.Get("/xxx", func(ctx *bast.Context) {
    //verify imput parameter
    err := ctx.Verify("name@required|min:1", "age@required|min:1", "addr/address@required|min:1")
    if err != nil {
        ctx.Failed(err.Error())
        return
    }

    name := ctx.GetString("name")
    age := ctx.GetInt("age")

    //handling
    //...

    person := &Person{
        Name: name,
        Age:  Age,
    }
    //handling
    //...
    ctx.JSON(person)
})  


``` 

## Post

``` golang 
//Person struct
type Person struct {
    Name string `json:"name" v:"required|min:1"`
    Age  int    `json:"age"  v:"min:1"`
}

bast.Post("/xxx", func(ctx *bast.Context) {
    person := &Person{}
    err := ctx.JSONObj(person) 
    //or ctx.JSONObj(person,true) //version of verify imput parameter
    if err != nil {
        ctx.Failed("sorry! invalid parameter")
        return
    }
    person.Age += 2

    //handling
    //...

    ctx.JSON(person)
})
    
``` 
---

# Validate

`a similar pipeline validator`

## Syntax

 ``` bash

  key[/key translator][/split divide (default is |)]@verify1[:verify1 param]|verify2

 ```

## Syntax such as 

>
    key1@required|int|min:1     
    key2/key2_translator@required|string|min:1|max:12      
    key3@sometimes|required|date      

## Global register keys translator

` note：only is key `

``` golang 

//global register keys translator //note：only is key
func init() { 
    //register keys translator //note：is key
    //suck as:verify error
    //en：    The name field is required
    //zh-cn： {0} 不能空 

    lang.RegisterKeys("zh-cn", map[string]string{
	    "name":    "姓名",
	    "age":     "年龄",
	    "address": "地址",
    })

    //other langs 
}

```

## Support for the validator 

` note：will continue to add new validator `

 - date
 - email
 - int
 - ip
 - match
 - max
 - min
 - required
 - sometimes 

---

# Http Client 

` httpc is HTTP client library `

## install 

``` bash

 go get -u github.com/aixiaoxiang/bast/httpc

``` 

## Support for GET POST HEAD POST PUT PATCH DELETE etc

`result support string,json,xml,yaml,file etc`

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

---
# Run 

``` golang

bast.Run(":9999")

```
  

# CommandLine

` Like nginx commandline `

### If Your program name is ``` Ai ```

#### -h | -help

``` bash

    ./Ai -h

```

#### -start   

` Run in background  `

``` bash

    ./Ai -start

```

#### -stop 

` stop program `

``` bash

    ./Ai -stop

```

#### -reload    

`graceful restart. stop and start`

``` bash

    ./Ai -reload

```

#### -conf 

` seting config files.(default is ./config.conf)`

``` bash

    ./Ai -conf=your path/config.conf 

```


#### -install 

`installed as service.(daemon) `


``` bash

    ./Ai -install

```


#### -uninstall 

`uninstall a service.(daemon) `


``` bash

    ./Ai -uninstall

```
 

#### -migration 
 
` migration or initial system(handle sql script ...) `

``` bash

    ./Ai -migration

```
 
### Such as

>` run program (run in background) `


``` bash  

    ./Ai -start -conf=./config.conf 

```


> ` deploy program (startup) `


``` bash  

    ./Ai -install

```

# config template

` support multiple instances` 
 

``` json
[
    {//a instance
        "key":"xxx-conf",
        "name":"xx",  
        "addr":":9999",
        "fileDir":"./file/",//(default is ./file/)
        "debug":false,
        "baseUrl":"", 
        "idNode":0,  
        "lang":"en",//en,zh-cn
        "sameSite":"none",//cookie sameSite strict、lax、none 
        "wrap":true,//wrap response body 
        "session":{//session conf
            "enable":false,
            "lifeTime":20,
            "name":"_sid",//session name
            "prefix":"",//session id prefix 
            "suffix":"",//session id suffix 
            "engine":"memory",//session engine memory、redis、redis-cluster 
            "source":"cookie",//cookie、header
            "redis":{//if source eq redis or redis-cluster
               "addrs":"ip:port,ip2:port2",
               "password":"",
               "poolSize":0
            }
        },
        "log":{
            "outPath":"./logs/logs.log", //(default is ./logs/logs.log)
            "level":"debug",
            "maxSize":10,
            "maxBackups":3,
            "maxAge":28,
            "debug":false,
            "logSelect":false
        },
        "cors":{//CORS https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Access_control_CORS
            "allowOrigin":"",
            "allowMethods":"GET, POST, OPTIONS, PATCH, PUT, DELETE, HEAD,UPDATE",
            "allowHeaders":"",
            "maxAge":"1728000",
            "allowCredentials":"true"
        },
        "conf":{//user config(non bast framework)
            "key":"app",
            "name":"xxx",
            "dbTitle":"xxx app",
            "dbName":"xxxx db",
            "dbUser":"xxx",
            "dbPwd":"******",
            "dbServer":"localhost"
            //..more field..//
        },
        "extend":""//user extend
    }
    //..more instances..//
]

```

# Distributed system unique ID    

> [snowflake-golang](https://github.com/bwmarrin/snowflake)  or [snowflake-twitter](https://github.com/twitter/snowflake)   
 

> use  

``` golang

  id := bast.ID()
  fmt.Printf("id=%d", id)

```

> benchmark test ``` go test  -bench=. -benchmem  ./ids```   
physics cpu ``` 4 ```

``` bash

    go test   -bench=. -benchmem  ./ids
    goos: darwin
    goarch: amd64 
    Benchmark_ID-4              20000000    72.1 ns/op       16 B/op     1 allocs/op
    Benchmark_Parallel_ID-4     10000000    150 ns/op        16 B/op     1 allocs/op
    PASS
    ok      github.com/aixiaoxiang/bast/ids 10.126s

```
