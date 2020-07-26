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
	lock     sync.RWMutex
	prefix   string
	nodes    map[string]map[string]string
	client   *clientv3.Client
	watching bool
}

//NewDiscovery create an clientv3 of etcd
func NewDiscovery(c *conf.DiscoveryConf) (*Discovery, error) {
	if c.DialTimeout <= 0 {
		c.DialTimeout = 10
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(c.Endpoints, ","),
		DialTimeout: time.Duration(c.DialTimeout) * time.Second,
	})

	if err != nil {
		logs.Errors("new disconver failed", err)
		return nil, err
	}

	d := &Discovery{
		prefix: c.Prefix,
		nodes:  map[string]map[string]string{},
		client: cli,
	}
	go d.Watch()
	return d, nil
}

//Name get service name
func (d *Discovery) Name(key string) string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	key = d.prefix + key
	nodes, ok := d.nodes[key]
	if ok {
		for _, item := range nodes {
			return item
		}
	}
	return ""
}

//Names get service  all name
func (d *Discovery) Names(key string) []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	key = d.prefix + key
	nodes, ok := d.nodes[key]
	if ok {
		name := make([]string, 0, len(nodes))
		for _, item := range nodes {
			name = append(name, item)
		}
		return name
	}
	return nil
}

//Sync sync all service nodes
func (d *Discovery) Sync() error {
	if d.prefix == "" {
		return nil
	}
	keys, err := d.client.Get(context.TODO(), d.prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, key := range keys.Kvs {
		k := string(key.Key)
		v := string(key.Value)
		if k != "" && v != "" {
			d.doUpdateNode(k, v)
		}
	}
	return nil
}

//Watch watch change of  all service nodes
func (d *Discovery) Watch() {
	if d.watching || d.prefix == "" {
		return
	}
	err := d.Sync()
	if err != nil {
		logs.Errors("disconver sync error", err)
	}
	d.watching = true
	rch := d.client.Watch(context.TODO(), d.prefix, clientv3.WithPrefix())
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
				logs.Debug("disconver delete", logs.Any("type", ev.Type), logs.String("key", key))
				d.deleteNode(key)
			}
		}
	}
}

func (d *Discovery) updateNode(key, value string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.doUpdateNode(key, value)
}

func (d *Discovery) doUpdateNode(key, value string) {
	realKey := key
	pos := strings.LastIndex(key, ".")
	if pos != -1 {
		realKey = key[0:pos]
	}
	nodes, ok := d.nodes[realKey]
	if !ok {
		nodes = map[string]string{}
	}
	nodes[key] = value
	d.nodes[realKey] = nodes
}

func (d *Discovery) deleteNode(key string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	realKey := key
	pos := strings.LastIndex(key, ".")
	if pos != -1 {
		realKey = key[0:pos]
	}
	nodes, ok := d.nodes[realKey]
	if ok {
		delete(nodes, key)
		d.nodes[realKey] = nodes
	}
	if len(nodes) <= 0 {
		delete(d.nodes, realKey)
	}
}

//Stop shuts down the client's etcd connections.
func (d *Discovery) Stop() {
	d.watching = false
	if d.client == nil {
		return
	}
	d.client.Close()
}
