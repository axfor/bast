//Copyright 2018 The axx Authors. All rights reserved.

package redis

import (
	"github.com/axfor/bast/session/conf"
	"github.com/axfor/bast/session/redis/cluster"
	"github.com/axfor/bast/session/redis/standalone"
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
