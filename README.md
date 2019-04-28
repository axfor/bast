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

### If Your program name is ``` Aibast ```

#### -h | -help

``` bash

    ./Aibast -h

```

#### -start   

``` bash

    ./Aibast -start

```

#### -stop

``` bash

    ./Aibast -stop

```

#### -reload    

> ``` graceful restart  ```

``` bash

    ./Aibast -reload

```

#### -conf 

``` bash

    ./Aibast -conf=your path/config.conf 

```


#### -install 

> ``` daemon ```


``` bash

    ./Aibast -install

```


#### -uninstall 

``` bash

    ./Aibast -uninstall

```
 

#### -migration 
 
` migration or initial system `

``` bash

    ./Aibast -migration

```



### Such as

> ``` run program ```


``` bash  

    ./Aibast -start -conf=./config.conf 

```


> ``` deploy program ```


``` bash  

    ./Aibast -install

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
