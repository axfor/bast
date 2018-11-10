//Copyright 2018 The axx Authors. All rights reserved.

package guid

import "github.com/aixiaoxiang/bast/objectid"

//GUID 快捷创建一个GUID
func GUID() string {
	return objectid.New().Hex()
}
