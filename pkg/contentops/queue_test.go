package contentops

import "testing"

func TestReviewQueueLifecycle(t *testing.T) {
	q := NewReviewQueue()
	d, err := q.Create("quest", "stub quest text")
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != StatusPending {
		t.Fatalf("status=%s", d.Status)
	}
	approved, err := q.Approve(d.ID)
	if err != nil || approved.Status != StatusApproved {
		t.Fatalf("approve: %v %v", approved, err)
	}
	rolled, err := q.Rollback(d.ID)
	if err != nil || rolled.Status != StatusRolledBack {
		t.Fatalf("rollback: %v %v", rolled, err)
	}
	if _, err := q.Approve(d.ID); err == nil {
		t.Fatal("expected approve after rollback to fail")
	}
}
