package janus

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type fakeResolver struct {
	target string
	err    error
}

func (f *fakeResolver) Lookup(_ context.Context, _ string, _ uint32) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.target, nil
}

type countingResolver struct {
	calls  atomic.Int32
	target string
	err    error
}

func (c *countingResolver) Lookup(_ context.Context, _ string, _ uint32) (string, error) {
	c.calls.Add(1)
	if c.err != nil {
		return "", c.err
	}
	return c.target, nil
}

func TestNormalizeZoneID(t *testing.T) {
	if NormalizeZoneID("  east  ") != "east" {
		t.Fatal("trim")
	}
	if NormalizeZoneID("") != "default" {
		t.Fatal("default")
	}
}

func TestDiscoveryPrefersEtcd(t *testing.T) {
	d := &Discovery{
		Etcd:    &fakeResolver{target: "etcd-roma:9092"},
		Static:  map[string]string{"default": "static:9092"},
		Default: "fallback:9092",
	}
	got, err := d.Lookup(context.Background(), "default")
	if err != nil || got != "etcd-roma:9092" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestDiscoveryStaticFallback(t *testing.T) {
	d := &Discovery{
		Etcd:   &fakeResolver{err: context.Canceled},
		Static: map[string]string{"east": "roma-east:9092"},
	}
	got, err := d.Lookup(context.Background(), "east")
	if err != nil || got != "roma-east:9092" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestDiscoveryCacheSkipsRepeatedEtcd(t *testing.T) {
	res := &countingResolver{target: "etcd-roma:9092"}
	d := &Discovery{
		Etcd:     res,
		CacheTTL: time.Minute,
	}
	ctx := context.Background()
	if _, err := d.Lookup(ctx, "default"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Lookup(ctx, "default"); err != nil {
		t.Fatal(err)
	}
	if res.calls.Load() != 1 {
		t.Fatalf("expected 1 etcd lookup, got %d", res.calls.Load())
	}
	ep, err := d.Resolve(ctx, "default")
	if err != nil || ep.Source != SourceCache {
		t.Fatalf("resolve %+v err=%v", ep, err)
	}
}

func TestDiscoveryStickyAfterEtcdFailure(t *testing.T) {
	res := &countingResolver{target: "etcd-roma:9092"}
	d := &Discovery{
		Etcd:      res,
		CacheTTL:  time.Millisecond,
		StickyTTL: time.Minute,
	}
	ctx := context.Background()
	if _, err := d.Lookup(ctx, "east"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)

	res.err = errors.New("etcd down")
	got, err := d.Lookup(ctx, "east")
	if err != nil || got != "etcd-roma:9092" {
		t.Fatalf("sticky got %q err=%v", got, err)
	}
	ep, _ := d.Resolve(ctx, "east")
	if ep.Source != SourceSticky || !ep.Degraded {
		t.Fatalf("expected sticky degraded, got %+v", ep)
	}
}

func TestDiscoveryBuiltinReportsEtcdError(t *testing.T) {
	d := &Discovery{
		Etcd: &fakeResolver{err: errors.New("unavailable")},
	}
	_, err := d.Lookup(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error when falling through to builtin without static/default")
	}
}

func TestParseZoneEndpointMap(t *testing.T) {
	m, err := ParseZoneEndpointMap("east=roma-east:9092, west=roma-west:9092")
	if err != nil {
		t.Fatal(err)
	}
	if m["east"] != "roma-east:9092" || m["west"] != "roma-west:9092" {
		t.Fatalf("%v", m)
	}
	if _, err := ParseZoneEndpointMap("bad-entry"); err == nil {
		t.Fatal("expected parse error")
	}
}
