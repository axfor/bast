package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aixiaoxiang/bast/conf"
	"github.com/aixiaoxiang/bast/logs"
	"github.com/coreos/etcd/clientv3"
)

//Discovery inwrap for clientv3 of etcd
type Discovery struct {
	lock   sync.RWMutex
	prefix string
	nodes  map[string]string
	client *clientv3.Client
}

//NewDiscovery create an clientv3 of etcd
func NewDiscovery(c *conf.Service) (*Discovery, error) {
	if c.DialTimeout <= 0 {
		c.DialTimeout = 5
	}
	if c.TTL <= 0 {
		c.TTL = 5
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(c.Endpoints, ","),
		DialTimeout: time.Duration(c.DialTimeout) * time.Second,
	})

	if err != nil {
		logs.Errors("new disconver failed", err)
		return nil, err
	}

	discovery := &Discovery{
		prefix: c.Prefix,
		nodes:  make(map[string]string),
		client: cli,
	}
	go discovery.Watch()
	return discovery, err
}

//Name get service name
func (d *Discovery) Name(key string) string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	n, ok := d.nodes[key]
	if !ok {
		return n
	}
	return ""
}

//Watch watch change of  all service nodes
func (d *Discovery) Watch() {
	rch := d.client.Watch(context.Background(), d.prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				key := string(ev.Kv.Key)
				value := string(ev.Kv.Value)
				if key != "" {
					logs.Debug("disconver put", logs.Any("type", ev.Type), logs.String("key", key),
						logs.String("value", value))
				}
				d.updateNode(key, value)
			case clientv3.EventTypeDelete:
				key := string(ev.Kv.Key)
				value := string(ev.Kv.Value)
				logs.Debug("disconver delete", logs.Any("type", ev.Type), logs.String("key", key),
					logs.String("value", value))
				d.deleteNode(key)
			}
		}
	}
}

func (d *Discovery) updateNode(key, value string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.nodes[key] = value
}

func (d *Discovery) deleteNode(key string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.nodes, key)
}

//Close shuts down the client's etcd connections.
func (d *Discovery) Close() {
	d.client.Close()
}
