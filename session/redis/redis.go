//Copyright 2018 The axx Authors. All rights reserved.

package redis

import (
	"github.com/aixiaoxiang/bast/session/conf"
	"github.com/aixiaoxiang/bast/session/redis/cluster"
	"github.com/aixiaoxiang/bast/session/redis/standalone"
)

//Init init
func Init(c *conf.Conf) error {
	if c.Engine == "redis" {
		err := standalone.Init(c)
		return err
	} else if c.Engine == "redis-cluster" {
		err := cluster.Init(c)
		return err
	}
	return nil
}
