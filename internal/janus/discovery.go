package janus

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	etcdreg "github.com/iiooiioo888/qianjunpo-reborn/pkg/discovery/etcd"
)

const (
	defaultDiscoveryCacheTTL  = 30 * time.Second
	defaultDiscoveryStickyTTL = 5 * time.Minute
	builtinRomaEndpoint       = "roma:9092"
)

// RomaResolver resolves Roma gRPC targets for a zone shard.
type RomaResolver interface {
	Lookup(ctx context.Context, zoneID string, shard uint32) (string, error)
}

// ResolutionSource describes how a Roma endpoint was chosen.
type ResolutionSource string

const (
	SourceEtcd    ResolutionSource = "etcd"
	SourceCache   ResolutionSource = "cache"
	SourceSticky  ResolutionSource = "sticky"
	SourceStatic  ResolutionSource = "static"
	SourceDefault ResolutionSource = "default"
	SourceBuiltin ResolutionSource = "builtin"
)

// ZoneEndpoint is the outcome of Roma discovery for a zone.
type ZoneEndpoint struct {
	Target string
	Source ResolutionSource
	// Degraded is true when etcd failed and a non-etcd fallback was used after a prior etcd success.
	Degraded bool
}

// Discovery resolves Roma endpoints via etcd with static fallback.
type Discovery struct {
	Etcd    RomaResolver
	Static  map[string]string
	Default string
	// CacheTTL limits repeated etcd lookups for the same zone (0 uses defaultDiscoveryCacheTTL).
	CacheTTL time.Duration
	// StickyTTL keeps the last etcd target when etcd is temporarily unavailable (0 uses defaultDiscoveryStickyTTL).
	StickyTTL time.Duration

	mu    sync.Mutex
	cache map[string]cachedZoneEndpoint
}

type cachedZoneEndpoint struct {
	target      string
	source      ResolutionSource
	freshUntil  time.Time
	stickyUntil time.Time
}

// NormalizeZoneID trims whitespace and maps empty ids to "default".
func NormalizeZoneID(zoneID string) string {
	zoneID = strings.TrimSpace(zoneID)
	if zoneID == "" {
		return "default"
	}
	return zoneID
}

// Lookup returns a Roma gRPC target for zone id (shard 0 when not specified).
func (d *Discovery) Lookup(ctx context.Context, zoneID string) (string, error) {
	ep, err := d.Resolve(ctx, zoneID)
	if err != nil {
		return "", err
	}
	return ep.Target, nil
}

// Resolve returns the Roma endpoint and how it was resolved (for tests and observability).
func (d *Discovery) Resolve(ctx context.Context, zoneID string) (ZoneEndpoint, error) {
	zoneID = NormalizeZoneID(zoneID)
	if d == nil {
		return ZoneEndpoint{Target: builtinRomaEndpoint, Source: SourceBuiltin}, nil
	}

	now := time.Now()
	if ep, ok := d.getFresh(zoneID, now); ok {
		return ep, nil
	}

	var etcdErr error
	if d.Etcd != nil {
		target, err := d.Etcd.Lookup(ctx, zoneID, 0)
		if err == nil && strings.TrimSpace(target) != "" {
			d.remember(zoneID, strings.TrimSpace(target), SourceEtcd, now)
			return ZoneEndpoint{Target: target, Source: SourceEtcd}, nil
		}
		etcdErr = err
		if ep, ok := d.getSticky(zoneID, now); ok {
			ep.Degraded = true
			return ep, nil
		}
	}

	if ep, ok := d.staticEndpoint(zoneID); ok {
		return ep, nil
	}
	if ep, ok := d.defaultEndpoint(); ok {
		return ep, nil
	}

	if etcdErr != nil {
		return ZoneEndpoint{Target: builtinRomaEndpoint, Source: SourceBuiltin},
			fmt.Errorf("janus: roma discovery for zone %q failed (etcd: %v); using builtin %q", zoneID, etcdErr, builtinRomaEndpoint)
	}
	return ZoneEndpoint{Target: builtinRomaEndpoint, Source: SourceBuiltin}, nil
}

