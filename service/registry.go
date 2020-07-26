package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aixiaoxiang/bast/conf"
	"github.com/aixiaoxiang/bast/logs"
	"go.etcd.io/etcd/clientv3"
)

//Registry inwrap for clientv3 of etcd
type Registry struct {
	client       *clientv3.Client
	leaseID      clientv3.LeaseID
	prefix       string
	postfix      string
	stop         chan struct{}
	keepAliveing bool
}

//NewRegistry create an clientv3 of etcd
func NewRegistry(c *conf.RegistryConf) (*Registry, error) {
	if c.DialTimeout <= 0 {
		c.DialTimeout = 10
	}

	if c.TTL <= 0 {
		c.TTL = 5
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(c.Endpoints, ","),
		DialTimeout: time.Duration(c.DialTimeout) * time.Second,
	})

	if err != nil {
		return nil, err
	}

	resp, err := cli.Grant(cli.Ctx(), c.TTL)
	if err != nil {
		return nil, err
	}
	r := &Registry{
		client:  cli,
		leaseID: resp.ID,
		prefix:  c.Prefix,
		postfix: fmt.Sprintf(".%d", resp.ID),
	}
	return r, nil
}

//Put inwrap for clientv3.Put of etcd
func (r *Registry) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	if r.client == nil {
		return nil, errors.New("registry etcd client is nil")
	}
	key = r.prefix + key + r.postfix
	return r.client.Put(ctx, key, val, r.op(opts)...)
}

//Get inwrap for clientv3.Get of etcd
func (r *Registry) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if r.client == nil {
		return nil, errors.New("registry etcd client is nil")
	}
	key = r.prefix + key + r.postfix
	return r.client.Get(ctx, key, r.op(opts)...)
}

//Delete inwrap for clientv3.Delete of etcd
func (r *Registry) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	if r.client == nil {
		return nil, errors.New("registry etcd client is nil")
	}
	key = r.prefix + key + r.postfix
	return r.client.Delete(ctx, key, r.op(opts)...)
}

//Compact inwrap for clientv3.Compact of etcd
func (r *Registry) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	if r.client == nil {
		return nil, errors.New("registry etcd client is nil")
	}
	return r.client.Compact(ctx, rev, opts...)
}

//Do inwrap for clientv3.Do of etcd
func (r *Registry) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	if r.client == nil {
		return clientv3.OpResponse{}, errors.New("registry etcd client is nil")
	}
	return r.client.Do(ctx, op)
}

//Txn inwrap for clientv3.Txn of etcd
func (r *Registry) Txn(ctx context.Context) clientv3.Txn {
	if r.client == nil {
		return nil
	}
	return r.client.Txn(ctx)
}

//Watch inwrap for clientv3.Watch of etcd
func (r *Registry) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	if r.client == nil {
		return nil
	}
	return r.client.Watch(ctx, key, r.op(opts)...)
}

//RequestProgress inwrap for clientv3.RequestProgress of etcd
func (r *Registry) RequestProgress(ctx context.Context) error {
	return r.client.RequestProgress(ctx)
}

//Close inwrap for clientv3.Close of etcd
func (r *Registry) Close() error {
	return r.client.Close()
}

//Start start keepAlive and discovery
func (r *Registry) Start() error {
	err := r.KeepAlive()
	if err != nil {
		return err
	}
	return nil
}

//KeepAlive start keepAlive
func (r *Registry) KeepAlive() error {
	if r.keepAliveing {
		return nil
	}
	if r.client == nil {
		logs.Error("registry etcd client is nil")
		return errors.New("registry etcd client is nil")
	}
	ch, err := r.doKeepAlive()
	if err != nil {
		logs.Errors("registry start", err)
		return err
	}
	r.keepAliveing = true
	for {
		select {
		case <-r.stop:
			r.revoke()
			r.keepAliveing = false
			return nil
		case <-r.client.Ctx().Done():
			r.keepAliveing = false
			logs.Error("registry keepAlive done", logs.String("server", "server closed"))
			return errors.New("server closed")
		case _, ok := <-ch:
			if !ok {
				r.keepAliveing = false
				logs.Error("registry keepAlive close", logs.String("keepAlive", "keep alive channel closed"))
				err := r.revoke()
				return err
			}
			// else {
			// 	// log.Printf("recv reply from registry: %s, ttl:%d", sv.String(), sv.TTL)
			// }
		}
	}
}

//Stop stop keepAlive
func (r *Registry) Stop() {
	if r.keepAliveing {
		if r.client == nil {
			return
		}
		r.stop <- struct{}{}
		r.keepAliveing = false
	}
}

func (r *Registry) op(opts []clientv3.OpOption) []clientv3.OpOption {
	if opts != nil {
		opts = append(opts, clientv3.WithLease(r.leaseID))
	} else {
		opts = []clientv3.OpOption{clientv3.WithLease(r.leaseID)}
	}
	return opts
}

func (r *Registry) doKeepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	return r.client.KeepAlive(context.TODO(), r.leaseID)
}

func (r *Registry) revoke() error {
	_, err := r.client.Revoke(context.TODO(), r.leaseID)
	return err
}
