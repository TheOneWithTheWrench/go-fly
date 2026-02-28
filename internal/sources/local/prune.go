package local

import "time"

const PruneTTL = 24 * time.Hour
const PruneLaunchCooldown = 30 * time.Second

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

func ShouldLaunchPrune(state PruneState, exists bool, now time.Time) bool {
	if !exists {
		return true
	}
	if state.StartedAt.IsZero() {
		return true
	}
	if now.Sub(state.StartedAt) >= PruneLaunchCooldown {
		return true
	}

	return false
}
