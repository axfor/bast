//Copyright 2018 The axx Authors. All rights reserved.
// see https://github.com/bwmarrin/snowflake
// Package snowflake provides a very simple Twitter snowflake generator and parser.

package ids

import "github.com/aixiaoxiang/bast/snowflake"

var currentIDNode uint8

//ID cenerate id
func ID() int64 {
	n, _ := snowflake.NewNode(currentIDNode)
	return n.GenerateWithInt64()
}

//SetCurrentIDNode set  current id node
func SetCurrentIDNode(idNode uint8) {
	currentIDNode = idNode
}

//IDWithNode  cenerate id
func IDWithNode(node uint8) int64 {
	n, _ := snowflake.NewNode(node)
	return n.GenerateWithInt64()
}

//IDStr  cenerate id
func IDStr(node ...uint8) string {
	if node != nil {
		n, _ := snowflake.NewNode(node[0])
		return n.Generate().String()
	}
	n, _ := snowflake.NewNode(currentIDNode)
	return n.Generate().String()
}

//Clear clear all
func Clear() {
	snowflake.Clear()
}
