// Package etcd provides Roma zone → gRPC target registration for Janus discovery.
package etcd

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const keyPrefix = "/qjp/roma/v1/zones"

// RomaRegistry registers and resolves Roma gRPC endpoints under etcd.
type RomaRegistry struct {
	cli    *clientv3.Client
	lease  clientv3.LeaseID
	cancel context.CancelFunc
	mu     sync.Mutex
}

// NewRomaRegistry connects to etcd. endpoints is a comma-separated client URL list.
func NewRomaRegistry(endpoints string) (*RomaRegistry, error) {
	endpoints = strings.TrimSpace(endpoints)
	if endpoints == "" {
		return nil, fmt.Errorf("etcd: empty endpoints")
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(endpoints, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &RomaRegistry{cli: cli}, nil
}

// Close releases the registration lease and closes the client.
func (r *RomaRegistry) Close() error {
	if r == nil || r.cli == nil {
		return nil
	}
	if r.cancel != nil {
		r.cancel()
	}
	if r.lease != 0 {
		_, _ = r.cli.Revoke(context.Background(), r.lease)
	}
	return r.cli.Close()
}

// zoneKey builds the etcd key for a zone shard.
func zoneKey(zoneID string, shard uint32) string {
	zoneID = strings.TrimSpace(zoneID)
	if zoneID == "" {
		zoneID = "default"
	}
	return path.Join(keyPrefix, zoneID, "shards", fmt.Sprintf("%d", shard))
}

// Register advertises grpcTarget with a TTL lease and keepalive.
func (r *RomaRegistry) Register(ctx context.Context, zoneID string, shard uint32, grpcTarget string) error {
	if r == nil || r.cli == nil {
		return fmt.Errorf("etcd: registry not connected")
	}
	grpcTarget = strings.TrimSpace(grpcTarget)
	if grpcTarget == "" {
		return fmt.Errorf("etcd: empty grpc target")
	}
	lease, err := r.cli.Grant(ctx, 10)
	if err != nil {
		return err
	}
	key := zoneKey(zoneID, shard)
	_, err = r.cli.Put(ctx, key, grpcTarget, clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.lease = lease.ID
	if r.cancel != nil {
		r.cancel()
	}
	keepCtx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	r.mu.Unlock()

	ch, kaErr := r.cli.KeepAlive(keepCtx, lease.ID)
	if kaErr != nil {
		return kaErr
	}
	go func() {
		for range ch {
		}
	}()
	return nil
}

// Lookup returns the Roma gRPC dial target for zone/shard.
func (r *RomaRegistry) Lookup(ctx context.Context, zoneID string, shard uint32) (string, error) {
	if r == nil || r.cli == nil {
		return "", fmt.Errorf("etcd: registry not connected")
	}
	resp, err := r.cli.Get(ctx, zoneKey(zoneID, shard))
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("etcd: no roma endpoint for zone %q shard %d", zoneID, shard)
	}
	return string(resp.Kvs[0].Value), nil
}
