package janus

import (
	"context"
	"fmt"
	"strings"

	etcdreg "github.com/iiooiioo888/qianjunpo-reborn/pkg/discovery/etcd"
)

// RomaResolver resolves Roma gRPC targets for a zone shard.
type RomaResolver interface {
	Lookup(ctx context.Context, zoneID string, shard uint32) (string, error)
}

// Discovery resolves Roma endpoints via etcd with static fallback.
type Discovery struct {
	Etcd    RomaResolver
	Static  map[string]string
	Default string
}

// Lookup returns a Roma gRPC target for zone id (shard 0 when not specified).
func (d *Discovery) Lookup(ctx context.Context, zoneID string) (string, error) {
	zoneID = strings.TrimSpace(zoneID)
	if zoneID == "" {
		zoneID = "default"
	}
	if d != nil && d.Etcd != nil {
		target, err := d.Etcd.Lookup(ctx, zoneID, 0)
		if err == nil && target != "" {
			return target, nil
		}
	}
	if d != nil && d.Static != nil {
		if ep, ok := d.Static[zoneID]; ok && ep != "" {
			return ep, nil
		}
		if ep, ok := d.Static["default"]; ok && ep != "" {
			return ep, nil
		}
	}
	if d != nil && d.Default != "" {
		return d.Default, nil
	}
	return "roma:9092", nil
}

// EtcdResolver adapts pkg/discovery/etcd.RomaRegistry to RomaResolver.
type EtcdResolver struct {
	Reg *etcdreg.RomaRegistry
}

func (e *EtcdResolver) Lookup(ctx context.Context, zoneID string, shard uint32) (string, error) {
	if e == nil || e.Reg == nil {
		return "", fmt.Errorf("janus: etcd resolver not configured")
	}
	return e.Reg.Lookup(ctx, zoneID, shard)
}
