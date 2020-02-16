
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

 etc