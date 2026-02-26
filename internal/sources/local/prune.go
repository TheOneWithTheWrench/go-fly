package local

import "time"

const PruneTTL = 24 * time.Hour

func ShouldPrune(state PruneState, exists bool) bool {
	if !exists {
		return true
	}
	if state.LastPrunedAt.IsZero() {
		return true
	}
	if time.Since(state.LastPrunedAt) >= PruneTTL {
		return true
	}
	return false
}
