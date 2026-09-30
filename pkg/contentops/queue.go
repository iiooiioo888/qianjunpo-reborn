package contentops

import (
	"fmt"
	"sync"
	"time"
)

// Status is the review lifecycle for LLM-generated dynamic content.
type Status string

const (
	StatusPending    Status = "pending"
	StatusApproved   Status = "approved"
	StatusRolledBack Status = "rolled_back"
)

// Draft is one generated content item awaiting ops review.
type Draft struct {
	ID        string
	Kind      string
	Body      string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int
}

// ReviewQueue is an in-memory ops-review queue (Phase 5 stub; replace with DB in prod).
type ReviewQueue struct {
	mu    sync.Mutex
	items map[string]*Draft
	seq   int
}

func NewReviewQueue() *ReviewQueue {
	return &ReviewQueue{items: map[string]*Draft{}}
}

// Create inserts a pending draft.
func (q *ReviewQueue) Create(kind, body string) (*Draft, error) {
	if body == "" {
		return nil, fmt.Errorf("contentops: empty body")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.seq++
	id := fmt.Sprintf("draft-%d", q.seq)
	now := time.Now().UTC()
	d := &Draft{
		ID:        id,
		Kind:      kind,
		Body:      body,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
	q.items[id] = d
	return cloneDraft(d), nil
}

// Approve marks a pending draft as approved.
func (q *ReviewQueue) Approve(id string) (*Draft, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	d, ok := q.items[id]
	if !ok {
		return nil, fmt.Errorf("contentops: not found %q", id)
	}
	if d.Status != StatusPending {
		return nil, fmt.Errorf("contentops: approve from %s", d.Status)
	}
	d.Status = StatusApproved
	d.UpdatedAt = time.Now().UTC()
	d.Version++
	return cloneDraft(d), nil
}

// Rollback reverts an approved draft (keeps audit trail in-memory).
func (q *ReviewQueue) Rollback(id string) (*Draft, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	d, ok := q.items[id]
	if !ok {
		return nil, fmt.Errorf("contentops: not found %q", id)
	}
	if d.Status != StatusApproved {
		return nil, fmt.Errorf("contentops: rollback from %s", d.Status)
	}
	d.Status = StatusRolledBack
	d.UpdatedAt = time.Now().UTC()
	d.Version++
	return cloneDraft(d), nil
}

// Get returns a copy of a draft by id.
func (q *ReviewQueue) Get(id string) (*Draft, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	d, ok := q.items[id]
	if !ok {
		return nil, false
	}
	return cloneDraft(d), true
}

func cloneDraft(d *Draft) *Draft {
	cp := *d
	return &cp
}
