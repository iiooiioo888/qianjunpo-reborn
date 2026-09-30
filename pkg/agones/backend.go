package agones

import "context"

// Allocator claims a GameServer from Fleet / buffer (mock or HTTP allocation API).
type Allocator interface {
	Allocate(ctx context.Context, req AllocateRequest) (RoomIdentity, error)
}

// GameServerSide reports health and drives Ready / Shutdown on the running Roma pod.
type GameServerSide interface {
	Identity(ctx context.Context) (RoomIdentity, error)
	Ready(ctx context.Context) error
	Shutdown(ctx context.Context) error
	Health(ctx context.Context) error
}
