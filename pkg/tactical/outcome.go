package tactical

import "encoding/json"

// NoWinner is the draw / unset winner sentinel (exported for tests and tooling).
const NoWinner = 255

// EndReason describes why a match stopped.
type EndReason uint8

const (
	EndNone         EndReason = 0
	EndAnnihilation EndReason = 1
	EndTimeout      EndReason = 2
	EndMutualWipe   EndReason = 3
	EndCapture      EndReason = 4
)

// String returns the stable gameplay token used in view snapshot JSON (endReason).
func (r EndReason) String() string {
	switch r {
	case EndAnnihilation:
		return "EndAnnihilation"
	case EndTimeout:
		return "EndTimeout"
	case EndMutualWipe:
		return "EndMutualWipe"
	case EndCapture:
		return "EndCapture"
	default:
		return "EndNone"
	}
}

// MarshalJSON emits the gameplay name (e.g. "EndAnnihilation") for display-layer snapshots.
func (r EndReason) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON accepts the gameplay name or a legacy numeric wire value.
func (r *EndReason) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var name string
		if err := json.Unmarshal(data, &name); err != nil {
			return err
		}
		*r = ParseEndReason(name)
		return nil
	}
	var n uint8
	if err := json.Unmarshal(data, &n); err != nil {
		return err
	}
	*r = EndReason(n)
	return nil
}

// ParseEndReason maps snapshot JSON tokens to EndReason.
func ParseEndReason(name string) EndReason {
	switch name {
	case "EndAnnihilation":
		return EndAnnihilation
	case "EndTimeout":
		return EndTimeout
	case "EndMutualWipe":
		return EndMutualWipe
	case "EndCapture":
		return EndCapture
	default:
		return EndNone
	}
}
