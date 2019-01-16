# bast

## A lightweight RESTful  for Golang


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
