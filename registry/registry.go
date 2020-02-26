package registry

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aixiaoxiang/bast/conf"
	"github.com/aixiaoxiang/bast/logs"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

//Service inwrap for clientv3 of etcd
type Service struct {
	Client       *clientv3.Client
	ID           clientv3.LeaseID
	stop         chan struct{}
	keepAliveing bool
}

//New create an clientv3 of etcd
func New(rc *conf.Registry) (*Service, error) {
	if rc.DialTimeout <= 0 {
		rc.DialTimeout = 5
	}
	if rc.TTL <= 0 {
		rc.TTL = 5
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(rc.Endpoints, ","),
		DialTimeout: time.Duration(rc.DialTimeout) * time.Second,
	})

	if err != nil {
		return nil, err
	}

	resp, err := cli.Grant(context.TODO(), rc.TTL)
	if err != nil {
		return nil, err
	}
	s := &Service{
		Client: cli,
		ID:     resp.ID,
	}
	return s, nil
}

//Put inwrap for clientv3.Put of etcd
func (s *Service) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	if s.Client == nil {
		return nil, errors.New("connection error")
	}
	return s.Client.Put(ctx, key, val, s.op(opts)...)
}

//Get inwrap for clientv3.Get of etcd
func (s *Service) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if s.Client == nil {
		return nil, errors.New("connection error")
	}
	return s.Client.Get(ctx, key, s.op(opts)...)
}

//Delete inwrap for clientv3.Delete of etcd
func (s *Service) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	if s.Client == nil {
		return nil, errors.New("connection error")
	}
	return s.Client.Delete(ctx, key, s.op(opts)...)
}

//Compact inwrap for clientv3.Compact of etcd
func (s *Service) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	if s.Client == nil {
		return nil, errors.New("connection error")
	}
	return s.Client.Compact(ctx, rev, opts...)
}

//Do inwrap for clientv3.Do of etcd
func (s *Service) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	if s.Client == nil {
		return clientv3.OpResponse{}, errors.New("connection error")
	}
	return s.Client.Do(ctx, op)
}

//Txn inwrap for clientv3.Txn of etcd
func (s *Service) Txn(ctx context.Context) clientv3.Txn {
	if s.Client == nil {
		return nil
	}
	return s.Client.Txn(ctx)
}

//Watch inwrap for clientv3.Watch of etcd
func (s *Service) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	if s.Client == nil {
		return nil
	}
	return s.Client.Watch(ctx, key, s.op(opts)...)
}

func (s *Service) op(opts []clientv3.OpOption) []clientv3.OpOption {
	if opts != nil {
		opts = append(opts, clientv3.WithLease(s.ID))
	} else {
		opts = []clientv3.OpOption{clientv3.WithLease(s.ID)}
	}
	return opts
}

//RequestProgress inwrap for clientv3.RequestProgress of etcd
func (s *Service) RequestProgress(ctx context.Context) error {
	return s.Client.RequestProgress(ctx)
}

//Close inwrap for clientv3.Close of etcd
func (s *Service) Close() error {
	return s.Client.Close()
}

//StartKeepAlive start keepAlive
func (s *Service) StartKeepAlive() error {
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
			logs.Error("registry-start", zap.String("server", "server closed"))
			return errors.New("server closed")
		case _, ok := <-ch:
			if !ok {
				s.keepAliveing = false
				logs.Error("registry-start", zap.String("keepAlive", "keep alive channel closed"))
				s.revoke()
				return nil
			}
			//  else {
			// 	// log.Printf("Recv reply from service: %s, ttl:%d", s.Name, ka.TTL)
			// }
		}
	}
}

//StopKeepAlive stop keepAlive
func (s *Service) StopKeepAlive() {
	if s.keepAliveing {
		if s.Client == nil {
			return
		}
		s.stop <- struct{}{}
		s.keepAliveing = false
	}
}

func (s *Service) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	return s.Client.KeepAlive(context.TODO(), s.ID)
}

func (s *Service) revoke() error {
	_, err := s.Client.Revoke(context.TODO(), s.ID)
	return err
}
