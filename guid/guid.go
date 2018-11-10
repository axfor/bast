package guid

import "github.com/aixiaoxiang/bast/objectid"

//GUID 快捷创建一个GUID
func GUID() string {
	return objectid.New().Hex()
}
