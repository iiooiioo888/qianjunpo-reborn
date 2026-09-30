package etcd

import "testing"

func TestZoneKeyDefault(t *testing.T) {
	if zoneKey("", 1) != "/qjp/roma/v1/zones/default/shards/1" {
		t.Fatal(zoneKey("", 1))
	}
}

func TestNewRomaRegistryEmptyEndpoints(t *testing.T) {
	if _, err := NewRomaRegistry(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestLookupKeyFormat(t *testing.T) {
	k := zoneKey("east", 2)
	if k != "/qjp/roma/v1/zones/east/shards/2" {
		t.Fatal(k)
	}
}
