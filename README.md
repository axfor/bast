# bast

## A lightweight RESTful  for Golang


> Install

``` bash

 go get -u github.com/aixiaoxiang/bast

 ```

# Router doc(Request example)

> Router
 

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
     //handling
     //...
     ctx.JSON(person)
})

```
 

### Post

``` golang

//Person struct 
type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"` 
} 

bast.Post("/xxx", func(ctx *bast.Context){
     person := &Person{}  
     err := ctx.JSONObj(person)
     if err != nil {
          ctx.Failed("Sorry! invalid parameter")
	   return
     }
     person.Age += 2
     //handling
     //...
     ctx.JSON(person)
}) 

```

### Run 

``` golang

bast.Run(":9999")

```
  

# CommandLine

### If Your program name is ``` aibast ```

#### -h | -help

``` bash

    ./aibast -h

```

#### -start   

``` bash

    ./aibast -start

```

#### -stop

``` bash

    ./aibast -stop

```

#### -reload    

> ``` graceful restart  ```

``` bash

    ./aibast -reload

```

#### -conf 

``` bash

    ./aibast -conf=your path/config.conf 

```

### Such as

> ``` deploy program ```


``` bash  

    ./aibast -start -conf=./config.conf 

```

# Distributed system unique ID    

> [snowflake-golang](https://github.com/bwmarrin/snowflake)  or [snowflake-twitter](https://github.com/twitter/snowflake)   
base snowflake-golang, 4x+ performance optimization。 5000000 -> 20000000

> use

``` golang

  id := bast.ID()
  fmt.Printf("id=%d", id)

```

> benchmark test ``` go test   -bench=. -cpu=12 -benchmem ```   
physics cpu ``` 4 ```

``` bash

    go test   -bench=. -cpu=12 -benchmem 
    goos: darwin
    goarch: amd64
    Benchmark_ID-12         20000000                93.2 ns/op            32 B/op          1 allocs/op
    PASS
    ok      _/xxx/bast/ids 1.970s

```
