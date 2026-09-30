package janus

import (
	"context"
	"testing"
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
