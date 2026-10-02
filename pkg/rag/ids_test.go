package rag

import "testing"

func TestDocumentIDs(t *testing.T) {
	ids := DocumentIDs([]Document{{ID: "a"}, {ID: "b"}})
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("%v", ids)
	}
	if DocumentIDs(nil) != nil {
		t.Fatal("expected nil")
	}
}
