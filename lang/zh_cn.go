//Copyright 2018 The axx Authors. All rights reserved.

package lang

//zh-cn
var zhCn = map[string]string{
	"v.required":   "{0}不能空",
	"v.date":       "{0}必须为有效日期类型",
	"v.int":        "{0}必须为数字类型",
	"v.max.string": "{0}必须小于{1}个字符",
	"v.max.int":    "{0}必须小于等于{1}",
	"v.min.string": "{0}必须大于{1}个字符",
	"v.min.int":    "{0}必须大于等于{1}",
	"v.email":      "{0}无效的邮件格式",
	"v.ip":         "{0}无效的IP地址",
	"v.match":      "{0}无效的数据格式",
}

func init() {
	Register("zh-cn", zhCn)
}
