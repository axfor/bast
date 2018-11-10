# bast

## golang 服务轻量 JSON API 框架


> 项目依赖

``` bash

 go get -u github.com/julienschmidt/httprouter

 // go get -u github.com/bwmarrin/snowflake

 // go get -u github.com/satori/go.uuid

 go get -u go.uber.org/zap

 go get -u gopkg.in/natefinch/lumberjack.v2

 ```


> 使用说明

``` bash

 go get -u github.com/aixiaoxiang/bast

 ```

# api 文档

> 路由

---

### Get

``` golang

bast.Get("/xxx", func(ctx *bast.Context){
  
})

```

---

### Post

``` golang

bast.Post("/xxx", func(ctx *bast.Context){

})

```

---

> 启动服务


``` golang

bast.Run(":9999")

```
---