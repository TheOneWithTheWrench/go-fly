package remote

import "time"

const RefreshTTL = 24 * time.Hour

func ShouldRefresh(cache Cache, exists bool) bool {
	if !exists {
		return true
	}
	if cache.FetchedAt.IsZero() {
		return true
	}
	if time.Since(cache.FetchedAt) >= RefreshTTL {
		return true
	}
	return false
}
