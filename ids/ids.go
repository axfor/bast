//Copyright 2018 The axx Authors. All rights reserved.
// see https://github.com/bwmarrin/snowflake
// Package snowflake provides a very simple Twitter snowflake generator and parser.

package ids

import "github.com/axfor/bast/snowflake"

var currentIDNode uint8

//New new id node
func New() *snowflake.Node {
	id, err := snowflake.NewNode(currentIDNode)
	if err != nil {
		return nil
	}
	return id
}

//NewWithIDNode new id node
func NewWithIDNode(idNode uint8) *snowflake.Node {
	id, err := snowflake.NewNode(idNode)
	if err != nil {
		return nil
	}
	return id
}

//SetIDNode set  current id node
func SetIDNode(idNode uint8) {
	currentIDNode = idNode
}
