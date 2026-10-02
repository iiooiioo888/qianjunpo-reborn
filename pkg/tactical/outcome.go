package tactical

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

// EndReasonName is the display-layer token for ViewSnapshot.endReason (JSON).
func EndReasonName(r EndReason) string {
	switch r {
	case EndAnnihilation:
		return "wipeout"
	case EndTimeout:
		return "timeout"
	case EndMutualWipe:
		return "mutual_wipe"
	case EndCapture:
		return "occupy"
	default:
		return "none"
	}
}
