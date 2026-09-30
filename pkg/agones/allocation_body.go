package agones

import (
	"encoding/json"
	"fmt"
)

// allocationRequest is the JSON body for Agones GameServerAllocation HTTP allocate.
// Legacy mode uses gameServerSelectors; players-counter mode uses selectors + counters
// matching deploy/agones/gameserverallocation-players.yaml.
type allocationRequest struct {
	Namespace           string               `json:"namespace"`
	Scheduling          string               `json:"scheduling,omitempty"`
	Priorities          []allocationPriority `json:"priorities,omitempty"`
	Selectors           []allocationSelector `json:"selectors,omitempty"`
	GameServerSelectors []allocationSelector `json:"gameServerSelectors,omitempty"`
	Counters            map[string]counterAction `json:"counters,omitempty"`
	Metadata            *allocationMetadata  `json:"metadata,omitempty"`
}

type allocationPriority struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Order string `json:"order"`
}

type allocationSelector struct {
	GameServerState string                       `json:"gameServerState,omitempty"`
	MatchLabels     map[string]string            `json:"matchLabels"`
	Counters        map[string]counterSelector   `json:"counters,omitempty"`
}

type counterSelector struct {
	MinAvailable int `json:"minAvailable"`
}

type counterAction struct {
	Action string `json:"action"`
	Amount int    `json:"amount"`
}

type allocationMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// buildAllocationRequest constructs the POST body for HTTPAllocator.
func buildAllocationRequest(cfg Config, req AllocateRequest) allocationRequest {
	meta := &allocationMetadata{
		Labels: map[string]string{
			"qianjunpo.dev/zone":  req.ZoneID,
			"qianjunpo.dev/shard": shardLabel(req.Shard),
		},
	}
	fleetLabel := map[string]string{
		"agones.dev/fleet": cfg.FleetName,
	}
	if cfg.AllocationPlayersCounter {
		minAvail := cfg.AllocationPlayersMinAvailable
		if minAvail <= 0 {
			minAvail = 1
		}
		incr := cfg.AllocationPlayersIncrement
		if incr <= 0 {
			incr = 1
		}
		playerFilter := map[string]counterSelector{
			"players": {MinAvailable: minAvail},
		}
		return allocationRequest{
			Namespace:  cfg.Namespace,
			Scheduling: "Packed",
			Priorities: []allocationPriority{{
				Type:  "Counter",
				Key:   "players",
				Order: "Ascending",
			}},
			Selectors: []allocationSelector{
				{
					GameServerState: "Ready",
					MatchLabels:     fleetLabel,
					Counters:        playerFilter,
				},
				{
					GameServerState: "Allocated",
					MatchLabels:     fleetLabel,
					Counters:        playerFilter,
				},
			},
			Counters: map[string]counterAction{
				"players": {Action: "Increment", Amount: incr},
			},
			Metadata: meta,
		}
	}
	return allocationRequest{
		Namespace: cfg.Namespace,
		GameServerSelectors: []allocationSelector{{
			MatchLabels: fleetLabel,
		}},
		Metadata: meta,
	}
}

// MarshalAllocationRequestJSON is used by tests and operators to verify body shape.
func MarshalAllocationRequestJSON(cfg Config, req AllocateRequest) ([]byte, error) {
	return json.Marshal(buildAllocationRequest(cfg, req))
}

func shardLabel(shard uint32) string {
	return fmt.Sprintf("%d", shard)
}
