package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aixiaoxiang/bast/conf"
	"github.com/aixiaoxiang/bast/logs"
	"go.etcd.io/etcd/clientv3"
)

//Registry inwrap for clientv3 of etcd
type Registry struct {
	Client       *clientv3.Client
	ID           clientv3.LeaseID
	stop         chan struct{}
	keepAliveing bool
}

//NewRegistry create an clientv3 of etcd
func NewRegistry(c *conf.Service) (*Registry, error) {
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
		return nil, err
	}

	resp, err := cli.Grant(context.TODO(), c.TTL)
	if err != nil {
		return nil, err
	}
	s := &Registry{
		Client: cli,
		ID:     resp.ID,
	}
	return s, nil
}

//Put inwrap for clientv3.Put of etcd
func (s *Registry) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	if s.Client == nil {
		return nil, errors.New("connection error")
	}
	return s.Client.Put(ctx, key, val, s.op(opts)...)
}

//Get inwrap for clientv3.Get of etcd
func (s *Registry) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if s.Client == nil {
		return nil, errors.New("connection error")
	}
	return s.Client.Get(ctx, key, s.op(opts)...)
}

//Delete inwrap for clientv3.Delete of etcd
func (s *Registry) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	if s.Client == nil {
		return nil, errors.New("connection error")
	}
	return s.Client.Delete(ctx, key, s.op(opts)...)
}

//Compact inwrap for clientv3.Compact of etcd
func (s *Registry) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	if s.Client == nil {
		return nil, errors.New("connection error")
	}
	return s.Client.Compact(ctx, rev, opts...)
}

//Do inwrap for clientv3.Do of etcd
func (s *Registry) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	if s.Client == nil {
		return clientv3.OpResponse{}, errors.New("connection error")
	}
	return s.Client.Do(ctx, op)
}

//Txn inwrap for clientv3.Txn of etcd
func (s *Registry) Txn(ctx context.Context) clientv3.Txn {
	if s.Client == nil {
		return nil
	}
	return s.Client.Txn(ctx)
}

//Watch inwrap for clientv3.Watch of etcd
func (s *Registry) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	if s.Client == nil {
		return nil
	}
	return s.Client.Watch(ctx, key, s.op(opts)...)
}

//RequestProgress inwrap for clientv3.RequestProgress of etcd
func (s *Registry) RequestProgress(ctx context.Context) error {
	return s.Client.RequestProgress(ctx)
}

//Close inwrap for clientv3.Close of etcd
func (s *Registry) Close() error {
	return s.Client.Close()
}

//KeepAlive start keepAlive
func (s *Registry) KeepAlive() error {
	if !s.keepAliveing {
		return nil
	}
	if s.Client == nil {
		return errors.New("connection error")
	}
	ch, err := s.keepAlive()
	if err != nil {
		logs.Errors("registry-start", err)
		return err
	}
	s.keepAliveing = true
	for {
		select {
		case <-s.stop:
			s.revoke()
			s.keepAliveing = false
			return nil
		case <-s.Client.Ctx().Done():
			s.keepAliveing = false
			logs.Error("registry-start", logs.String("server", "server closed"))
			return errors.New("server closed")
		case _, ok := <-ch:
			if !ok {
				s.keepAliveing = false
				logs.Error("registry-start", logs.String("keepAlive", "keep alive channel closed"))
				s.revoke()
				return nil
			}
			//  else {
			// 	// log.Printf("Recv reply from Registry: %s, ttl:%d", s.Name, ka.TTL)
			// }
		}
	}
}

//Stop stop keepAlive
func (s *Registry) Stop() {
	if s.keepAliveing {
		if s.Client == nil {
			return
		}
		s.stop <- struct{}{}
		s.keepAliveing = false
	}
}

func (s *Registry) op(opts []clientv3.OpOption) []clientv3.OpOption {
	if opts != nil {
		opts = append(opts, clientv3.WithLease(s.ID))
	} else {
		opts = []clientv3.OpOption{clientv3.WithLease(s.ID)}
	}
	return opts
}

func (s *Registry) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {

	return s.Client.KeepAlive(context.TODO(), s.ID)
}

func (s *Registry) revoke() error {
	_, err := s.Client.Revoke(context.TODO(), s.ID)
	return err
}
