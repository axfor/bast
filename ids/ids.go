//Copyright 2018 The axx Authors. All rights reserved.
// see https://github.com/bwmarrin/snowflake
// Package snowflake provides a very simple Twitter snowflake generator and parser.

package ids

import "github.com/aixiaoxiang/bast/snowflake"

var currentIDNode int64

//ID cenerate id
func ID() int64 {
	n, _ := snowflake.NewNode(currentIDNode)
	return n.Generate().Int64()
}

//SetCurrentIDNode set  current id node
func SetCurrentIDNode(idNode int) {
	currentIDNode = int64(idNode)
}

//IDWithNode  cenerate id
func IDWithNode(node int64) int64 {
	n, _ := snowflake.NewNode(node)
	return n.Generate().Int64()
}

//IDStr  cenerate id
func IDStr(node ...int64) string {
	if node != nil {
		n, _ := snowflake.NewNode(node[0])
		return n.Generate().String()
	}
	n, _ := snowflake.NewNode(0)
	return n.Generate().String()
}

//IDClear clear
func IDClear() {
	snowflake.Clear()
}