func (d *Discovery) cacheTTL() time.Duration {
	if d == nil || d.CacheTTL <= 0 {
		return defaultDiscoveryCacheTTL
	}
	return d.CacheTTL
}

func (d *Discovery) stickyTTL() time.Duration {
	if d == nil || d.StickyTTL <= 0 {
		return defaultDiscoveryStickyTTL
	}
	return d.StickyTTL
}

func (d *Discovery) getFresh(zoneID string, now time.Time) (ZoneEndpoint, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.cacheEntry(zoneID)
	if !ok || now.After(entry.freshUntil) {
		return ZoneEndpoint{}, false
	}
	return ZoneEndpoint{Target: entry.target, Source: SourceCache}, true
}

func (d *Discovery) getSticky(zoneID string, now time.Time) (ZoneEndpoint, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	entry, ok := d.cacheEntry(zoneID)
	if !ok || entry.source != SourceEtcd || now.After(entry.stickyUntil) {
		return ZoneEndpoint{}, false
	}
	return ZoneEndpoint{Target: entry.target, Source: SourceSticky}, true
}

func (d *Discovery) remember(zoneID, target string, source ResolutionSource, now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cache == nil {
		d.cache = make(map[string]cachedZoneEndpoint)
	}
	d.cache[zoneID] = cachedZoneEndpoint{
		target:      target,
		source:      source,
		freshUntil:  now.Add(d.cacheTTL()),
		stickyUntil: now.Add(d.stickyTTL()),
	}
}

func (d *Discovery) cacheEntry(zoneID string) (cachedZoneEndpoint, bool) {
	if d.cache == nil {
		return cachedZoneEndpoint{}, false
	}
	entry, ok := d.cache[zoneID]
	return entry, ok
}

func (d *Discovery) staticEndpoint(zoneID string) (ZoneEndpoint, bool) {
	if d.Static == nil {
		return ZoneEndpoint{}, false
	}
	if ep, ok := d.Static[zoneID]; ok && strings.TrimSpace(ep) != "" {
		return ZoneEndpoint{Target: strings.TrimSpace(ep), Source: SourceStatic}, true
	}
	if zoneID != "default" {
		if ep, ok := d.Static["default"]; ok && strings.TrimSpace(ep) != "" {
			return ZoneEndpoint{Target: strings.TrimSpace(ep), Source: SourceStatic}, true
		}
	}
	return ZoneEndpoint{}, false
}

func (d *Discovery) defaultEndpoint() (ZoneEndpoint, bool) {
	if d.Default != "" {
		return ZoneEndpoint{Target: strings.TrimSpace(d.Default), Source: SourceDefault}, true
	}
	return ZoneEndpoint{}, false
}

// ParseZoneEndpointMap parses comma-separated zone=target pairs (target may contain ':').
// Example: "east=roma-east:9092,west=roma-west:9092".
func ParseZoneEndpointMap(raw string) (map[string]string, error) {
	out := make(map[string]string)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out, nil
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		zone, target, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("janus: invalid zone endpoint entry %q (want zone=host:port)", part)
		}
		zone = NormalizeZoneID(zone)
		target = strings.TrimSpace(target)
		if target == "" {
			return nil, fmt.Errorf("janus: empty roma target for zone %q", zone)
		}
		out[zone] = target
	}
	return out, nil
}

// EtcdResolver adapts pkg/discovery/etcd.RomaRegistry to RomaResolver.
type EtcdResolver struct {
	Reg *etcdreg.RomaRegistry
}

func (e *EtcdResolver) Lookup(ctx context.Context, zoneID string, shard uint32) (string, error) {
	if e == nil || e.Reg == nil {
		return "", fmt.Errorf("janus: etcd resolver not configured")
	}
	zoneID = NormalizeZoneID(zoneID)
	return e.Reg.Lookup(ctx, zoneID, shard)
}
