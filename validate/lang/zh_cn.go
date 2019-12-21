package lang

//zh-cn
var zhCn = map[string]string{
	"required":   "{0} 不能空",
	"date":       "{0} 必须为有效日期类型",
	"int":        "{0} 必须为数字类型",
	"max.string": "{0} 必须小于 {1} 个字符",
	"max.int":    "{0} 必须小于等于 {1}",
	"min.string": "{0} 必须大于 {1} 个字符",
	"min.int":    "{0} 必须大于等于 {1}",
}

func init() {
	Register("zh-cn", zhCn)
}
