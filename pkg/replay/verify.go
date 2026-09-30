package replay

import "errors"

// VerifyTerminal checks recording metadata and that recomputed state hash matches the trailer.
func VerifyTerminal(rec Recording, stateHash uint64) error {
	if err := VerifyConsistency(rec); err != nil {
		return err
	}
	if rec.FinalHash != stateHash {
		return errors.New("final state hash mismatch")
	}
	return nil
}
