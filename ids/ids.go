//https://github.com/bwmarrin/snowflake
// Package snowflake provides a very simple Twitter snowflake generator and parser.

package ids

import "github.com/aixiaoxiang/bast/snowflake"

//ID 快捷ID生成
func ID() int64 {
	// n, _ := snowflake.NewNode(0)
	// return n.Generate().Int64()
	return IDX()
}

//IDX 快捷ID生成
func IDX() int64 {
	n, _ := snowflake.NewNodeX(0)
	return n.GenerateX().Int64()
}

//IDWithNode 快捷ID生成
func IDWithNode(node int64) int64 {
	n, _ := snowflake.NewNode(node)
	return n.Generate().Int64()
}

//IDStr 快捷ID生成
func IDStr(node ...int64) string {
	if node != nil {
		n, _ := snowflake.NewNode(node[0])
		return n.Generate().String()
	}
	n, _ := snowflake.NewNode(0)
	return n.Generate().String()
}

//IDClear 清空
func IDClear() {
	snowflake.Clear()
}
