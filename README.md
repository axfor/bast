# bast

## A lightweight RESTful  for Golang


> install

``` bash

 go get -u github.com/aixiaoxiang/bast

 ```

# API doc

> Router

---

### Get

``` golang


//Person struct 
type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"` 
}

bast.Get("/xxx", func(ctx *bast.Context){
     name := ctx.GetString("name")
     age := ctx.GetInt("age") 
     person := &Person{
        Name:name,
        Age:Age, 
     }
     ctx.JSON(person)
})

```

---

### Post

``` golang

//Person struct 
type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"` 
}

bast.Post("/xxx", func(ctx *bast.Context){
     person := &Person{}  
     ctx.JSONObj(person)
     person.Age += 2
     ctx.JSON(person)
})

```

---

# Run


``` golang

bast.Run(":9999")

```
---


# CommandLine

### If Your program name is ``` aibast ```

####  -h | -help

``` bash

    ./aibast -h

```

####  -start   

``` bash

    ./aibast -start

```

#### -stop

``` bash

    ./aibast -stop

```

#### -reload    

> ``` Graceful Restart  ```

``` bash

    ./aibast -reload

```

#### -conf 

``` bash

    ./aibast -conf=your path/config.conf 

```

### Such as

> deploy program


``` bash

    ./aibast -start -conf=./config.conf 

```

---
